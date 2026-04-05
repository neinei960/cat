package model

import "gorm.io/gorm"

const (
	BoardingCabinetStatusEnabled  = "enabled"
	BoardingCabinetStatusCleaning = "cleaning"
	BoardingCabinetStatusDisabled = "disabled"
)

type BoardingCabinet struct {
	gorm.Model
	ShopID         uint    `json:"shop_id" gorm:"not null;index"`
	Code           string  `json:"code" gorm:"size:50;not null"`
	CabinetType    string  `json:"cabinet_type" gorm:"size:50;not null;index"`
	RoomCount      int     `json:"room_count" gorm:"default:1"`
	Capacity       int     `json:"capacity" gorm:"default:1"`
	BasePrice      float64 `json:"base_price" gorm:"type:decimal(10,2);default:0"`
	Status         string  `json:"status" gorm:"size:20;default:enabled;index"`
	Remark         string  `json:"remark" gorm:"size:500"`
	OccupiedRooms  int     `json:"occupied_rooms" gorm:"-"`
	ReservedRooms  int     `json:"reserved_rooms" gorm:"-"`
	RemainingRooms int     `json:"remaining_rooms" gorm:"-"`
}
