package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type ShopRepository struct{}

func NewShopRepository() *ShopRepository {
	return &ShopRepository{}
}

func (r *ShopRepository) Create(shop *model.Shop) error {
	return database.DB.Create(shop).Error
}

func (r *ShopRepository) FindByID(id uint) (*model.Shop, error) {
	var shop model.Shop
	err := database.DB.First(&shop, id).Error
	return &shop, err
}

func (r *ShopRepository) Update(shop *model.Shop) error {
	return database.DB.Save(shop).Error
}
