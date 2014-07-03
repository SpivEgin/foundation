package rest

import (
	"net/http"
	"sort"

	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/api"

	"github.com/julienschmidt/httprouter"
)

func init() {
	instance := new(DefaultRestService)

	api.RegisterRestService(instance)
	env.RegisterOnConfigIniStart( instance.startup )
}

func (it *DefaultRestService) startup() error {

	it.ListenOn = ":3000"
	if iniConfig := env.GetIniConfig(); iniConfig != nil {
		if iniValue := iniConfig.GetValue("rest.listenOn", it.ListenOn); iniValue != "" {
			it.ListenOn = iniValue
		}
	}

	it.Router = httprouter.New()
	it.Router.GET("/",
		func( resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
			newline := []byte( "\n" )

			resp.Header().Add("Content-Type", "text")

			resp.Write( []byte( "Ottemo REST API:" ) )
			resp.Write( newline )

			// sorting handlers
			handlers := make([]string, 0, len(it.Handlers))
			for handlerPath := range it.Handlers {
				handlers = append(handlers, handlerPath)
			}
			sort.Strings(handlers)


			for _, handlerPath := range handlers {
				resp.Write( []byte( handlerPath ) )
				resp.Write( newline )
			}
		})

	it.Handlers = make( map[string]httprouter.Handle )

	api.OnRestServiceStart()

	return nil
}
