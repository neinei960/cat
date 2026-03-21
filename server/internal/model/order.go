package model

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderNo        string     `json:"order_no" gorm:"size:50;uniqueIndex;not null"`
	ShopID         uint       `json:"shop_id" gorm:"not null;index"`
	CustomerID     *uint      `json:"customer_id" gorm:"index"`
	PetID          *uint      `json:"pet_id" gorm:"index;comment:关联猫咪"`
	AppointmentID  *uint      `json:"appointment_id" gorm:"index"`
	StaffID        *uint      `json:"staff_id" gorm:"index"`
	TotalAmount    float64    `json:"total_amount" gorm:"type:decimal(10,2);default:0;comment:基础+附加合计"`
	DiscountAmount float64    `json:"discount_amount" gorm:"type:decimal(10,2);default:0;comment:折扣减免金额"`
	DiscountRate   float64    `json:"discount_rate" gorm:"type:decimal(3,2);default:1;comment:会员折扣率"`
	PayAmount      float64    `json:"pay_amount" gorm:"type:decimal(10,2);default:0;comment:实付金额"`
	Commission     float64    `json:"commission" gorm:"type:decimal(10,2);default:0;comment:洗护师提成"`
	PayMethod      string     `json:"pay_method" gorm:"size:20;comment:wechat/alipay/cash/card"`
	PayStatus      int        `json:"pay_status" gorm:"default:0;comment:0未付 1已付 2已退"`
	PayTime        *time.Time `json:"pay_time"`
	TransactionID  string     `json:"transaction_id" gorm:"size:100"`
	Status         int        `json:"status" gorm:"default:0;comment:0待付 1完成 2取消 3退款"`
	Remark         string     `json:"remark" gorm:"size:500"`

	Shop        *Shop        `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	Customer    *Customer    `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Pet         *Pet         `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Appointment *Appointment `json:"appointment,omitempty" gorm:"foreignKey:AppointmentID"`
	Staff       *Staff       `json:"staff,omitempty" gorm:"foreignKey:StaffID"`

	Items []OrderItem `json:"items,omitempty" gorm:"foreignKey:OrderID"`
}
