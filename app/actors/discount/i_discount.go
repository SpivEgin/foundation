package discount

import (
	"time"

	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/utils"
)

// GetName returns name of current discount implementation
func (it *DefaultDiscount) GetName() string {
	return "CouponDiscount"
}

// GetCode returns code of current discount implementation
func (it *DefaultDiscount) GetCode() string {
	return "coupon_discount"
}

// CalculateDiscount calculates and returns a set of discounts applied to given checkout
func (it *DefaultDiscount) CalculateDiscount(checkoutInstance checkout.I_Checkout) []checkout.T_Discount {

	result := make([]checkout.T_Discount, 0)

	// checking session for applied coupon codes
	if currentSession := checkoutInstance.GetSession(); currentSession != nil {
		if appliedCodes, ok := currentSession.Get(SESSION_KEY_APPLIED_DISCOUNT_CODES).([]string); ok && len(appliedCodes) > 0 {

			// getting order information will use in calculations
			discountableAmount := checkoutInstance.GetGrandTotal()
			grandTotalAmount := checkoutInstance.GetGrandTotal()

			// loading information about applied discounts
			collection, err := db.GetCollection(COLLECTION_NAME_COUPON_DISCOUNTS)
			if err != nil {
				return result
			}
			err = collection.AddFilter("code", "in", appliedCodes)
			if err != nil {
				return result
			}

			records, err := collection.Load()
			if err != nil {
				return result
			}

			// making coupon code map for right apply order
			discountCodes := make(map[string]map[string]interface{})
			for _, record := range records {
				if discountCode := utils.InterfaceToString(record["code"]); discountCode != "" {
					discountCodes[discountCode] = record
				}
			}

			// applying discount codes
			for appliedCodesIdx, discountCode := range appliedCodes {
				if discountCoupon, ok := discountCodes[discountCode]; ok {

					applyTimes := utils.InterfaceToInt(discountCoupon["times"])
					workSince := utils.InterfaceToTime(discountCoupon["since"])
					workUntil := utils.InterfaceToTime(discountCoupon["until"])

					currentTime := time.Now()

					// to be applicable coupon should satisfy following conditions:
					//   [applyTimes] should be -1 or >0 and [workSince] >= currentTime <= [workUntil] if set
					if (applyTimes == -1 || applyTimes > 0) &&
						(utils.IsZeroTime(workSince) || workSince.Unix() <= currentTime.Unix()) &&
						(utils.IsZeroTime(workUntil) || workUntil.Unix() >= currentTime.Unix()) {

						// calculating coupon discount amount
						discountAmount := utils.InterfaceToFloat64(discountCoupon["amount"])
						discountPercent := utils.InterfaceToFloat64(discountCoupon["percent"])

						discountAmount = utils.RoundPrice(discountAmount) + utils.RoundPrice(grandTotalAmount/100*discountPercent)

						if discountableAmount > discountAmount {
							discountableAmount -= discountAmount
						} else {
							discountAmount = discountableAmount
							discountableAmount = 0
						}

						result = append(result, checkout.T_Discount{
							Name:   utils.InterfaceToString(discountCoupon["name"]),
							Code:   utils.InterfaceToString(discountCoupon["code"]),
							Amount: discountAmount,
						})

					} else {
						// we have not applicable coupon - removing it from applied coupons list
						newAppliedCodes := make([]string, 0, len(appliedCodes)-1)
						for idx, value := range appliedCodes {
							if idx != appliedCodesIdx {
								newAppliedCodes = append(newAppliedCodes, value)
							}
						}
						currentSession.Set(SESSION_KEY_APPLIED_DISCOUNT_CODES, newAppliedCodes)
					}
				}
			}
		}
	}

	return result
}
