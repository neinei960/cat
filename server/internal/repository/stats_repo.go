package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type StatsRepository struct{}

func NewStatsRepository() *StatsRepository {
	return &StatsRepository{}
}

func (r *StatsRepository) Upsert(stats *model.DailyStats) error {
	return database.DB.Where("shop_id = ? AND date = ?", stats.ShopID, stats.Date).
		Assign(*stats).FirstOrCreate(stats).Error
}

func (r *StatsRepository) FindByDateRange(shopID uint, startDate, endDate string) ([]model.DailyStats, error) {
	var stats []model.DailyStats
	err := database.DB.Where("shop_id = ? AND date >= ? AND date <= ?", shopID, startDate, endDate).
		Order("date ASC").Find(&stats).Error
	return stats, err
}

// RevenueTrendItem is a single day's revenue data
type RevenueTrendItem struct {
	Date       string  `json:"date"`
	Revenue    float64 `json:"revenue"`
	OrderCount int     `json:"order_count"`
}

// GetRevenueTrendRealtime queries orders table directly for daily revenue (no dependency on daily_stats)
func (r *StatsRepository) GetRevenueTrendRealtime(shopID uint, startDate, endDate string) ([]RevenueTrendItem, error) {
	var items []RevenueTrendItem
	err := database.DB.Model(&model.Order{}).
		Select("DATE(created_at) as date, COALESCE(SUM(pay_amount), 0) as revenue, COUNT(*) as order_count").
		Where("shop_id = ? AND status = 1 AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Find(&items).Error
	return items, err
}

type OverviewStats struct {
	TodayRevenue          float64 `json:"today_revenue"`
	TodayOrderCount       int     `json:"today_order_count"`
	TodayAppointmentCount int     `json:"today_appointment_count"`
	TodayNewCustomers     int     `json:"today_new_customers"`
	PendingAppointments   int64   `json:"pending_appointments"`
	TotalCustomers        int64   `json:"total_customers"`
}

type MemberTemplateStat struct {
	TemplateID   uint   `json:"template_id"`
	TemplateName string `json:"template_name"`
	Count        int    `json:"count"`
}

type MemberStats struct {
	ActiveMembers      int64                `json:"active_members"`
	FrozenMembers      int64                `json:"frozen_members"`
	TotalBalance       float64              `json:"total_balance"`
	TotalMemberSpent   float64              `json:"total_member_spent"`
	RangeRecharge      float64              `json:"range_recharge"`
	RangeConsumption   float64              `json:"range_consumption"`
	TemplateBreakdown  []MemberTemplateStat `json:"template_breakdown"`
}

func (r *StatsRepository) GetOverview(shopID uint, today string) (*OverviewStats, error) {
	var stats OverviewStats

	// Today's revenue and order count - query orders table directly for real-time data
	var revenueResult struct {
		Total float64
		Count int64
	}
	database.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount), 0) as total, COUNT(*) as count").
		Where("shop_id = ? AND status = 1 AND DATE(created_at) = ?", shopID, today).
		Scan(&revenueResult)
	stats.TodayRevenue = revenueResult.Total
	stats.TodayOrderCount = int(revenueResult.Count)

	// Today's appointment count
	var apptCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ?", shopID, today).Count(&apptCount)
	stats.TodayAppointmentCount = int(apptCount)

	// Today's new customers
	var newCustCount int64
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ? AND DATE(created_at) = ?", shopID, today).Count(&newCustCount)
	stats.TodayNewCustomers = int(newCustCount)

	// Pending appointments count
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status IN (0,1,6)", shopID).Count(&stats.PendingAppointments)

	// Total customers
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ?", shopID).Count(&stats.TotalCustomers)

	return &stats, nil
}

// GetOverviewByRange returns aggregated stats for a date range
func (r *StatsRepository) GetOverviewByRange(shopID uint, startDate, endDate string) (*OverviewStats, error) {
	var stats OverviewStats

	var revenueResult struct {
		Total float64
		Count int64
	}
	database.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(pay_amount), 0) as total, COUNT(*) as count").
		Where("shop_id = ? AND status = 1 AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).
		Scan(&revenueResult)
	stats.TodayRevenue = revenueResult.Total
	stats.TodayOrderCount = int(revenueResult.Count)

	var apptCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ?", shopID, startDate, endDate).Count(&apptCount)
	stats.TodayAppointmentCount = int(apptCount)

	var newCustCount int64
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).Count(&newCustCount)
	stats.TodayNewCustomers = int(newCustCount)

	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status IN (0,1,6)", shopID).Count(&stats.PendingAppointments)
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ?", shopID).Count(&stats.TotalCustomers)

	return &stats, nil
}

