package service

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestListCalendarResourcesReturnsActiveStaffWithSchedules(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:staff-calendar-resources?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&model.Staff{}, &model.StaffSchedule{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	activeA := model.Staff{ShopID: 1, Phone: "13800138001", PasswordHash: "x", Name: "A", Role: model.StaffRoleStaff, Status: 1, SortOrder: 2}
	activeB := model.Staff{ShopID: 1, Phone: "13800138002", PasswordHash: "x", Name: "B", Role: model.StaffRoleStaff, Status: 1, SortOrder: 1}
	inactive := model.Staff{ShopID: 1, Phone: "13800138003", PasswordHash: "x", Name: "离职", Role: model.StaffRoleStaff, Status: 2, SortOrder: 3}
	otherShop := model.Staff{ShopID: 2, Phone: "13800138004", PasswordHash: "x", Name: "其他店", Role: model.StaffRoleStaff, Status: 1, SortOrder: 1}
	for _, staff := range []*model.Staff{&activeA, &activeB, &inactive, &otherShop} {
		if err := database.DB.Create(staff).Error; err != nil {
			t.Fatalf("create staff: %v", err)
		}
	}
	if err := database.DB.Create(&model.StaffSchedule{
		ShopID: 1, StaffID: activeA.ID, Date: "2026-05-02", StartTime: "12:00", EndTime: "20:00", IsDayOff: false,
	}).Error; err != nil {
		t.Fatalf("create active schedule: %v", err)
	}
	if err := database.DB.Create(&model.StaffSchedule{
		ShopID: 1, StaffID: inactive.ID, Date: "2026-05-02", StartTime: "12:00", EndTime: "20:00", IsDayOff: false,
	}).Error; err != nil {
		t.Fatalf("create inactive schedule: %v", err)
	}

	svc := NewStaffService(repository.NewStaffRepository(), repository.NewScheduleRepository(), nil)
	resources, err := svc.ListCalendarResources(1, "2026-05-02")
	if err != nil {
		t.Fatalf("list calendar resources: %v", err)
	}
	if len(resources.Staffs) != 2 {
		t.Fatalf("expected 2 active shop staffs, got %d", len(resources.Staffs))
	}
	if resources.Staffs[0].ID != activeB.ID || resources.Staffs[1].ID != activeA.ID {
		t.Fatalf("expected staff sort order B,A, got %+v", resources.Staffs)
	}
	if len(resources.Schedules) != 1 {
		t.Fatalf("expected only active staff schedule, got %d", len(resources.Schedules))
	}
	if resources.Schedules[0].StaffID != activeA.ID || resources.Schedules[0].Date != "2026-05-02" {
		t.Fatalf("unexpected schedule: %+v", resources.Schedules[0])
	}
}
