package handler

import (
	"fmt"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMemberCardHandlerTestDB(t *testing.T) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	database.DB = db
	if err := database.DB.AutoMigrate(
		&model.Customer{},
		&model.MemberCardTemplate{},
		&model.MemberCard{},
		&model.RechargeRecord{},
		&model.Staff{},
		&model.Order{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func TestBalancePaymentSnapshotsOrderMemberBalance(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138002",
		Nickname:      "余额客户",
		MemberBalance: 100,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "金卡",
		MinRecharge:  100,
		DiscountRate: 1,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "金卡",
		Balance:             100,
		TotalSpent:          0,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139002",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:               "TEST-BALANCE-PAYMENT-SNAPSHOT",
		ShopID:                1,
		CustomerID:            &customer.ID,
		StaffID:               &staff.ID,
		ServiceTotal:          100,
		ServiceDiscountAmount: 10,
		PayAmount:             90,
		PayStatus:             0,
		PayMethod:             "",
		TotalAmount:           100,
		Commission:            20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 90, 1); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.MemberBalanceBefore == nil || *saved.MemberBalanceBefore != 100 {
		t.Fatalf("expected member balance before snapshot 100.00, got %#v", saved.MemberBalanceBefore)
	}
	if saved.MemberBalanceAfter == nil || *saved.MemberBalanceAfter != 10 {
		t.Fatalf("expected member balance after snapshot 10.00, got %#v", saved.MemberBalanceAfter)
	}
	if saved.Commission != 18 {
		t.Fatalf("expected commission recalculated from discounted service amount 18.00, got %.2f", saved.Commission)
	}
}

func TestBalancePaymentRejectsAlreadyPaidOrderBeforeDeductingBalance(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138004",
		Nickname:      "已支付客户",
		MemberBalance: 100,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "金卡",
		MinRecharge:  100,
		DiscountRate: 1,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "金卡",
		Balance:             100,
		TotalSpent:          20,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139004",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:      "TEST-BALANCE-ALREADY-PAID",
		ShopID:       1,
		CustomerID:   &customer.ID,
		StaffID:      &staff.ID,
		ServiceTotal: 100,
		PayAmount:    100,
		PayStatus:    1,
		Status:       1,
		PayMethod:    "balance",
		TotalAmount:  100,
		Commission:   20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 100, 1); err == nil {
		t.Fatalf("expected already paid error")
	}

	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Balance != 100 || savedCard.TotalSpent != 20 {
		t.Fatalf("expected card unchanged, got balance %.2f spent %.2f", savedCard.Balance, savedCard.TotalSpent)
	}
}

func TestMixedBalancePaymentUsesBalanceDiscountOnlyForCoveredAmount(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138003",
		Nickname:      "余额不足客户",
		MemberBalance: 60,
		DiscountRate:  0.8,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "八折卡",
		MinRecharge:  100,
		DiscountRate: 0.8,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "八折卡",
		Balance:             60,
		DiscountRate:        0.8,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139003",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:               "TEST-MIXED-BALANCE-PAYMENT",
		ShopID:                1,
		CustomerID:            &customer.ID,
		StaffID:               &staff.ID,
		TotalAmount:           100,
		ServiceTotal:          100,
		ServiceDiscountAmount: 20,
		DiscountAmount:        20,
		DiscountRate:          0.8,
		PayAmount:             80,
		PayStatus:             0,
		Status:                0,
		Commission:            20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	result, err := MixedBalancePayment(1, customer.ID, order.ID, "qrcode", 1)
	if err != nil {
		t.Fatalf("mixed balance payment: %v", err)
	}

	if result.BalanceUsed != 60 {
		t.Fatalf("expected balance used 60.00, got %.2f", result.BalanceUsed)
	}
	if result.CashPayAmount != 25 {
		t.Fatalf("expected cash pay amount 25.00, got %.2f", result.CashPayAmount)
	}
	if result.FinalPayAmount != 85 {
		t.Fatalf("expected final pay amount 85.00, got %.2f", result.FinalPayAmount)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.PayMethod != "mixed_balance" {
		t.Fatalf("expected pay method mixed_balance, got %q", saved.PayMethod)
	}
	if saved.CashPayMethod != "qrcode" {
		t.Fatalf("expected cash pay method qrcode, got %q", saved.CashPayMethod)
	}
	if saved.MemberBalanceUsed != 60 {
		t.Fatalf("expected saved balance used 60.00, got %.2f", saved.MemberBalanceUsed)
	}
	if saved.CashPayAmount != 25 {
		t.Fatalf("expected saved cash pay amount 25.00, got %.2f", saved.CashPayAmount)
	}
	if saved.PayAmount != 85 {
		t.Fatalf("expected saved pay amount 85.00, got %.2f", saved.PayAmount)
	}
	if saved.ServiceDiscountAmount != 15 || saved.DiscountAmount != 15 {
		t.Fatalf("expected partial discount 15.00, got service %.2f total %.2f", saved.ServiceDiscountAmount, saved.DiscountAmount)
	}
	if saved.Commission != 17 {
		t.Fatalf("expected commission recalculated from discounted service amount 17.00, got %.2f", saved.Commission)
	}
	if saved.MemberBalanceBefore == nil || *saved.MemberBalanceBefore != 60 {
		t.Fatalf("expected member balance before 60.00, got %#v", saved.MemberBalanceBefore)
	}
	if saved.MemberBalanceAfter == nil || *saved.MemberBalanceAfter != 0 {
		t.Fatalf("expected member balance after 0.00, got %#v", saved.MemberBalanceAfter)
	}

	var reloadedCard model.MemberCard
	if err := database.DB.First(&reloadedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if reloadedCard.Balance != 0 || reloadedCard.TotalSpent != 60 {
		t.Fatalf("expected card balance 0 and spent 60, got balance %.2f spent %.2f", reloadedCard.Balance, reloadedCard.TotalSpent)
	}
}
