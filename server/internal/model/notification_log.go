package model

import "gorm.io/gorm"

type NotificationLog struct {
	gorm.Model
	ShopID     uint   `json:"shop_id" gorm:"not null;index"`
	CustomerID uint   `json:"customer_id" gorm:"not null;index"`
	Scene      string `json:"scene" gorm:"size:20;not null;comment:confirm/reminder/complete/cancel"`
	Channel    string `json:"channel" gorm:"size:20;not null;comment:wechat/sms"`
	Content    string `json:"content" gorm:"type:text"`
	Status     int    `json:"status" gorm:"default:0;comment:0发送中 1成功 2失败"`
	ErrorMsg   string `json:"error_msg" gorm:"size:500"`
}
