package coupon

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/impex"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	// cart endpoints
	service.POST("cart/coupons", Apply)
	service.DELETE("cart/coupons/:code", Remove)

	// Admin Only
	service.GET("coupons", api.IsAdminHandler(List))
	service.POST("coupons", api.IsAdminHandler(Create))
	service.GET("csv/coupons", api.IsAdminHandler(DownloadCSV))
	service.POST("csv/coupons", api.IsAdminHandler(
		impex.ImportStartHandler(
			api.AsyncHandler(UploadCSV, impex.ImportResultHandler))))
	service.GET("coupons/:id", api.IsAdminHandler(GetByID))
	service.PUT("coupons/:id", api.IsAdminHandler(UpdateByID))
	service.DELETE("coupons/:id", api.IsAdminHandler(DeleteByID))

	return nil
}

// List returns a list registered coupons and is an protected resource that requires authentication to access.
func List(context api.InterfaceApplicationContext) (interface{}, error) {

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	records, err := collection.Load()

	return records, nil
}

// Create will generate a new coupon code when supplied the following required keys,
// they are not required to match.
//   * "name" is the desired reference key for the coupon
//   * "code" is the text visitors must enter to apply a coupon in checkout
func Create(context api.InterfaceApplicationContext) (interface{}, error) {

	// checking request context
	postValues, err := api.GetRequestContentAsMap(context)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	if !utils.KeysInMapAndNotBlank(postValues, "code", "name") {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "842d3ba9-3354-4470-a85f-cbaf909c3827", "Required fields, 'code' and 'name', cannot be blank.")
	}

	valueCode := utils.InterfaceToString(postValues["code"])
	valueName := utils.InterfaceToString(postValues["name"])

	valueUntil := time.Now()
	if value, present := postValues["until"]; present {
		valueUntil = utils.InterfaceToTime(value)
	}

	valueSince := time.Now()
	if value, present := postValues["since"]; present {
		valueSince = utils.InterfaceToTime(value)
	}

	valueLimits := make(map[string]interface{})
	if value, present := postValues["limits"]; present {
		valueLimits = utils.InterfaceToMap(value)
	}

	valueTarget := checkout.ConstDiscountObjectCart
	if targetValue, present := postValues["target"]; present {
		target := strings.ToLower(utils.InterfaceToString(targetValue))
		if target != "" && !strings.Contains(target, checkout.ConstDiscountObjectCart) {
			valueTarget = target
		}
	}

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	if err := collection.AddFilter("code", "=", valueCode); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "099748a0-4348-4d5c-93ab-4f2859f04f0c", err.Error())
	}
	recordsNumber, err := collection.Count()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}
	if recordsNumber > 0 {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "34cb6cfe-fba3-4c1f-afc5-1ff7266a9a86", "A Discount with the provided code: '"+valueCode+"', already exists.")
	}

	// making new record and storing it
	//---------------------------------
	newRecord := map[string]interface{}{
		"code":    valueCode,
		"name":    valueName,
		"amount":  0,
		"percent": 0,
		"times":   -1,
		"since":   valueSince,
		"until":   valueUntil,
		"limits":  valueLimits,
		"target":  valueTarget,
	}

	attributes := []string{"amount", "percent", "times"}
	for _, attribute := range attributes {
		if value, present := postValues[attribute]; present {
			newRecord[attribute] = value
		}
	}

	newID, err := collection.Save(newRecord)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	newRecord["_id"] = newID

	return newRecord, nil
}

