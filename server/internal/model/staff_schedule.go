package model

import "gorm.io/gorm"

type StaffSchedule struct {
	gorm.Model
	StaffID     uint   `json:"staff_id" gorm:"not null;uniqueIndex:idx_staff_date"`
	ShopID      uint   `json:"shop_id" gorm:"not null;index"`
	Date        string `json:"date" gorm:"size:10;not null;uniqueIndex:idx_staff_date;comment:YYYY-MM-DD"`
	StartTime   string `json:"start_time" gorm:"size:5;not null;comment:HH:MM"`
	EndTime     string `json:"end_time" gorm:"size:5;not null;comment:HH:MM"`
	BreakStart  string `json:"break_start" gorm:"size:5;comment:HH:MM"`
	BreakEnd    string `json:"break_end" gorm:"size:5;comment:HH:MM"`
	MaxCapacity int    `json:"max_capacity" gorm:"default:1;comment:同时服务数"`
	IsDayOff    bool   `json:"is_day_off" gorm:"default:false"`

	Staff *Staff `json:"staff,omitempty" gorm:"foreignKey:StaffID"`
	Shop  *Shop  `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
}
