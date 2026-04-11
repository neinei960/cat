package repository

import (
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type OrderRepository struct{}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{}
}

func (r *OrderRepository) Create(order *model.Order) error {
	return database.DB.Create(order).Error
}

func (r *OrderRepository) FindByID(id uint) (*model.Order, error) {
	var order model.Order
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").Preload("Appointment.Pets").Preload("Appointment.Pets.Pet").
		Preload("FeedingPlan").Preload("FeedingPlan.Pets").Preload("FeedingPlan.Pets.Pet").
		First(&order, id).Error
	return &order, err
}

func (r *OrderRepository) FindByIDUnscoped(id uint) (*model.Order, error) {
	var order model.Order
	err := database.DB.Unscoped().
		Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").Preload("Appointment.Pets").Preload("Appointment.Pets.Pet").
		Preload("FeedingPlan").Preload("FeedingPlan.Pets").Preload("FeedingPlan.Pets.Pet").
		First(&order, id).Error
	return &order, err
}

func (r *OrderRepository) FindByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").Preload("Appointment.Pets").Preload("Appointment.Pets.Pet").
		Preload("FeedingPlan").Preload("FeedingPlan.Pets").Preload("FeedingPlan.Pets.Pet").
		Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

type OrderFilter struct {
	Status         *int
	DateFrom       string
	DateTo         string
	CustomerID     uint
	StaffID        uint
	PayMethod      string
	ProductKeyword string
	OrderKind      string
}

