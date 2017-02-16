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
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"io/ioutil"
	"time"
	"strings"
	"net/http"
	"github.com/ottemo/foundation/media"
)

func createMagentoIdAttribute() (bool, error) {

	customAttributesCollection, err := db.GetCollection(attributes.ConstCollectionNameCustomAttributes)
	if err != nil {
		// todo
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b8b1e23-c2ad-45c5-9252-215084a8cd81", "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
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
		return false, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "3b8b1e23-c2ad-45c5-9252-215084a8cd81", "Can't get collection '" + attributes.ConstCollectionNameCustomAttributes + "': " + err.Error())
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

func AddImageForCategory(categoryModel category.InterfaceCategory, mediaName string, imageUrl string) (interface{}, error) {

	// check request context
	//---------------------

	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "82581008-b40e-47b7-ab55-3a0a704eeccd", "media name was not specified")
	}

	mediaType := media.ConstMediaTypeImage

	// get file by url and processing
	//-----------------------
	fileContents, err := getFileContentByUrl(imageUrl)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Adding timestamp to image name to prevent overwriting
	mediaNameParts := strings.SplitN(mediaName, ".", 2)
	mediaName = mediaNameParts[0] + "_" + utils.InterfaceToString(time.Now().Unix()) + "." + mediaNameParts[1]

	err = categoryModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}


func AddImageForProduct(productModel product.InterfaceProduct, mediaName string, imageUrl string) (interface{}, error) {

	// check request context
	//---------------------

	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "82581008-b40e-47b7-ab55-3a0a704eeccd", "media name was not specified")
	}

	mediaType := media.ConstMediaTypeImage

	// get file by url and processing
	//-----------------------
	fileContents, err := getFileContentByUrl(imageUrl)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// Adding timestamp to image name to prevent overwriting
	mediaNameParts := strings.SplitN(mediaName, ".", 2)
	mediaName = mediaNameParts[0] + "_" + utils.InterfaceToString(time.Now().Unix()) + "." + mediaNameParts[1]

	err = productModel.AddMedia(mediaType, mediaName, fileContents)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

func getFileContentByUrl(url string) ([]byte, error) {

	var fileContents []byte

	response, err := http.Get(url)
	if err != nil {
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5fcd6e5b-af73-4969-aebc-1a57686c6b40", "Error while downloading " + url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fileContents, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "9e22d74b-a037-4def-b8f1-265cb5fab735", "File does not exit by url " + url)
	}

	fileContents, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return fileContents, nil
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

func addItemsToOrder(items []interface{}, orderModel order.InterfaceOrder) (error) {

	idx := 1
	for _, item := range items {
		itemData := utils.InterfaceToMap(item)
		// saving category products assignment
		orderItemCollection, err := db.GetCollection(orderActor.ConstCollectionNameOrderItems)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		productData, err := getProductByMagentoId(utils.InterfaceToInt(itemData["product_id"]))
		if len(productData) == 0 || err != nil {
			fmt.Println("continue")
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
				fmt.Println(err)
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
