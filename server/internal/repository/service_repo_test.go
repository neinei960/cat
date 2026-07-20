package repository

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceRepoTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:service-repo-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(
		&model.Shop{},
		&model.Customer{},
		&model.Service{},
		&model.ServicePriceRule{},
		&model.Appointment{},
		&model.AppointmentService{},
		&model.AppointmentPet{},
		&model.AppointmentPetService{},
		&model.Order{},
		&model.OrderItem{},
	); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
}

func TestFindByShopIDCustomerUsageOrdersFrequentServicesFirst(t *testing.T) {
	setupServiceRepoTestDB(t)

	shop := model.Shop{Name: "test shop"}
	if err := database.DB.Create(&shop).Error; err != nil {
		t.Fatalf("create shop: %v", err)
	}
	customer := model.Customer{ShopID: shop.ID, Nickname: "常客"}
	otherCustomer := model.Customer{ShopID: shop.ID, Nickname: "其他客户"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if err := database.DB.Create(&otherCustomer).Error; err != nil {
		t.Fatalf("create other customer: %v", err)
	}

	rare := model.Service{ShopID: shop.ID, Name: "很少买", BasePrice: 88, Duration: 60, SortOrder: 1, Status: 1}
	favorite := model.Service{ShopID: shop.ID, Name: "常买服务", BasePrice: 128, Duration: 60, SortOrder: 2, Status: 1}
	otherCustomerFavorite := model.Service{ShopID: shop.ID, Name: "别人常买", BasePrice: 108, Duration: 60, SortOrder: 3, Status: 1}
	if err := database.DB.Create(&rare).Error; err != nil {
		t.Fatalf("create rare service: %v", err)
	}
	if err := database.DB.Create(&favorite).Error; err != nil {
		t.Fatalf("create favorite service: %v", err)
	}
	if err := database.DB.Create(&otherCustomerFavorite).Error; err != nil {
		t.Fatalf("create other favorite service: %v", err)
	}

	orderA := model.Order{ShopID: shop.ID, CustomerID: &customer.ID, OrderNo: "A", Status: 1}
	orderB := model.Order{ShopID: shop.ID, CustomerID: &customer.ID, OrderNo: "B", Status: 1}
	orderOther := model.Order{ShopID: shop.ID, CustomerID: &otherCustomer.ID, OrderNo: "C", Status: 1}
	if err := database.DB.Create(&[]model.Order{orderA, orderB, orderOther}).Error; err != nil {
		t.Fatalf("create orders: %v", err)
	}
	var orders []model.Order
	if err := database.DB.Order("order_no ASC").Find(&orders).Error; err != nil {
		t.Fatalf("load orders: %v", err)
	}

	items := []model.OrderItem{
		{OrderID: orders[0].ID, ItemType: 1, ItemID: favorite.ID, Name: favorite.Name, Quantity: 1, UnitPrice: 128, Amount: 128},
		{OrderID: orders[1].ID, ItemType: 1, ItemID: favorite.ID, Name: favorite.Name, Quantity: 1, UnitPrice: 128, Amount: 128},
		{OrderID: orders[2].ID, ItemType: 1, ItemID: otherCustomerFavorite.ID, Name: otherCustomerFavorite.Name, Quantity: 1, UnitPrice: 108, Amount: 108},
	}
	if err := database.DB.Create(&items).Error; err != nil {
		t.Fatalf("create order items: %v", err)
	}

	services, _, err := NewServiceRepository().FindByShopID(shop.ID, 1, 20, "customer_usage", customer.ID)
	if err != nil {
		t.Fatalf("list services: %v", err)
	}
	if len(services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(services))
	}
	if services[0].ID != favorite.ID {
		t.Fatalf("expected favorite service first, got %s", services[0].Name)
	}
	if services[0].CustomerUsageCount != 2 {
		t.Fatalf("expected favorite customer usage 2, got %d", services[0].CustomerUsageCount)
	}
	if services[1].ID != rare.ID {
		t.Fatalf("expected original sort order after customer favorite, got %s", services[1].Name)
	}
}
