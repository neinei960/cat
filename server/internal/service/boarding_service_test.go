package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCancelRoomClearsMissingLinkedPayOrder(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedPendingBoardingOrderWithMissingPayOrder(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CancelRoom(state.shopID, state.order.ID, state.room.ID, state.operatorID)
	if err != nil {
		t.Fatalf("cancel room: %v", err)
	}
	if order.OrderID != nil {
		t.Fatalf("expected stale order link to be cleared, got %v", *order.OrderID)
	}
	if order.Status != model.BoardingOrderStatusCancelled {
		t.Fatalf("expected boarding order status %q, got %q", model.BoardingOrderStatusCancelled, order.Status)
	}
	if len(order.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(order.Rooms))
	}
	if order.Rooms[0].Status != model.BoardingOrderStatusCancelled {
		t.Fatalf("expected room status %q, got %q", model.BoardingOrderStatusCancelled, order.Rooms[0].Status)
	}

	var persisted model.BoardingOrder
	if err := database.DB.First(&persisted, state.order.ID).Error; err != nil {
		t.Fatalf("load persisted boarding order: %v", err)
	}
	if persisted.OrderID != nil {
		t.Fatalf("expected persisted stale order link to be cleared, got %v", *persisted.OrderID)
	}
}

func TestDashboardSortsCabinetsByBasePrice(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "测试门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	cabinets := []model.BoardingCabinet{
		{
			ShopID:      shop.ID,
			Code:        "high-price",
			CabinetType: "高价房",
			RoomCount:   2,
			Capacity:    1,
			BasePrice:   165,
			Status:      model.BoardingCabinetStatusEnabled,
		},
		{
			ShopID:      shop.ID,
			Code:        "low-price",
			CabinetType: "低价房",
			RoomCount:   2,
			Capacity:    1,
			BasePrice:   85,
			Status:      model.BoardingCabinetStatusEnabled,
		},
		{
			ShopID:      shop.ID,
			Code:        "mid-price",
			CabinetType: "中价房",
			RoomCount:   2,
			Capacity:    1,
			BasePrice:   120,
			Status:      model.BoardingCabinetStatusEnabled,
		},
	}
	if err := database.DB.Create(&cabinets).Error; err != nil {
		t.Fatalf("create cabinets: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	groups, err := svc.Dashboard(shop.ID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].CabinetType != "低价房" || groups[1].CabinetType != "中价房" || groups[2].CabinetType != "高价房" {
		t.Fatalf("expected groups sorted by base price asc, got %q, %q, %q", groups[0].CabinetType, groups[1].CabinetType, groups[2].CabinetType)
	}
}

