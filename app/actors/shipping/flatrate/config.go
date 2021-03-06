package flatrate

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"regexp"
	"strings"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	if config := env.GetConfig(); config != nil {
		err := config.RegisterItem(env.StructConfigItem{
			Path:        ConstConfigPathGroup,
			Value:       nil,
			Type:        env.ConstConfigTypeGroup,
			Editor:      "",
			Options:     nil,
			Label:       "Flat Rate",
			Description: "flat rate shipping method",
			Image:       "",
		}, nil)

		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = config.RegisterItem(env.StructConfigItem{
			Path:        ConstConfigPathEnabled,
			Value:       false,
			Type:        env.ConstConfigTypeBoolean,
			Editor:      "boolean",
			Options:     nil,
			Label:       "Enabled",
			Description: "enables/disables shipping method for storefront",
			Image:       "",
		}, nil)

		if err != nil {
			return env.ErrorDispatch(err)
		}

		// validateNewRates validate structure of new shipping rates
		validateNewRates := func(newRatesValues interface{}) (interface{}, error) {

			if utils.InterfaceToString(newRatesValues) == "[]" || newRatesValues == nil || utils.InterfaceToString(newRatesValues) == "" {
				flatRates = make([]interface{}, 0)
				return newRatesValues, nil
			}

			// taking rules as array
			var newRatesArray []interface{}
			switch value := newRatesValues.(type) {
			case string:
				newRatesArray, err = utils.DecodeJSONToArray(value)
				if err != nil {
					return nil, env.ErrorDispatch(err)
				}
			case []interface{}:
				newRatesArray = value
			default:
				err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "df1ccfbd-90ce-412a-b638-5211f23ef525", "can't convert to array")
				return nil, env.ErrorDispatch(err)
			}

			methods := make(map[string]map[string]interface{})

			// checking rules array
			for _, rate := range newRatesArray {
				shippingRate := utils.InterfaceToMap(rate)

				if !utils.KeysInMapAndNotBlank(shippingRate, "title", "code", "price") {
					err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "9593d30e-0df3-4571-8bb1-2e29656b9fe2", "keys 'title', 'code' and 'price' should be not null")
					return nil, env.ErrorDispatch(err)
				}

				shippingRateCode := strings.Replace(strings.ToLower(utils.InterfaceToString(shippingRate["code"])), " ", "_", -1)
				shippingRatePrice := utils.InterfaceToFloat64(shippingRate["price"])

				matched, err := regexp.MatchString("^[a-z0-9_]+$", shippingRateCode)
				if !matched || err != nil {
					err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ea826349-4823-45fe-93c0-46c529f6bcac", "code must be only alphanumeric")
					return nil, env.ErrorDispatch(err)
				}

				if shippingRatePrice < 0 {
					err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "0db031e3-6961-4d90-99e9-736e156acbed", "price can't have negative value")
					return nil, env.ErrorDispatch(err)
				}

				if _, present := methods[shippingRateCode]; present {
					err := env.ErrorNew(ConstErrorModule, ConstErrorLevel, "beeccbfb-68b4-4379-8d9a-78a834c5911c", "duplicate code - "+shippingRateCode)
					return nil, env.ErrorDispatch(err)
				}

				methods[shippingRateCode] = shippingRate
			}

			var rates []interface{}

			for rateCode, rate := range methods {
				rate["code"] = rateCode
				rates = append(rates, rate)
			}
			flatRates = rates

			return newRatesValues, nil
		}

		// grouping rules config setup
		//----------------------------
		err = config.RegisterItem(env.StructConfigItem{
			Path:    ConstConfigPathRates,
			Value:   `[]`,
			Type:    env.ConstConfigTypeJSON,
			Editor:  "multiline_text",
			Options: "",
			Label:   "Shipping Rates",
			Description: `shipping rates, pattern:
[
	{"title": "State Shipping",         "code": "State", "price": 4.99},
	{"title": "Expedited Shipping",     "code": "expedited_shipping", "price": 8, "price_from": 50, "price_to": 160},
	{"title": "International Shipping", "code": "international_shipping", "price": 18, "banned_countries": "Qatar, Mexico, Indonesia", "allowed_countries":"Canada"},
	...
]
make it "[]" to use default method any of additional params such as "banned_countries", "price_from" etc. will be limiting parameters (banned country) `,
			Image: "",
		}, env.FuncConfigValueValidator(validateNewRates))

		if err != nil {
			return env.ErrorDispatch(err)
		}
	} else {
		err := env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "b05c2437-d93f-4b03-8811-2eea9f1c65ef", "Unable to obtain configuration for Flat Rate")
		return env.ErrorDispatch(err)
	}

	return nil
}
