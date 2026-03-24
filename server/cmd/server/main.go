package main

import (
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/router"
	"github.com/neinei960/cat/server/pkg/database"
)

func main() {
	syscall.Umask(0022)
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := database.AutoMigrate(
		&model.Shop{},
		&model.Staff{},
		&model.Customer{},
		&model.CustomerTag{},
		&model.CustomerTagRelation{},
		&model.Pet{},
		&model.MemberCardTemplate{},
		&model.MemberCardDiscount{},
		&model.MemberCard{},
		&model.RechargeRecord{},
		&model.ServiceCategory{},
		&model.Service{},
		&model.ServicePriceRule{},
		&model.ServiceAddon{},
		&model.StaffService{},
		&model.StaffSchedule{},
		&model.Appointment{},
		&model.AppointmentService{},
		&model.AppointmentPet{},
		&model.AppointmentPetService{},
		&model.Order{},
		&model.OrderItem{},
		&model.NotificationLog{},
		&model.DailyStats{},
		&model.Product{},
		&model.ProductSKU{},
		&model.ProductCategory{},
	); err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	// 商品分类种子数据
	seedProductCategories()

	// 定时清理已删除超过1天的客户数据
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		repo := repository.NewCustomerRepository()
		for range ticker.C {
			before := time.Now().Add(-24 * time.Hour)
			if n, err := repo.CleanupExpired(before); err != nil {
				log.Printf("Cleanup expired customers error: %v", err)
			} else if n > 0 {
				log.Printf("Cleaned up %d expired deleted customers", n)
			}
		}
	}()

	r := router.Setup(config.AppConfig.Server.Mode)

	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func seedProductCategories() {
	var count int64
	database.DB.Model(&model.ProductCategory{}).Count(&count)
	if count > 0 {
		return
	}

	categories := []model.ProductCategory{
		{ShopID: 1, Name: "猫粮", SortOrder: 1, Status: 1},
		{ShopID: 1, Name: "猫砂", SortOrder: 2, Status: 1},
		{ShopID: 1, Name: "洗护用品", SortOrder: 3, Status: 1},
		{ShopID: 1, Name: "营养品", SortOrder: 4, Status: 1},
		{ShopID: 1, Name: "猫玩具", SortOrder: 5, Status: 1},
		{ShopID: 1, Name: "猫日用品", SortOrder: 6, Status: 1},
		{ShopID: 1, Name: "其他", SortOrder: 7, Status: 1},
	}
	for _, c := range categories {
		database.DB.Create(&c)
	}
	log.Println("Seeded product categories")
}
