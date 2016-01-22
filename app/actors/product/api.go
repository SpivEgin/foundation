package product

import (
	"io/ioutil"
	"math/rand"
	"mime"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("product/:productID", api.ConstRESTOperationGet, APIGetProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("product", api.ConstRESTOperationCreate, APICreateProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("product/:productID", api.ConstRESTOperationUpdate, APIUpdateProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("product/:productID", api.ConstRESTOperationDelete, APIDeleteProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("product/:productID/media/:mediaType/:mediaName", api.ConstRESTOperationGet, APIGetMedia)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("product/:productID/media/:mediaType", api.ConstRESTOperationGet, APIListMedia)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("product/:productID/media/:mediaType/:mediaName", api.ConstRESTOperationCreate, APIAddMediaForProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("product/:productID/media/:mediaType/:mediaName", api.ConstRESTOperationDelete, APIRemoveMediaForProduct)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("product/:productID/mediapath/:mediaType", api.ConstRESTOperationGet, APIGetMediaPath)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("product/:productID/related", api.ConstRESTOperationGet, APIListRelatedProducts)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("products", api.ConstRESTOperationGet, APIListProducts)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("products/attributes", api.ConstRESTOperationGet, APIListProductAttributes)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("products/attribute", api.ConstRESTOperationCreate, APICreateProductAttribute)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("products/attribute/:attribute", api.ConstRESTOperationUpdate, APIUpdateProductAttribute)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("products/attribute/:attribute", api.ConstRESTOperationDelete, APIDeleteProductsAttribute)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("products/shop", api.ConstRESTOperationGet, APIListShopProducts)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("products/shop/layers", api.ConstRESTOperationGet, APIGetShopLayers)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// APIListProductAttributes returns a list of product attributes
func APIListProductAttributes(context api.InterfaceApplicationContext) (interface{}, error) {
	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attrInfo := productModel.GetAttributesInfo()

	return attrInfo, nil
}

// APIUpdateProductAttribute updates existing custom attribute of product model
//   - attribute name/code should be provided in "attribute" argument
//   - attribute parameters should be provided in request content
//   - attribute parameters "id" and "name" will be ignored
//   - static attributes can not be changed
func APIUpdateProductAttribute(context api.InterfaceApplicationContext) (interface{}, error) {
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attributeName := context.GetRequestArgument("attribute")
	if attributeName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "cb8f7251-e22b-4605-97bb-e239df6c7aac", "attribute name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for _, attribute := range productModel.GetAttributesInfo() {
		if attribute.Attribute == attributeName {
			if attribute.IsStatic == true {
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2893262f-a61a-42f8-9c75-e763e0a5c8ca", "can't edit static attributes")
			}

			for key, value := range requestData {
				switch strings.ToLower(key) {
				case "label":
					attribute.Label = utils.InterfaceToString(value)
				case "group":
					attribute.Group = utils.InterfaceToString(value)
				case "editors":
					attribute.Editors = utils.InterfaceToString(value)
				case "options":
					attribute.Options = utils.InterfaceToString(value)
				case "default":
					attribute.Default = utils.InterfaceToString(value)
				case "validators":
					attribute.Validators = utils.InterfaceToString(value)
				case "isrequired", "required":
					attribute.IsRequired = utils.InterfaceToBool(value)
				case "islayered", "layered":
					attribute.IsLayered = utils.InterfaceToBool(value)
				case "ispublic", "public":
					attribute.IsPublic = utils.InterfaceToBool(value)
				}
			}
			err := productModel.EditAttribute(attributeName, attribute)
			if err != nil {
				return nil, err
			}
			return attribute, nil
		}
	}

	return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2893262f-a61a-42f8-9c75-e763e0a5c8ca", "attribute not found")
}

// APICreateProductAttribute creates a new custom attribute for a product model
//   - attribute parameters "Attribute" and "Label" are required
func APICreateProductAttribute(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attributeName, isSpecified := requestData["Attribute"]
	if !isSpecified {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2f7aec81-dba8-4cad-b683-23c5d0a08cf5", "attribute name was not specified")
	}

	attributeLabel, isSpecified := requestData["Label"]
	if !isSpecified {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "93457847-8e4d-4536-8985-43f340a1abc4", "attribute label was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// make product attribute operation
	//---------------------------------
	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attribute := models.StructAttributeInfo{
		Model:      product.ConstModelNameProduct,
		Collection: ConstCollectionNameProduct,
		Attribute:  utils.InterfaceToString(attributeName),
		Type:       utils.ConstDataTypeText,
		IsRequired: false,
		IsStatic:   false,
		Label:      utils.InterfaceToString(attributeLabel),
		Group:      "General",
		Editors:    "text",
		Options:    "",
		Default:    "",
		Validators: "",
		IsLayered:  false,
		IsPublic:   false,
	}

	for key, value := range requestData {
		switch strings.ToLower(key) {
		case "type":
			attribute.Type = utils.InterfaceToString(value)
		case "group":
			attribute.Group = utils.InterfaceToString(value)
		case "editors":
			attribute.Editors = utils.InterfaceToString(value)
		case "options":
			attribute.Options = utils.InterfaceToString(value)
		case "default":
			attribute.Default = utils.InterfaceToString(value)
		case "validators":
			attribute.Validators = utils.InterfaceToString(value)
		case "isrequired", "required":
			attribute.IsRequired = utils.InterfaceToBool(value)
		case "islayered", "layered":
			attribute.IsLayered = utils.InterfaceToBool(value)
		case "ispublic", "public":
			attribute.IsPublic = utils.InterfaceToBool(value)
		}
	}

	err = productModel.AddNewAttribute(attribute)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return attribute, nil
}

// APIDeleteProductsAttribute removes existing custom attribute of a product model
//   - attribute name/code should be provided in "attribute" argument
func APIDeleteProductsAttribute(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//--------------------
	attributeName := context.GetRequestArgument("attribute")
	if attributeName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "cb8f7251-e22b-4605-97bb-e239df6c7aac", "attribute name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// remove attribute actions
	//-------------------------
	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = productModel.RemoveAttribute(attributeName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIGetProduct return specified product information
//   - product id should be specified in "productID" argument
func APIGetProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "feb3a463-622b-477e-a22d-c0a3fd1972dc", "product id was not specified")
	}

	// load product operation
	//-----------------------
	productModel, err := product.LoadProductByID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// not allowing to see disabled products if not admin
	if api.ValidateAdminRights(context) != nil && productModel.GetEnabled() == false {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "153673ac-1008-40b5-ada9-2286ad3f02b0", "product not available")
	}

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	result := productModel.ToHashMap()

	itemImages, err := mediaStorage.GetAllSizes(product.ConstModelNameProduct, productModel.GetID(), ConstProductMediaTypeImage)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	defaultImage := productModel.GetDefaultImage()

	// move default image to first position in array
	if defaultImage != "" && len(itemImages) > 1 {
		defaultImageName := defaultImage[strings.LastIndex(defaultImage, "/")+1 : strings.Index(defaultImage, ".")]
		found := false
		for index, images := range itemImages {
			for _, sizeValue := range images {
				if strings.Contains(sizeValue, defaultImageName) {
					found = true
					itemImages = append(itemImages[:index], itemImages[index+1:]...)
					itemImages = append([]map[string]string{images}, itemImages...)
				}
				break
			}
			if found {
				break
			}
		}
	}

	result["images"] = itemImages

	return result, nil
}

// APICreateProduct creates a new product
//   - product attributes must be provided in request content
//   - "sku" and "name" attributes are required
func APICreateProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if !utils.KeysInMapAndNotBlank(context.GetRequestArguments(), "sku", "name") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2a0cf2b0-215e-4b53-bf55-98fbfe22cd27", "product name and/or sku were not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// create product operation
	//-------------------------
	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range requestData {
		err := productModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = productModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return productModel.ToHashMap(), nil
}

// APIDeleteProduct deletes existing product
//   - product id must be specified in "productID" argument
func APIDeleteProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//--------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f35af170-8172-4ec0-b30d-ab883231d222", "product id was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// delete operation
	//-----------------
	productModel, err := product.GetProductModelAndSetID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = productModel.Delete()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIUpdateProduct updates existing product
//   - product id should be specified in "productID" argument
//   - product attributes should be specified in content
func APIUpdateProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "c91e8fc7-ca77-40d1-823c-e50f90b8b4b5", "product id was not specified")
	}

	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fffccbad-455a-4fff-81d4-8919ae3a5c35", "unexpected request content")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// update operations
	//------------------
	productModel, err := product.LoadProductByID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attrName, attrVal := range requestData {
		err = productModel.Set(attrName, attrVal)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = productModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return productModel.ToHashMap(), nil
}

// APIGetMediaPath returns relative path to product media files within media library
//   - product id, media type must be specified in "productID" and "mediaType" arguments
func APIGetMediaPath(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6597ff92-f2ee-4233-bcf9-eb73b957fb05", "product id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "75c00741-5873-4be1-9fa0-df9d2956d3de", "media type was not specified")
	}

	// list media operation
	//---------------------
	productModel, err := product.GetProductModelAndSetID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaList, err := productModel.GetMediaPath(mediaType)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return mediaList, nil
}

