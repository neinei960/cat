package model

import "gorm.io/gorm"

const (
	BoardingOrderStatusPendingCheckin = "pending_checkin"
	BoardingOrderStatusCheckedIn      = "checked_in"
	BoardingOrderStatusCheckedOut     = "checked_out"
	BoardingOrderStatusCancelled      = "cancelled"
)

type BoardingOrder struct {
	gorm.Model
	ShopID                 uint    `json:"shop_id" gorm:"not null;index"`
	OrderID                *uint   `json:"order_id" gorm:"index"`
	CustomerID             uint    `json:"customer_id" gorm:"not null;index"`
	StaffID                uint    `json:"staff_id" gorm:"not null;index"`
	CabinetID              uint    `json:"cabinet_id" gorm:"not null;index"`
	CheckInAt              string  `json:"check_in_at" gorm:"size:10;not null;index"`
	CheckOutAt             string  `json:"check_out_at" gorm:"size:10;not null;index"`
	ActualCheckOutAt       string  `json:"actual_check_out_at" gorm:"size:10"`
	Nights                 int     `json:"nights" gorm:"default:0"`
	BaseAmount             float64 `json:"base_amount" gorm:"type:decimal(10,2);default:0"`
	HolidaySurchargeAmount float64 `json:"holiday_surcharge_amount" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount         float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	ManualDiscountAmount   float64 `json:"manual_discount_amount" gorm:"type:decimal(10,2);default:0"`
	PayAmount              float64 `json:"pay_amount" gorm:"type:decimal(10,2);default:0"`
	Status                 string  `json:"status" gorm:"size:30;default:pending_checkin;index"`
	HasDeworming           *bool   `json:"has_deworming" gorm:"default:null"`
	Remark                 string  `json:"remark" gorm:"type:text"`
	PolicySnapshotJSON     string  `json:"policy_snapshot_json" gorm:"type:text"`
	PriceSnapshotJSON      string  `json:"price_snapshot_json" gorm:"type:text"`

	Order    *Order             `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Customer *Customer          `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Staff    *Staff             `json:"staff,omitempty" gorm:"foreignKey:StaffID"`
	Cabinet  *BoardingCabinet   `json:"cabinet,omitempty" gorm:"foreignKey:CabinetID"`
	Pets     []BoardingOrderPet `json:"pets,omitempty" gorm:"foreignKey:BoardingOrderID"`
	Logs     []BoardingOrderLog `json:"logs,omitempty" gorm:"foreignKey:BoardingOrderID"`
}

type BoardingOrderPet struct {
	gorm.Model
	BoardingOrderID uint   `json:"boarding_order_id" gorm:"not null;index"`
	PetID           uint   `json:"pet_id" gorm:"not null;index"`
	PetNameSnapshot string `json:"pet_name_snapshot" gorm:"size:100"`
	Remark          string `json:"remark" gorm:"size:500"`

	Pet *Pet `json:"pet,omitempty" gorm:"foreignKey:PetID"`
}

type BoardingOrderLog struct {
	gorm.Model
	BoardingOrderID uint   `json:"boarding_order_id" gorm:"not null;index"`
	OperatorID      uint   `json:"operator_id" gorm:"not null;index"`
	Action          string `json:"action" gorm:"size:50;not null"`
	Content         string `json:"content" gorm:"size:1000"`

	Operator *Staff `json:"operator,omitempty" gorm:"foreignKey:OperatorID"`
}
