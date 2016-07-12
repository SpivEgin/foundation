package composer

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/utils"
	"strings"
)

// setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()
	service.GET("composer/units/:names", APIComposerUnits)
	service.GET("composer", APIComposerInfo)
	// gets correspondence between go and js types
	service.GET("composer/go-json", APIComposerGoTypes)

	// get type info {types: {}, units:{}, types_units_binding:{}}
	service.GET("composer/types/:names", APIComposerTypes)

	// add Check!!!!

	service.GET("composer/elements", APIComposerElements)

	return nil
}

// APIComposerTypes returns related to specified types information
//    types - types of items which could be selected for specified type value
//    units - expressions which could be applied to specified value
//    binding - which units could be applied to specified types
func APIComposerTypes(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})
	typesResult := make(map[string]interface{})
	unitsResult := make(map[string]interface{})
	bindingResult := make(map[string]interface{})

	composer := GetComposer()
	baseForAny := map[string]int{"string": 1, "int": 1, "float": 1, "boolean": 1}
	typeNames := strings.Split(context.GetRequestArgument("names"), ",")
	listUnits := composer.ListUnits()

	for _, typeName := range typeNames {
		// types definition
		if typeInfo := composer.GetType(typeName); typeInfo != nil {
			keyInfo := make(map[string]interface{})
			for _, item := range typeInfo.ListItems() {
				keyInfo[item] = map[string]interface{}{
					"label":       typeInfo.GetLabel(item),
					"description": typeInfo.GetDescription(item),
					"type":        typeInfo.GetType(item),
				}
			}

			typesResult[typeName] = keyInfo
		}

		// units definition
		var binding []string
		for _, unitInfo := range listUnits {
			unitType := unitInfo.GetType(ConstPrefixUnit)

			if unitType == typeName || (baseForAny[typeName] == 1 && unitType == "any") {

				unitName := unitInfo.GetName()
				// binding definition
				binding = append(binding, unitName)

				if unitsResult[unitName] == nil {
					keyInfo := make(map[string]interface{})
					for _, item := range unitInfo.ListItems() {
						keyInfo[item] = map[string]interface{}{
							"label":       unitInfo.GetLabel(item),
							"description": unitInfo.GetDescription(item),
							"type":        unitInfo.GetType(item),
							"required":    unitInfo.IsRequired(item),
						}
					}

					unitsResult[unitName] = keyInfo
				}
			}
		}
		bindingResult[typeName] = binding
	}

	result["types"] = typesResult
	result["units"] = unitsResult
	result["binding"] = bindingResult

	return result, nil
}

// APIComposerGoTypes returns corresponding JS types for GO types
func APIComposerGoTypes(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})
	composer := GetComposer()

	for _, goType := range []string{
		utils.ConstDataTypeID,
		utils.ConstDataTypeBoolean,
		utils.ConstDataTypeVarchar,
		utils.ConstDataTypeText,
		utils.ConstDataTypeDecimal,
		utils.ConstDataTypeMoney,
		utils.ConstDataTypeDatetime,
		utils.ConstDataTypeJSON,
	} {
		typeInfo := composer.GetType(goType)
		for _, item := range typeInfo.ListItems() {
			result[item] = map[string]interface{}{
				"label": typeInfo.GetLabel(item),
				"desc":  typeInfo.GetDescription(item),
				"type":  typeInfo.GetType(item),
			}
		}
	}

	return result, nil
}

// APIComposerUnits returns related to specified units information
func APIComposerUnits(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})

	if composer := GetComposer(); composer != nil {
		units := strings.Split(context.GetRequestArgument("names"), ",")
		for _, unitName := range units {
			if unit := composer.GetUnit(unitName); unit != nil {
				unitInfo := make(map[string]interface{})

				for _, item := range unit.ListItems() {
					unitInfo[item] = map[string]interface{}{
						"label":       unit.GetLabel(item),
						"description": unit.GetDescription(item),
						"type":        unit.GetType(item),
						"required":    unit.IsRequired(item),
					}
				}

				result[unitName] = unitInfo
			}
		}
	}

	return result, nil
}

// APIComposerElements returns all registered units and types
func APIComposerElements(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]interface{})
	if composer := GetComposer(); composer != nil {
		units := make(map[string]interface{})
		for _, unit := range composer.ListUnits() {
			unitInfo := make(map[string]interface{})

			for _, item := range unit.ListItems() {
				unitInfo[item] = map[string]interface{}{
					"label":       unit.GetLabel(item),
					"description": unit.GetDescription(item),
					"type":        unit.GetType(item),
					"required":    unit.IsRequired(item),
				}
			}
			units[unit.GetName()] = unitInfo
		}
		result["units"] = units

		types := make(map[string]interface{})
		for _, unit := range composer.ListTypes() {
			unitInfo := make(map[string]interface{})

			for _, item := range unit.ListItems() {
				unitInfo[item] = map[string]interface{}{
					"label":       unit.GetLabel(item),
					"description": unit.GetDescription(item),
					"type":        unit.GetType(item),
				}
			}
			types[unit.GetName()] = unitInfo
		}
		result["types"] = types
	}

	return result, nil
}

// APIComposerInfo returns composer description and information
func APIComposerInfo(context api.InterfaceApplicationContext) (interface{}, error) {

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
