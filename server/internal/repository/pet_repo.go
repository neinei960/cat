package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type PetRepository struct{}

const petSpentJoinSQL = `
LEFT JOIN (
	SELECT pet_id, COALESCE(SUM(pay_amount), 0) AS total_spent
	FROM orders
	WHERE deleted_at IS NULL
	  AND pet_id IS NOT NULL
	  AND pay_status = 1
	  AND status = 1
	GROUP BY pet_id
) pet_spend ON pet_spend.pet_id = pets.id
`

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

func (r *PetRepository) FindByShopID(shopID uint, petTag string, page, pageSize int) ([]model.Pet, int64, error) {
	return r.list(shopID, "", petTag, page, pageSize)
}

func (r *PetRepository) Search(shopID uint, keyword, petTag string, page, pageSize int) ([]model.Pet, int64, error) {
	return r.list(shopID, keyword, petTag, page, pageSize)
}

func (r *PetRepository) list(shopID uint, keyword, petTag string, page, pageSize int) ([]model.Pet, int64, error) {
	var pets []model.Pet
	var total int64
	db := database.DB.Model(&model.Pet{}).
		Joins("LEFT JOIN customers ON customers.id = pets.customer_id").
		Joins(petSpentJoinSQL).
		Where("pets.shop_id = ?", shopID)

	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where(`(
			pets.name LIKE ?
			OR customers.nickname LIKE ?
			OR customers.phone LIKE ?
			OR pets.fur_level LIKE ?
			OR pets.personality LIKE ?
			OR pets.aggression LIKE ?
			OR pets.forbidden_zones LIKE ?
			OR pets.care_notes LIKE ?
			OR pets.behavior_notes LIKE ?
			OR (? = '已绝育' AND pets.neutered = ?)
		)`, like, like, like, like, like, like, like, like, like, keyword, keyword == "已绝育")
	}

	if petTag != "" {
		db = applyPetTagFilter(db, petTag)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Customer.MemberCard").
		Order("COALESCE(pet_spend.total_spent, 0) DESC, CASE WHEN pets.customer_id IS NULL THEN 1 ELSE 0 END ASC, pets.customer_id ASC, pets.id ASC").
		Offset(offset).Limit(pageSize).Find(&pets).Error
	return pets, total, err
}

func applyPetTagFilter(db *gorm.DB, petTag string) *gorm.DB {
	switch petTag {
	case "已绝育":
		return db.Where("pets.neutered = ?", true)
	case "未绝育":
		return db.Where("pets.neutered = ?", false)
	default:
		return db.Where("(pets.fur_level = ? OR pets.personality = ? OR pets.aggression = ?)", petTag, petTag, petTag)
	}
}
