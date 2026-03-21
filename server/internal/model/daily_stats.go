package model

import "gorm.io/gorm"

type DailyStats struct {
	gorm.Model
	ShopID           uint    `json:"shop_id" gorm:"not null;uniqueIndex:idx_shop_date"`
	Date             string  `json:"date" gorm:"size:10;not null;uniqueIndex:idx_shop_date;comment:YYYY-MM-DD"`
	Revenue          float64 `json:"revenue" gorm:"type:decimal(12,2);default:0"`
	OrderCount       int     `json:"order_count" gorm:"default:0"`
	AppointmentCount int     `json:"appointment_count" gorm:"default:0"`
	NewCustomerCount int     `json:"new_customer_count" gorm:"default:0"`
	CancelCount      int     `json:"cancel_count" gorm:"default:0"`
}
