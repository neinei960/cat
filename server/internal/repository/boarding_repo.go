package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type BoardingRepository struct{}

func NewBoardingRepository() *BoardingRepository {
	return &BoardingRepository{}
}

func (r *BoardingRepository) ListCabinets(shopID uint) ([]model.BoardingCabinet, error) {
	var cabinets []model.BoardingCabinet
	err := database.DB.Where("shop_id = ?", shopID).Order("cabinet_type ASC, id ASC").Find(&cabinets).Error
	return cabinets, err
}

func (r *BoardingRepository) FindCabinetByID(shopID, id uint) (*model.BoardingCabinet, error) {
	var cabinet model.BoardingCabinet
	err := database.DB.Where("shop_id = ?", shopID).First(&cabinet, id).Error
	return &cabinet, err
}

func (r *BoardingRepository) ListHolidays(shopID uint) ([]model.BoardingHoliday, error) {
	var holidays []model.BoardingHoliday
	err := database.DB.Where("shop_id = ?", shopID).Order("holiday_date ASC, id ASC").Find(&holidays).Error
	return holidays, err
}

func (r *BoardingRepository) ListHolidaysInRange(shopID uint, startDate, endDate string) ([]model.BoardingHoliday, error) {
	var holidays []model.BoardingHoliday
	err := database.DB.Where("shop_id = ? AND holiday_date >= ? AND holiday_date < ?", shopID, startDate, endDate).
		Order("holiday_date ASC").Find(&holidays).Error
	return holidays, err
}

func (r *BoardingRepository) ListPolicies(shopID uint) ([]model.BoardingDiscountPolicy, error) {
	var policies []model.BoardingDiscountPolicy
	err := database.DB.Where("shop_id = ?", shopID).Order("policy_type ASC, priority DESC, id DESC").Find(&policies).Error
	return policies, err
}

func (r *BoardingRepository) FindPoliciesByIDs(shopID uint, ids []uint) ([]model.BoardingDiscountPolicy, error) {
	var policies []model.BoardingDiscountPolicy
	if len(ids) == 0 {
		return policies, nil
	}
	err := database.DB.Where("shop_id = ? AND id IN ?", shopID, ids).
		Order("policy_type ASC, priority DESC, id DESC").Find(&policies).Error
	return policies, err
}

func (r *BoardingRepository) FindBoardingOrderByID(shopID, id uint) (*model.BoardingOrder, error) {
	var order model.BoardingOrder
	err := database.DB.Preload("Order.Items").
		Preload("Customer").
		Preload("Staff").
		Preload("Cabinet").
		Preload("Pets.Pet").
		Preload("Logs.Operator").
		Where("shop_id = ?", shopID).
		First(&order, id).Error
	return &order, err
}

func (r *BoardingRepository) ListBoardingOrders(shopID uint, status string, page, pageSize int) ([]model.BoardingOrder, int64, error) {
	var list []model.BoardingOrder
	var total int64
	db := database.DB.Model(&model.BoardingOrder{}).Where("shop_id = ?", shopID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").
		Preload("Cabinet").
		Preload("Pets.Pet").
		Order("id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

func (r *BoardingRepository) ListActiveOrders(shopID uint) ([]model.BoardingOrder, error) {
	var orders []model.BoardingOrder
	err := database.DB.Preload("Customer").
		Preload("Cabinet").
		Preload("Pets.Pet").
		Where("shop_id = ? AND status IN ?", shopID, []string{model.BoardingOrderStatusPendingCheckin, model.BoardingOrderStatusCheckedIn}).
		Order("check_in_at ASC, id ASC").
		Find(&orders).Error
	return orders, err
}
