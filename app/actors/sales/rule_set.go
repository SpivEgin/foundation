package sales

import (
	"time"
)

func (it *RuleSet) IsEnabled() bool {
	return it.Enabled
}

func (it *RuleSet) GetName() string {
	return it.Name
}

func (it *RuleSet) GetKind() string {
	return it.Kind
}

func (it *RuleSet) GetStartDate() time.Time {
	return nil
}

func (it *RuleSet) GetEndDate() time.Time {
	return nil
}

func (it *RuleSet) ToHashMap() map[string]interface{} {
	return nil
}


func (it *RuleSet) Validate() bool {
	return false
}

func (it *RuleSet) Apply() error {
	return nil
}