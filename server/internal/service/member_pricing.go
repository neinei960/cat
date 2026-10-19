package service

import (
	"math"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type MemberPricingInput struct {
	ServiceTotal                      float64
	ProductTotal                      float64
	AddonTotal                        float64
	AppointmentDepositDeductionAmount float64
	ServiceDiscountRate               float64
	ProductDiscountRate               float64
	AvailableBalance                  float64
	ServiceDiscountEligible           bool
}

type MemberPricingBreakdown struct {
	TotalAmount           float64
	ServiceDiscountAmount float64
	ProductDiscountAmount float64
	DiscountAmount        float64
	DiscountRate          float64
	BalanceUsed           float64
	CashPayAmount         float64
	FinalPayAmount        float64
}

type MemberPricingContext struct {
	CardID              uint
	ServiceDiscountRate float64
	ProductDiscountRate float64
	AvailableBalance    float64
}

func ResolveActiveMemberPricingContext(tx *gorm.DB, customerID *uint) MemberPricingContext {
	context := MemberPricingContext{
		ServiceDiscountRate: 1,
		ProductDiscountRate: 1,
	}
	if customerID == nil || *customerID == 0 {
		return context
	}
	if tx == nil {
		tx = database.DB
	}

	var card model.MemberCard
	if err := tx.Where("customer_id = ? AND status = 1", *customerID).Order("id DESC").First(&card).Error; err != nil {
		return context
	}
	if card.ExpireAt != nil && card.ExpireAt.Before(time.Now()) {
		return context
	}

	serviceRate := normalizeMemberPricingRate(card.DiscountRate)
	var customer model.Customer
	if err := tx.Select("id", "discount_rate").First(&customer, *customerID).Error; err == nil {
		customerRate := normalizeMemberPricingRate(customer.DiscountRate)
		if customerRate < 1 {
			serviceRate = customerRate
		}
	}

	return MemberPricingContext{
		CardID:              card.ID,
		ServiceDiscountRate: serviceRate,
		ProductDiscountRate: normalizeMemberPricingRate(card.ProductDiscountRate),
		AvailableBalance:    roundOrderAmount(math.Max(card.Balance, 0)),
	}
}

func CalculateMemberPricingForCustomer(tx *gorm.DB, customerID *uint, input MemberPricingInput) MemberPricingBreakdown {
	context := ResolveActiveMemberPricingContext(tx, customerID)
	input.ServiceDiscountRate = context.ServiceDiscountRate
	input.ProductDiscountRate = context.ProductDiscountRate
	input.AvailableBalance = context.AvailableBalance
	return CalculateMemberPricing(input)
}

func CalculateMemberPricing(input MemberPricingInput) MemberPricingBreakdown {
	serviceTotal := roundOrderAmount(input.ServiceTotal)
	productTotal := roundOrderAmount(input.ProductTotal)
	addonTotal := roundOrderAmount(input.AddonTotal)
	depositDeduction := roundOrderAmount(input.AppointmentDepositDeductionAmount)
	availableBalance := roundOrderAmount(math.Max(input.AvailableBalance, 0))

	serviceRate := normalizeMemberPricingRate(input.ServiceDiscountRate)
	if !input.ServiceDiscountEligible {
		serviceRate = 1
	}
	productRate := normalizeMemberPricingRate(input.ProductDiscountRate)

	serviceDiscount, serviceBalanceUsed := consumeMemberDiscountBucket(serviceTotal, serviceRate, availableBalance)
	remainingBalance := roundOrderAmount(availableBalance - serviceBalanceUsed)
	productDiscount, _ := consumeMemberDiscountBucket(productTotal, productRate, remainingBalance)

	discountAmount := roundOrderAmount(serviceDiscount + productDiscount)
	totalAmount := roundOrderAmount(serviceTotal + productTotal + addonTotal)
	finalPayAmount := roundOrderAmount(math.Max(totalAmount-discountAmount-depositDeduction, 0))
	balanceUsed := roundOrderAmount(math.Min(availableBalance, finalPayAmount))
	cashPayAmount := roundOrderAmount(math.Max(finalPayAmount-balanceUsed, 0))

	return MemberPricingBreakdown{
		TotalAmount:           totalAmount,
		ServiceDiscountAmount: serviceDiscount,
		ProductDiscountAmount: productDiscount,
		DiscountAmount:        discountAmount,
		DiscountRate:          calculateOrderDiscountRate(totalAmount, finalPayAmount),
		BalanceUsed:           balanceUsed,
		CashPayAmount:         cashPayAmount,
		FinalPayAmount:        finalPayAmount,
	}
}

func consumeMemberDiscountBucket(total, rate, availableBalance float64) (float64, float64) {
	total = roundOrderAmount(total)
	rate = normalizeMemberPricingRate(rate)
	availableBalance = roundOrderAmount(availableBalance)
	if total <= 0 || availableBalance <= 0 || rate >= 1 {
		return 0, 0
	}

	fullDiscountedPay := roundOrderAmount(total * rate)
	if availableBalance >= fullDiscountedPay {
		return roundOrderAmount(total - fullDiscountedPay), fullDiscountedPay
	}

	coveredOriginal := roundOrderAmount(availableBalance / rate)
	discount := roundOrderAmount(coveredOriginal * (1 - rate))
	return discount, availableBalance
}

func normalizeMemberPricingRate(rate float64) float64 {
	if rate > 0 && rate < 1 {
		return rate
	}
	return 1
}
