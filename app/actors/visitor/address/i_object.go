package address

import (
	"strings"
	"github.com/ottemo/foundation/app/models"
)



func (it *DefaultVisitorAddress) Get(attribute string) interface{} {
	switch strings.ToLower(attribute) {
	case "_id", "id":
		return it.id
	case "street":
		return it.Street
	case "city":
		return it.City
	case "state":
		return it.State
	case "phone":
		return it.Phone
	case "zip", "zip_code":
		return it.ZipCode
	}

	return nil
}



func (it *DefaultVisitorAddress) Set(attribute string, value interface{}) error {
	switch strings.ToLower(attribute) {
	case "_id", "id":
		it.id = value.(string)
	case "street":
		it.Street = value.(string)
	case "city":
		it.City = value.(string)
	case "state":
		it.State = value.(string)
	case "phone":
		it.Phone = value.(string)
	case "zip", "zip_code":
		it.ZipCode = value.(string)
	}
	return nil
}


func (it *DefaultVisitorAddress) FromHashMap(input map[string]interface{}) error {

	for attribute, value := range input {
		if err := it.Set(attribute, value); err != nil {
			return err
		}
	}

	return nil
}

func (it *DefaultVisitorAddress) ToHashMap() map[string]interface{} {

	result := make( map[string]interface{} )

	result["_id"] = it.id

	result["street"] = it.Street
	result["city"] = it.City
	result["state"] = it.State
	result["phone"] = it.Phone
	result["zip_code"] = it.ZipCode

	return result
}



func (it *DefaultVisitorAddress) GetAttributesInfo() []models.T_AttributeInfo {
	info := []models.T_AttributeInfo {
		models.T_AttributeInfo {
			Model: "VisitorAddress",
			Collection: "visitor_address",
			Attribute: "_id",
			Type: "text",
			Label: "ID",
			Group: "General",
			Editors: "not_editable",
			Options: "",
			Default: "",
		},
		models.T_AttributeInfo {
			Model: "VisitorAddress",
			Collection: "visitor_address",
			Attribute: "street",
			Type: "text",
			Label: "Street",
			Group: "General",
			Editors: "line_text",
			Options: "",
			Default: "",
		},
		models.T_AttributeInfo {
			Model: "VisitorAddress",
			Collection: "visitor_address",
			Attribute: "city",
			Type: "text",
			Label: "City",
			Group: "General",
			Editors: "line_text",
			Options: "",
			Default: "",
		},
		models.T_AttributeInfo {
			Model: "VisitorAddress",
			Collection: "visitor_address",
			Attribute: "phone",
			Type: "text",
			Label: "Phone",
			Group: "General",
			Editors: "line_text",
			Options: "",
			Default: "",
		},
		models.T_AttributeInfo {
			Model: "VisitorAddress",
			Collection: "visitor_address",
			Attribute: "zip_code",
			Type: "text",
			Label: "Zip",
			Group: "General",
			Editors: "line_text",
			Options: "",
			Default: "",
		},
	}

	return info
}
