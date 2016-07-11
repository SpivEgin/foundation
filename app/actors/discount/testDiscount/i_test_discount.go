package testDiscount

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/composer"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// GetName returns name of current discount implementation
func (it *DefaultTestDiscount) GetName() string {
	return "Test Discount"
}

// GetCode returns code of current discount implementation
func (it *DefaultTestDiscount) GetCode() string {
	return "test_discount"
}

// GetPriority returns the code of the current coupon implementation
func (it *DefaultTestDiscount) GetPriority() []float64 {
	baseTestDiscountPriority := utils.InterfaceToFloat64(env.ConfigGetValue(ConstConfigPathTestDiscountApplyPriority))
	cartCalculationPriority := baseTestDiscountPriority + 0.01
	return []float64{baseTestDiscountPriority, cartCalculationPriority}
}

// CalculateDiscount calculates and returns amount and set of applied gift card discounts to given checkout
func (it *DefaultTestDiscount) Calculate(checkoutInstance checkout.InterfaceCheckout, currentPriority float64) []checkout.StructPriceAdjustment {
	var result []checkout.StructPriceAdjustment

	// checking
	in := map[string]interface{}{
			"Cart": map[string]interface{}{
				"cartAmount": checkoutInstance.GetGrandTotal(),
				"visitorIsLogin": checkoutInstance.GetVisitor() != nil,
			},
			//		"vsitor": map[string]interface{}{
			//			"is_admin": checkoutInstance.GetGrandTotal(),
			//		},
	}

//	in := checkoutInstance.GetCart();
	rule := utils.InterfaceToMap(env.ConfigGetValue(ConstConfigPathTestDiscountRule))
	action := utils.InterfaceToMap(env.ConfigGetValue(ConstConfigPathTestDiscountAction))

	for _, object := range rule {
		rule = utils.InterfaceToMap(object);
	}
	for _, object := range action {
		action = utils.InterfaceToMap(object);
	}

	check, err := composer.GetComposer().Check(in, rule)
	if err != nil {
		env.LogError(err)
	}

	if check {
		result = append(result, checkout.StructPriceAdjustment{
			Name:      action["name"].(string),
			Code:      action["code"].(string),
			Amount:    action["amount"].(float64),
			IsPercent: action["is_percent"].(bool),
			Priority:  action["priority"].(float64),
		})
	}
	return result
}