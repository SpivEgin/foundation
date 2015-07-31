package sales

import (
	"bytes"
	"runtime"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("promotions/list", api.ConstRESTOperationGet, getSessionTimeZone)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}