package service

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type PetService struct {
	repo *repository.PetRepository
}

func NewPetService(repo *repository.PetRepository) *PetService {
	return &PetService{repo: repo}
}

func (s *PetService) Create(pet *model.Pet) error {
	return s.repo.Create(pet)
}

func (s *PetService) GetByID(id uint) (*model.Pet, error) {
	return s.repo.FindByID(id)
}

func (s *PetService) List(shopID uint, page, pageSize int) ([]model.Pet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShopID(shopID, page, pageSize)
}

func (s *PetService) FindByCustomer(customerID uint) ([]model.Pet, error) {
	return s.repo.FindByCustomerID(customerID)
}

func (s *PetService) Update(pet *model.Pet) error {
	return s.repo.Update(pet)
}

func (s *PetService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *PetService) Search(shopID uint, keyword string, page, pageSize int) ([]model.Pet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.Search(shopID, keyword, page, pageSize)
}
