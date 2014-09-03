package address

import (
	"errors"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/visitor"
)

// REST API registration function
func setupAPI() error {
	err := api.GetRestService().RegisterAPI("visitor/address", "POST", "create", restCreateVisitorAddress)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("visitor/address", "PUT", "update/:id", restUpdateVisitorAddress)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("visitor/address", "DELETE", "delete/:id", restDeleteVisitorAddress)
	if err != nil {
		return err
	}

	err = api.GetRestService().RegisterAPI("visitor/address", "GET", "attribute/list", restListVisitorAddressAttributes)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("visitor/address", "GET", "list", restListVisitorAddress)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("visitor/address", "GET", "list/:visitorId", restListVisitorAddress)
	if err != nil {
		return err
	}
	err = api.GetRestService().RegisterAPI("visitor/address", "GET", "load/:id", restGetVisitorAddress)
	if err != nil {
		return err
	}

	return nil
}

// WEB REST API used to create new visitor address
//   - visitor address attributes must be included in POST form
//   - visitor id required
func restCreateVisitorAddress(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//---------------------
	reqData, err := api.GetRequestContentAsMap(params)
	if err != nil {
		return nil, err
	}

	if _, ok := reqData["visitor_id"]; !ok {
		return nil, errors.New("visitor id was not specified")
	}

	// create visitor address operation
	//---------------------------------
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, err
	}

	for attribute, value := range reqData {
		err := visitorAddressModel.Set(attribute, value)
		if err != nil {
			return nil, err
		}
	}

	err = visitorAddressModel.Save()
	if err != nil {
		return nil, err
	}

	return visitorAddressModel.ToHashMap(), nil
}

// WEB REST API used to update existing visitor address
//   - visitor address id must be specified in request URI
//   - visitor address attributes must be included in POST form
func restUpdateVisitorAddress(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//---------------------
	addressId, isSpecifiedId := params.RequestURLParams["id"]
	if !isSpecifiedId {
		return nil, errors.New("visitor address 'id' was not specified")
	}

	reqData, err := api.GetRequestContentAsMap(params)
	if err != nil {
		return nil, err
	}

	// update operation
	//-----------------
	visitorAddressModel, err := visitor.LoadVisitorAddressById(addressId)
	if err != nil {
		return nil, err
	}

	for attribute, value := range reqData {
		err := visitorAddressModel.Set(attribute, value)
		if err != nil {
			return nil, err
		}
	}

	err = visitorAddressModel.Save()
	if err != nil {
		return nil, err
	}

	return visitorAddressModel.ToHashMap(), nil
}

// WEB REST API used to delete visitor address
//   - visitor address attributes must be included in POST form
func restDeleteVisitorAddress(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//--------------------
	addressId, isSpecifiedId := params.RequestURLParams["id"]
	if !isSpecifiedId {
		return nil, errors.New("visitor address id was not specified")
	}

	// delete operation
	//-----------------
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, err
	}

	err = visitorAddressModel.Delete(addressId)
	if err != nil {
		return nil, err
	}

	return "ok", nil
}

// WEB REST API function used to obtain visitor address attributes information
func restListVisitorAddressAttributes(params *api.T_APIHandlerParams) (interface{}, error) {
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, err
	}

	attrInfo := visitorAddressModel.GetAttributesInfo()
	return attrInfo, nil
}

// WEB REST API function used to obtain visitor addresses list
//   - visitor id must be specified in request URI
func restListVisitorAddress(params *api.T_APIHandlerParams) (interface{}, error) {

	// check request params
	//---------------------
	visitorId, isSpecifiedId := params.RequestURLParams["visitorId"]
	if !isSpecifiedId {

		sessionValue := params.Session.Get("visitor_id")
		sessionVisitorId, ok := sessionValue.(string)
		if !ok {
			return nil, errors.New("you are not logined in")
		}
		visitorId = sessionVisitorId
	}

	// list operation
	//---------------
	visitorAddressModel, err := visitor.GetVisitorAddressModel()
	if err != nil {
		return nil, err
	}

	visitorAddressModel.ListFilterAdd("visitor_id", "=", visitorId)

	return visitorAddressModel.List()
}

// WEB REST API used to get visitor address object
//   - visitor address id must be specified in request URI
func restGetVisitorAddress(params *api.T_APIHandlerParams) (interface{}, error) {
	visitorAddressId, isSpecifiedId := params.RequestURLParams["id"]
	if !isSpecifiedId {
		return nil, errors.New("visitor 'id' was not specified")
	}

	visitorAddressModel, err := visitor.LoadVisitorAddressById(visitorAddressId)
	if err != nil {
		return nil, err
	}

	return visitorAddressModel.ToHashMap(), nil
}