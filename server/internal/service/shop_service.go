package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type ShopService struct {
	repo *repository.ShopRepository
}

func NewShopService(repo *repository.ShopRepository) *ShopService {
	return &ShopService{repo: repo}
}

func (s *ShopService) GetByID(id uint) (*model.Shop, error) {
	return s.repo.FindByID(id)
}

func (s *ShopService) Update(shop *model.Shop) error {
	return s.repo.Update(shop)
}
