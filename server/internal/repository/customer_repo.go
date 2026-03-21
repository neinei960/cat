package repository

import (
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{}
}

func (r *CustomerRepository) Create(customer *model.Customer) error {
	return database.DB.Create(customer).Error
}

func (r *CustomerRepository) FindByID(id uint) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.First(&customer, id).Error
	return &customer, err
}

func (r *CustomerRepository) FindByOpenID(openID string) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.Where("open_id = ?", openID).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) FindByPhone(phone string, shopID uint) (*model.Customer, error) {
	var customer model.Customer
	err := database.DB.Where("phone = ? AND shop_id = ?", phone, shopID).First(&customer).Error
	return &customer, err
}

func (r *CustomerRepository) FindByShopID(shopID uint, page, pageSize int) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64
	db := database.DB.Model(&model.Customer{}).Where("shop_id = ?", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Search(shopID uint, keyword string, page, pageSize int) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64
	db := database.DB.Model(&model.Customer{}).Where("shop_id = ?", shopID)
	if keyword != "" {
		db = db.Where("nickname LIKE ? OR phone LIKE ? OR remark LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Update(customer *model.Customer) error {
	return database.DB.Save(customer).Error
}

func (r *CustomerRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Customer{}, id).Error
}

func (r *CustomerRepository) FindDeleted(shopID uint, page, pageSize int) ([]model.Customer, int64, error) {
	var customers []model.Customer
	var total int64
	db := database.DB.Unscoped().Model(&model.Customer{}).Where("shop_id = ? AND deleted_at IS NOT NULL", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&customers).Error
	return customers, total, err
}

func (r *CustomerRepository) Restore(id uint) error {
	return database.DB.Unscoped().Model(&model.Customer{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

func (r *CustomerRepository) CleanupExpired(before time.Time) (int64, error) {
	result := database.DB.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", before).Delete(&model.Customer{})
	return result.RowsAffected, result.Error
}
