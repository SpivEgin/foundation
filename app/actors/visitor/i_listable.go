package visitor

import (
	"errors"
	"github.com/ottemo/foundation/db"
)

//---------------------------------
// Implementation specific methods
//---------------------------------

var listCollection db.I_DBCollection = nil

func getListCollection() (db.I_DBCollection, error) {
	if listCollection != nil {
		return listCollection, nil
	} else {
		var err error = nil

		dbEngine := db.GetDBEngine()
		if dbEngine == nil {
			return nil, errors.New("Can't obtain DBEngine")
		}

		listCollection, err = dbEngine.GetCollection("Visitor")

		return listCollection, err
	}
}

//--------------------------
// Interface implementation
//--------------------------

func (it *DefaultVisitor) List() ([]interface{}, error) {
	result := make([]interface{}, 0)

	collection, err := getListCollection()
	if err != nil {
		return result, err
	}

	dbItems, err := collection.Load()
	if err != nil {
		return result, err
	}

	for _, dbItemData := range dbItems {
		model, err := it.New()
		if err != nil {
			return result, err
		}

		visitor := model.(*DefaultVisitor)
		visitor.FromHashMap(dbItemData)

		result = append(result, visitor)
	}

	return result, nil
}

func (it *DefaultVisitor) ListFilterAdd(Attribute string, Operator string, Value interface{}) error {
	collection, err := getListCollection()
	if err != nil {
		return err
	}

	collection.AddFilter(Attribute, Operator, Value.(string))
	return nil
}

func (it *DefaultVisitor) ListFilterReset() error {
	collection, err := getListCollection()
	if err != nil {
		return err
	}

	collection.ClearFilters()
	return nil
}
