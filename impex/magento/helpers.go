package magento

import (
	"fmt"
	"github.com/ottemo/foundation/api"

	"github.com/ottemo/foundation/app/helpers/attributes"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	categoryActor "github.com/ottemo/foundation/app/actors/category"
	"github.com/ottemo/foundation/app/models/order"
	orderActor "github.com/ottemo/foundation/app/actors/order"
	"github.com/ottemo/foundation/app/models/product"
	productActor "github.com/ottemo/foundation/app/actors/product"
	"github.com/ottemo/foundation/app/models/visitor"
	visitorActor "github.com/ottemo/foundation/app/actors/visitor"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"io/ioutil"
	"time"
	"strings"
	"net/http"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/app"
	"crypto/md5"
	"encoding/hex"
)

func createMagentoIdAttributeToProduct() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b0c94c2-440c-4fd7-9cee-da58e2f97dac", "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
	}
	attributeName := "magento_id"

	customAttributesCollection.AddFilter("model", "=", product.ConstModelNameProduct)
	customAttributesCollection.AddFilter("attribute", "=", attributeName)
	records, err := customAttributesCollection.Load()

	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return false, env.ErrorDispatch(err)
	}

	if len(records) != 0 {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Attribute magento_id is exist")
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "343b1eb7-07a2-435c-86a8-93da702d17f8", "Attribute magento_id is exist")
	}

	// make product attribute operation
	//---------------------------------
	productModel, err := product.GetProductModel()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return false, env.ErrorDispatch(err)
	}

	attribute := models.StructAttributeInfo{
		Model:      product.ConstModelNameProduct,
		Collection: productActor.ConstCollectionNameProduct,
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

func createMagentoIdAttributeToVisitor() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b0c94c2-440c-4fd7-9cee-da58e2f97dac", "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
	}
	attributeName := "magento_id"

	customAttributesCollection.AddFilter("model", "=", visitor.ConstModelNameVisitor)
	customAttributesCollection.AddFilter("attribute", "=", attributeName)
	records, err := customAttributesCollection.Load()

	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return false, env.ErrorDispatch(err)
	}

	if len(records) != 0 {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Attribute magento_id is exist")
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "343b1eb7-07a2-435c-86a8-93da702d17f8", "Attribute magento_id is exist")
	}

	// make visitor attribute operation
	//---------------------------------
	visitorModel, err := visitor.GetVisitorModel()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return false, env.ErrorDispatch(err)
	}

	attribute := models.StructAttributeInfo{
		Model:      visitor.ConstModelNameVisitor,
		Collection: visitorActor.ConstCollectionNameVisitor,
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

	visitorModel.AddNewAttribute(attribute)

	return true, nil
}

func AddImageForCategory(categoryModel category.InterfaceCategory, mediaName string, imageUrl string) (interface{}, error) {

	// check request context
	//---------------------

	if mediaName == "" {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "media name was not specified")
		}
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "82581008-b40e-47b7-ab55-3a0a704eeccd", "media name was not specified")
	}

	mediaType := media.ConstMediaTypeImage

	// get file by url and processing
	//-----------------------
	fileContents, err := getFileContentByUrl(imageUrl)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	// Adding timestamp to image name to prevent overwriting
	mediaNameParts := strings.SplitN(mediaName, ".", 2)
	mediaName = mediaNameParts[0] + "_" + utils.InterfaceToString(time.Now().Unix()) + "." + mediaNameParts[1]

	err = categoryModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}


func AddImageForProduct(productModel product.InterfaceProduct, mediaName string, imageUrl string) (interface{}, error) {

	// check request context
	//---------------------

	if mediaName == "" {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "media name was not specified")
		}
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fef8633f-ce68-49d5-bbfb-03db95731609", "media name was not specified")
	}

	mediaType := media.ConstMediaTypeImage

	// get file by url and processing
	//-----------------------
	fileContents, err := getFileContentByUrl(imageUrl)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	// Adding timestamp to image name to prevent overwriting
	mediaNameParts := strings.SplitN(mediaName, ".", 2)
	mediaName = mediaNameParts[0] + "_" + utils.InterfaceToString(time.Now().Unix()) + "." + mediaNameParts[1]

	err = productModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

