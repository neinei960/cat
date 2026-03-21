package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type ServiceService struct {
	repo *repository.ServiceRepository
}

func NewServiceService(repo *repository.ServiceRepository) *ServiceService {
	return &ServiceService{repo: repo}
}

func (s *ServiceService) Create(svc *model.Service) error {
	return s.repo.Create(svc)
}

func (s *ServiceService) GetByID(id uint) (*model.Service, error) {
	return s.repo.FindByID(id)
}

func (s *ServiceService) List(shopID uint, page, pageSize int) ([]model.Service, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShopID(shopID, page, pageSize)
}

func (s *ServiceService) ListActive(shopID uint) ([]model.Service, error) {
	return s.repo.FindActiveByShopID(shopID)
}

func (s *ServiceService) Update(svc *model.Service) error {
	return s.repo.Update(svc)
}

func (s *ServiceService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// Price rules

func (s *ServiceService) CreatePriceRule(rule *model.ServicePriceRule) error {
	return s.repo.CreatePriceRule(rule)
}

func (s *ServiceService) GetPriceRules(serviceID uint) ([]model.ServicePriceRule, error) {
	return s.repo.FindPriceRules(serviceID)
}

func (s *ServiceService) UpdatePriceRule(rule *model.ServicePriceRule) error {
	return s.repo.UpdatePriceRule(rule)
}

func (s *ServiceService) DeletePriceRule(id uint) error {
	return s.repo.DeletePriceRule(id)
}
