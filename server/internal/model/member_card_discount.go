package model

import "gorm.io/gorm"

// MemberCardDiscount stores per-category discount rates for a card template
type MemberCardDiscount struct {
	gorm.Model
	TemplateID   uint    `json:"template_id" gorm:"not null;index;comment:会员卡模板ID"`
	CategoryID   uint    `json:"category_id" gorm:"not null;index;comment:服务一级分类ID"`
	CategoryName string  `json:"category_name" gorm:"size:50;comment:分类名称(冗余)"`
	DiscountRate float64 `json:"discount_rate" gorm:"type:decimal(3,2);not null;comment:折扣率如0.8"`
}
