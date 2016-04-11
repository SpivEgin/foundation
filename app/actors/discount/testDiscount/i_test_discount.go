package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
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

	// checking
	in := map[string]interface{}{
		"cart": map[string]interface{}{
			"cartAmount": checkoutInstance.GetGrandTotal(),
			"visitorIsLogin": checkoutInstance.GetVisitor() != nil,
		},
	}

//	in := checkoutInstance.GetCart();
	rule := env.ConfigGetValue(ConstConfigPathTestDiscountRule)
	action := env.ConfigGetValue(ConstConfigPathTestDiscountAction).(map[string]interface{})

	check, err := composer.GetComposer().Check(in, rule)
	if err != nil {
		env.LogError(err)
	}

	if check {
		result = append(result, checkout.StructDiscount{
			Name:      action["name"].(string),
			Code:      action["code"].(string),
			Amount:    action["amount"].(float64),
			IsPercent: action["is_percent"].(bool),
			Priority:  action["priority"].(float64),
		})
	}
	return result
}