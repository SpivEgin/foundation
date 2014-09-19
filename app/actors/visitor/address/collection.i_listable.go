package address

import (
	"errors"

	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/visitor"
)

// enumerates items of VisitorAddress model type
func (it *DefaultVisitorAddressCollection) List() ([]models.T_ListItem, error) {
	result := make([]models.T_ListItem, 0)

	dbRecords, err := it.listCollection.Load()
	if err != nil {
		return result, err
	}

	for _, dbRecordData := range dbRecords {
		visitorAddressModel, err := visitor.GetVisitorAddressModel()
		if err != nil {
			return result, err
		}
		visitorAddressModel.FromHashMap(dbRecordData)

		// retrieving minimal data needed for list
		resultItem := new(models.T_ListItem)

		resultItem.Id = visitorAddressModel.GetId()
		resultItem.Name = visitorAddressModel.GetZipCode() + " " + visitorAddressModel.GetState() + ", " +
			visitorAddressModel.GetCity() + ", " + visitorAddressModel.GetAddress()
		resultItem.Image = ""
		resultItem.Desc = "Zip: " + visitorAddressModel.GetZipCode() + ", State: " + visitorAddressModel.GetState() +
			", City: " + visitorAddressModel.GetCity() + ", Address: " + visitorAddressModel.GetAddress() +
			", Phone: " + visitorAddressModel.GetPhone()

		// if extra attributes were required
		if len(it.listExtraAtributes) > 0 {
			resultItem.Extra = make(map[string]interface{})

			for _, attributeName := range it.listExtraAtributes {
				resultItem.Extra[attributeName] = visitorAddressModel.Get(attributeName)
			}
		}

		result = append(result, *resultItem)
	}

	return result, nil
}

// allows to obtain additional attributes from  List() function
func (it *DefaultVisitorAddressCollection) ListAddExtraAttribute(attribute string) error {

	if utils.IsAmongStr(attribute, "_id", "id", "visitor_id", "visitorId", "fname", "first_name", "lname", "last_name",
		"address_line1", "address_line2", "company", "country", "city", "state", "phone", "zip", "zip_code") {

		if utils.IsInListStr(attribute, it.listExtraAtributes) {
			it.listExtraAtributes = append(it.listExtraAtributes, attribute)
		} else {
			return errors.New("attribute already in list")
		}
	} else {
		return errors.New("not allowed attribute")
	}

	return nil
}

// adds selection filter to List() function
func (it *DefaultVisitorAddressCollection) ListFilterAdd(Attribute string, Operator string, Value interface{}) error {
	it.listCollection.AddFilter(Attribute, Operator, Value.(string))
	return nil
}

// clears presets made by ListFilterAdd() and ListAddExtraAttribute() functions
func (it *DefaultVisitorAddressCollection) ListFilterReset() error {
	it.listCollection.ClearFilters()
	return nil
}

// sets select pagination
func (it *DefaultVisitorAddressCollection) ListLimit(offset int, limit int) error {
	return it.listCollection.SetLimit(offset, limit)
}
