package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"fmt"
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

	// a date on which client set a date to bill order
	collection.AddColumn("date", db.ConstTypeDatetime, true)
	collection.AddColumn("period", db.ConstTypeInteger, false)

	collection.AddColumn("status", db.TypeWPrecision(db.ConstTypeVarchar, 50), false)

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

	fmt.Println(currentDay, currentDay.Add(ConstTimeDay))

	// bill orders which subscription date is today and status is confirmed
	subscriptionCollection.AddFilter("date", ">=", currentDay)
	subscriptionCollection.AddFilter("date", "<", currentDay.Add(ConstTimeDay))
	subscriptionCollection.AddFilter("status", "=", ConstSubscriptionStatusConfirmed)

	subscriptionsOnSubmit, err := subscriptionCollection.Load()
	if err == nil {
		for subscriptionRecord := range subscriptionsOnSubmit {

			subscriptionRecord := utils.InterfaceToMap(subscriptionRecord)
			orderID, present := subscriptionRecord["order_id"]
			if !present {
				fmt.Println("71")
				env.LogError(env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "946c3598-53b4-4dad-9d6f-23bf1ed6440f", "orderID not present in subscription record"))
				continue
			}

			orderModel, err := order.LoadOrderByID(utils.InterfaceToString(orderID))
			if err != nil {
				fmt.Println(err)
				env.LogError(err)
				continue
			}

			// check for stock availability of products
			newCheckout, err := orderModel.DuplicateOrder(nil)
			if err != nil {
				fmt.Println(err)
				env.LogError(err)
				continue
			}

			checkoutInstance, ok := newCheckout.(checkout.InterfaceCheckout)
			if !ok {
				env.LogError(env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "946c3598-53b4-4dad-9d6f-23bf1ed6440f", "order can't be typed"))
				continue
			}

			// need to check for unreached payment
			// to send email to user in case of low balance on credit card
			// also here possible some another payment errors
			_, err = checkoutInstance.Submit()
			if err != nil {
				fmt.Println(err)
				env.LogError(err)
				continue
			}

			// add to date month * period
			subscriptionNextDate := currentDay.AddDate(0, utils.InterfaceToInt(subscriptionRecord["period"]), 0)
			subscriptionRecord["date"] = subscriptionNextDate
			subscriptionRecord["status"] = ConstSubscriptionStatusSuspended

			_, err = subscriptionCollection.Save(subscriptionRecord)
			if err != nil {
				fmt.Println(err)
				env.LogError(err)
				continue
			}
		}
	} else {
		fmt.Println(err)
		env.LogError(err)
	}

	// send email to subscribers to confirm order placing
	subscriptionCollection.ClearFilters()
	subscriptionCollection.AddFilter("date", ">=", currentDay.Add(-ConstTimeDay*8))
	subscriptionCollection.AddFilter("date", "<", currentDay.Add(-ConstTimeDay*7))
	subscriptionCollection.AddFilter("status", "=", ConstSubscriptionStatusSuspended)

	fmt.Println(currentDay.Add(-ConstTimeDay*8), currentDay.Add(-ConstTimeDay*7))

	subscriptionsToConfirm, err := subscriptionCollection.Load()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	emailSubject := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailSubject))
	emailTemplate := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailTemplate))
	storefrontConfirmationLink := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathSubscriptionEmailTemplate))

	for record := range subscriptionsToConfirm {
		subscriptionRecord := utils.InterfaceToMap(record)
		orderID := utils.InterfaceToString(subscriptionRecord["order_id"])
		subscriptionID := utils.InterfaceToString(subscriptionRecord["_id"])

		orderModel, err := order.LoadOrderByID(orderID)
		if err != nil {
			fmt.Println(err)
			env.LogError(err)
			continue
		}

		visitorMap := map[string]interface{}{
			"name":  orderModel.Get("customer_name"),
			"email": orderModel.Get("customer_email"),
		}

		linkHref := app.GetStorefrontURL(strings.Replace(storefrontConfirmationLink, "{{subscriptionID}}", subscriptionID, 1))

		customInfo := map[string]interface{}{
			"link": linkHref,
		}

		orderMap := orderModel.ToHashMap()

		var orderItems []map[string]interface{}

		for _, item := range orderModel.GetItems() {
			options := make(map[string]interface{})

			for _, optionKeys := range item.GetOptions() {
				optionMap := utils.InterfaceToMap(optionKeys)
				options[utils.InterfaceToString(optionMap["label"])] = optionMap["value"]
			}
			orderItems = append(orderItems, map[string]interface{}{
				"name":    item.GetName(),
				"options": options,
				"sku":     item.GetSku(),
				"qty":     item.GetQty(),
				"price":   item.GetPrice()})
		}

		orderMap["items"] = orderItems

		confirmationEmail, err := utils.TextTemplate(emailTemplate,
			map[string]interface{}{
				"Order":   orderMap,
				"Visitor": visitorMap,
				"Info":    customInfo,
			})

		err = app.SendMail(utils.InterfaceToString(orderModel.Get("customer_email")), emailSubject, confirmationEmail)
		if err != nil {
			return env.ErrorDispatch(err)
		}

	}

	return nil
}

// onAppStart makes module initialization on application startup
func onAppStart() error {

	if scheduler := env.GetScheduler(); scheduler != nil {
		scheduler.RegisterTask("subscriptionProcess", schedulerFunc)
		scheduler.ScheduleRepeat("*/10 * * * *", "checkOrdersToSent", nil)
	}

	return nil
}