// Apply will add the coupon code to the current checkout
//   - coupon code should be specified in "coupon" argument
func Apply(context api.InterfaceApplicationContext) (interface{}, error) {

	var couponCode string
	var present bool

	// check request context
	postValues, err := api.GetRequestContentAsMap(context)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// validate presence of code in post
	if _, present = postValues["code"]; present {
		couponCode = utils.InterfaceToString(postValues["code"])
	} else {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "085b8e25-7939-4b94-93f1-1007ada357d4", "Required key 'code' cannot have a blank value.")
	}

	currentSession := context.GetSession()

	// get applied coupons array for current cart
	currentRedemptions := utils.InterfaceToStringArray(currentSession.Get(ConstSessionKeyCurrentRedemptions))

	// check if coupon has already been applied
	if utils.IsInArray(couponCode, currentRedemptions) {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "29c4c963-0940-4780-8ad2-9ed5ca7c97ff", "Coupon code, "+couponCode+" has already been applied in this cart.")
	}

	// load coupon for specified code
	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}
	err = collection.AddFilter("code", "=", couponCode)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	records, err := collection.Load()
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// verify and apply obtained coupon
	if len(records) > 0 {
		discountCoupon := records[0]

		applyTimes := utils.InterfaceToInt(discountCoupon["times"])

		validStart := isValidStart(discountCoupon["since"])
		validEnd := isValidEnd(discountCoupon["until"])

		// check if subtotal is more then required by the discount
		currentCheckout, err := checkout.GetCurrentCheckout(context, true)
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		limits := utils.InterfaceToMap(discountCoupon["limits"])
		minimumCartAmount := utils.InterfaceToFloat64(limits["minimum_cart_amount"])

		if minimumCartAmount > currentCheckout.GetSubtotal() {

			context.SetResponseStatusBadRequest()
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "023c3e22-aff2-40eb-b75c-60834f49b951", "minium purchase amount of $"+fmt.Sprintf("%.2f", minimumCartAmount)+" not met.")
		}

		// to be applicable, the coupon should satisfy following conditions:
		//   [applyTimes] should be -1 or >0 and [workSince] >= currentTime <= [workUntil] if set
		if (applyTimes == -1 || applyTimes > 0) && validStart && validEnd {

			// TODO: applied coupons are lost with session clear, probably should be made on order creation,
			// or add an event handler to add to session # of times used
			if applyTimes > 0 {
				discountCoupon["times"] = applyTimes - 1
				_, err := collection.Save(discountCoupon)
				if err != nil {
					context.SetResponseStatusInternalServerError()
					return nil, env.ErrorDispatch(err)
				}
			}

			// coupon is working - applying it
			currentRedemptions = append(currentRedemptions, couponCode)
			currentSession.Set(ConstSessionKeyCurrentRedemptions, currentRedemptions)

		} else {
			context.SetResponseStatusBadRequest()
			if !validStart {
				context.SetResponseStatusBadRequest()
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "63442858-bd71-4f10-855a-b5975fc2dd16", "Coupon code, "+strings.ToUpper(couponCode)+", has an start time outside valid time constraints.")
			} else if !validEnd {
				context.SetResponseStatusInternalServerError()
				return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6f286e8a-e649-4535-ae5a-7b792e1f38fe", "Coupon code, "+strings.ToUpper(couponCode)+", has an end time outside valid time constraints.")
			}
			context.SetResponseStatusInternalServerError()
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3cbc500c-e091-4b61-bc95-d34e3b9422aa", "Coupon code, "+strings.ToUpper(couponCode)+", cannot be applied, exceeded usage limits.")
		}
	} else {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "b2934505-06e9-4250-bb98-c22e4918799e", "Coupon code, "+strings.ToUpper(couponCode)+", is not a valid coupon code.")
	}

	return "Coupon applied", nil
}

