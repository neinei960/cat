package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type ScheduleRepository struct{}

func NewScheduleRepository() *ScheduleRepository {
	return &ScheduleRepository{}
}

func (r *ScheduleRepository) Upsert(schedule *model.StaffSchedule) error {
	return database.DB.Where("staff_id = ? AND date = ?", schedule.StaffID, schedule.Date).
		Assign(*schedule).FirstOrCreate(schedule).Error
}

func (r *ScheduleRepository) FindByStaffAndDateRange(staffID uint, startDate, endDate string) ([]model.StaffSchedule, error) {
	var schedules []model.StaffSchedule
	err := database.DB.Where("staff_id = ? AND date >= ? AND date <= ?", staffID, startDate, endDate).
		Order("date ASC").Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) FindByShopAndDate(shopID uint, date string) ([]model.StaffSchedule, error) {
	var schedules []model.StaffSchedule
	err := database.DB.Preload("Staff").Where("shop_id = ? AND date = ? AND is_day_off = false", shopID, date).
		Find(&schedules).Error
	return schedules, err
}

func (r *ScheduleRepository) BatchUpsert(schedules []model.StaffSchedule) error {
	tx := database.DB.Begin()
	for _, s := range schedules {
		schedule := s
		if err := tx.Where("staff_id = ? AND date = ?", schedule.StaffID, schedule.Date).
			Assign(schedule).FirstOrCreate(&schedule).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
