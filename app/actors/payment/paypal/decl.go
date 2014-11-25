// Package paypal is a PayPal implementation of payment method interface declared in
// "github.com/ottemo/foundation/app/models/checkout" package
package paypal

import (
	"sync"
)

// Package global constants
const (
	ConstPaymentCode = "paypal_express"
	ConstPaymentName = "PayPal Express"

	ConstPaymentActionSale          = "Sale"
	ConstPaymentActionAuthorization = "Authorization"

	ConstConfigPathGroup = "payment.paypal"

	ConstConfigPathEnabled = "payment.paypal.enabled"
	ConstConfigPathTitle   = "payment.paypal.title"

	ConstConfigPathNVP     = "payment.paypal.nvp"
	ConstConfigPathGateway = "payment.paypal.gateway"

	ConstConfigPathUser = "payment.paypal.user"
	ConstConfigPathPass = "payment.paypal.password"

	ConstConfigPathSignature = "payment.paypal.signature"
	ConstConfigPathAction    = "payment.paypal.action"
)

// Package global variables
var (
	waitingTokens      = make(map[string]interface{})
	waitingTokensMutex sync.RWMutex
)

// Express is a implementer of InterfacePaymentMethod for a PayPal Express method
type Express struct{}

// RestAPI is a implementer of InterfacePaymentMethod for a PayPal REST API method (currently not working)
type RestAPI struct{}
