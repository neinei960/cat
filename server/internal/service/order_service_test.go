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

func TestBuildOrderKindRecognizesBoardingOrders(t *testing.T) {
	order := &model.Order{
		ServiceTotal: 455,
		Items: []model.OrderItem{
			{ItemType: 4, Name: "康娜温柔乡 寄养住宿", Quantity: 8, UnitPrice: 65, Amount: 520},
			{ItemType: 6, Name: "会员折扣", Quantity: 1, UnitPrice: -65, Amount: -65},
		},
	}

	if got := buildOrderKind(order); got != "boarding" {
		t.Fatalf("expected boarding order kind, got %q", got)
	}
}

func TestOrderRepositoryFiltersBoardingAndFeedingOrders(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	feedingPlan := seedOrderFilterFeedingPlan(t, customer.ID)
	repo := repository.NewOrderRepository()

	serviceOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:      "TEST-SERVICE-ORDER",
		ShopID:       1,
		CustomerID:   customer.ID,
		ServiceTotal: 88,
		PayAmount:    88,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 101, Name: "基础洗护", Quantity: 1, UnitPrice: 88, Amount: 88},
		},
	})
	feedingOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-FEEDING-ORDER",
		ShopID:        1,
		CustomerID:    customer.ID,
		FeedingPlanID: &feedingPlan.ID,
		ServiceTotal:  120,
		PayAmount:     120,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 201, Name: "上门喂养服务", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})
	boardingOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:      "TEST-BOARDING-ORDER",
		ShopID:       1,
		CustomerID:   customer.ID,
		ServiceTotal: 455,
		PayAmount:    455,
		Items: []model.OrderItem{
			{ItemType: 4, ItemID: 301, Name: "康娜温柔乡 寄养住宿", Quantity: 8, UnitPrice: 65, Amount: 520},
			{ItemType: 6, ItemID: 301, Name: "会员折扣", Quantity: 1, UnitPrice: -65, Amount: -65},
		},
	})

	list, total, err := repo.FindByShopPaged(1, repository.OrderFilter{OrderKind: "feeding"}, 1, 20)
	if err != nil {
		t.Fatalf("filter feeding orders: %v", err)
	}
	assertOrderFilterResult(t, "feeding paged filter", list, total, feedingOrder.ID)

	list, total, err = repo.FindByShopPaged(1, repository.OrderFilter{OrderKind: "boarding"}, 1, 20)
	if err != nil {
		t.Fatalf("filter boarding orders: %v", err)
	}
	assertOrderFilterResult(t, "boarding paged filter", list, total, boardingOrder.ID)

	searchList, searchTotal, err := repo.Search(1, customer.Nickname, repository.OrderFilter{OrderKind: "feeding"}, 1, 20)
	if err != nil {
		t.Fatalf("search feeding orders: %v", err)
	}
	assertOrderFilterResult(t, "feeding search filter", searchList, searchTotal, feedingOrder.ID)

	searchList, searchTotal, err = repo.Search(1, customer.Nickname, repository.OrderFilter{OrderKind: "boarding"}, 1, 20)
	if err != nil {
		t.Fatalf("search boarding orders: %v", err)
	}
	assertOrderFilterResult(t, "boarding search filter", searchList, searchTotal, boardingOrder.ID)

	unfilteredList, unfilteredTotal, err := repo.FindByShopPaged(1, repository.OrderFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("list unfiltered orders: %v", err)
	}
	if unfilteredTotal != 3 || len(unfilteredList) != 3 {
		t.Fatalf("expected 3 unfiltered orders including service=%d, got total=%d len=%d", serviceOrder.ID, unfilteredTotal, len(unfilteredList))
	}
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

type seedOrderFilterOrderInput struct {
	OrderNo       string
	ShopID        uint
	CustomerID    uint
	FeedingPlanID *uint
	ServiceTotal  float64
	PayAmount     float64
	Items         []model.OrderItem
}

func seedOrderFilterCustomer(t *testing.T, shopID uint) model.Customer {
	t.Helper()

	customer := model.Customer{
		ShopID:       shopID,
		Phone:        "13800138111",
		Nickname:     "筛选客户",
		DiscountRate: 1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create filter customer: %v", err)
	}
	return customer
}

func seedOrderFilterFeedingPlan(t *testing.T, customerID uint) model.FeedingPlan {
	t.Helper()

	plan := model.FeedingPlan{
		ShopID:          1,
		CustomerID:      customerID,
		ContactName:     "筛选客户",
		ContactPhone:    "13800138111",
		StartDate:       "2026-04-18",
		EndDate:         "2026-04-20",
		TimeGranularity: model.FeedingWindowMorning,
		Status:          model.FeedingPlanStatusActive,
		TotalAmount:     120,
		UnpaidAmount:    120,
	}
	if err := database.DB.Create(&plan).Error; err != nil {
		t.Fatalf("create feeding plan: %v", err)
	}
	return plan
}

func seedOrderFilterOrder(t *testing.T, input seedOrderFilterOrderInput) model.Order {
	t.Helper()

	order := model.Order{
		OrderNo:       input.OrderNo,
		ShopID:        input.ShopID,
		CustomerID:    &input.CustomerID,
		FeedingPlanID: input.FeedingPlanID,
		TotalAmount:   input.PayAmount,
		ServiceTotal:  input.ServiceTotal,
		PayAmount:     input.PayAmount,
		Status:        1,
		PayStatus:     1,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order %s: %v", input.OrderNo, err)
	}
	if len(input.Items) == 0 {
		return order
	}
	items := make([]model.OrderItem, 0, len(input.Items))
	for _, item := range input.Items {
		item.OrderID = order.ID
		items = append(items, item)
	}
	if err := database.DB.Create(&items).Error; err != nil {
		t.Fatalf("create order items %s: %v", input.OrderNo, err)
	}
	return order
}

func assertOrderFilterResult(t *testing.T, label string, list []model.Order, total int64, wantID uint) {
	t.Helper()

	if total != 1 {
		t.Fatalf("%s: expected total 1, got %d", label, total)
	}
	if len(list) != 1 {
		t.Fatalf("%s: expected 1 order in page, got %d", label, len(list))
	}
	if list[0].ID != wantID {
		t.Fatalf("%s: expected order %d, got %d", label, wantID, list[0].ID)
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