func (r *StatsRepository) GetMemberStats(shopID uint, startDate, endDate string) (*MemberStats, error) {
	var stats MemberStats

	database.DB.Model(&model.MemberCard{}).
		Where("shop_id = ? AND status = 1", shopID).
		Count(&stats.ActiveMembers)

	database.DB.Model(&model.MemberCard{}).
		Where("shop_id = ? AND status = 0", shopID).
		Count(&stats.FrozenMembers)

	database.DB.Model(&model.MemberCard{}).
		Select("COALESCE(SUM(balance), 0)").
		Where("shop_id = ? AND status = 1", shopID).
		Scan(&stats.TotalBalance)

	database.DB.Model(&model.MemberCard{}).
		Select("COALESCE(SUM(total_spent), 0)").
		Where("shop_id = ? AND status = 1", shopID).
		Scan(&stats.TotalMemberSpent)

	database.DB.Model(&model.RechargeRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("shop_id = ? AND type = 1 AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).
		Scan(&stats.RangeRecharge)

	database.DB.Model(&model.RechargeRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("shop_id = ? AND type = 2 AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).
		Scan(&stats.RangeConsumption)

	database.DB.Table("member_cards").
		Select("member_cards.template_id, member_card_templates.name as template_name, COUNT(*) as count").
		Joins("JOIN member_card_templates ON member_card_templates.id = member_cards.template_id").
		Where("member_cards.shop_id = ? AND member_cards.status = 1 AND member_cards.deleted_at IS NULL AND member_card_templates.deleted_at IS NULL", shopID).
		Group("member_cards.template_id, member_card_templates.name").
		Order("count DESC, member_cards.template_id ASC").
		Find(&stats.TemplateBreakdown)

	return &stats, nil
}

type ServiceRanking struct {
	ServiceName string  `json:"service_name"`
	Count       int     `json:"count"`
	Revenue     float64 `json:"revenue"`
}

func (r *StatsRepository) GetServiceRanking(shopID uint, startDate, endDate string) ([]ServiceRanking, error) {
	var rankings []ServiceRanking
	err := database.DB.Table("appointment_services").
		Select("appointment_services.service_name, COUNT(*) as count, SUM(appointment_services.price) as revenue").
		Joins("JOIN appointments ON appointments.id = appointment_services.appointment_id").
		Where("appointments.shop_id = ? AND appointments.date >= ? AND appointments.date <= ? AND appointments.status = 3 AND appointments.deleted_at IS NULL AND appointment_services.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group("appointment_services.service_name").
		Order("count DESC").
		Limit(10).
		Find(&rankings).Error
	return rankings, err
}

type CategoryStat struct {
	ServiceName string  `json:"service_name"`
	FurLevel    string  `json:"fur_level"`
	Count       int     `json:"count"`
	Revenue     float64 `json:"revenue"`
}

func (r *StatsRepository) GetCategoryStats(shopID uint, startDate, endDate string) ([]CategoryStat, error) {
	var stats []CategoryStat
	err := database.DB.Table("order_items").
		Select("order_items.name as service_name, pets.fur_level, COUNT(*) as count, SUM(order_items.amount) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Where("orders.shop_id = ? AND DATE(orders.created_at) >= ? AND DATE(orders.created_at) <= ? AND orders.status = 1 AND order_items.item_type = 1 AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group("order_items.name, pets.fur_level").
		Order("order_items.name ASC, pets.fur_level ASC").
		Find(&stats).Error
	return stats, err
}

type StaffPerformance struct {
	StaffID        uint    `json:"staff_id"`
	StaffName      string  `json:"staff_name"`
	ApptCount      int     `json:"appointment_count"`
	Revenue        float64 `json:"revenue"`
	CommissionRate float64 `json:"commission_rate"`
	Commission     float64 `json:"commission"`
}

func (r *StatsRepository) GetStaffPerformance(shopID uint, startDate, endDate string) ([]StaffPerformance, error) {
	var perfs []StaffPerformance
	err := database.DB.Table("appointments").
		Select("appointments.staff_id, staffs.name as staff_name, staffs.commission_rate, COUNT(*) as appt_count, SUM(appointments.total_amount) as revenue").
		Joins("JOIN staffs ON staffs.id = appointments.staff_id").
		Where("appointments.shop_id = ? AND appointments.date >= ? AND appointments.date <= ? AND appointments.status = 3 AND appointments.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group("appointments.staff_id, staffs.name, staffs.commission_rate").
		Order("revenue DESC").
		Find(&perfs).Error
	// Calculate commission
	for i := range perfs {
		perfs[i].Commission = perfs[i].Revenue * perfs[i].CommissionRate / 100
	}
	return perfs, err
}
