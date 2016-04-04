package testDiscount

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstSessionKeyAppliedTestDiscountCodes = "applied_test_discount_codes"
	//ConstCollectionNameTestDiscount  = "test_discount"

	ConstErrorModule = "testDiscount"
	ConstErrorLevel  = env.ConstErrorLevelActor
)

// DefaultTestDiscount is a default implementer of InterfaceTestDiscount
type DefaultTestDiscount struct{}

