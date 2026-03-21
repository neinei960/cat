package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type PetRepository struct{}

func NewPetRepository() *PetRepository {
	return &PetRepository{}
}

func (r *PetRepository) Create(pet *model.Pet) error {
	return database.DB.Create(pet).Error
}

func (r *PetRepository) FindByID(id uint) (*model.Pet, error) {
	var pet model.Pet
	err := database.DB.Preload("Customer").First(&pet, id).Error
	return &pet, err
}

func (r *PetRepository) FindByShopID(shopID uint, page, pageSize int) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64
	db := database.DB.Model(&model.Pet{}).Where("shop_id = ?", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Customer.MemberCard").Order("COALESCE(customer_id, 999999999) ASC, customer_id ASC, id ASC").Offset(offset).Limit(pageSize).Find(&pets).Error
	return pets, total, err
}

func (r *PetRepository) FindByCustomerID(customerID uint) ([]model.Pet, error) {
	var pets []model.Pet
	err := database.DB.Where("customer_id = ?", customerID).Find(&pets).Error
	return pets, err
}

func (r *PetRepository) Update(pet *model.Pet) error {
	// Clear preloaded associations to avoid GORM ignoring FK changes
	pet.Customer = nil
	return database.DB.Select("*").Save(pet).Error
}

func (r *PetRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Pet{}, id).Error
}

func (r *PetRepository) Search(shopID uint, keyword string, page, pageSize int) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64
	like := "%" + keyword + "%"
	db := database.DB.Model(&model.Pet{}).
		Joins("LEFT JOIN customers ON customers.id = pets.customer_id").
		Where("pets.shop_id = ? AND (pets.name LIKE ? OR customers.nickname LIKE ? OR customers.phone LIKE ?)", shopID, like, like, like)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Customer.MemberCard").Order("COALESCE(pets.customer_id, 999999999) ASC, pets.customer_id ASC, pets.id ASC").Offset(offset).Limit(pageSize).Find(&pets).Error
	return pets, total, err
}