func TestPreviewSupportsMultipleSpecialItems(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "多加收门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}
	customer := model.Customer{ShopID: shop.ID, Phone: "13800138888", Nickname: "寄养客户", DiscountRate: 1}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: shop.ID, CustomerID: &customer.ID, Name: "薯条", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	cabinet := model.BoardingCabinet{ShopID: shop.ID, Code: "multi-special", CabinetType: "普通房", RoomCount: 2, Capacity: 1, BasePrice: 100, Status: model.BoardingCabinetStatusEnabled}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		t.Fatalf("create cabinet: %v", err)
	}
	items := []model.BoardingSpecialItem{
		{ShopID: shop.ID, Name: "特殊护理", DefaultDailyPrice: 10, Status: 1},
		{ShopID: shop.ID, Name: "节假日加收", DefaultDailyPrice: 20, Status: 1},
	}
	if err := database.DB.Create(&items).Error; err != nil {
		t.Fatalf("create special items: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)
	preview, _, _, err := svc.Preview(shop.ID, BoardingPreviewInput{
		CustomerID: customer.ID,
		PetIDs:     []uint{pet.ID},
		CabinetID:  cabinet.ID,
		CheckInAt:  "2026-05-01",
		CheckOutAt: "2026-05-04",
		SpecialItems: []BoardingSpecialItemSelection{
			{ID: items[0].ID, DailyPrice: 10, Days: 3},
			{ID: items[1].ID, DailyPrice: 20, Days: 2},
		},
	})
	if err != nil {
		t.Fatalf("preview: %v", err)
	}
	if preview.SpecialItemAmount != 70 {
		t.Fatalf("expected special item amount 70.00, got %.2f", preview.SpecialItemAmount)
	}
	if preview.PayAmount != 370 {
		t.Fatalf("expected pay amount 370.00, got %.2f", preview.PayAmount)
	}
	var specialLines []BoardingPriceLine
	for _, line := range preview.Lines {
		if line.Type == "special_item" {
			specialLines = append(specialLines, line)
		}
	}
	if len(specialLines) != 2 {
		t.Fatalf("expected 2 special item lines, got %d", len(specialLines))
	}
	if specialLines[0].SpecialItemID == 0 || specialLines[1].SpecialItemID == 0 {
		t.Fatalf("expected special item line ids to be preserved: %+v", specialLines)
	}
}

func TestBoardingHistoryUsesCheckoutDateInsteadOfPayment(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "历史规则门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}
	customer := model.Customer{ShopID: shop.ID, Phone: "13800138999", Nickname: "寄养客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	staff := model.Staff{ShopID: shop.ID, Phone: "13800138001", Name: "员工", Role: model.StaffRoleStaff, Status: 1}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	cabinet := model.BoardingCabinet{ShopID: shop.ID, Code: "history-room", CabinetType: "历史房", RoomCount: 2, Capacity: 1, BasePrice: 100, Status: model.BoardingCabinetStatusEnabled}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		t.Fatalf("create cabinet: %v", err)
	}

	today := time.Now()
	yesterday := today.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := today.AddDate(0, 0, 1).Format("2006-01-02")
	nextDay := today.AddDate(0, 0, 2).Format("2006-01-02")

	paidPastOrder := model.Order{OrderNo: "BHIST-PAST", ShopID: shop.ID, CustomerID: &customer.ID, StaffID: &staff.ID, PayAmount: 100, PayStatus: 1, Status: 1}
	if err := database.DB.Create(&paidPastOrder).Error; err != nil {
		t.Fatalf("create paid past order: %v", err)
	}
	pastBoarding := model.BoardingOrder{
		ShopID: shop.ID, OrderID: &paidPastOrder.ID, CustomerID: customer.ID, StaffID: staff.ID,
		CabinetID: cabinet.ID, CheckInAt: today.AddDate(0, 0, -3).Format("2006-01-02"), CheckOutAt: yesterday,
		Status: model.BoardingOrderStatusCheckedIn, PayAmount: 100,
	}
	if err := database.DB.Create(&pastBoarding).Error; err != nil {
		t.Fatalf("create past boarding: %v", err)
	}
	pastRoom := model.BoardingOrderRoom{
		BoardingOrderID: pastBoarding.ID, CabinetID: cabinet.ID, RoomIndex: 1,
		CheckInAt: pastBoarding.CheckInAt, CheckOutAt: pastBoarding.CheckOutAt,
		Status: model.BoardingOrderStatusCheckedIn, PayAmount: 100,
	}
	if err := database.DB.Create(&pastRoom).Error; err != nil {
		t.Fatalf("create past room: %v", err)
	}

	paidFutureOrder := model.Order{OrderNo: "BHIST-FUTURE", ShopID: shop.ID, CustomerID: &customer.ID, StaffID: &staff.ID, PayAmount: 100, PayStatus: 1, Status: 1}
	if err := database.DB.Create(&paidFutureOrder).Error; err != nil {
		t.Fatalf("create paid future order: %v", err)
	}
	futureBoarding := model.BoardingOrder{
		ShopID: shop.ID, OrderID: &paidFutureOrder.ID, CustomerID: customer.ID, StaffID: staff.ID,
		CabinetID: cabinet.ID, CheckInAt: tomorrow, CheckOutAt: nextDay,
		Status: model.BoardingOrderStatusCheckedIn, PayAmount: 100,
	}
	if err := database.DB.Create(&futureBoarding).Error; err != nil {
		t.Fatalf("create future boarding: %v", err)
	}
	futureRoom := model.BoardingOrderRoom{
		BoardingOrderID: futureBoarding.ID, CabinetID: cabinet.ID, RoomIndex: 1,
		CheckInAt: futureBoarding.CheckInAt, CheckOutAt: futureBoarding.CheckOutAt,
		Status: model.BoardingOrderStatusCheckedIn, PayAmount: 100,
	}
	if err := database.DB.Create(&futureRoom).Error; err != nil {
		t.Fatalf("create future room: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)
	groups, err := svc.Dashboard(shop.ID)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected one cabinet group, got %d", len(groups))
	}
	if got := len(groups[0].Orders); got != 1 {
		t.Fatalf("expected only the not-yet-ended boarding order on dashboard, got %d", got)
	}
	if groups[0].Orders[0].ID != futureBoarding.ID {
		t.Fatalf("expected future paid boarding to stay active, got order #%d", groups[0].Orders[0].ID)
	}

	history, _, err := repository.NewBoardingRepository().ListBoardingOrders(shop.ID, model.BoardingOrderStatusCheckedOut, "", "", 0, 1, 20)
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected only ended boarding in history, got %d", len(history))
	}
	if history[0].ID != pastBoarding.ID {
		t.Fatalf("expected past boarding in history, got order #%d", history[0].ID)
	}
}

func TestRoundMoneyHandlesNegativeValues(t *testing.T) {
	if got := roundMoney(-200); got != -200 {
		t.Fatalf("expected negative roundMoney to preserve -200.00, got %.2f", got)
	}
	if got := roundMoney(-19.995); got != -20 {
		t.Fatalf("expected negative roundMoney to round to -20.00, got %.2f", got)
	}
}

func TestCreateHolidayEnsuresDefaultHolidaySurchargePolicy(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "节假日门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)
	err := svc.CreateHoliday(&model.BoardingHoliday{
		ShopID:      shop.ID,
		HolidayDate: "2026-05-01",
		Name:        "五一",
	})
	if err != nil {
		t.Fatalf("create holiday: %v", err)
	}

	var policy model.BoardingDiscountPolicy
	if err := database.DB.Where("shop_id = ? AND policy_type = ?", shop.ID, model.BoardingPolicyTypeHolidaySurcharge).First(&policy).Error; err != nil {
		t.Fatalf("load default holiday surcharge policy: %v", err)
	}
	if policy.Status != 1 {
		t.Fatalf("expected default holiday surcharge policy enabled, got %d", policy.Status)
	}
	var rule surchargeRule
	if err := json.Unmarshal([]byte(policy.RuleJSON), &rule); err != nil {
		t.Fatalf("unmarshal default surcharge rule: %v", err)
	}
	if got := roundMoney(rule.Surcharge); got != defaultHolidaySurchargeAmount {
		t.Fatalf("expected default surcharge %.2f, got %.2f", defaultHolidaySurchargeAmount, got)
	}
}

func TestCreateHolidayRangeCreatesInclusiveDailyRecords(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "范围节假日门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	created, err := svc.CreateHolidayRange(shop.ID, "2026-05-01", "2026-05-03", "五一")
	if err != nil {
		t.Fatalf("create holiday range: %v", err)
	}
	if len(created) != 3 {
		t.Fatalf("expected 3 holidays created, got %d", len(created))
	}

	wantDates := []string{"2026-05-01", "2026-05-02", "2026-05-03"}
	for i, want := range wantDates {
		if created[i].HolidayDate != want {
			t.Fatalf("expected created holiday %d date %s, got %s", i, want, created[i].HolidayDate)
		}
		if created[i].Name != "五一" {
			t.Fatalf("expected created holiday %d name 五一, got %s", i, created[i].Name)
		}
	}

	list, err := svc.ListHolidays(shop.ID)
	if err != nil {
		t.Fatalf("list holidays: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 persisted holidays, got %d", len(list))
	}
	for i, want := range wantDates {
		if list[i].HolidayDate != want {
			t.Fatalf("expected persisted holiday %d date %s, got %s", i, want, list[i].HolidayDate)
		}
	}
}

func TestCreateHolidayRangeSkipsExistingDates(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "去重节假日门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	if err := svc.CreateHoliday(&model.BoardingHoliday{
		ShopID:      shop.ID,
		HolidayDate: "2026-05-02",
		Name:        "已有日期",
	}); err != nil {
		t.Fatalf("seed existing holiday: %v", err)
	}

	created, err := svc.CreateHolidayRange(shop.ID, "2026-05-01", "2026-05-03", "五一")
	if err != nil {
		t.Fatalf("create holiday range: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 newly created holidays, got %d", len(created))
	}
	if created[0].HolidayDate != "2026-05-01" || created[1].HolidayDate != "2026-05-03" {
		t.Fatalf("expected created dates 2026-05-01 and 2026-05-03, got %s and %s", created[0].HolidayDate, created[1].HolidayDate)
	}

	list, err := svc.ListHolidays(shop.ID)
	if err != nil {
		t.Fatalf("list holidays: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 persisted holidays after skipping duplicate, got %d", len(list))
	}
}

