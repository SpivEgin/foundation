package friendmail

import (
	"bytes"
	"encoding/base64"
	"github.com/dchest/captcha"
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"time"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	var err error

	err = api.GetRestService().RegisterAPI("friend/email", api.ConstRESTOperationCreate, APIFriendEmail)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("friend/captcha", api.ConstRESTOperationGet, APIFriendCaptcha)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// APIFriendCaptcha generates captcha for a email form
func APIFriendCaptcha(context api.InterfaceApplicationContext) (interface{}, error) {

	var captchaDigits []byte
	var captchaValue string
	var result interface{}

	captchaValuesMutex.Lock()
	if len(captchaValues) < ConstMaxCaptchaItems {
		captchaDigits = captcha.RandomDigits(captcha.DefaultLen)
		for i := range captchaDigits {
			captchaValue += string(captchaDigits[i] + '0')
		}
	} else {
		for key, _ := range captchaValues {
			captchaValue = key
			break
		}
		captchaDigits = make([]byte, len(captchaValue))
		for idx, digit := range []byte(captchaValue) {
			captchaDigits[idx] = digit - '0'
		}
	}
	captchaValues[captchaValue] = time.Now()
	captchaValuesMutex.Unlock()

	image := captcha.NewImage(captchaValue, captchaDigits, captcha.StdWidth, captcha.StdHeight)
	if image == nil {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "7224c8fe-6079-4bb3-9dc3-ad0847db8e29", "captcha image generation error")
	}

	if context.GetRequestArguments()["json"] != "" {
		buffer := new(bytes.Buffer)
		buffer.WriteString("data:image/png;base64,")
		image.WriteTo(base64.NewEncoder(base64.StdEncoding, buffer))

		result = map[string]interface{}{
			"captcha": buffer.String(),
		}
	} else {
		context.SetResponseContentType("image/png")
		image.WriteTo(context.GetResponseWriter())
	}

	return result, nil
}

// APIFriendEmail sends an email to a friend
func APIFriendEmail(context api.InterfaceApplicationContext) (interface{}, error) {

	var email string

	// checking request values
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if !utils.KeysInMapAndNotBlank(requestData, "captcha", "email") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3c3d0918-b951-43b7-943c-54e8571d0c32", "'captcha' and/or 'email' fields are not specified or blank")
	}

	email = utils.InterfaceToString(requestData["email"])
	if !utils.ValidEmailAddress(email) {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "5821734e-4c84-449b-9f75-fd1154623c42", "invalid email")
	}

	// checking captcha
	captchaValue := utils.InterfaceToString(requestData["captcha"])

	captchaValuesMutex.Lock()
	if _, present := captchaValues[captchaValue]; !present {
		captchaValuesMutex.Unlock()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "8bd3ad79-e464-4355-8a13-27ff55980fbb", "invalid captcha value")
	}
	captchaStore.Get(captchaValue, true)
	delete(captchaValues, captchaValue)
	captchaValuesMutex.Unlock()

	// sending an e-mail
	emailSubject := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathEmailSubject))

	emailTemplate := utils.InterfaceToString(env.ConfigGetValue(ConstConfigPathEmailTemplate))
	emailTemplate, err = utils.TextTemplate(emailTemplate, requestData)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	app.SendMail(email, emailSubject, emailTemplate)

	// storing data to database
	saveData := map[string]interface{}{
		"date":  time.Now(),
		"email": email,
		"data":  requestData,
	}

	dbCollection, err := db.GetCollection(ConstCollectionNameFriendMail)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	_, err = dbCollection.Save(saveData)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	return "ok", nil
}
