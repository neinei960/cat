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

func TestMarkPaidSnapshotsMemberBalanceForCustomerWithActiveCard(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138001",
		Nickname:      "会员客户",
		MemberBalance: 188.8,
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
		Balance:             188.8,
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

	order := model.Order{
		OrderNo:      "TEST-MARK-PAID-SNAPSHOT",
		ShopID:       1,
		CustomerID:   &customer.ID,
		TotalAmount:  50,
		ServiceTotal: 50,
		PayAmount:    50,
		PayStatus:    0,
		Status:       0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	if err := svc.MarkPaid(order.ID, "wechat", ""); err != nil {
		t.Fatalf("mark paid: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.MemberBalanceBefore == nil || *saved.MemberBalanceBefore != 188.8 {
		t.Fatalf("expected member balance before snapshot 188.80, got %#v", saved.MemberBalanceBefore)
	}
	if saved.MemberBalanceAfter == nil || *saved.MemberBalanceAfter != 188.8 {
		t.Fatalf("expected member balance after snapshot 188.80, got %#v", saved.MemberBalanceAfter)
	}
}

func TestUpdateCustomerPetAllowsManualCustomerAndPetSelection(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "024578", Nickname: "Nil"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "奶盖", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	order := model.Order{
		OrderNo:      "TEST-ORDER-CUSTOMER-PET",
		ShopID:       1,
		TotalAmount:  20,
		ProductTotal: 20,
		PayAmount:    20,
		Status:       1,
		PayStatus:    1,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	if err := svc.UpdateCustomerPet(1, order.ID, &customer.ID, &pet.ID); err != nil {
		t.Fatalf("update customer pet: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.CustomerID == nil || *saved.CustomerID != customer.ID {
		t.Fatalf("expected customer_id %d, got %v", customer.ID, saved.CustomerID)
	}
	if saved.PetID == nil || *saved.PetID != pet.ID {
		t.Fatalf("expected pet_id %d, got %v", pet.ID, saved.PetID)
	}
}

func TestUpdateCustomerPetRejectsPetFromAnotherCustomer(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "024578", Nickname: "Nil"}
	otherCustomer := model.Customer{ShopID: 1, Phone: "13900000000", Nickname: "Other"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if err := database.DB.Create(&otherCustomer).Error; err != nil {
		t.Fatalf("create other customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &otherCustomer.ID, Name: "别人的猫", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	order := model.Order{OrderNo: "TEST-ORDER-PET-OWNER", ShopID: 1, Status: 1, PayStatus: 1}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	err := svc.UpdateCustomerPet(1, order.ID, &customer.ID, &pet.ID)
	if err == nil || err.Error() != "猫咪不属于所选客户" {
		t.Fatalf("expected pet ownership error, got %v", err)
	}
}

func TestUpdateCustomerPetCanClearPetWhileKeepingCustomer(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "024578", Nickname: "Nil"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "糊糊", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	order := model.Order{
		OrderNo:    "TEST-ORDER-CLEAR-PET",
		ShopID:     1,
		CustomerID: &customer.ID,
		PetID:      &pet.ID,
		Status:     1,
		PayStatus:  1,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	if err := svc.UpdateCustomerPet(1, order.ID, &customer.ID, nil); err != nil {
		t.Fatalf("clear pet: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.CustomerID == nil || *saved.CustomerID != customer.ID {
		t.Fatalf("expected customer_id %d, got %v", customer.ID, saved.CustomerID)
	}
	if saved.PetID != nil {
		t.Fatalf("expected cleared pet_id, got %v", *saved.PetID)
	}
}

func TestCancelFeedingOrderCancelsLinkedPlan(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	plan := seedOrderFilterFeedingPlan(t, customer.ID)
	visit := model.FeedingVisit{
		ShopID:        1,
		FeedingPlanID: plan.ID,
		ScheduledDate: "2026-05-02",
		WindowCode:    model.FeedingWindowAllDay,
		Status:        model.FeedingVisitStatusPending,
		VisitPrice:    95,
	}
	if err := database.DB.Create(&visit).Error; err != nil {
		t.Fatalf("create feeding visit: %v", err)
	}
	order := model.Order{
		OrderNo:       "TEST-CANCEL-FEEDING-ORDER",
		ShopID:        1,
		CustomerID:    &customer.ID,
		FeedingPlanID: &plan.ID,
		TotalAmount:   95,
		ServiceTotal:  95,
		PayAmount:     95,
		Status:        0,
		PayStatus:     0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create feeding order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	if err := svc.Cancel(order.ID); err != nil {
		t.Fatalf("cancel feeding order: %v", err)
	}

	var savedPlan model.FeedingPlan
	if err := database.DB.First(&savedPlan, plan.ID).Error; err != nil {
		t.Fatalf("load feeding plan: %v", err)
	}
	if savedPlan.Status != model.FeedingPlanStatusCancelled {
		t.Fatalf("expected linked feeding plan cancelled, got %q", savedPlan.Status)
	}
	if savedPlan.UnpaidAmount != 0 {
		t.Fatalf("expected linked feeding plan unpaid amount 0, got %.2f", savedPlan.UnpaidAmount)
	}
	var savedVisit model.FeedingVisit
	if err := database.DB.First(&savedVisit, visit.ID).Error; err != nil {
		t.Fatalf("load feeding visit: %v", err)
	}
	if savedVisit.Status != model.FeedingVisitStatusCancelled {
		t.Fatalf("expected pending feeding visit cancelled, got %q", savedVisit.Status)
	}
}

func TestDeleteFeedingOrderCancelsLinkedPlan(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	plan := seedOrderFilterFeedingPlan(t, customer.ID)
	order := model.Order{
		OrderNo:       "TEST-DELETE-FEEDING-ORDER",
		ShopID:        1,
		CustomerID:    &customer.ID,
		FeedingPlanID: &plan.ID,
		TotalAmount:   95,
		ServiceTotal:  95,
		PayAmount:     95,
		Status:        0,
		PayStatus:     0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create feeding order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	if err := svc.Delete(1, order.ID); err != nil {
		t.Fatalf("delete feeding order: %v", err)
	}

	var savedPlan model.FeedingPlan
	if err := database.DB.First(&savedPlan, plan.ID).Error; err != nil {
		t.Fatalf("load feeding plan: %v", err)
	}
	if savedPlan.Status != model.FeedingPlanStatusCancelled {
		t.Fatalf("expected linked feeding plan cancelled, got %q", savedPlan.Status)
	}
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

func TestGetByIDBuildsBoardingPetSummaryFromBoardingPets(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	petA := model.Pet{
		ShopID:     1,
		CustomerID: &customer.ID,
		Name:       "茶茶",
		Species:    "猫",
	}
	if err := database.DB.Create(&petA).Error; err != nil {
		t.Fatalf("create boarding pet A: %v", err)
	}
	petB := model.Pet{
		ShopID:     1,
		CustomerID: &customer.ID,
		Name:       "五角",
		Species:    "猫",
	}
	if err := database.DB.Create(&petB).Error; err != nil {
		t.Fatalf("create boarding pet B: %v", err)
	}

	order := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:      "TEST-BOARDING-PET-SUMMARY",
		ShopID:       1,
		CustomerID:   customer.ID,
		ServiceTotal: 240,
		PayAmount:    240,
		Items: []model.OrderItem{
			{ItemType: 4, ItemID: 301, Name: "房间1 · 茶茶的空中后花园 寄养住宿", Quantity: 2, UnitPrice: 120, Amount: 240},
		},
	})

	boardingOrder := model.BoardingOrder{
		ShopID:     1,
		OrderID:    &order.ID,
		CustomerID: customer.ID,
		StaffID:    1,
		CabinetID:  1,
		CheckInAt:  "2026-04-20",
		CheckOutAt: "2026-04-22",
		Nights:     2,
		BaseAmount: 240,
		PayAmount:  240,
		Status:     model.BoardingOrderStatusPendingCheckin,
	}
	if err := database.DB.Create(&boardingOrder).Error; err != nil {
		t.Fatalf("create boarding order: %v", err)
	}
	boardingPets := []model.BoardingOrderPet{
		{BoardingOrderID: boardingOrder.ID, PetID: petA.ID, PetNameSnapshot: "茶茶"},
		{BoardingOrderID: boardingOrder.ID, PetID: petB.ID, PetNameSnapshot: "五角"},
	}
	if err := database.DB.Create(&boardingPets).Error; err != nil {
		t.Fatalf("create boarding pets: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	savedOrder, err := svc.GetByID(order.ID)
	if err != nil {
		t.Fatalf("get order by id: %v", err)
	}

	if savedOrder.PetSummary != "茶茶 / 五角" {
		t.Fatalf("expected boarding pet summary \"茶茶 / 五角\", got %q", savedOrder.PetSummary)
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

func TestStatsRepositoryGetOverviewByRangeIncludesPaymentBreakdown(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	appointment := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      1,
		Date:       "2026-04-20",
		StartTime:  "10:00",
		EndTime:    "11:00",
		Status:     3,
		Source:     2,
	}
	if err := database.DB.Create(&appointment).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-WECHAT",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  100,
		PayAmount:     100,
		PayMethod:     "wechat",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "洗护", Quantity: 1, UnitPrice: 100, Amount: 100},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-MEITUAN",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  230,
		PayAmount:     230,
		PayMethod:     "meituan",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9002, Name: "美容", Quantity: 1, UnitPrice: 230, Amount: 230},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-BALANCE",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  80,
		PayAmount:     80,
		PayMethod:     "balance",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9003, Name: "刷牙", Quantity: 1, UnitPrice: 80, Amount: 80},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-MIXED-BALANCE",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  60,
		PayAmount:     60,
		PayMethod:     "mixed_balance",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9006, Name: "补差服务", Quantity: 1, UnitPrice: 60, Amount: 60},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-QRCODE",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  40,
		PayAmount:     40,
		PayMethod:     "qrcode",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9005, Name: "年卡服务", Quantity: 1, UnitPrice: 40, Amount: 40},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-OTHER",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  50,
		PayAmount:     50,
		PayMethod:     "cash",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9004, Name: "修甲", Quantity: 1, UnitPrice: 50, Amount: 50},
		},
	})

	stats, err := repository.NewStatsRepository().GetOverviewByRange(1, "2026-04-20", "2026-04-20")
	if err != nil {
		t.Fatalf("get overview by range: %v", err)
	}

	if got := roundOrderAmount(stats.TodayRevenue); got != 560 {
		t.Fatalf("expected total revenue 560, got %.2f", got)
	}
	if got := roundOrderAmount(stats.MonthRevenue); got != 560 {
		t.Fatalf("expected range revenue 560 in month_revenue field, got %.2f", got)
	}

	breakdown := make(map[string]float64)
	breakdownLabels := make(map[string]string)
	for _, item := range stats.PaymentBreakdown {
		breakdown[item.Key] = item.Amount
		breakdownLabels[item.Key] = item.Label
	}
	if got := roundOrderAmount(breakdown["wechat"]); got != 100 {
		t.Fatalf("expected wechat 100, got %.2f", got)
	}
	if got := roundOrderAmount(breakdown["meituan"]); got != 230 {
		t.Fatalf("expected meituan 230, got %.2f", got)
	}
	if got := roundOrderAmount(breakdown["balance"]); got != 80 {
		t.Fatalf("expected balance 80, got %.2f", got)
	}
	if got := breakdownLabels["balance"]; got != "会员" {
		t.Fatalf("expected balance label 会员, got %q", got)
	}
	if got := roundOrderAmount(breakdown["mixed_balance"]); got != 60 {
		t.Fatalf("expected mixed_balance 60, got %.2f", got)
	}
	if got := breakdownLabels["mixed_balance"]; got != "会员+补差" {
		t.Fatalf("expected mixed_balance label 会员+补差, got %q", got)
	}
	if got := roundOrderAmount(breakdown["qrcode"]); got != 40 {
		t.Fatalf("expected qrcode 40, got %.2f", got)
	}
	if got := roundOrderAmount(breakdown["other"]); got != 50 {
		t.Fatalf("expected other 50, got %.2f", got)
	}
}

func TestStatsRepositoryGetStaffPerformanceUsesStoredOrderCommission(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	staff := model.Staff{
		ShopID:                1,
		Phone:                 "13800139200",
		PasswordHash:          "hash",
		Name:                  "乐乐",
		Role:                  model.StaffRoleStaff,
		Status:                1,
		CommissionRate:        20,
		ProductCommissionRate: 0,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	appointment := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      1,
		StaffID:    &staff.ID,
		Date:       "2026-04-20",
		StartTime:  "10:00",
		EndTime:    "11:00",
		Status:     3,
		Source:     2,
	}
	if err := database.DB.Create(&appointment).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-STAFF-PERF-ZERO-COMMISSION",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		StaffID:       &staff.ID,
		ServiceTotal:  100,
		PayAmount:     100,
		PayMethod:     "wechat",
		Commission:    0,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "洗护", Quantity: 1, UnitPrice: 100, Amount: 100},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-STAFF-PERF-STORED-COMMISSION",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		StaffID:       &staff.ID,
		ServiceTotal:  80,
		ProductTotal:  20,
		PayAmount:     100,
		PayMethod:     "wechat",
		Commission:    12,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9002, Name: "刷牙", Quantity: 1, UnitPrice: 80, Amount: 80},
			{ItemType: 2, ItemID: 9102, Name: "商品", Quantity: 1, UnitPrice: 20, Amount: 20},
		},
	})

	perfs, err := repository.NewStatsRepository().GetStaffPerformance(1, "2026-04-20", "2026-04-20")
	if err != nil {
		t.Fatalf("get staff performance: %v", err)
	}
	if len(perfs) != 1 {
		t.Fatalf("expected one staff performance row, got %d: %+v", len(perfs), perfs)
	}
	if got := roundOrderAmount(perfs[0].Commission); got != 12 {
		t.Fatalf("expected stored commission sum 12.00, got %.2f", got)
	}
}

