package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/utils"
	"github.com/ottemo/foundation/composer"
	"github.com/ottemo/foundation/env"
)

// GetName returns name of current discount implementation
func (it *DefaultTestDiscount) GetName() string {
	return "Test Discount"
}

// GetCode returns code of current discount implementation
func (it *DefaultTestDiscount) GetCode() string {
	return "test_discount"
}

// CalculateDiscount calculates and returns amount and set of applied gift card discounts to given checkout
func (it *DefaultTestDiscount) CalculateDiscount(checkoutInstance checkout.InterfaceCheckout) []checkout.StructDiscount {
	var result []checkout.StructDiscount

	rules, err := utils.DecodeJSONToStringKeyMap(`{
		"cartAmount": {">gt": 15},
		"visitorIsLogin": true,
	}`)

	// checking
	input := map[string]interface{}{
		"cartAmount": checkoutInstance.GetGrandTotal(),
		"visitorIsLogin": checkoutInstance.GetVisitor() != nil,
	}

	check, err := composer.GetComposer().Check(input, rules)
	if err != nil {
		env.LogError(err)
	}

	if check {
		result = append(result, checkout.StructDiscount{
			Name:      "test",
			Code:      "test",
			Amount:    15,
			IsPercent: false,
			Priority:  2.2,
		})
	}
	return result
}