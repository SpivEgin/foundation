package sales

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/app/models/sales"
)


func (it *RuleEngine) RegisterRule(rule sales.InterfaceRuleItem) error {
	ruleCode := rule.GetCode()

	if ruleCode == "" {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "2bd26072-1a0c-4011-9d34-f2be20edf7ea", "Rule code should be specified")
	}

	if present, _ := it.ruleItems[ruleCode]; present {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "7aa573c7-cfe4-4970-81a5-8de3ba156d57", "Rule already registered")
	}

	it.ruleItems[ruleCode] = rule

	ruleInType := rule.GetInType()
	if present, value := it.inTypeRules[ruleInType]; present {
		it.inTypeRules[ruleInType] = append(value, ruleInType)
	} else {
		it.inTypeRules[ruleInType] = []sales.InterfaceRuleItem{rule}
	}

	ruleOutType := rule.GetOutType()
	if present, value := it.inTypeRules[ruleInType]; present {
		it.outTypeRules[ruleOutType] = append(value, ruleOutType)
	} else {
		it.outTypeRules[ruleOutType] = []sales.InterfaceRuleItem{rule}
	}

	return nil
}


func (it *RuleEngine) GetRuleByCode(ruleCode string) (sales.InterfaceRuleItem, error) {
	return it.ruleItems[ruleCode]
}


func (it *RuleEngine) GetRegisteredRules() []sales.InterfaceRuleItem {
	return it.ruleItems
}


func (it *RuleEngine) GetInTypeRules(typeName string) []sales.InterfaceRuleItem {
	if present, value := it.inTypeRules[typeName]; present {
		return value
	}

	return []sales.InterfaceRuleItem{}
}


func (it *RuleEngine) GetOutTypeRules(typeName string) []sales.InterfaceRuleItem {
	if present, value := it.outTypeRules[typeName]; present {
		return value
	}

	return []sales.InterfaceRuleItem{}
}


func (it *RuleEngine) AddRuleSet(name string, ruleSet map[string]interface{}) error {
	it.ruleSets[name] = ruleSet
	return nil
}


func (it *RuleEngine) RemoveRuleSet(name string) error {
	if _, present := it.ruleSets[name]; !present {
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "39b026c8-fa30-4d4a-ad41-e563724619b0", "Rule set not exists")
	}
	delete(it.ruleSets, name)
	return nil
}


func (it *RuleEngine) GetRuleSets() []map[string]interface{} {
	return it.ruleSets
}