func TestPreviewBoardingOrderAppliesFixedDepositDeduction(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	preview, err := svc.PreviewOrder(state.shopID, BoardingPreviewInput{
		CustomerID:     state.customer.ID,
		DepositEnabled: true,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
	})
	if err != nil {
		t.Fatalf("preview boarding order: %v", err)
	}

	if got := roundMoney(preview.PayAmount); got != 100 {
		t.Fatalf("expected preview pay amount 100 after deposit deduction, got %.2f", got)
	}
	assertBoardingDepositLine(t, preview.Lines, -200)
}

func TestPreviewBoardingOrderSkipsDepositWhenNotSelected(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	preview, err := svc.PreviewOrder(state.shopID, BoardingPreviewInput{
		CustomerID: state.customer.ID,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
	})
	if err != nil {
		t.Fatalf("preview boarding order: %v", err)
	}

	if got := roundMoney(preview.PayAmount); got != 300 {
		t.Fatalf("expected preview pay amount 300 without deposit, got %.2f", got)
	}
	assertNoBoardingDepositLine(t, preview.Lines)
}

func TestPreviewBoardingOrderAppliesDepositOnlyWhenSelected(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	preview, err := svc.PreviewOrder(state.shopID, BoardingPreviewInput{
		CustomerID:     state.customer.ID,
		DepositEnabled: true,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
	})
	if err != nil {
		t.Fatalf("preview boarding order: %v", err)
	}

	if got := roundMoney(preview.PayAmount); got != 100 {
		t.Fatalf("expected preview pay amount 100 after optional deposit deduction, got %.2f", got)
	}
	assertBoardingDepositLine(t, preview.Lines, -200)
}

