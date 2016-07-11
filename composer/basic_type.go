package composer

import (
	"sort"
)

// GetName returns the type name
func (it *BasicType) GetName() string {
	return it.Name
}

// ListItems returns list of items which could be selected for specified type
func (it *BasicType) ListItems() []string {
	var result []string

	for itemName := range it.Type {
		result = append(result, itemName)
	}

	sort.Strings(result)
	return result
}

// GetType returns type for specified item
func (it *BasicType) GetType(item string) string {
	if value, present := it.Type[item]; present {
		return value
	}
	return ""
}

// GetLabel returns label for specified item
func (it *BasicType) GetLabel(item string) string {
	if value, present := it.Label[item]; present {
		return value
	}
	return ""
}

// GetDescription returns description for specified item
func (it *BasicType) GetDescription(item string) string {
	if value, present := it.Description[item]; present {
		return value
	}
	return ""
}
