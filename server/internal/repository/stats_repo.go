package repository

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
// Uses appointment date when order is from an appointment, falls back to order created_at for walk-ins
func (r *StatsRepository) GetRevenueTrendRealtime(shopID uint, startDate, endDate string) ([]RevenueTrendItem, error) {
	var items []RevenueTrendItem
	err := database.DB.Model(&model.Order{}).
		Select("COALESCE(appointments.date, DATE(orders.created_at)) as date, COALESCE(SUM(orders.pay_amount), 0) as revenue, COUNT(*) as order_count").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Group("COALESCE(appointments.date, DATE(orders.created_at))").
		Order("date ASC").
		Find(&items).Error
	return items, err
}

type OverviewStats struct {
	TodayRevenue            float64                `json:"today_revenue"`
	MonthRevenue            float64                `json:"month_revenue"`
	MonthRecharge           float64                `json:"month_recharge"`
	MonthCollection         float64                `json:"month_collection"`
	TodayOrderCount         int                    `json:"today_order_count"`
	TodayAppointmentCount   int                    `json:"today_appointment_count"`
	TodayServiceCompleted   int                    `json:"today_service_completed_count"`
	TodayPendingSettlement  int                    `json:"today_pending_settlement_count"`
	TodayRefundedOrderCount int                    `json:"today_refunded_order_count"`
	TodayNewCustomers       int                    `json:"today_new_customers"`
	RegularCustomerCount    int                    `json:"regular_customer_count"`
	PendingAppointments     int64                  `json:"pending_appointments"`
	TotalCustomers          int64                  `json:"total_customers"`
	AvgOrderValue           float64                `json:"avg_order_value"`
	NoShowRate              float64                `json:"no_show_rate"`
	NoShowCount             int64                  `json:"no_show_count"`
	TotalAppointments       int64                  `json:"total_appointments"`
	PaymentBreakdown        []PaymentBreakdownItem `json:"payment_breakdown"`
}

type PaymentBreakdownItem struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
}

type MemberTemplateStat struct {
	TemplateID   uint   `json:"template_id"`
	TemplateName string `json:"template_name"`
	Count        int    `json:"count"`
}

type MemberStats struct {
	ActiveMembers     int64                `json:"active_members"`
	FrozenMembers     int64                `json:"frozen_members"`
	TotalBalance      float64              `json:"total_balance"`
	TotalMemberSpent  float64              `json:"total_member_spent"`
	RangeRecharge     float64              `json:"range_recharge"`
	RangeConsumption  float64              `json:"range_consumption"`
	TemplateBreakdown []MemberTemplateStat `json:"template_breakdown"`
}