func TestCreateBoardingOrderAppliesFixedDepositDeduction(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID:     state.customer.ID,
		DepositEnabled: true,
		CheckInAt:      "2026-04-20",
		CheckOutAt:     "2026-04-23",
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order: %v", err)
	}
	if got := roundMoney(order.PayAmount); got != 100 {
		t.Fatalf("expected boarding pay amount 100 after deposit deduction, got %.2f", got)
	}
	if order.OrderID == nil || *order.OrderID == 0 {
		t.Fatalf("expected linked pay order to be created")
	}

	var payOrder model.Order
	if err := database.DB.First(&payOrder, *order.OrderID).Error; err != nil {
		t.Fatalf("load pay order: %v", err)
	}
	if got := roundMoney(payOrder.AppointmentDepositAmount); got != 200 {
		t.Fatalf("expected stored deposit amount 200, got %.2f", got)
	}
	if got := roundMoney(payOrder.AppointmentDepositDeductionAmount); got != 200 {
		t.Fatalf("expected deposit deduction 200, got %.2f", got)
	}
	if got := roundMoney(payOrder.PayAmount); got != 100 {
		t.Fatalf("expected linked order pay amount 100 after deduction, got %.2f", got)
	}

	var snapshot BoardingPricePreview
	if err := json.Unmarshal([]byte(order.PriceSnapshotJSON), &snapshot); err != nil {
		t.Fatalf("unmarshal boarding price snapshot: %v", err)
	}
	assertBoardingDepositLine(t, snapshot.Lines, -200)
}

func TestCreateBoardingOrderSkipsDepositWhenNotSelected(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID: state.customer.ID,
		CheckInAt:  "2026-04-20",
		CheckOutAt: "2026-04-23",
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order: %v", err)
	}
	if got := roundMoney(order.PayAmount); got != 300 {
		t.Fatalf("expected boarding pay amount 300 without deposit, got %.2f", got)
	}
	if order.OrderID == nil || *order.OrderID == 0 {
		t.Fatalf("expected linked pay order to be created")
	}

	var payOrder model.Order
	if err := database.DB.First(&payOrder, *order.OrderID).Error; err != nil {
		t.Fatalf("load pay order: %v", err)
	}
	if got := roundMoney(payOrder.AppointmentDepositAmount); got != 0 {
		t.Fatalf("expected stored deposit amount 0, got %.2f", got)
	}
	if got := roundMoney(payOrder.AppointmentDepositDeductionAmount); got != 0 {
		t.Fatalf("expected deposit deduction 0, got %.2f", got)
	}

	var snapshot BoardingPricePreview
	if err := json.Unmarshal([]byte(order.PriceSnapshotJSON), &snapshot); err != nil {
		t.Fatalf("unmarshal boarding price snapshot: %v", err)
	}
	assertNoBoardingDepositLine(t, snapshot.Lines)
}

func TestExtendBoardingOrderReappliesFixedDepositDeduction(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID:     state.customer.ID,
		DepositEnabled: true,
		CheckInAt:      "2026-04-20",
		CheckOutAt:     "2026-04-23",
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-23",
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order: %v", err)
	}

	extended, err := svc.Extend(state.shopID, order.ID, state.operator.ID, "2026-04-24")
	if err != nil {
		t.Fatalf("extend boarding order: %v", err)
	}
	if got := roundMoney(extended.PayAmount); got != 200 {
		t.Fatalf("expected extended boarding pay amount 200 after deduction, got %.2f", got)
	}
	if extended.OrderID == nil || *extended.OrderID == 0 {
		t.Fatalf("expected linked pay order after extension")
	}

	var payOrder model.Order
	if err := database.DB.First(&payOrder, *extended.OrderID).Error; err != nil {
		t.Fatalf("load extended pay order: %v", err)
	}
	if got := roundMoney(payOrder.AppointmentDepositAmount); got != 200 {
		t.Fatalf("expected stored deposit amount 200 after extension, got %.2f", got)
	}
	if got := roundMoney(payOrder.AppointmentDepositDeductionAmount); got != 200 {
		t.Fatalf("expected deposit deduction 200 after extension, got %.2f", got)
	}
	if got := roundMoney(payOrder.PayAmount); got != 200 {
		t.Fatalf("expected linked order pay amount 200 after extension, got %.2f", got)
	}
}

func TestAdjustRoomPriceAllowsUpdatingDiscountAfterCheckIn(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID:     state.customer.ID,
		DepositEnabled: true,
		CheckInAt:      "2026-04-20",
		CheckOutAt:     "2026-04-24",
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  state.cabinet.ID,
				CheckInAt:  "2026-04-20",
				CheckOutAt: "2026-04-24",
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order: %v", err)
	}
	if len(order.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(order.Rooms))
	}

	checkedIn, err := svc.CheckInRoom(state.shopID, order.ID, order.Rooms[0].ID, state.operator.ID, BoardingCheckInInput{
		DiscountAmount: 20,
	})
	if err != nil {
		t.Fatalf("check in room: %v", err)
	}
	if got := roundMoney(checkedIn.Rooms[0].ManualDiscountAmount); got != 20 {
		t.Fatalf("expected initial room discount 20, got %.2f", got)
	}

	adjusted, err := svc.AdjustRoomPrice(state.shopID, order.ID, checkedIn.Rooms[0].ID, state.operator.ID, BoardingCheckInInput{
		DiscountAmount: 35,
	})
	if err != nil {
		t.Fatalf("adjust room price: %v", err)
	}
	if got := roundMoney(adjusted.Rooms[0].ManualDiscountAmount); got != 35 {
		t.Fatalf("expected adjusted room discount 35, got %.2f", got)
	}
	if got := roundMoney(adjusted.Rooms[0].PayAmount); got != 400 {
		t.Fatalf("expected stored room pay amount 400 before manual discount display, got %.2f", got)
	}
	if adjusted.OrderID == nil || *adjusted.OrderID == 0 {
		t.Fatalf("expected linked pay order after adjustment")
	}

	var payOrder model.Order
	if err := database.DB.First(&payOrder, *adjusted.OrderID).Error; err != nil {
		t.Fatalf("load adjusted pay order: %v", err)
	}
	if got := roundMoney(payOrder.PayAmount); got != 165 {
		t.Fatalf("expected linked pay order pay amount 165, got %.2f", got)
	}
	if got := roundMoney(payOrder.AppointmentDepositDeductionAmount); got != 200 {
		t.Fatalf("expected deposit deduction to stay 200, got %.2f", got)
	}

	var latestLog model.BoardingOrderLog
	if err := database.DB.Where("boarding_order_id = ?", adjusted.ID).Order("id DESC").First(&latestLog).Error; err != nil {
		t.Fatalf("load latest boarding log: %v", err)
	}
	if latestLog.Action != "adjust_price" {
		t.Fatalf("expected latest log action adjust_price, got %q", latestLog.Action)
	}
}