func getFileContentByUrl(url string) ([]byte, error) {

	var fileContents []byte

	response, err := http.Get(url)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Error while downloading " + url)
		}
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5fcd6e5b-af73-4969-aebc-1a57686c6b40", "Error while downloading " + url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "File does not exit by url " + url)
		}
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "9e22d74b-a037-4def-b8f1-265cb5fab735", "File does not exit by url " + url)
	}

	fileContents, err = ioutil.ReadAll(response.Body)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	return fileContents, nil
}


func getProductByMagentoId(magentoId int) ([]map[string]interface{}, error) {
	// todo check magentoId
	var result []map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't obtain DBEngine")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d48f88f1-f0c0-4719-9acc-01c67db50d39", "Can't obtain DBEngine")
	}

	productCollectionModel, err := dbEngine.GetCollection(product.ConstModelNameProduct)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	err = productCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoId))
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	result, err = productCollectionModel.Load()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	return result, nil
}

func getVisitorByMagentoId(magentoId int) ([]map[string]interface{}, error) {
	// todo check magentoId
	var result []map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't obtain DBEngine")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d115a091-7d9a-4d14-807f-7aa8f11eebb3", "Can't obtain DBEngine")
	}

	visitorCollectionModel, err := dbEngine.GetCollection(visitor.ConstModelNameVisitor)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	err = visitorCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoId))
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	result, err = visitorCollectionModel.Load()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	return result, nil
}

func getVisitorByMail(email string) (map[string]interface{}, error) {
	// todo check magentoId
	var result map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't obtain DBEngine")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d115a091-7d9a-4d14-807f-7aa8f11eebb3", "Can't obtain DBEngine")
	}

	visitorCollectionModel, err := dbEngine.GetCollection(visitor.ConstModelNameVisitor)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	err = visitorCollectionModel.AddFilter("email", "=", utils.InterfaceToString(email))
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	data, err := visitorCollectionModel.Load()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	if (len(data) == 1 ) {
		return data[0], nil
	}

	return make(map[string]interface{}), nil
}

func getCategoryByMagentoId(magentoId int) ([]map[string]interface{}, error) {
	// todo check magentoId
	var result []map[string]interface{}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't obtain DBEngine")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "9ea4a67a-eb2a-4bcb-8bfa-37efc3341a53", "Can't obtain DBEngine")
	}

	categoryCollectionModel, err := dbEngine.GetCollection(category.ConstModelNameCategory)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	err = categoryCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoId))
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	result, err = categoryCollectionModel.Load()
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	return result, nil
}

func getDataFromContext(context api.InterfaceApplicationContext) ([]interface{}, error) {

	responseBody, err := ioutil.ReadAll(context.GetRequestFile("import.json"))
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return nil, env.ErrorDispatch(err)
	}

	jsonResponse, err := utils.DecodeJSONToArray(responseBody)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
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


		// save visitor default address
		if (utils.InterfaceToBool(addresseMap["default_billing"])) {
			visitorModel.Set("billing_address_id", visitorAddressModel.GetID())
		}

		if (utils.InterfaceToBool(addresseMap["default_shipping"])) {
			visitorModel.Set("shipping_address_id", visitorAddressModel.GetID())
		}


		if (utils.InterfaceToBool(addresseMap["default_billing"]) && utils.InterfaceToBool(addresseMap["default_shipping"])) {
			err = visitorModel.Save()
			if err != nil {
				fmt.Println(err)
				return false
			}
		}


	}
	return true

}

