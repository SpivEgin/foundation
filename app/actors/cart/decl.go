// Package cart is a default implementation of interfaces declared in
// "github.com/ottemo/foundation/app/models/cart" package
package cart

import (
	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstCartCollectionName      = "cart"
	ConstCartItemsCollectionName = "cart_items"

	ConstErrorModule = "cart"
	ConstErrorLevel  = env.ConstErrorLevelActor
)

// DefaultCart is a default implementer of InterfaceCart
type DefaultCart struct {
	id string

	VisitorID string

	Info map[string]interface{}

	Items map[int]cart.InterfaceCartItem

	Active bool

	Subtotal float64

	maxIdx int
}

// DefaultCartItem is a default implementer of InterfaceCart
type DefaultCartItem struct {
	id string

	idx int

	ProductID string

	Qty int

	Options map[string]interface{}

	Cart *DefaultCart
}
