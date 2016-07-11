package testDiscount

import (
	"github.com/ottemo/foundation/env"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	config := env.GetConfig()
	if config == nil {
		return env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "15859fac-8fc0-4fbf-a801-b9cacf70d356", "can't obtain config")
	}

	err := config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscounts,
		Value:       nil,
		Type:        env.ConstConfigTypeGroup,
		Editor:      "",
		Options:     nil,
		Label:       "Test-Discounts",
		Description: "Test Discounts related options",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscountRule,
		Value:       "{}",
		Type:        env.ConstConfigTypeText,
		Editor:      "JSON_editor",
		Options:     nil,
		Label:       "Rule",
		Description: "Rule description",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscountAction,
		Value:       "{}",
		Type:        env.ConstConfigTypeText,
		Editor:      "JSON_editor",
		Options:     nil,
		Label:       "Action",
		Description: "Action description",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscountApplyPriority,
		Value:       2.10,
		Type:        env.ConstConfigTypeFloat,
		Editor:      "line_text",
		Options:     nil,
		Label:       "Discounts calculating position",
		Description: "This value used for using position to calculate it's possible applicable amount (Subtotal - 1, Shipping - 2, Grand total - 3)",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}
