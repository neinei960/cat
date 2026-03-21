package model

import "gorm.io/gorm"

type ServiceCategory struct {
	gorm.Model
	ShopID    uint   `json:"shop_id" gorm:"not null;index"`
	ParentID  *uint  `json:"parent_id" gorm:"index;comment:父分类ID,NULL表示一级分类"`
	Name      string `json:"name" gorm:"size:50;not null"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
	Status    int    `json:"status" gorm:"default:1;comment:1启用 2禁用"`

	Parent   *ServiceCategory  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []ServiceCategory `json:"children,omitempty" gorm:"foreignKey:ParentID"`
}
