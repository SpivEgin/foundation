package subscription

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/app/models/subscription"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	// Administrative
	service.GET("subscriptions", api.IsAdminHandler(APIListSubscriptions))
	service.GET("subscriptions/:id", api.IsAdminHandler(APIGetSubscription))
	service.PUT("subscriptions/:id", APIUpdateSubscription)
	service.GET("update/subscriptions", api.IsAdminHandler(APIUpdateSubscriptionInfo))

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

	if err := models.ApplyFilters(context, subscriptionCollectionModel.GetDBCollection()); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e7f9dbaf-15f0-4f9b-b56c-34a2111c7981", err.Error())
	}

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return subscriptionCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	if err := subscriptionCollectionModel.ListLimit(models.GetListLimit(context)); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "98ac65cb-9394-49bf-83c0-1fc4cbba0128", err.Error())
	}

	// extra parameter handle
	if err := models.ApplyExtraAttributes(context, subscriptionCollectionModel); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "b46343e6-b707-40fe-a842-680b98456aca", err.Error())
	}

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
	if err := dbCollection.AddStaticFilter("visitor_id", "=", visitorID); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "68d3bbd1-a93f-46d9-8fd2-bd675ebe26c9", err.Error())
	}
	if err := dbCollection.AddStaticFilter("status", "=", subscription.ConstSubscriptionStatusConfirmed); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "8ee26edd-21cf-4f1c-870b-004714a8899e", err.Error())
	}
	if err := models.ApplyFilters(context, dbCollection); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "8dc49237-b7f7-42e2-a03a-2df839679340", err.Error())
	}

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return dbCollection.Count()
	}

	// limit parameter handle
	if err := dbCollection.SetLimit(models.GetListLimit(context)); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "b7bcf8a3-340d-428a-b722-64d890f68863", err.Error())
	}

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
	isOwner := subscriptionInstance.GetVisitorID() == visitor.GetCurrentVisitorID(context)

	if !api.IsAdminSession(context) && !isOwner {
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
	if err := app.SendMail(email, subject, body); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "b9e13b60-c4b5-4934-8fc9-d0f520323b6b", err.Error())
	}
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
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "68461077-353d-4c50-8e62-9447f6a259a6", err.Error())
	}

	return subject, body
}

// APIUpdateSubscriptionInfo allows run and update info of all existing subscriptions
//  - if id is provided in request it is used to filter category
func APIUpdateSubscriptionInfo(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	subscriptionID := context.GetRequestArgument("id")

	subscriptionCollection, err := subscription.GetSubscriptionCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if subscriptionID != "" {
		if err := subscriptionCollection.ListFilterAdd("_id", "=", subscriptionID); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "a5987106-59f0-4210-9f33-1d20a633dfd5", err.Error())
		}
	}

	for _, currentSubscription := range subscriptionCollection.ListSubscriptions() {

		for _, subscriptionItem := range currentSubscription.GetItems() {
			productModel, err := product.LoadProductByID(subscriptionItem.ProductID)
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}
			if err = productModel.ApplyOptions(subscriptionItem.Options); err != nil {
				// no need to return here as it's possible that some options was already changed
				_ = env.ErrorDispatch(err)
				continue
			}
			productOptions := make(map[string]interface{})

			// add options to subscription info as description that used to show on FED
			for key, value := range productModel.GetOptions() {
				option := utils.InterfaceToMap(value)
				optionLabel := key
				if labelValue, optionLabelPresent := option["label"]; optionLabelPresent {
					optionLabel = utils.InterfaceToString(labelValue)
				}

				productOptions[optionLabel] = value
				optionValue, optionValuePresent := option["value"]
				// in this case looks like structure of options was changed or it's not a map
				if !optionValuePresent {
					continue
				}
				productOptions[optionLabel] = optionValue

				optionType := ""
				if val, present := option["type"]; present {
					optionType = utils.InterfaceToString(val)
				}
				if options, present := option["options"]; present {
					optionsMap := utils.InterfaceToMap(options)

					if optionType == "multi_select" {
						selectedOptions := ""
						for i, optionValue := range utils.InterfaceToArray(optionValue) {
							if optionValueParameters, ok := optionsMap[utils.InterfaceToString(optionValue)]; ok {
								optionValueParametersMap := utils.InterfaceToMap(optionValueParameters)
								if labelValue, labelValuePresent := optionValueParametersMap["label"]; labelValuePresent {
									productOptions[optionLabel] = labelValue
									if i > 0 {
										selectedOptions = selectedOptions + ", "
									}
									selectedOptions = selectedOptions + utils.InterfaceToString(labelValue)
								}
							}
						}
						productOptions[optionLabel] = selectedOptions

					} else if optionValueParameters, ok := optionsMap[utils.InterfaceToString(optionValue)]; ok {
						optionValueParametersMap := utils.InterfaceToMap(optionValueParameters)
						if labelValue, labelValuePresent := optionValueParametersMap["label"]; labelValuePresent {
							productOptions[optionLabel] = labelValue
						}

					}
				}
			}

			currentSubscription.SetInfo("options", productOptions)
		}

		err = currentSubscription.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	return "ok", nil
}
