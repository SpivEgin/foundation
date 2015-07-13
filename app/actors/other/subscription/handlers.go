package subscription

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// checkoutSuccessHandler is a handler for checkout success event which sends order information to TrustPilot
func checkoutSuccessHandler(event string, eventData map[string]interface{}) bool {

	var currentCheckout checkout.InterfaceCheckout
	if eventItem, present := eventData["checkout"]; present {
		if typedItem, ok := eventItem.(checkout.InterfaceCheckout); ok {
			currentCheckout = typedItem
		}
	}

	subscription := currentCheckout.GetInfo("subscription")
	if subscription == nil {
		return true
	}

	var checkoutOrder order.InterfaceOrder
	if eventItem, present := eventData["order"]; present {
		if typedItem, ok := eventItem.(order.InterfaceOrder); ok {
			checkoutOrder = typedItem
		}
	}

	if checkoutOrder != nil && currentCheckout != nil {
		go subscriptionUpdate(checkoutOrder, utils.InterfaceToMap(subscription))
	}

	return true
}

// subscriptionUpdate is a asynchronously update subscription with new state of the order
func subscriptionUpdate(checkoutOrder order.InterfaceOrder, subscriptionMap map[string]interface{}) error {

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	subscriptionCollection.Load()

	return nil
}
