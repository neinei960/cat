package model

import "gorm.io/gorm"

type BoardingHoliday struct {
	gorm.Model
	ShopID      uint   `json:"shop_id" gorm:"not null;index"`
	HolidayDate string `json:"holiday_date" gorm:"size:10;not null;index"`
	Name        string `json:"name" gorm:"size:100"`
}