func (r *StatsRepository) GetOverview(shopID uint, today string) (*OverviewStats, error) {
	var stats OverviewStats

	// Today's revenue and order count - query orders table directly for real-time data
	var revenueResult struct {
		Total float64
		Count int64
	}
	database.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(orders.pay_amount), 0) as total, COUNT(*) as count").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) = ?", shopID, today).
		Scan(&revenueResult)
	stats.TodayRevenue = revenueResult.Total
	stats.TodayOrderCount = int(revenueResult.Count)
	r.fillMonthCollectionStats(shopID, &stats)

	// Today's appointment count
	var apptCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ?", shopID, today).Count(&apptCount)
	stats.TodayAppointmentCount = int(apptCount)

	var completedCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ? AND status = 3", shopID, today).
		Count(&completedCount)
	stats.TodayServiceCompleted = int(completedCount)

	var pendingSettlementCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ? AND status = 3 AND paid_amount + 0.0001 < total_amount", shopID, today).
		Count(&pendingSettlementCount)
	stats.TodayPendingSettlement = int(pendingSettlementCount)

	var refundedCount int64
	database.DB.Model(&model.Order{}).
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 3 AND COALESCE(appointments.date, DATE(orders.created_at)) = ?", shopID, today).
		Count(&refundedCount)
	stats.TodayRefundedOrderCount = int(refundedCount)

	// Today's new customers
	var newCustCount int64
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ? AND DATE(created_at) = ?", shopID, today).Count(&newCustCount)
	stats.TodayNewCustomers = int(newCustCount)

	var regularCustomerCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date = ? AND customer_type = ?", shopID, today, model.AppointmentCustomerTypeRegular).
		Count(&regularCustomerCount)
	stats.RegularCustomerCount = int(regularCustomerCount)

	// Pending appointments count
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status IN (0,1,6)", shopID).Count(&stats.PendingAppointments)

	stats.TotalCustomers = r.countServedCustomers(shopID, today, today)

	// 客单价 (AOV) - 近30天已完成订单
	var aovResult struct{ Avg float64 }
	database.DB.Model(&model.Order{}).
		Select("COALESCE(AVG(pay_amount), 0) as avg").
		Where("shop_id = ? AND status = 1 AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)", shopID).
		Scan(&aovResult)
	stats.AvgOrderValue = aovResult.Avg

	// 爽约率 - 近30天
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)", shopID).
		Count(&stats.TotalAppointments)
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status = 5 AND date >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)", shopID).
		Count(&stats.NoShowCount)
	if stats.TotalAppointments > 0 {
		stats.NoShowRate = float64(stats.NoShowCount) / float64(stats.TotalAppointments)
	}

	stats.PaymentBreakdown = r.getPaymentBreakdown(shopID, today, today)

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
		Select("COALESCE(SUM(orders.pay_amount), 0) as total, COUNT(*) as count").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Scan(&revenueResult)
	stats.TodayRevenue = revenueResult.Total
	stats.TodayOrderCount = int(revenueResult.Count)
	r.fillCollectionStatsByRange(shopID, startDate, endDate, &stats)

	var apptCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ?", shopID, startDate, endDate).Count(&apptCount)
	stats.TodayAppointmentCount = int(apptCount)

	var completedCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ? AND status = 3", shopID, startDate, endDate).
		Count(&completedCount)
	stats.TodayServiceCompleted = int(completedCount)

	var pendingSettlementCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ? AND status = 3 AND paid_amount + 0.0001 < total_amount", shopID, startDate, endDate).
		Count(&pendingSettlementCount)
	stats.TodayPendingSettlement = int(pendingSettlementCount)

	var refundedCount int64
	database.DB.Model(&model.Order{}).
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 3 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Count(&refundedCount)
	stats.TodayRefundedOrderCount = int(refundedCount)

	var newCustCount int64
	database.DB.Model(&model.Customer{}).
		Where("shop_id = ? AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).Count(&newCustCount)
	stats.TodayNewCustomers = int(newCustCount)

	var regularCustomerCount int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ? AND customer_type = ?", shopID, startDate, endDate, model.AppointmentCustomerTypeRegular).
		Count(&regularCustomerCount)
	stats.RegularCustomerCount = int(regularCustomerCount)

	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status IN (0,1,6)", shopID).Count(&stats.PendingAppointments)
	stats.TotalCustomers = r.countServedCustomers(shopID, startDate, endDate)

	// AOV for range
	var aovResult struct{ Avg float64 }
	database.DB.Model(&model.Order{}).
		Select("COALESCE(AVG(orders.pay_amount), 0) as avg").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Scan(&aovResult)
	stats.AvgOrderValue = aovResult.Avg

	// No-show for range
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ?", shopID, startDate, endDate).
		Count(&stats.TotalAppointments)
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND status = 5 AND date >= ? AND date <= ?", shopID, startDate, endDate).
		Count(&stats.NoShowCount)
	if stats.TotalAppointments > 0 {
		stats.NoShowRate = float64(stats.NoShowCount) / float64(stats.TotalAppointments)
	}

	stats.PaymentBreakdown = r.getPaymentBreakdown(shopID, startDate, endDate)

	return &stats, nil
}

func (r *StatsRepository) countServedCustomers(shopID uint, startDate, endDate string) int64 {
	var count int64
	database.DB.Model(&model.Appointment{}).
		Where("shop_id = ? AND date >= ? AND date <= ? AND status NOT IN ?", shopID, startDate, endDate, []int{4, 5}).
		Distinct("customer_id").
		Count(&count)
	return count
}

