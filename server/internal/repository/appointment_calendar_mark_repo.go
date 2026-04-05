package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

func (r *AppointmentRepository) FindCalendarMarks(shopID uint, startDate, endDate string) ([]model.AppointmentCalendarMark, error) {
	var marks []model.AppointmentCalendarMark
	err := database.DB.
		Where("shop_id = ? AND date >= ? AND date <= ?", shopID, startDate, endDate).
		Order("date ASC").
		Find(&marks).Error
	return marks, err
}

func (r *AppointmentRepository) SetCalendarMark(shopID uint, date string, marked bool) error {
	if marked {
		mark := model.AppointmentCalendarMark{
			ShopID: shopID,
			Date:   date,
		}
		return database.DB.Where("shop_id = ? AND date = ?", shopID, date).FirstOrCreate(&mark).Error
	}
	return database.DB.Where("shop_id = ? AND date = ?", shopID, date).Delete(&model.AppointmentCalendarMark{}).Error
}
