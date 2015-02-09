package checkout

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"

	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("checkout", api.ConstRESTOperationGet, restCheckoutInfo)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/payment/methods", api.ConstRESTOperationGet, restCheckoutPaymentMethods)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/shipping/methods", api.ConstRESTOperationGet, restCheckoutShippingMethods)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/info", api.ConstRESTOperationUpdate, restCheckoutSetInfo)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/shipping/address", api.ConstRESTOperationCreate, restCheckoutSetShippingAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/billing/address", api.ConstRESTOperationCreate, restCheckoutSetBillingAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/payment/method/:method", api.ConstRESTOperationCreate, restCheckoutSetPaymentMethod)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/shipping/method/:method/:rate", api.ConstRESTOperationCreate, restCheckoutSetShippingMethod)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("checkout/submit", api.ConstRESTOperationGet, restSubmit)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// WEB REST API function to get current checkout process status
func restCheckoutInfo(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := map[string]interface{}{
		"billing_address":  nil,
		"shipping_address": nil,

		"payment_method_name": nil,
		"payment_method_code": nil,

		"shipping_method_name": nil,
		"shipping_method_code": nil,

		"shipping_rate":   nil,
		"shipping_amount": nil,

		"discounts":       nil,
		"discount_amount": nil,

		"taxes":      nil,
		"tax_amount": nil,

		"subtotal":   nil,
		"grandtotal": nil,
		"info":       nil,
	}

	if billingAddress := currentCheckout.GetBillingAddress(); billingAddress != nil {
		result["billing_address"] = billingAddress.ToHashMap()
	}

	if shippingAddress := currentCheckout.GetShippingAddress(); shippingAddress != nil {
		result["shipping_address"] = shippingAddress.ToHashMap()
	}

	if paymentMethod := currentCheckout.GetPaymentMethod(); paymentMethod != nil {
		result["payment_method_name"] = paymentMethod.GetName()
		result["payment_method_code"] = paymentMethod.GetCode()
	}

	if shippingMethod := currentCheckout.GetShippingMethod(); shippingMethod != nil {
		result["shipping_method_name"] = shippingMethod.GetName()
		result["shipping_method_code"] = shippingMethod.GetCode()
	}

	if shippingRate := currentCheckout.GetShippingRate(); shippingRate != nil {
		result["shipping_rate"] = shippingRate
		result["shipping_amount"] = shippingRate.Price
	}

	if checkoutCart := currentCheckout.GetCart(); checkoutCart != nil {
		result["subtotal"] = checkoutCart.GetSubtotal()
	}

	result["discount_amount"], result["discounts"] = currentCheckout.GetDiscounts()

	result["tax_amount"], result["taxes"] = currentCheckout.GetTaxes()

	result["grandtotal"] = currentCheckout.GetGrandTotal()
	result["info"] = currentCheckout.GetInfo("*")

	return result, nil
}

// WEB REST API function to get possible payment methods for checkout
func restCheckoutPaymentMethods(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	type ResultValue struct {
		Name string
		Code string
		Type string
	}
	var result []ResultValue

	for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
		if paymentMethod.IsAllowed(currentCheckout) {
			result = append(result, ResultValue{Name: paymentMethod.GetName(), Code: paymentMethod.GetCode(), Type: paymentMethod.GetType()})
		}
	}

	return result, nil
}

// WEB REST API function to get possible shipping methods for checkout
func restCheckoutShippingMethods(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	type ResultValue struct {
		Name  string
		Code  string
		Rates []checkout.StructShippingRate
	}
	var result []ResultValue

	for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
		if shippingMethod.IsAllowed(currentCheckout) {
			result = append(result, ResultValue{Name: shippingMethod.GetName(), Code: shippingMethod.GetCode(), Rates: shippingMethod.GetRates(currentCheckout)})
		}
	}

	return result, nil
}

