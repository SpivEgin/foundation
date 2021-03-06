package coupon

import (
	"strings"
	"time"

	"github.com/ottemo/foundation/app/models/checkout"
	"github.com/ottemo/foundation/db"
	"github.com/ottemo/foundation/env"
	"github.com/ottemo/foundation/utils"
)

// GetName returns the name of the current coupon implementation
func (it *Coupon) GetName() string {
	return "Coupon"
}

// GetCode returns the code of the current coupon implementation
func (it *Coupon) GetCode() string {
	return "coupon"
}

// GetPriority returns the code of the current coupon implementation
func (it *Coupon) GetPriority() []float64 {
	baseCouponPriority := utils.InterfaceToFloat64(env.ConfigGetValue(ConstConfigPathDiscountApplyPriority))
	cartCalculationPriority := baseCouponPriority + 0.01
	return []float64{baseCouponPriority, cartCalculationPriority}
}

// Calculate calculates and returns a set of coupons applied to the provided checkout
func (it *Coupon) Calculate(checkoutInstance checkout.InterfaceCheckout, currentPriority float64) []checkout.StructPriceAdjustment {

	var result []checkout.StructPriceAdjustment

	// check session for applied coupon codes
	if currentSession := checkoutInstance.GetSession(); currentSession != nil {

		redeemedCodes := utils.InterfaceToStringArray(currentSession.Get(ConstSessionKeyCurrentRedemptions))

		if len(redeemedCodes) > 0 {

			// loading information about applied coupons
			collection, err := db.GetCollection(ConstCollectionNameCouponDiscounts)
			if err != nil {
				return result
			}
			err = collection.AddFilter("code", "in", redeemedCodes)
			if err != nil {
				return result
			}

			records, err := collection.Load()
			if err != nil || len(records) == 0 {
				return result
			}

			applicableProductDiscounts := make(map[string][]discount)
			applicableCartDiscounts := make([]discount, 0)

			// collect products to one map, that holds productID: qty and used to get apply qty
			productsInCart := make(map[string]int)

			items := checkoutInstance.GetItems()
			for _, productInCart := range items {
				productID := productInCart.GetProductID()
				productQty := productInCart.GetQty()

				if qty, present := productsInCart[productID]; present {
					productsInCart[productID] = qty + productQty
					continue
				}
				productsInCart[productID] = productQty
				applicableProductDiscounts[productID] = make([]discount, 0)
			}

			// use coupon map to hold the correct application order and ignore previously used coupons
			discountCodes := make(map[string]map[string]interface{})
			for _, record := range records {

				discountsUsageQty := getCouponApplyQty(productsInCart, record)
				discountCode := utils.InterfaceToString(record["code"])

				if discountCode != "" && discountsUsageQty > 0 {
					record["usage_qty"] = discountsUsageQty
					discountCodes[discountCode] = record
				}
			}

			couponPriorityValue := utils.InterfaceToFloat64(env.ConfigGetValue(ConstConfigPathDiscountApplyPriority))
			productCouponsCalculation := couponPriorityValue == currentPriority

			// accumulation of coupon discounts for cart to result and for products to applicableProductDiscounts
			for appliedCodesIdx, discountCode := range redeemedCodes {
				discountCoupon, present := discountCodes[discountCode]
				if !present {
					continue
				}

				validStart := isValidStart(discountCoupon["since"])
				validEnd := isValidEnd(discountCoupon["until"])

				// to be applicable coupon should satisfy following conditions:
				//   [begin] >= currentTime <= [end] if set
				if !validStart || !validEnd {
					// we have not applicable coupon - removing it from applied coupons list
					newRedemptions := make([]string, 0, len(redeemedCodes)-1)
					for idx, value := range redeemedCodes {
						if idx != appliedCodesIdx {
							newRedemptions = append(newRedemptions, value)
						}
					}
					currentSession.Set(ConstSessionKeyCurrentRedemptions, newRedemptions)
					continue
				}

				discountTarget := utils.InterfaceToString(discountCoupon["target"])

				// add discount object for every product id that it can affect
				applicableDiscount := discount{
					Code:     utils.InterfaceToString(discountCoupon["code"]),
					Name:     utils.InterfaceToString(discountCoupon["name"]),
					Amount:   utils.InterfaceToFloat64(discountCoupon["amount"]),
					Percents: utils.InterfaceToFloat64(discountCoupon["percent"]),
					Qty:      utils.InterfaceToInt(discountCoupon["usage_qty"]),
				}

				// if we in product coupons calculation then skip cart coupons
				if strings.Contains(discountTarget, checkout.ConstDiscountObjectCart) || discountTarget == "" {
					if !productCouponsCalculation {
						applicableCartDiscounts = append(applicableCartDiscounts, applicableDiscount)
					}

					continue
				}

				// collect only discounts for productIDs that are in cart
				for _, productID := range utils.InterfaceToStringArray(discountTarget) {
					if discounts, present := applicableProductDiscounts[productID]; present {
						applicableProductDiscounts[productID] = append(discounts, applicableDiscount)
					}
				}
			}

			// handle cart coupon discounts to find one that is biggest and append it to result
			// as price adjustments in % and in $ amount
			if !productCouponsCalculation {
				currentCartAmount := checkoutInstance.GetItemSpecificTotal(0, checkout.ConstLabelGrandTotal)
				if len(applicableCartDiscounts) > 0 && currentCartAmount > 0 {
					applicableDiscount, _ := findBiggestDiscount(applicableCartDiscounts, currentCartAmount)

					currentPriceAdjustment := checkout.StructPriceAdjustment{
						Code:      applicableDiscount.Code,
						Name:      applicableDiscount.Name,
						Amount:    applicableDiscount.Percents * -1,
						IsPercent: true,
						Priority:  currentPriority,
						Labels:    []string{checkout.ConstLabelDiscount},
						PerItem:   nil,
					}

					if applicableDiscount.Percents > 0 {
						currentPriceAdjustment.Priority += float64(0.00001)

						result = append(result, currentPriceAdjustment)
					}

					if applicableDiscount.Amount > 0 {
						currentPriceAdjustment.Amount = applicableDiscount.Amount * -1
						currentPriceAdjustment.IsPercent = false
						currentPriceAdjustment.Priority += float64(0.0001)

						result = append(result, currentPriceAdjustment)
					}
				}

				return result
			}

			// hold price adjustment for every coupon code ( to make total details with right description)
			priceAdjustments := make(map[string]checkout.StructPriceAdjustment)

			// adding to discounts the biggest applicable discount per product
			for _, cartItem := range checkoutInstance.GetDiscountableItems() {
				index := utils.InterfaceToString(cartItem.GetIdx())

				if cartProduct := cartItem.GetProduct(); cartProduct != nil {
					productPrice := cartProduct.GetPrice()
					productID := cartItem.GetProductID()
					productQty := cartItem.GetQty()

					// discount will be applied for every single product and grouped per item
					for i := 0; i < productQty; i++ {
						productDiscounts, present := applicableProductDiscounts[productID]
						if !present || len(productDiscounts) <= 0 {
							break
						}

						// looking for biggest applicable discount for current item
						biggestAppliedDiscount, biggestAppliedDiscountIndex := findBiggestDiscount(productDiscounts, productPrice)

						// update used discount and change qty of chosen discount to number of usage
						discountUsed := productDiscounts[biggestAppliedDiscountIndex].Qty
						if discountableProductsQty := productQty - i; discountableProductsQty < discountUsed {
							discountUsed = discountableProductsQty
						}
						i += discountUsed - 1

						productDiscounts[biggestAppliedDiscountIndex].Qty -= discountUsed

						// remove fully used discount from discounts list
						var newProductDiscounts []discount
						for _, currentDiscount := range productDiscounts {
							if currentDiscount.Qty > 0 {
								newProductDiscounts = append(newProductDiscounts, currentDiscount)
							}
						}
						applicableProductDiscounts[productID] = newProductDiscounts

						// making from discount price adjustment
						// calculating amount that will be discounted from item
						amount := float64(discountUsed) * biggestAppliedDiscount.Total * -1

						// add this amount to already existing PA (with the same coupon code) or creating new
						if priceAdjustment, present := priceAdjustments[biggestAppliedDiscount.Code]; present {
							priceAdjustment.PerItem[index] = utils.RoundPrice(priceAdjustment.PerItem[index] + amount)
							priceAdjustments[biggestAppliedDiscount.Code] = priceAdjustment
						} else {
							currentPriority += float64(0.000001)
							priceAdjustments[biggestAppliedDiscount.Code] = checkout.StructPriceAdjustment{
								Code:      biggestAppliedDiscount.Code,
								Name:      biggestAppliedDiscount.Name,
								Amount:    0,
								IsPercent: false,
								Priority:  currentPriority,
								Labels:    []string{checkout.ConstLabelDiscount},
								PerItem: map[string]float64{
									index: amount,
								},
							}
						}
					}
				}
			}

			// attach price adjustments on products to result
			for _, priceAdjustment := range priceAdjustments {
				result = append(result, priceAdjustment)
			}
		}
	}

	return result
}

