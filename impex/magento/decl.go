// Package impex is a implementation of import/export service
package magento

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstErrorModule = "impex_magento"
	ConstErrorLevel  = env.ConstErrorLevelAPI

	ConstLogFileName                     = "impex_magento.log"
	ConstSessionKeyMagentoRequestToken   = "magentoRequestToken"
	ConstSessionKeyMagentoRequestSecret  = "magentoRequestSecret"
	ConstSessionKeyMagentoSiteAdminUrl   = "magentoSiteAdminUrl"
	ConstSessionKeyMagentoSiteUrl        = "magentoSiteUrl"
	ConstSessionKeyMagentoConsumerKey    = "magentoConsumerKey"
	ConstSessionKeyMagentoConsumerSecret = "magentoConsumerSecret"
)

// Package global variables
var (
	ConversionFuncs = map[string]interface{}{}
)
