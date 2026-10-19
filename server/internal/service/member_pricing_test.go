package service

import "testing"

func TestCalculateMemberPricingZeroBalanceRemovesDiscount(t *testing.T) {
	got := CalculateMemberPricing(MemberPricingInput{
		ServiceTotal:            188,
		ServiceDiscountRate:     0.9,
		AvailableBalance:        0,
		ServiceDiscountEligible: true,
	})

	if got.ServiceDiscountAmount != 0 || got.DiscountAmount != 0 || got.FinalPayAmount != 188 || got.DiscountRate != 1 {
		t.Fatalf("expected zero balance to keep original 188.00, got %+v", got)
	}
}

func TestCalculateMemberPricingOnlyDiscountsBalanceCoveredAmount(t *testing.T) {
	got := CalculateMemberPricing(MemberPricingInput{
		ServiceTotal:            188,
		ServiceDiscountRate:     0.9,
		AvailableBalance:        2,
		ServiceDiscountEligible: true,
	})

	if got.ServiceDiscountAmount != 0.22 || got.DiscountAmount != 0.22 || got.FinalPayAmount != 187.78 || got.BalanceUsed != 2 || got.CashPayAmount != 185.78 {
		t.Fatalf("expected only balance-covered service amount to receive discount, got %+v", got)
	}
}

func TestCalculateMemberPricingFullBalanceAppliesMemberDiscount(t *testing.T) {
	got := CalculateMemberPricing(MemberPricingInput{
		ServiceTotal:            188,
		ServiceDiscountRate:     0.9,
		AvailableBalance:        200,
		ServiceDiscountEligible: true,
	})

	if got.ServiceDiscountAmount != 18.8 || got.FinalPayAmount != 169.2 || got.BalanceUsed != 169.2 || got.CashPayAmount != 0 || got.DiscountRate != 0.9 {
		t.Fatalf("expected full member discount with sufficient balance, got %+v", got)
	}
}

func TestCalculateMemberPricingCanExcludeServiceDiscount(t *testing.T) {
	got := CalculateMemberPricing(MemberPricingInput{
		ServiceTotal:            188,
		ServiceDiscountRate:     0.9,
		AvailableBalance:        200,
		ServiceDiscountEligible: false,
	})

	if got.ServiceDiscountAmount != 0 || got.FinalPayAmount != 188 || got.DiscountRate != 1 {
		t.Fatalf("expected excluded service order to remain undiscounted, got %+v", got)
	}
}
