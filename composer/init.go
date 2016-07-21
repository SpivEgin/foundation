package composer

import (
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
		for _, check := range utils.InterfaceToArray(in) {
			if result, err := composer.Check(check, args); !result || err != nil {
				return false, err
			}
		}

		return true, nil
	}

	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*allCartItems",
		Type: map[string]string{
			ConstPrefixUnit: "Cart", // apply check to each element of list
			//ConstPrefixArg:  ConstTypeAny, // rules should be passed as an arguments that are checked per item

			ConstPrefixArg + "amount":         "float",
			ConstPrefixArg + "visitorIsLogin": "bool",
			ConstPrefixArg + "cartItems":      "object",
			ConstPrefixOut:                    "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: "allItemsCheck", ConstPrefixArg: "Rules"},
		Description: map[string]string{ConstPrefixUnit: "Checks list of condiions for items and reuturn false on first negative"},
		Action:      action,
	})

	action = func(in interface{}, args map[string]interface{}, composer InterfaceComposer) (interface{}, error) {

		return "ok", nil
	}
	registeredComposer.RegisterUnit(&BasicUnit{
		Name: "*test",
		Type: map[string]string{
			ConstPrefixOut:  "",
			ConstPrefixUnit: "Cart",
			//ConstPrefixArg:  "object",
			ConstPrefixArg + "amount":         "float",
			ConstPrefixArg + "visitorIsLogin": "bool",
		},
		Label:       map[string]string{ConstPrefixUnit: ""},
		Description: map[string]string{ConstPrefixUnit: "Temporary test unit"},
		Action:      action,
	})

}

func initTest() error {
	testType := &BasicType{
		Name: "Test",
		Label: map[string]string{
			"a": "IntegerTest",
			"b": "FloatTest",
			"c": "StringTest",
			"d": "ProductTest",
		},
		Type: map[string]string{
			"a": "any",
			"b": "float",
			"c": "string",
			"d": "Checkout",
			"e": "[]Product",
		},
		Description: map[string]string{
			"a": "Description for Test type",
		},
	}

	// can we represent using this type array of any elements?
	registeredComposer.RegisterType(testType)
	arrayType := &BasicType{
		Name: "object",
		Label: map[string]string{
			"": "object",
		},
		Type: map[string]string{
			"": "object",
		},
		Description: map[string]string{
			"": "object",
		},
	}

	registeredComposer.RegisterType(arrayType)

	testCartItemType := &BasicType{
		Name: "Items",
		Label: map[string]string{
			"":          "Items",
			"ProductID": "ProductID",
			"Qty":       "Qty",
			"Options":   "Options",
		},
		Type: map[string]string{
			"":          "CartItem",
			"ProductID": "string",
			"Qty":       utils.ConstDataTypeInteger,
			"Options":   "object",
		},
		Description: map[string]string{
			"":          "Cart Item object",
			"ProductID": "ProductID",
			"Qty":       "Qty",
			"Options":   "Options",
		},
	}

	/*
		idx       int
		ProductID string
		Qty       int
		Options   map[string]interface{}
	*/

	registeredComposer.RegisterType(testCartItemType)

	testCartType := &BasicType{
		Name: "Cart",
		Label: map[string]string{
			"":               "Cart",
			"subtotal":       "Amount",
			"visitorIsLogin": "Visitor is login",
			"cartItems":      "Items",
		},
		Type: map[string]string{
			"":               "Cart",
			"subtotal":       "float",
			"visitorIsLogin": "boolean",
			"cartItems":      "Items",
		},
		Description: map[string]string{
			"":          "Cart model object",
			"subtotal":  "Cart amount",
			"cartItems": "array of CartItems",
		},
	}

	registeredComposer.RegisterType(testCartType)

	testCheckoutType := &BasicType{
		Name: "Checkout",
		Label: map[string]string{
			"cart":            "Cart",
			"paymentMethods":  "Payment Methods",
			"shippingMethods": "Shippin Methods",
		},
		Type: map[string]string{
			"cart":            "Cart",
			"paymentMethods":  "[]Payment",
			"shippingMethods": "[]Shippin",
		},
		Description: map[string]string{
			"cart": "current Cart",
		},
	}

	registeredComposer.RegisterType(testCheckoutType)

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
