package cart

import (
	"github.com/ottemo/foundation/app/models/cart"
	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/app/models/product"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// GetID returns id of cart item
func (it *DefaultCartItem) GetID() string {
	return it.id
}

// SetID sets id to cart item
func (it *DefaultCartItem) SetID(newID string) error {
	it.id = newID
	return nil
}

// GetIdx returns index value for current cart item
func (it *DefaultCartItem) GetIdx() int {
	return it.idx
}

// SetIdx changes index value for current cart item if it is possible
func (it *DefaultCartItem) SetIdx(newIdx int) error {

	if newIdx < 0 {
		return env.ErrorNew("wrong cart item index")
	}

	if value, present := it.Cart.Items[newIdx]; present {
		it.Cart.Items[newIdx] = it
		it.Cart.Items[it.idx] = value
		it.idx = newIdx
	} else {
		it.Cart.Items[newIdx] = it
		it.idx = newIdx
	}

	return nil
}

// GetProductID returns product id which cart item represents
func (it *DefaultCartItem) GetProductID() string {
	return it.ProductID
}

// GetProduct returns product instance which cart item represents
func (it *DefaultCartItem) GetProduct() product.InterfaceProduct {
	if it.ProductID != "" {
		product, err := product.LoadProductByID(it.ProductID)
		if err == nil {
			return product
		}
	}
	return nil
}

// GetQty returns current cart item qty
func (it *DefaultCartItem) GetQty() int {
	return it.Qty
}

// SetQty sets qty for current cart item
func (it *DefaultCartItem) SetQty(qty int) error {
	if qty > 0 {
		it.Qty = qty
	} else {
		return env.ErrorNew("qty must be greater then 0")
	}

	if err := it.ValidateProduct(); err != nil {
		return err
	}

	it.Cart.cartChanged()

	return nil
}

// Remove removes item from the cart
func (it *DefaultCartItem) Remove() error {

	if it.Cart != nil {
		return it.Cart.RemoveItem(it.idx)
	}

	return env.ErrorNew("item is not bound to cart")
}

// GetOptions returns all item options or nil
func (it *DefaultCartItem) GetOptions() map[string]interface{} {
	return it.Options
}

// SetOption sets an option to cart item
func (it *DefaultCartItem) SetOption(optionName string, optionValue interface{}) error {
	if it.Options == nil {
		it.Options = make(map[string]interface{})
	}

	it.Options[optionName] = optionValue

	return nil
}

// GetCart returns cart that item belongs to
func (it *DefaultCartItem) GetCart() cart.InterfaceCart {
	return it.Cart
}

// ValidateProduct checks that cart product is existent and have available qty
func (it *DefaultCartItem) ValidateProduct() error {
	cartProduct := it.GetProduct()
	if cartProduct == nil {
		return env.ErrorNew("product " + it.GetProductID() + " is not exists currently")
	}

	if cartProduct.GetEnabled() == false {
		return env.ErrorNew("product " + it.GetProductID() + " is not available")
	}

	err := cartProduct.ApplyOptions(it.Options)
	if err != nil {
		return err
	}

	allowBackorders := utils.InterfaceToBool(env.ConfigGetValue(checkout.ConstConfigPathBackorders))
	if !allowBackorders && product.GetRegisteredStock() != nil {
		if qty := cartProduct.GetQty(); qty < it.GetQty() {
			return env.ErrorNew("item out of stock")
		}
	}

	return nil
}
