package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type FeedingPlanFilter struct {
	Status string
	Page   int
	Size   int
}

type FeedingVisitFilter struct {
	ID            uint
	PlanID        uint
	ScheduledDate string
	Status        string
	StaffID       uint
	WindowCode    string
}

type FeedingRepository struct{}

func NewFeedingRepository() *FeedingRepository {
	return &FeedingRepository{}
}

func (r *FeedingRepository) GetSetting(shopID uint) (*model.FeedingSetting, error) {
	var setting model.FeedingSetting
	err := database.DB.Where("shop_id = ?", shopID).First(&setting).Error
	return &setting, err
}

func (r *FeedingRepository) SaveSetting(setting *model.FeedingSetting) error {
	return database.DB.Save(setting).Error
}

func (r *FeedingRepository) FindPlanByID(shopID, id uint) (*model.FeedingPlan, error) {
	var plan model.FeedingPlan
	err := database.DB.Preload("Customer").
		Preload("Order.Items").
		Preload("Pets.Pet").
		Preload("Rules").
		Preload("Visits.Staff").
		Preload("Visits.Items").
		Preload("Visits.Logs.Operator").
		Preload("Visits.Media").
		Where("shop_id = ?", shopID).
		First(&plan, id).Error
	return &plan, err
}

func (r *FeedingRepository) ListPlans(shopID uint, filter FeedingPlanFilter) ([]model.FeedingPlan, int64, error) {
	var list []model.FeedingPlan
	var total int64

	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 {
		size = 20
	}

	db := database.DB.Model(&model.FeedingPlan{}).Where("shop_id = ?", shopID)
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Customer").
		Preload("Pets.Pet").
		Preload("Visits").
		Order("id DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&list).Error
	return list, total, err
}

func (r *FeedingRepository) ListVisits(shopID uint, filter FeedingVisitFilter, preloadPlan bool) ([]model.FeedingVisit, error) {
	var list []model.FeedingVisit
	db := database.DB.Model(&model.FeedingVisit{}).Where("shop_id = ?", shopID)
	if filter.ID > 0 {
		db = db.Where("id = ?", filter.ID)
	}
	if filter.PlanID > 0 {
		db = db.Where("feeding_plan_id = ?", filter.PlanID)
	}
	if filter.ScheduledDate != "" {
		db = db.Where("scheduled_date = ?", filter.ScheduledDate)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.StaffID > 0 {
		db = db.Where("staff_id = ?", filter.StaffID)
	}
	if filter.WindowCode != "" {
		db = db.Where("window_code = ?", filter.WindowCode)
	}
	if preloadPlan {
		db = db.Preload("Plan.Customer").Preload("Plan.Pets.Pet").Preload("Plan.Order")
	}
	err := db.Preload("Staff").
		Preload("Items").
		Preload("Logs.Operator").
		Preload("Media").
		Order("scheduled_date ASC, id ASC").
		Find(&list).Error
	return list, err
}

func (r *FeedingRepository) FindVisitByID(shopID, id uint) (*model.FeedingVisit, error) {
	var visit model.FeedingVisit
	err := database.DB.Preload("Plan.Customer").
		Preload("Plan.Pets.Pet").
		Preload("Plan.Order").
		Preload("Staff").
		Preload("Items").
		Preload("Logs.Operator").
		Preload("Media").
		Where("shop_id = ?", shopID).
		First(&visit, id).Error
	return &visit, err
}

func (r *FeedingRepository) UpsertPlan(tx *gorm.DB, plan *model.FeedingPlan) error {
	return tx.Save(plan).Error
}
