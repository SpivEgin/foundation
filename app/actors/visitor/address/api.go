package address

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/visitor"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	err := api.GetRestService().RegisterAPI("visitor/:visitorID/address", api.ConstRESTOperationCreate, restCreateVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitor/:visitorID/address/update/:addressID", api.ConstRESTOperationUpdate, restUpdateVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitor/:visitorID/address/delete/:addressID", api.ConstRESTOperationDelete, restDeleteVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("visitor/:visitorID/addresses", api.ConstRESTOperationGet, restListVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitor/:visitorID/addresses/count", api.ConstRESTOperationGet, restCountVisitorAddresses)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("visitors/addresses/count", api.ConstRESTOperationGet, restCountVisitorAddresses)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitors/addresses/attributes", api.ConstRESTOperationGet, restListVisitorAddressAttributes)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("visitors/address", api.ConstRESTOperationCreate, restListVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitors/address/:addressID", api.ConstRESTOperationDelete, restDeleteVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	err = api.GetRestService().RegisterAPI("visitors/address/:addressID", api.ConstRESTOperationGet, restGetVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// WEB REST API used to create new visitor address
//   - visitor address attributes must be included in POST form
//   - visitor id required
func restCreateVisitorAddress(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	reqData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if _, ok := reqData["visitor_id"]; !ok {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "a9da4ac4-d073-48f3-b062-2ba536d2c577", "visitor id was not specified")
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		if reqData["visitor_id"] != visitor.GetCurrentVisitorID(context) {
			return nil, env.ErrorDispatch(err)
		}
	}

	// create visitor address operation
	//---------------------------------
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	for attribute, value := range reqData {
		err := visitorAddressModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = visitorAddressModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return visitorAddressModel.ToHashMap(), nil
}

// WEB REST API used to update existing visitor address
//   - visitor address id must be specified in request URI
//   - visitor address attributes must be included in POST form
func restUpdateVisitorAddress(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	addressID := context.GetRequestArgument("addressID")
	if addressID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "fe7814c0-85fe-4d60-a134-415f7ac12075", "visitor address 'id' was not specified")
	}

	reqData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	visitorAddressModel, err := visitor.LoadVisitorAddressByID(addressID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		if visitorAddressModel.GetVisitorID() != visitor.GetCurrentVisitorID(context) {
			return nil, env.ErrorDispatch(err)
		}
	}

	// update operation
	//-----------------
	for attribute, value := range reqData {
		err := visitorAddressModel.Set(attribute, value)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}
	}

	err = visitorAddressModel.Save()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return visitorAddressModel.ToHashMap(), nil
}

// WEB REST API used to delete visitor address
//   - visitor address attributes must be included in POST form
func restDeleteVisitorAddress(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//--------------------
	addressID := context.GetRequestArgument("addressID")
	if addressID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "eec1ef1b-25d9-4dbe-8bd2-b907a0897203", "visitor address id was not specified")
	}

	visitorAddressModel, err := visitor.LoadVisitorAddressByID(addressID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		if visitorAddressModel.GetVisitorID() != visitor.GetCurrentVisitorID(context) {
			return nil, env.ErrorDispatch(err)
		}
	}

	// delete operation
	err = visitorAddressModel.Delete()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}

// WEB REST API function used to obtain visitor address attributes information
func restListVisitorAddressAttributes(context api.InterfaceApplicationContext) (interface{}, error) {

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	attrInfo := visitorAddressModel.GetAttributesInfo()
	return attrInfo, nil
}

// WEB REST API function used to obtain visitors addresses count in model collection
func restCountVisitorAddresses(context api.InterfaceApplicationContext) (interface{}, error) {

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	visitorAddressCollectionModel, err := visitor.GetVisitorAddressCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	dbCollection := visitorAddressCollectionModel.GetDBCollection()

	// filters handle
	api.ApplyFilters(context, dbCollection)

	return dbCollection.Count()
}

// WEB REST API function used to obtain visitor addresses list
//   - visitor id must be specified in request URI
func restListVisitorAddress(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	//---------------------
	reqData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, err
	}

	visitorID := context.GetRequestArgument("visitorID")
	if visitorID == "" {

		sessionVisitorID := visitor.GetCurrentVisitorID(context)
		if sessionVisitorID == "" {
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "2ac4c16b-9241-406e-b35a-399813bb6ca5", "you are not logined in")
		}
		visitorID = sessionVisitorID
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		if visitorID != visitor.GetCurrentVisitorID(context) {
			return nil, env.ErrorDispatch(err)
		}
	}

	// list operation
	//---------------
	visitorAddressCollectionModel, err := visitor.GetVisitorAddressCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	dbCollection := visitorAddressCollectionModel.GetDBCollection()
	dbCollection.AddStaticFilter("visitor_id", "=", visitorID)

	// limit parameter handle
	visitorAddressCollectionModel.ListLimit(api.GetListLimit(context))

	// filters handle
	api.ApplyFilters(context, dbCollection)

	// extra parameter handle
	if extra, isExtra := reqData["extra"]; isExtra {
		extra := utils.Explode(utils.InterfaceToString(extra), ",")
		for _, value := range extra {
			err := visitorAddressCollectionModel.ListAddExtraAttribute(value)
			if err != nil {
				return nil, env.ErrorDispatch(err)
			}
		}
	}

	return visitorAddressCollectionModel.List()
}

// WEB REST API used to get visitor address object
//   - visitor address id must be specified in request URI
func restGetVisitorAddress(context api.InterfaceApplicationContext) (interface{}, error) {
	visitorAddressID := context.GetRequestArgument("addressID")
	if visitorAddressID == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b94882c6-bbdd-428d-88b0-7ea5623d80f7", "visitor 'id' was not specified")
	}

	visitorAddressModel, err := visitor.LoadVisitorAddressByID(visitorAddressID)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// check rights
	if err := api.ValidateAdminRights(context); err != nil {
		if visitorAddressModel.GetVisitorID() != visitor.GetCurrentVisitorID(context) {
			return nil, env.ErrorDispatch(err)
		}
	}

	return visitorAddressModel.ToHashMap(), nil
}
