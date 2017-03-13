package magento

import (
	"fmt"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/product"
	productActor "github.com/ottemo/foundation/app/actors/product"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

)

// setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	service.GET("impex/magento/options", api.IsAdminHandler(magentoOptionsRequest))

	service.POST("impex/magento/visitor", IsMagentoHandler(magentoVisitorRequest))
	service.POST("impex/magento/order", IsMagentoHandler(magentoOrderRequest))
	service.POST("impex/magento/category", IsMagentoHandler(magentoCategoryRequest))
	service.POST("impex/magento/product/attributes", IsMagentoHandler(magentoProductAttributesRequest))
	service.POST("impex/magento/products", IsMagentoHandler(magentoProductRequest))
	service.POST("impex/magento/stock", IsMagentoHandler(magentoStockRequest))

	return nil
}

func magentoOptionsRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	return generateMagentoApiData()
}

func magentoVisitorRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoVisitorRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	createMagentoIdAttributeToVisitor()
	var count int

	for _, value := range jsonResponse {

		visitorModel, err := visitor.GetVisitorModel()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		v := utils.InterfaceToMap(value)
		email := utils.InterfaceToString(v["email"])
		visitorData, err := getVisitorByEmail(email)
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		if visitorData["_id"] != nil {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "7620b94a-8abe-4f50-a279-d9c254b86b25", "Customer exist with email "+email)
			continue
		}

		// visitor map with info
		visitorRecord := map[string]interface{}{
			"magento_id": utils.InterfaceToString(v["entity_id"]),
			"email":      email,
			"first_name": utils.InterfaceToString(v["firstname"]),
			"last_name":  utils.InterfaceToString(v["lastname"]),
			"is_admin":   false,
			"password":   "test",
			"created_at": time.Now(),
		}

		visitorModel.FromHashMap(visitorRecord)

		err = visitorModel.Save()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		if _, present := v["address"]; present {
			addCustomerAddresses(utils.InterfaceToArray(v["address"]), visitorModel)
		}

		count++
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}

func magentoCategoryRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoCategoryRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	var count int

	//fmt.Println(jsonResponse)

	for _, value := range jsonResponse {

		categoryModel, err := category.GetCategoryModel()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
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
			"created_at":  time.Now(),
		}

		if utils.InterfaceToInt(v["parent_id"]) > 0 {
			rowData, err := getCategoryByMagentoId(utils.InterfaceToInt(v["parent_id"]))
			if err != nil {
				if ConstMagentoLog || ConstDebugLog {
					env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
				}
				return nil, env.ErrorDispatch(err)
			}

			if len(rowData) == 1 {
				categoryRecord["parent_id"] = rowData[0]["_id"]
			}
		}

		categoryModel.FromHashMap(categoryRecord)

		err = categoryModel.Save()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		if imageName, present := v["image"]; present && imageName != nil {
			_, err = AddImageForCategory(categoryModel, utils.InterfaceToString(imageName), utils.InterfaceToString(v["image_url"]))
			if err != nil {
				if ConstMagentoLog || ConstDebugLog {
					env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
				}
				//return nil, env.ErrorDispatch(err)
			}
		}

		count++
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}

func magentoOrderRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoOrderRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	var count int

	//fmt.Println(jsonResponse)

	for _, value := range jsonResponse {

		orderModel, err := order.GetOrderModel()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}
		v := utils.InterfaceToMap(value)
		visitorId := ""
		// todo visitor
		visitorData, err := getVisitorByMagentoId(utils.InterfaceToInt(v["customer_id"]))
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		if len(visitorData) == 1 {
			visitorDataArray := utils.InterfaceToArray(visitorData)[0]

			visitorDataMap := utils.InterfaceToMap(visitorDataArray)
			visitorId = utils.InterfaceToString(visitorDataMap["_id"])
		}

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
			"visitor_id":      visitorId,
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
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		count++

		addItemsToOrder(utils.InterfaceToArray(v["items"]), orderModel)
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}

func magentoProductAttributesRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoProductAttributesRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}
	//fmt.Println(jsonResponse)

	createMagentoIdAttributeToProduct()

	dataTypeMap := map[string]interface{}{
		"boolean":     utils.ConstDataTypeBoolean,
		"textarea":    utils.ConstDataTypeText,
		"text":        utils.ConstDataTypeText,
		"media_image": utils.ConstDataTypeText,
		"select":      utils.ConstDataTypeText,
		"price":       utils.ConstDataTypeMoney,
		"date":        utils.ConstDataTypeDatetime,
		"multiselect": utils.ConstDataTypeJSON,
	}
	editorMap := map[string]interface{}{
		"text":        "text",
		"textarea":    "multiline_text",
		"date":        "text",
		"boolean":     "boolean",
		"select":      "select",
		"multiselect": "multi_select",
	}

	var count int

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		attributeName := utils.InterfaceToString(v["attribute_code"])
		if attributeName == "" {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "080df187-4c40-478a-9323-5533a1dfef96", "attribute name was not specified")
			continue
		}

		attributeLabel := utils.InterfaceToString(v["frontend_label"])
		if attributeLabel == "" {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "34847e62-1dd7-41c3-b3d9-6d637c8d9de5", "attribute frontend label was not specified")
			continue
		}

		attributeFrontendInput := utils.InterfaceToString(v["frontend_input"])
		if attributeFrontendInput == "" {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "904598bc-0c7e-4050-b1ba-b8d164468858", "attribute frontend input was not specified")
			continue
		}

		if _, present := editorMap[attributeFrontendInput]; !present {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5c3934ed-9e6a-48bc-85d5-d42bfd332fa6", "attribute editor was not specified")
			continue
		}

		if _, present := dataTypeMap[attributeFrontendInput]; !present {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "592ca965-401e-4a9c-98e4-5eb5add650b0", "attribute type was not specified")
			continue
		}

		// make product attribute operation
		//---------------------------------
		productModel, err := product.GetProductModel()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
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
		}

		err = productModel.AddNewAttribute(attribute)
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			//return nil, env.ErrorDispatch(err)
		}

		count++
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}

func magentoProductRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoProductRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	var count int

	//fmt.Println(jsonResponse)

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		if !utils.KeysInMapAndNotBlank(v, "sku", "name") {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6aaa7688-fb36-4a6b-a59c-bb4709c5df9b", "product name and/or sku were not specified")
			continue
		}

		// create product operation
		//-------------------------
		productModel, err := product.GetProductModel()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(v["entity_id"]))
		if len(productData) == 1 && err == nil {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "02f31b09-c0ae-493a-9202-65cf7ee92177", "Product exists with magento_id:"+utils.InterfaceToString(v["product_id"]))
			continue
		}

		for attribute, value := range v {
			if attribute == "entity_id" {
				attribute = "magento_id"
			}
			err := productModel.Set(attribute, value)
			if err != nil {
				if ConstMagentoLog || ConstDebugLog {
					env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
				}
				//return nil, env.ErrorDispatch(err)
			}
		}

		err = productModel.Save()
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		AddImagesToProduct(utils.InterfaceToArray(v["category_ids"]), productModel)

		AddProductToCategories(utils.InterfaceToArray(v["images"]), productModel)

		AddProductSEO(productModel.GetID(),
			utils.InterfaceToString(v["url_path"]),
			utils.InterfaceToString(v["meta_title"]),
			utils.InterfaceToString(v["meta_keyword"]),
			utils.InterfaceToString(v["meta_description"]))

		count++
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}

func magentoStockRequest(context api.InterfaceApplicationContext) (interface{}, error) {
	//fmt.Println("magentoStockRequest")
	//fmt.Println(context)
	//fmt.Println(context.GetRequestFile("import.json"))

	jsonResponse, err := getDataFromContext(context)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	var count int

	EnableStock()

	//fmt.Println(jsonResponse)
	options := make(map[string]interface{})

	for _, value := range jsonResponse {
		v := utils.InterfaceToMap(value)

		stockManager := product.GetRegisteredStock()
		if stockManager == nil {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "90bef837-a41c-4a9b-a071-8547be5ba966", "no registered stock manager")
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(v["product_id"]))
		if len(productData) == 0 || err != nil {
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d37e7a80-4d8d-41cf-a2f5-55c0af7fa2e6", "Product does not exist with magento_id:"+utils.InterfaceToString(v["product_id"]))
			continue
		}

		qty := utils.InterfaceToInt(v["qty"])

		err = stockManager.SetProductQty(utils.InterfaceToString(productData[0]["_id"]), options, qty)
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return nil, env.ErrorDispatch(err)
		}

		count++
	}

	result := map[string]interface{}{
		"count": count,
	}

	return result, nil
}