func TestUpdateDewormingAllowsCheckedOutOrderAndClearingState(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedPendingBoardingOrderWithMissingPayOrder(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	initial := false
	if err := database.DB.Model(&model.BoardingOrder{}).
		Where("id = ?", state.order.ID).
		Updates(map[string]any{
			"status":              model.BoardingOrderStatusCheckedOut,
			"has_deworming":       &initial,
			"actual_check_out_at": "2026-04-17",
		}).Error; err != nil {
		t.Fatalf("update boarding order status: %v", err)
	}
	if err := database.DB.Model(&model.BoardingOrderRoom{}).
		Where("id = ?", state.room.ID).
		Updates(map[string]any{
			"status":              model.BoardingOrderStatusCheckedOut,
			"actual_check_out_at": "2026-04-17",
		}).Error; err != nil {
		t.Fatalf("update boarding room status: %v", err)
	}

	target := true
	updated, err := svc.UpdateDeworming(state.shopID, state.order.ID, state.operatorID, &target)
	if err != nil {
		t.Fatalf("update deworming on checked out order: %v", err)
	}
	if updated.HasDeworming == nil || !*updated.HasDeworming {
		t.Fatalf("expected checked out order deworming status to be true, got %#v", updated.HasDeworming)
	}
	if updated.Status != model.BoardingOrderStatusCheckedOut {
		t.Fatalf("expected checked out status to remain unchanged, got %q", updated.Status)
	}

	cleared, err := svc.UpdateDeworming(state.shopID, state.order.ID, state.operatorID, nil)
	if err != nil {
		t.Fatalf("clear deworming status: %v", err)
	}
	if cleared.HasDeworming != nil {
		t.Fatalf("expected deworming status to be cleared, got %#v", cleared.HasDeworming)
	}

	var latestLog model.BoardingOrderLog
	if err := database.DB.Where("boarding_order_id = ?", state.order.ID).Order("id DESC").First(&latestLog).Error; err != nil {
		t.Fatalf("load latest boarding log: %v", err)
	}
	if latestLog.Action != "update_deworming" {
		t.Fatalf("expected latest log action update_deworming, got %q", latestLog.Action)
	}
	if !strings.Contains(latestLog.Content, "已驱虫 -> 未填写") {
		t.Fatalf("expected deworming log to record status transition, got %q", latestLog.Content)
	}
}

func TestPreviewBoardingOrderDoesNotDiscountSpecialItem(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingSpecialItemFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	preview, err := svc.PreviewOrder(state.shopID, BoardingPreviewInput{
		CustomerID: state.customer.ID,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:              1,
				CabinetID:             state.cabinet.ID,
				CheckInAt:             "2026-04-20",
				CheckOutAt:            "2026-04-27",
				SpecialItemID:         state.specialItem.ID,
				SpecialItemDailyPrice: 30,
				SpecialItemDays:       2,
			},
		},
	})
	if err != nil {
		t.Fatalf("preview boarding order with special item: %v", err)
	}
	if got := roundMoney(preview.SpecialItemAmount); got != 60 {
		t.Fatalf("expected preview special item amount 60, got %.2f", got)
	}
	if got := roundMoney(preview.DiscountAmount); got != 70 {
		t.Fatalf("expected member discount amount 70 without discounting special item, got %.2f", got)
	}
	if got := roundMoney(preview.PayAmount); got != 690 {
		t.Fatalf("expected preview pay amount 690 after member discount without deposit, got %.2f", got)
	}
	if len(preview.Rooms) != 1 {
		t.Fatalf("expected 1 preview room, got %d", len(preview.Rooms))
	}
	if got := roundMoney(preview.Rooms[0].SpecialItemAmount); got != 60 {
		t.Fatalf("expected preview room special item amount 60, got %.2f", got)
	}

	found := false
	for _, line := range preview.Lines {
		if line.Type == "special_item" {
			found = true
			if got := roundMoney(line.Amount); got != 60 {
				t.Fatalf("expected special item line amount 60, got %.2f", got)
			}
		}
	}
	if !found {
		t.Fatalf("expected special item line in preview lines, got %+v", preview.Lines)
	}
}

