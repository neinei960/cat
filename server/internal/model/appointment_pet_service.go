package model

import "gorm.io/gorm"

type AppointmentPetService struct {
	gorm.Model
	AppointmentPetID uint    `json:"appointment_pet_id" gorm:"not null;index"`
	ServiceID        uint    `json:"service_id" gorm:"not null"`
	ServiceName      string  `json:"service_name" gorm:"size:100;not null;comment:快照"`
	Price            float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Duration         int     `json:"duration" gorm:"not null;comment:分钟"`

	AppointmentPet *AppointmentPet `json:"appointment_pet,omitempty" gorm:"foreignKey:AppointmentPetID"`
	Service        *Service        `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
