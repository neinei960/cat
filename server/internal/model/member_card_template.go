package model

import "gorm.io/gorm"

type MemberCardTemplate struct {
	gorm.Model
	ShopID       uint    `json:"shop_id" gorm:"not null;index"`
	Name         string  `json:"name" gorm:"size:50;not null;comment:卡名称如I/II/III/IV"`
	MinRecharge  float64 `json:"min_recharge" gorm:"type:decimal(10,2);not null;comment:储值门槛"`
	DiscountRate        float64 `json:"discount_rate" gorm:"type:decimal(3,2);not null;default:1;comment:服务折扣率0.8=八折"`
	ProductDiscountRate float64 `json:"product_discount_rate" gorm:"type:decimal(3,2);not null;default:1;comment:商品折扣率0.8=八折"`
	ValidDays           int     `json:"valid_days" gorm:"default:0;comment:有效天数0=永久"`
	Color        string  `json:"color" gorm:"size:100;default:'linear-gradient(135deg, #4F46E5, #7C3AED)';comment:卡片渐变色"`
	Status       int     `json:"status" gorm:"default:1;comment:0停售 1在售"`
	SortOrder    int     `json:"sort_order" gorm:"default:0"`

	Discounts []MemberCardDiscount `json:"discounts,omitempty" gorm:"foreignKey:TemplateID"`
}
