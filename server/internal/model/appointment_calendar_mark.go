package model

import "gorm.io/gorm"

// AppointmentCalendarMark stores manually highlighted appointment dates.
type AppointmentCalendarMark struct {
	gorm.Model
	ShopID uint   `json:"shop_id" gorm:"not null;uniqueIndex:idx_appt_calendar_mark_shop_date"`
	Date   string `json:"date" gorm:"size:10;not null;uniqueIndex:idx_appt_calendar_mark_shop_date;index;comment:YYYY-MM-DD"`
}