func TestCreateBoardingOrderSyncsSpecialItemToAddonTotals(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingSpecialItemFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID: state.customer.ID,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:              1,
				CabinetID:             state.cabinet.ID,
				CheckInAt:             "2026-04-20",
				CheckOutAt:            "2026-04-27",
				SpecialItemID:         state.specialItem.ID,
				SpecialItemDailyPrice: 30,
				SpecialItemDays:       2,
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order with special item: %v", err)
	}
	if got := roundMoney(order.SpecialItemAmount); got != 60 {
		t.Fatalf("expected boarding order special item amount 60, got %.2f", got)
	}
	if len(order.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(order.Rooms))
	}
	if got := roundMoney(order.Rooms[0].SpecialItemAmount); got != 60 {
		t.Fatalf("expected room special item amount 60, got %.2f", got)
	}
	if order.OrderID == nil || *order.OrderID == 0 {
		t.Fatalf("expected linked pay order")
	}

	var payOrder model.Order
	if err := database.DB.First(&payOrder, *order.OrderID).Error; err != nil {
		t.Fatalf("load linked pay order: %v", err)
	}
	if got := roundMoney(payOrder.ServiceTotal); got != 700 {
		t.Fatalf("expected pay order service total 700, got %.2f", got)
	}
	if got := roundMoney(payOrder.AddonTotal); got != 60 {
		t.Fatalf("expected pay order addon total 60, got %.2f", got)
	}
	if got := roundMoney(payOrder.TotalAmount); got != 760 {
		t.Fatalf("expected pay order total amount 760, got %.2f", got)
	}
	if got := roundMoney(payOrder.DiscountAmount); got != 70 {
		t.Fatalf("expected pay order discount amount 70, got %.2f", got)
	}
	if got := roundMoney(payOrder.PayAmount); got != 690 {
		t.Fatalf("expected pay order pay amount 690, got %.2f", got)
	}

	var items []model.OrderItem
	if err := database.DB.Where("order_id = ?", payOrder.ID).Order("id ASC").Find(&items).Error; err != nil {
		t.Fatalf("load order items: %v", err)
	}
	addonFound := false
	for _, item := range items {
		if item.ItemType == 3 {
			addonFound = true
			if got := roundMoney(item.Amount); got != 60 {
				t.Fatalf("expected addon item amount 60, got %.2f", got)
			}
		}
	}
	if !addonFound {
		t.Fatalf("expected addon order item for special boarding fee, got %+v", items)
	}
}

func TestCheckOutRoomRejectsSpecialItemDaysExceedNights(t *testing.T) {
	setupBoardingServiceTestDB(t)

	state := seedBoardingSpecialItemFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(state.shopID, BoardingCreateInput{
		CustomerID: state.customer.ID,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:              1,
				CabinetID:             state.cabinet.ID,
				CheckInAt:             "2026-04-20",
				CheckOutAt:            "2026-04-27",
				SpecialItemID:         state.specialItem.ID,
				SpecialItemDailyPrice: 30,
				SpecialItemDays:       2,
			},
		},
		OperatorID: state.operator.ID,
	})
	if err != nil {
		t.Fatalf("create boarding order with special item: %v", err)
	}
	if len(order.Rooms) != 1 {
		t.Fatalf("expected 1 room, got %d", len(order.Rooms))
	}

	checkedIn, err := svc.CheckInRoom(state.shopID, order.ID, order.Rooms[0].ID, state.operator.ID, BoardingCheckInInput{})
	if err != nil {
		t.Fatalf("check in room: %v", err)
	}

	_, err = svc.CheckOutRoom(state.shopID, checkedIn.ID, checkedIn.Rooms[0].ID, state.operator.ID, "2026-04-21")
	if err == nil {
		t.Fatalf("expected early checkout to reject special item days greater than nights")
	}
	if err.Error() != "特殊寄养天数不能超过寄养晚数" {
		t.Fatalf("expected special item day validation error, got %v", err)
	}
}