func TestStatsRepositoryGetStaffCommissionDetailsIncludesFormula(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	staff := model.Staff{
		ShopID:                1,
		Phone:                 "13800139201",
		PasswordHash:          "hash",
		Name:                  "乐乐",
		Role:                  model.StaffRoleStaff,
		Status:                1,
		CommissionRate:        20,
		ProductCommissionRate: 0,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	appointment := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      1,
		StaffID:    &staff.ID,
		Date:       "2026-04-20",
		StartTime:  "10:00",
		EndTime:    "11:00",
		Status:     3,
		Source:     2,
	}
	if err := database.DB.Create(&appointment).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-STAFF-COMMISSION-MEITUAN",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		StaffID:       &staff.ID,
		ServiceTotal:  100,
		PayAmount:     100,
		PayMethod:     "meituan",
		Commission:    18,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "洗护", Quantity: 1, UnitPrice: 100, Amount: 100},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-STAFF-COMMISSION-PRODUCT",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		StaffID:       &staff.ID,
		ServiceTotal:  100,
		ProductTotal:  75,
		PayAmount:     175,
		PayMethod:     "wechat",
		Commission:    20,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9002, Name: "洗护", Quantity: 1, UnitPrice: 100, Amount: 100},
			{ItemType: 2, ItemID: 9102, Name: "商品", Quantity: 1, UnitPrice: 75, Amount: 75},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-STAFF-COMMISSION-ZERO",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		StaffID:       &staff.ID,
		ServiceTotal:  80,
		PayAmount:     80,
		PayMethod:     "wechat",
		Commission:    0,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9003, Name: "洗护", Quantity: 1, UnitPrice: 80, Amount: 80},
		},
	})

	details, err := repository.NewStatsRepository().GetStaffCommissionDetails(1, staff.ID, "2026-04-20", "2026-04-20")
	if err != nil {
		t.Fatalf("get staff commission details: %v", err)
	}
	if len(details) != 2 {
		t.Fatalf("expected two detail rows, got %d: %+v", len(details), details)
	}
	if details[0].Formula != "¥100.00 × 0.9 × 20% = ¥18.00" {
		t.Fatalf("unexpected meituan formula: %s", details[0].Formula)
	}
	if details[1].Formula != "(¥175.00 - 商品¥75.00) × 20% = ¥20.00" {
		t.Fatalf("unexpected product formula: %s", details[1].Formula)
	}
}

