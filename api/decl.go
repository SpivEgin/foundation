// Package api is a set of interfaces representing API endpoint services.
//
// (currently only "InterfaceRestService" endpoint interface supported)
package api

import (
	"net/http"
)

// Package global constants
var (
	ConstSessionKeyAdminRights = "adminRights" // session key used to flag that user have admin rights
)

// structure to hold API request related information
type StructAPIHandlerParams struct {
	ResponseWriter   http.ResponseWriter
	Request          *http.Request
	RequestGETParams map[string]string
	RequestURLParams map[string]string
	RequestContent   interface{}
	Session          InterfaceSession
}

// structure you should return in API handler function if redirect needed
type StructRestRedirect struct {
	Result   interface{}
	Location string

	DoRedirect bool
}

// API handler callback function type
type FuncAPIHandler func(params *StructAPIHandlerParams) (interface{}, error)
