package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type ServiceRepository struct{}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{}
}

func (r *ServiceRepository) Create(svc *model.Service) error {
	return database.DB.Create(svc).Error
}

func (r *ServiceRepository) FindByID(id uint) (*model.Service, error) {
	var svc model.Service
	err := database.DB.Preload("PriceRules").First(&svc, id).Error
	return &svc, err
}

func (r *ServiceRepository) FindByShopID(shopID uint, page, pageSize int) ([]model.Service, int64, error) {
	var services []model.Service
	var total int64
	db := database.DB.Model(&model.Service{}).Where("shop_id = ?", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&services).Error
	return services, total, err
}

func (r *ServiceRepository) FindActiveByShopID(shopID uint) ([]model.Service, error) {
	var services []model.Service
	err := database.DB.Where("shop_id = ? AND status = 1", shopID).
		Order("sort_order ASC, id ASC").Find(&services).Error
	return services, err
}

func (r *ServiceRepository) Update(svc *model.Service) error {
	return database.DB.Save(svc).Error
}

func (r *ServiceRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Service{}, id).Error
}

// Price rules

func (r *ServiceRepository) CreatePriceRule(rule *model.ServicePriceRule) error {
	return database.DB.Create(rule).Error
}

func (r *ServiceRepository) FindPriceRules(serviceID uint) ([]model.ServicePriceRule, error) {
	var rules []model.ServicePriceRule
	err := database.DB.Where("service_id = ?", serviceID).Find(&rules).Error
	return rules, err
}

func (r *ServiceRepository) UpdatePriceRule(rule *model.ServicePriceRule) error {
	return database.DB.Save(rule).Error
}

func (r *ServiceRepository) DeletePriceRule(id uint) error {
	return database.DB.Delete(&model.ServicePriceRule{}, id).Error
}

// Staff services

func (r *ServiceRepository) SetStaffServices(staffID uint, serviceIDs []uint) error {
	tx := database.DB.Begin()
	if err := tx.Where("staff_id = ?", staffID).Delete(&model.StaffService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, sid := range serviceIDs {
		ss := model.StaffService{StaffID: staffID, ServiceID: sid}
		if err := tx.Create(&ss).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *ServiceRepository) FindStaffServices(staffID uint) ([]model.Service, error) {
	var services []model.Service
	err := database.DB.Joins("JOIN staff_services ON staff_services.service_id = services.id AND staff_services.staff_id = ? AND staff_services.deleted_at IS NULL", staffID).
		Find(&services).Error
	return services, err
}
