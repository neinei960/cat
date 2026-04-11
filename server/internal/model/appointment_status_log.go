package model

import "gorm.io/gorm"

type AppointmentStatusLog struct {
	gorm.Model
	ShopID        uint   `json:"shop_id" gorm:"not null;index"`
	AppointmentID uint   `json:"appointment_id" gorm:"not null;index"`
	OperatorID    uint   `json:"operator_id" gorm:"default:0;index"`
	FromStatus    int    `json:"from_status" gorm:"default:0"`
	ToStatus      int    `json:"to_status" gorm:"default:0;index"`
	Action        string `json:"action" gorm:"size:50;not null"`
	Note          string `json:"note" gorm:"size:1000"`

	Operator *Staff `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
}
