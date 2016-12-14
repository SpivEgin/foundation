package magento

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mrjones/oauth"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/utils"
	"github.com/ottemo/foundation/app/models/visitor"
	"time"
)
var siteApiRestUrl string
var httpClient *http.Client

// setups package related API endpoint routines
func setupAPI() error {

	service := api.GetRestService()

	service.GET("impex/magento", restMagento)
	service.POST("impex/magento", restMagentoImport)

	return nil
}

// WEB REST API used to list available models for Impex system
func restMagento(context api.InterfaceApplicationContext) (interface{}, error) {
	//var result []string

	var baseURL = utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathDashboardURL))

	var consumerKey = "2f026794738f4dcf5c2519da1025a085"
	var consumerSecret = "15d7850d214561e7f2431d9330c9171c"
	var siteUrl = "http://ee.test.taa.speroteck-dev.com"
	var siteAdminUrl = "http://ee.test.taa.speroteck-dev.com/admin"

	//return nil, env.ErrorNew(ConstErrorModule, ConstErrorLevel, "6f5a582f-46f1-4b46-834d-e6b39da1ca68", "no csv file was attached")

	if baseURL == "" {
		baseURL = "http://localhost:9000"
	}

	c := oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   siteUrl + "/oauth/initiate",
			AuthorizeTokenUrl: siteAdminUrl + "/oauth_authorize",
			AccessTokenUrl:    siteUrl + "/oauth/token",
		},
	)
	c.Debug(true)
	token, requestUrl, err := c.GetRequestTokenAndUrl(baseURL + "/impex")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println(token)

	context.GetSession().Set(ConstSessionKeyMagentoRequestToken, token.Token)
	context.GetSession().Set(ConstSessionKeyMagentoRequestSecret, token.Secret)
	context.GetSession().Set(ConstSessionKeyMagentoConsumerKey, consumerKey)
	context.GetSession().Set(ConstSessionKeyMagentoConsumerSecret, consumerSecret)
	context.GetSession().Set(ConstSessionKeyMagentoSiteAdminUrl, siteAdminUrl)
	context.GetSession().Set(ConstSessionKeyMagentoSiteUrl, siteUrl)

	return api.StructRestRedirect{
		Result:   "redirect",
		Location: requestUrl,
		DoRedirect: false,
	}, nil
}


func restMagentoImport(context api.InterfaceApplicationContext) (interface{}, error) {
 	var result interface{}
	var baseURL = utils.InterfaceToString(env.ConfigGetValue(app.ConstConfigPathDashboardURL))

	if baseURL == "" {
		baseURL = "http://localhost:9000"
	}

	// check request context
	//---------------------
	requestData, err := api.GetRequestContentAsMap(context)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}


	oauthToken := utils.InterfaceToString(requestData["oauthToken"])
	if oauthToken == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6372b9a3-29f3-4ea4-a19f-40051a8f330b", "email was not specified")
	}

	oauthVerifier := utils.InterfaceToString(requestData["oauthVerifier"])
	if oauthVerifier == "" {
		return nil, env.ErrorNew(ConstErrorModule, env.ConstErrorLevelAPI, "6372b9a3-29f3-4ea4-a19f-40051a8f330b", "email was not specified")
	}

	requestToken := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoRequestToken))
	requestSecret := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoRequestSecret))
	consumerKey := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoConsumerKey))
	consumerSecret := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoConsumerSecret))
	siteAdminUrl := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoSiteAdminUrl))
	siteUrl := utils.InterfaceToString(context.GetSession().Get(ConstSessionKeyMagentoSiteUrl))
	siteApiRestUrl = siteUrl + "/api/rest"

	c := oauth.NewConsumer(
		consumerKey,
		consumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   siteUrl + "/oauth/initiate",
			AuthorizeTokenUrl: siteAdminUrl + "/oauth_authorize",
			AccessTokenUrl:    siteUrl + "/oauth/token",
		},
	)
	c.Debug(true)

	oauthRequestToken := &oauth.RequestToken{
		Secret: requestSecret,
		Token: requestToken,
	}

	accessToken, err := c.AuthorizeToken(oauthRequestToken, oauthVerifier)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	client, err := c.MakeHttpClient(accessToken)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	httpClient = client
	//_, err = getApiData("/customers/27/addresses")
	//if err != nil {
	//	return nil, err
	//}
	customersByte, err := getApiData("/customers")
	if err != nil {
		return nil, err
	}
	result = saveCustomersData(customersByte)

	productsByte, err := getApiData("/products")
	if err != nil {
		return nil, err
	}
	result = saveProductsData(productsByte)

	//data, err := getApiData("/orders")
	//if err != nil {
	//	return nil, err
	//}
	//result = saveOrdersData(data)

	return result, nil
}

func getApiData(apiUrl string) ([]byte, error) {
	request, err := http.NewRequest("GET", siteApiRestUrl + apiUrl, nil)
	if err != nil {
		return make([]byte, 0), err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "*/*")

	response, err := httpClient.Do(request)
	//response, err := client.Get(siteApiRestUrl + "/customers")
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return make([]byte, 0), err
	}

	//data, err := utils.DecodeJSONToStringKeyMap(body)
	//if len(data) > 0 {
	//	data, _ = utils.InterfaceToArray(v)
	//}
	//if err != nil {
	//	data, err = utils.DecodeJSONToArray(body)
	//	if err != nil {
	//		fmt.Println(err)
	//		return make(map[string]interface{}, 0), err
	//	}
	//}
	//
	//fmt.Println(data)

	return body, nil
}

