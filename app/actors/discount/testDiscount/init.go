package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/api"
)

// init makes package self-initialization routine
func init() {
	instance := new(DefaultTestDiscount)
	var _ checkout.InterfacePriceAdjustment = instance
	checkout.RegisterPriceAdjustment(instance)

	env.RegisterOnConfigStart(setupConfig)
	api.RegisterOnRestServiceStart(setupAPI)
}