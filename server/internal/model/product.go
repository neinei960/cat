package model

import "gorm.io/gorm"

// ProductCategory 商品分类
type ProductCategory struct {
	gorm.Model
	ShopID    uint   `json:"shop_id" gorm:"not null;index"`
	Name      string `json:"name" gorm:"size:100"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status" gorm:"default:1"`
}

// Product 商品
type Product struct {
	gorm.Model
	ShopID      uint             `json:"shop_id" gorm:"not null;index"`
	CategoryID  uint             `json:"category_id" gorm:"index"`
	Name        string           `json:"name" gorm:"size:200"`
	Brand       string           `json:"brand" gorm:"size:100"`
	Description string           `json:"description" gorm:"size:500"`
	MultiSpec   bool             `json:"multi_spec"`
	Status      int              `json:"status" gorm:"default:1"`
	Category    *ProductCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	SKUs        []ProductSKU     `json:"skus,omitempty" gorm:"foreignKey:ProductID"`
}

// ProductSKU 商品规格
type ProductSKU struct {
	gorm.Model
	ProductID uint    `json:"product_id" gorm:"not null;index"`
	SpecName  string  `json:"spec_name" gorm:"size:100"`
	Price     float64 `json:"price" gorm:"type:decimal(10,2)"`
	Weight    float64 `json:"weight"`
	Sellable  bool    `json:"sellable" gorm:"default:true"`
}
