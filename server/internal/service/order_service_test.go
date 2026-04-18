package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteBalancePaidOrderRestoresMemberCardBalance(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedBalancePaidOrderState(t)
	svc := NewOrderService(repository.NewOrderRepository(), nil)

	if err := svc.Delete(state.shopID, state.order.ID); err != nil {
		t.Fatalf("delete order: %v", err)
	}

	assertOrderCardState(t, state.card.ID, 120, 20, 0)
	assertCustomerMemberBalance(t, state.customer.ID, 120)
	assertRechargeRecordDeleted(t, state.orderRecord.ID, true, 40)
	assertRechargeBalanceAfter(t, state.laterRecord.ID, 120)
	assertOrderDeleted(t, state.order.ID, true)
}

func TestRestoreBalancePaidOrderReappliesMemberCardDeduction(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedBalancePaidOrderState(t)
	svc := NewOrderService(repository.NewOrderRepository(), nil)

	if err := svc.Delete(state.shopID, state.order.ID); err != nil {
		t.Fatalf("delete order: %v", err)
	}
	if err := svc.Restore(state.shopID, state.order.ID); err != nil {
		t.Fatalf("restore order: %v", err)
	}

	assertOrderCardState(t, state.card.ID, 60, 20, 60)
	assertCustomerMemberBalance(t, state.customer.ID, 60)
	assertRechargeRecordDeleted(t, state.orderRecord.ID, false, 40)
	assertRechargeBalanceAfter(t, state.laterRecord.ID, 60)
	assertOrderDeleted(t, state.order.ID, false)
}

type balanceOrderTestState struct {
	shopID      uint
	customer    model.Customer
	card        model.MemberCard
	order       model.Order
	orderRecord model.RechargeRecord
	laterRecord model.RechargeRecord
}

func setupOrderServiceTestDB(t *testing.T) {
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
		&model.Order{},
		&model.OrderItem{},
		&model.Pet{},
		&model.Staff{},
		&model.Appointment{},
		&model.AppointmentPet{},
		&model.FeedingPlan{},
		&model.FeedingPlanPet{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func seedBalancePaidOrderState(t *testing.T) balanceOrderTestState {
	t.Helper()

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138000",
		Nickname:      "余额客户",
		MemberBalance: 60,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "金卡",
		MinRecharge:  1000,
		DiscountRate: 1,
		ValidDays:    0,
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
		Balance:             60,
		TotalRecharge:       20,
		TotalSpent:          60,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	if err := database.DB.Model(&customer).Updates(map[string]any{
		"member_card_id": card.ID,
		"member_balance": card.Balance,
	}).Error; err != nil {
		t.Fatalf("link card to customer: %v", err)
	}

	now := time.Now()
	order := model.Order{
		OrderNo:      "TEST-ORDER-DELETE-BALANCE",
		ShopID:       1,
		CustomerID:   &customer.ID,
		TotalAmount:  60,
		ServiceTotal: 60,
		PayAmount:    60,
		PayMethod:    "balance",
		PayStatus:    1,
		Status:       1,
		PayTime:      &now,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	item := model.OrderItem{
		OrderID:   order.ID,
		ItemType:  1,
		ItemID:    101,
		Name:      "测试服务",
		Quantity:  1,
		UnitPrice: 60,
		Amount:    60,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}

	orderRecordCreatedAt := now.Add(-2 * time.Hour)
	orderRecord := model.RechargeRecord{
		ShopID:       1,
		CustomerID:   customer.ID,
		CardID:       card.ID,
		Type:         2,
		Amount:       60,
		BalanceAfter: 40,
		OrderID:      &order.ID,
		Remark:       "订单消费",
	}
	if err := database.DB.Create(&orderRecord).Error; err != nil {
		t.Fatalf("create order recharge record: %v", err)
	}
	if err := database.DB.Model(&orderRecord).Update("created_at", orderRecordCreatedAt).Error; err != nil {
		t.Fatalf("set order record created_at: %v", err)
	}

	laterRecord := model.RechargeRecord{
		ShopID:       1,
		CustomerID:   customer.ID,
		CardID:       card.ID,
		Type:         1,
		Amount:       20,
		BalanceAfter: 60,
		Remark:       "后续充值",
	}
	if err := database.DB.Create(&laterRecord).Error; err != nil {
		t.Fatalf("create later recharge record: %v", err)
	}
	if err := database.DB.Model(&laterRecord).Update("created_at", now.Add(-1*time.Hour)).Error; err != nil {
		t.Fatalf("set later record created_at: %v", err)
	}

	return balanceOrderTestState{
		shopID:      1,
		customer:    customer,
		card:        card,
		order:       order,
		orderRecord: orderRecord,
		laterRecord: laterRecord,
	}
}

func assertOrderCardState(t *testing.T, cardID uint, wantBalance, wantRecharge, wantSpent float64) {
	t.Helper()

	var card model.MemberCard
	if err := database.DB.First(&card, cardID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if card.Balance != wantBalance || card.TotalRecharge != wantRecharge || card.TotalSpent != wantSpent {
		t.Fatalf("unexpected card state: balance=%.2f recharge=%.2f spent=%.2f", card.Balance, card.TotalRecharge, card.TotalSpent)
	}
}

func assertCustomerMemberBalance(t *testing.T, customerID uint, want float64) {
	t.Helper()

	var customer model.Customer
	if err := database.DB.First(&customer, customerID).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	if customer.MemberBalance != want {
		t.Fatalf("unexpected customer balance: got %.2f want %.2f", customer.MemberBalance, want)
	}
}

func assertRechargeRecordDeleted(t *testing.T, recordID uint, wantDeleted bool, wantBalanceAfter float64) {
	t.Helper()

	var record model.RechargeRecord
	if err := database.DB.Unscoped().First(&record, recordID).Error; err != nil {
		t.Fatalf("load recharge record: %v", err)
	}
	if record.BalanceAfter != wantBalanceAfter {
		t.Fatalf("unexpected order recharge balance_after: got %.2f want %.2f", record.BalanceAfter, wantBalanceAfter)
	}
	gotDeleted := record.DeletedAt.Valid
	if gotDeleted != wantDeleted {
		t.Fatalf("unexpected order recharge deleted state: got %v want %v", gotDeleted, wantDeleted)
	}
}

func assertRechargeBalanceAfter(t *testing.T, recordID uint, want float64) {
	t.Helper()

	var record model.RechargeRecord
	if err := database.DB.First(&record, recordID).Error; err != nil {
		t.Fatalf("load later recharge record: %v", err)
	}
	if record.BalanceAfter != want {
		t.Fatalf("unexpected later recharge balance_after: got %.2f want %.2f", record.BalanceAfter, want)
	}
}

func assertOrderDeleted(t *testing.T, orderID uint, wantDeleted bool) {
	t.Helper()

	var order model.Order
	if err := database.DB.Unscoped().First(&order, orderID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if order.DeletedAt.Valid != wantDeleted {
		t.Fatalf("unexpected order deleted state: got %v want %v", order.DeletedAt.Valid, wantDeleted)
	}
}
