package sales

var (
	registeredRuleEngine InterfaceRuleEngine
)

func RegisterRuleEngine(ruleEngine InterfaceRuleEngine) error {
	registeredRuleEngine = ruleEngine
	return nil
}

func GetRegisterRuleEngine() InterfaceRuleEngine {
	return registeredRuleEngine
}
