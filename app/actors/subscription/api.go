package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/subscription"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	// Administrative
	service.GET("subscriptions", api.IsAdmin(APIListSubscriptions))
	service.GET("subscriptions/:id", api.IsAdmin(APIGetSubscription))
	service.PUT("subscriptions/:id", APIUpdateSubscription)

	// Public
	service.GET("visit/subscriptions", APIListVisitorSubscriptions)
	service.PUT("visit/subscriptions/:id", APIUpdateSubscription)

	// Other thing
	service.GET("subscriptional/checkout", APICheckCheckoutSubscription)

	return nil
}

// APIListSubscriptions returns a list of subscriptions for visitor
//   - if "action" parameter is set to "count" result value will be just a number of list items
func APIListSubscriptions(context api.InterfaceApplicationContext) (interface{}, error) {

	// list operation
	subscriptionCollectionModel, err := subscription.GetSubscriptionCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	models.ApplyFilters(context, subscriptionCollectionModel.GetDBCollection())

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return subscriptionCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	subscriptionCollectionModel.ListLimit(models.GetListLimit(context))

	// extra parameter handle
	models.ApplyExtraAttributes(context, subscriptionCollectionModel)

	return subscriptionCollectionModel.List()
}

// APIListVisitorSubscriptions returns a list of subscriptions for visitor
//   - if "action" parameter is set to "count" result value will be just a number of list items
func APIListVisitorSubscriptions(context api.InterfaceApplicationContext) (interface{}, error) {

	visitorID := visitor.GetCurrentVisitorID(context)
	if visitorID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c73e39c9-dc23-463b-9792-a5d3f7e4d9dd", "You should log in first")
	}

	// for showing subscriptions to a visitor, request is specific so handle it in different way from default List
	subscriptionCollectionModel, err := subscription.GetSubscriptionCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	dbCollection := subscriptionCollectionModel.GetDBCollection()
	dbCollection.AddStaticFilter("visitor_id", "=", visitorID)
	dbCollection.AddStaticFilter("status", "=", subscription.ConstSubscriptionStatusConfirmed)
	models.ApplyFilters(context, dbCollection)

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return dbCollection.Count()
	}

	// limit parameter handle
	dbCollection.SetLimit(models.GetListLimit(context))

	subscriptions := subscriptionCollectionModel.ListSubscriptions()
	var result []map[string]interface{}

	for _, subscriptionItem := range subscriptions {
		result = append(result, subscriptionItem.ToHashMap())
	}

	return result, nil
}

// APIGetSubscription return specified subscription information
//   - subscription id should be specified in "id" argument
func APIGetSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	subscriptionID := context.GetRequestArgument("id")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b626ec0a-a317-4b63-bd05-cc23932bdfe0", "subscription id should be specified")
	}

	subscriptionModel, err := subscription.LoadSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := subscriptionModel.ToHashMap()

	result["payment_method_name"] = subscriptionModel.GetPaymentMethod().GetName()
	result["shipping_method_name"] = subscriptionModel.GetShippingMethod().GetName()

	return result, nil
}

// APICheckCheckoutSubscription provide check is current checkout allows to create new subscription
func APICheckCheckoutSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// check visitor to be registered
	visitorID := visitor.GetCurrentVisitorID(context)
	if visitorID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e6109c04-e35a-4a90-9593-4cc1f141a358", "you are not logged in")
	}

	currentCheckout, err := checkout.GetCurrentCheckout(context, false)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if err := validateCheckoutToSubscribe(currentCheckout); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIUpdateSubscription allows to change status of subscription for visitor and for administrator
func APIUpdateSubscription(context api.InterfaceApplicationContext) (interface{}, error) {

	// validate params
	subscriptionID := context.GetRequestArgument("id")
	if subscriptionID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4e8f9873-9144-42ae-b119-d1e95bb1bbfd", "subscription id should be specified")
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	requestedStatus := utils.InterfaceToString(requestData["status"])
	if requestedStatus == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "71fc926c-d2a0-4c8a-9462-b5274346ed23", "status should be specified")
	}

	subscriptionInstance, err := subscription.LoadSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// validate ownership
	isAdmin := api.ValidateAdminRights(context) == nil
	isOwner := subscriptionInstance.GetVisitorID() == visitor.GetCurrentVisitorID(context)

	if !isAdmin && !isOwner {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bae87bfa-0fa2-4256-ab11-2fffa20bfa00", "Subscription ownership could not be verified")
	}

	err = subscriptionInstance.SetStatus(requestedStatus)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Send cancellation emails
	isCancelled := requestedStatus == subscription.ConstSubscriptionStatusCanceled
	if isCancelled {
		sendCancellationEmail(subscriptionInstance)
	}

	return "ok", subscriptionInstance.Save()
}

func sendCancellationEmail(subscriptionItem subscription.InterfaceSubscription) {
	email := utils.InterfaceToString(subscriptionItem.GetCustomerEmail())
	subject, body := getEmailInfo(subscriptionItem)
	app.SendMail(email, subject, body)
}

func getEmailInfo(subscriptionItem subscription.InterfaceSubscription) (string, string) {
	subject := utils.InterfaceToString(env.ConfigGetValue(subscription.ConstConfigPathSubscriptionCancelEmailSubject))

	siteVariables := map[string]interface{}{
		"Url": app.GetStorefrontURL(""),
	}

	templateVariables := map[string]interface{}{
		"Subscription": subscriptionItem.ToHashMap(),
		"Site":         siteVariables,
	}

	body := utils.InterfaceToString(env.ConfigGetValue(subscription.ConstConfigPathSubscriptionCancelEmailTemplate))
	body, err := utils.TextTemplate(body, templateVariables)
	if err != nil {
		env.ErrorDispatch(err)
	}

	return subject, body
}