// Remove will remove the coupon code and its value from the current checkout
//   * "coupon" key refers to the coupon code
//   * use a "*" as the coupon code to revert all discounts
func Remove(context api.InterfaceApplicationContext) (interface{}, error) {

	couponCode := context.GetRequestArgument("code")

	currentRedemptions := utils.InterfaceToStringArray(context.GetSession().Get(ConstSessionKeyCurrentRedemptions))
	if !utils.IsInArray(couponCode, currentRedemptions) {
		return "ok", nil
	}

	if len(currentRedemptions) > 0 {
		var newAppliedCoupons []string
		for _, value := range currentRedemptions {
			if value != couponCode {
				newAppliedCoupons = append(newAppliedCoupons, value)
			}
		}
		context.GetSession().Set(ConstSessionKeyCurrentRedemptions, newAppliedCoupons)

		// times used increase
		collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
		if err != nil {
			context.SetResponseStatusInternalServerError()
			return nil, env.ErrorDispatch(err)
		}
		err = collection.AddFilter("code", "=", couponCode)
		if err != nil {
			context.SetResponseStatusInternalServerError()
			return nil, env.ErrorDispatch(err)
		}
		records, err := collection.Load()
		if err != nil {
			context.SetResponseStatusInternalServerError()
			return nil, env.ErrorDispatch(err)
		}
		if len(records) > 0 {
			applyTimes := utils.InterfaceToInt(records[0]["times"])
			if applyTimes >= 0 {
				records[0]["times"] = applyTimes + 1

				_, err := collection.Save(records[0])
				if err != nil {
					context.SetResponseStatusInternalServerError()
					return nil, env.ErrorDispatch(err)
				}
			}
		}
	}

	return "Removed successful", nil
}

// DownloadCSV returns a csv file with the current coupons and their configuration
//   * returns a csv file
func DownloadCSV(context api.InterfaceApplicationContext) (interface{}, error) {

	// preparing csv writer
	csvWriter := csv.NewWriter(context.GetResponseWriter())

	if err := context.SetResponseContentType("text/csv"); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "e6a797ba-db60-41fa-a5ef-876f77d7b7a1", err.Error())
	}
	if err := context.SetResponseSetting("Content-disposition", "attachment;filename=discount_coupons.csv"); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "54f5243c-fd94-4462-b99b-d1d23e0bc401", err.Error())
	}

	if err := csvWriter.Write([]string{"Code", "Name", "Amount", "Percent", "Times", "Since", "Until", "Limits", "Target"}); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "8dc86f61-f256-4fe3-85a1-3522edd43c12", err.Error())
	}

	csvWriter.Flush()

	// loading records from DB and writing them in csv format
	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	err = collection.Iterate(func(record map[string]interface{}) bool {
		err := csvWriter.Write([]string{
			utils.InterfaceToString(record["code"]),
			utils.InterfaceToString(record["name"]),
			utils.InterfaceToString(record["amount"]),
			utils.InterfaceToString(record["percent"]),
			utils.InterfaceToString(record["times"]),
			utils.InterfaceToString(record["since"]),
			utils.InterfaceToString(record["until"]),
			utils.InterfaceToString(record["limits"]),
			utils.InterfaceToString(record["target"]),
		})
		if err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "0f6db32f-a0c9-4cf0-9c61-e3b9ba679af4", err.Error())
		}

		csvWriter.Flush()

		return true
	})

	return "Download Complete", nil
}

