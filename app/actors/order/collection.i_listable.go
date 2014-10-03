package order

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/order"
)

// enumerates items of Product model type
func (it *DefaultOrderCollection) List() ([]models.T_ListItem, error) {
	result := make([]models.T_ListItem, 0)

	dbRecords, err := it.listCollection.Load()
	if err != nil {
		return result, err
	}

	for _, dbRecordData := range dbRecords {

		orderModel, err := order.GetOrderModel()
		if err != nil {
			return result, err
		}
		err = orderModel.FromHashMap(dbRecordData)
		if err != nil {
			return result, err
		}

		// retrieving minimal data needed for list
		resultItem := new(models.T_ListItem)

		resultItem.Id = orderModel.GetId()
		resultItem.Name = orderModel.GetIncrementId()
		resultItem.Image = ""
		resultItem.Desc = utils.InterfaceToString(orderModel.Get("description"))

		// if extra attributes were required
		if len(it.listExtraAtributes) > 0 {
			resultItem.Extra = make(map[string]interface{})

			for _, attributeName := range it.listExtraAtributes {
				resultItem.Extra[attributeName] = orderModel.Get(attributeName)
			}
		}

		result = append(result, *resultItem)
	}

	return result, nil
}

// allows to obtain additional attributes from  List() function
func (it *DefaultOrderCollection) ListAddExtraAttribute(attribute string) error {

	orderModel, err := order.GetOrderModel()
	if err != nil {
		return err
	}

	allowedAttributes := make([]string, 0)
	for _, attributeInfo := range orderModel.GetAttributesInfo() {
		allowedAttributes = append(allowedAttributes, attributeInfo.Attribute)
	}

	if utils.IsInArray(attribute, allowedAttributes) {
		if !utils.IsInListStr(attribute, it.listExtraAtributes) {
			it.listExtraAtributes = append(it.listExtraAtributes, attribute)
		} else {
			return env.ErrorNew("attribute already in list")
		}
	} else {
		return env.ErrorNew("not allowed attribute")
	}

	return nil
}

// adds selection filter to List() function
func (it *DefaultOrderCollection) ListFilterAdd(Attribute string, Operator string, Value interface{}) error {
	it.listCollection.AddFilter(Attribute, Operator, Value.(string))
	return nil
}

// clears presets made by ListFilterAdd() and ListAddExtraAttribute() functions
func (it *DefaultOrderCollection) ListFilterReset() error {
	it.listCollection.ClearFilters()
	return nil
}

// specifies selection paging
func (it *DefaultOrderCollection) ListLimit(offset int, limit int) error {
	return it.listCollection.SetLimit(offset, limit)
}
