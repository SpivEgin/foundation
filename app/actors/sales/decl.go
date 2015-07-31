package sales

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/app/models/sales"
)

// Package global constants
const (
	ConstErrorModule = "promotion"
	ConstErrorLevel  = env.ConstErrorLevelModel
)

type RuleEngine struct {
	ruleItems    map[string]sales.InterfaceRuleItem
	inTypeRules  map[string][]sales.InterfaceRuleItem
	outTypeRules map[string][]sales.InterfaceRuleItem

	ruleSets map[string]sales.InterfaceRuleSet
}

type RuleItem struct {
	Code        string
	Name        string
	Description string

	RuleType string

	InType  string
	OutType string

	RequiredArgs map[string]string
	OptionalArgs map[string]string

	Function func (in interface{}, args map[string]interface{}) interface{}
}

type RuleSet struct {
	Enabled bool
	Name string
	Kind string
}