package service

import (
	"encoding/json"
	"fmt"
	"testing"

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

type boardingServiceTestState struct {
	shopID     uint
	operatorID uint
	order      model.BoardingOrder
	room       model.BoardingOrderRoom
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