// APIListMedia returns lost of media files assigned to specified product
//   - product id, media type must be specified in "productID" and "mediaType" arguments
func APIListMedia(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "52677450-8a7f-49c9-a472-51d0e80bc7ca", "product id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b8b31a9f-6fac-47b3-89e2-c9b3e589a8f6", "media type was not specified")
	}

	// list media operation
	//---------------------
	productModel, err := product.GetProductModelAndSetID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaList, err := productModel.ListMedia(mediaType)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return mediaList, nil
}

// APIAddMediaForProduct uploads and assigns media file send in request for a specified product
//   - product id, media type and media name should be specified in "productID", "mediaType" and "mediaName" arguments
//   - media file should be provided in "file" field
func APIAddMediaForProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "a4696c5d-3276-4272-8d86-8061e57743a5", "product id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f3ea9a01-412a-4af2-9496-cb58cdb8139d", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "23fb7617-f19a-4505-b706-10f7898fd980", "media name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// income file processing
	//-----------------------
	files := context.GetRequestFiles()
	if len(files) == 0 {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "75a2ddaf-b63d-4eed-b16d-4b32778f5fc1", "media file was not specified")
	}

	var fileContents []byte
	for _, fileReader := range files {
		contents, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
		fileContents = contents
		break
	}

	// add media operation
	//--------------------
	productModel, err := product.GetProductModelAndSetID(productID)
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

