package model

import "gorm.io/gorm"

type ServiceAddon struct {
	gorm.Model
	ShopID       uint    `json:"shop_id" gorm:"not null;index"`
	Name         string  `json:"name" gorm:"size:50;not null;comment:超重费/去油费/药浴/刷牙/开结/攻击费/春节加收"`
	DefaultPrice float64 `json:"default_price" gorm:"type:decimal(10,2);default:0;comment:默认金额"`
	IsVariable   bool    `json:"is_variable" gorm:"default:false;comment:金额是否可变"`
	SortOrder    int     `json:"sort_order" gorm:"default:0"`
	Status       int     `json:"status" gorm:"default:1;comment:1启用 2停用"`

	Shop *Shop `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
}
