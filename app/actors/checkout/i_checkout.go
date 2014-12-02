package checkout

import (
	"fmt"
	"time"

	"github.com/ottemo/foundation/api"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"

	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/app/models/visitor"
)

// SetShippingAddress sets shipping address for checkout
func (it *DefaultCheckout) SetShippingAddress(address visitor.InterfaceVisitorAddress) error {
	it.ShippingAddressID = address.GetID()
	return nil
}

// GetShippingAddress returns checkout shipping address
func (it *DefaultCheckout) GetShippingAddress() visitor.InterfaceVisitorAddress {
	shippingAddress, _ := visitor.LoadVisitorAddressByID(it.ShippingAddressID)
	return shippingAddress
}

// SetBillingAddress sets billing address for checkout
func (it *DefaultCheckout) SetBillingAddress(address visitor.InterfaceVisitorAddress) error {
	it.BillingAddressID = address.GetID()
	return nil
}

// GetBillingAddress returns checkout billing address
func (it *DefaultCheckout) GetBillingAddress() visitor.InterfaceVisitorAddress {
	billingAddress, _ := visitor.LoadVisitorAddressByID(it.BillingAddressID)
	return billingAddress
}

// SetPaymentMethod sets payment method for checkout
func (it *DefaultCheckout) SetPaymentMethod(paymentMethod checkout.InterfacePaymentMethod) error {
	it.PaymentMethodCode = paymentMethod.GetCode()
	return nil
}

// GetPaymentMethod returns checkout payment method
func (it *DefaultCheckout) GetPaymentMethod() checkout.InterfacePaymentMethod {
	if paymentMethods := checkout.GetRegisteredPaymentMethods(); paymentMethods != nil {
		for _, paymentMethod := range paymentMethods {
			if paymentMethod.GetCode() == it.PaymentMethodCode {
				return paymentMethod
			}
		}
	}
	return nil
}

// SetShippingMethod sets payment method for checkout
func (it *DefaultCheckout) SetShippingMethod(shippingMethod checkout.InterfaceShippingMethod) error {
	it.ShippingMethodCode = shippingMethod.GetCode()
	return nil
}

// GetShippingMethod returns a checkout shipping method
func (it *DefaultCheckout) GetShippingMethod() checkout.InterfaceShippingMethod {
	if shippingMethods := checkout.GetRegisteredShippingMethods(); shippingMethods != nil {
		for _, shippingMethod := range shippingMethods {
			if shippingMethod.GetCode() == it.ShippingMethodCode {
				return shippingMethod
			}
		}
	}
	return nil
}

// SetShippingRate sets shipping rate for checkout
func (it *DefaultCheckout) SetShippingRate(shippingRate checkout.StructShippingRate) error {
	it.ShippingRate = shippingRate
	return nil
}

// GetShippingRate returns a checkout shipping rate
func (it *DefaultCheckout) GetShippingRate() *checkout.StructShippingRate {
	return &it.ShippingRate
}

// SetCart sets cart for checkout
func (it *DefaultCheckout) SetCart(checkoutCart cart.InterfaceCart) error {
	it.CartID = checkoutCart.GetID()
	return nil
}

// GetCart returns a shopping cart
func (it *DefaultCheckout) GetCart() cart.InterfaceCart {
	cartInstance, _ := cart.LoadCartByID(it.CartID)
	return cartInstance
}

// SetVisitor sets visitor for checkout
func (it *DefaultCheckout) SetVisitor(checkoutVisitor visitor.InterfaceVisitor) error {
	it.VisitorID = checkoutVisitor.GetID()

	if it.BillingAddressID == "" && checkoutVisitor.GetBillingAddress() != nil {
		it.BillingAddressID = checkoutVisitor.GetBillingAddress().GetID()
	}

	if it.ShippingAddressID == "" && checkoutVisitor.GetShippingAddress() != nil {
		it.ShippingAddressID = checkoutVisitor.GetShippingAddress().GetID()
	}

	return nil
}

// GetVisitor return checkout visitor
func (it *DefaultCheckout) GetVisitor() visitor.InterfaceVisitor {
	visitorInstance, _ := visitor.LoadVisitorByID(it.VisitorID)
	return visitorInstance
}

// SetSession sets visitor for checkout
func (it *DefaultCheckout) SetSession(checkoutSession api.InterfaceSession) error {
	it.SessionID = checkoutSession.GetID()
	return nil
}

// GetSession return checkout visitor
func (it *DefaultCheckout) GetSession() api.InterfaceSession {
	return api.GetSessionByID(it.SessionID)
}

// GetTaxes collects taxes applied for current checkout
func (it *DefaultCheckout) GetTaxes() (float64, []checkout.StructTaxRate) {

	var amount float64

	if !it.taxesCalculateFlag {
		it.taxesCalculateFlag = true

		it.Taxes = make([]checkout.StructTaxRate, 0)
		for _, tax := range checkout.GetRegisteredTaxes() {
			for _, taxRate := range tax.CalculateTax(it) {
				it.Taxes = append(it.Taxes, taxRate)
				amount += taxRate.Amount
			}
		}

		it.taxesCalculateFlag = false
	} else {
		for _, taxRate := range it.Taxes {
			amount += taxRate.Amount
		}
	}

	return amount, it.Taxes
}

// GetDiscounts collects discounts applied for current checkout
func (it *DefaultCheckout) GetDiscounts() (float64, []checkout.StructDiscount) {

	var amount float64

	if !it.discountsCalculateFlag {
		it.discountsCalculateFlag = true

		it.Discounts = make([]checkout.StructDiscount, 0)
		for _, discount := range checkout.GetRegisteredDiscounts() {
			for _, discountValue := range discount.CalculateDiscount(it) {
				it.Discounts = append(it.Discounts, discountValue)
				amount += discountValue.Amount
			}
		}

		it.discountsCalculateFlag = false
	} else {
		for _, discount := range it.Discounts {
			amount += discount.Amount
		}
	}

	return amount, it.Discounts
}

