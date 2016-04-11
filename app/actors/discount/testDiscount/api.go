package testDiscount

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("testDiscount/CreateTestRule", api.ConstRESTOperationCreate, CreateTestRule)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("testDiscount/CreateTestAction", api.ConstRESTOperationCreate, CreateTestAction)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

func CreateTestRule(context api.InterfaceApplicationContext) (interface{}, error) {
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

func CreateTestAction(context api.InterfaceApplicationContext) (interface{}, error) {
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

