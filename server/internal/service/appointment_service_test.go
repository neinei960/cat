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

func TestCreateMultiAllowsFiveMinuteManualEndTime(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138002", Nickname: "短项目客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "短毛", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	svcModel := model.Service{ShopID: 1, Name: "刷牙剪指甲", BasePrice: 20, Duration: 30, Status: 1}
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
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      pet.ID,
		Date:       "2026-06-02",
		StartTime:  "12:00",
		EndTime:    "12:10",
		Status:     1,
		Source:     2,
	}

	if err := svc.CreateMulti(appt, []AppointmentPetSelection{{PetID: pet.ID, ServiceIDs: []uint{svcModel.ID}}}); err != nil {
		t.Fatalf("create short appointment: %v", err)
	}

	var saved model.Appointment
	if err := database.DB.First(&saved, appt.ID).Error; err != nil {
		t.Fatalf("load appointment: %v", err)
	}
	if saved.EndTime != "12:10" {
		t.Fatalf("expected short end time 12:10, got %q", saved.EndTime)
	}
}

func TestUpdateStatusAllowsNoShowBackToConfirmed(t *testing.T) {
	setupOrderServiceTestDB(t)
	if err := database.DB.AutoMigrate(&model.AppointmentStatusLog{}); err != nil {
		t.Fatalf("migrate appointment status log: %v", err)
	}

	customer := model.Customer{ShopID: 1, Phone: "13800138003", Nickname: "恢复确认客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "恢复确认猫", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	appt := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      pet.ID,
		Date:       "2026-06-18",
		StartTime:  "15:00",
		EndTime:    "16:00",
		Status:     5,
		Source:     2,
	}
	if err := database.DB.Create(&appt).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	svc := NewAppointmentService(
		repository.NewAppointmentRepository(),
		nil,
		repository.NewServiceRepository(),
		nil,
		nil,
	)
	if err := svc.UpdateStatusWithOperator(appt.ID, 1, "恢复为已确认", "", "", 9); err != nil {
		t.Fatalf("restore no-show appointment to confirmed: %v", err)
	}

	var saved model.Appointment
	if err := database.DB.First(&saved, appt.ID).Error; err != nil {
		t.Fatalf("load appointment: %v", err)
	}
	if saved.Status != 1 {
		t.Fatalf("expected status 1 after restore, got %d", saved.Status)
	}

	var log model.AppointmentStatusLog
	if err := database.DB.Where("appointment_id = ?", appt.ID).First(&log).Error; err != nil {
		t.Fatalf("load status log: %v", err)
	}
	if log.FromStatus != 5 || log.ToStatus != 1 || log.OperatorID != 9 {
		t.Fatalf("unexpected status log: from=%d to=%d operator=%d", log.FromStatus, log.ToStatus, log.OperatorID)
	}
}

func TestUpdateNotesAllowsBilledAppointment(t *testing.T) {
	setupOrderServiceTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138004", Nickname: "已开单备注客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "已开单猫", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	appt := model.Appointment{
		ShopID:     1,
		CustomerID: customer.ID,
		PetID:      pet.ID,
		Date:       "2026-07-04",
		StartTime:  "14:00",
		EndTime:    "15:00",
		Status:     7,
		Source:     2,
		Notes:      "旧备注",
	}
	if err := database.DB.Create(&appt).Error; err != nil {
		t.Fatalf("create appointment: %v", err)
	}

	svc := NewAppointmentService(
		repository.NewAppointmentRepository(),
		nil,
		repository.NewServiceRepository(),
		nil,
		nil,
	)
	if err := svc.UpdateNotes(appt.ID, 1, "已开单后补充备注"); err != nil {
		t.Fatalf("update billed appointment notes: %v", err)
	}

	var saved model.Appointment
	if err := database.DB.First(&saved, appt.ID).Error; err != nil {
		t.Fatalf("load appointment: %v", err)
	}
	if saved.Notes != "已开单后补充备注" {
		t.Fatalf("expected updated notes, got %q", saved.Notes)
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