// GetGrandTotal returns grand total for current checkout: [cart subtotal] + [shipping rate] + [taxes] - [discounts]
func (it *DefaultCheckout) GetGrandTotal() float64 {
	var amount float64

	currentCart := it.GetCart()
	if currentCart != nil {
		amount += currentCart.GetSubtotal()
	}

	if shippingRate := it.GetShippingRate(); shippingRate != nil {
		amount += shippingRate.Price
	}

	taxAmount, _ := it.GetTaxes()
	amount += taxAmount

	discountAmount, _ := it.GetDiscounts()
	amount -= discountAmount

	return amount
}

// SetInfo sets additional info for checkout - any values related to checkout process
func (it *DefaultCheckout) SetInfo(key string, value interface{}) error {
	it.Info[key] = value

	return nil
}

// GetInfo returns additional checkout info value or nil
func (it *DefaultCheckout) GetInfo(key string) interface{} {
	if value, present := it.Info[key]; present {
		return value
	}
	return nil
}

// SetOrder sets order for current checkout
func (it *DefaultCheckout) SetOrder(checkoutOrder order.InterfaceOrder) error {
	it.OrderID = checkoutOrder.GetID()
	return nil
}

// GetOrder returns current checkout related order or nil if not created yet
func (it *DefaultCheckout) GetOrder() order.InterfaceOrder {
	if it.OrderID != "" {
		orderInstance, err := order.LoadOrderByID(it.OrderID)
		if err == nil {
			return orderInstance
		}
	}
	return nil
}

// Submit creates the order with provided information
func (it *DefaultCheckout) Submit() (interface{}, error) {

	if it.GetBillingAddress() == nil {
		return nil, env.ErrorNew("Billing address is not set")
	}

	if it.GetShippingAddress() == nil {
		return nil, env.ErrorNew("Shipping address is not set")
	}

	if it.GetPaymentMethod() == nil {
		return nil, env.ErrorNew("Payment method is not set")
	}

	if it.GetShippingMethod() == nil {
		return nil, env.ErrorNew("Shipping method is not set")
	}

	currentCart := it.GetCart()
	if currentCart == nil {
		return nil, env.ErrorNew("Cart is not specified")
	}

	cartItems := currentCart.GetItems()
	if len(cartItems) == 0 {
		return nil, env.ErrorNew("Cart is empty")
	}

	if err := currentCart.ValidateCart(); err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// making new order if needed
	//---------------------------
	currentTime := time.Now()

	checkoutOrder := it.GetOrder()
	if checkoutOrder == nil {
		newOrder, err := order.GetOrderModel()
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		newOrder.Set("created_at", currentTime)

		checkoutOrder = newOrder
	}

	// updating order information
	//---------------------------
	checkoutOrder.Set("updated_at", currentTime)

	checkoutOrder.Set("status", order.ConstOrderStatusNew)
	if currentVisitor := it.GetVisitor(); currentVisitor != nil {
		checkoutOrder.Set("visitor_id", currentVisitor.GetID())

		checkoutOrder.Set("customer_email", currentVisitor.GetEmail())
		checkoutOrder.Set("customer_name", currentVisitor.GetFullName())
	}

	billingAddress := it.GetBillingAddress().ToHashMap()
	checkoutOrder.Set("billing_address", billingAddress)

	shippingAddress := it.GetShippingAddress().ToHashMap()
	checkoutOrder.Set("shipping_address", shippingAddress)

	checkoutOrder.Set("cart_id", currentCart.GetID())
	checkoutOrder.Set("payment_method", it.GetPaymentMethod().GetCode())
	checkoutOrder.Set("shipping_method", it.GetShippingMethod().GetCode()+"/"+it.GetShippingRate().Code)

	discountAmount, _ := it.GetDiscounts()
	taxAmount, _ := it.GetTaxes()

	checkoutOrder.Set("discount", discountAmount)
	checkoutOrder.Set("tax_amount", taxAmount)
	checkoutOrder.Set("shipping_amount", it.GetShippingRate().Price)

	generateDescriptionFlag := false
	orderDescription := utils.InterfaceToString(it.GetInfo("order_description"))
	if orderDescription == "" {
		generateDescriptionFlag = true
	}

	for _, cartItem := range cartItems {
		orderItem, err := checkoutOrder.AddItem(cartItem.GetProductID(), cartItem.GetQty(), cartItem.GetOptions())
		if err != nil {
			return nil, env.ErrorDispatch(err)
		}

		if generateDescriptionFlag {
			if orderDescription != "" {
				orderDescription += ", "
			}
			orderDescription += fmt.Sprintf("%dx %s", cartItem.GetQty(), orderItem.GetName())
		}
	}
	checkoutOrder.Set("description", orderDescription)

	err := checkoutOrder.CalculateTotals()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	err = checkoutOrder.Proceed()
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	it.SetOrder(checkoutOrder)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// trying to process payment
	//--------------------------
	paymentInfo := make(map[string]interface{})
	paymentInfo["sessionID"] = it.GetSession().GetID()

	result, err := it.GetPaymentMethod().Authorize(checkoutOrder, paymentInfo)
	if err != nil {
		return nil, env.ErrorDispatch(err)
	}

	// TODO: should be different way to do that, result should be some interface or just error but not this
	if result != nil {
		return result, nil
	}

	err = it.CheckoutSuccess(checkoutOrder, it.GetSession())
	if err != nil {
		return nil, err
	}

	return checkoutOrder.ToHashMap(), nil
}
