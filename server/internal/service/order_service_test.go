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

	if got := roundOrderAmount(stats.TodayRevenue); got != 460 {
		t.Fatalf("expected total revenue 460, got %.2f", got)
	}

	breakdown := make(map[string]float64)
	for _, item := range stats.PaymentBreakdown {
		breakdown[item.Key] = item.Amount
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
	if got := roundOrderAmount(breakdown["other"]); got != 50 {
		t.Fatalf("expected other 50, got %.2f", got)
	}
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
		&model.BoardingCabinet{},
		&model.BoardingOrder{},
		&model.BoardingOrderRoom{},
		&model.BoardingOrderPet{},
		&model.ServiceCategory{},
		&model.Service{},
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
	ServiceTotal  float64
	PayAmount     float64
	PayMethod     string
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
		TotalAmount:   input.PayAmount,
		ServiceTotal:  input.ServiceTotal,
		PayAmount:     input.PayAmount,
		PayMethod:     input.PayMethod,
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
