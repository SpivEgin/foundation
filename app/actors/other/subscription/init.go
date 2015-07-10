package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"time"
	"fmt"
	"github.com/ottemo/foundation/utils"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/checkout"
)

// init makes package self-initialization routine before app start
func init() {
	db.RegisterOnDatabaseStart(setupDB)
	api.RegisterOnRestServiceStart(setupAPI)
	app.OnAppStart(onAppStart)
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

	// TODO: check database type for duration (we need month or some think like this)
	collection.AddColumn("period", db.ConstTypeDatetime, false)

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
		for record := range subscriptionsOnSubmit {

			subscriptionRecord := utils.InterfaceToMap(record)
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
			_, err := checkoutInstance.Submit()
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
	subscriptionCollection.AddFilter("date", ">=", currentDay.Add(-ConstTimeDay * 8))
	subscriptionCollection.AddFilter("date", "<", currentDay.Add(-ConstTimeDay * 7))
	subscriptionCollection.AddFilter("status", "=", ConstSubscriptionStatusSuspended)

	fmt.Println(currentDay.Add(-ConstTimeDay * 8), currentDay.Add(-ConstTimeDay * 7))

	subscriptionsToConfirm, err := subscriptionCollection.Load()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	for record := range subscriptionsToConfirm {

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