func TestStatsRepositoryGetOverviewByRangeUsesRangeCollectionStats(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	appointment := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      1,
		Date:       "2026-04-20",
		StartTime:  "10:00",
		EndTime:    "11:00",
		Status:     3,
		Source:     2,
	}
	if err := database.DB.Create(&appointment).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-DASHBOARD-RANGE-ORDER",
		ShopID:        1,
		CustomerID:    customer.ID,
		AppointmentID: &appointment.ID,
		ServiceTotal:  120,
		PayAmount:     120,
		PayMethod:     "wechat",
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9101, Name: "洗护", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})
	if err := database.DB.Create(&model.RechargeRecord{
		ShopID: 1, CustomerID: customer.ID, CardID: 1, Type: 1, Amount: 300,
		Model: gorm.Model{CreatedAt: time.Date(2026, 4, 20, 12, 0, 0, 0, time.Local)},
	}).Error; err != nil {
		t.Fatalf("create in-range recharge: %v", err)
	}
	if err := database.DB.Create(&model.RechargeRecord{
		ShopID: 1, CustomerID: customer.ID, CardID: 1, Type: 1, Amount: 900,
		Model: gorm.Model{CreatedAt: time.Date(2026, 5, 2, 12, 0, 0, 0, time.Local)},
	}).Error; err != nil {
		t.Fatalf("create outside recharge: %v", err)
	}

	stats, err := repository.NewStatsRepository().GetOverviewByRange(1, "2026-04-20", "2026-04-20")
	if err != nil {
		t.Fatalf("get overview by range: %v", err)
	}

	if got := roundOrderAmount(stats.MonthRevenue); got != 120 {
		t.Fatalf("expected range revenue 120, got %.2f", got)
	}
	if got := roundOrderAmount(stats.MonthRecharge); got != 300 {
		t.Fatalf("expected range recharge 300, got %.2f", got)
	}
	if got := roundOrderAmount(stats.MonthCollection); got != 420 {
		t.Fatalf("expected range collection 420, got %.2f", got)
	}
}

func TestStatsRepositoryGetOverviewByRangeCountsServedCustomers(t *testing.T) {
	setupOrderServiceTestDB(t)

	customers := []model.Customer{
		{ShopID: 1, Phone: "13800139001", Nickname: "本月服务客户A", DiscountRate: 1},
		{ShopID: 1, Phone: "13800139002", Nickname: "本月服务客户B", DiscountRate: 1},
		{ShopID: 1, Phone: "13800139003", Nickname: "取消客户", DiscountRate: 1},
		{ShopID: 1, Phone: "13800139004", Nickname: "未服务客户", DiscountRate: 1},
	}
	if err := database.DB.Create(&customers).Error; err != nil {
		t.Fatalf("create customers: %v", err)
	}

	appointments := []model.Appointment{
		{ShopID: 1, CustomerID: customers[0].ID, PetID: 1, Date: "2026-04-03", StartTime: "10:00", EndTime: "11:00", Status: 3, Source: 2},
		{ShopID: 1, CustomerID: customers[0].ID, PetID: 1, Date: "2026-04-08", StartTime: "10:00", EndTime: "11:00", Status: 7, Source: 2},
		{ShopID: 1, CustomerID: customers[1].ID, PetID: 1, Date: "2026-04-12", StartTime: "10:00", EndTime: "11:00", Status: 1, Source: 2},
		{ShopID: 1, CustomerID: customers[2].ID, PetID: 1, Date: "2026-04-13", StartTime: "10:00", EndTime: "11:00", Status: 4, Source: 2},
		{ShopID: 1, CustomerID: customers[2].ID, PetID: 1, Date: "2026-04-14", StartTime: "10:00", EndTime: "11:00", Status: 5, Source: 2},
	}
	if err := database.DB.Create(&appointments).Error; err != nil {
		t.Fatalf("create appointments: %v", err)
	}

	stats, err := repository.NewStatsRepository().GetOverviewByRange(1, "2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatalf("get overview by range: %v", err)
	}
	if stats.TotalCustomers != 2 {
		t.Fatalf("expected 2 served customers, got %d", stats.TotalCustomers)
	}
}

