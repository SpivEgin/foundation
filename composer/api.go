package composer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/composer"
)

// setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("composer/units", api.ConstRESTOperationGet, composerUnits)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// WEB REST API used to list available models for Impex system
func composerUnits(context api.InterfaceApplicationContext) (interface{}, error) {
	var result []string

	for _, unit := range composer.ListUnits() {
		result = append(result, unit.GetName())
	}

	return result, nil
}