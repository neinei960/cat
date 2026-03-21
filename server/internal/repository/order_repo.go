package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
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

func (r *OrderRepository) FindByShopPaged(shopID uint, status *int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	db := database.DB.Model(&model.Order{}).Where("shop_id = ?", shopID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Items").
		Order("id DESC").Offset(offset).Limit(pageSize).Find(&orders).Error
	return orders, total, err
}

func (r *OrderRepository) Search(shopID uint, keyword string, status *int, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	like := "%" + keyword + "%"
	db := database.DB.Model(&model.Order{}).
		Joins("LEFT JOIN customers ON customers.id = orders.customer_id").
		Joins("LEFT JOIN pets ON pets.id = orders.pet_id").
		Where("orders.shop_id = ? AND (orders.order_no LIKE ? OR customers.nickname LIKE ? OR customers.phone LIKE ? OR pets.name LIKE ?)",
			shopID, like, like, like, like)
	if status != nil {
		db = db.Where("orders.status = ?", *status)
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
