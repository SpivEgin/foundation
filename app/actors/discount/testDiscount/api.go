package testDiscount

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("testDiscount/CreateTestDiscount", api.ConstRESTOperationCreate, CreateTestDiscount)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

func CreateTestDiscount(context api.InterfaceApplicationContext) (interface{}, error) {
	//CalculateDiscount

	return "", nil
}

