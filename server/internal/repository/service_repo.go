package repository

import (
	"sort"
	"time"

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

func (r *ServiceRepository) FindActiveByIDAndShop(id, shopID uint) (*model.Service, error) {
	var svc model.Service
	err := database.DB.Where("id = ? AND shop_id = ? AND status = 1", id, shopID).First(&svc).Error
	return &svc, err
}

type serviceUsageRow struct {
	ServiceID uint
	Count     int64
}

func (r *ServiceRepository) FindByShopID(shopID uint, page, pageSize int, orderBy string) ([]model.Service, int64, error) {
	var services []model.Service
	var total int64
	db := database.DB.Model(&model.Service{}).Where("shop_id = ?", shopID)
	db.Count(&total)

	if orderBy == "monthly_usage" {
		err := db.Preload("PriceRules").Order("sort_order ASC, id ASC").Find(&services).Error
		if err != nil {
			return nil, total, err
		}

		usageMap, err := r.GetCurrentMonthUsageCounts(shopID)
		if err != nil {
			return nil, total, err
		}

		for i := range services {
			services[i].MonthlyUsageCount = usageMap[services[i].ID]
		}

		sort.SliceStable(services, func(i, j int) bool {
			if services[i].MonthlyUsageCount != services[j].MonthlyUsageCount {
				return services[i].MonthlyUsageCount > services[j].MonthlyUsageCount
			}
			if services[i].SortOrder != services[j].SortOrder {
				return services[i].SortOrder < services[j].SortOrder
			}
			return services[i].ID < services[j].ID
		})

		offset := (page - 1) * pageSize
		if offset >= len(services) {
			return []model.Service{}, total, nil
		}
		end := offset + pageSize
		if end > len(services) {
			end = len(services)
		}
		return services[offset:end], total, nil
	}

	offset := (page - 1) * pageSize
	err := db.Preload("PriceRules").Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&services).Error
	return services, total, err
}

func (r *ServiceRepository) FindActiveByShopID(shopID uint) ([]model.Service, error) {
	var services []model.Service
	err := database.DB.Preload("PriceRules").Where("shop_id = ? AND status = 1", shopID).
		Order("sort_order ASC, id ASC").Find(&services).Error
	return services, err
}

func (r *ServiceRepository) GetCurrentMonthUsageCounts(shopID uint) (map[uint]int64, error) {
	counts := make(map[uint]int64)

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	nextMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0).Format("2006-01-02")

	var rows []serviceUsageRow
	if err := database.DB.Table("appointment_services").
		Select("appointment_services.service_id, COUNT(*) as count").
		Joins("JOIN appointments ON appointments.id = appointment_services.appointment_id").
		Where("appointments.shop_id = ? AND appointments.status <> 4 AND appointments.date >= ? AND appointments.date < ? AND appointments.deleted_at IS NULL AND appointment_services.deleted_at IS NULL",
			shopID, monthStart, nextMonthStart).
		Group("appointment_services.service_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ServiceID] += row.Count
	}

	rows = nil
	if err := database.DB.Table("appointment_pet_services").
		Select("appointment_pet_services.service_id, COUNT(*) as count").
		Joins("JOIN appointment_pets ON appointment_pets.id = appointment_pet_services.appointment_pet_id").
		Joins("JOIN appointments ON appointments.id = appointment_pets.appointment_id").
		Where("appointments.shop_id = ? AND appointments.status <> 4 AND appointments.date >= ? AND appointments.date < ? AND appointments.deleted_at IS NULL AND appointment_pets.deleted_at IS NULL AND appointment_pet_services.deleted_at IS NULL",
			shopID, monthStart, nextMonthStart).
		Group("appointment_pet_services.service_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ServiceID] += row.Count
	}

	return counts, nil
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

// Discounts

func (r *ServiceRepository) CreateDiscount(d *model.ServiceDiscount) error {
	return database.DB.Create(d).Error
}

func (r *ServiceRepository) FindDiscounts(serviceID uint) ([]model.ServiceDiscount, error) {
	var discounts []model.ServiceDiscount
	err := database.DB.Where("service_id = ?", serviceID).Find(&discounts).Error
	return discounts, err
}

func (r *ServiceRepository) DeleteDiscount(id uint) error {
	return database.DB.Delete(&model.ServiceDiscount{}, id).Error
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