func TestStatsRepositoryGetProjectRevenueTreeGroupsAllBusinessLines(t *testing.T) {
	setupOrderServiceTestDB(t)

	today := time.Now().Format("2006-01-02")
	customer := seedOrderFilterCustomer(t, 1)
	rootCategory := model.ServiceCategory{ShopID: 1, Name: "洗护", Status: 1}
	if err := database.DB.Create(&rootCategory).Error; err != nil {
		t.Fatalf("create service root category: %v", err)
	}
	childCategory := model.ServiceCategory{ShopID: 1, ParentID: &rootCategory.ID, Name: "基础护理", Status: 1}
	if err := database.DB.Create(&childCategory).Error; err != nil {
		t.Fatalf("create service child category: %v", err)
	}
	bathService := model.Service{ShopID: 1, Name: "日常护理", CategoryID: &childCategory.ID, BasePrice: 100, Duration: 60, Status: 1}
	if err := database.DB.Create(&bathService).Error; err != nil {
		t.Fatalf("create service: %v", err)
	}
	productCategory := model.ProductCategory{ShopID: 1, Name: "驱虫", Status: 1}
	if err := database.DB.Create(&productCategory).Error; err != nil {
		t.Fatalf("create product category: %v", err)
	}
	product := model.Product{ShopID: 1, CategoryID: productCategory.ID, Name: "大宠爱", Status: 1}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	feedingPlan := seedOrderFilterFeedingPlan(t, customer.ID)
	feedingID := feedingPlan.ID

	serviceOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:    "TEST-PROJECT-SERVICE",
		ShopID:     1,
		CustomerID: customer.ID,
		PayAmount:  90,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: bathService.ID, Name: "日常护理", Quantity: 1, UnitPrice: 100, Amount: 100},
		},
	})
	if err := database.DB.Model(&model.Order{}).Where("id = ?", serviceOrder.ID).Updates(map[string]interface{}{
		"total_amount":               100,
		"service_total":              100,
		"discount_amount":            10,
		"service_discount_amount":    10,
		"product_discount_amount":    0,
		"pay_amount":                 90,
		"appointment_deposit_amount": 0,
	}).Error; err != nil {
		t.Fatalf("update service discount: %v", err)
	}
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:    "TEST-PROJECT-PRODUCT",
		ShopID:     1,
		CustomerID: customer.ID,
		PayAmount:  136,
		Items: []model.OrderItem{
			{ItemType: 2, ItemID: product.ID, Name: "大宠爱", Quantity: 2, UnitPrice: 68, Amount: 136},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-PROJECT-FEEDING",
		ShopID:        1,
		CustomerID:    customer.ID,
		FeedingPlanID: &feedingID,
		PayAmount:     120,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 0, Name: "花花 · 上门喂养 × 1天", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})
	seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:    "TEST-PROJECT-BOARDING",
		ShopID:     1,
		CustomerID: customer.ID,
		PayAmount:  240,
		Items: []model.OrderItem{
			{ItemType: 4, ItemID: 301, Name: "房间1 · 寄养住宿", Quantity: 2, UnitPrice: 120, Amount: 240},
			{ItemType: 6, ItemID: 301, Name: "定金抵扣", Quantity: 1, UnitPrice: -200, Amount: -200},
		},
	})

	tree, err := repository.NewStatsRepository().GetProjectRevenueTree(1, today, today)
	if err != nil {
		t.Fatalf("get project revenue tree: %v", err)
	}

	roots := make(map[string]repository.ProjectRevenueNode)
	for _, node := range tree {
		roots[node.Name] = node
	}
	if got := roundOrderAmount(roots["服务"].Revenue); got != 90 {
		t.Fatalf("expected service root revenue 90 after discount, got %.2f", got)
	}
	if got := roundOrderAmount(roots["商品"].Revenue); got != 136 {
		t.Fatalf("expected product root revenue 136, got %.2f", got)
	}
	if got := roundOrderAmount(roots["上门喂养"].Revenue); got != 120 {
		t.Fatalf("expected feeding root revenue 120, got %.2f", got)
	}
	if got := roundOrderAmount(roots["寄养"].Revenue); got != 40 {
		t.Fatalf("expected boarding root revenue 40 after deposit deduction, got %.2f", got)
	}
	if !projectTreeContainsPath(roots["服务"], "洗护", "基础护理", "日常护理") {
		t.Fatalf("expected service category path in tree: %+v", roots["服务"])
	}
	if !projectTreeContainsPath(roots["商品"], "驱虫", "大宠爱") {
		t.Fatalf("expected product category path in tree: %+v", roots["商品"])
	}
	if !projectTreeContainsPath(roots["寄养"], "住宿", "寄养住宿") {
		t.Fatalf("expected boarding path in tree: %+v", roots["寄养"])
	}
	if projectTreeContainsPath(roots["寄养"], "优惠/抵扣") {
		t.Fatalf("expected boarding deductions to be excluded from project revenue tree: %+v", roots["寄养"])
	}
}

func projectTreeContainsPath(node repository.ProjectRevenueNode, names ...string) bool {
	current := node
	for _, name := range names {
		found := false
		for _, child := range current.Children {
			if child.Name == name {
				current = child
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestOrderRepositoryFiltersByServiceCategory(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := seedOrderFilterCustomer(t, 1)
	repo := repository.NewOrderRepository()

	groomingRoot := model.ServiceCategory{
		ShopID: 1,
		Name:   "洗护",
		Status: 1,
	}
	if err := database.DB.Create(&groomingRoot).Error; err != nil {
		t.Fatalf("create grooming root category: %v", err)
	}

	bathChild := model.ServiceCategory{
		ShopID:   1,
		ParentID: &groomingRoot.ID,
		Name:     "洗澡",
		Status:   1,
	}
	if err := database.DB.Create(&bathChild).Error; err != nil {
		t.Fatalf("create bath child category: %v", err)
	}

	medicalRoot := model.ServiceCategory{
		ShopID: 1,
		Name:   "医疗",
		Status: 1,
	}
	if err := database.DB.Create(&medicalRoot).Error; err != nil {
		t.Fatalf("create medical root category: %v", err)
	}

	bathService := model.Service{
		ShopID:     1,
		Name:       "香波洗护",
		CategoryID: &bathChild.ID,
		BasePrice:  88,
		Duration:   60,
		Status:     1,
	}
	if err := database.DB.Create(&bathService).Error; err != nil {
		t.Fatalf("create bath service: %v", err)
	}

	medicalService := model.Service{
		ShopID:     1,
		Name:       "基础问诊",
		CategoryID: &medicalRoot.ID,
		BasePrice:  120,
		Duration:   45,
		Status:     1,
	}
	if err := database.DB.Create(&medicalService).Error; err != nil {
		t.Fatalf("create medical service: %v", err)
	}

	groomingOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:      "TEST-SERVICE-CATEGORY-GROOMING",
		ShopID:       1,
		CustomerID:   customer.ID,
		ServiceTotal: 88,
		PayAmount:    88,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: bathService.ID, Name: "香波洗护", Quantity: 1, UnitPrice: 88, Amount: 88},
		},
	})
	medicalOrder := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:      "TEST-SERVICE-CATEGORY-MEDICAL",
		ShopID:       1,
		CustomerID:   customer.ID,
		ServiceTotal: 120,
		PayAmount:    120,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: medicalService.ID, Name: "基础问诊", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})

	list, total, err := repo.FindByShopPaged(1, repository.OrderFilter{CategoryID: groomingRoot.ID}, 1, 20)
	if err != nil {
		t.Fatalf("filter orders by service category: %v", err)
	}
	assertOrderFilterResult(t, "service category paged filter", list, total, groomingOrder.ID)

	searchList, searchTotal, err := repo.Search(1, customer.Nickname, repository.OrderFilter{CategoryID: groomingRoot.ID}, 1, 20)
	if err != nil {
		t.Fatalf("search orders by service category: %v", err)
	}
	assertOrderFilterResult(t, "service category search filter", searchList, searchTotal, groomingOrder.ID)

	if groomingOrder.ID == medicalOrder.ID {
		t.Fatalf("expected distinct seeded orders")
	}
}

