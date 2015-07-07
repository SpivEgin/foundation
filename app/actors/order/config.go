package order

import (
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	config := env.GetConfig()

	config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathLastIncrementID,
		Value:       0,
		Type:        env.ConstConfigTypeInteger,
		Editor:      "integer",
		Options:     "",
		Label:       "Last Order Increment ID: ",
		Description: "Do not change this value unless you know what you doing",
		Image:       "",
	},
		func(value interface{}) (interface{}, error) {
			return utils.InterfaceToInt(value), nil
		})

	lastIncrementID = utils.InterfaceToInt(config.GetValue(ConstConfigPathLastIncrementID))

	err := config.RegisterItem(env.StructConfigItem{
		Path: ConstConfigPathSubscriptionConfirmationEmail,
		Value: `Dear {{.Visitor.name}}
<br />
<br />
Please confirm placing of new order.
<br />
<h3>Duplicated order #{{.Order.increment_id}}: </h3><br />
Order summary<br />
Subtotal: ${{.Order.subtotal}}<br />
Tax: ${{.Order.tax_amount}}<br />
Shipping: ${{.Order.shipping_amount}}<br />
Total: ${{.Order.grand_total}}<br />
<br />
<a href={{.Info.Link}}>Submit</a><br />`,
		Type:        env.ConstConfigTypeText,
		Editor:      "multiline_text",
		Options:     "",
		Label:       "Duplicate order confirmation e-mail: ",
		Description: "contents of email will be sent to customer to confirm duplication",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}
