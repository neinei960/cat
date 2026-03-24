package model

import (
	"time"

	"gorm.io/gorm"
)

type Staff struct {
	gorm.Model
	ShopID         uint    `json:"shop_id" gorm:"not null;index"`
	Phone          string  `json:"phone" gorm:"size:20;uniqueIndex;not null"`
	PasswordHash   string  `json:"-" gorm:"size:255;not null"`
	Name           string  `json:"name" gorm:"size:50;not null"`
	Avatar         string  `json:"avatar" gorm:"size:500"`
	Role           string  `json:"role" gorm:"size:20;not null;default:staff;comment:admin/manager/staff"`
	Status         int     `json:"status" gorm:"default:1;comment:1在职 2离职"`
	CommissionRate        float64 `json:"commission_rate" gorm:"type:decimal(5,2);default:0;comment:洗浴提成百分比"`
	ProductCommissionRate float64 `json:"product_commission_rate" gorm:"type:decimal(5,2);default:0;comment:商品提成百分比"`
	FeedingCommissionRate float64 `json:"feeding_commission_rate" gorm:"type:decimal(5,2);default:0;comment:上门喂养提成百分比"`

	Shop *Shop `json:"shop,omitempty" gorm:"foreignKey:ShopID"`

	LastLoginAt *time.Time `json:"last_login_at"`
}