func (r *OrderRepository) applyFilters(db *gorm.DB, shopID uint, f OrderFilter) *gorm.DB {
	db = db.Where("shop_id = ?", shopID)
	if f.Status != nil {
		db = db.Where("status = ?", *f.Status)
	}
	if f.DateFrom != "" {
		db = db.Where("DATE(created_at) >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		db = db.Where("DATE(created_at) <= ?", f.DateTo)
	}
	if f.StaffID > 0 {
		db = db.Where("staff_id = ?", f.StaffID)
	}
	if f.CustomerID > 0 {
		db = db.Where("customer_id = ?", f.CustomerID)
	}
	if f.PayMethod != "" {
		db = applyPayMethodFilter(db, "pay_method", f.PayMethod)
	}
	if f.ProductKeyword != "" {
		like := "%" + f.ProductKeyword + "%"
		db = db.Where(
			"EXISTS (?)",
			database.DB.Model(&model.OrderItem{}).
				Select("1").
				Where("order_items.order_id = orders.id AND order_items.deleted_at IS NULL AND order_items.item_type = ? AND order_items.name LIKE ?", 2, like),
		)
	}
	if f.OrderKind == "feeding" {
		db = db.Where("feeding_plan_id IS NOT NULL AND feeding_plan_id > 0")
	}
	return db
}

func (r *OrderRepository) FindByShopPaged(shopID uint, f OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := r.applyFilters(database.DB.Model(&model.Order{}), shopID, f)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").Preload("Appointment.Pets").Preload("Appointment.Pets.Pet").
		Preload("FeedingPlan").Preload("FeedingPlan.Pets").Preload("FeedingPlan.Pets.Pet").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Search(shopID uint, keyword string, f OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	like := "%" + keyword + "%"
	subQuery := database.DB.Model(&model.Order{}).
		Select("DISTINCT orders.id").
		Joins("LEFT JOIN customers ON customers.id = orders.customer_id").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Where("orders.shop_id = ? AND (orders.order_no LIKE ? OR customers.nickname LIKE ? OR customers.phone LIKE ? OR pets.name LIKE ? OR order_items.name LIKE ?)",
			shopID, like, like, like, like, like)
	db := database.DB.Model(&model.Order{}).Where("orders.id IN (?)", subQuery)
	if f.Status != nil {
		db = db.Where("orders.status = ?", *f.Status)
	}
	if f.DateFrom != "" {
		db = db.Where("DATE(orders.created_at) >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		db = db.Where("DATE(orders.created_at) <= ?", f.DateTo)
	}
	if f.StaffID > 0 {
		db = db.Where("orders.staff_id = ?", f.StaffID)
	}
	if f.CustomerID > 0 {
		db = db.Where("orders.customer_id = ?", f.CustomerID)
	}
	if f.PayMethod != "" {
		db = applyPayMethodFilter(db, "orders.pay_method", f.PayMethod)
	}
	if f.ProductKeyword != "" {
		productLike := "%" + f.ProductKeyword + "%"
		db = db.Where(
			"EXISTS (?)",
			database.DB.Model(&model.OrderItem{}).
				Select("1").
				Where("order_items.order_id = orders.id AND order_items.deleted_at IS NULL AND order_items.item_type = ? AND order_items.name LIKE ?", 2, productLike),
		)
	}
	if f.OrderKind == "feeding" {
		db = db.Where("orders.feeding_plan_id IS NOT NULL AND orders.feeding_plan_id > 0")
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").Preload("Appointment.Pets").Preload("Appointment.Pets.Pet").
		Preload("FeedingPlan").Preload("FeedingPlan.Pets").Preload("FeedingPlan.Pets.Pet").
		Order("orders.id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func applyPayMethodFilter(db *gorm.DB, column string, payMethod string) *gorm.DB {
	switch payMethod {
	case "qrcode":
		return db.Where(column+" IN ?", []string{"qrcode", "alipay"})
	case "wechat":
		return db.Where(column+" = ?", "wechat")
	case "meituan":
		return db.Where(column+" = ?", "meituan")
	case "balance":
		return db.Where(column+" IN ?", []string{"balance", "card"})
	case "other":
		return db.Where(column+" IN ?", []string{"other", "cash"})
	default:
		return db.Where(column+" = ?", payMethod)
	}
}

func (r *OrderRepository) FindByCustomer(customerID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := database.DB.Model(&model.Order{}).Where("customer_id = ?", customerID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Pet").Preload("Items").Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Update(order *model.Order) error {
	return database.DB.Save(order).Error
}

func (r *OrderRepository) UpdateRemark(id uint, remark string) error {
	return database.DB.Model(&model.Order{}).Where("id = ?", id).Update("remark", remark).Error
}

func (r *OrderRepository) CreateItems(items []model.OrderItem) error {
	return database.DB.Create(&items).Error
}

func (r *OrderRepository) CountByAppointment(appointmentID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Order{}).Where("appointment_id = ?", appointmentID).Count(&count).Error
	return count, err
}

func (r *OrderRepository) FindByAppointment(appointmentID uint) ([]model.Order, error) {
	var orders []model.Order
	err := database.DB.Where("appointment_id = ?", appointmentID).Order("id ASC").Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) FindDeleted(shopID uint, page, pageSize int, since time.Time) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := database.DB.Unscoped().Model(&model.Order{}).
		Where("shop_id = ? AND deleted_at IS NOT NULL AND deleted_at >= ?", shopID, since)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Appointment").Preload("FeedingPlan").
		Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Restore(shopID, id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Unscoped().Where("id = ? AND shop_id = ? AND deleted_at IS NOT NULL", id, shopID).First(&order).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.OrderItem{}).Where("order_id = ?", id).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return tx.Unscoped().Model(&model.Order{}).Where("id = ?", id).Update("deleted_at", nil).Error
	})
}

func (r *OrderRepository) CleanupExpired(before time.Time) (int64, error) {
	var orders []model.Order
	if err := database.DB.Unscoped().
		Select("id").
		Where("deleted_at IS NOT NULL AND deleted_at < ?", before).
		Find(&orders).Error; err != nil {
		return 0, err
	}
	if len(orders) == 0 {
		return 0, nil
	}

	ids := make([]uint, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}

	var deletedCount int64
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("order_id IN ?", ids).Delete(&model.OrderItem{}).Error; err != nil {
			return err
		}
		result := tx.Unscoped().Where("id IN ?", ids).Delete(&model.Order{})
		deletedCount = result.RowsAffected
		return result.Error
	})
	return deletedCount, err
}
