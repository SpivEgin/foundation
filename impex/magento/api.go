package magento

import (
	"fmt"
	"github.com/ottemo/foundation/api"

	"github.com/ottemo/foundation/app/helpers/attributes"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"io/ioutil"
	"time"
)

// setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	service.POST("impex/magento/visitor", magentoVisitorRequest)
	service.POST("impex/magento/order", magentoOrderRequest)
	service.POST("impex/magento/category", magentoCategoryRequest)
	service.POST("impex/magento/product/attributes", magentoProductAttributesRequest)
	service.POST("impex/magento/products", magentoProductRequest)
	service.POST("impex/magento/stock", magentoStockRequest)

	return nil
}

func magentoVisitorRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoVisitorRequest")
	fmt.Println(context)
	fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for _, value := range jsonResponse {

		visitorModel, err := visitor.GetVisitorModel()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}

		v := utils.InterfaceToMap(value)

		//// visitor map with info
		visitorRecord := map[string]interface{}{
			"magento_id": utils.InterfaceToString(v["entity_id"]),
			"email":      utils.InterfaceToString(v["email"]),
			"first_name": utils.InterfaceToString(v["first_name"]),
			"last_name":  utils.InterfaceToString(v["last_name"]),
			"is_admin":   false,
			"password":   "test",
			"created_at": time.Now(),
		}

		visitorModel.FromHashMap(visitorRecord)

		err = visitorModel.Save()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}

		//if utils.InterfaceToArray(v["address"]) {
		addCustomerAddresses(utils.InterfaceToArray(v["address"]), visitorModel)
		//}
	}
	var result []string

	return result, nil
}

func magentoCategoryRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoCategoryRequest")
	fmt.Println(context)
	fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	//fmt.Println(jsonResponse)

	for _, value := range jsonResponse {

		categoryModel, err := category.GetCategoryModel()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}
		v := utils.InterfaceToMap(value)

		// category map with info
		categoryRecord := map[string]interface{}{
			"name":        utils.InterfaceToString(v["name"]),
			"description": utils.InterfaceToString(v["description"]),
			"last_name":   utils.InterfaceToString(v["last_name"]),
			"enabled":     utils.InterfaceToBool(v["is_active"]),
			"magento_id":  utils.InterfaceToString(v["entity_id"]),
			// todo image
			//"image":  utils.InterfaceToString(v["image"]),
			"created_at": time.Now(),
		}

		if utils.InterfaceToInt(v["parent_id"]) > 0 {
			fmt.Println(v["parent_id"])

			rowData, err := getCategoryByMagentoId(utils.InterfaceToInt(v["parent_id"]))
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}

			if len(rowData) == 1 {
				categoryRecord["parent_id"] = rowData[0]["_id"]
			}
		}

		categoryModel.FromHashMap(categoryRecord)

		err = categoryModel.Save()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}
	}

	var result []string

	return result, nil
}

func magentoOrderRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoOrderRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	//fmt.Println(jsonResponse)

	statesList := map[string]string{}
	for code, name := range models.ConstStatesList {
		statesList[name] = code
	}

	fmt.Println(statesList)

	for _, value := range jsonResponse {
		//fmt.Println(value)
		orderModel, err := order.GetOrderModel()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}
		v := utils.InterfaceToMap(value)
		fmt.Println("")
		fmt.Println("")
		fmt.Println(v)
		fmt.Println("")
		fmt.Println("")

		// get state code
		// models.ConstStatesList
		// order map with info
		orderRecord := map[string]interface{}{
			"status":          utils.InterfaceToString(v["status"]),
			"increment_id":    utils.InterfaceToString(v["increment_id"]),
			"magento_id":      utils.InterfaceToString(v["entity_id"]),
			"grand_total":     utils.InterfaceToFloat64(v["grand_total"]),
			"shipping_amount": utils.InterfaceToFloat64(v["base_shipping_amount"]),
			"subtotal":        utils.InterfaceToFloat64(v["subtotal"]),
			"tax_amount":      utils.InterfaceToFloat64(v["tax_amount"]),
			"discount":        utils.InterfaceToFloat64(v["discount_amount"]),
			"created_at":      time.Now(),
		}

		shippingAddress := utils.InterfaceToMap(v["shippingAddress"])
		orderRecord["shipping_address"] = map[string]interface{}{
			"country":       utils.InterfaceToString(shippingAddress["country_id"]),
			"address_line1": utils.InterfaceToString(shippingAddress["street"]),
			"zip_code":      utils.InterfaceToString(shippingAddress["postcode"]),
			"_id":           "",
			"last_name":     utils.InterfaceToString(shippingAddress["lastname"]),
			"state":         statesList[utils.InterfaceToString(shippingAddress["region"])],
			"company":       utils.InterfaceToString(shippingAddress["company"]),
			"phone":         utils.InterfaceToString(shippingAddress["telephone"]),
			"visitor_id":    "",
			"address_line2": "",
			"first_name":    utils.InterfaceToString(shippingAddress["firstname"]),
			"city":          utils.InterfaceToString(shippingAddress["city"]),
		}

		billingAddress := utils.InterfaceToMap(v["billingAddress"])
		orderRecord["billing_address"] = map[string]interface{}{
			"country":       utils.InterfaceToString(billingAddress["country_id"]),
			"address_line1": utils.InterfaceToString(billingAddress["street"]),
			"zip_code":      utils.InterfaceToString(billingAddress["postcode"]),
			"_id":           "",
			"last_name":     utils.InterfaceToString(billingAddress["lastname"]),
			"state":         statesList[utils.InterfaceToString(billingAddress["region"])],
			"company":       utils.InterfaceToString(billingAddress["company"]),
			"phone":         utils.InterfaceToString(billingAddress["telephone"]),
			"visitor_id":    "",
			"address_line2": "",
			"first_name":    utils.InterfaceToString(billingAddress["firstname"]),
			"city":          utils.InterfaceToString(billingAddress["city"]),
		}

		orderRecord["shipping_info"] = map[string]interface{}{
			"shipping_method_name": utils.InterfaceToString(shippingAddress["shipping_description"]),
		}

		orderModel.FromHashMap(orderRecord)

		err = orderModel.Save()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}
	}

	var result []string

	return result, nil
}

func magentoProductAttributesRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoProductAttributesRequest")
	fmt.Println(context)
	fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	fmt.Println(jsonResponse)

	createMagentoIdAttribute()
	//<select id="frontend_input" name="frontend_input" title="Catalog Input Type for Store Owner" class=" select">
	//<option value="text" selected="selected">Text Field</option>
	//<option value="textarea">Text Area</option>
	//<option value="date">Date</option>
	//<option value="boolean">Yes/No</option>
	//<option value="multiselect">Multiple Select</option>
	//<option value="select">Dropdown</option>
	//<option value="price">Price</option>
	//<option value="media_image">Media Image</option>
	//<option value="weee">Fixed Product Tax</option>
	//</select>
	//dataTypeMap := map[string]interface{}{
	//	"boolean": utils.ConstDataTypeBoolean,
	//	"": utils.ConstDataTypeVarchar,
	//	"text": utils.ConstDataTypeText,
	//	"": utils.ConstDataTypeInteger,
	//	"": utils.ConstDataTypeDecimal,
	//	"price": utils.ConstDataTypeMoney,
	//	"": utils.ConstDataTypeFloat,
	//	"date": utils.ConstDataTypeDatetime,
	//	"": utils.ConstDataTypeJSON,
	//	"": utils.ConstDataTypeHTML,
	//}

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		attributeName := utils.InterfaceToString(v["attribute_code"])
		if attributeName == "" {
			continue
			//return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2f7aec81-dba8-4cad-b683-23c5d0a08cf5", "attribute name was not specified")
		}

		attributeLabel := utils.InterfaceToString(v["frontend_label"])
		if attributeLabel == "" {
			continue
			//return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "93457847-8e4d-4536-8985-43f340a1abc4", "attribute label was not specified")
		}

		// make product attribute operation
		//---------------------------------
		productModel, err := product.GetProductModel()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		attribute := models.StructAttributeInfo{
			Model:      product.ConstModelNameProduct,
			Collection: "product",
			Attribute:  utils.InterfaceToString(attributeName),
			Type:       utils.ConstDataTypeText,
			IsRequired: utils.InterfaceToBool(v["is_required"]),
			IsStatic:   false,
			Label:      utils.InterfaceToString(attributeLabel),
			Group:      "Magento",
			Editors:    "text",
			Options:    "",
			Default:    utils.InterfaceToString(v["default_value"]),
			Validators: "",
			IsLayered:  false,
			IsPublic:   utils.InterfaceToBool(v["is_visible"]),
			//magento_id:   utils.InterfaceToInt(v["attribute_id"]),
		}

		productModel.AddNewAttribute(attribute)
		//err = productModel.AddNewAttribute(attribute)
		//if err != nil {
		//	return nil, env.ErrorDispatch(err)
		//}
	}

	var result []string

	return result, nil
}

func magentoProductRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoProductRequest")
	fmt.Println(context)
	fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	//fmt.Println(jsonResponse)

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		if !utils.KeysInMapAndNotBlank(v, "sku", "name") {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2a0cf2b0-215e-4b53-bf55-98fbfe22cd27", "product name and/or sku were not specified")
		}

		// create product operation
		//-------------------------
		productModel, err := product.GetProductModel()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(v["entity_id"]))
		if len(productData) == 1 && err == nil {
			continue
		}

		for attribute, value := range v {
			if attribute == "entity_id" {
				attribute = "magento_id"
			}
			err := productModel.Set(attribute, value)
			if err != nil {
				fmt.Println(err)
				//return nil, env.ErrorDispatch(err)
			}
		}

		err = productModel.Save()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	var result []string

	return result, nil
}

func magentoStockRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	fmt.Println("magentoStockRequest")
	fmt.Println(context)
	fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	//fmt.Println(jsonResponse)
	options := make(map[string]interface{})

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		stockManager := product.GetRegisteredStock()
		if stockManager == nil {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c03d0b95-400e-415f-8c4a-26863993adbc", "no registered stock manager")
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(v["product_id"]))
		if len(productData) == 0 || err != nil {
			fmt.Println("continue")
			continue
		}

		qty := utils.InterfaceToInt(v["qty"])

		err = stockManager.SetProductQty(utils.InterfaceToString(productData[0]["_id"]), options, qty)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	var result []string

	return result, nil
}

func getDataFromContext(context api.InterfaceApplicationContext) ([]interface{}, error) {

	responseBody, err := ioutil.ReadAll(context.GetRequestFile("import.json"))
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	jsonResponse, err := utils.DecodeJSONToArray(responseBody)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	//fmt.Println(utils.InterfaceToString(jsonResponse))

	return jsonResponse, nil
}

func addCustomerAddresses(addresses []interface{}, visitorModel visitor.InterfaceVisitor) bool {

	for _, addresse := range addresses {
		visitorAddressModel, err := visitor.GetVisitorAddressModel()
		if err != nil {
			return false
		}

		addresseMap := utils.InterfaceToMap(addresse)

		// todo set default address
		// visitor addresse map with info
		addresseRecord := map[string]interface{}{
			"visitor_id":    visitorModel.GetID(),
			"first_name":    utils.InterfaceToString(addresseMap["firstname"]),
			"last_name":     utils.InterfaceToString(addresseMap["lastname"]),
			"city":          utils.InterfaceToString(addresseMap["city"]),
			"zip_code":      utils.InterfaceToString(addresseMap["postcode"]),
			"company":       utils.InterfaceToString(addresseMap["company"]),
			"phone":         utils.InterfaceToString(addresseMap["telephone"]),
			"country":       utils.InterfaceToString(addresseMap["country_id"]),
			"address_line1": utils.InterfaceToArray(addresseMap["street"])[0],
			"address_line2": "",
		}

		visitorAddressModel.FromHashMap(addresseRecord)
		err = visitorAddressModel.Save()
		if err != nil {
			fmt.Println(err)
			return false
		}

	}
	return true

}

func getProductByMagentoId(magentoId int) ([]map[string]interface{}, error) {
	// todo check magentoId
	var result []map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "642ed88a-6d8b-48a1-9b3c-feac54c4d9a3", "Can't obtain DBEngine")
	}

	productCollectionModel, err := dbEngine.GetCollection(product.ConstModelNameProduct)
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	err = productCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoId))
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	result, err = productCollectionModel.Load()
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	return result, nil
}

func getCategoryByMagentoId(magentoId int) ([]map[string]interface{}, error) {
	// todo check magentoId
	var result []map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "642ed88a-6d8b-48a1-9b3c-feac54c4d9a3", "Can't obtain DBEngine")
	}

	categoryCollectionModel, err := dbEngine.GetCollection(category.ConstModelNameCategory)
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	err = categoryCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoId))
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	result, err = categoryCollectionModel.Load()
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	return result, nil
}

func createMagentoIdAttribute() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {
		// todo
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b8b1e23-c2ad-45c5-9252-215084a8cd81", "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
	}
	attributeName := "magento_id"

	customAttributesCollection.AddFilter("model", "=", product.ConstModelNameProduct)
	customAttributesCollection.AddFilter("attribute", "=", attributeName)
	records, err := customAttributesCollection.Load()

	if err != nil {
		return false, env.ErrorDispatch(err)
	}

	if len(records) > 0 {
		// todo
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b8b1e23-c2ad-45c5-9252-215084a8cd81", "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
	}

	// make product attribute operation
	//---------------------------------
	productModel, err := product.GetProductModel()
	if err != nil {
		return false, env.ErrorDispatch(err)
	}

	attribute := models.StructAttributeInfo{
		Model:      product.ConstModelNameProduct,
		Collection: "product",
		Attribute:  attributeName,
		Type:       utils.ConstDataTypeText,
		IsRequired: false,
		IsStatic:   false,
		Label:      "Magento Id",
		Group:      "Magento",
		Editors:    "text",
		Options:    "",
		Default:    "",
		Validators: "",
		IsLayered:  false,
		IsPublic:   false,
	}

	productModel.AddNewAttribute(attribute)

	return true, nil
}
