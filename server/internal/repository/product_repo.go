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
	var total int64

	db := database.DB.Model(&model.Product{}).Where("products.shop_id = ?", shopID)

	if categoryID > 0 {
		db = db.Where("products.category_id = ?", categoryID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.
			Joins("LEFT JOIN product_categories ON product_categories.id = products.category_id AND product_categories.deleted_at IS NULL").
			Joins("LEFT JOIN product_skus ON product_skus.product_id = products.id AND product_skus.deleted_at IS NULL").
			Where("(products.name LIKE ? OR products.brand LIKE ? OR product_categories.name LIKE ? OR product_skus.spec_name LIKE ?)", like, like, like, like)
	}

	if err := db.Distinct("products.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var ids []uint
	if err := db.
		Select("products.id").
		Distinct().
		Order("products.id DESC").
		Offset(offset).
		Limit(pageSize).
		Pluck("products.id", &ids).Error; err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return []model.Product{}, total, nil
	}

	var products []model.Product
	if err := database.DB.
		Where("id IN ?", ids).
		Preload("SKUs").
		Preload("Category").
		Find(&products).Error; err != nil {
		return nil, 0, err
	}

	orderMap := make(map[uint]int, len(ids))
	for idx, id := range ids {
		orderMap[id] = idx
	}
	orderedProducts := make([]model.Product, len(ids))
	for _, product := range products {
		if idx, ok := orderMap[product.ID]; ok {
			orderedProducts[idx] = product
		}
	}
	return orderedProducts, total, nil
}

func (r *ProductRepository) Update(p *model.Product) error {
	return database.DB.Model(&model.Product{}).
		Where("id = ? AND shop_id = ?", p.ID, p.ShopID).
		Updates(map[string]any{
			"category_id":  p.CategoryID,
			"name":         p.Name,
			"brand":        p.Brand,
			"description":  p.Description,
			"multi_spec":   p.MultiSpec,
			"status":       p.Status,
		}).Error
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
