package friendmail

import (
	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
	"github.com/dchest/captcha"
	"time"
	"github.com/ottemo/foundation/app"
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

	context.SetResponseContentType("image/png")
	image.WriteTo(context.GetResponseWriter())

	/* context.SetResponseContentType("image/png")
	err := captcha.WriteImage(context.GetResponseWriter(), captchaValue, captcha.StdWidth, captcha.StdHeight)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	} */

	return []byte{}, nil
}

// APIFriendEmail sends an email to a friend
func APIFriendEmail(context api.InterfaceApplicationContext) (interface{}, error) {

	var email string
	var name string

	// checking request values
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	if !utils.KeysInMapAndNotBlank(requestData, "captcha", "name", "email") {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3c3d0918-b951-43b7-943c-54e8571d0c32", "'captcha', 'name' and/or 'email' fields are not specified or blank")
	}

	email = utils.InterfaceToString(requestData["email"])
	name = utils.InterfaceToString(requestData["name"])
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
	emailSubject := utils.InterfaceToString( env.ConfigGetValue(ConstConfigPathEmailSubject) )

	emailTemplate := utils.InterfaceToString( env.ConfigGetValue(ConstConfigPathEmailTemplate) )
	emailTemplate, err = utils.TextTemplate(emailTemplate, requestData)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	app.SendMail(email, emailSubject, emailTemplate)

	// storing data to database
	saveData := map[string]interface{} {
		"date": time.Now(),
		"email": email,
		"name": name,
		"data": requestData,
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