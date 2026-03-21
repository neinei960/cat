package model

import "gorm.io/gorm"

type ServicePriceRule struct {
	gorm.Model
	ServiceID uint    `json:"service_id" gorm:"not null;index"`
	Name      string  `json:"name" gorm:"size:50;comment:规格名称"`
	FurLevel  string  `json:"fur_level" gorm:"size:20;comment:毛发等级:短毛猫/长毛猫/A/B/C/D"`
	PetSize   string  `json:"pet_size" gorm:"size:20;comment:小/中/大/特大(兼容旧数据)"`
	Breed     string  `json:"breed" gorm:"size:100"`
	Price     float64 `json:"price" gorm:"type:decimal(10,2);not null"`
	Duration  int     `json:"duration" gorm:"comment:时长(分钟)"`

	Service *Service `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
