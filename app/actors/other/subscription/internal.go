package subscription

import (
	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// sendConfirmationEmail used to send confirmation emails about subscription change status or to proceed checkout
func sendConfirmationEmail(subscriptionRecord map[string]interface{}, storefrontConfirmationLink, emailTemplate, emailSubject string) error {

	visitorMap := make(map[string]interface{})
	templateMap := make(map[string]interface{})

	customInfo := map[string]interface{}{
		"link": storefrontConfirmationLink,
	}

	templateMap["Info"] = customInfo

	if value, present := subscriptionRecord["order_id"]; present {
		orderID := utils.InterfaceToString(value)

		orderModel, err := order.LoadOrderByID(orderID)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		orderMap := orderModel.ToHashMap()

		var orderItems []map[string]interface{}

		for _, item := range orderModel.GetItems() {
			options := make(map[string]interface{})

			for _, optionKeys := range item.GetOptions() {
				optionMap := utils.InterfaceToMap(optionKeys)
				options[utils.InterfaceToString(optionMap["label"])] = optionMap["value"]
			}
			orderItems = append(orderItems, map[string]interface{}{
				"name":    item.GetName(),
				"options": options,
				"sku":     item.GetSku(),
				"qty":     item.GetQty(),
				"price":   item.GetPrice()})
		}

		orderMap["items"] = orderItems

		templateMap["Order"] = orderMap

		visitorMap = map[string]interface{}{
			"name":  orderModel.Get("customer_name"),
			"email": orderModel.Get("customer_email"),
		}

		templateMap["Visitor"] = visitorMap

	} else {
		visitorID := utils.InterfaceToString(subscriptionRecord["visitor_id"])

		visitorModel, err := visitor.LoadVisitorByID(visitorID)
		if err != nil {
			return env.ErrorDispatch(err)
		}

		visitorMap = map[string]interface{}{
			"name":  visitorModel.GetFullName(),
			"email": visitorModel.GetEmail(),
		}

		templateMap["Visitor"] = visitorMap
	}

	confirmationEmail, err := utils.TextTemplate(emailTemplate, templateMap)

	err = app.SendMail(utils.InterfaceToString(visitorMap["email"]), emailSubject, confirmationEmail)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}