func TestCreateFromAppointmentAppliesDepositDeductionForNonMember(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138112",
		CustomerNickname:   "预约金客户",
		AppointmentAmount:  100,
		AppointmentDeposit: 30,
	})
	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())

	order, err := svc.CreateFromAppointment(state.appointment.ID, false)
	if err != nil {
		t.Fatalf("create order from appointment: %v", err)
	}

	if got := roundOrderAmount(order.PayAmount); got != 70 {
		t.Fatalf("expected pay amount 70, got %.2f", got)
	}
	if got := roundOrderAmount(order.AppointmentDepositAmount); got != 30 {
		t.Fatalf("expected appointment deposit amount 30, got %.2f", got)
	}
	if got := roundOrderAmount(order.AppointmentDepositDeductionAmount); got != 30 {
		t.Fatalf("expected appointment deposit deduction 30, got %.2f", got)
	}
	if order.AppointmentIsLate {
		t.Fatalf("expected appointment late false by default")
	}
}

func TestCreateFromAppointmentStoresCommissionForAssignedStaff(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:            1,
		CustomerPhone:     "13800138116",
		CustomerNickname:  "提成预约客户",
		AppointmentAmount: 128,
	})
	staff := model.Staff{
		ShopID:                1,
		Phone:                 "13800139116",
		PasswordHash:          "hash",
		Name:                  "乐乐",
		Role:                  model.StaffRoleStaff,
		Status:                1,
		CommissionRate:        20,
		ProductCommissionRate: 0,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	if err := database.DB.Model(&model.Appointment{}).Where("id = ?", state.appointment.ID).Update("staff_id", staff.ID).Error; err != nil {
		t.Fatalf("assign appointment staff: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	order, err := svc.CreateFromAppointment(state.appointment.ID, false)
	if err != nil {
		t.Fatalf("create order from appointment: %v", err)
	}

	if order.StaffID == nil || *order.StaffID != staff.ID {
		t.Fatalf("expected order staff %d, got %+v", staff.ID, order.StaffID)
	}
	if got := roundOrderAmount(order.Commission); got != 25.6 {
		t.Fatalf("expected commission 25.60, got %.2f", got)
	}
}

func TestCreateFromAppointmentCalculatesCommissionFromDiscountedServiceAmount(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:            1,
		CustomerPhone:     "13800138117",
		CustomerNickname:  "会员提成预约客户",
		AppointmentAmount: 100,
	})
	staff := model.Staff{
		ShopID:                1,
		Phone:                 "13800139117",
		PasswordHash:          "hash",
		Name:                  "乐乐",
		Role:                  model.StaffRoleStaff,
		Status:                1,
		CommissionRate:        20,
		ProductCommissionRate: 0,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	if err := database.DB.Model(&model.Customer{}).Where("id = ?", state.customer.ID).Update("discount_rate", 0.9).Error; err != nil {
		t.Fatalf("update customer discount: %v", err)
	}
	if err := database.DB.Model(&model.Appointment{}).Where("id = ?", state.appointment.ID).Update("staff_id", staff.ID).Error; err != nil {
		t.Fatalf("assign appointment staff: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	order, err := svc.CreateFromAppointment(state.appointment.ID, false)
	if err != nil {
		t.Fatalf("create order from appointment: %v", err)
	}

	if got := roundOrderAmount(order.PayAmount); got != 90 {
		t.Fatalf("expected discounted pay amount 90.00, got %.2f", got)
	}
	if got := roundOrderAmount(order.Commission); got != 18 {
		t.Fatalf("expected discounted commission 18.00, got %.2f", got)
	}
}

func TestCreateFromAppointmentAppliesReducedDepositDeductionWhenLate(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138115",
		CustomerNickname:   "迟到预约金客户",
		AppointmentAmount:  100,
		AppointmentDeposit: 30,
	})
	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())

	order, err := svc.CreateFromAppointment(state.appointment.ID, true)
	if err != nil {
		t.Fatalf("create late order from appointment: %v", err)
	}

	if got := roundOrderAmount(order.PayAmount); got != 79 {
		t.Fatalf("expected pay amount 79 after late deduction, got %.2f", got)
	}
	if got := roundOrderAmount(order.AppointmentDepositDeductionAmount); got != 21 {
		t.Fatalf("expected appointment deposit deduction 21 after late penalty, got %.2f", got)
	}
	if !order.AppointmentIsLate {
		t.Fatalf("expected appointment late true")
	}
}

func TestCreateFromAppointmentSkipsDepositDeductionForMember(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138113",
		CustomerNickname:   "会员预约金客户",
		AppointmentAmount:  100,
		AppointmentDeposit: 30,
		WithMemberCard:     true,
	})
	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())

	order, err := svc.CreateFromAppointment(state.appointment.ID, false)
	if err != nil {
		t.Fatalf("create member order from appointment: %v", err)
	}

	if got := roundOrderAmount(order.PayAmount); got != 100 {
		t.Fatalf("expected member pay amount 100, got %.2f", got)
	}
	if got := roundOrderAmount(order.AppointmentDepositAmount); got != 30 {
		t.Fatalf("expected appointment deposit amount 30, got %.2f", got)
	}
	if got := roundOrderAmount(order.AppointmentDepositDeductionAmount); got != 0 {
		t.Fatalf("expected appointment deposit deduction 0 for member, got %.2f", got)
	}
}

