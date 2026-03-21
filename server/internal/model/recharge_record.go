package model

import "gorm.io/gorm"

type RechargeRecord struct {
	gorm.Model
	ShopID       uint    `json:"shop_id" gorm:"not null;index"`
	CustomerID   uint    `json:"customer_id" gorm:"not null;index"`
	CardID       uint    `json:"card_id" gorm:"not null;index"`
	Type         int     `json:"type" gorm:"not null;comment:1充值 2消费 3退款"`
	Amount       float64 `json:"amount" gorm:"type:decimal(10,2);not null"`
	BalanceAfter float64 `json:"balance_after" gorm:"type:decimal(10,2)"`
	OrderID      *uint   `json:"order_id" gorm:"index;comment:关联订单ID"`
	Remark       string  `json:"remark" gorm:"size:200"`
	OperatorID   *uint   `json:"operator_id" gorm:"comment:操作员工ID"`
}
