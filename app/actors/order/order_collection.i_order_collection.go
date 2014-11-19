package order

import (
	"github.com/ottemo/foundation/app/models/order"
	"github.com/ottemo/foundation/db"
)

// returns database collection
func (it *DefaultOrderCollection) GetDBCollection() db.InterfaceDBCollection {
	return it.listCollection
}

// returns array of products in model instance form
func (it *DefaultOrderCollection) ListOrders() []order.InterfaceOrder {
	result := make([]order.InterfaceOrder, 0)

	dbRecords, err := it.listCollection.Load()
	if err != nil {
		return result
	}

	for _, dbRecordData := range dbRecords {
		orderModel, err := order.GetOrderModel()
		if err != nil {
			return result
		}
		orderModel.FromHashMap(dbRecordData)

		result = append(result, orderModel)
	}

	return result
}
