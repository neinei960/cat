package model

import "gorm.io/gorm"

type ServiceDiscount struct {
	gorm.Model
	ServiceID     uint    `json:"service_id" gorm:"not null;index"`
	Type          int     `json:"type" gorm:"not null;comment:1满天折扣 2住N免M"`
	MinDays       int     `json:"min_days" gorm:"not null;comment:满N天触发"`
	DiscountPrice float64 `json:"discount_price" gorm:"type:decimal(10,2);comment:type=1时的优惠单价"`
	FreeDays      int     `json:"free_days" gorm:"default:0;comment:type=2时免M天"`
	IsHoliday     bool    `json:"is_holiday" gorm:"default:false;comment:是否节假日策略"`
	Status        int     `json:"status" gorm:"default:1;comment:1启用 2停用"`

	Service *Service `json:"service,omitempty" gorm:"foreignKey:ServiceID"`
}
