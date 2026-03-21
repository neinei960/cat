package model

import "gorm.io/gorm"

type FurCategory struct {
	gorm.Model
	ShopID    uint   `json:"shop_id" gorm:"not null;index"`
	Name      string `json:"name" gorm:"size:20;not null;comment:类别名称如短毛猫/长毛猫/A/B/C/D"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
	Status    int    `json:"status" gorm:"default:1;comment:1启用 2停用"`

	Shop *Shop `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
}
