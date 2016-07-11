package testDiscount

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()
	service.PUT("testDiscount/CreateTestRule", APICreateTestRule)
	service.PUT("testDiscount/CreateTestAction", APICreateTestAction)
	service.GET("testDiscount/GetConfigForTestDiscount", APIGetConfigForTestDiscount)

	return nil
}

func APICreateTestRule(context api.InterfaceApplicationContext) (interface{}, error) {
	config := env.GetConfig()

	var setValue interface{}

	setValue = context.GetRequestContent()
	configPath := ConstConfigPathTestDiscountRule

	err := config.SetValue(configPath, setValue)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "rule was saved successfully", nil
}

func APICreateTestAction(context api.InterfaceApplicationContext) (interface{}, error) {
	config := env.GetConfig()

	var setValue interface{}

	setValue = context.GetRequestContent()
	configPath := ConstConfigPathTestDiscountAction

	err := config.SetValue(configPath, setValue)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "action was saved successfully", nil
}

func APIGetConfigForTestDiscount(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]interface{})

	config := env.GetConfig()

	rule := make(map[string]interface{})
	rule["type"] = "DiscountRule"
	rule["json"] = config.GetValue(ConstConfigPathTestDiscountRule)
	result["rule"] = rule

	action := make(map[string]interface{})
	action["type"] = "DiscountAction"
	action["json"] = config.GetValue(ConstConfigPathTestDiscountAction)
	result["action"] = action

	return result, nil
}
