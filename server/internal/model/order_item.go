package model

import "gorm.io/gorm"

type OrderItem struct {
	gorm.Model
	OrderID   uint    `json:"order_id" gorm:"not null;index"`
	ItemType  int     `json:"item_type" gorm:"not null;comment:1服务 2商品 3附加"`
	ItemID    uint    `json:"item_id" gorm:"not null"`
	Name      string  `json:"name" gorm:"size:100;not null"`
	Quantity  int     `json:"quantity" gorm:"default:1"`
	UnitPrice float64 `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	Amount    float64 `json:"amount" gorm:"type:decimal(10,2);not null"`

	Order *Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`
}
