package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type CustomerService struct {
	repo *repository.CustomerRepository
}

func NewCustomerService(repo *repository.CustomerRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) Create(customer *model.Customer) error {
	return s.repo.Create(customer)
}

func (s *CustomerService) GetByID(id uint) (*model.Customer, error) {
	return s.repo.FindByID(id)
}

func (s *CustomerService) List(shopID uint, page, pageSize int) ([]model.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShopID(shopID, page, pageSize)
}

func (s *CustomerService) Update(customer *model.Customer) error {
	return s.repo.Update(customer)
}

func (s *CustomerService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *CustomerService) ListDeleted(shopID uint, page, pageSize int) ([]model.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindDeleted(shopID, page, pageSize)
}

func (s *CustomerService) Restore(id uint) error {
	return s.repo.Restore(id)
}

func (s *CustomerService) GetByPhone(phone string, shopID uint) (*model.Customer, error) {
	return s.repo.FindByPhone(phone, shopID)
}

func (s *CustomerService) Search(shopID uint, keyword string, page, pageSize int) ([]model.Customer, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.Search(shopID, keyword, page, pageSize)
}
