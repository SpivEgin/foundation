package rest

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
)

// GetName returns implementation name of our REST API service
func (it *DefaultRestService) GetName() string {
	return "httprouter"
}

// RegisterAPI is available for modules to call in order to provide their own REST API functionality
func (it *DefaultRestService) RegisterAPI(service string, method string, uri string, handler api.FuncAPIHandler) error {

	// httprouter supposes other format of handler than we use, so we need wrapper
	wrappedHandler := func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {

		// catching API handler fails
		defer func() {
			if recoverResult := recover(); recoverResult != nil {
				env.ErrorNew(ConstErrorModule, ConstErrorLevel, "28d7ef2f-631f-4f38-a916-579bf822908b", "API call fail")
			}
		}()

		// debug log related variables initialization
		var startTime time.Time
		var debugRequestIdentifier string

		if ConstUseDebugLog {
			startTime = time.Now()
			debugRequestIdentifier = startTime.Format("20060102150405")
		}

		// Request URL parameters detection
		//----------------------------------

		// getting URL params of request
		reqArguments := make(map[string]string)
		for _, param := range params {
			reqArguments[param.Key] = param.Value
		}

		// getting params from URL, those after "?"
		reqParameters := make(map[string]string)
		urlParsedParams, err := url.ParseQuery(req.URL.RawQuery)
		if err == nil {
			for key, value := range urlParsedParams {
				reqParameters[key] = value[0]
			}
		}

		// Request content detection
		//----------------------------

		var content interface{}
		contentType := req.Header.Get("Content-Type")

		switch {

		// request contains JSON content
		case strings.Contains(contentType, "json"):
			newContent := map[string]interface{}{}

			buf := make([]byte, req.ContentLength)
			req.Body.Read(buf)
			json.Unmarshal(buf, &newContent)

			content = newContent

		// request contains POST form data
		case strings.Contains(contentType, "form-data"):
			newContent := map[string]interface{}{}

			req.ParseForm()
			for attribute, value := range req.PostForm {
				newContent[attribute], _ = url.QueryUnescape(value[0])
			}

			req.ParseMultipartForm(32 << 20) // 32 MB
			if req.MultipartForm != nil {
				for attribute, value := range req.MultipartForm.Value {
					newContent[attribute], _ = url.QueryUnescape(value[0])
				}
			}

			content = newContent

		// request contains "x-www-form-urlencoded" data
		case strings.Contains(contentType, "urlencode"):
			newContent := map[string]interface{}{}

			req.ParseForm()
			for attribute, value := range req.PostForm {
				newContent[attribute], _ = url.QueryUnescape(value[0])
			}

			content = newContent

		default:
			var body []byte

			if req.ContentLength > 0 {
				body = make([]byte, req.ContentLength)
				req.Body.Read(body)
			} else {
				body, _ = ioutil.ReadAll(req.Body)
			}

			content = string(body)
		}

		// Handling request
		//------------------

		// preparing struct for API handler
		applicationContext := new(RestApplicationContext)
		applicationContext.Request = req
		applicationContext.RequestArguments = reqArguments
		applicationContext.RequestParameters = reqParameters
		applicationContext.RequestContent = content
		applicationContext.RequestFiles = make(map[string]io.Reader)
		applicationContext.ResponseWriter = resp
		applicationContext.ContextValues = make(map[string]interface{})

		// collecting request files
		if req.MultipartForm != nil && req.MultipartForm.File != nil {
			for _, fileInfoArray := range req.MultipartForm.File {
				for _, fileInfo := range fileInfoArray {
					attachedFile, err := fileInfo.Open()
					if err == nil {
						applicationContext.RequestFiles[fileInfo.Filename] = attachedFile
					}
				}
			}
		}

		// starting session for request
		currentSession, err := api.StartSession(applicationContext)
		if err != nil {
			env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c8a3bbf8-215f-4dff-b0e7-3d0d102ad02d", "Session init fail: "+err.Error())
		}
		applicationContext.Session = currentSession

		if ConstUseDebugLog {
			env.Log(ConstDebugLogStorage, "REQUEST_"+debugRequestIdentifier, fmt.Sprintf("%s [%s]\n%#v\n", req.RequestURI, currentSession.GetID(), content))
		}

		// event for request
		eventData := map[string]interface{}{"session": currentSession, "context": applicationContext}
		cookieReferrer, err := req.Cookie("X_Referrer")
		if err != nil {
			eventData["referrer"] = ""
		} else {
			eventData["referrer"] = cookieReferrer.Value
		}
		env.Event("api.request", eventData)

		// API handler processing
		result, err := handler(applicationContext)

		if err == nil {
			applicationContext.Result = result
		}

		// event for response
		eventData["response"] = result
		env.Event("api.response", eventData)
		result = eventData["response"]

		// result conversion before output
		redirectLocation := ""
		if redirect, ok := result.(api.StructRestRedirect); ok {
			if redirect.DoRedirect {
				resp.Header().Add("Location", redirect.Location)
				resp.WriteHeader(301)
				result = []byte("")
			} else {
				redirectLocation = redirect.Location
				result = redirect.Result
			}
		}

		// converting result to []byte if it is not already done
		if _, ok := result.([]byte); !ok {

			// JSON encode
			if resp.Header().Get("Content-Type") == "application/json" {
				var errorMsg map[string]interface{}
				if err != nil {
					if _, ok := err.(env.InterfaceOttemoError); !ok {
						err = env.ErrorDispatch(err)
					}

					if ottemoError, ok := err.(env.InterfaceOttemoError); ok {
						errorMsg = map[string]interface{}{
							"message": ottemoError.Error(),
							"level":   ottemoError.ErrorLevel(),
							"code":    ottemoError.ErrorCode(),
						}
					} else {
						env.ErrorNew(ConstErrorModule, ConstErrorLevel, "bdbb8627-18e8-4969-a048-c8b482235f39", "can't convert error to ottemoError")
						errorMsg = map[string]interface{}{
							"message": err.Error(),
							"level":   env.ConstErrorLevelAPI,
							"code":    "bdbb8627-18e8-4969-a048-c8b482235f39",
						}
					}
				}

				result, _ = json.Marshal(map[string]interface{}{"result": result, "error": errorMsg, "redirect": redirectLocation})
			}

			// XML encode
			if resp.Header().Get("Content-Type") == "text/xml" {
				result, _ = xml.Marshal(result)
			}
		}

		if ConstUseDebugLog {
			responseTime := time.Now().Sub(startTime)
			env.Log(ConstDebugLogStorage, "RESPONSE_"+debugRequestIdentifier, fmt.Sprintf("%s (%dns)\n%s\n", req.RequestURI, responseTime, result))
		}

		resp.Write(result.([]byte))
	}

	// registration of handler within httprouter
	//-------------------------------------------
	path := "/" + service + "/" + uri

	switch method {
	case "GET":
		it.Router.GET(path, wrappedHandler)
	case "PUT":
		it.Router.PUT(path, wrappedHandler)
	case "POST":
		it.Router.POST(path, wrappedHandler)
	case "DELETE":
		it.Router.DELETE(path, wrappedHandler)
	default:
		return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "58228dcc-f5e4-4aae-b6df-9dd55041a21e", "unsupported method '"+method+"'")
	}

	key := path + " {" + method + "}"
	it.Handlers[key] = wrappedHandler

	return nil
}

// ServeHTTP is an entry point for HTTP request, it takes control before request handled
// (go lang "http.server" package "Handler" interface implementation)
func (it DefaultRestService) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	// CORS fix-up
	responseWriter.Header().Set("Access-Control-Allow-Origin", request.Header.Get("Origin"))
	responseWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	responseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
	responseWriter.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie, X-Referer, Content-Length, Accept-Encoding, X-CSRF-Token")

	responseWriter.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
	responseWriter.Header().Set("Pragma", "no-cache") // HTTP 1.0.
	responseWriter.Header().Set("Expires", "0") // Proxies

	if request.Method == "GET" || request.Method == "POST" || request.Method == "PUT" || request.Method == "DELETE" {

		// default output format
		responseWriter.Header().Set("Content-Type", "application/json")

		request.URL.Path = strings.Replace(request.URL.Path, "/foundation", "", -1)

		it.Router.ServeHTTP(responseWriter, request)
	}
}

// Run is the Ottemo REST server startup function, analogous to "ListenAndServe"
func (it *DefaultRestService) Run() error {
	fmt.Println("REST API Service [HTTPRouter] starting to listen on " + it.ListenOn)
	env.ErrorDispatch(http.ListenAndServe(it.ListenOn, it))

	return nil
}
