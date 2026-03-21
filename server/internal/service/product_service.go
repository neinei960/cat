package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(repo *repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(p *model.Product) error {
	return s.repo.Create(p)
}

func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	return s.repo.FindByID(id)
}

func (s *ProductService) List(shopID uint, categoryID uint, keyword string, page, pageSize int) ([]model.Product, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShopID(shopID, categoryID, keyword, page, pageSize)
}

func (s *ProductService) Update(p *model.Product) error {
	return s.repo.Update(p)
}

func (s *ProductService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProductService) GetBrands(shopID uint) ([]string, error) {
	return s.repo.GetBrands(shopID)
}

func (s *ProductService) ReplaceSKUs(productID uint, skus []model.ProductSKU) error {
	return s.repo.ReplaceSKUs(productID, skus)
}