// WEB REST API function to set checkout related extra information
func restCheckoutSetInfo(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	reqData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for key, value := range reqData {
		err := currentCheckout.SetInfo(key, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	// updating session
	checkout.SetCurrentCheckout(context, currentCheckout)

	return "ok", nil
}

// internal function for  restCheckoutSetShippingAddress() and restCheckoutSetBillingAddress()
func checkoutObtainAddress(context api.InterfaceApplicationContext) (visitor.InterfaceVisitorAddress, error) {

	reqData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if addressID, present := reqData["id"]; present {

		// Address id was specified - trying to load
		visitorAddress, err := visitor.LoadVisitorAddressByID(utils.InterfaceToString(addressID))
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		currentVisitorID := utils.InterfaceToString(context.GetSession().Get(visitor.ConstSessionKeyVisitorID))
		if visitorAddress.GetVisitorID() != currentVisitorID {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bef27714-4ac5-4705-b59a-47c8e0bc5aa4", "address id is not related to current visitor")
		}

		return visitorAddress, nil
	}

	// supposedly address data was specified
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range reqData {
		err := visitorAddressModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	visitorID := utils.InterfaceToString(context.GetSession().Get(visitor.ConstSessionKeyVisitorID))
	visitorAddressModel.Set("visitor_id", visitorID)

	if visitorAddressModel.GetID() != "" {
		err = visitorAddressModel.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	return visitorAddressModel, nil
}

// WEB REST API function to set shipping address
func restCheckoutSetShippingAddress(context api.InterfaceApplicationContext) (interface{}, error) {
	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	address, err := checkoutObtainAddress(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCheckout.SetShippingAddress(address)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// updating session
	checkout.SetCurrentCheckout(context, currentCheckout)

	return address.ToHashMap(), nil
}

// WEB REST API function to set billing address
func restCheckoutSetBillingAddress(context api.InterfaceApplicationContext) (interface{}, error) {
	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	address, err := checkoutObtainAddress(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = currentCheckout.SetBillingAddress(address)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// updating session
	checkout.SetCurrentCheckout(context, currentCheckout)

	return address.ToHashMap(), nil
}

// WEB REST API function to set payment method
func restCheckoutSetPaymentMethod(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// looking for payment method
	for _, paymentMethod := range checkout.GetRegisteredPaymentMethods() {
		if paymentMethod.GetCode() == context.GetRequestArgument("method") {
			if paymentMethod.IsAllowed(currentCheckout) {

				// updating checkout payment method
				err := currentCheckout.SetPaymentMethod(paymentMethod)
				if err != nil {
					return nil, env.ErrorDispatch(err)
				}

				// checking for additional info
				contentValues, _ := api.GetRequestContentAsMap(context)
				for key, value := range contentValues {
					currentCheckout.SetInfo(key, value)
				}

				eventData := map[string]interface{}{"session": context.GetSession(), "paymentMethod": paymentMethod, "checkout": currentCheckout}
				env.Event("api.checkout.setPayment", eventData)

				// updating session
				checkout.SetCurrentCheckout(context, currentCheckout)

				return "ok", nil
			}
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "bd07849e-8789-4316-924c-9c754efbc348", "payment method not allowed")
		}
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b8384a47-8806-4a54-90fc-cccb5e958b4e", "payment method not found")
}

// WEB REST API function to set payment method
func restCheckoutSetShippingMethod(context api.InterfaceApplicationContext) (interface{}, error) {
	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// looking for shipping method
	for _, shippingMethod := range checkout.GetRegisteredShippingMethods() {
		if shippingMethod.GetCode() == context.GetRequestArgument("method") {
			if shippingMethod.IsAllowed(currentCheckout) {

				// looking for shipping rate
				for _, shippingRate := range shippingMethod.GetRates(currentCheckout) {
					if shippingRate.Code == context.GetRequestArgument("rate") {

						err := currentCheckout.SetShippingMethod(shippingMethod)
						if err != nil {
							return nil, env.ErrorDispatch(err)
						}

						err = currentCheckout.SetShippingRate(shippingRate)
						if err != nil {
							return nil, env.ErrorDispatch(err)
						}

						// updating session
						checkout.SetCurrentCheckout(context, currentCheckout)

						return "ok", nil
					}
				}

			} else {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d7fb6ff2-b914-467b-bf56-b8d2bea472ef", "shipping method not allowed")
			}
		}
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "279a645c-6a03-44de-95c0-2651a51440fa", "shipping method and/or rate were not found")
}

// WEB REST API function to submit checkout information and make order
func restSubmit(context api.InterfaceApplicationContext) (interface{}, error) {

	currentCheckout, err := checkout.GetCurrentCheckout(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return currentCheckout.Submit()
}
