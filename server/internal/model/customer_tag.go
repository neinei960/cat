package model

import "gorm.io/gorm"

type CustomerTag struct {
	gorm.Model
	ShopID        uint   `json:"shop_id" gorm:"not null;index"`
	Name          string `json:"name" gorm:"size:50;not null"`
	Description   string `json:"description" gorm:"size:255"`
	Color         string `json:"color" gorm:"size:20;default:'#6366F1'"`
	SortOrder     int    `json:"sort_order" gorm:"default:0"`
	Status        int    `json:"status" gorm:"default:1;comment:0停用 1启用"`
	RelationCount int64  `json:"relation_count" gorm:"-"`
}