func TestCheckInRoomPreservesProductItemsOnLinkedPayOrderSync(t *testing.T) {
	setupBoardingServiceTestDB(t)

	fixture := seedBoardingDepositFixture(t)
	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	order, err := svc.CreateOrder(fixture.shopID, BoardingCreateInput{
		CustomerID: fixture.customer.ID,
		RoomGroups: []BoardingRoomGroupInput{
			{
				PetCount:   1,
				CabinetID:  fixture.cabinet.ID,
				CheckInAt:  "2026-04-22",
				CheckOutAt: "2026-04-25",
			},
		},
		OperatorID:     fixture.operator.ID,
		DepositEnabled: true,
	})
	if err != nil {
		t.Fatalf("create boarding order: %v", err)
	}
	if order.OrderID == nil || *order.OrderID == 0 {
		t.Fatalf("expected linked pay order")
	}

	productItems := []model.OrderItem{
		{OrderID: *order.OrderID, ItemType: 2, ItemID: 801, Name: "柴柒·体内外同驱 · 单支", Quantity: 1, UnitPrice: 68, Amount: 68},
		{OrderID: *order.OrderID, ItemType: 2, ItemID: 802, Name: "大威宠·体内外同驱 · 幼猫", Quantity: 1, UnitPrice: 48, Amount: 48},
	}
	if err := database.DB.Create(&productItems).Error; err != nil {
		t.Fatalf("create linked product items: %v", err)
	}
	if err := database.DB.Model(&model.Order{}).
		Where("id = ?", *order.OrderID).
		Updates(map[string]any{
			"product_total": 116.0,
			"total_amount":  416.0,
			"pay_amount":    216.0,
		}).Error; err != nil {
		t.Fatalf("update linked pay order products: %v", err)
	}

	checkedIn, err := svc.CheckInRoom(fixture.shopID, order.ID, order.Rooms[0].ID, fixture.operator.ID, BoardingCheckInInput{})
	if err != nil {
		t.Fatalf("check in room: %v", err)
	}
	if checkedIn.OrderID == nil || *checkedIn.OrderID == 0 {
		t.Fatalf("expected checked in order to keep linked pay order")
	}

	var payOrder model.Order
	if err := database.DB.Preload("Items").First(&payOrder, *checkedIn.OrderID).Error; err != nil {
		t.Fatalf("load linked pay order: %v", err)
	}
	if got := roundMoney(payOrder.ServiceTotal); got != 300 {
		t.Fatalf("expected preserved boarding service total 300, got %.2f", got)
	}
	if got := roundMoney(payOrder.ProductTotal); got != 116 {
		t.Fatalf("expected preserved product total 116, got %.2f", got)
	}
	if got := roundMoney(payOrder.TotalAmount); got != 416 {
		t.Fatalf("expected combined total amount 416, got %.2f", got)
	}
	if got := roundMoney(payOrder.PayAmount); got != 216 {
		t.Fatalf("expected combined pay amount 216, got %.2f", got)
	}

	productCount := 0
	boardingCount := 0
	for _, item := range payOrder.Items {
		switch item.ItemType {
		case 2:
			productCount++
		case 4, 5, 6:
			boardingCount++
		}
	}
	if productCount != 2 {
		t.Fatalf("expected 2 preserved product items, got %d", productCount)
	}
	if boardingCount == 0 {
		t.Fatalf("expected boarding items to remain on linked pay order")
	}
}

type boardingServiceTestState struct {
	shopID     uint
	operatorID uint
	order      model.BoardingOrder
	room       model.BoardingOrderRoom
}

type boardingDepositFixture struct {
	shopID   uint
	customer model.Customer
	operator model.Staff
	cabinet  model.BoardingCabinet
}

type boardingSpecialItemFixture struct {
	shopID      uint
	customer    model.Customer
	operator    model.Staff
	cabinet     model.BoardingCabinet
	specialItem model.BoardingSpecialItem
}

