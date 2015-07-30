package order

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/env"

	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("orders/attributes", api.ConstRESTOperationGet, APIListOrderAttributes)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("orders", api.ConstRESTOperationGet, APIListOrders)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("order/:orderID", api.ConstRESTOperationGet, APIGetOrder)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	// err = api.GetRestService().RegisterAPI("order", api.ConstRESTOperationCreate, APICreateOrder)
	// if err != nil {
	// 	return env.ErrorDispatch(err)
	// }

	err = api.GetRestService().RegisterAPI("order/:orderID", api.ConstRESTOperationUpdate, APIUpdateOrder)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("order/:orderID", api.ConstRESTOperationDelete, APIDeleteOrder)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// APIListOrderAttributes returns a list of purchase order attributes
func APIListOrderAttributes(context api.InterfaceApplicationContext) (interface{}, error) {

	orderModel, err := order.GetOrderModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return orderModel.GetAttributesInfo(), nil
}

// APIListOrders returns a list of existing purchase orders
//   - if "action" parameter is set to "count" result value will be just a number of list items
func APIListOrders(context api.InterfaceApplicationContext) (interface{}, error) {

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// taking orders collection model
	orderCollectionModel, err := order.GetOrderCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// filters handle
	models.ApplyFilters(context, orderCollectionModel.GetDBCollection())

	// checking for a "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return orderCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	orderCollectionModel.ListLimit(models.GetListLimit(context))

	// extra parameter handle
	models.ApplyExtraAttributes(context, orderCollectionModel)

	return orderCollectionModel.List()
}

// APIGetOrder return specified purchase order information
//   - order id should be specified in "orderID" argument
func APIGetOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	blockID := context.GetRequestArgument("orderID")
	if blockID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "723ef443-f974-4455-9be0-a8af13916554", "order id should be specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// operation
	//----------
	orderModel, err := order.LoadOrderByID(blockID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := orderModel.ToHashMap()
	if notes, present := utils.InterfaceToMap(result["shipping_info"])["notes"]; present {
		utils.InterfaceToMap(result["shipping_address"])["notes"] = notes
	}

	result["items"] = orderModel.GetItems()
	return result, nil
}

// APIUpdateOrder update existing purchase order
//   - order id should be specified in "orderID" argument
func APIUpdateOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	blockID := context.GetRequestArgument("orderID")
	if blockID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "20a08638-e9e6-428b-b70c-a418d7821e4b", "order id should be specified")
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// operation
	//----------
	orderModel, err := order.LoadOrderByID(blockID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range requestData {
		orderModel.Set(attribute, value)
	}

	orderModel.SetID(blockID)
	orderModel.Save()

	return orderModel.ToHashMap(), nil
}

// APIDeleteOrder deletes existing purchase order
//   - order id should be specified in "orderID" argument
func APIDeleteOrder(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	blockID := context.GetRequestArgument("orderID")
	if blockID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fc3011c7-e58c-4433-b9b0-881a7ba005cf", "order id should be specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// operation
	//----------
	orderModel, err := order.GetOrderModelAndSetID(blockID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	orderModel.Delete()

	return "ok", nil
}

// APIConfirmOrder return specified purchase order information for duplicate
//   - order id should be specified in "orderID" argument
func APIConfirmOrder(context api.InterfaceApplicationContext) (interface{}, error) {

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
