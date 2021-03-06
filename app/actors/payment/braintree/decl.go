// Package braintree is a "braintree payments" implementation of payment method interface declared in
// "github.com/ottemo/foundation/app/models/checkout" package
package braintree

import (
	"github.com/lionelbarrow/braintree-go"

	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	// --------------------------------------
	// Because of multiple payment modules supported by Braintree constant names and values are divided into
	// General - overall values
	// Method  - specific per method values
	//
	// Note: group name is prefix of elements grouped in frontend

	// --------------------------------------
	// General

	ConstGeneralConfigPathGroup       = "payment.braintreeGeneral"

	ConstGeneralConfigPathEnabled = "payment.braintreeGeneral.enabled"
	ConstGeneralMethodConfigPathName    = "payment.braintreeGeneral.name" // User customized name of the payment method

	ConstGeneralConfigPathEnvironment = "payment.braintreeGeneral.environment"
	ConstGeneralConfigPathMerchantID  = "payment.braintreeGeneral.merchantID"
	ConstGeneralConfigPathPublicKey   = "payment.braintreeGeneral.publicKey"
	ConstGeneralConfigPathPrivateKey  = "payment.braintreeGeneral.privateKey"


	ConstEnvironmentSandbox    = string(braintree.Sandbox)
	ConstEnvironmentProduction = string(braintree.Production)

	ConstErrorModule = "payment/braintree"
	ConstErrorLevel = env.ConstErrorLevelActor

	constLogStorage = "braintree.log"


	constCCMethodCode         = "braintree"           // Method code used in business logic
	constCCMethodInternalName = "Braintree Credit Card" // Human readable name of payment method

)

// CreditCardMethod is a implementer of InterfacePaymentMethod for a Credit Card payment method
type CreditCardMethod struct{}
