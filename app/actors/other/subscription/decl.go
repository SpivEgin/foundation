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

	ConstConfigPathSubscription                 = "general.subscription"
	ConstConfigPathSubscriptionEnabled          = "general.subscription.enabled"
	ConstConfigPathSubscriptionEmailSubject     = "general.subscription.emailSubject"
	ConstConfigPathSubscriptionEmailTemplate    = "general.subscription.emailTemplate"
	ConstConfigPathSubscriptionConfirmationLink = "general.subscription.confirmationLink"

	ConstCollectionNameSubscription = "subscription"

	ConstSubscriptionStatusSuspended = "suspended"
	ConstSubscriptionStatusConfirmed = "confirmed"

	ConstTimeDay = time.Hour * 24

//	ConstGiftCardStatusCanceled = "canceled"
)
