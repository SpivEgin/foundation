package swatch

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
)

// init makes package self-initialization routine
func init() {
	app.OnAppStart(onAppStart)
	api.RegisterOnRestServiceStart(setupAPI)
}

func onAppStart() error {

	var err error
	mediaStorage, err = media.GetMediaStorage()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}
