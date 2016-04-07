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

	//config.UnregisterItem(ConstConfigPathTestDiscountRule)
	err := config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscountRule,
		Value:       "rule",
		Type:        env.ConstConfigTypeJSON,
		Editor:      "JSON_composer",
		Options:     nil,
		Label:       "Rule",
		Description: "Rule description",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	//config.UnregisterItem(ConstConfigPathTestDiscountAction)
	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathTestDiscountAction,
		Value:       "action",
		Type:        env.ConstConfigTypeJSON,
		Editor:      "JSON_composer",
		Options:     nil,
		Label:       "Action",
		Description: "Action description",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}
