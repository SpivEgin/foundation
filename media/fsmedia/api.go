package fsmedia

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/utils"
)

// configures package related API endpoint routines
func setupAPI() error {

	api.GetRestService().GET("media", api.IsAdminHandler(APIGetMediaInfo))

	return nil
}

// APIGetMediaInfo will resize all images if the params of the request contain 'resizeAll' with a value of true
func APIGetMediaInfo(context api.InterfaceApplicationContext) (interface{}, error) {
	// TODO: add example api call or add this to Apiary - jwv

	requestParams := context.GetRequestArguments()
	resizeAll := utils.GetFirstMapValue(requestParams, "resizeAll", "resizeImages", "resizeAllImages")

	if resizeAll != nil && utils.InterfaceToBool(resizeAll) {
		mediaStorage, err := media.GetMediaStorage()
		if err != nil {
			return nil, err
		}

		err = mediaStorage.ResizeAllMediaImages()
		if err != nil {
			return nil, err
		}
	}

	return "ok", nil
}
