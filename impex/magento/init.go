package magento

import (
	"fmt"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// init makes package self-initialization routine
func init() {
	api.RegisterOnRestServiceStart(setupAPI)

	// initializing column conversion functions
	ConversionFuncs["log"] = func(args ...interface{}) string {
		env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprint(args))
		return ""
	}

}
