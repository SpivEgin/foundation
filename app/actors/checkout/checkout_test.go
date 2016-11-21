package checkout

import (
	"github.com/ottemo/foundation/app/models/checkout"
	"testing"
)

var (
	TaxPriority = 2.50
)

type sampleShipping struct {}

func (*sampleShipping) GetName() string {
	return "sampleShipping"
}

func (*sampleShipping) GetCode() string {
	return "sample"
}

func (*sampleShipping) IsAllowed(checkoutInstance checkout.InterfaceCheckout) bool {
	return true
}

func (it *sampleShipping) GetRates(checkoutInstance checkout.InterfaceCheckout) []checkout.StructShippingRate {
	return []checkout.StructShippingRate {
		checkout.StructShippingRate {
			Name: "Sample Shipping",
			Code: "sample",
			Price: 10,
		},
	}
}

func (it *sampleShipping) GetAllRates() []checkout.StructShippingRate {
	return it.GetRates()
}

type sampleTax struct {}

func (it *sampleTax) GetName() string {
	return "sampleTax"
}

func (it *sampleTax) GetCode() string {
	return "sampleTax"
}

func (it *sampleTax) GetPriority() []float64 {
	return TaxPriority
}

func (it *sampleTax) Calculate(checkoutInstance checkout.InterfaceCheckout, currentPriority float64) []checkout.StructPriceAdjustment {
	return []checkout.StructPriceAdjustment {
		checkout.StructPriceAdjustment {
			Code: "tax",
			Name: "tax",
			Priority: 1,
			Amount: 10,
			IsPercent: true,
			Labels: checkout.ConstLabelDiscount,
			PerItem: nil,
		},
	}
}

func TestCalculateTotals(t *testing.T) {
	err := checkout.RegisterShippingMethod(new(sampleShipping))
	if err != nil {
		t.Error(err)
		return;
	}

	err = checkout.RegisterPriceAdjustment(new(sampleTax))
	if err != nil {
		t.Error(err)
		return;
	}

	checkoutInstance, err := checkout.GetCheckoutModel()
	if err != nil {
		t.Error(err)
		return;
	}

	checkoutInstance.GetSubtotal()
}