// UploadCSV will overwrite and replace the current coupon configuration with the uploaded CSV
//   NOTE: the csv file should be provided in a "file" field when sent as a multipart form
func UploadCSV(context api.InterfaceApplicationContext) (interface{}, error) {

	csvFileName := context.GetRequestArgument("file")
	if csvFileName == "" {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "3398f40a-726b-48ad-9f29-9dd390b7e952", "A file name must be specified.")
	}

	csvFile := context.GetRequestFile(csvFileName)
	if csvFile == nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6b0cf271-ce1c-43ae-8f18-261120972bd0", "A file must be specified.")
	}

	csvReader := csv.NewReader(csvFile)
	csvReader.Comma = ','

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}
	if _, err := collection.Delete(); err != nil {
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "008caf00-e1d8-440f-8682-9299ac9e7d99", err.Error())
	}

	if _, err := csvReader.Read(); err != nil { //skipping header
		return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "43a3cf52-983a-4173-b48d-62af65fef015", err.Error())
	}
	for csvRecord, err := csvReader.Read(); err == nil; csvRecord, err = csvReader.Read() {
		if len(csvRecord) >= 7 {
			record := make(map[string]interface{})

			code := utils.InterfaceToString(csvRecord[0])
			name := utils.InterfaceToString(csvRecord[1])
			if code == "" || name == "" {
				continue
			}

			times := utils.InterfaceToInt(csvRecord[4])
			if csvRecord[4] == "" {
				times = -1
			}

			record["code"] = code
			record["name"] = name
			record["amount"] = utils.InterfaceToFloat64(csvRecord[2])
			record["percent"] = utils.InterfaceToFloat64(csvRecord[3])
			record["times"] = times
			record["since"] = utils.InterfaceToTime(csvRecord[5])
			record["until"] = utils.InterfaceToTime(csvRecord[6])
			record["limits"] = utils.InterfaceToMap(csvRecord[7])
			record["target"] = utils.InterfaceToString(csvRecord[8])

			if _, err := collection.Save(record); err != nil {
				_ = env.ErrorDispatch(err)
				return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "c5307a1b-104c-4fef-ad2c-c5cb118386b5", "Unable to store coupon")
			}
		}
	}

	return "Upload Complete", nil
}

// GetByID returns a coupon with the specified ID
// * coupon id should be specified in the "id" argument
func GetByID(context api.InterfaceApplicationContext) (interface{}, error) {

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	id := context.GetRequestArgument("id")
	records, err := collection.LoadByID(id)

	return records, nil
}

// UpdateByID updates existing coupon specified in the request argument
//   * coupon id should be specified in "couponID" argument
func UpdateByID(context api.InterfaceApplicationContext) (interface{}, error) {

	// check request context
	postValues, err := api.GetRequestContentAsMap(context)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	couponID := context.GetRequestArgument("id")
	record, err := collection.LoadByID(couponID)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	// if discount 'code' was changed - checking new value for duplicates
	if codeValue, present := postValues["code"]; present && codeValue != record["code"] {
		codeValue := utils.InterfaceToString(codeValue)

		if err := collection.AddFilter("code", "=", codeValue); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "1b28bceb-e799-460c-b74f-b5d14511b247", err.Error())
		}
		recordsNumber, err := collection.Count()
		if err != nil {
			context.SetResponseStatusInternalServerError()
			return nil, env.ErrorDispatch(err)
		}
		if recordsNumber > 0 {
			context.SetResponseStatusBadRequest()
			return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "e49e5e01-4f6f-4ff0-bd28-dfb616308aa7", "A Discount with the provided code: '"+codeValue+"', already exists.")
		}

		record["code"] = codeValue
	}

	// updating other attributes
	attributes := []string{"amount", "percent", "times", "limits"}
	for _, attribute := range attributes {
		if value, present := postValues[attribute]; present {
			record[attribute] = value
		}
	}

	record["target"] = checkout.ConstDiscountObjectCart
	if targetValue, present := postValues["target"]; present {
		target := strings.ToLower(utils.InterfaceToString(targetValue))
		if target != "" && !strings.Contains(target, checkout.ConstDiscountObjectCart) {
			record["target"] = target
		}
	}

	if value, present := postValues["until"]; present {
		record["until"] = utils.InterfaceToTime(value)
	}

	if value, present := postValues["since"]; present {
		record["since"] = utils.InterfaceToTime(value)
	}

	// saving updates
	_, err = collection.Save(record)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	return record, nil
}

// DeleteByID deletes specified SEO item
//   * discount id should be specified in the "couponID" argument
func DeleteByID(context api.InterfaceApplicationContext) (interface{}, error) {

	collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
	if err != nil {
		context.SetResponseStatusInternalServerError()
		return nil, env.ErrorDispatch(err)
	}

	err = collection.DeleteByID(context.GetRequestArgument("id"))
	if err != nil {
		context.SetResponseStatusBadRequest()
		return nil, env.ErrorDispatch(err)
	}

	return "Delete Successful", nil
}
