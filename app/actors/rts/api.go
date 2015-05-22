package rts

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupAPI setups package related API endpoint routines
func setupAPI() error {
	var err error

	err = api.GetRestService().RegisterAPI("rts/visit", api.ConstRESTOperationCreate, APIRegisterVisit)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/referrers", api.ConstRESTOperationGet, APIGetReferrers)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/visits", api.ConstRESTOperationGet, APIGetVisits)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/visits/detail/:from/:to", api.ConstRESTOperationGet, APIGetVisitsDetails)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/conversion", api.ConstRESTOperationGet, APIGetConversion)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/sales", api.ConstRESTOperationGet, APIGetSales)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/sales/detail/:from/:to", api.ConstRESTOperationGet, APIGetSalesDetails)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/bestsellers", api.ConstRESTOperationGet, APIGetBestsellers)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = api.GetRestService().RegisterAPI("rts/visits/realtime", api.ConstRESTOperationGet, APIGetVisitsRealtime)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// APIRegisterVisit registers request for a statistics
func APIRegisterVisit(context api.InterfaceApplicationContext) (interface{}, error) {
	if httpRequest, ok := context.GetRequest().(*http.Request); ok && httpRequest != nil {
		if httpResponseWriter, ok := context.GetResponseWriter().(http.ResponseWriter); ok && httpResponseWriter != nil {
			xReferrer := utils.InterfaceToString(httpRequest.Header.Get("X-Referer"))

			http.SetCookie(httpResponseWriter, &http.Cookie{Name: "X_Referrer", Value: xReferrer, Path: "/"})

			eventData := map[string]interface{}{"session": context.GetSession(), "context": context}
			env.Event("api.rts.visit", eventData)

			return nil, nil
		}
	}
	return nil, nil
}

// APIGetReferrers returns list of unique referrers were registered
func APIGetReferrers(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]int)
	limit := 20

	for url, count := range referrers {
		result[url] = count

		limit++
		if limit == 20 {
			break
		}
	}

	return result, nil
}

// APIGetVisits returns site visit information for a specified local day
func APIGetVisits(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]interface{})
	timeZone := context.GetRequestArgument("tz")

	// get a hours pasted for local day and count for them and for previous day
	todayTo := time.Now().Truncate(time.Hour).Add(time.Hour)
	todayFrom, _ := utils.ApplyTimeZone(todayTo, timeZone)
	todayHoursPast := todayFrom.Sub(todayFrom.Truncate(time.Hour * 24))

	todayFrom = todayTo.Add(-todayHoursPast)
	yesterdayFrom := todayFrom.AddDate(0, 0, -1)
	weekFrom := yesterdayFrom.AddDate(0, 0, -5)

	// get data for visits
	todayStats, err := GetRangeStats(todayFrom, todayTo)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	todayVisits := todayStats.Visit
	todayTotalVisits := todayStats.TotalVisits

	yesterdayStats, err := GetRangeStats(yesterdayFrom, todayFrom)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	yesterdayVisits := yesterdayStats.Visit
	yesterdayTotalVisits := yesterdayStats.TotalVisits

	weekStats, err := GetRangeStats(weekFrom, yesterdayFrom)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	weekVisits := yesterdayVisits + todayVisits + weekStats.Visit
	weekTotalVisits := yesterdayVisits + todayVisits + weekStats.TotalVisits

	result["total"] = map[string]int{
		"today":     todayTotalVisits,
		"yesterday": yesterdayTotalVisits,
		"week":      weekTotalVisits,
	}
	result["unique"] = map[string]int{
		"today":     todayVisits,
		"yesterday": yesterdayVisits,
		"week":      weekVisits,
	}

	return result, nil
}

