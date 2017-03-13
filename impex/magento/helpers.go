package magento

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"crypto/md5"
	"encoding/hex"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/utils"
	"github.com/ottemo/foundation/app/actors/seo"
	"github.com/ottemo/foundation/app/actors/stock"
	visitorActor "github.com/ottemo/foundation/app/actors/visitor"
	"github.com/ottemo/foundation/app/helpers/attributes"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/category"
	categoryActor "github.com/ottemo/foundation/app/actors/category"
	"github.com/ottemo/foundation/app/models/order"
	orderActor "github.com/ottemo/foundation/app/actors/order"
	"github.com/ottemo/foundation/app/models/product"
	productActor "github.com/ottemo/foundation/app/actors/product"
	"github.com/ottemo/foundation/app/models/visitor"
)

// create attribute `magento_id` to product
func createMagentoIDAttributeToProduct() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "aa206fd6-9788-46c7-943c-a7064183c529", "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
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
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "176cd03e-9227-4dce-8738-0a787f11e4cf", "Attribute magento_id is exist")
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

// create attribute `magento_id` to visitor
func createMagentoIDAttributeToVisitor() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
		}
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b0c94c2-440c-4fd7-9cee-da58e2f97dac", "Can't get collection '"+attributes.ConstCollectionNameCustomAttributes+"': "+err.Error())
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

// AddImageForCategory - add image to category
func AddImageForCategory(categoryModel category.InterfaceCategory, mediaName string, imageURL string) (interface{}, error) {

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
	fileContents, err := getFileContentByURL(imageURL)
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

// AddImageForProduct - add image to product
func AddImageForProduct(productModel product.InterfaceProduct, mediaName string, imageURL string) (interface{}, error) {

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
	fileContents, err := getFileContentByURL(imageURL)
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

// download file by url
func getFileContentByURL(url string) ([]byte, error) {

	var fileContents []byte

	response, err := http.Get(url)
	if err != nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Error while downloading "+url)
		}
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5fcd6e5b-af73-4969-aebc-1a57686c6b40", "Error while downloading "+url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "File does not exit by url "+url)
		}
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "9e22d74b-a037-4def-b8f1-265cb5fab735", "File does not exit by url "+url)
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

// find product by magento id attribute
func getProductByMagentoID(magentoID int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	if (magentoID == 0) {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "magentoID was not specified")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "0edf9c6f-2a3c-4d5f-8fa2-72c4b955782a", "magentoID was not specified")
	}

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

	err = productCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoID))
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

// find visitor by magento id attribute
func getVisitorByMagentoID(magentoID int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	if (magentoID == 0) {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "magentoID was not specified")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ea17f941-bd04-4177-8621-1af202535e22", "magentoID was not specified")
	}

	dbEngine := db.GetDBEngine()
	if dbEngine == nil {

		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Can't obtain DBEngine")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "1bc4f026-a793-444d-bfd1-e79d1294750e", "Can't obtain DBEngine")
	}

	visitorCollectionModel, err := dbEngine.GetCollection(visitor.ConstModelNameVisitor)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return result, env.ErrorDispatch(err)
	}

	err = visitorCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoID))
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

// find visitor by email
func getVisitorByEmail(email string) (map[string]interface{}, error) {
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

	if len(data) == 1 {
		return data[0], nil
	}

	return make(map[string]interface{}), nil
}

// find category by magento id attribute
func getCategoryByMagentoID(magentoID int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	if (magentoID == 0) {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "magentoID was not specified")
		}
		return result, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "5448a7fc-392a-47b1-b242-89d65260c300", "magentoID was not specified")
	}

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

	err = categoryCollectionModel.AddFilter("magento_id", "=", utils.InterfaceToInt(magentoID))
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

// get data from import.json
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

	return jsonResponse, nil
}

// addCustomerAddresses - add addresses to customer
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
			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
			}
			continue
		}

		// save visitor default address
		if utils.InterfaceToBool(addresseMap["default_billing"]) {
			visitorModel.Set("billing_address_id", visitorAddressModel.GetID())
		}

		if utils.InterfaceToBool(addresseMap["default_shipping"]) {
			visitorModel.Set("shipping_address_id", visitorAddressModel.GetID())
		}

		if utils.InterfaceToBool(addresseMap["default_billing"]) && utils.InterfaceToBool(addresseMap["default_shipping"]) {
			err = visitorModel.Save()
			if err != nil {
				if ConstMagentoLog || ConstDebugLog {
					env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
				}
				continue
			}
		}

	}
	return true

}

