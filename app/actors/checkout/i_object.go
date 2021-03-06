package checkout

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// Get returns object attribute value or nil
func (it *DefaultCheckout) Get(attribute string) interface{} {

	switch attribute {
	case "CartID":
		return it.CartID
	case "VisitorID":
		return it.VisitorID
	case "OrderID":
		return it.OrderID
	case "SessionID":
		return it.SessionID
	case "ShippingAddress":
		return it.ShippingAddress
	case "BillingAddress":
		return it.BillingAddress
	case "PaymentMethodCode":
		return it.PaymentMethodCode
	case "ShippingMethodCode":
		return it.ShippingMethodCode
	case "ShippingRate":
		return it.ShippingRate
	case "Info":
		return it.Info
	}

	return nil
}

// Set sets attribute value to object or returns error
func (it *DefaultCheckout) Set(attribute string, value interface{}) error {
	switch attribute {
	case "CartID":
		it.CartID = utils.InterfaceToString(value)

	case "VisitorID":
		it.VisitorID = utils.InterfaceToString(value)

	case "OrderID":
		it.OrderID = utils.InterfaceToString(value)

	case "SessionID":
		it.SessionID = utils.InterfaceToString(value)

	case "ShippingAddress":
		shippingAddress := utils.InterfaceToMap(value)
		if len(shippingAddress) > 0 {
			it.ShippingAddress = shippingAddress
		}

	case "BillingAddress":
		billingAddress := utils.InterfaceToMap(value)
		if len(billingAddress) > 0 {
			it.BillingAddress = billingAddress
		}

	case "PaymentMethodCode":
		paymentMethodCode := utils.InterfaceToString(value)
		for _, method := range checkout.GetRegisteredPaymentMethods() {
			if method.GetCode() == paymentMethodCode {
				it.PaymentMethodCode = paymentMethodCode
				break
			}
		}
	case "ShippingMethodCode":
		shippingMethodCode := utils.InterfaceToString(value)
		for _, method := range checkout.GetRegisteredShippingMethods() {
			if method.GetCode() == shippingMethodCode {
				it.ShippingMethodCode = shippingMethodCode
				break
			}
		}

	case "ShippingRate":
		mapValue := utils.InterfaceToMap(value)
		if utils.StrKeysInMap(mapValue, "Name", "Code", "Price") {
			it.ShippingRate.Name = utils.InterfaceToString(mapValue["Name"])
			it.ShippingRate.Code = utils.InterfaceToString(mapValue["Code"])
			it.ShippingRate.Price = utils.InterfaceToFloat64(mapValue["Price"])
		}

		// leave this on it's one place to prevent some checkout from drop
	case "Taxes":

	case "Discounts":

	case "Info":
		info := utils.InterfaceToMap(value)
		if len(info) > 0 {
			it.Info = info
		}

	default:
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6280473d-bea1-46ee-90af-af70aed317f0", "unknown checkout attribute - "+attribute)
	}
	return nil
}

// FromHashMap fills object attributes from map[string]interface{}
func (it *DefaultCheckout) FromHashMap(input map[string]interface{}) error {

	for attribute, value := range input {
		if err := it.Set(attribute, value); err != nil {
			return env.ErrorDispatch(err)
		}
	}

	return nil
}

// ToHashMap represents object as map[string]interface{}
func (it *DefaultCheckout) ToHashMap() map[string]interface{} {

	result := make(map[string]interface{})

	result["CartID"] = it.CartID
	result["VisitorID"] = it.VisitorID
	result["OrderID"] = it.OrderID
	result["SessionID"] = it.SessionID
	result["ShippingAddress"] = it.ShippingAddress
	result["BillingAddress"] = it.BillingAddress
	result["PaymentMethodCode"] = it.PaymentMethodCode
	result["ShippingMethodCode"] = it.ShippingMethodCode
	result["ShippingRate"] = it.ShippingRate
	result["Info"] = it.Info

	return result
}

// GetAttributesInfo returns information about object attributes
func (it *DefaultCheckout) GetAttributesInfo() []models.StructAttributeInfo {

	info := []models.StructAttributeInfo{
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "CartID",
			Type:       db.ConstTypeID,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Cart ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "VisitorID",
			Type:       db.ConstTypeID,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Visitor ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "SessionID",
			Type:       db.ConstTypeID,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Session ID",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingAddress",
			Type:       db.ConstTypeJSON,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Shipping Address",
			Group:      "General",
			Editors:    "model",
			Options:    "model:VisitorAddress",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "BillingAddress",
			Type:       db.ConstTypeJSON,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Billing Address",
			Group:      "General",
			Editors:    "model",
			Options:    "model:VisitorAddress",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "PaymentMethodCode",
			Type:       db.ConstTypeVarchar,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Payment Method",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingMethodCode",
			Type:       db.ConstTypeVarchar,
			IsRequired: true,
			IsStatic:   true,
			Label:      "Shipping Method",
			Group:      "General",
			Editors:    "line_text",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "ShippingRate",
			Type:       db.ConstTypeJSON,
			IsRequired: false,
			IsStatic:   true,
			Label:      "ShippingRate",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Taxes",
			Type:       db.ConstTypeJSON,
			IsRequired: false,
			IsStatic:   true,
			Label:      "Taxes",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Discounts",
			Type:       db.ConstTypeJSON,
			IsRequired: false,
			IsStatic:   true,
			Label:      "Discounts",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
		models.StructAttributeInfo{
			Model:      checkout.ConstCheckoutModelName,
			Collection: "",
			Attribute:  "Info",
			Type:       db.ConstTypeJSON,
			IsRequired: false,
			IsStatic:   true,
			Label:      "Info",
			Group:      "General",
			Editors:    "not_editable",
			Options:    "",
			Default:    "",
		},
	}

	return info
}
