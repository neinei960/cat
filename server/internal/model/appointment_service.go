package model

import "gorm.io/gorm"

type AppointmentService struct {
	gorm.Model
	AppointmentID uint    `json:"appointment_id" gorm:"not null;index"`
	ServiceID     uint    `json:"service_id" gorm:"not null"`
	ServiceName   string  `json:"service_name" gorm:"size:100;not null;comment:快照"`
	Price         float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Duration      int     `json:"duration" gorm:"not null;comment:分钟"`

	Appointment *Appointment `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`
	Service     *Service     `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