// APIGetVisitsDetails returns detailed site visit information for a specified period
//   - period start and end dates should be specified in "from" and "to" attributes in DD-MM-YYY format
func APIGetVisitsDetails(context api.InterfaceApplicationContext) (interface{}, error) {

	// getting initial values
	result := make(map[string]int)
	timeZone := context.GetRequestArgument("tz")
	dateFrom := utils.InterfaceToTime(context.GetRequestArgument("from"))
	dateTo := utils.InterfaceToTime(context.GetRequestArgument("to"))

	// checking if user specified correct from and to dates
	if dateFrom.IsZero() {
		dateFrom = time.Now().Truncate(time.Hour * 24)
	}

	if dateTo.IsZero() {
		dateTo = time.Now().Truncate(time.Hour * 24)
	}

	if dateFrom == dateTo {
		dateTo = dateTo.Add(time.Hour * 24)
	}

	// time zone recognize routines save time difference to show in graph by local time
	hoursOffset := time.Hour * 0

	if timeZone != "" {
		dateFrom, hoursOffset = utils.ApplyTimeZone(dateFrom, timeZone)
		dateTo, _ = utils.ApplyTimeZone(dateTo, timeZone)
	}

	// determining required scope
	delta := dateTo.Sub(dateFrom)

	timeScope := time.Hour
	if delta.Hours() > 48 {
		timeScope = timeScope * 24
	}
	dateFrom = dateFrom.Truncate(time.Hour)
	dateTo = dateTo.Truncate(time.Hour)

	// making database request
	visitorInfoCollection, err := db.GetCollection(ConstCollectionNameRTSVisitors)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	visitorInfoCollection.AddFilter("day", ">=", dateFrom)
	visitorInfoCollection.AddFilter("day", "<", dateTo)
	visitorInfoCollection.AddSort("day", false)

	dbRecords, err := visitorInfoCollection.Load()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// filling requested period
	timeIterator := dateFrom
	for timeIterator.Before(dateTo) {
		result[fmt.Sprint(timeIterator.Add(hoursOffset).Unix())] = 0
		timeIterator = timeIterator.Add(timeScope)
	}

	// grouping database records
	for _, item := range dbRecords {
		timestamp := fmt.Sprint(utils.InterfaceToTime(item["day"]).Truncate(timeScope).Unix())
		visits := utils.InterfaceToInt(item["visitors"])

		if value, present := result[timestamp]; present {
			result[timestamp] = value + visits
		}
	}

	var arrayResult [][]int

	for key, item := range result {
		arrayResult = append(arrayResult, []int{utils.InterfaceToInt(key), item})
	}

	return arrayResult, nil
}

// APIGetConversion returns site conversion information
func APIGetConversion(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]interface{})

	timeZone := context.GetRequestArgument("tz")

	// get a hours pasted for local day and count only for them
	todayTo := time.Now().Truncate(time.Hour).Add(time.Hour)
	todayFrom, _ := utils.ApplyTimeZone(todayTo, timeZone)
	todayHoursPast := todayFrom.Sub(todayFrom.Truncate(time.Hour * 24))
	todayFrom = todayTo.Add(-todayHoursPast)

	visits := 0
	sales := 0
	addToCart := 0

	// Go thrue period and summarise a visits
	for todayFrom.Before(todayTo) {

		if _, ok := statistic[todayFrom.Unix()]; ok {
			visits = visits + statistic[todayFrom.Unix()].Visit
			sales = sales + statistic[todayFrom.Unix()].Sales
			addToCart = addToCart + statistic[todayFrom.Unix()].Cart
		}

		todayFrom = todayFrom.Add(time.Hour)
	}

	result["totalVisitors"] = visits
	result["addedToCart"] = addToCart
	result["reachedCheckout"] = sales
	result["purchased"] = sales

	return result, nil
}

//APIGetSales returns information about sales in the recent period, taking into account time zone
func APIGetSales(context api.InterfaceApplicationContext) (interface{}, error) {

	result := make(map[string]interface{})
	timeZone := context.GetRequestArgument("tz")

	// get a hours pasted for local day and count for them and for previous day
	todayTo := time.Now().Truncate(time.Hour).Add(time.Hour)
	todayFrom, _ := utils.ApplyTimeZone(todayTo, timeZone)
	todayHoursPast := todayFrom.Sub(todayFrom.Truncate(time.Hour * 24))

	todayFrom = todayTo.Add(-todayHoursPast)
	yesterdayFrom := todayFrom.AddDate(0, 0, -1)
	weekFrom := yesterdayFrom.AddDate(0, 0, -5)

	// get data for sales
	todayStats, err := GetRangeStats(todayFrom, todayTo)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	todaySales := todayStats.Sales
	todaySalesAmount := todayStats.SalesAmount

	yesterdayStats, err := GetRangeStats(yesterdayFrom, todayFrom)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	yesterdaySales := yesterdayStats.Sales
	yesterdaySalesAmount := yesterdayStats.SalesAmount

	weekStats, err := GetRangeStats(weekFrom, yesterdayFrom)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}
	weekSales := todaySales + yesterdaySales + weekStats.Sales
	weekSalesAmount := todaySalesAmount + yesterdaySalesAmount + weekStats.SalesAmount

	result["sales"] = map[string]float64{
		"today":     todaySalesAmount,
		"yesterday": yesterdaySalesAmount,
		"week":      weekSalesAmount,
	}
	result["orders"] = map[string]int{
		"today":     todaySales,
		"yesterday": yesterdaySales,
		"week":      weekSales,
	}

	return result, nil
}