// addItemsToOrder - add products to order
func addItemsToOrder(items []interface{}, orderModel order.InterfaceOrder) error {

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

		productData, err := getProductByMagentoID(utils.InterfaceToInt(itemData["product_id"]))
		if len(productData) == 0 || err != nil {

			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Product does not exist with magento_id:"+utils.InterfaceToString(itemData["product_id"]))
			}
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "4782fd27-a054-4e92-9176-e716a7bbff65", "Product does not exist with magento_id:"+utils.InterfaceToString(itemData["product_id"]))
			continue
		}

		orderItemData := map[string]interface{}{
			"sku":               utils.InterfaceToString(itemData["sku"]),
			"name":              utils.InterfaceToString(itemData["name"]),
			"short_description": "",
			"idx":               idx,
			"product_id":        productData[0]["_id"],
			"weight":            utils.InterfaceToFloat64(itemData["weight"]),
			"price":             utils.InterfaceToFloat64(itemData["price"]),
			"qty":               utils.InterfaceToInt(itemData["qty_ordered"]),
			"order_id":          orderModel.GetID(),
		}

		orderItemData["options"] = map[string]interface{}{}
		orderItemCollection.Save(orderItemData)
		idx++
	}

	return nil
}

// AddImagesToProduct - add images to product
func AddImagesToProduct(images []interface{}, productModel product.InterfaceProduct) error {

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

// AddProductToCategories - add product to categories
func AddProductToCategories(categories []interface{}, productModel product.InterfaceProduct) error {

	for _, categoryID := range categories {
		categoryData, err := getCategoryByMagentoID(utils.InterfaceToInt(categoryID))
		if len(categoryData) != 1 {

			if ConstMagentoLog || ConstDebugLog {
				env.Log(ConstLogFileName, env.ConstLogPrefixDebug, "Category does not exist with magento_id:"+utils.InterfaceToString(categoryID))
			}
			env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "8fdadaac-88ed-4204-9697-8505b0375b4b", "Category does not exist with magento_id:"+utils.InterfaceToString(categoryID))
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

func generateMagentoAPIData() (map[string]interface{}, error) {
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

	return result, nil
}

// IsMagentoHandler returns middleware API Handler that checks magento import rights
func IsMagentoHandler(next api.FuncAPIHandler) api.FuncAPIHandler {
	return func(context api.InterfaceApplicationContext) (interface{}, error) {
		err := ValidateMagentoRights(context)

		if err != nil {
			context.SetResponseStatusForbidden()
			return nil, err
		}

		return next(context)
	}
}

// ValidateMagentoRights returns nil if context have correct api_key
func ValidateMagentoRights(context api.InterfaceApplicationContext) error {
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	if utils.KeysInMapAndNotBlank(requestData, ConstGETApiKeyParamName) {
		apiKey := utils.InterfaceToString(requestData[ConstGETApiKeyParamName])
		data, err := generateMagentoAPIData()
		if err == nil && data["api_key"] == apiKey {
			return nil
		}
	}

	return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "8afbaca6-e1ec-435a-8208-d427ceb05d71", "Forbidden")
}

// AddProductSEO - add seo data to product
func AddProductSEO(productID string, url string, title string, metaKeywords string, metaDescription string) (interface{}, error) {

	valueURL := utils.InterfaceToString(url)
	valueRewrite := utils.InterfaceToString(productID)

	// looking for duplicated 'url'
	//-----------------------------
	collection, err := db.GetCollection(seo.ConstCollectionNameURLRewrites)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if err := collection.AddFilter("url", "=", valueURL); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "ee7abbc8-ee1f-43c2-b118-5cf6016d6c34", err.Error())
	}

	recordsNumber, err := collection.Count()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if recordsNumber > 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b1ed0906-6a1b-4791-829d-4573660e5621", "rewrite for url '"+valueURL+"' already exists")
	}

	// making new record and storing it
	//---------------------------------
	newRecord := map[string]interface{}{
		"url":              valueURL,
		"type":             "product",
		"rewrite":          valueRewrite,
		"title":            title,
		"meta_keywords":    metaKeywords,
		"meta_description": metaDescription,
	}

	newID, err := collection.Save(newRecord)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	newRecord["_id"] = newID

	return newRecord, nil
}

// EnableStock - enable stock
func EnableStock() error {
	config := env.GetConfig()

	err := config.SetValue(stock.ConstConfigPathEnabled, true)
	if err != nil {
		if ConstMagentoLog || ConstDebugLog {
			env.Log(ConstLogFileName, env.ConstLogPrefixDebug, fmt.Sprintf("Error: %s", err.Error()))
		}
		return env.ErrorDispatch(err)
	}

	return nil
}
