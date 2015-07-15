package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"strings"
	"time"
)

// init makes package self-initialization routine before app start
func init() {
	db.RegisterOnDatabaseStart(setupDB)
	api.RegisterOnRestServiceStart(setupAPI)
	app.OnAppStart(onAppStart)
	env.RegisterOnConfigStart(setupConfig)
}

// setupDB prepares system database for package usage
func setupDB() error {

	collection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	collection.AddColumn("order_id", db.ConstTypeID, true)
	collection.AddColumn("visitor_id", db.ConstTypeID, true)
	collection.AddColumn("cart_id", db.ConstTypeID, true)

	// a date on which client set a date to bill order
	collection.AddColumn("date", db.ConstTypeDatetime, true)
	collection.AddColumn("period", db.ConstTypeInteger, false)

	collection.AddColumn("status", db.TypeWPrecision(db.ConstTypeVarchar, 50), false)
	collection.AddColumn("action", db.TypeWPrecision(db.ConstTypeVarchar, 50), false)

	return nil
}

// Function for every day checking for email sent to customers whoes subscription need to be confirmed
// and when need to bill orders
func schedulerFunc(params map[string]interface{}) error {

	currentDay := time.Now().Truncate(ConstTimeDay)

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	submitEmailSubject := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionSubmitEmailSubject))
	submitEmailTemplate := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionSubmitEmailTemplate))
	submitEmailLink := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionSubmitEmailLink))

	subscriptionCollection.AddFilter("date", ">=", currentDay)
	subscriptionCollection.AddFilter("date", "<", currentDay.Add(ConstTimeDay))

	//	get subscriptions with current day date and do action
	subscriptionsOnSubmit, err := subscriptionCollection.Load()
	if err == nil {
		for _, record := range subscriptionsOnSubmit {

			subscriptionRecord := utils.InterfaceToMap(record)

			subscriptionID := utils.InterfaceToString(subscriptionRecord["_id"])
			subscriptionCheckoutAction := utils.InterfaceToString(subscriptionRecord["action"])

			// subscriptionNextDate add to subscriptionDate month * period
			subscriptionDate := utils.InterfaceToTime(subscriptionRecord["date"])
			subscriptionNextDate := subscriptionDate.AddDate(0, utils.InterfaceToInt(subscriptionRecord["period"]), 0)

			subscriptionStatus := utils.InterfaceToString(subscriptionRecord["status"])

			// bill orders which subscription date is today and status is confirmed
			if subscriptionStatus == ConstSubscriptionStatusConfirmed {

				proceedCheckoutLink := app.GetStorefrontURL(strings.Replace(submitEmailLink, "{{subscriptionID}}", subscriptionID, 1))

				// submitting orders which orders are allow to do this, in case of submit error we make a go to checkout email
				orderID, orderPresent := subscriptionRecord["order_id"]

				if orderPresent && subscriptionCheckoutAction == ConstSubscriptionActionSubmit {
					orderModel, err := order.LoadOrderByID(utils.InterfaceToString(orderID))
					if err != nil {
						env.LogError(err)
						continue
					}

					// check for stock availability of products
					newCheckout, err := orderModel.DuplicateOrder(nil)
					if err != nil {
						env.LogError(err)
						continue
					}

					checkoutInstance, ok := newCheckout.(checkout.InterfaceCheckout)
					if !ok {
						env.LogError(env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e0e5b596-fbb7-406b-b540-445c2f2e1790", "order can't be typed"))
						continue
					}

					err = checkoutInstance.SetInfo("subscription", subscriptionID)
					if err != nil {
						env.LogError(err)
					}

					// need to check for unreached payment
					// to send email to user in case of low balance on credit card
					_, err = checkoutInstance.Submit()
					if err != nil {
						env.LogError(err)

						err = sendConfirmationEmail(subscriptionRecord, proceedCheckoutLink, submitEmailTemplate, submitEmailSubject)
						if err != nil {
							env.LogError(err)
							continue
						}

						subscriptionRecord["action"] = ConstSubscriptionActionUpdate
					}

					subscriptionRecord["date"] = subscriptionNextDate
					subscriptionRecord["status"] = ConstSubscriptionStatusSuspended

					_, err = subscriptionCollection.Save(subscriptionRecord)
					if err != nil {
						env.LogError(err)
						continue
					}

				} else {

					err = sendConfirmationEmail(subscriptionRecord, proceedCheckoutLink, submitEmailTemplate, submitEmailSubject)
					if err != nil {
						env.LogError(err)
					}
				}
			} else {
				subscriptionRecord["date"] = subscriptionNextDate
				_, err = subscriptionCollection.Save(subscriptionRecord)
				if err != nil {
					env.LogError(err)
					continue
				}
			}

		}
	} else {
		env.LogError(err)
	}

	// send email to subscribers that notifies they are about to receive a shipment for a recurring order 1 week before being billed
	subscriptionCollection.ClearFilters()
	subscriptionCollection.AddFilter("date", ">=", currentDay.Add(-ConstTimeDay*8))
	subscriptionCollection.AddFilter("date", "<", currentDay.Add(-ConstTimeDay*7))
	subscriptionCollection.AddFilter("status", "=", ConstSubscriptionStatusSuspended)

	subscriptionsToConfirm, err := subscriptionCollection.Load()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	// email elements for confirmation emails to subscriptions for which the date of payment will be in a week
	confirmationEmailSubject := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailSubject))
	confirmationEmailTemplate := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailTemplate))
	confirmationEmailLink := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailTemplate))

	for _, record := range subscriptionsToConfirm {
		subscriptionRecord := utils.InterfaceToMap(record)

		subscriptionID := utils.InterfaceToString(subscriptionRecord["_id"])
		confirmationLink := app.GetStorefrontURL(strings.Replace(confirmationEmailLink, "{{subscriptionID}}", subscriptionID, 1))

		err = sendConfirmationEmail(subscriptionRecord, confirmationLink, confirmationEmailTemplate, confirmationEmailSubject)
		if err != nil {

			env.LogError(err)
			continue
		}
	}

	return nil
}

// onAppStart makes module initialization on application startup
func onAppStart() error {

	env.EventRegisterListener("checkout.success", checkoutSuccessHandler)

	if scheduler := env.GetScheduler(); scheduler != nil {
		scheduler.RegisterTask("subscriptionProcess", schedulerFunc)
		scheduler.ScheduleRepeat("0 12 * * *", "subscriptionProcess", nil)
	}

	return nil
}
