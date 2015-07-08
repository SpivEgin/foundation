package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("subscriptions", api.ConstRESTOperationGet, APIListSubscriptions)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("subscription", api.ConstRESTOperationCreate, APICreateSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("subscription/:subscriptionID", api.ConstRESTOperationGet, APIGetSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("subscription/:subscriptionID", api.ConstRESTOperationDelete, APIDeleteSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("subscription/:subscriptionID/suspend", api.ConstRESTOperationGet, APISuspendSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// APIListSubscriptions returns a list of subscriptions for visitor
//   - if "action" parameter is set to "count" result value will be just a number of list items
func APIListSubscriptions(context api.InterfaceApplicationContext) (interface{}, error) {

	visitorID := visitor.GetCurrentVisitorID(context)
	if visitorID == "" {
		return "you are not logined in", nil
	}

	// making database request
	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("visitor_id", "=", visitorID)

	dbRecords, err := subscriptionCollection.Load()

	return dbRecords, env.ErrorDispatch(err)
}

// APIGetSubscription return specified subscription information
//   - subscription id should be specified in "subscriptionID" argument
func APIGetSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	subscriptionID := context.GetRequestArgument("subscriptionID")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "99b0a49b-9fe4-4f64-9879-bf5a45ff5ac7", "subscription id should be specified")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("_id", "=", subscriptionID)

	dbRecords, err := subscriptionCollection.Load()

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d724cdc3-5bb7-494b-9a8a-952fdc311bd0", "subscription not found")
	}

	return dbRecords[0], nil
}

// APIDeleteSubscription deletes existing purchase order
//   - subscription id should be specified in "subscriptionID" argument
func APIDeleteSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	subscriptionID := context.GetRequestArgument("subscriptionID")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "67bedbe8-7426-437b-9dbc-4840f13e619e", "subscription id should be specified")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("_id", "=", subscriptionID)

	dbRecords, err := subscriptionCollection.Load()

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6c9559d5-c0fe-4fa1-a07b-4e7b6ac1dad6", "subscription not found")
	}

	visitorID := visitor.GetCurrentVisitorID(context)
	if api.ValidateAdminRights(context) == nil || visitorID == dbRecords[0]["visitor_id"] {
		subscriptionCollection.DeleteByID(subscriptionID)
	} else {
		return "you are not logined in", nil
	}

	return "ok", nil
}

// APISuspendSubscription suspend subscription
//   - subscription id should be specified in "subscriptionID" argument
func APISuspendSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	subscriptionID := context.GetRequestArgument("subscriptionID")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4e8f9873-9144-42ae-b119-d1e95bb1bbfd", "subscription id should be specified")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("_id", "=", subscriptionID)

	dbRecords, err := subscriptionCollection.Load()

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "59e4ab86-7726-4e3d-bec8-7ef5bf0ebbbf", "subscription not found")
	}

	subscription := utils.InterfaceToMap(dbRecords[0])

	subscription["status"] = ConstSubscriptionStatusSuspended

	visitorID := visitor.GetCurrentVisitorID(context)
	if api.ValidateAdminRights(context) != nil && visitorID != subscription["visitor_id"] {
		return "you are not logined in", nil
	}

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APICreateSubscription provide mechanism to create new subscription
func APICreateSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	orderID := context.GetRequestArgument("confirm")
	if orderID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "8e115c53-caa0-44d1-87f6-27cc5062aca3", "something go wrong")
	}

	// rewrite current checkout and cart by newly created from duplicate order
	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentSession := context.GetSession()

	currentCart := currentCheckout.GetCart()

	err = currentCart.Deactivate()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCart.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// update cart and checkout object for current session
	orderModel, err := order.LoadOrderByID(orderID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	duplicateCheckout, err := orderModel.DuplicateOrder(nil)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	checkoutInstance, ok := duplicateCheckout.(checkout.InterfaceCheckout)
	if !ok {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "946c3598-53b4-4dad-9d6f-23bf1ed6440f", "order can't be typed")
	}

	err = checkoutInstance.SetSession(currentSession)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentCart = checkoutInstance.GetCart()

	err = currentCart.SetSessionID(currentSession.GetID())
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCart.Activate()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCart.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentSession.Set(cart.ConstSessionKeyCurrentCart, currentCart.GetID())
	currentSession.Set(checkout.ConstSessionKeyCurrentCheckout, checkoutInstance)

	return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, nil
}