// APIRemoveMediaForProduct removes media content from specified product
//   - product id, media type and media name should be specified in "productID", "mediaType" and "mediaName" arguments
func APIRemoveMediaForProduct(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "f5f77b7f-6606-4bdd-a113-0a3b26f5759c", "product id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e81b841f-8253-4b66-ac7d-2cc9a484044c", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "63b37b08-3b21-48b7-9058-291bb7e635a1", "media name was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// list media operation
	//---------------------
	productModel, err := product.GetProductModelAndSetID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = productModel.RemoveMedia(mediaType, mediaName)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// APIGetMedia returns media contents for a product (file assigned to a product)
//   - product id, media type and media name must be specified in "productID", "mediaType" and "mediaName" arguments
//   - on success case not a JSON data returns, but media file
func APIGetMedia(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d33b8a67-359f-4a3e-b626-f58b6c70f09f", "product id was not specified")
	}

	mediaType := context.GetRequestArgument("mediaType")
	if mediaType == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "d081b726-caf4-4694-baaa-7b1801ca9713", "media type was not specified")
	}

	mediaName := context.GetRequestArgument("mediaName")
	if mediaName == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "124c8b9d-1a6b-491c-97ba-a03e8c828337", "media name was not specified")
	}

	context.SetResponseContentType(mime.TypeByExtension(mediaName))

	// list media operation
	//---------------------
	productModel, err := product.GetProductModelAndSetID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return productModel.GetMedia(mediaType, mediaName)
}

