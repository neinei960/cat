package model

import (
	"time"

	"gorm.io/gorm"
)

type Pet struct {
	gorm.Model
	ShopID         uint       `json:"shop_id" gorm:"not null;index"`
	CustomerID     *uint      `json:"customer_id" gorm:"index;comment:可选,猫咪可无主人"`
	Name           string     `json:"name" gorm:"size:100;not null"`
	Species        string     `json:"species" gorm:"size:20;not null;default:猫;comment:犬/猫"`
	Breed          string     `json:"breed" gorm:"size:100"`
	Gender         int        `json:"gender" gorm:"default:0;comment:0未知 1公 2母"`
	BirthDate      *time.Time `json:"birth_date"`
	Weight         float64    `json:"weight" gorm:"type:decimal(5,2);comment:体重kg"`
	CoatType       string     `json:"coat_type" gorm:"size:20;comment:短毛/长毛/卷毛"`
	CoatColor      string     `json:"coat_color" gorm:"size:50"`
	FurLevel       string     `json:"fur_level" gorm:"size:20;comment:毛发等级:短毛猫/长毛猫/A/B/C/D(定价关键)"`
	Personality    string     `json:"personality" gorm:"size:20;comment:性格:胆小敏感/神仙宝贝/笑里藏刀/绝世凶兽/过度活跃/胆大开放"`
	Aggression     string     `json:"aggression" gorm:"size:10;comment:攻击性:无/可能/有"`
	ForbiddenZones string     `json:"forbidden_zones" gorm:"size:200;comment:禁区-不能碰的部位"`
	BathFrequency  string     `json:"bath_frequency" gorm:"size:20;comment:洗澡频率:每月/两月/三月/半年等"`
	Neutered       bool       `json:"neutered" gorm:"default:false"`
	Avatar         string     `json:"avatar" gorm:"size:500"`
	CareNotes      string     `json:"care_notes" gorm:"type:text;comment:洗护注意事项"`
	BehaviorNotes  string     `json:"behavior_notes" gorm:"type:text;comment:行为备注"`
	Status         int        `json:"status" gorm:"default:1;comment:1正常 2已故 3转让"`

	Shop     *Shop     `json:"shop,omitempty" gorm:"foreignKey:ShopID"`
	Customer *Customer `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
}
