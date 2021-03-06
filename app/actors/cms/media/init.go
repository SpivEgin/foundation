package media

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/media"
	"github.com/ottemo/foundation/utils"
)

// init makes package self-initialization routine
func init() {
	api.RegisterOnRestServiceStart(setupAPI)
	app.OnAppStart(onAppStart)

	if err := utils.RegisterTemplateFunction("media", mediaTemplateDirective); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "706a59bb-cfdd-4e26-b8f1-42444daa3170", err.Error())
	}
}

func onAppStart() error {
	mediaStorageInstance, err := media.GetMediaStorage()
	if err != nil {
		return env.ErrorDispatch(err)
	}

	mediaStorage = mediaStorageInstance
	return nil
}

// mediaTemplateDirective - for adding image to pages
//   - use {{media "mediaName" .}} to fetch image URL
//   - Currently this method supports only images, but it is not used anywhere
func mediaTemplateDirective(args ...interface{}) (string, error) {
	mediaName := ""
	if len(args) > 0 {
		mediaName = utils.InterfaceToString(args[0])
	}
	imagePath, err := mediaStorage.GetMediaPath(ConstStorageModel, ConstStorageObject, media.ConstMediaTypeImage)
	if err != nil {
		return "", env.ErrorDispatch(err)
	}

	foundationURL := app.GetFoundationURL(imagePath + mediaName)

	return "<img src=\"" + foundationURL + "\" />", nil
}
