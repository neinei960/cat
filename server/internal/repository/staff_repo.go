package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type StaffRepository struct{}

func NewStaffRepository() *StaffRepository {
	return &StaffRepository{}
}

func (r *StaffRepository) Create(staff *model.Staff) error {
	return database.DB.Create(staff).Error
}

func (r *StaffRepository) FindByID(id uint) (*model.Staff, error) {
	var staff model.Staff
	err := database.DB.First(&staff, id).Error
	return &staff, err
}

func (r *StaffRepository) FindByPhone(phone string) (*model.Staff, error) {
	var staff model.Staff
	err := database.DB.Where("phone = ?", phone).First(&staff).Error
	return &staff, err
}

func (r *StaffRepository) FindByShopID(shopID uint, page, pageSize int) ([]model.Staff, int64, error) {
	var staffs []model.Staff
	var total int64
	db := database.DB.Model(&model.Staff{}).Where("shop_id = ?", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Find(&staffs).Error
	return staffs, total, err
}

func (r *StaffRepository) Update(staff *model.Staff) error {
	return database.DB.Save(staff).Error
}

func (r *StaffRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Staff{}, id).Error
}
