package service

import (
	"errors"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"gorm.io/gorm"
	"strings"
)

var ErrPetOwnerPhoneConflict = errors.New("该手机号已属于其他客户，未修改主人关联；请先核对客户档案")
var ErrPetOwnerChanged = errors.New("主人关联已变更，请刷新后重试；更换主人请单独处理")

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

func (s *PetService) List(shopID uint, petTag string, page, pageSize int) ([]model.Pet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.FindByShopID(shopID, petTag, page, pageSize)
}

func (s *PetService) FindByCustomer(customerID uint) ([]model.Pet, error) {
	return s.repo.FindByCustomerID(customerID)
}

func (s *PetService) Update(pet *model.Pet) error {
	return s.repo.Update(pet)
}

// Editing contact details preserves the identity referenced by appointments,
// orders and membership records. A phone edit is never an implicit transfer.
func (s *PetService) UpdateProfile(pet *model.Pet, phone string, requestedOwnerID *uint) error {
	phone = strings.TrimSpace(phone)
	return s.repo.Transaction(func(repo *repository.PetRepository) error {
		current, err := repo.LockPet(pet.ShopID, pet.ID)
		if err != nil {
			return err
		}
		ownerID := func(id *uint) uint {
			if id == nil {
				return 0
			}
			return *id
		}
		if ownerID(current.CustomerID) != ownerID(pet.CustomerID) {
			return ErrPetOwnerChanged
		}
		pet.CustomerID = current.CustomerID
		if ownerID(current.CustomerID) != 0 {
			if requestedOwnerID != nil && *requestedOwnerID != *current.CustomerID {
				return ErrPetOwnerChanged
			}
			owner, err := repo.LockOwner(pet.ShopID, *current.CustomerID)
			if err != nil {
				return err
			}
			if phone != "" && phone != owner.Phone {
				other, err := repo.FindOwnerByPhone(pet.ShopID, phone)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				if err == nil && other.ID != owner.ID {
					return ErrPetOwnerPhoneConflict
				}
				if err := repo.UpdateOwnerPhone(owner, phone); err != nil {
					return err
				}
			}
		} else {
			var owner *model.Customer
			if requestedOwnerID != nil && *requestedOwnerID > 0 {
				owner, err = repo.LockOwner(pet.ShopID, *requestedOwnerID)
				if err != nil {
					return err
				}
				if phone != "" && phone != owner.Phone {
					return ErrPetOwnerPhoneConflict
				}
			} else if phone != "" {
				owner, err = repo.FindOwnerByPhone(pet.ShopID, phone)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					owner = &model.Customer{ShopID: pet.ShopID, Phone: phone}
					err = repo.CreateOwner(owner)
				}
				if err != nil {
					return err
				}
			}
			if owner != nil {
				pet.CustomerID = &owner.ID
			}
		}
		return repo.Update(pet)
	})
}

func (s *PetService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *PetService) Search(shopID uint, keyword, petTag string, page, pageSize int) ([]model.Pet, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.Search(shopID, keyword, petTag, page, pageSize)
}
