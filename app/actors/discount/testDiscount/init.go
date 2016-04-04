package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
)

// init makes package self-initialization routine
func init() {
	instance := new(DefaultTestDiscount)
	var _ checkout.InterfaceDiscount = instance
	checkout.RegisterDiscount(instance)
}