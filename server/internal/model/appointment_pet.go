package model

import "gorm.io/gorm"

type AppointmentPet struct {
	gorm.Model
	AppointmentID uint    `json:"appointment_id" gorm:"not null;index"`
	PetID         uint    `json:"pet_id" gorm:"not null;index"`
	SortOrder     int     `json:"sort_order" gorm:"default:0"`
	TotalAmount   float64 `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	TotalDuration int     `json:"total_duration" gorm:"default:0;comment:分钟"`

	Appointment *Appointment            `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`
	Pet         *Pet                    `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Services    []AppointmentPetService `json:"services,omitempty" gorm:"foreignKey:AppointmentPetID"`
}
