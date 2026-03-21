package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type ProductRepository struct{}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{}
}

func (r *ProductRepository) Create(p *model.Product) error {
	return database.DB.Create(p).Error
}

func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
	var p model.Product
	err := database.DB.Preload("SKUs").Preload("Category").First(&p, id).Error
	return &p, err
}

func (r *ProductRepository) FindByShopID(shopID uint, categoryID uint, keyword string, page, pageSize int) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	db := database.DB.Model(&model.Product{}).Where("shop_id = ?", shopID)

	if categoryID > 0 {
		db = db.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("(name LIKE ? OR brand LIKE ?)", like, like)
	}

	db.Count(&total)

	offset := (page - 1) * pageSize
	err := db.Preload("SKUs").Preload("Category").
		Order("id DESC").Offset(offset).Limit(pageSize).
		Find(&products).Error
	return products, total, err
}

func (r *ProductRepository) Update(p *model.Product) error {
	return database.DB.Save(p).Error
}

func (r *ProductRepository) Delete(id uint) error {
	tx := database.DB.Begin()
	if err := tx.Where("product_id = ?", id).Delete(&model.ProductSKU{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.Product{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *ProductRepository) GetBrands(shopID uint) ([]string, error) {
	var brands []string
	err := database.DB.Model(&model.Product{}).
		Where("shop_id = ? AND brand != ''", shopID).
		Distinct("brand").Pluck("brand", &brands).Error
	return brands, err
}

func (r *ProductRepository) ReplaceSKUs(productID uint, skus []model.ProductSKU) error {
	tx := database.DB.Begin()
	// 硬删旧 SKU
	if err := tx.Unscoped().Where("product_id = ?", productID).Delete(&model.ProductSKU{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range skus {
		skus[i].ProductID = productID
		skus[i].ID = 0
		if err := tx.Create(&skus[i]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
