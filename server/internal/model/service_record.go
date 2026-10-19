package model

import "gorm.io/gorm"

// ServiceRecord 服务完成后的记录（技师填写）
type ServiceRecord struct {
	gorm.Model
	ShopID          uint   `json:"shop_id" gorm:"not null;index"`
	AppointmentID   uint   `json:"appointment_id" gorm:"not null;index"`
	PetID           uint   `json:"pet_id" gorm:"not null;index"`
	StaffID         uint   `json:"staff_id" gorm:"not null;index"`
	Notes           string `json:"notes" gorm:"type:text;comment:服务记录文字"`
	Photos          string `json:"photos" gorm:"type:text;comment:照片URL,逗号分隔"`
	SkinIssues      string `json:"skin_issues" gorm:"size:500;comment:皮肤问题记录"`
	DentalCondition string `json:"dental_condition" gorm:"size:500;comment:牙齿状况"`
	OtherIssues     string `json:"other_issues" gorm:"size:300;comment:其他问题"`
	FurCondition    string `json:"fur_condition" gorm:"size:100;comment:毛发状况"`
	Weight          string `json:"weight" gorm:"size:20;comment:体重"`

	Pet         *Pet         `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Staff       *Staff       `json:"staff,omitempty" gorm:"foreignKey:StaffID"`
	Appointment *Appointment `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`

	OrderItemSummary string `json:"order_item_summary,omitempty" gorm:"-"`
}
