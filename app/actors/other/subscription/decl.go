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

	ConstCollectionNameSubscription = "subscription"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"

	ConstTimeDay = time.Hour *24
//	ConstGiftCardStatusCanceled = "canceled"
)