// APIListProducts returns a list of available products
//   - if "action" parameter is set to "count" result value will be just a number of list items
//   - visitors can not see disabled products, but administrators can
func APIListProducts(context api.InterfaceApplicationContext) (interface{}, error) {

	var productCollectionModel product.InterfaceProductCollection
	var err error

	if productCollectionModel, err = product.GetProductCollectionModel(); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// filters handle
	models.ApplyFilters(context, productCollectionModel.GetDBCollection())

	// exclude disabled products for visitors, but not Admins
	if err := api.ValidateAdminRights(context); err != nil {
		productCollectionModel.GetDBCollection().AddFilter("enabled", "=", true)
	}

	// check "count" request
	if context.GetRequestArgument(api.ConstRESTActionParameter) == "count" {
		return productCollectionModel.GetDBCollection().Count()
	}

	// limit parameter handle
	productCollectionModel.ListLimit(models.GetListLimit(context))

	// extra parameter handle
	models.ApplyExtraAttributes(context, productCollectionModel)

	listItems, err := productCollectionModel.List()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	var result []map[string]interface{}

	for _, listItem := range listItems {

		itemImages, err := mediaStorage.GetAllSizes(product.ConstModelNameProduct, listItem.ID, ConstProductMediaTypeImage)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		// move default image to first position in array
		if listItem.Image != "" && len(itemImages) > 1 {
			defaultImageName := listItem.Image[strings.LastIndex(listItem.Image, "/")+1 : strings.Index(listItem.Image, ".")]
			found := false
			for index, images := range itemImages {
				for _, sizeValue := range images {
					if strings.Contains(sizeValue, defaultImageName) {
						found = true
						itemImages = append(itemImages[:index], itemImages[index+1:]...)
						itemImages = append([]map[string]string{images}, itemImages...)
					}
					break
				}
				if found {
					break
				}
			}
		}

		item := map[string]interface{}{
			"ID":     listItem.ID,
			"Name":   listItem.Name,
			"Desc":   listItem.Desc,
			"Extra":  listItem.Extra,
			"Image":  listItem.Image,
			"Images": itemImages,
		}

		result = append(result, item)
	}

	return result, nil
}

