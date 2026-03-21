package model

import "gorm.io/gorm"

type Appointment struct {
	gorm.Model
	ShopID       uint    `json:"shop_id" gorm:"not null;index"`
	CustomerID   uint    `json:"customer_id" gorm:"not null;index"`
	PetID        uint    `json:"pet_id" gorm:"not null;index"`
	StaffID      *uint   `json:"staff_id" gorm:"index;comment:可空=待分配"`
	Date         string  `json:"date" gorm:"size:10;not null;index;comment:YYYY-MM-DD"`
	StartTime    string  `json:"start_time" gorm:"size:5;not null;comment:HH:MM"`
	EndTime      string  `json:"end_time" gorm:"size:5;not null;comment:HH:MM"`
	Status       int     `json:"status" gorm:"default:0;index;comment:0待确认 1已确认 2进行中 3已完成 4已取消 5未到店"`
	Source       int     `json:"source" gorm:"default:1;comment:1小程序 2商家创建 3电话"`
	Notes        string  `json:"notes" gorm:"size:500;comment:客户备注"`
	StaffNotes   string  `json:"staff_notes" gorm:"size:500;comment:技师备注"`
	CancelReason string  `json:"cancel_reason" gorm:"size:500"`
	CancelledBy  string  `json:"cancelled_by" gorm:"size:20;comment:customer/staff"`
	TotalAmount  float64 `json:"total_amount" gorm:"type:decimal(10,2);default:0"`
	PaidAmount   float64 `json:"paid_amount" gorm:"type:decimal(10,2);default:0"`

	Shop     *Shop     `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	Customer *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Pet      *Pet      `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Staff    *Staff    `json:"staff,omitempty" gorm:"foreignKey:StaffID"`

	Services []AppointmentService `json:"services,omitempty" gorm:"foreignKey:AppointmentID"`
}
