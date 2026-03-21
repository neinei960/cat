package model

import "gorm.io/gorm"

type Service struct {
	gorm.Model
	ShopID           uint    `json:"shop_id" gorm:"not null;index"`
	Name             string  `json:"name" gorm:"size:100;not null"`
	Category         string  `json:"category" gorm:"size:20;comment:旧分类字段,兼容"`
	CategoryID       *uint   `json:"category_id" gorm:"index;comment:服务分类ID"`
	Description      string  `json:"description" gorm:"size:500"`
	BasePrice        float64 `json:"base_price" gorm:"type:decimal(10,2);not null"`
	Duration         int     `json:"duration" gorm:"not null;comment:时长(分钟)"`
	ApplicableSpecies string `json:"applicable_species" gorm:"size:100;comment:适用物种,逗号分隔"`
	ApplicableSizes  string  `json:"applicable_sizes" gorm:"size:100;comment:适用体型,逗号分隔"`
	SortOrder        int     `json:"sort_order" gorm:"default:0"`
	Status           int     `json:"status" gorm:"default:1;comment:1启用 2停用"`

	Shop             *Shop              `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	ServiceCategory  *ServiceCategory   `json:"service_category,omitempty" gorm:"foreignKey:CategoryID"`
	PriceRules []ServicePriceRule `json:"price_rules,omitempty" gorm:"foreignKey:ServiceID"`
}