// APIGetSalesDetails returns site sales information for a specified period
//   - period start and end dates should be specified in "from" and "to" attributes in DD-MM-YYY format
func APIGetSalesDetails(context api.InterfaceApplicationContext) (interface{}, error) {

	// getting initial values
	result := make(map[string]int)
	timeZone := context.GetRequestArgument("tz")
	dateFrom := utils.InterfaceToTime(context.GetRequestArgument("from"))
	dateTo := utils.InterfaceToTime(context.GetRequestArgument("to"))

	// checking if user specified correct from and to dates
	if dateFrom.IsZero() {
		dateFrom = time.Now().Truncate(time.Hour)
	}

	if dateTo.IsZero() {
		dateTo = time.Now().Truncate(time.Hour)
	}

	if dateFrom == dateTo {
		dateTo = dateTo.Add(time.Hour * 24)
	}

	// time zone recognize routines save time difference to show in graph by local time
	hoursOffset := time.Hour * 0

	if timeZone != "" {
		dateFrom, hoursOffset = utils.ApplyTimeZone(dateFrom, timeZone)
		dateTo, _ = utils.ApplyTimeZone(dateTo, timeZone)
	}

	// determining required scope
	delta := dateTo.Sub(dateFrom)

	timeScope := time.Hour
	if delta.Hours() > 48 {
		timeScope = timeScope * 24
	}
	dateFrom = dateFrom.Truncate(time.Hour)
	dateTo = dateTo.Truncate(time.Hour)

	// set database request settings
	orderCollectionModelT, err := order.GetOrderCollectionModel()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	dbCollection := orderCollectionModelT.GetDBCollection()
	dbCollection.SetResultColumns("_id", "created_at")
	dbCollection.AddSort("created_at", false)
	dbCollection.AddFilter("created_at", ">=", dateFrom)
	dbCollection.AddFilter("created_at", "<=", dateTo)

	// get database records
	dbRecords, err := dbCollection.Load()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// filling requested period
	timeIterator := dateFrom
	for timeIterator.Before(dateTo) {
		result[fmt.Sprint(timeIterator.Add(hoursOffset).Unix())] = 0
		timeIterator = timeIterator.Add(timeScope)
	}

	// grouping database records
	for _, order := range dbRecords {
		timestamp := fmt.Sprint(utils.InterfaceToTime(order["created_at"]).Truncate(timeScope).Unix())

		if _, present := result[timestamp]; present {
			result[timestamp]++
		}
	}

	var arrayResult [][]int

	for key, item := range result {
		arrayResult = append(arrayResult, []int{utils.InterfaceToInt(key), item})
	}

	return arrayResult, nil
}

// APIGetBestsellers returns information on site bestsellers top five existing products
func APIGetBestsellers(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make([]map[string]interface{}, 5)

	salesCollection, err := db.GetCollection(ConstCollectionNameRTSSales)
	if err != nil {
		return result, env.ErrorDispatch(err)
	}
	salesCollection.AddFilter("count", ">", 0)
	salesCollection.AddFilter("range", "=", GetSalesRange())
	salesCollection.AddSort("count", true)
	collectionRecords, err := salesCollection.Load()
	if err != nil {
		return result, env.ErrorDispatch(err)
	}

	topFiveCounter := 0
	for _, item := range collectionRecords {
		productID := utils.InterfaceToString(item["product_id"])

		productInstance, err := product.LoadProductByID(productID)
		if err != nil {
			continue
		}

		mediaPath, err := productInstance.GetMediaPath("image")
		if err != nil {
			continue
		}

		result[topFiveCounter] = make(map[string]interface{})

		result[topFiveCounter]["pid"] = productID
		if productInstance.GetDefaultImage() != "" {
			result[topFiveCounter]["image"] = mediaPath + productInstance.GetDefaultImage()
		}

		result[topFiveCounter]["name"] = productInstance.GetName()
		result[topFiveCounter]["count"] = utils.InterfaceToInt(item["count"])
		topFiveCounter++

		if topFiveCounter == 10 {
			break
		}
	}

	return result, nil
}

// APIGetVisitsRealtime returns real-time information on current visits
func APIGetVisitsRealtime(context api.InterfaceApplicationContext) (interface{}, error) {
	result := make(map[string]interface{})
	ratio := float64(0)

	result["Online"] = len(OnlineSessions)
	if OnlineSessionsMax == 0 || len(OnlineSessions) == 0 {
		ratio = float64(0)
	} else {
		ratio = float64(len(OnlineSessions)) / float64(OnlineSessionsMax)
	}
	result["OnlineRatio"] = utils.Round(ratio, 0.5, 2)

	result["Direct"] = OnlineDirect
	if OnlineDirectMax == 0 || OnlineDirect == 0 {
		ratio = float64(0)
	} else {
		ratio = float64(OnlineDirect) / float64(OnlineDirectMax)
	}
	result["DirectRatio"] = utils.Round(ratio, 0.5, 2)

	result["Search"] = OnlineSearch
	if OnlineSearchMax == 0 || OnlineSearch == 0 {
		ratio = float64(0)
	} else {
		ratio = float64(OnlineSearch) / float64(OnlineSearchMax)
	}
	result["SearchRatio"] = utils.Round(ratio, 0.5, 2)

	result["Site"] = OnlineSite
	if OnlineSiteMax == 0 || OnlineSite == 0 {
		ratio = float64(0)
	} else {
		ratio = float64(OnlineSite) / float64(OnlineSiteMax)
	}
	result["SiteRatio"] = utils.Round(ratio, 0.5, 2)

	return result, nil
}