func saveCustomersData(dataByte []byte) (bool)  {
	data, err := utils.DecodeJSONToStringKeyMap(dataByte)
	if err != nil {
		return false
	}

	for _, value := range data {

		visitorModel, err := visitor.GetVisitorModel()
		if err != nil {
			fmt.Println(err)
			return false
		}

		v := utils.InterfaceToMap(value)

		//// visitor map with info
		visitorRecord := map[string]interface{}{
			"email":       utils.InterfaceToString(v["email"]),
			"first_name":  utils.InterfaceToString(v["first_name"]),
			"last_name":   utils.InterfaceToString(v["last_name"]),
			"is_admin":    false,
			"password":    "test",
			"created_at":  time.Now(),
		}

		visitorModel.FromHashMap(visitorRecord)

		err = visitorModel.Save()
		if err != nil {
			fmt.Println(err)
			return false
		}

		if visitorModel.GetID() != "" {
			addCustomerAddresses(utils.InterfaceToString(v["entity_id"]), visitorModel)
		}
	}

	return true
}

func saveProductsData(dataByte []byte) (bool)  {
	data, err := utils.DecodeJSONToStringKeyMap(dataByte)
	if err != nil {
		return false
	}
fmt.Println(data)
	//for _, value := range data {
	//
	//	productModel, err := product.GetProductModel()
	//	if err != nil {
	//		fmt.Println(err)
	//		return false
	//	}
	//
	//	v := utils.InterfaceToMap(value)
	//
	//	//// visitor map with info
	//	visitorRecord := map[string]interface{}{
	//		"email":       utils.InterfaceToString(v["email"]),
	//		"first_name":  utils.InterfaceToString(v["first_name"]),
	//		"last_name":   utils.InterfaceToString(v["last_name"]),
	//		"is_admin":    false,
	//		"password":    "test",
	//		"created_at":  time.Now(),
	//	}
	//
	//	productModel.FromHashMap(visitorRecord)
	//
	//	err = productModel.Save()
	//	if err != nil {
	//		fmt.Println(err)
	//		return false
	//	}
	//
	//	// todo save products images
	//	//if productModel.GetID() != "" {
	//	//	addCustomerAddresses(utils.InterfaceToString(v["entity_id"]), productModel)
	//	//}
	//}

	return true
}

func saveOrdersData(data []byte) (bool)  {
	//fmt.Println("\n\n\n\n\n")
	//fmt.Println(data)
	//fmt.Println("\n\n\n\n\n")
	//for _, value := range data {
	//	v := utils.InterfaceToMap(value)
	//
	//	//// visitor map with info
	//	visitorRecord := map[string]interface{}{
	//		"email":       utils.InterfaceToString(v["email"]),
	//		"first_name":  utils.InterfaceToString(v["first_name"]),
	//		"last_name":   utils.InterfaceToString(v["last_name"]),
	//		"is_admin":    false,
	//		"password":    "test",
	//		"created_at":  time.Now(),
	//	}
	//	visitorModel, err := saveCustomer(visitorRecord)
	//	if err != nil {
	//		fmt.Println(err)
	//		return false
	//	}
	//	addCustomerAddresses(utils.InterfaceToString(v["entity_id"]), visitorModel)
	//	fmt.Println(visitorModel)
	//	fmt.Println(visitorModel.GetID())
	//}

	return true
}

func addCustomerAddresses(customerId string, visitorModel visitor.InterfaceVisitor) (bool) {

	fmt.Println("/customers/" + customerId + "/addresses")
	fmt.Println(visitorModel)
	fmt.Println(visitorModel.GetID())

	body, _ := getApiData("/customers/" + customerId + "/addresses")
	addresses, err := utils.DecodeJSONToArray(body)
	if err != nil {
		return false
	}

	for _, addresse := range addresses {
		visitorAddressModel, err := visitor.GetVisitorAddressModel()
		if err != nil {
			return false
		}

		addresseMap := utils.InterfaceToMap(addresse)

		// todo set default address
		// visitor addresse map with info
		addresseRecord := map[string]interface{}{
			"visitor_id":       visitorModel.GetID(),
			"first_name":  utils.InterfaceToString(addresseMap["firstname"]),
			"last_name":   utils.InterfaceToString(addresseMap["lastname"]),
			"city":    utils.InterfaceToString(addresseMap["city"]),
			"zip_code":    utils.InterfaceToString(addresseMap["postcode"]),
			"company":    utils.InterfaceToString(addresseMap["company"]),
			"phone":    utils.InterfaceToString(addresseMap["telephone"]),
			"country":    utils.InterfaceToString(addresseMap["country_id"]),
			"address_line1":    utils.InterfaceToArray(addresseMap["street"])[0],
			"address_line2":    "",
		}

		visitorAddressModel.FromHashMap(addresseRecord)
		err = visitorAddressModel.Save()
		if err != nil {
			fmt.Println(err)
			return false
		}

	}
	return true

}

