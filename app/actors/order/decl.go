// Package order is a default implementation of interfaces declared in
// "github.com/ottemo/foundation/app/models/order" package
package order

import (
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/db"

	"sync"
	"time"
)

// Package global variables
var (
	lastIncrementId      int = 0
	lastIncrementIdMutex sync.Mutex
)

// Package global constants
const (
	ConstCollectionNameOrder      = "orders"
	ConstCollectionNameOrderItems = "order_items"

	ConstIncrementIDFormat = "%0.10d"

	ConstConfigPathLastIncrementID = "internal.order.increment_id"
)

// DefaultOrderItem is a default implementer of InterfaceOrderItem
type DefaultOrderItem struct {
	id  string
	idx int

	OrderId string

	ProductId string

	Qty int

	Name string
	Sku  string

	ShortDescription string

	Options map[string]interface{}

	Price  float64
	Weight float64
}

// DefaultOrder is a default implementer of InterfaceOrder
type DefaultOrder struct {
	id string

	IncrementId string
	Status      string

	VisitorId string
	CartId    string

	Description string
	PaymentInfo map[string]interface{}

	BillingAddress  map[string]interface{}
	ShippingAddress map[string]interface{}

	CustomerEmail string
	CustomerName  string

	PaymentMethod  string
	ShippingMethod string

	Subtotal       float64
	Discount       float64
	TaxAmount      float64
	ShippingAmount float64
	GrandTotal     float64

	CreatedAt time.Time
	UpdatedAt time.Time

	Items map[int]order.InterfaceOrderItem

	maxIdx int
}

// DefaultOrderItemCollection is a default implementer of InterfaceOrderCollection
type DefaultOrderCollection struct {
	listCollection     db.InterfaceDBCollection
	listExtraAtributes []string
}

// DefaultOrderItemCollection is a default implementer of InterfaceOrderItemCollection
type DefaultOrderItemCollection struct {
	listCollection     db.InterfaceDBCollection
	listExtraAtributes []string
}
