package model

import "gorm.io/gorm"

type StaffService struct {
	gorm.Model
	StaffID   uint `json:"staff_id" gorm:"not null;uniqueIndex:idx_staff_service"`
	ServiceID uint `json:"service_id" gorm:"not null;uniqueIndex:idx_staff_service"`

	Staff   *Staff   `json:"staff,omitempty" gorm:"foreignKey:StaffID"`
	Service *Service `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
