package model

import (
	"encoding/json"

	"gorm.io/gorm"
)

type BusinessHours struct {
	Monday    []string `json:"monday"`
	Tuesday   []string `json:"tuesday"`
	Wednesday []string `json:"wednesday"`
	Thursday  []string `json:"thursday"`
	Friday    []string `json:"friday"`
	Saturday  []string `json:"saturday"`
	Sunday    []string `json:"sunday"`
}

type Shop struct {
	gorm.Model
	Name          string          `json:"name" gorm:"size:100;not null"`
	Logo          string          `json:"logo" gorm:"size:500"`
	Phone         string          `json:"phone" gorm:"size:20"`
	Address       string          `json:"address" gorm:"size:500"`
	Latitude      float64         `json:"latitude" gorm:"type:decimal(10,7)"`
	Longitude     float64         `json:"longitude" gorm:"type:decimal(10,7)"`
	BusinessHours json.RawMessage `json:"business_hours" gorm:"type:json;comment:营业时间"`
	OpenTime      string          `json:"open_time" gorm:"size:10;default:10:00;comment:营业开始时间"`
	CloseTime     string          `json:"close_time" gorm:"size:10;default:22:00;comment:营业结束时间"`
	Status        int             `json:"status" gorm:"default:1;comment:1正常 2停业"`
}
