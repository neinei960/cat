package service

import (
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

type DashboardService struct {
	statsRepo *repository.StatsRepository
}

func NewDashboardService(statsRepo *repository.StatsRepository) *DashboardService {
	return &DashboardService{statsRepo: statsRepo}
}

func (s *DashboardService) GetOverview(shopID uint) (*repository.OverviewStats, error) {
	today := time.Now().Format("2006-01-02")
	return s.statsRepo.GetOverview(shopID, today)
}

func (s *DashboardService) GetOverviewByRange(shopID uint, startDate, endDate string) (*repository.OverviewStats, error) {
	return s.statsRepo.GetOverviewByRange(shopID, startDate, endDate)
}

func (s *DashboardService) GetRevenueChart(shopID uint, startDate, endDate string) ([]repository.RevenueTrendItem, error) {
	return s.statsRepo.GetRevenueTrendRealtime(shopID, startDate, endDate)
}

func (s *DashboardService) GetServiceRanking(shopID uint, startDate, endDate string) ([]repository.ServiceRanking, error) {
	return s.statsRepo.GetServiceRanking(shopID, startDate, endDate)
}

func (s *DashboardService) GetStaffPerformance(shopID uint, startDate, endDate string) ([]repository.StaffPerformance, error) {
	return s.statsRepo.GetStaffPerformance(shopID, startDate, endDate)
}

func (s *DashboardService) GetCategoryStats(shopID uint, startDate, endDate string) ([]repository.CategoryStat, error) {
	return s.statsRepo.GetCategoryStats(shopID, startDate, endDate)
}

func (s *DashboardService) GetMemberStats(shopID uint, startDate, endDate string) (*repository.MemberStats, error) {
	return s.statsRepo.GetMemberStats(shopID, startDate, endDate)
}

// AggregateDaily recalculates daily stats for a given date
func (s *DashboardService) AggregateDaily(shopID uint, date string) error {
	var revenue float64
	var orderCount, apptCount, newCustomerCount, cancelCount int64

	database.DB.Model(&model.Order{}).
		Where("shop_id = ? AND DATE(created_at) = ? AND status = 1", shopID, date).
		Select("COALESCE(SUM(pay_amount), 0)").Row().Scan(&revenue)

	database.DB.Model(&model.Order{}).
		Where("shop_id = ? AND DATE(created_at) = ? AND status = 1", shopID, date).
		Count(&orderCount)

	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ?", shopID, date).
		Count(&apptCount)

	database.DB.Model(&model.Customer{}).
		Where("shop_id = ? AND DATE(created_at) = ?", shopID, date).
		Count(&newCustomerCount)

	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ? AND status = 4", shopID, date).
		Count(&cancelCount)

	stats := &model.DailyStats{
		ShopID:           shopID,
		Date:             date,
		Revenue:          revenue,
		OrderCount:       int(orderCount),
		AppointmentCount: int(apptCount),
		NewCustomerCount: int(newCustomerCount),
		CancelCount:      int(cancelCount),
	}

	return s.statsRepo.Upsert(stats)
}
