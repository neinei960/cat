package model

import (
	"time"

	"gorm.io/gorm"
)

type PetBathReport struct {
	gorm.Model
	ShopID   uint       `json:"shop_id" gorm:"not null;index"`
	PetID    uint       `json:"pet_id" gorm:"not null;index"`
	ImageURL string     `json:"image_url" gorm:"size:500;not null"`
	BathDate *time.Time `json:"bath_date" gorm:"index;comment:洗浴日期"`
	SortOrder int       `json:"sort_order" gorm:"default:0;index;comment:手动排序"`

	Pet *Pet `json:"pet,omitempty" gorm:"foreignKey:PetID"`
}