func TestUpdateDraftReappliesAppointmentDepositDeduction(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138114",
		CustomerNickname:   "改单预约金客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 20,
	})
	order := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-APPOINTMENT-DEPOSIT-DRAFT",
		ShopID:        1,
		CustomerID:    state.customer.ID,
		AppointmentID: &state.appointment.ID,
		ServiceTotal:  120,
		PayAmount:     100,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	patch := &model.Order{
		CustomerID:            &state.customer.ID,
		AppointmentID:         &state.appointment.ID,
		AppointmentIsLate:     false,
		TotalAmount:           120,
		ServiceTotal:          120,
		DiscountRate:          1,
		DiscountAmount:        0,
		ServiceDiscountAmount: 0,
		ProductDiscountAmount: 0,
		PayAmount:             120,
	}
	items := []model.OrderItem{
		{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
	}

	if err := svc.UpdateDraft(1, model.StaffRoleManager, order.ID, patch, items); err != nil {
		t.Fatalf("update appointment order draft: %v", err)
	}

	updated, err := svc.GetByID(order.ID)
	if err != nil {
		t.Fatalf("reload updated order: %v", err)
	}
	if got := roundOrderAmount(updated.PayAmount); got != 100 {
		t.Fatalf("expected updated pay amount 100, got %.2f", got)
	}
	if got := roundOrderAmount(updated.AppointmentDepositAmount); got != 20 {
		t.Fatalf("expected updated appointment deposit amount 20, got %.2f", got)
	}
	if got := roundOrderAmount(updated.AppointmentDepositDeductionAmount); got != 20 {
		t.Fatalf("expected updated appointment deposit deduction 20, got %.2f", got)
	}
}

func TestUpdateDraftPersistsChangedStaffIDWithPreloadedStaff(t *testing.T) {
	setupOrderServiceTestDB(t)

	originalStaff := model.Staff{
		ShopID: 1,
		Phone:  "13800138190",
		Name:   "原员工",
		Role:   model.StaffRoleStaff,
		Status: 1,
	}
	if err := database.DB.Create(&originalStaff).Error; err != nil {
		t.Fatalf("create original staff: %v", err)
	}
	nextStaff := model.Staff{
		ShopID: 1,
		Phone:  "13800138191",
		Name:   "新员工",
		Role:   model.StaffRoleStaff,
		Status: 1,
	}
	if err := database.DB.Create(&nextStaff).Error; err != nil {
		t.Fatalf("create next staff: %v", err)
	}

	order := model.Order{
		OrderNo:      "TEST-UPDATE-DRAFT-STAFF",
		ShopID:       1,
		StaffID:      &originalStaff.ID,
		TotalAmount:  99,
		ProductTotal: 99,
		PayAmount:    99,
		Status:       1,
		PayStatus:    1,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	items := []model.OrderItem{
		{OrderID: order.ID, ItemType: 2, ItemID: 287, Name: "商品", Quantity: 1, UnitPrice: 99, Amount: 99},
	}
	if err := database.DB.Create(&items).Error; err != nil {
		t.Fatalf("create order items: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), nil)
	patch := &model.Order{
		StaffID:      &nextStaff.ID,
		TotalAmount:  99,
		ProductTotal: 99,
		DiscountRate: 1,
		PayAmount:    99,
	}
	patchItems := []model.OrderItem{
		{ItemType: 2, ItemID: 287, Name: "商品", Quantity: 1, UnitPrice: 99, Amount: 99},
	}
	if err := svc.UpdateDraft(1, model.StaffRoleManager, order.ID, patch, patchItems); err != nil {
		t.Fatalf("update draft staff: %v", err)
	}

	updated, err := svc.GetByID(order.ID)
	if err != nil {
		t.Fatalf("reload updated order: %v", err)
	}
	if updated.StaffID == nil || *updated.StaffID != nextStaff.ID {
		t.Fatalf("expected staff id %d, got %#v", nextStaff.ID, updated.StaffID)
	}
	if updated.Staff == nil || updated.Staff.Name != nextStaff.Name {
		t.Fatalf("expected preloaded staff %q, got %#v", nextStaff.Name, updated.Staff)
	}
}

func TestCreateSplitFromAppointmentSyncsOverriddenServicesToAppointment(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138118",
		CustomerNickname:   "同步预约项目客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 0,
	})
	changedService := model.Service{
		ShopID:    1,
		Name:      "实际洗护项目",
		BasePrice: 168,
		Duration:  75,
		Status:    1,
	}
	if err := database.DB.Create(&changedService).Error; err != nil {
		t.Fatalf("create changed service: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	_, err := svc.CreateSplitFromAppointment(state.appointment.ID, false, nil, map[uint]PetOverrideData{
		state.pet.ID: {
			Services: []ServiceOverride{{
				ServiceID:   changedService.ID,
				ServiceName: changedService.Name,
				Price:       168,
				Duration:    75,
			}},
		},
	})
	if err != nil {
		t.Fatalf("create split order from appointment: %v", err)
	}

	assertAppointmentPetServices(t, state.appointment.ID, state.pet.ID, []wantAppointmentService{{
		ServiceID:   changedService.ID,
		ServiceName: changedService.Name,
		Price:       168,
		Duration:    75,
	}})
}

func TestCreateSplitFromAppointmentUsesStaffOverride(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138121",
		CustomerNickname:   "合单员工客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 0,
	})
	overrideStaff := model.Staff{
		ShopID: 1,
		Name:   "改后员工",
		Phone:  "13900138121",
		Role:   model.StaffRoleStaff,
		Status: 1,
	}
	if err := database.DB.Create(&overrideStaff).Error; err != nil {
		t.Fatalf("create override staff: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	orders, err := svc.CreateSplitFromAppointment(state.appointment.ID, false, &overrideStaff.ID, nil)
	if err != nil {
		t.Fatalf("create split order from appointment: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected one order, got %d", len(orders))
	}
	if orders[0].StaffID == nil || *orders[0].StaffID != overrideStaff.ID {
		t.Fatalf("expected staff override %d, got %#v", overrideStaff.ID, orders[0].StaffID)
	}
}

func TestCreateSplitFromAppointmentAllowsOnePetWithoutServicesWhenAnotherHasServices(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138122",
		CustomerNickname:   "多猫部分服务客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 0,
	})
	secondPet := model.Pet{
		ShopID:     1,
		CustomerID: &state.customer.ID,
		Name:       "实际洗护猫",
		Species:    "猫",
	}
	if err := database.DB.Create(&secondPet).Error; err != nil {
		t.Fatalf("create second pet: %v", err)
	}
	secondApptPet := model.AppointmentPet{
		AppointmentID: state.appointment.ID,
		PetID:         secondPet.ID,
		SortOrder:     2,
		TotalAmount:   168,
		TotalDuration: 75,
	}
	if err := database.DB.Create(&secondApptPet).Error; err != nil {
		t.Fatalf("create second appointment pet: %v", err)
	}
	secondService := model.AppointmentPetService{
		AppointmentPetID: secondApptPet.ID,
		ServiceID:        9002,
		ServiceName:      "实际完成服务",
		Price:            168,
		Duration:         75,
	}
	if err := database.DB.Create(&secondService).Error; err != nil {
		t.Fatalf("create second appointment pet service: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	orders, err := svc.CreateSplitFromAppointment(state.appointment.ID, false, nil, map[uint]PetOverrideData{
		state.pet.ID: {
			Services: []ServiceOverride{},
		},
		secondPet.ID: {
			Services: []ServiceOverride{{
				ServiceID:   secondService.ServiceID,
				ServiceName: secondService.ServiceName,
				Price:       secondService.Price,
				Duration:    secondService.Duration,
			}},
		},
	})
	if err != nil {
		t.Fatalf("create split order from appointment: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected one order, got %d", len(orders))
	}

	var items []model.OrderItem
	if err := database.DB.Where("order_id = ?", orders[0].ID).Order("id ASC").Find(&items).Error; err != nil {
		t.Fatalf("load order items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one item for the pet that had service, got %d: %+v", len(items), items)
	}
	if items[0].Name != "实际洗护猫 · 实际完成服务" {
		t.Fatalf("unexpected item name: %q", items[0].Name)
	}
}

