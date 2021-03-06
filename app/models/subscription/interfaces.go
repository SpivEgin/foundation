// Package subscription represents abstraction of business layer purchase subscription object
package subscription

import (
	"time"

	"github.com/ottemo/foundation/app/models"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/visitor"
)

// InterfaceSubscription represents interface to access business layer implementation of purchase subscription object
type InterfaceSubscription interface {
	GetCustomerEmail() string
	GetCustomerName() string

	GetOrderID() string
	GetVisitorID() string

	GetItems() []StructSubscriptionItem

	SetShippingAddress(address visitor.InterfaceVisitorAddress) error
	GetShippingAddress() visitor.InterfaceVisitorAddress

	SetBillingAddress(address visitor.InterfaceVisitorAddress) error
	GetBillingAddress() visitor.InterfaceVisitorAddress

	SetCreditCard(creditCard visitor.InterfaceVisitorCard) error
	GetCreditCard() visitor.InterfaceVisitorCard

	GetPaymentMethod() checkout.InterfacePaymentMethod

	SetShippingMethod(shippingMethod checkout.InterfaceShippingMethod) error
	GetShippingMethod() checkout.InterfaceShippingMethod

	SetShippingRate(shippingRate checkout.StructShippingRate) error
	GetShippingRate() checkout.StructShippingRate

	GetStatus() string
	SetStatus(status string) error

	GetActionDate() time.Time
	SetActionDate(actionDate time.Time) error

	GetPeriod() int
	SetPeriod(days int) error

	SetInfo(key string, value interface{})
	GetInfo(key string) interface{}

	UpdateActionDate() error

	Validate() error
	GetCheckout() (checkout.InterfaceCheckout, error)

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceListable
}

// StructSubscriptionItem hold data related to subscription item
type StructSubscriptionItem struct {
	ProductID string                 `json:"product_id"`
	Options   map[string]interface{} `json:"options"`
	Qty       int                    `json:"qty"`
	Name      string                 `json:"name"`
	Sku       string                 `json:"sku"`
	Price     float64                `json:"price"`
}

// InterfaceSubscriptionCollection represents interface to access business layer implementation of purchase subscription collection
type InterfaceSubscriptionCollection interface {
	ListSubscriptions() []InterfaceSubscription

	models.InterfaceCollection
}