// APIListRelatedProducts returns related products list for a given product
func APIListRelatedProducts(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	productID := context.GetRequestArgument("productID")
	if productID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "55aa2eee-0407-4094-a90a-5d69d8c1efcc", "product id was not specified")
	}

	count := 5
	if countValue := utils.InterfaceToInt(api.GetArgumentOrContentValue(context, "count")); countValue > 0 {
		count = countValue
	}

	// load product operation
	//-----------------------
	productModel, err := product.LoadProductByID(productID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// result := make([]models.StructListItem, 0)
	var result []models.StructListItem

	relatedPids := utils.InterfaceToArray(productModel.Get("related_pids"))

	if len(relatedPids) > count {
		// indexes := make([]int, 0)
		var indexes []int
		for len(indexes) < count {

			new := rand.Intn(len(relatedPids))

			inArray := false
			for _, b := range indexes {
				if utils.InterfaceToInt(b) == new {
					inArray = true
				}
			}
			if !inArray {
				indexes = append(indexes, new)
			}
		}
		for _, index := range indexes {
			if productID := utils.InterfaceToString(relatedPids[index]); productID != "" {
				if productModel, err := product.LoadProductByID(productID); err == nil {

					// not allowing to see disabled products if not admin
					if err := api.ValidateAdminRights(context); err != nil {
						if productModel.GetEnabled() == false {
							continue
						}
					}

					if err == nil {
						resultItem := new(models.StructListItem)

						mediaPath, err := productModel.GetMediaPath("image")
						if err != nil {
							return result, env.ErrorDispatch(err)
						}

						resultItem.ID = productModel.GetID()
						resultItem.Name = "[" + productModel.GetSku() + "] " + productModel.GetName()
						resultItem.Image = ""
						resultItem.Desc = productModel.GetShortDescription()

						if productModel.GetDefaultImage() != "" {
							resultItem.Image = mediaPath + productModel.GetDefaultImage()
						}

						extra := utils.InterfaceToString(api.GetArgumentOrContentValue(context, "extra"))
						if extra != "" {
							resultItem.Extra = make(map[string]interface{})
							extra := utils.Explode(utils.InterfaceToString(extra), ",")
							for _, value := range extra {
								resultItem.Extra[value] = productModel.Get(value)
							}
						}

						result = append(result, *resultItem)
					}
				}
			}
		}
	} else {
		for _, productID := range relatedPids {
			if productID == "" {
				continue
			}

			productModel, err := product.LoadProductByID(utils.InterfaceToString(productID))
			if err == nil {
				resultItem := new(models.StructListItem)

				mediaPath, err := productModel.GetMediaPath("image")
				if err != nil {
					return result, env.ErrorDispatch(err)
				}

				resultItem.ID = productModel.GetID()
				resultItem.Name = "[" + productModel.GetSku() + "] " + productModel.GetName()
				resultItem.Image = ""
				resultItem.Desc = productModel.GetShortDescription()

				if productModel.GetDefaultImage() != "" {
					resultItem.Image = mediaPath + productModel.GetDefaultImage()
				}

				extra := utils.InterfaceToString(api.GetArgumentOrContentValue(context, "extra"))
				if extra != "" {
					resultItem.Extra = make(map[string]interface{})
					extra := utils.Explode(utils.InterfaceToString(extra), ",")
					for _, value := range extra {
						resultItem.Extra[value] = productModel.Get(value)
					}
				}

				result = append(result, *resultItem)
			}
		}
	}

	return result, nil
}

// APIListShopProducts returns a list of available products for a shop
//   - for a not admins available products are limited to enabled ones
func APIListShopProducts(context api.InterfaceApplicationContext) (interface{}, error) {

	productsCollection, err := product.GetProductCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	dbCollection := productsCollection.GetDBCollection()

	// filters handle
	models.ApplyFilters(context, dbCollection)

	// not allowing to see disabled products if not admin
	if err := api.ValidateAdminRights(context); err != nil {
		productsCollection.GetDBCollection().AddFilter("enabled", "=", true)
	}

	mediaStorage, err := media.GetMediaStorage()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// preparing product information
	var result []map[string]interface{}

	for _, productModel := range productsCollection.ListProducts() {
		productInfo := productModel.ToHashMap()

		defaultImage := utils.InterfaceToString(productInfo["default_image"])
		itemImages, err := mediaStorage.GetSizes(product.ConstModelNameProduct, productModel.GetID(), ConstProductMediaTypeImage, defaultImage)
		if err != nil {
			env.LogError(err)
		}

		productInfo["image"] = itemImages
		result = append(result, productInfo)
	}

	return result, nil
}

// APIGetShopLayers returns layered navigation options for a shop products list
//   - for a not admins available products are limited to enabled ones
func APIGetShopLayers(context api.InterfaceApplicationContext) (interface{}, error) {

	productsCollection, err := product.GetProductCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	productsDBCollection := productsCollection.GetDBCollection()

	productModel, err := product.GetProductModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	productAttributesInfo := productModel.GetAttributesInfo()

	result := make(map[string]interface{})

	models.ApplyFilters(context, productsDBCollection)

	// not allowing to see disabled products if not admin
	if err := api.ValidateAdminRights(context); err != nil {
		productsDBCollection.AddFilter("enabled", "=", true)
	}

	for _, productAttribute := range productAttributesInfo {
		if productAttribute.IsLayered {
			distinctValues, _ := productsDBCollection.Distinct(productAttribute.Attribute)
			result[productAttribute.Attribute] = distinctValues
		}
	}

	return result, nil
}
