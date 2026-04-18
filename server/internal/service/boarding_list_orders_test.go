package service

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

func TestListOrdersSupportsDateAndCabinetFilters(t *testing.T) {
	setupBoardingServiceTestDB(t)

	shop := model.Shop{Name: "测试门店"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}

	customer := model.Customer{
		ShopID:       shop.ID,
		Phone:        "13800138000",
		Nickname:     "历史客户",
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

	cabinetA := model.BoardingCabinet{
		ShopID:      shop.ID,
		Code:        "warm-room",
		CabinetType: "康娜温柔乡",
		RoomCount:   3,
		Capacity:    1,
		BasePrice:   65,
		Status:      model.BoardingCabinetStatusEnabled,
	}
	if err := database.DB.Create(&cabinetA).Error; err != nil {
		t.Fatalf("create cabinet A: %v", err)
	}

	cabinetB := model.BoardingCabinet{
		ShopID:      shop.ID,
		Code:        "cool-room",
		CabinetType: "波妞的游乐园",
		RoomCount:   2,
		Capacity:    1,
		BasePrice:   95,
		Status:      model.BoardingCabinetStatusEnabled,
	}
	if err := database.DB.Create(&cabinetB).Error; err != nil {
		t.Fatalf("create cabinet B: %v", err)
	}

	orderA := model.BoardingOrder{
		ShopID:           shop.ID,
		CustomerID:       customer.ID,
		StaffID:          operator.ID,
		CabinetID:        cabinetA.ID,
		CheckInAt:        "2026-04-09",
		CheckOutAt:       "2026-04-17",
		ActualCheckOutAt: "2026-04-17",
		Nights:           8,
		PayAmount:        455,
		Status:           model.BoardingOrderStatusCheckedOut,
	}
	if err := database.DB.Create(&orderA).Error; err != nil {
		t.Fatalf("create order A: %v", err)
	}
	if err := database.DB.Create(&model.BoardingOrderRoom{
		BoardingOrderID:  orderA.ID,
		CabinetID:        cabinetA.ID,
		RoomIndex:        1,
		CheckInAt:        orderA.CheckInAt,
		CheckOutAt:       orderA.CheckOutAt,
		ActualCheckOutAt: orderA.ActualCheckOutAt,
		Nights:           orderA.Nights,
		PayAmount:        orderA.PayAmount,
		Status:           model.BoardingOrderStatusCheckedOut,
	}).Error; err != nil {
		t.Fatalf("create room A: %v", err)
	}

	orderB := model.BoardingOrder{
		ShopID:           shop.ID,
		CustomerID:       customer.ID,
		StaffID:          operator.ID,
		CabinetID:        cabinetB.ID,
		CheckInAt:        "2026-04-16",
		CheckOutAt:       "2026-04-17",
		ActualCheckOutAt: "2026-04-17",
		Nights:           1,
		PayAmount:        65,
		Status:           model.BoardingOrderStatusCheckedOut,
	}
	if err := database.DB.Create(&orderB).Error; err != nil {
		t.Fatalf("create order B: %v", err)
	}
	if err := database.DB.Create(&model.BoardingOrderRoom{
		BoardingOrderID:  orderB.ID,
		CabinetID:        cabinetB.ID,
		RoomIndex:        1,
		CheckInAt:        orderB.CheckInAt,
		CheckOutAt:       orderB.CheckOutAt,
		ActualCheckOutAt: orderB.ActualCheckOutAt,
		Nights:           orderB.Nights,
		PayAmount:        orderB.PayAmount,
		Status:           model.BoardingOrderStatusCheckedOut,
	}).Error; err != nil {
		t.Fatalf("create room B: %v", err)
	}

	svc := NewBoardingService(
		repository.NewBoardingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	list, total, err := svc.ListOrders(shop.ID, model.BoardingOrderStatusCheckedOut, "2026-04-15", "2026-04-15", 0, 1, 20)
	if err != nil {
		t.Fatalf("list orders by date: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 order for date filter, got %d", total)
	}
	if len(list) != 1 || list[0].ID != orderA.ID {
		t.Fatalf("expected only order A for date filter, got %+v", list)
	}

	list, total, err = svc.ListOrders(shop.ID, model.BoardingOrderStatusCheckedOut, "", "", cabinetB.ID, 1, 20)
	if err != nil {
		t.Fatalf("list orders by cabinet: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 order for cabinet filter, got %d", total)
	}
	if len(list) != 1 || list[0].ID != orderB.ID {
		t.Fatalf("expected only order B for cabinet filter, got %+v", list)
	}
}
