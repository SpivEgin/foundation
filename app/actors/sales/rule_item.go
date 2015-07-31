package sales

func (it *RuleItem) GetCode() string {
	return it.Code
}

func (it *RuleItem) GetName() string {
	return it.Name
}

func (it *RuleItem) GetDescription() string {
	return it.Description
}

func (it *RuleItem) GetRuleType() string {
	return it.RuleType
}

func (it *RuleItem) GetInType() string {
	return it.InType
}

func (it *RuleItem) GetOutType() string {
	return it.OutType
}

func (it *RuleItem) GetRequiredArguments() map[string]string {
	return it.RequiredArgs
}

func (it *RuleItem) GetOptionalArguments() map[string]string {
	return it.OptionalArgs
}

func (it *RuleItem) Apply(in interface{}, args map[string]interface{}) interface{} {
	return in
}