package checkout

import (
	"errors"

	"github.com/ottemo/foundation/app"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// SetSession sets visitor for checkout
func (it *DefaultCheckout) SendOrderConfirmationMail() error {

	checkoutOrder := it.GetOrder()
	if checkoutOrder == nil {
		return errors.New("checkout order is not exists")
	}

	confirmationEmail := utils.InterfaceToString(env.ConfigGetValue(checkout.CONFIG_PATH_CONFIRMATION_EMAIL))
	if confirmationEmail != "" {
		email := utils.InterfaceToString(checkoutOrder.Get("customer_email"))
		if email == "" {
			return errors.New("customer email for order is not set")
		}

		confirmationEmail, err := utils.TextTemplate(confirmationEmail,
			map[string]interface{}{
				"Order":   checkoutOrder.ToHashMap(),
				"Visitor": it.GetVisitor().ToHashMap(),
			})
		if err != nil {
			return err
		}

		err = app.SendMail(email, "Order confirmation", confirmationEmail)
		if err != nil {
			return err
		}
	}

	return nil
}