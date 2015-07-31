// Package subscription implements base subscription functionality
package subscription

import (
	"github.com/ottemo/foundation/env"
	"time"
)

// Package global constants
const (
	ConstErrorModule = "subscription"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstConfigPathSubscription        = "general.subscription"
	ConstConfigPathSubscriptionEnabled = "general.subscription.enabled"

	ConstConfigPathSubscriptionEmailSubject     = "general.subscription.emailSubject"
	ConstConfigPathSubscriptionEmailTemplate    = "general.subscription.emailTemplate"
	ConstConfigPathSubscriptionConfirmationLink = "general.subscription.confirmationLink"

	ConstConfigPathSubscriptionSubmitEmailSubject  = "general.subscription.emailSubmitSubject"
	ConstConfigPathSubscriptionSubmitEmailTemplate = "general.subscription.emailSubmitTemplate"
	ConstConfigPathSubscriptionSubmitEmailLink     = "general.subscription.SubmitLink"

	ConstCollectionNameSubscription = "subscription"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"
	ConstSubscriptionStatusCanceled  = "canceled"

	ConstSubscriptionActionSubmit = "submit"
	ConstSubscriptionActionUpdate = "update"
	ConstSubscriptionActionCreate = "create"

	ConstTimeDay = time.Hour * 24

	ConstCreationDaysDelay = 33
)

var (
	nextCreationDate time.Time
)

// DefaultSubscription just for controlling values
type DefaultSubscription struct {
	id        string
	OrderID   string
	CartID    string
	VisitorID string

	Email string
	Name  string

	Status string
	State  string
	Action string

	ShippingAddress map[string]string

	LastSubmit time.Time
	NextAction time.Time

	Period int
}