func TestUpdateDraftSyncsServiceItemsToAppointment(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138119",
		CustomerNickname:   "改单同步预约项目客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 0,
	})
	changedService := model.Service{
		ShopID:    1,
		Name:      "改单后项目",
		BasePrice: 188,
		Duration:  80,
		Status:    1,
	}
	if err := database.DB.Create(&changedService).Error; err != nil {
		t.Fatalf("create update changed service: %v", err)
	}
	order := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-APPOINTMENT-SERVICE-SYNC",
		ShopID:        1,
		CustomerID:    state.customer.ID,
		AppointmentID: &state.appointment.ID,
		ServiceTotal:  120,
		PayAmount:     120,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})
	if err := database.DB.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"pet_id":     state.pet.ID,
		"pay_status": 0,
		"status":     0,
	}).Error; err != nil {
		t.Fatalf("prepare appointment order: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	patch := &model.Order{
		CustomerID:            &state.customer.ID,
		PetID:                 &state.pet.ID,
		AppointmentID:         &state.appointment.ID,
		AppointmentIsLate:     false,
		TotalAmount:           188,
		ServiceTotal:          188,
		DiscountRate:          1,
		DiscountAmount:        0,
		ServiceDiscountAmount: 0,
		ProductDiscountAmount: 0,
		PayAmount:             188,
	}
	items := []model.OrderItem{
		{ItemType: 1, ItemID: changedService.ID, Name: changedService.Name, Quantity: 1, UnitPrice: 188, Amount: 188},
	}

	if err := svc.UpdateDraft(1, model.StaffRoleManager, order.ID, patch, items); err != nil {
		t.Fatalf("update appointment order draft: %v", err)
	}

	assertAppointmentPetServices(t, state.appointment.ID, state.pet.ID, []wantAppointmentService{{
		ServiceID:   changedService.ID,
		ServiceName: changedService.Name,
		Price:       188,
		Duration:    80,
	}})
}

