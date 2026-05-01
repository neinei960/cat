package model

import "gorm.io/gorm"

type BoardingSpecialItem struct {
	gorm.Model
	ShopID            uint    `json:"shop_id" gorm:"not null;index"`
	Name              string  `json:"name" gorm:"size:100;not null"`
	DefaultDailyPrice float64 `json:"default_daily_price" gorm:"type:decimal(10,2);default:0"`
	SortOrder         int     `json:"sort_order" gorm:"default:0"`
	Status            int     `json:"status" gorm:"default:1;comment:1启用 2停用"`
	Remark            string  `json:"remark" gorm:"size:500"`
}
