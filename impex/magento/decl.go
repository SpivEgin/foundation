// Package magento is a implementation of import magento data service
package magento

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstErrorModule = "impex_magento"
	ConstErrorLevel  = env.ConstErrorLevelAPI
	ConstGETApiKeyParamName            = "api_key"
	ConstLogFileName                     = "impex_magento.log"
	ConstSessionKeyMagentoRequestToken   = "magentoRequestToken"
)

// Package global variables
var (
	ConstMagentoLog = true  // flag indicates to make log of values going to be processed
	ConstDebugLog = false // flag indicates to have extra log information

	ConversionFuncs = map[string]interface{}{}

	statesList = map[string]string{}

	orderStatusMapping = map[string]interface{}{
		"pending": "pending",
		"pending_ogone": "pending",
		"pending_payment": "pending",
		"pending_paypal": "pending",
		"processing": "processed",
		"payment_review": "processed",
		"paypal_reversed": "processed",
		"paypal_canceled_reversal": "processed",
		"processed_ogone": "processed",
		"processing_ogone": "processed",
		"decline_ogone": "declined",
		"closed": "completed",
		"complete": "completed",
		"canceled": "cancelled",
		"cancel_ogone": "cancelled",
		"holded": "cancelled",
		"fraud": "cancelled",
	}
)
