package service

import (
	"fmt"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

func TestLegacySingleAppointmentPetSurvivesRename(t *testing.T) {
	order := &model.Order{
		Appointment: &model.Appointment{Pets: []model.AppointmentPet{{PetID: 1159, Pet: &model.Pet{Model: gorm.Model{ID: 1159}, Name: "曼尼"}}}},
		Items:       []model.OrderItem{{ItemType: 1, Name: "money/曼妮 · 基础洗护", Amount: 88}},
	}
	groups := buildOrderPetGroups(order)
	if len(groups) != 1 || groups[0].PetID == nil || *groups[0].PetID != 1159 {
		t.Fatalf("renamed pet lost identity: %+v", groups)
	}
	if groups[0].Items[0].Amount != 88 || groups[0].PetName != "money/曼妮" {
		t.Fatal("historical snapshot changed")
	}
}

func TestOrderItemPetIdentitySurvivesRenameAndDuplicateNames(t *testing.T) {
	order := &model.Order{
		Appointment: &model.Appointment{Pets: []model.AppointmentPet{
			{PetID: 1, Pet: &model.Pet{Name: "new name"}},
			{PetID: 2, Pet: &model.Pet{Name: "new name"}},
		}},
		Items: []model.OrderItem{
			{PetID: uintPtr(1), ItemType: 1, Name: "old name · wash", Amount: 88},
			{PetID: uintPtr(2), ItemType: 1, Name: "old name · wash", Amount: 108},
		},
	}
	groups := buildOrderPetGroups(order)
	if len(groups) != 2 || groups[0].PetID == nil || *groups[0].PetID != 1 || groups[1].PetID == nil || *groups[1].PetID != 2 {
		t.Fatalf("distinct pet identities merged or lost: %+v", groups)
	}
	order.Items = []model.OrderItem{{ItemType: 1, Name: "unmatched old name · wash"}}
	groups = buildOrderPetGroups(order)
	if groups[0].PetID != nil {
		t.Fatal("ambiguous legacy item assigned to an unrelated pet")
	}
}

func TestAppointmentOrderPersistsPetIdentity(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(fmt.Sprintf("batch=%t", batch), func(t *testing.T) {
			setupOrderServiceTestDB(t)
			state := seedAppointmentOrderFixture(t, seedAppointmentOrderStateInput{ShopID: 1, CustomerPhone: "13800138999", AppointmentAmount: 100})
			svc := NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository())
			var id uint
			if batch {
				orders, err := svc.CreateSplitFromAppointment(state.appointment.ID, false, nil)
				if err != nil {
					t.Fatal(err)
				}
				id = orders[0].ID
			} else {
				order, err := svc.CreateFromAppointment(state.appointment.ID, false)
				if err != nil {
					t.Fatal(err)
				}
				id = order.ID
			}
			var items []model.OrderItem
			if err := database.DB.Where("order_id = ?", id).Find(&items).Error; err != nil {
				t.Fatal(err)
			}
			if len(items) == 0 {
				t.Fatal("missing persisted items")
			}
			for _, item := range items {
				if item.PetID == nil || *item.PetID != state.appointment.PetID {
					t.Fatalf("missing stored identity: %+v", item)
				}
			}
			if err := database.DB.Model(&model.Pet{}).Where("id = ?", state.appointment.PetID).Update("name", "renamed").Error; err != nil {
				t.Fatal(err)
			}
			order, err := svc.GetByID(id)
			if err != nil {
				t.Fatal(err)
			}
			if len(order.PetGroups) != 1 || order.PetGroups[0].PetID == nil || *order.PetGroups[0].PetID != state.appointment.PetID {
				t.Fatalf("rename broke persisted identity: %+v", order.PetGroups)
			}
		})
	}
}
