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

	err = api.GetRestService().RegisterAPI("subscription/:subscriptionID/confirm", api.ConstRESTOperationGet, APIConfirmSubscription)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("subscription/:subscriptionID/submit", api.ConstRESTOperationGet, APISubmitSubscription)
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
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "29e6d0c1-c7d0-433f-8d4f-a9bebafef76b", "59f8171f-af26-403f-93fa-f67ab6103adc")
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
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5d438cbd-60a3-44af-838f-bddf4e19364e", "59f8171f-af26-403f-93fa-f67ab6103adc")
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
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5d438cbd-60a3-44af-838f-bddf4e19364e", "you are not logined in")
	}

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIConfirmSubscription set subscription status to confirmed that allow it to be procceed
//   - subscription id should be specified in "subscriptionID" argument
func APIConfirmSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	subscriptionID := context.GetRequestArgument("subscriptionID")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d61ff7fe-2a22-43be-8b23-3d56f39c94db", "subscription id should be specified")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("_id", "=", subscriptionID)

	dbRecords, err := subscriptionCollection.Load()

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6b37c9c0-1f6d-4b00-b7de-de326d30f8dd", "subscription not found")
	}

	subscription := utils.InterfaceToMap(dbRecords[0])
	if utils.InterfaceToString(subscription["status"]) == ConstSubscriptionStatusConfirmed {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e411dcd1-234b-4dcf-8072-239755b5fa34", "subscription already confirmed")
	}

	subscription["status"] = ConstSubscriptionStatusConfirmed

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APICreateSubscription provide mechanism to create new subscription
func APICreateSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check visitor rights
	visitorID := visitor.GetCurrentVisitorID(context)
	if api.ValidateAdminRights(context) != nil && visitorID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e6109c04-e35a-4a90-9593-4cc1f141a358", "you are not logined in")
	}

	result := make(map[string]interface{})
	submittableOrder := false

	// check request context
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionDate := utils.GetFirstMapValue(requestData, "date")
	if subscriptionDate == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "43873ddc-a817-4216-aa3c-9b004d96a539", "subscription Date can't be blank")
	}

	subscriptionPeriod := utils.GetFirstMapValue(requestData, "period")
	if subscriptionPeriod == nil {
		subscriptionPeriod = 1
	}

	orderID, present := requestData["orderID"]

	// try to create new subscription with existing order
	if present {

		orderModel, err := order.LoadOrderByID(utils.InterfaceToString(orderID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		orderVisitorID := utils.InterfaceToString(orderModel.Get("visitor_id"))

		if api.ValidateAdminRights(context) != nil && visitorID != orderVisitorID {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4916bf20-e053-472e-98e1-bb28b7c867a1", "you are trying to use vicarious order")
		}

		// update cart and checkout object for current session
		duplicateCheckout, err := orderModel.DuplicateOrder(nil)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		checkoutInstance, ok := duplicateCheckout.(checkout.InterfaceCheckout)
		if !ok {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "946c3598-53b4-4dad-9d6f-23bf1ed6440f", "order can't be typed")
		}

		duplicateCart := checkoutInstance.GetCart()

		err = duplicateCart.Delete()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// check order for possibility to proceed automatically
		if paymentInfo := utils.InterfaceToMap(orderModel.Get("payment_info")); paymentInfo != nil {
			if _, present := paymentInfo["transactionID"]; present {
				submittableOrder = true
			}
		}
	} else {
		// this case can be removed but need to be sure in handler for it (ConstSubscriptionCheckoutProcessingCreate)
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d13e49d1-f73a-4bbc-b564-4258c99921c9", "orderID can't be blank")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result = map[string]interface{}{
		"visitor_id":         visitorID,
		"order_id":           utils.InterfaceToString(orderID),
		"date":               utils.InterfaceToTime(subscriptionDate),
		"period":             utils.InterfaceToInt(subscriptionPeriod),
		"status":             ConstSubscriptionStatusSuspended,
		"checkoutProcessing": ConstSubscriptionCheckoutProcessingUpdate,
	}

	if submittableOrder {
		result["checkoutProcessing"] = ConstSubscriptionCheckoutProcessingSubmit
	}

	//	if orderID == nil {
	//		result["checkoutProcessing"] = ConstSubscriptionCheckoutProcessingCreate
	//	}

	_, err = subscriptionCollection.Save(result)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return result, nil
}

// APISubmitSubscription give current session new checkout and card from subscription  and try to proceed it
//   - subscription id should be specified in "subscriptionID" argument
func APISubmitSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	subscriptionID := context.GetRequestArgument("subscriptionID")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "027e7ef9-b202-475b-a242-02e2d0d74ce6", "subscription id should be specified")
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	subscriptionCollection.AddFilter("_id", "=", subscriptionID)

	dbRecords, err := subscriptionCollection.Load()

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "aecec9b2-3b02-40cb-b163-39cb03b53252", "subscription not found")
	}

	subscription := utils.InterfaceToMap(dbRecords[0])

	if utils.InterfaceToString(subscription["status"]) != ConstSubscriptionStatusConfirmed {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "153ee2dd-3e3f-42ac-b669-1d15ec741547", "subscription not confirmed")
	}

	currentSession := context.GetSession()

	if orderID, present := subscription["order_id"]; present && orderID != nil {

		orderModel, err := order.LoadOrderByID(utils.InterfaceToString(orderID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// update cart and checkout object for current session
		duplicateCheckout, err := orderModel.DuplicateOrder(nil)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		checkoutInstance, ok := duplicateCheckout.(checkout.InterfaceCheckout)
		if !ok {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3788a54b-6ef6-486f-9819-c85e34ff43c5", "order can't be typed")
		}

		// rewrite current checkout and cart by newly created from duplicate order
		currentCheckout, err := checkout.GetCurrentCheckout(context)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		currentCart := currentCheckout.GetCart()

		err = currentCart.Deactivate()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = currentCart.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = checkoutInstance.SetSession(currentSession)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = checkoutInstance.SetInfo("subscription", subscription)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		duplicateCart := checkoutInstance.GetCart()

		err = duplicateCart.SetSessionID(currentSession.GetID())
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = duplicateCart.Activate()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = duplicateCart.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		if subscription["checkoutProcessing"] == ConstSubscriptionCheckoutProcessingSubmit {
			_, err := checkoutInstance.Submit()
			if err != nil {
				currentSession.Set(cart.ConstSessionKeyCurrentCart, currentCart.GetID())
				currentSession.Set(checkout.ConstSessionKeyCurrentCheckout, checkoutInstance)
				return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, env.ErrorDispatch(err)
			}
		} else {
			currentSession.Set(cart.ConstSessionKeyCurrentCart, currentCart.GetID())
			currentSession.Set(checkout.ConstSessionKeyCurrentCheckout, checkoutInstance)
			return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, nil
		}
	}

	subscription["status"] = ConstSubscriptionStatusSuspended

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}
