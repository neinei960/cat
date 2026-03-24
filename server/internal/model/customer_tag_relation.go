package model

type CustomerTagRelation struct {
	CustomerID uint `json:"customer_id" gorm:"primaryKey"`
	TagID      uint `json:"tag_id" gorm:"primaryKey"`
}
