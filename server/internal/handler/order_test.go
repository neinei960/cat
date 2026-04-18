package handler

import (
	"errors"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
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
