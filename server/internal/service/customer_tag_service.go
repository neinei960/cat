package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type CustomerTagService struct {
	repo *repository.CustomerTagRepository
}

func NewCustomerTagService(repo *repository.CustomerTagRepository) *CustomerTagService {
	return &CustomerTagService{repo: repo}
}

func (s *CustomerTagService) Create(tag *model.CustomerTag) error {
	return s.repo.Create(tag)
}

func (s *CustomerTagService) List(shopID uint) ([]model.CustomerTag, error) {
	return s.repo.List(shopID)
}

func (s *CustomerTagService) GetByID(id uint, shopID uint) (*model.CustomerTag, error) {
	return s.repo.FindByID(id, shopID)
}

func (s *CustomerTagService) Update(tag *model.CustomerTag) error {
	return s.repo.Update(tag)
}

func (s *CustomerTagService) Delete(id uint, shopID uint) error {
	return s.repo.Delete(id, shopID)
}