//finds biggest discount amount if applied more than one coupon
func findBiggestDiscount(discounts []discount, total float64) (discount, int) {

	var biggestAppliedDiscount discount
	var biggestAppliedDiscountIndex int

	// looking for biggest applicable discount for current item
	for index, discount := range discounts {
		if (discount.Qty) > 0 {
			productDiscountableAmount := discount.Amount + total*discount.Percents/100

			// if we have discount that is bigger then a price we will apply it
			if productDiscountableAmount > total {
				biggestAppliedDiscount = discount
				biggestAppliedDiscount.Total = total
				biggestAppliedDiscountIndex = index
				break
			}

			if biggestAppliedDiscount.Total < productDiscountableAmount {
				biggestAppliedDiscount = discount
				biggestAppliedDiscount.Total = productDiscountableAmount
				biggestAppliedDiscountIndex = index
			}
		}
	}

	return biggestAppliedDiscount, biggestAppliedDiscountIndex
}

// check coupon limitation parameters for correspondence to current checkout values
// return qty of usages if coupon is allowed for current checkout and satisfies all conditions
func getCouponApplyQty(productsInCart map[string]int, couponDiscount map[string]interface{}) int {

	result := -1
	if limits, present := couponDiscount["limits"]; present {
		limitations := utils.InterfaceToMap(limits)
		if len(limitations) > 0 {
			for limitingKey, limitingValue := range limitations {

				switch strings.ToLower(limitingKey) {
				case "product_in_cart":
					requiredProduct := utils.InterfaceToStringArray(limitingValue)
					for index, productID := range requiredProduct {
						if _, present := productsInCart[productID]; present {
							break
						}
						if index == (len(requiredProduct) - 1) {
							return 0
						}
					}

				case "products_in_cart":
					requiredProducts := utils.InterfaceToStringArray(limitingValue)
					for _, productID := range requiredProducts {
						if _, present := productsInCart[productID]; !present {
							return 0
						}
					}

				case "products_in_qty":
					requiredProducts := utils.InterfaceToMap(limitingValue)
					for requiredProductID, requiredQty := range requiredProducts {
						requiredQty := utils.InterfaceToInt(requiredQty)
						if requiredQty < 1 {
							requiredQty = 1
						}
						productQty, present := productsInCart[requiredProductID]
						limitingQty := utils.InterfaceToInt(productQty / requiredQty)

						if !present || limitingQty < 1 {
							return 0
						}

						if result == -1 || limitingQty < result {
							result = limitingQty
						}

					}
				} // end of switch
			} // end of loop

			if maxLimitValue, present := limitations["max_usage_qty"]; present {
				limitingQty := utils.InterfaceToInt(maxLimitValue)
				if limitingQty > 0 && (result > limitingQty || result == -1) {
					result = limitingQty
				}
				if result == -1 && limitingQty == -1 {
					result = 9999
				}
			}
		}
	}
	if result == -1 {
		result = 1
	}

	return result
}

// validStart returns a boolean value of the datetame passed is valid
func isValidStart(start interface{}) bool {

	couponStart := utils.InterfaceToTime(start)
	currentTime := time.Now()

	isValidStart := (utils.IsZeroTime(couponStart) || couponStart.Unix() <= currentTime.Unix())

	return isValidStart
}

// validEnd returns a boolean value of the datetame passed is valid
func isValidEnd(end interface{}) bool {

	couponEnd := utils.InterfaceToTime(end)
	currentTime := time.Now()

	// to be applicable coupon should satisfy following conditions:
	isValidEnd := (utils.IsZeroTime(couponEnd) || couponEnd.Unix() >= currentTime.Unix())

	return isValidEnd
}
