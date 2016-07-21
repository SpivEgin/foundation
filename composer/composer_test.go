package composer

import (
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/utils"
	"testing"
)

type testIObject struct{ data map[string]interface{} }

func (it *testIObject) Get(attribute string) interface{} {
	if value, present := it.data[attribute]; present {
		return value
	}
	return nil
}

func (it *testIObject) Set(attribute string, value interface{}) error {
	it.data[attribute] = value
	return nil
}

func (it *testIObject) FromHashMap(hashMap map[string]interface{}) error {
	return nil
}

func (it *testIObject) ToHashMap() map[string]interface{} {
	return it.data
}

func (it *testIObject) GetAttributesInfo() []models.StructAttributeInfo {
	return []models.StructAttributeInfo{}
}

func TestOperations(tst *testing.T) {

	var object models.InterfaceObject = &testIObject{data: map[string]interface{}{
		"sku":   "test_product",
		"name":  "Test Product",
		"price": 1.1,
	}}

	tst.Log(object.Get("sku"))
	input := map[string]interface{}{
		"cart": map[string]interface{}{
			"id":       "cart_id",
			"subtotal": 45,
			"items": []map[string]interface{}{
				{
					"id":    "cart_item_1",
					"sku":   "cart_item_sku_1",
					"price": 25,
					"qty":   1,
				},
				{
					"id":    "cart_item_2",
					"sku":   "cart_item_sku_2",
					"price": 10,
					"qty":   2,
				},
			},
		},
	}

	rules, err := utils.DecodeJSONToStringKeyMap(`{
		"cart": {
			"subtotal": {"*lt":{"@": 35,"#": false}, "*gt":15},
			"items": {"*any":{"@id":{"*contains":"cart_item_2"}, "@qty":{"*gt":1}}, "*all":{"@id":{"*contains":"cart_item_"}}}
		}
	}`)
	/*

		, "@test1":true
		"*filter":{}
		"items": {
			"*any":{"id":"cart_item_sku_2"}
		}
		*any:{"id":"cart_item_sku_2"}
	*/

	tst.Log(rules)
	if err != nil {
		tst.Errorf("JSON decode fail: %v", err)
	}

	result, err := GetComposer().Check(input, rules)
	if err != nil {
		tst.Errorf("Validation fail: %v", err)
	} else if !result {
		tst.Error("Validation fail")
	}
	tst.Log(result)
}