func (r *StatsRepository) fillMonthCollectionStats(shopID uint, stats *OverviewStats) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	monthEnd := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	r.fillCollectionStatsByRange(shopID, monthStart, monthEnd, stats)
}

func (r *StatsRepository) fillCollectionStatsByRange(shopID uint, startDate, endDate string, stats *OverviewStats) {
	database.DB.Model(&model.Order{}).
		Select("COALESCE(SUM(orders.pay_amount), 0)").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Scan(&stats.MonthRevenue)

	database.DB.Model(&model.RechargeRecord{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("shop_id = ? AND type = 1 AND DATE(created_at) >= ? AND DATE(created_at) <= ?", shopID, startDate, endDate).
		Scan(&stats.MonthRecharge)

	stats.MonthCollection = stats.MonthRevenue + stats.MonthRecharge
}

func (r *StatsRepository) getPaymentBreakdown(shopID uint, startDate, endDate string) []PaymentBreakdownItem {
	type paymentBreakdownRow struct {
		PayGroup string  `gorm:"column:pay_group" json:"pay_group"`
		Amount   float64 `json:"amount"`
	}

	var rows []paymentBreakdownRow
	payMethodGroupExpr := `
		CASE
			WHEN orders.pay_method = 'wechat' THEN 'wechat'
				WHEN orders.pay_method IN ('qrcode', 'alipay') THEN 'qrcode'
				WHEN orders.pay_method = 'meituan' THEN 'meituan'
				WHEN orders.pay_method IN ('balance', 'card') THEN 'balance'
				WHEN orders.pay_method = 'mixed_balance' THEN 'mixed_balance'
				ELSE 'other'
			END
		`
	if err := database.DB.Model(&model.Order{}).
		Select(payMethodGroupExpr+" as pay_group, COALESCE(SUM(orders.pay_amount), 0) as amount").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND orders.status = 1 AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ?", shopID, startDate, endDate).
		Group(payMethodGroupExpr).
		Order("amount DESC").
		Find(&rows).Error; err != nil {
		return nil
	}

	items := make([]PaymentBreakdownItem, 0, len(rows))
	for _, row := range rows {
		if row.Amount <= 0 {
			continue
		}
		items = append(items, PaymentBreakdownItem{
			Key:    row.PayGroup,
			Label:  paymentBreakdownLabel(row.PayGroup),
			Amount: row.Amount,
		})
	}
	return items
}

func paymentBreakdownLabel(key string) string {
	switch key {
	case "wechat":
		return "微信"
	case "qrcode":
		return "扫码"
	case "meituan":
		return "美团"
	case "balance":
		return "会员"
	case "mixed_balance":
		return "会员+补差"
	default:
		return "其他"
	}
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

type ProjectRevenueNode struct {
	Key      string               `json:"key"`
	Name     string               `json:"name"`
	Kind     string               `json:"kind"`
	Count    int                  `json:"count"`
	Revenue  float64              `json:"revenue"`
	Children []ProjectRevenueNode `json:"children,omitempty"`
}

type projectRevenueOrderMix struct {
	Service  float64
	Product  float64
	Feeding  float64
	Boarding float64
}

func (r *StatsRepository) GetServiceRanking(shopID uint, startDate, endDate string) ([]ServiceRanking, error) {
	var rankings []ServiceRanking
	serviceNameExpr := `
		COALESCE(
			NULLIF(services.name, ''),
			CASE
				WHEN order_items.name LIKE '% · %' THEN SUBSTRING_INDEX(order_items.name, ' · ', -1)
				ELSE order_items.name
			END
		)
	`
	err := database.DB.Table("order_items").
		Select(serviceNameExpr+" as service_name, COUNT(*) as count, SUM(order_items.amount) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Joins("LEFT JOIN services ON services.id = order_items.item_id AND order_items.item_type = 1 AND services.deleted_at IS NULL").
		Where("orders.shop_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND order_items.item_type = 1 AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group(serviceNameExpr).
		Order("count DESC").
		Limit(10).
		Find(&rankings).Error
	return rankings, err
}

func (r *StatsRepository) GetProjectRevenueTree(shopID uint, startDate, endDate string) ([]ProjectRevenueNode, error) {
	type itemRevenueRow struct {
		OrderID       uint
		ItemType      int
		ItemID        uint
		Name          string
		Quantity      int
		Amount        float64
		FeedingPlanID *uint
	}
	type orderAdjustmentRow struct {
		ID                                uint
		DiscountAmount                    float64
		ServiceDiscountAmount             float64
		ProductDiscountAmount             float64
		AppointmentDepositDeductionAmount float64
	}
	var rows []itemRevenueRow
	if err := database.DB.Table("order_items").
		Select("order_items.order_id, order_items.item_type, order_items.item_id, order_items.name, order_items.quantity, order_items.amount, orders.feeding_plan_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL",
			shopID, startDate, endDate).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	serviceByID, serviceCategoryPath, err := r.projectRevenueServiceLookups(shopID)
	if err != nil {
		return nil, err
	}
	productByID, productCategoryByID, skuProductIDByID, err := r.projectRevenueProductLookups(shopID)
	if err != nil {
		return nil, err
	}

	roots := []ProjectRevenueNode{
		{Key: "service", Name: "服务", Kind: "root"},
		{Key: "product", Name: "商品", Kind: "root"},
		{Key: "feeding", Name: "上门喂养", Kind: "root"},
		{Key: "boarding", Name: "寄养", Kind: "root"},
	}
	rootIndexes := map[string]int{"service": 0, "product": 1, "feeding": 2, "boarding": 3}

	boardingOrderIDs := make(map[uint]struct{})
	for _, row := range rows {
		if row.ItemType == 4 || row.ItemType == 5 {
			boardingOrderIDs[row.OrderID] = struct{}{}
		}
	}

	revenueMixByOrderID := make(map[uint]*projectRevenueOrderMix)
	for _, row := range rows {
		mix := revenueMixByOrderID[row.OrderID]
		if mix == nil {
			mix = &projectRevenueOrderMix{}
			revenueMixByOrderID[row.OrderID] = mix
		}
		switch {
		case row.FeedingPlanID != nil:
			mix.Feeding += row.Amount
		case row.ItemType == 1:
			mix.Service += row.Amount
		case row.ItemType == 2:
			mix.Product += row.Amount
		case row.ItemType == 4 || row.ItemType == 5:
			mix.Boarding += row.Amount
		case row.ItemType == 3:
			if _, ok := boardingOrderIDs[row.OrderID]; ok {
				mix.Boarding += row.Amount
			}
		}
	}

	for _, row := range rows {
		count := row.Quantity
		if count < 1 {
			count = 1
		}
		name := cleanProjectItemName(row.Name)
		if name == "" {
			name = "未命名项目"
		}

		switch {
		case row.FeedingPlanID != nil:
			kind := "feeding"
			addProjectRevenuePath(&roots[rootIndexes[kind]], []string{name}, kind, count, row.Amount)
		case row.ItemType == 1:
			kind := "service"
			serviceName := name
			path := []string{"未分类服务"}
			if service, ok := serviceByID[row.ItemID]; ok {
				if strings.TrimSpace(service.Name) != "" {
					serviceName = service.Name
				}
				if service.CategoryID != nil {
					if categoryPath := serviceCategoryPath[*service.CategoryID]; len(categoryPath) > 0 {
						path = categoryPath
					}
				}
			}
			addProjectRevenuePath(&roots[rootIndexes[kind]], append(path, serviceName), kind, count, row.Amount)
		case row.ItemType == 2:
			kind := "product"
			productID := row.ItemID
			if mappedProductID, ok := skuProductIDByID[row.ItemID]; ok {
				productID = mappedProductID
			}
			productName := name
			categoryName := "未分类商品"
			if product, ok := productByID[productID]; ok {
				if strings.TrimSpace(product.Name) != "" {
					productName = product.Name
				}
				if category, ok := productCategoryByID[product.CategoryID]; ok && strings.TrimSpace(category.Name) != "" {
					categoryName = category.Name
				}
			}
			addProjectRevenuePath(&roots[rootIndexes[kind]], []string{categoryName, productName}, kind, count, row.Amount)
		case row.ItemType == 4 || row.ItemType == 5:
			kind := "boarding"
			addProjectRevenuePath(&roots[rootIndexes[kind]], []string{boardingItemGroup(row.ItemType), name}, kind, count, row.Amount)
		case row.ItemType == 6:
			if _, ok := boardingOrderIDs[row.OrderID]; ok {
				addProjectRevenueRootAdjustment(&roots[rootIndexes["boarding"]], row.Amount)
			}
		case row.ItemType == 3:
			if _, ok := boardingOrderIDs[row.OrderID]; ok {
				kind := "boarding"
				addProjectRevenuePath(&roots[rootIndexes[kind]], []string{"增值服务", name}, kind, count, row.Amount)
			}
		}
	}

	var adjustments []orderAdjustmentRow
	if err := database.DB.Table("orders").
		Select("orders.id, orders.discount_amount, orders.service_discount_amount, orders.product_discount_amount, orders.appointment_deposit_deduction_amount").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND orders.deleted_at IS NULL",
			shopID, startDate, endDate).
		Find(&adjustments).Error; err != nil {
		return nil, err
	}
	for _, adjustment := range adjustments {
		mix := revenueMixByOrderID[adjustment.ID]
		if mix == nil || mix.Boarding > 0 {
			continue
		}
		if adjustment.ServiceDiscountAmount > 0 {
			addProjectRevenueRootAdjustment(&roots[rootIndexes["service"]], -adjustment.ServiceDiscountAmount)
		}
		if adjustment.ProductDiscountAmount > 0 {
			addProjectRevenueRootAdjustment(&roots[rootIndexes["product"]], -adjustment.ProductDiscountAmount)
		}
		unallocatedDiscount := adjustment.DiscountAmount - adjustment.ServiceDiscountAmount - adjustment.ProductDiscountAmount
		if unallocatedDiscount < 0 {
			unallocatedDiscount = 0
		}
		applyProjectRevenueUnallocatedDiscount(roots, rootIndexes, mix, unallocatedDiscount+adjustment.AppointmentDepositDeductionAmount)
	}

	for i := range roots {
		sortProjectRevenueChildren(&roots[i])
	}
	return roots, nil
}

type CategoryStat struct {
	ServiceName string  `json:"service_name"`
	FurLevel    string  `json:"fur_level"`
	Count       int     `json:"count"`
	Revenue     float64 `json:"revenue"`
}

func (r *StatsRepository) GetCategoryStats(shopID uint, startDate, endDate string) ([]CategoryStat, error) {
	var stats []CategoryStat
	serviceNameExpr := `
		COALESCE(
			NULLIF(services.name, ''),
			CASE
				WHEN order_items.name LIKE '% · %' THEN SUBSTRING_INDEX(order_items.name, ' · ', -1)
				ELSE order_items.name
			END
		)
	`
	furLevelExpr := `
		COALESCE(
			NULLIF((
				SELECT pets_inner.fur_level
				FROM appointment_pets ap_inner
				JOIN pets pets_inner ON pets_inner.id = ap_inner.pet_id AND pets_inner.deleted_at IS NULL
				WHERE ap_inner.appointment_id = orders.appointment_id
					AND ap_inner.deleted_at IS NULL
					AND (
						order_items.name NOT LIKE '% · %'
						OR pets_inner.name = SUBSTRING_INDEX(order_items.name, ' · ', 1)
					)
				ORDER BY ap_inner.sort_order ASC, ap_inner.id ASC
				LIMIT 1
			), ''),
			NULLIF(pets.fur_level, ''),
			NULLIF(service_price_rules.name, ''),
			NULLIF(service_price_rules.fur_level, ''),
			'-'
		)
	`
	err := database.DB.Table("order_items").
		Select(serviceNameExpr+" as service_name, "+furLevelExpr+" as fur_level, COUNT(*) as count, SUM(order_items.amount) as revenue").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Joins("LEFT JOIN services ON services.id = order_items.item_id AND order_items.item_type = 1 AND services.deleted_at IS NULL").
		Joins("LEFT JOIN service_price_rules ON service_price_rules.service_id = order_items.item_id AND service_price_rules.price = order_items.unit_price AND service_price_rules.deleted_at IS NULL").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND order_items.item_type = 1 AND orders.deleted_at IS NULL AND order_items.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group(serviceNameExpr + ", " + furLevelExpr).
		Order("revenue DESC, service_name ASC, fur_level ASC").
		Find(&stats).Error
	return stats, err
}

func addProjectRevenuePath(root *ProjectRevenueNode, path []string, kind string, count int, revenue float64) {
	root.Count += count
	root.Revenue += revenue
	current := root
	for depth, name := range path {
		cleanName := strings.TrimSpace(name)
		if cleanName == "" {
			cleanName = "未分类"
		}
		key := fmt.Sprintf("%s:%d:%s", current.Key, depth, cleanName)
		childIndex := -1
		for i := range current.Children {
			if current.Children[i].Key == key {
				childIndex = i
				break
			}
		}
		if childIndex < 0 {
			current.Children = append(current.Children, ProjectRevenueNode{
				Key:  key,
				Name: cleanName,
				Kind: kind,
			})
			childIndex = len(current.Children) - 1
		}
		current = &current.Children[childIndex]
		current.Count += count
		current.Revenue += revenue
	}
}

func addProjectRevenueRootAdjustment(root *ProjectRevenueNode, revenue float64) {
	root.Revenue += revenue
}

func applyProjectRevenueUnallocatedDiscount(roots []ProjectRevenueNode, rootIndexes map[string]int, mix *projectRevenueOrderMix, discount float64) {
	if mix == nil || discount <= 0 {
		return
	}
	total := mix.Service + mix.Product + mix.Feeding
	if total <= 0 {
		return
	}
	if mix.Service > 0 {
		addProjectRevenueRootAdjustment(&roots[rootIndexes["service"]], -discount*mix.Service/total)
	}
	if mix.Product > 0 {
		addProjectRevenueRootAdjustment(&roots[rootIndexes["product"]], -discount*mix.Product/total)
	}
	if mix.Feeding > 0 {
		addProjectRevenueRootAdjustment(&roots[rootIndexes["feeding"]], -discount*mix.Feeding/total)
	}
}

func sortProjectRevenueChildren(node *ProjectRevenueNode) {
	sort.SliceStable(node.Children, func(i, j int) bool {
		if node.Children[i].Revenue != node.Children[j].Revenue {
			return node.Children[i].Revenue > node.Children[j].Revenue
		}
		if node.Children[i].Count != node.Children[j].Count {
			return node.Children[i].Count > node.Children[j].Count
		}
		return node.Children[i].Name < node.Children[j].Name
	})
	for i := range node.Children {
		sortProjectRevenueChildren(&node.Children[i])
	}
}

func cleanProjectItemName(name string) string {
	cleaned := strings.TrimSpace(name)
	if parts := strings.Split(cleaned, " · "); len(parts) > 1 {
		cleaned = strings.TrimSpace(parts[len(parts)-1])
	}
	return cleaned
}

func boardingItemGroup(itemType int) string {
	switch itemType {
	case 4:
		return "住宿"
	case 5:
		return "节假日加收"
	default:
		return "其他"
	}
}

func (r *StatsRepository) projectRevenueServiceLookups(shopID uint) (map[uint]model.Service, map[uint][]string, error) {
	var services []model.Service
	if err := database.DB.Where("shop_id = ?", shopID).Find(&services).Error; err != nil {
		return nil, nil, err
	}
	serviceByID := make(map[uint]model.Service, len(services))
	for _, service := range services {
		serviceByID[service.ID] = service
	}

	var categories []model.ServiceCategory
	if err := database.DB.Where("shop_id = ?", shopID).Find(&categories).Error; err != nil {
		return nil, nil, err
	}
	categoryByID := make(map[uint]model.ServiceCategory, len(categories))
	for _, category := range categories {
		categoryByID[category.ID] = category
	}
	pathCache := make(map[uint][]string, len(categories))
	var buildPath func(id uint) []string
	buildPath = func(id uint) []string {
		if path, ok := pathCache[id]; ok {
			return path
		}
		category, ok := categoryByID[id]
		if !ok {
			return nil
		}
		path := []string{}
		if category.ParentID != nil && *category.ParentID > 0 && *category.ParentID != category.ID {
			path = append(path, buildPath(*category.ParentID)...)
		}
		path = append(path, category.Name)
		pathCache[id] = path
		return path
	}
	for id := range categoryByID {
		buildPath(id)
	}
	return serviceByID, pathCache, nil
}

func (r *StatsRepository) projectRevenueProductLookups(shopID uint) (map[uint]model.Product, map[uint]model.ProductCategory, map[uint]uint, error) {
	var products []model.Product
	if err := database.DB.Where("shop_id = ?", shopID).Find(&products).Error; err != nil {
		return nil, nil, nil, err
	}
	productByID := make(map[uint]model.Product, len(products))
	for _, product := range products {
		productByID[product.ID] = product
	}

	var categories []model.ProductCategory
	if err := database.DB.Where("shop_id = ?", shopID).Find(&categories).Error; err != nil {
		return nil, nil, nil, err
	}
	categoryByID := make(map[uint]model.ProductCategory, len(categories))
	for _, category := range categories {
		categoryByID[category.ID] = category
	}

	var skus []model.ProductSKU
	if err := database.DB.Joins("JOIN products ON products.id = product_skus.product_id AND products.shop_id = ? AND products.deleted_at IS NULL", shopID).
		Find(&skus).Error; err != nil {
		return nil, nil, nil, err
	}
	skuProductIDByID := make(map[uint]uint, len(skus))
	for _, sku := range skus {
		skuProductIDByID[sku.ID] = sku.ProductID
	}
	return productByID, categoryByID, skuProductIDByID, nil
}

type StaffPerformance struct {
	StaffID               uint    `json:"staff_id"`
	StaffName             string  `json:"staff_name"`
	ApptCount             int     `json:"appointment_count"`
	Revenue               float64 `json:"revenue"`
	ProductRevenue        float64 `json:"product_revenue"`
	CommissionRate        float64 `json:"commission_rate"`
	ProductCommissionRate float64 `json:"product_commission_rate"`
	Commission            float64 `json:"commission"`
}

type StaffCommissionDetail struct {
	OrderID        uint    `json:"order_id"`
	OrderNo        string  `json:"order_no"`
	Date           string  `json:"date"`
	PayMethod      string  `json:"pay_method"`
	PayMethodLabel string  `json:"pay_method_label"`
	PayAmount      float64 `json:"pay_amount"`
	ServiceAmount  float64 `json:"service_amount"`
	ProductAmount  float64 `json:"product_amount"`
	CommissionRate float64 `json:"commission_rate"`
	Commission     float64 `json:"commission"`
	Formula        string  `json:"formula"`
	CustomerName   string  `json:"customer_name"`
	PetSummary     string  `json:"pet_summary"`
	Remark         string  `json:"remark"`
}

func (r *StatsRepository) GetStaffPerformance(shopID uint, startDate, endDate string) ([]StaffPerformance, error) {
	var perfs []StaffPerformance
	err := database.DB.Table("orders").
		Select(`
			orders.staff_id,
			staffs.name as staff_name,
			staffs.commission_rate,
			staffs.product_commission_rate,
			COUNT(*) as appt_count,
			SUM(orders.pay_amount) as revenue,
			COALESCE(SUM(CASE WHEN orders.product_total - orders.product_discount_amount > 0 THEN orders.product_total - orders.product_discount_amount ELSE 0 END), 0) as product_revenue,
			COALESCE(SUM(orders.commission), 0) as commission
		`).
		Joins("JOIN staffs ON staffs.id = orders.staff_id").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Where("orders.shop_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND orders.staff_id IS NOT NULL AND orders.deleted_at IS NULL",
			shopID, startDate, endDate).
		Group("orders.staff_id, staffs.name, staffs.commission_rate, staffs.product_commission_rate").
		Order("revenue DESC").
		Find(&perfs).Error
	return perfs, err
}

func (r *StatsRepository) GetStaffCommissionDetails(shopID, staffID uint, startDate, endDate string) ([]StaffCommissionDetail, error) {
	type staffCommissionDetailRow struct {
		OrderID               uint
		OrderNo               string
		Date                  string
		PayMethod             string
		PayAmount             float64
		ServiceTotal          float64
		ServiceDiscountAmount float64
		ProductAmount         float64
		CommissionRate        float64
		Commission            float64
		CustomerName          string
		PetName               string
		Remark                string
	}

	var rows []staffCommissionDetailRow
	err := database.DB.Table("orders").
		Select(`
			orders.id AS order_id,
			orders.order_no,
			COALESCE(appointments.date, DATE(orders.created_at)) AS date,
			orders.pay_method,
			orders.pay_amount,
			orders.service_total,
			orders.service_discount_amount,
			CASE WHEN orders.product_total - orders.product_discount_amount > 0 THEN orders.product_total - orders.product_discount_amount ELSE 0 END AS product_amount,
			staffs.commission_rate,
			orders.commission,
			COALESCE(NULLIF(customers.nickname, ''), customers.phone, '-') AS customer_name,
			COALESCE(pets.name, '') AS pet_name,
			orders.remark
		`).
		Joins("JOIN staffs ON staffs.id = orders.staff_id").
		Joins("LEFT JOIN appointments ON appointments.id = orders.appointment_id AND appointments.deleted_at IS NULL").
		Joins("LEFT JOIN customers ON customers.id = orders.customer_id").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Where("orders.shop_id = ? AND orders.staff_id = ? AND COALESCE(appointments.date, DATE(orders.created_at)) >= ? AND COALESCE(appointments.date, DATE(orders.created_at)) <= ? AND orders.status = 1 AND orders.commission > 0 AND orders.deleted_at IS NULL",
			shopID, staffID, startDate, endDate).
		Order("COALESCE(appointments.date, DATE(orders.created_at)) ASC, orders.id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	details := make([]StaffCommissionDetail, 0, len(rows))
	for _, row := range rows {
		serviceAmount := row.PayAmount - row.ProductAmount
		if row.ServiceTotal > 0 {
			serviceAmount = row.ServiceTotal - row.ServiceDiscountAmount
		}
		if serviceAmount < 0 {
			serviceAmount = 0
		}
		details = append(details, StaffCommissionDetail{
			OrderID:        row.OrderID,
			OrderNo:        row.OrderNo,
			Date:           row.Date,
			PayMethod:      row.PayMethod,
			PayMethodLabel: paymentBreakdownLabel(normalizeDashboardPayMethod(row.PayMethod)),
			PayAmount:      row.PayAmount,
			ServiceAmount:  serviceAmount,
			ProductAmount:  row.ProductAmount,
			CommissionRate: row.CommissionRate,
			Commission:     row.Commission,
			Formula:        formatCommissionFormula(row.PayAmount, row.ProductAmount, serviceAmount, row.PayMethod, row.CommissionRate, row.Commission),
			CustomerName:   row.CustomerName,
			PetSummary:     row.PetName,
			Remark:         row.Remark,
		})
	}
	return details, nil
}

func normalizeDashboardPayMethod(method string) string {
	switch method {
	case "wechat":
		return "wechat"
	case "qrcode", "alipay":
		return "qrcode"
	case "meituan":
		return "meituan"
	case "balance", "card":
		return "balance"
	case "mixed_balance":
		return "mixed_balance"
	default:
		return "other"
	}
}

func formatCommissionFormula(payAmount, productAmount, serviceAmount float64, payMethod string, rate float64, commission float64) string {
	rateText := fmt.Sprintf("%.0f%%", rate)
	if productAmount > 0 && payMethod == "meituan" {
		return fmt.Sprintf("(¥%.2f - 商品¥%.2f) × 0.9 × %s = ¥%.2f", payAmount, productAmount, rateText, commission)
	}
	if productAmount > 0 {
		return fmt.Sprintf("(¥%.2f - 商品¥%.2f) × %s = ¥%.2f", payAmount, productAmount, rateText, commission)
	}
	if payMethod == "meituan" {
		return fmt.Sprintf("¥%.2f × 0.9 × %s = ¥%.2f", serviceAmount, rateText, commission)
	}
	return fmt.Sprintf("¥%.2f × %s = ¥%.2f", serviceAmount, rateText, commission)
}
