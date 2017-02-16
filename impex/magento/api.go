package magento

import (
	"fmt"
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/product"
	productActor "github.com/ottemo/foundation/app/actors/product"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/app/actors/stock"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
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

		if _, present := v["address"]; present {
			addCustomerAddresses(utils.InterfaceToArray(v["address"]), visitorModel)
		}
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

		if imageName, present := v["image"]; present && imageName != nil {
			_, err = AddImageForCategory(categoryModel, utils.InterfaceToString(imageName), utils.InterfaceToString(v["image_url"]))
			if err != nil {
				fmt.Println(err)
				//return nil, env.ErrorDispatch(err)
			}
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

	for _, value := range jsonResponse {

		orderModel, err := order.GetOrderModel()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}
		v := utils.InterfaceToMap(value)

		// get state code
		// models.ConstStatesList
		// order map with info
		orderRecord := map[string]interface{}{
			"status":          orderStatusMapping[utils.InterfaceToString(v["status"])],
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
		orderRecord["shipping_method"] = ""

			orderModel.FromHashMap(orderRecord)

		err = orderModel.Save()
		if err != nil {
			fmt.Println(err)
			return nil, env.ErrorDispatch(err)
		}

		addItemsToOrder(utils.InterfaceToArray(v["items"]), orderModel)
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
	//fmt.Println(jsonResponse)

	createMagentoIdAttribute()

	dataTypeMap := map[string]interface{}{
		"boolean": utils.ConstDataTypeBoolean,
		"textarea": utils.ConstDataTypeText,
		"text": utils.ConstDataTypeText,
		"media_image": utils.ConstDataTypeText,
		"select": utils.ConstDataTypeText,
		"price": utils.ConstDataTypeMoney,
		"date": utils.ConstDataTypeDatetime,
		"multiselect": utils.ConstDataTypeJSON,
	}
	editorMap := map[string]interface{}{
		"text": "text",
		"textarea": "multiline_text",
		"date": "text",
		"boolean": "boolean",
		"select": "select",
		"multiselect": "multi_select",
	}

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

		attributeFrontendInput := utils.InterfaceToString(v["frontend_input"])
		if attributeFrontendInput == "" {
			continue
			//return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "93457847-8e4d-4536-8985-43f340a1abc4", "attribute label was not specified")
		}


		if _, present := editorMap[attributeFrontendInput]; !present {
			continue
			//return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "93457847-8e4d-4536-8985-43f340a1abc4", "attribute label was not specified")
		}
		fmt.Println(editorMap[attributeFrontendInput])

		if _, present := dataTypeMap[attributeFrontendInput]; !present {
			continue
			//return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "93457847-8e4d-4536-8985-43f340a1abc4", "attribute label was not specified")
		}

		fmt.Println(dataTypeMap[attributeFrontendInput])

		// make product attribute operation
		//---------------------------------
		productModel, err := product.GetProductModel()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		attribute := models.StructAttributeInfo{
			Model:      product.ConstModelNameProduct,
			Collection: productActor.ConstCollectionNameProduct,
			Attribute:  utils.InterfaceToString(attributeName),
			Type:       utils.InterfaceToString(dataTypeMap[attributeFrontendInput]),
			IsRequired: utils.InterfaceToBool(v["is_required"]),
			IsStatic:   false,
			Label:      utils.InterfaceToString(attributeLabel),
			Group:      "Magento",
			Editors:    utils.InterfaceToString(editorMap[attributeFrontendInput]),
			Options:    utils.InterfaceToString(v["options"]),
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

		addImagesToProduct(utils.InterfaceToArray(v["category_ids"]), productModel)

		addProductToCategories(utils.InterfaceToArray(v["images"]), productModel)
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

	// todo move enable stock
	config := env.GetConfig()

	err = config.SetValue(stock.ConstConfigPathEnabled, true)
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


