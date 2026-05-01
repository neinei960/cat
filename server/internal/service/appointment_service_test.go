package service

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

func TestCreateMultiPersistsCustomerTypeSnapshot(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138001", Nickname: "老客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "奶茶", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	svcModel := model.Service{ShopID: 1, Name: "日常护理", BasePrice: 88, Duration: 60, Status: 1}
	if err := database.DB.Create(&svcModel).Error; err != nil {
		t.Fatalf("create service: %v", err)
	}

	svc := NewAppointmentService(
		repository.NewAppointmentRepository(),
		nil,
		repository.NewServiceRepository(),
		nil,
		nil,
	)
	appt := &model.Appointment{
		ShopID:       1,
		CustomerID:   customer.ID,
		PetID:        pet.ID,
		Date:         "2026-04-25",
		StartTime:    "12:00",
		Status:       1,
		Source:       2,
		CustomerType: model.AppointmentCustomerTypeNew,
	}

	if err := svc.CreateMulti(appt, []AppointmentPetSelection{{PetID: pet.ID, ServiceIDs: []uint{svcModel.ID}}}); err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	var saved model.Appointment
	if err := database.DB.First(&saved, appt.ID).Error; err != nil {
		t.Fatalf("load appointment: %v", err)
	}
	if saved.CustomerType != model.AppointmentCustomerTypeNew {
		t.Fatalf("expected new customer type snapshot, got %d", saved.CustomerType)
	}
}

func TestNormalizeAppointmentDepositForMemberCustomerReturnsZero(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:             1,
		CustomerPhone:      "13800138115",
		CustomerNickname:   "会员预约",
		AppointmentAmount:  88,
		AppointmentDeposit: 18,
		WithMemberCard:     true,
	})

	if got := normalizeAppointmentDeposit(&state.customer.ID, 18, 88); got != 0 {
		t.Fatalf("expected member appointment deposit to normalize to 0, got %.2f", got)
	}
}

func TestNormalizeAppointmentDepositClampsToRangeForNonMember(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{
		ShopID:            1,
		CustomerPhone:     "13800138116",
		CustomerNickname:  "普通预约",
		AppointmentAmount: 88,
	})

	if got := normalizeAppointmentDeposit(&state.customer.ID, -5, 88); got != 0 {
		t.Fatalf("expected negative deposit to clamp to 0, got %.2f", got)
	}
	if got := normalizeAppointmentDeposit(&state.customer.ID, 120, 88); got != 88 {
		t.Fatalf("expected deposit to clamp to total amount 88, got %.2f", got)
	}
}
