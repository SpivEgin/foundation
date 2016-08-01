package composer

import (
	"fmt"
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/utils"
	"regexp"
	"strings"
)

// init makes package self-initialization routine
func init() {
	instance := new(DefaultComposer)
	instance.units = make(map[string]InterfaceComposeUnit)
	instance.types = make(map[string]InterfaceComposeType)

	registeredComposer = instance

	api.RegisterOnRestServiceStart(setupAPI)
	app.OnAppStart(initModelTypes)
	initBaseUnits()
	initTest()
	initBaseTypes()
}

// initBaseUnits register simple units
func initBaseUnits() {

	//	action := func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
	//		if argValue, present := args[""]; present {
	//			return utils.Equals(in, argValue), nil
	//		}
	//		return false, nil
	//	}
	//
	//	registeredComposer.RegisterUnit(&BasicUnit{
	//		Name: "*eq",
	//		Type: map[string]string{
	//			ConstPrefixUnit: ConstTypeAny, // input type
	//			ConstPrefixArg:  ConstTypeAny, // operand type (unnamed argument is a key for rule right-side value if it is not a map)
	//			ConstPrefixOut:  "boolean",       // output type
	//		},
	//		Label:       map[string]string{ConstPrefixUnit: "equals"},
	//		Description: map[string]string{ConstPrefixUnit: "Checks if value equals to other value"},
	//		Action:      action,
	//	})

	action := func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
		if argValue, present := args[""]; present {
			if utils.InterfaceToFloat64(in) > utils.InterfaceToFloat64(argValue) {
				return true, nil
			}
		}
		return false, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*gt",
		Type: map[string]string{
			ConstPrefixUnit: ConstTypeAny,
			ConstPrefixArg:  ConstTypeAny,
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "greater then"},
		Description: map[string]string{ConstPrefixUnit: "Checks if value is greater then other value"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
		if argValue, present := args[""]; present {
			if utils.InterfaceToFloat64(in) < utils.InterfaceToFloat64(argValue) {
				return true, nil
			}
		}
		return false, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*lt",
		Type: map[string]string{
			ConstPrefixUnit: ConstTypeAny,
			ConstPrefixArg:  ConstTypeAny,
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "less then"},
		Description: map[string]string{ConstPrefixUnit: "Checks if value if lower then other value"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
		if argValue, present := args[""]; present {
			if strings.Contains(utils.InterfaceToString(in), utils.InterfaceToString(argValue)) {
				return true, nil
			}
		}
		return false, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*contains",
		Type: map[string]string{
			ConstPrefixUnit: "string",
			ConstPrefixArg:  "string",
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "contains"},
		Description: map[string]string{ConstPrefixUnit: "Checks if value containt other value"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
		if argValue, present := args[""]; present {
			if matched, err := regexp.MatchString(utils.InterfaceToString(argValue), utils.InterfaceToString(in)); err == nil {
				return matched, nil
			}
		}
		return false, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*regex",
		Type: map[string]string{
			ConstPrefixUnit: "string",
			ConstPrefixArg:  "string",
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "regex"},
		Description: map[string]string{ConstPrefixUnit: "Checks regular expression over value"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {
		// in should be list of checks
		for _, check := range utils.InterfaceToArray(in) {
			if result, err := composer.Check(check, args); result && err == nil {
				return true, nil
			}
		}

		return false, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*any",
		Type: map[string]string{
			ConstPrefixUnit: "array", // apply check to each element of list
			ConstPrefixArg:  ConstTypeAny,
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "any"},
		Description: map[string]string{ConstPrefixUnit: "Checks list of condiions and reuturn first positive"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		// in should be list of checks
		for _, check := range utils.InterfaceToArray(in) {
			if result, err := composer.Check(check, args); !result || err != nil {
				return false, err
			}
		}

		return true, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*all",
		Type: map[string]string{
			ConstPrefixUnit: "array",      // apply check to each element of list
			ConstPrefixArg:  ConstTypeAny, // rules should be passed as an arguments that are checked per item
			ConstPrefixOut:  "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "all"},
		Description: map[string]string{ConstPrefixUnit: "Checks list of condiions and reuturn false and first negative"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		// in should be list of checks
		i := 0

		for _, check := range utils.InterfaceToArray(in) {
			if result, err := composer.Check(check, args); result && err == nil {
				i++
			}
		}
		fmt.Println(args, i)

		return i, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*have",
		Type: map[string]string{
			ConstPrefixUnit:              "CartItems", // apply check to each element of list
			ConstPrefixArg + "productID": "string",
			ConstPrefixArg + "sku":       "string",
			ConstPrefixArg + "qty":       utils.ConstDataTypeInteger,
			ConstPrefixArg + "options":   "object",
			ConstPrefixOut:               utils.ConstDataTypeInteger,
		},
		Label:       map[string]string{ConstPrefixUnit: "allItemsCheck"},
		Description: map[string]string{ConstPrefixUnit: "Checks list of condiions for items and reuturn false on first negative"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		// in should be list of checks
		result := make(map[string]interface{})
		//		groupedProducts := make(map[string]int)
		//
		//		for key, value := range utils.InterfaceToMap(in) {
		//			if key == "items" {
		//				for _, productInCart := range utils.InterfaceToArray(value) {
		//					productID := productInCart.GetProductID()
		//					productQty := productInCart.GetQty()
		//
		//					if qty, present := productsInCart[productID]; present {
		//						productsInCart[productID] = qty + productQty
		//						continue
		//					}
		//					productsInCart[productID] = productQty
		//					applicableProductDiscounts[productID] = make([]discount, 0)
		//				}
		//			}
		//			result[key] = value
		//		}

		return result, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*group",
		Type: map[string]string{
			ConstPrefixUnit:              "Cart",
			ConstPrefixArg + "productID": "boolean",
			ConstPrefixOut:               "FlatCart",
		},
		Label:       map[string]string{ConstPrefixUnit: "allItemsCheck"},
		Description: map[string]string{ConstPrefixUnit: "Checks list of condiions for items and reuturn false on first negative"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		// in should be list of checks
		result := make(map[string]interface{})
		fmt.Println(in, args)

		return result, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*global",
		Type: map[string]string{
			ConstPrefixUnit:                   "Visitor",
			ConstPrefixArg + "startDate":      utils.ConstDataTypeDatetime,
			ConstPrefixArg + "endDate":        utils.ConstDataTypeDatetime,
			ConstPrefixArg + "customerGroups": "[]" + utils.ConstDataTypeText,
			ConstPrefixArg + "sourceCodes":    "[]" + utils.ConstDataTypeText,
			ConstPrefixArg + "couponCodes":    "[]" + utils.ConstDataTypeText,
		},
		Label:       map[string]string{ConstPrefixUnit: "Global"},
		Description: map[string]string{ConstPrefixUnit: "Global set of rules :"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		return "ok", nil
	}
	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*rule",
		Type: map[string]string{
			ConstPrefixOut:               "",
			ConstPrefixUnit:              "CartItem",
			ConstPrefixArg + "productID": "string",
			ConstPrefixArg + "exclusive": "boolean",
			ConstPrefixArg + "sku":       "string",
			ConstPrefixArg + "qty":       utils.ConstDataTypeInteger,
			ConstPrefixArg + "options":   "object",
		},
		Label:       map[string]string{ConstPrefixUnit: ""},
		Description: map[string]string{ConstPrefixUnit: "Rule to apply on cart items", ConstPrefixArg + "exclusive": "For all items, or first possible"},
		Action:      action,
	})

}

func initTest() error {

	CartItemType := &BasicType{
		Name: "CartItem",
		Label: map[string]string{
			"":          "CartItem",
			"productID": "ProductID",
			"sku":       "Sku",
			"qty":       "Qty",
			"options":   "Options",
		},
		Type: map[string]string{
			"":          "CartItem",
			"productID": "string",
			"sku":       "string",
			"qty":       utils.ConstDataTypeInteger,
			"options":   "object",
		},
		Description: map[string]string{
			"":          "Cart Item",
			"productID": "ProductID",
			"sku":       "Sku",
			"qty":       "Qty",
			"options":   "Options",
		},
	}
	registeredComposer.RegisterType(CartItemType)

	// this type is temp usage to work with cart items and apply rules to each of them
	CartItemsType := &BasicType{
		Name: "CartItems",
		Label: map[string]string{
			"":          "CartItems",
			"productID": "ProductID",
			"sku":       "Sku",
			"qty":       "Qty",
			"options":   "Options",
		},
		Type: map[string]string{
			"":    "CartItems",
			"qty": "Qty",
		},
		Description: map[string]string{
			"":          "Cart Item",
			"productID": "ProductID",
			"sku":       "Sku",
			"qty":       "Qty of cart items",
			"options":   "Options",
		},
	}
	registeredComposer.RegisterType(CartItemsType)

	CartType := &BasicType{
		Name: "Cart",
		Label: map[string]string{
			"":          "Cart",
			"subtotal":  "Amount",
			"cartItems": "List of Cart Items",
			"items":     "List of Cart Items",
		},
		Type: map[string]string{
			"":          "Cart",
			"subtotal":  "float",
			"cartItems": "[]CartItem",
			"items":     "CartItems",
		},
		Description: map[string]string{
			"":          "Cart model object",
			"subtotal":  "Cart amount",
			"cartItems": "array of CartItems",
		},
	}

	registeredComposer.RegisterType(CartType)

	PaymentType := &BasicType{
		Name: "Payment",
		Label: map[string]string{
			"":     "Payment",
			"name": "Name",
			"code": "Code",
			"type": "Type",
		},
		Type: map[string]string{
			"":     "Payment",
			"name": "string",
			"code": "string",
			"type": "string",
		},
		Description: map[string]string{
			"":     "Payment",
			"name": "Name",
			"code": "Code",
			"type": "Type",
		},
	}

	registeredComposer.RegisterType(PaymentType)

	ShippingType := &BasicType{
		Name: "Shipping",
		Label: map[string]string{
			"":     "Shipping",
			"name": "Name",
			"code": "Code",
			"type": "Type",
		},
		Type: map[string]string{
			"":     "Shipping",
			"name": "string",
			"code": "string",
			"type": "string",
		},
		Description: map[string]string{
			"":     "Shipping",
			"name": "Name",
			"code": "Code",
			"type": "Type",
		},
	}

	registeredComposer.RegisterType(ShippingType)

	// this type is not response to existing one at 100%
	CheckoutType := &BasicType{
		Name: "Checkout",
		Label: map[string]string{
			"cart":           "Cart",
			"paymentMethod":  "Payment Method",
			"shippingMethod": "Shipping Method",
			"email":          "string",
			"subtotal":       "float",
			"shipping":       "float",
			"discount":       "float",
			"grandtotal":     "float",
		},
		Type: map[string]string{
			"cart":           "Cart",
			"paymentMethod":  "Payment",
			"shippingMethod": "Shipping",
			"email":          "string",
			"subtotal":       "float",
			"shipping":       "float",
			"discount":       "float",
			"grandtotal":     "float",
		},
		Description: map[string]string{
			"cart":           "Cart",
			"paymentMethod":  "Payment",
			"shippingMethod": "Shipping",
			"email":          "string",
			"subtotal":       "float",
			"shipping":       "float",
			"discount":       "float",
			"grandtotal":     "float",
		},
	}

	registeredComposer.RegisterType(CheckoutType)

	testVisitorType := &BasicType{
		Name: "Visitor",
		Label: map[string]string{
			"id":             "ID",
			"name":           "Name",
			"country":        "Country",
			"visitorIsLogin": "Visitor is login",
		},
		Type: map[string]string{
			"id":             "string",
			"name":           "string",
			"country":        "string",
			"visitorIsLogin": "boolean",
		},
		Description: map[string]string{},
	}

	registeredComposer.RegisterType(testVisitorType)

	testDiscountRule := &BasicType{
		Name: "DiscountRule",
		Label: map[string]string{
			"":         "Discount Rule",
			"Cart":     "Cart",
			"Visitor":  "Visitor",
			"Checkout": "Checkout",
		},
		Type: map[string]string{
			"":         "DiscountRule",
			"Cart":     "Cart",
			"Visitor":  "Visitor",
			"Checkout": "Checkout",
		},
		Description: map[string]string{
			"":         "DiscountRule model object",
			"Cart":     "cart description",
			"Visitor":  "visitor description",
			"Checkout": "checkout description",
		},
	}

	registeredComposer.RegisterType(testDiscountRule)

	testDiscountAction := &BasicType{
		Name: "DiscountAction",
		Label: map[string]string{
			"":           "DiscountAction",
			"name":       "Name",
			"code":       "Code",
			"amount":     "Discount amount",
			"is_percent": "Is percent",
			"priority":   "Priority",
		},
		Type: map[string]string{
			"":           "DiscountAction",
			"name":       "string",
			"code":       "string",
			"amount":     "float",
			"is_percent": "boolean",
			"priority":   "float",
		},
		Description: map[string]string{},
	}

	registeredComposer.RegisterType(testDiscountAction)
	return nil
}

// initModelTypes register base types into composer
func initBaseTypes() error {

	for goType, jsonType := range map[string]string{
		utils.ConstDataTypeID:      "string",
		utils.ConstDataTypeBoolean: "boolean",
		utils.ConstDataTypeVarchar: "string",
		utils.ConstDataTypeText:    "string",
		//utils.ConstDataTypeInteger:  "int",
		utils.ConstDataTypeDecimal: "float",
		utils.ConstDataTypeMoney:   "float",
		//utils.ConstDataTypeFloat:    "float",
		utils.ConstDataTypeDatetime: "string",
		utils.ConstDataTypeJSON:     "object",
	} {

		registeredComposer.RegisterType(&BasicType{
			Name:        goType,
			Label:       map[string]string{goType: strings.Title(goType)},
			Type:        map[string]string{goType: jsonType},
			Description: map[string]string{goType: "Basic Ottemo type {" + goType + "}"},
		})

	}

	return nil
}

// initModelTypes register all foundation models that implements interface object with their type including
//  all attributes provided by GetAttributesInfo
func initModelTypes() error {

	for modelName, modelInstance := range models.GetDeclaredModels() {
		if modelInstance == nil {
			continue
		}

		modelInstance, err := modelInstance.New()
		if err != nil || modelInstance == nil {
			continue
		}

		if objectInstance, ok := modelInstance.(models.InterfaceObject); ok {
			baseType := &BasicType{
				Name:        modelName,
				Label:       map[string]string{"": modelName},
				Type:        map[string]string{"": modelName},
				Description: map[string]string{"": modelName + " model object"},
			}

			for _, v := range objectInstance.GetAttributesInfo() {
				baseType.Label[v.Attribute] = v.Label
				baseType.Type[v.Attribute] = v.Type
				baseType.Description[v.Attribute] = "The '" + v.Label + "' attribute"
			}

			registeredComposer.RegisterType(baseType)
		}
	}

	return nil
}
