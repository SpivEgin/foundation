package order

import (
	"github.com/ottemo/foundation/db"
)

// returns database collection
func (it *DefaultOrderItemCollection) GetDBCollection() db.InterfaceDBCollection {
	return it.listCollection
}
