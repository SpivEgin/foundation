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
	"time"
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

// APISuspendSubscription change status of subscription to suspended, it will pass any action on next date
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

	visitorID := visitor.GetCurrentVisitorID(context)
	if api.ValidateAdminRights(context) != nil && visitorID != subscription["visitor_id"] {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5d438cbd-60a3-44af-838f-bddf4e19364e", "you are not logined in")
	}

	if utils.InterfaceToString(subscription["status"]) == ConstSubscriptionStatusSuspended {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "ceb58d22-876a-47fb-a686-017508618313", "subscription already suspended")
	}

	subscription["status"] = ConstSubscriptionStatusSuspended

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIConfirmSubscription set subscription status to confirmed that allow it to be proceed on it's date
//   - subscription id should be specified in "subscriptionID" argument
// TODO: check requirements for this action maybe block unregistered or create key to perform action
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

	subscriptionDateValue := utils.GetFirstMapValue(requestData, "date")
	if subscriptionDateValue == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "43873ddc-a817-4216-aa3c-9b004d96a539", "subscription Date can't be blank")
	}

	timeZone := utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathStoreTimeZone))

	subscriptionDate, _ := utils.MakeUTCTime(utils.InterfaceToTime(subscriptionDateValue), timeZone)
	if subscriptionDate.Before(time.Now().Truncate(ConstTimeDay).Add(ConstTimeDay)) {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c4881529-8b05-4a16-8cd4-6c79d0d79856", "subscription Date cannot be today or earlier")
	}

	subscriptionPeriod := utils.GetFirstMapValue(requestData, "period", "recurrence_period", "recurring")
	if subscriptionPeriod == nil || utils.InterfaceToInt(subscriptionPeriod) < 1 {
		subscriptionPeriod = 1
	}

	if utils.InterfaceToInt(subscriptionPeriod) > 3 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "85f539fa-89fe-4ad8-b171-3b66910bad3f", "subscription recurrence period cannot be more than 3 month")
	}

	orderID, orderPresent := requestData["orderID"]

	cartID := ""

	currentCart, err := cart.GetCurrentCart(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentSession := context.GetSession()

	if !orderPresent && len(currentCart.GetItems()) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d0484d9d-cb6d-48ed-be5f-f77fe19c6dca", "No items in cart or no order subscription specified")
	}

	// try to create new subscription with existing order
	if orderPresent {

		orderModel, err := order.LoadOrderByID(utils.InterfaceToString(orderID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		orderVisitorID := utils.InterfaceToString(orderModel.Get("visitor_id"))
		cartID = utils.InterfaceToString(orderModel.Get("cart_id"))

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
		err = currentCart.ValidateCart()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		cartID = currentCart.GetID()

		err = currentCart.Deactivate()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = currentCart.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		currentSession.Set(cart.ConstSessionKeyCurrentCart, nil)
	}

	subscriptionCollection, err := db.GetCollection(ConstCollectionNameSubscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result = map[string]interface{}{
		"visitor_id": visitorID,
		"order_id":   utils.InterfaceToString(orderID),
		"cart_id":    cartID,
		"date":       subscriptionDate,
		"period":     utils.InterfaceToInt(subscriptionPeriod),
		"status":     ConstSubscriptionStatusSuspended,
		"action":     ConstSubscriptionActionUpdate,
	}

	// for orders that not have transaction by default we set action value to Update or Create
	// that means on subscription Date they will need to proceed checkout one more time
	if submittableOrder && orderID != nil {
		result["action"] = ConstSubscriptionActionSubmit
	}

	if orderID == nil && cartID != "" {
		result["action"] = ConstSubscriptionActionCreate
	}

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
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if len(dbRecords) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "aecec9b2-3b02-40cb-b163-39cb03b53252", "subscription not found")
	}

	subscription := utils.InterfaceToMap(dbRecords[0])

	if utils.InterfaceToString(subscription["status"]) != ConstSubscriptionStatusConfirmed {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "153ee2dd-3e3f-42ac-b669-1d15ec741547", "subscription not confirmed")
	}

	subscriptionDate := utils.InterfaceToTime(subscription["date"])

	currentDay := time.Now().Truncate(ConstTimeDay)

	// when someone try to submit subscription before available date (means submitting email wasn't sented yet)
	if currentDay.Before(subscriptionDate.Truncate(ConstTimeDay)) {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "747ae177-a295-4029-b1dc-4abcce319d7b", "subscription can't be submited yet")
	}

	subscriptionOrderID := subscription["order_id"]
	subscriptionCartID := subscription["cart_id"]

	if subscriptionOrderID == nil && subscriptionCartID == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "780975f0-c24d-452c-ae43-cfbef64b9a1a", "this subscription can't be submited (no cart and order in)")
	}

	// obtain user current cart and checkout for future operations
	currentSession := context.GetSession()

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	currentCart := currentCheckout.GetCart()

	// Duplicating order and set new checkout and cart to current session with redirect to checkout if need to update info
	if subscriptionOrderID != nil {

		orderModel, err := order.LoadOrderByID(utils.InterfaceToString(subscriptionOrderID))
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
		err = checkoutInstance.SetSession(currentSession)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = checkoutInstance.SetInfo("subscription", subscriptionID)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// rewrite current curt with duplicated
		err = currentCart.Deactivate()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		err = currentCart.Save()
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

		if subscription["action"] == ConstSubscriptionActionSubmit {
			_, err := checkoutInstance.Submit()
			if err != nil {
				currentSession.Set(cart.ConstSessionKeyCurrentCart, duplicateCart.GetID())
				currentSession.Set(checkout.ConstSessionKeyCurrentCheckout, checkoutInstance)
				return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, env.ErrorDispatch(err)
			}
		} else {
			currentSession.Set(cart.ConstSessionKeyCurrentCart, duplicateCart.GetID())
			currentSession.Set(checkout.ConstSessionKeyCurrentCheckout, checkoutInstance)
			return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, nil
		}

		// We need to set for user his subscription cart and add to checkout info subscription to handle it on success
	} else {

		err = currentCheckout.SetInfo("subscription", subscriptionID)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		subscriptionCart, err := cart.LoadCartByID(utils.InterfaceToString(subscriptionCartID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		for _, cartItem := range currentCart.GetItems() {
			currentCart.RemoveItem(cartItem.GetIdx())
		}

		for _, cartItem := range subscriptionCart.GetItems() {
			_, err = currentCart.AddItem(cartItem.GetProductID(), cartItem.GetQty(), cartItem.GetOptions())
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}
		}

		err = currentCart.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		currentSession.Set(cart.ConstSessionKeyCurrentCart, utils.InterfaceToString(subscriptionCartID))
		return api.StructRestRedirect{Result: "ok", Location: app.GetStorefrontURL("checkout")}, nil

	}

	// in case of instant checkout submit
	subscriptionNextDate := subscriptionDate.AddDate(0, utils.InterfaceToInt(subscription["period"]), 0)
	subscription["date"] = subscriptionNextDate
	subscription["status"] = ConstSubscriptionStatusSuspended

	_, err = subscriptionCollection.Save(subscription)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}
