package model

import "gorm.io/gorm"

const (
	BoardingPolicyTypeStayNFreeM       = "stay_n_free_m"
	BoardingPolicyTypeHolidaySurcharge = "holiday_surcharge"
)

type BoardingDiscountPolicy struct {
	gorm.Model
	ShopID     uint   `json:"shop_id" gorm:"not null;index"`
	Name       string `json:"name" gorm:"size:100;not null"`
	PolicyType string `json:"policy_type" gorm:"size:50;not null;index"`
	RuleJSON   string `json:"rule_json" gorm:"type:text"`
	ValidFrom  string `json:"valid_from" gorm:"size:10"`
	ValidTo    string `json:"valid_to" gorm:"size:10"`
	Priority   int    `json:"priority" gorm:"default:0"`
	Stackable  bool   `json:"stackable" gorm:"default:true"`
	Status     int    `json:"status" gorm:"default:1"`
	Remark     string `json:"remark" gorm:"size:500"`
}
