package repository

import (
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
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").Preload("Appointment").
		First(&order, id).Error
	return &order, err
}

func (r *OrderRepository) FindByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := database.DB.Preload("Customer").Preload("Staff").Preload("Items").
		Where("order_no = ?", orderNo).First(&order).Error
	return &order, err
}

type OrderFilter struct {
	Status    *int
	DateFrom  string
	DateTo    string
	StaffID   uint
	PayMethod string
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
	if f.PayMethod != "" {
		db = db.Where("pay_method = ?", f.PayMethod)
	}
	return db
}

func (r *OrderRepository) FindByShopPaged(shopID uint, f OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := r.applyFilters(database.DB.Model(&model.Order{}), shopID, f)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Search(shopID uint, keyword string, f OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	like := "%" + keyword + "%"
	db := database.DB.Model(&model.Order{}).
		Joins("LEFT JOIN customers ON customers.id = orders.customer_id").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id AND order_items.deleted_at IS NULL").
		Where("orders.shop_id = ? AND (orders.order_no LIKE ? OR customers.nickname LIKE ? OR customers.phone LIKE ? OR pets.name LIKE ? OR order_items.name LIKE ?)",
			shopID, like, like, like, like, like).
		Distinct("orders.id")
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
	if f.PayMethod != "" {
		db = db.Where("orders.pay_method = ?", f.PayMethod)
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").
		Order("orders.id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
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
