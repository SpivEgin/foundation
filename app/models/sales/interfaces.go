// Package promotion represents abstraction of business layer of promotion objects
package sales

import "time"

const (
	SalesRuleTypeCondition = "condition"
	SalesRuleTypeAction    = "action"
)

// InterfaceSalesRule represents interface to work with sales rule
type InterfaceRuleSet interface {
	IsEnabled() bool

	GetName() string
	GetKind() string

	GetStartDate() time.Time
	GetEndDate() time.Time

	ToHashMap() map[string]interface{}
	// FromHashMap(data map[string]interface{}) error

	Validate() bool
	Apply() error
}

// InterfaceRuleItem represents interface to work with particular sales_rule item
type InterfaceRuleItem interface {
	GetCode() string
	GetName() string
	GetDescription() string

	GetRuleType() string

	GetInType() string
	GetOutType() string

	GetRequiredArguments() map[string]string
	GetOptionalArguments() map[string]string

	Apply(in interface{}, args map[string]interface{}) interface{}
}

// InterfacePromotionEngine represents interface to access promotion engine
type InterfaceRuleEngine interface {
	RegisterRule(rule InterfaceRuleItem) error

	GetRuleByCode(ruleCode string) (InterfaceRuleItem, error)

	GetRegisteredRules() []InterfaceRuleItem
	GetInTypeRules(typeName string) []InterfaceRuleItem
	GetOutTypeRules(typeName string) []InterfaceRuleItem

	AddRuleSet(name string, ruleSet map[string]interface{}) error
	RemoveRuleSet(name string) error
	GetRuleSets() []map[string]interface{}
}