package listable

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/db"
)

type ListableHelperDelegates struct {
	CollectionName string
	GetCollection  func() db.I_DBCollection

	ValidateExtraAttributeFunc func(attribute string) bool

	RecordToObjectFunc   func(recordData map[string]interface{}, extraAttributes []string) interface{}
	RecordToListItemFunc func(recordData map[string]interface{}, extraAttributes []string) models.T_ListItem
}

type ListableHelper struct {
	delegate ListableHelperDelegates

	listCollection     db.I_DBCollection
	listExtraAtributes []string
}

// use this function to obtain Listable struct for your object
func NewListable(delegates ListableHelperDelegates) *ListableHelper {
	return &ListableHelper{delegate: delegates}
}