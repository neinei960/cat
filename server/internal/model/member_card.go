package model

import (
	"time"

	"gorm.io/gorm"
)

type MemberCard struct {
	gorm.Model
	ShopID         uint       `json:"shop_id" gorm:"not null;index"`
	CustomerID     uint       `json:"customer_id" gorm:"not null;uniqueIndex"`
	TemplateID     uint       `json:"template_id" gorm:"not null;index"`
	CardName       string     `json:"card_name" gorm:"size:50;comment:开卡时的卡名"`
	CardType       int        `json:"card_type" gorm:"default:1;comment:1储值卡 2次卡"`
	Balance        float64    `json:"balance" gorm:"type:decimal(10,2);default:0"`
	TotalRecharge  float64    `json:"total_recharge" gorm:"type:decimal(10,2);default:0"`
	TotalSpent     float64    `json:"total_spent" gorm:"type:decimal(10,2);default:0"`
	TotalTimes     int        `json:"total_times" gorm:"default:0;comment:次卡总次数"`
	UsedTimes      int        `json:"used_times" gorm:"default:0;comment:次卡已用次数"`
	RemainingTimes int        `json:"remaining_times" gorm:"-"` // 计算字段
	DiscountRate        float64    `json:"discount_rate" gorm:"type:decimal(3,2);not null;default:1"`
	ProductDiscountRate float64    `json:"product_discount_rate" gorm:"type:decimal(3,2);not null;default:1"`
	ExpireAt       *time.Time `json:"expire_at" gorm:"comment:过期时间null=永久"`
	Status         int        `json:"status" gorm:"default:1;comment:0冻结 1正常"`

	Customer *Customer          `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Template *MemberCardTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}