func TestUpdateDraftReappliesReducedAppointmentDepositDeductionWhenLate(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138116",
		CustomerNickname:   "迟到改单客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 20,
	})
	order := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-APPOINTMENT-DEPOSIT-DRAFT-LATE",
		ShopID:        1,
		CustomerID:    state.customer.ID,
		AppointmentID: &state.appointment.ID,
		ServiceTotal:  120,
		PayAmount:     120,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})
	if err := database.DB.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"pay_status": 0,
		"status":     0,
	}).Error; err != nil {
		t.Fatalf("mark order unpaid: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	patch := &model.Order{
		CustomerID:            &state.customer.ID,
		AppointmentID:         &state.appointment.ID,
		AppointmentIsLate:     true,
		TotalAmount:           120,
		ServiceTotal:          120,
		DiscountRate:          1,
		DiscountAmount:        0,
		ServiceDiscountAmount: 0,
		ProductDiscountAmount: 0,
		PayAmount:             120,
	}
	items := []model.OrderItem{
		{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
	}

	if err := svc.UpdateDraft(1, model.StaffRoleManager, order.ID, patch, items); err != nil {
		t.Fatalf("update appointment order draft late: %v", err)
	}

	updated, err := svc.GetByID(order.ID)
	if err != nil {
		t.Fatalf("reload late updated order: %v", err)
	}
	if got := roundOrderAmount(updated.PayAmount); got != 106 {
		t.Fatalf("expected updated pay amount 106, got %.2f", got)
	}
	if got := roundOrderAmount(updated.AppointmentDepositDeductionAmount); got != 14 {
		t.Fatalf("expected updated late appointment deposit deduction 14, got %.2f", got)
	}
	if !updated.AppointmentIsLate {
		t.Fatalf("expected updated appointment late true")
	}
}

func TestUpdateDraftRejectsAppointmentLateChangeForPaidOrder(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138117",
		CustomerNickname:   "已支付迟到客户",
		AppointmentAmount:  120,
		AppointmentDeposit: 20,
	})
	order := seedOrderFilterOrder(t, seedOrderFilterOrderInput{
		OrderNo:       "TEST-APPOINTMENT-DEPOSIT-PAID-LATE",
		ShopID:        1,
		CustomerID:    state.customer.ID,
		AppointmentID: &state.appointment.ID,
		ServiceTotal:  120,
		PayAmount:     100,
		Items: []model.OrderItem{
			{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
		},
	})

	if err := database.DB.Model(&model.Order{}).Where("id = ?", order.ID).Updates(map[string]any{
		"pay_status":          1,
		"status":              1,
		"appointment_is_late": false,
	}).Error; err != nil {
		t.Fatalf("mark order paid: %v", err)
	}

	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	patch := &model.Order{
		CustomerID:            &state.customer.ID,
		AppointmentID:         &state.appointment.ID,
		AppointmentIsLate:     true,
		TotalAmount:           120,
		ServiceTotal:          120,
		DiscountRate:          1,
		DiscountAmount:        0,
		ServiceDiscountAmount: 0,
		ProductDiscountAmount: 0,
		PayAmount:             120,
	}
	items := []model.OrderItem{
		{ItemType: 1, ItemID: 9001, Name: "预约服务", Quantity: 1, UnitPrice: 120, Amount: 120},
	}

	err := svc.UpdateDraft(1, model.StaffRoleManager, order.ID, patch, items)
	if err == nil {
		t.Fatalf("expected late status change on paid order to fail")
	}
	if err.Error() != "已支付订单不可修改迟到状态" {
		t.Fatalf("unexpected error: %v", err)
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
		&model.AppointmentService{},
		&model.AppointmentPetService{},
		&model.FeedingPlan{},
		&model.FeedingPlanPet{},
		&model.FeedingVisit{},
		&model.BoardingCabinet{},
		&model.BoardingOrder{},
		&model.BoardingOrderRoom{},
		&model.BoardingOrderPet{},
		&model.ServiceCategory{},
		&model.Service{},
		&model.ProductCategory{},
		&model.Product{},
		&model.ProductSKU{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

type seedOrderFilterOrderInput struct {
	OrderNo       string
	ShopID        uint
	CustomerID    uint
	AppointmentID *uint
	FeedingPlanID *uint
	StaffID       *uint
	ServiceTotal  float64
	ProductTotal  float64
	PayAmount     float64
	PayMethod     string
	Commission    float64
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
		AppointmentID: input.AppointmentID,
		FeedingPlanID: input.FeedingPlanID,
		StaffID:       input.StaffID,
		TotalAmount:   input.PayAmount,
		ServiceTotal:  input.ServiceTotal,
		ProductTotal:  input.ProductTotal,
		PayAmount:     input.PayAmount,
		PayMethod:     input.PayMethod,
		Commission:    input.Commission,
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

type seedAppointmentOrderStateInput struct {
	ShopID             uint
	CustomerPhone      string
	CustomerNickname   string
	AppointmentAmount  float64
	AppointmentDeposit float64
	WithMemberCard     bool
}

type seedAppointmentOrderState struct {
	customer    model.Customer
	pet         model.Pet
	appointment model.Appointment
}

type wantAppointmentService struct {
	ServiceID   uint
	ServiceName string
	Price       float64
	Duration    int
}

func assertAppointmentPetServices(t *testing.T, appointmentID, petID uint, want []wantAppointmentService) {
	t.Helper()

	var apptPet model.AppointmentPet
	if err := database.DB.Where("appointment_id = ? AND pet_id = ?", appointmentID, petID).First(&apptPet).Error; err != nil {
		t.Fatalf("load appointment pet: %v", err)
	}
	var services []model.AppointmentPetService
	if err := database.DB.Where("appointment_pet_id = ?", apptPet.ID).Order("id ASC").Find(&services).Error; err != nil {
		t.Fatalf("load appointment pet services: %v", err)
	}
	if len(services) != len(want) {
		t.Fatalf("expected %d appointment services, got %d: %+v", len(want), len(services), services)
	}
	for i := range want {
		if services[i].ServiceID != want[i].ServiceID {
			t.Fatalf("service %d: expected id %d, got %d", i, want[i].ServiceID, services[i].ServiceID)
		}
		if services[i].ServiceName != want[i].ServiceName {
			t.Fatalf("service %d: expected name %q, got %q", i, want[i].ServiceName, services[i].ServiceName)
		}
		if got := roundOrderAmount(services[i].Price); got != roundOrderAmount(want[i].Price) {
			t.Fatalf("service %d: expected price %.2f, got %.2f", i, want[i].Price, got)
		}
		if services[i].Duration != want[i].Duration {
			t.Fatalf("service %d: expected duration %d, got %d", i, want[i].Duration, services[i].Duration)
		}
	}

	var activeCount int64
	petSubQuery := database.DB.Model(&model.AppointmentPet{}).Select("id").Where("appointment_id = ?", appointmentID)
	if err := database.DB.Model(&model.AppointmentPetService{}).Where("appointment_pet_id IN (?)", petSubQuery).Count(&activeCount).Error; err != nil {
		t.Fatalf("count appointment pet services: %v", err)
	}
	if activeCount != int64(len(want)) {
		t.Fatalf("expected %d active appointment pet service rows, got %d", len(want), activeCount)
	}
}

func TestSyncAppointmentServicesPreservesOrderItemPriceRuleName(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:            1,
		CustomerPhone:     "13800138088",
		CustomerNickname:  "规格客户",
		AppointmentAmount: 108,
	})
	baseService := model.Service{
		ShopID:    1,
		Name:      "基础洗护",
		BasePrice: 108,
		Duration:  60,
		Status:    1,
	}
	if err := database.DB.Create(&baseService).Error; err != nil {
		t.Fatalf("create service: %v", err)
	}

	items := []model.OrderItem{
		{ItemType: 1, ItemID: baseService.ID, Name: state.pet.Name + " · 基础洗护(长毛猫)", Quantity: 1, UnitPrice: 108, Amount: 108},
		{ItemType: 1, ItemID: baseService.ID, Name: state.pet.Name + " · 基础洗护(超重)", Quantity: 1, UnitPrice: 10, Amount: 10},
	}
	if err := syncAppointmentServicesFromOrderItems(database.DB, state.appointment.ID, nil, items); err != nil {
		t.Fatalf("sync appointment services: %v", err)
	}

	assertAppointmentPetServices(t, state.appointment.ID, state.pet.ID, []wantAppointmentService{
		{ServiceID: baseService.ID, ServiceName: "基础洗护(长毛猫)", Price: 108, Duration: 60},
		{ServiceID: baseService.ID, ServiceName: "基础洗护(超重)", Price: 10, Duration: 60},
	})
}

func seedAppointmentOrderFixture(t *testing.T, input seedAppointmentOrderStateInput) seedAppointmentOrderState {
	t.Helper()

	customer := model.Customer{
		ShopID:       input.ShopID,
		Phone:        input.CustomerPhone,
		Nickname:     input.CustomerNickname,
		DiscountRate: 1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create appointment customer: %v", err)
	}

	if input.WithMemberCard {
		template := model.MemberCardTemplate{
			ShopID:              input.ShopID,
			Name:                "预约金卡",
			MinRecharge:         0,
			DiscountRate:        1,
			ProductDiscountRate: 1,
			Status:              1,
		}
		if err := database.DB.Create(&template).Error; err != nil {
			t.Fatalf("create member card template: %v", err)
		}
		card := model.MemberCard{
			ShopID:              input.ShopID,
			CustomerID:          customer.ID,
			TemplateID:          template.ID,
			CardName:            "预约金会员卡",
			DiscountRate:        1,
			ProductDiscountRate: 1,
			Status:              1,
		}
		if err := database.DB.Create(&card).Error; err != nil {
			t.Fatalf("create member card: %v", err)
		}
	}

	pet := model.Pet{
		ShopID:     input.ShopID,
		CustomerID: &customer.ID,
		Name:       "测试猫咪",
		Species:    "猫",
	}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create appointment pet: %v", err)
	}

	appointment := model.Appointment{
		ShopID:      input.ShopID,
		CustomerID:  customer.ID,
		PetID:       pet.ID,
		Date:        "2026-04-19",
		StartTime:   "10:00",
		EndTime:     "11:00",
		Status:      3,
		Source:      2,
		TotalAmount: input.AppointmentAmount,
		Deposit:     input.AppointmentDeposit,
	}
	if err := database.DB.Create(&appointment).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	apptPet := model.AppointmentPet{
		AppointmentID: appointment.ID,
		PetID:         pet.ID,
		SortOrder:     1,
		TotalAmount:   input.AppointmentAmount,
		TotalDuration: 60,
	}
	if err := database.DB.Create(&apptPet).Error; err != nil {
		t.Fatalf("create appointment pet row: %v", err)
	}

	apptPetService := model.AppointmentPetService{
		AppointmentPetID: apptPet.ID,
		ServiceID:        9001,
		ServiceName:      "预约服务",
		Price:            input.AppointmentAmount,
		Duration:         60,
	}
	if err := database.DB.Create(&apptPetService).Error; err != nil {
		t.Fatalf("create appointment pet service: %v", err)
	}

	return seedAppointmentOrderState{
		customer:    customer,
		pet:         pet,
		appointment: appointment,
	}
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
