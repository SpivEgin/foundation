package composer

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"strings"
)

// setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("composer/unit/:unit", api.ConstRESTOperationGet, composerUnit)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer/units/:typeNames", api.ConstRESTOperationGet, composerUnits)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer/search-unit/:namePattern", api.ConstRESTOperationGet, composerUnitSearch)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer", api.ConstRESTOperationGet, composerInfo)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer/aliases", api.ConstRESTOperationGet, composerAliases)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer/type/:names", api.ConstRESTOperationGet, composerType)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("composer/check", api.ConstRESTOperationCreate, composerCheck)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

func composerCheck(context api.InterfaceApplicationContext) (interface{}, error) {

	input := map[string]interface{}{
		"a": 10,
		"b": 25,
		"c": "123",
		"d": map[string]interface{}{
			"_id":   "0bf7939973984c67b6b56b1c098edfca",
			"name":  "Product1",
			"sku":   "PR-1",
			"price": 10.5,
		},
	}

	rules, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result, err := GetComposer().Check(input, rules)

	if err != nil {
		return "Validation fail", err
	} else if !result {
		return "Validation fail", nil
	}

	return result, nil
}

func composerType(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})

	composer := GetComposer()
	typeNames := strings.Split(context.GetRequestArgument("names"), ",")
	for _, typeName := range typeNames {
		typeInfo := composer.GetType(typeName)
		if typeInfo == nil {
			return result, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "24bb5e98-b5de-4d4a-a5dc-cf2573dae3dd", "Type "+typeName+" is not defined")
		}
		res := make(map[string]interface{})
		for _, item := range typeInfo.ListItems() {
			res[item] = map[string]interface{}{
				"label": typeInfo.GetLabel(item),
				"desc":  typeInfo.GetDescription(item),
				"type":  typeInfo.GetType(item),
			}
		}
		result[typeName] = res
	}
	return result, nil
}

func composerAliases(context api.InterfaceApplicationContext) (interface{}, error) {

	result := map[string]string{
		utils.ConstDataTypeID:       "string",
		utils.ConstDataTypeBoolean:  "boolean",
		utils.ConstDataTypeVarchar:  "string",
		utils.ConstDataTypeText:     "string",
		utils.ConstDataTypeInteger:  "int",
		utils.ConstDataTypeDecimal:  "float",
		utils.ConstDataTypeMoney:    "float",
		utils.ConstDataTypeFloat:    "float",
		utils.ConstDataTypeDatetime: "string",
		utils.ConstDataTypeJSON:     "object",
	}

	return result, nil
}

func composerUnit(context api.InterfaceApplicationContext) (interface{}, error) {

	var result map[string]interface{}

	if composer := GetComposer(); composer != nil {
		unit := composer.GetUnit(context.GetRequestArgument("unit"))
		if unit != nil {
			result = make(map[string]interface{})

			for _, item := range unit.ListItems() {
				result[item] = map[string]interface{}{
					"label":       unit.GetLabel(item),
					"description": unit.GetDescription(item),
					"type":        unit.GetType(item),
					"required":    unit.IsRequired(item),
				}
			}
		}
	}

	return result, nil
}

func composerUnits(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})
	types := context.GetRequestArgument("typeNames")
	typeNames := strings.Split(types, ",")

	for _, typeName := range typeNames {
		typeFilledByUnits := make(map[string]interface{})

		if composer := GetComposer(); composer != nil {
			for _, unit := range composer.ListUnits() {
				unitName := unit.GetName();
				unitType := unit.GetType(ConstPrefixUnit)

				if unitType == typeName {
					unitInfo := make(map[string]interface{})

					for _, item := range unit.ListItems() {
						unitInfo[item] = map[string]interface{}{
							"label":       unit.GetLabel(item),
							"description": unit.GetDescription(item),
							"type":        unit.GetType(item),
							"required":    unit.IsRequired(item),
						}
					}
					typeFilledByUnits[unitName] = unitInfo
				}
			}
			result[typeName] = typeFilledByUnits
		}
	}

	return result, nil
}

func composerUnitSearch(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})

	namePattern := context.GetRequestArgument("namePattern")
	typeFilter := context.GetRequestArguments()
	if _, present := typeFilter["namePattern"]; present {
		delete(typeFilter, "namePattern")
	}

	if composer := GetComposer(); composer != nil {
		for _, unit := range composer.SearchUnits(namePattern, typeFilter) {
			if unitName := unit.GetName(); unitName != "" {
				result[unitName] = map[string]interface{}{
					"name":        unit.GetName(),
					"label":       unit.GetLabel(ConstPrefixUnit),
					"description": unit.GetLabel(ConstPrefixUnit),
					"type":        unit.GetType(ConstPrefixUnit),
					"in_required": unit.IsRequired(ConstPrefixUnit),
				}
			}
		}
	}

	return result, nil
}

func composerInfo(context api.InterfaceApplicationContext) (interface{}, error) {

	result := map[string]interface{}{
		"item_prefix": map[string]interface{}{
			"unit": ConstPrefixUnit,
			"in":   ConstPrefixArg,
			"out":  "",
		},
	}

	if composer := GetComposer(); composer != nil {
		result["composer"] = composer.GetName()
		result["units_count"] = len(composer.ListUnits())
	}

	return result, nil
}
