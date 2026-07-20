package handler

import (
	"errors"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
)

func TestBuildOrderDraftTreatsSingleGroupMergedOrderAsBatch(t *testing.T) {
	h := &OrderHandler{}
	appointmentID := uint(159)

	existing := &model.Order{
		AppointmentID: &appointmentID,
		PetID:         nil,
		PetGroups: []model.OrderPetGroup{
			{PetName: "皮卡"},
		},
	}
	req := createOrderReq{
		Items: []orderItemInput{
			{
				ItemType:  1,
				ItemID:    6,
				Name:      "皮卡 · 日常皮毛护理",
				Quantity:  1,
				UnitPrice: 88,
			},
		},
	}

	_, _, err := h.buildOrderDraft(1, req, existing)
	if !errors.Is(err, errDraftStaffRequired) {
		t.Fatalf("expected merged order to skip pet-required validation and fail on missing staff, got %v", err)
	}
}

func TestBuildOrderDraftPreservesBoardingAmountsWhenAddingProducts(t *testing.T) {
	h := &OrderHandler{}

	existing := &model.Order{
		OrderKind:                         "boarding",
		ServiceTotal:                      255,
		ProductTotal:                      68,
		AddonTotal:                        0,
		DiscountAmount:                    0,
		ServiceDiscountAmount:             0,
		ProductDiscountAmount:             0,
		AppointmentDepositAmount:          200,
		AppointmentDepositDeductionAmount: 200,
		PayAmount:                         123,
		Items: []model.OrderItem{
			{ItemType: 4, Name: "房间1 · 娜娜米的度假屋 寄养住宿", Quantity: 3, UnitPrice: 85, Amount: 255},
			{ItemType: 6, Name: "定金抵扣", Quantity: 1, UnitPrice: -200, Amount: -200},
			{ItemType: 2, Name: "柴柒·体内外同驱 · 单支", Quantity: 1, UnitPrice: 68, Amount: 68},
		},
	}
	req := createOrderReq{
		Items: []orderItemInput{
			{
				ItemType:  2,
				ItemID:    88,
				Name:      "大威宠·体内外同驱 · 幼猫",
				Quantity:  2,
				UnitPrice: 48,
			},
		},
	}

	order, items, err := h.buildOrderDraft(1, req, existing)
	if err != nil {
		t.Fatalf("build boarding retail draft: %v", err)
	}
	if got := order.ServiceTotal; got != 255 {
		t.Fatalf("expected preserved boarding service total 255, got %.2f", got)
	}
	if got := order.ProductTotal; got != 96 {
		t.Fatalf("expected updated product total 96, got %.2f", got)
	}
	if got := order.TotalAmount; got != 351 {
		t.Fatalf("expected combined total amount 351, got %.2f", got)
	}
	if got := order.PayAmount; got != 151 {
		t.Fatalf("expected combined pay amount 151 after deposit deduction, got %.2f", got)
	}
	if len(items) != 3 {
		t.Fatalf("expected preserved boarding items plus updated product item, got %d items", len(items))
	}
	if items[0].ItemType != 4 || items[1].ItemType != 6 || items[2].ItemType != 2 {
		t.Fatalf("unexpected merged item types: %+v", items)
	}
}

func TestBuildOrderDraftBindsExistingCustomerByPhone(t *testing.T) {
	setupCustomerTestDB(t)
	customer := seedCustomer(t, 1, "024578", "老家长")
	h := &OrderHandler{
		customerService: service.NewCustomerService(repository.NewCustomerRepository()),
	}

	req := createOrderReq{
		CustomerPhone: "024578",
		Items: []orderItemInput{
			{
				ItemType:  2,
				ItemID:    88,
				Name:      "零售商品",
				Quantity:  1,
				UnitPrice: 20,
			},
		},
	}

	order, _, err := h.buildOrderDraft(1, req, nil)
	if err != nil {
		t.Fatalf("build retail draft: %v", err)
	}
	if order.CustomerID == nil || *order.CustomerID != customer.ID {
		t.Fatalf("expected customer_id %d from phone match, got %v", customer.ID, order.CustomerID)
	}
}
