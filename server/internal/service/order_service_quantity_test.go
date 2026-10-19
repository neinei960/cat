package service

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

func TestOrderServiceQuantitySurvivesAppointmentSync(t *testing.T) {
	setupOrderServiceTestDB(t)
	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{ShopID: 1, CustomerPhone: "13800138991", AppointmentAmount: 40})
	duration := 5
	items := []model.OrderItem{{PetID: &state.appointment.PetID, ItemType: 1, ItemID: 19, Name: "刷牙 · 自带牙刷", Quantity: 2, UnitPrice: 20, Amount: 40, Duration: &duration}}
	if err := syncAppointmentServicesFromOrderItems(database.DB, state.appointment.ID, &state.appointment.PetID, items); err != nil {
		t.Fatal(err)
	}
	var services []model.AppointmentPetService
	if err := database.DB.Find(&services).Error; err != nil {
		t.Fatal(err)
	}
	if len(services) != 2 {
		t.Fatalf("expected two per-visit snapshots, got %+v", services)
	}
	for _, svc := range services {
		if svc.Price != 20 || svc.Duration != 5 || svc.ServiceName != "刷牙 · 自带牙刷" {
			t.Fatalf("unit price lost: %+v", svc)
		}
	}
}

func TestBatchOrderStoresQuantityAndDuration(t *testing.T) {
	setupOrderServiceTestDB(t)
	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{ShopID: 1, CustomerPhone: "13800138992", AppointmentAmount: 40})
	svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
	orders, err := svc.CreateSplitFromAppointment(state.appointment.ID, false, nil, map[uint]PetOverrideData{
		state.appointment.PetID: {Services: []ServiceOverride{{ServiceID: 19, ServiceName: "刷牙(自带牙刷)", Price: 20, Quantity: 2, Duration: 5, ServiceSpecID: 42}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	order, err := svc.GetByID(orders[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if order.PayAmount != 40 || len(order.Items) != 1 || order.Items[0].Quantity != 2 || order.Items[0].UnitPrice != 20 || order.Items[0].ServiceSpecID != 42 || order.Items[0].Duration == nil || *order.Items[0].Duration != 5 {
		t.Fatalf("lost quantity/spec: %+v", order)
	}
}

func TestAnnualPackageQuantityLimit(t *testing.T) {
	if err := ValidateServiceQuantities([]model.OrderItem{{ItemType: 1, Name: "刷牙(全年口腔护理)", Quantity: 2}}); err == nil {
		t.Fatal("annual package accepted quantity two")
	}
	if err := ValidateServiceQuantities([]model.OrderItem{{ItemType: 1, Name: "刷牙(已购口腔护理年卡)", Quantity: 2}}); err != nil {
		t.Fatal(err)
	}
}
