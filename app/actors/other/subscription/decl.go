// Package subscription implements base subscription functionality
package subscription

import (
	"github.com/ottemo/foundation/env"
)

// Package global constants
const (
	ConstErrorModule = "subscription"
	ConstErrorLevel  = env.ConstErrorLevelActor

	ConstCollectionNameSubscription = "subscription"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"

//	ConstGiftCardStatusCanceled = "canceled"
)
