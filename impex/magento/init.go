package magento

import (
	"fmt"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/app/models"
)

// init makes package self-initialization routine
func init() {
	api.RegisterOnRestServiceStart(setupAPI)

	// initializing column conversion functions
	ConversionFuncs["log"] = func(args ...interface{}) string {
		env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprint(args))
		return ""
	}

	for code, name := range models.ConstStatesList {
		statesList[name] = code
	}

}