func addItemsToOrder(items []interface{}, orderModel order.InterfaceOrder) (error) {

	idx := 1
	for _, item := range items {
		itemData := utils.InterfaceToMap(item)
		// saving category products assignment
		orderItemCollection, err := db.GetCollection(orderActor.ConstCollectionNameOrderItems)
		if err != nil {
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return env.ErrorDispatch(err)
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(itemData["product_id"]))
		if len(productData) == 0 || err != nil {

			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Product does not exist with magento_id:" + utils.InterfaceToString(itemData["product_id"]))
			}
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4782fd27-a054-4e92-9176-e716a7bbff65", "Product does not exist with magento_id:" + utils.InterfaceToString(itemData["product_id"]))
			continue
		}

		orderItemData := map[string]interface{}{
			"sku": utils.InterfaceToString(itemData["sku"]),
			"name": utils.InterfaceToString(itemData["name"]),
			"short_description": "",
			"idx" : idx,
			"product_id": productData[0]["_id"],
			"weight": utils.InterfaceToFloat64(itemData["weight"]),
			"price": utils.InterfaceToFloat64(itemData["price"]),
			"qty": utils.InterfaceToInt(itemData["qty_ordered"]),
			"order_id": orderModel.GetID(),
		}

		orderItemData["options"] = map[string]interface{}{

		}
		orderItemCollection.Save(orderItemData)
		idx++
	}

	return nil
}

func addImagesToProduct(images []interface{}, productModel product.InterfaceProduct) (error) {

	for _, image := range images {
		imageData := utils.InterfaceToMap(image)
		if imageName, present := imageData["image_name"]; present && imageName != nil {

			_, err := AddImageForProduct(productModel, utils.InterfaceToString(imageName), utils.InterfaceToString(imageData["image_url"]))
			if err != nil {
				if ConstMagentoLog || ConstDebugLog {
					env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
				}
				return env.ErrorDispatch(err)
			}
		}
	}
	return nil
}

func addProductToCategories(categories []interface{}, productModel product.InterfaceProduct) (error) {


	for _, categoryId := range categories {
		categoryData, err := getCategoryByMagentoId(utils.InterfaceToInt(categoryId))
		if (len(categoryData) != 1 ) {

			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Category does not exist with magento_id:" + utils.InterfaceToString(categoryId))
			}
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "8fdadaac-88ed-4204-9697-8505b0375b4b", "Category does not exist with magento_id:" + utils.InterfaceToString(categoryId))
			continue
		}

		// saving category products assignment
		junctionCollection, err := db.GetCollection(categoryActor.ConstCollectionNameCategoryProductJunction)
		if err != nil {

			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			return env.ErrorDispatch(err)
		}
		junctionCollection.Save(map[string]interface{}{"category_id": categoryData[0]["_id"], "product_id": productModel.GetID()})
	}

	return nil
}

func generateMagentoApiData() (map[string]interface{}, error) {
	var foundationURL = utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathFoundationURL))
	if foundationURL == "" {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e6b45640-7ab4-4a7a-bcf9-7a873c42ec36", "Foundation URL empty")
	}
	result := map[string]interface{}{
		"foundation_url": foundationURL,
	}

	md5Model := md5.New()
	md5Model.Write([]byte(foundationURL))
	md5Hash := hex.EncodeToString(md5Model.Sum(nil))

	result["api_key"] = md5Hash

	return  result, nil
}

func IsMagentoHandler(next api.FuncAPIHandler) api.FuncAPIHandler {
	return func(context api.InterfaceApplicationContext) (interface{}, error) {
		fmt.Println(context)
		err := ValidateMagentoRights(context)

		if err != nil {
			context.SetResponseStatusForbidden()
			return nil, err
		}

		return next(context)
	}
}

func ValidateMagentoRights(context api.InterfaceApplicationContext) error {
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	if utils.KeysInMapAndNotBlank(requestData, ConstGETApiKeyParamName) {
		apiKey := utils.InterfaceToString(requestData[ConstGETApiKeyParamName])
		data, err := generateMagentoApiData()
		if (err == nil && data["api_key"] == apiKey) {
			return nil
		}
	}

	return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "8afbaca6-e1ec-435a-8208-d427ceb05d71", "Forbidden")
}
