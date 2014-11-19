// Package visitor represents abstraction of business layer visitor object
package visitor

import (
	"github.com/ottemo/foundation/app/models"
	"time"
)

// Package global constants
const (
	ConstModelNameVisitor                  = "Visitor"
	ConstModelNameVisitorCollection        = "VisitorCollection"
	ConstModelNameVisitorAddress           = "VisitorAddress"
	ConstModelNameVisitorAddressCollection = "VisitorAddressCollection"

	ConstSessionKeyVisitorID = "visitor_id"
)

// InterfaceVisitor represents interface to access business layer implementation of visitor object
type InterfaceVisitor interface {
	GetEmail() string
	GetFacebookID() string
	GetGoogleID() string

	GetFullName() string
	GetFirstName() string
	GetLastName() string

	GetBirthday() time.Time
	GetCreatedAt() time.Time

	GetShippingAddress() InterfaceVisitorAddress
	GetBillingAddress() InterfaceVisitorAddress

	SetShippingAddress(address InterfaceVisitorAddress) error
	SetBillingAddress(address InterfaceVisitorAddress) error

	SetPassword(passwd string) error
	CheckPassword(passwd string) bool
	GenerateNewPassword() error

	IsAdmin() bool

	IsValidated() bool
	Invalidate() error
	Validate(key string) error

	LoadByEmail(email string) error
	LoadByFacebookID(facebookID string) error
	LoadByGoogleID(googleID string) error

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
	models.InterfaceListable
	models.InterfaceCustomAttributes
}

// InterfaceVisitorCollection represents interface to access business layer implementation of visitor collection
type InterfaceVisitorCollection interface {
	ListVisitors() []InterfaceVisitor

	models.InterfaceCollection
}

// InterfaceVisitorAddress represents interface to access business layer implementation of visitor address object
type InterfaceVisitorAddress interface {
	GetVisitorID() string

	GetFirstName() string
	GetLastName() string

	GetCompany() string

	GetCountry() string
	GetState() string
	GetCity() string

	GetAddress() string
	GetAddressLine1() string
	GetAddressLine2() string

	GetPhone() string
	GetZipCode() string

	models.InterfaceModel
	models.InterfaceObject
	models.InterfaceStorable
}

// InterfaceVisitorAddressCollection represents interface to access business layer implementation of visitor address collection
type InterfaceVisitorAddressCollection interface {
	ListVisitorsAddresses() []InterfaceVisitorAddress

	models.InterfaceCollection
}
