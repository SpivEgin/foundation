package testDiscount

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("testDiscount/setRule", api.ConstRESTOperationCreate, setRule)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("testDiscount/setAction", api.ConstRESTOperationCreate, setAction)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

func setRule(context api.InterfaceApplicationContext) (interface{}, error) {

	result := ""

	return result, nil
}

// APIGetGiftCard return gift card info buy it's code
func setAction(context api.InterfaceApplicationContext) (interface{}, error) {

	result := ""

	return result, nil
}


