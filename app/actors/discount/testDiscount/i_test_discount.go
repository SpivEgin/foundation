package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/composer"
	"github.com/ottemo/foundation/env"
	"fmt"
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
			"amount": checkoutInstance.GetGrandTotal(),
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

	fmt.Printf("action: (%T)%v", action, action)
	if check {
		result = append(result, checkout.StructDiscount{
			Name:      action["Name"].(string),
			Code:      action["Code"].(string),
			Amount:    action["Amount"].(float64),
			IsPercent: action["IsPercent"].(bool),
			Priority:  action["Priority"].(float64),
		})
	}
	return result
}