func setupBoardingServiceTestDB(t *testing.T) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	database.DB = db
	if err := database.DB.AutoMigrate(
		&model.Shop{},
		&model.Customer{},
		&model.MemberCard{},
		&model.Staff{},
		&model.Pet{},
		&model.Order{},
		&model.OrderItem{},
		&model.BoardingCabinet{},
		&model.BoardingDiscountPolicy{},
		&model.BoardingHoliday{},
		&model.BoardingSpecialItem{},
		&model.BoardingOrder{},
		&model.BoardingOrderRoom{},
		&model.BoardingOrderPet{},
		&model.BoardingOrderLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func seedPendingBoardingOrderWithMissingPayOrder(t *testing.T) boardingServiceTestState {
	t.Helper()

	shop := model.Shop{Name: "测试门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	customer := model.Customer{
		ShopID:       shop.ID,
		Phone:        "13800138000",
		Nickname:     "寄养客户",
		DiscountRate: 1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	operator := model.Staff{
		ShopID:       shop.ID,
		Phone:        "13900139000",
		PasswordHash: "hash",
		Name:         "店员A",
		Role:         model.StaffRoleStaff,
		Status:       1,
	}
	if err := database.DB.Create(&operator).Error; err != nil {
		t.Fatalf("create operator: %v", err)
	}

	cabinet := model.BoardingCabinet{
		ShopID:      shop.ID,
		Code:        "warm-room",
		CabinetType: "康娜温柔乡",
		RoomCount:   3,
		Capacity:    1,
		BasePrice:   65,
		Status:      model.BoardingCabinetStatusEnabled,
	}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		t.Fatalf("create cabinet: %v", err)
	}

	preview := BoardingPricePreview{
		CheckInAt:  "2026-04-09",
		CheckOutAt: "2026-04-17",
		Nights:     8,
		PetCount:   1,
		BaseAmount: 520,
		PayAmount:  455,
		Lines: []BoardingPriceLine{
			{Type: "base", Label: "康娜温柔乡 寄养住宿", Quantity: 8, UnitPrice: 65, Amount: 520},
			{Type: "manual_discount", Label: "入住优惠", Quantity: 1, UnitPrice: -65, Amount: -65},
		},
	}
	priceSnapshot, err := json.Marshal(preview)
	if err != nil {
		t.Fatalf("marshal preview: %v", err)
	}

	missingOrderID := uint(999999)
	order := model.BoardingOrder{
		ShopID:                 shop.ID,
		OrderID:                &missingOrderID,
		CustomerID:             customer.ID,
		StaffID:                operator.ID,
		CabinetID:              cabinet.ID,
		CheckInAt:              preview.CheckInAt,
		CheckOutAt:             preview.CheckOutAt,
		Nights:                 preview.Nights,
		BaseAmount:             preview.BaseAmount,
		DiscountAmount:         65,
		ManualDiscountAmount:   65,
		PayAmount:              preview.PayAmount,
		Status:                 model.BoardingOrderStatusPendingCheckin,
		PriceSnapshotJSON:      string(priceSnapshot),
		PolicySnapshotJSON:     "[]",
		HolidaySurchargeAmount: 0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create boarding order: %v", err)
	}

	room := model.BoardingOrderRoom{
		BoardingOrderID:        order.ID,
		CabinetID:              cabinet.ID,
		RoomIndex:              1,
		CheckInAt:              preview.CheckInAt,
		CheckOutAt:             preview.CheckOutAt,
		Nights:                 preview.Nights,
		BaseAmount:             preview.BaseAmount,
		DiscountAmount:         65,
		ManualDiscountAmount:   65,
		PayAmount:              preview.PayAmount,
		Status:                 model.BoardingOrderStatusPendingCheckin,
		PriceSnapshotJSON:      string(priceSnapshot),
		PolicySnapshotJSON:     "[]",
		HolidaySurchargeAmount: 0,
	}
	if err := database.DB.Create(&room).Error; err != nil {
		t.Fatalf("create boarding room: %v", err)
	}

	return boardingServiceTestState{
		shopID:     shop.ID,
		operatorID: operator.ID,
		order:      order,
		room:       room,
	}
}

func seedBoardingDepositFixture(t *testing.T) boardingDepositFixture {
	t.Helper()

	shop := model.Shop{Name: "寄养定金门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	customer := model.Customer{
		ShopID:       shop.ID,
		Phone:        "13800138001",
		Nickname:     "寄养定金客户",
		DiscountRate: 1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	operator := model.Staff{
		ShopID:       shop.ID,
		Phone:        "13900139001",
		PasswordHash: "hash",
		Name:         "店员B",
		Role:         model.StaffRoleStaff,
		Status:       1,
	}
	if err := database.DB.Create(&operator).Error; err != nil {
		t.Fatalf("create operator: %v", err)
	}

	cabinet := model.BoardingCabinet{
		ShopID:      shop.ID,
		Code:        "deposit-room",
		CabinetType: "定金房型",
		RoomCount:   3,
		Capacity:    1,
		BasePrice:   100,
		Status:      model.BoardingCabinetStatusEnabled,
	}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		t.Fatalf("create cabinet: %v", err)
	}

	return boardingDepositFixture{
		shopID:   shop.ID,
		customer: customer,
		operator: operator,
		cabinet:  cabinet,
	}
}

func seedBoardingSpecialItemFixture(t *testing.T) boardingSpecialItemFixture {
	t.Helper()

	shop := model.Shop{Name: "寄养特殊项目门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	customer := model.Customer{
		ShopID:       shop.ID,
		Phone:        "13800138002",
		Nickname:     "特殊寄养客户",
		DiscountRate: 0.9,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	operator := model.Staff{
		ShopID:       shop.ID,
		Phone:        "13900139002",
		PasswordHash: "hash",
		Name:         "店员C",
		Role:         model.StaffRoleStaff,
		Status:       1,
	}
	if err := database.DB.Create(&operator).Error; err != nil {
		t.Fatalf("create operator: %v", err)
	}

	cabinet := model.BoardingCabinet{
		ShopID:      shop.ID,
		Code:        "special-room",
		CabinetType: "特殊护理房",
		RoomCount:   3,
		Capacity:    1,
		BasePrice:   100,
		Status:      model.BoardingCabinetStatusEnabled,
	}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		t.Fatalf("create cabinet: %v", err)
	}

	specialItem := model.BoardingSpecialItem{
		ShopID:            shop.ID,
		Name:              "用药护理",
		DefaultDailyPrice: 30,
		SortOrder:         1,
		Status:            1,
	}
	if err := database.DB.Create(&specialItem).Error; err != nil {
		t.Fatalf("create boarding special item: %v", err)
	}

	return boardingSpecialItemFixture{
		shopID:      shop.ID,
		customer:    customer,
		operator:    operator,
		cabinet:     cabinet,
		specialItem: specialItem,
	}
}

func assertBoardingDepositLine(t *testing.T, lines []BoardingPriceLine, wantAmount float64) {
	t.Helper()

	for _, line := range lines {
		if line.Type == "boarding_deposit" && line.Label == "定金抵扣" {
			if got := roundMoney(line.Amount); got != wantAmount {
				t.Fatalf("expected deposit line amount %.2f, got %.2f", wantAmount, got)
			}
			return
		}
	}
	t.Fatalf("expected deposit deduction line in preview, got %+v", lines)
}

func assertNoBoardingDepositLine(t *testing.T, lines []BoardingPriceLine) {
	t.Helper()

	for _, line := range lines {
		if line.Type == "boarding_deposit" {
			t.Fatalf("expected no deposit deduction line, got %+v", line)
		}
	}
}
