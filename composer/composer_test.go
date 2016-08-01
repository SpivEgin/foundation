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
		"Cart": map[string]interface{}{
			"id":       "cart_id",
			"subtotal": 45,
			"items": []map[string]interface{}{
				{
					"id":        "cart_item_1",
					"productID": "cart_item_1",
					"sku":       "cart_item_sku_1",
					"price":     25,
					"qty":       1,
					"options":   map[string]interface{}{},
				},
				{
					"id":        "cart_item_2",
					"productID": "cart_item_2",
					"sku":       "cart_item_sku_2",
					"price":     10,
					"qty":       1,
					"options":   map[string]interface{}{},
				},
				{
					"id":        "cart_item_2",
					"productID": "cart_item_2",
					"sku":       "cart_item_sku_2",
					"price":     10,
					"qty":       1,
					"options":   map[string]interface{}{},
				},
			},
		},
	}

	rules, err := utils.DecodeJSONToStringKeyMap(`{
		"Cart": {
			"subtotal": {"*lt":{"@": 35,"#": false}, "*gt":15},
			"items": [{"*have": {"@productID": "cart_item_1", "*gt":0}}, {"*have": {"@productID": "cart_item_2", "*gt":0}}]
		}
	}`)
	/*
		"items": {"*have": {"*gt": 0,"@qty": { "*gt": 1}, "@sku": {"*contains":"cart_item"}}}

		"cartItems": {"*any":{"@id":{"*contains":"cart_item_2"}, "@qty":{"*gt":1}}, "*all":{"@id":{"*contains":"cart_item_"}}}

		, "@test1":true
		"*filter":{}
		"items": {
			"*any":{"id":"cart_item_sku_2"}
		}
		*any:{"id":"cart_item_sku_2"}

		// 1. Rule interpretation: one of this products should be in the Cart
		// in this case qty isn't used and valuable only presence of such items
		{"product_in_cart": ["54cf601a42189a77b5fe56ec", "54cf601900c58bbe78123064"]}
		"Cart": {
			"items": {"*have": {"@productID": "cart_item_1"}}
		}
		"Cart": {
			"items": {"*have": {"@productID": "cart_item_2"}}
		}

		// 2. Rule interpretation: both of this products should be in the Cart
		// in this case qty isn't used and valuable only presence of such items
		{ "products_in_cart": ["cart_item_1", "cart_item_2"]}

		"items": [{"*have": {"@productID": "cart_item_1", "*gt":0}}, {"*have": {"@productID": "cart_item_2", "*gt":0}}]

		// 3. Rule interpretation: all products should be in cart with specified qty
		// in this case qty is used directly
		{"products_in_qty":{"54cf601a42189a77b5fe56ec":1, "54cf601900c58bbe78123064": 1}}



		ADD unit to convert cart to flat cart by some params (args)
		Simplify rules so they can define only condition:
		1. Once; // just true
		2. By amount of qualified products; // true and divide on module level
		3. Same as above but below maximum application
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
