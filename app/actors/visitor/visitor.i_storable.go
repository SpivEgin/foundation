package visitor

import (
	"github.com/ottemo/foundation/app/actors/visitor/address"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
)

// GetID returns current ID of the Visitor
func (it *DefaultVisitor) GetID() string {
	return it.id
}

// SetID sets current ID of the current Visitor
func (it *DefaultVisitor) SetID(NewID string) error {
	it.id = NewID
	return nil
}

// Load will retrieve the Visitor information from database
func (it *DefaultVisitor) Load(ID string) error {

	collection, err := db.GetCollection(ConstCollectionNameVisitor)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	values, err := collection.LoadByID(ID)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	err = it.FromHashMap(values)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// Delete removes current Visitor from the database
func (it *DefaultVisitor) Delete() error {

	collection, err := db.GetCollection(ConstCollectionNameVisitor)
	if err != nil {
		return env.ErrorDispatch(err)
	}
	addressCollection, err := db.GetCollection(address.ConstCollectionNameVisitorAddress)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	if err := addressCollection.AddFilter("visitor_id", "=", it.GetID()); err != nil {
		_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "d35f39ef-1e4d-4072-9cce-11f3dceb0b5c", err.Error())
	}
	if _, err := addressCollection.Delete(); err != nil {
		return env.ErrorDispatch(err)
	}

	err = collection.DeleteByID(it.GetID())
	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}

// Save stores current Visitor to the database
func (it *DefaultVisitor) Save() error {

	collection, err := db.GetCollection(ConstCollectionNameVisitor)
	if err != nil {
		return env.ErrorDispatch(err)
	}

	if it.GetID() == "" {
		if err := collection.AddFilter("email", "=", it.GetEmail()); err != nil {
			_ = env.ErrorNew(ConstErrorModule, ConstErrorLevel, "105b8b9f-d901-4ec2-94eb-691da841a481", err.Error())
		}
		n, err := collection.Count()
		if err != nil {
			return env.ErrorDispatch(err)
		}
		if n > 0 {
			return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "29be1531-cb6b-44cf-a78e-f1bf9aae1163", "The requested email address is already in use.")
		}
	}

	storableValues := it.ToHashMap()

	delete(storableValues, "billing_address")
	delete(storableValues, "shipping_address")
	delete(storableValues, "token")

	storableValues["facebook_id"] = it.FacebookID
	storableValues["google_id"] = it.GoogleID
	storableValues["password"] = it.Password
	storableValues["validate"] = it.VerificationKey

	// shipping address save
	if it.ShippingAddress != nil {
		err := it.ShippingAddress.Save()
		if err != nil {
			return env.ErrorDispatch(err)
		}

		storableValues["shipping_address_id"] = it.ShippingAddress.GetID()
	}

	// billing address save
	if it.BillingAddress != nil {
		err := it.BillingAddress.Save()
		if err != nil {
			return env.ErrorDispatch(err)
		}

		storableValues["billing_address_id"] = it.BillingAddress.GetID()
	}

	// token save
	if it.Token != nil {
		err := it.Token.Save()
		if err != nil {
			return env.ErrorDispatch(err)
		}

		storableValues["token_id"] = it.Token.GetID()
	}

	// saving visitor
	if newID, err := collection.Save(storableValues); err == nil {
		if err := it.Set("_id", newID); err != nil {
			return env.ErrorNew(ConstErrorModule, ConstErrorLevel, "4dd5f3ca-188f-4ddd-a654-45418caf2009", err.Error())
		}
	} else {
		return env.ErrorDispatch(err)
	}

	return nil
}
