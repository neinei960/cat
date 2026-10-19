package repository

import (
	"fmt"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOrderSelfUseFilter(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = oldDB; sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}, &model.FeedingPlanPet{}); err != nil {
		t.Fatal(err)
	}
	for _, o := range []model.Order{
		{OrderNo: "test-self-paid", ShopID: 1, Remark: "员工自用商品", Status: 1},
		{OrderNo: "test-self-unpaid", ShopID: 1, Remark: "自用", Status: 0},
		{OrderNo: "test-normal", ShopID: 1, Remark: "普通订单", Status: 1},
		{OrderNo: "test-other-shop", ShopID: 2, Remark: "自用", Status: 1},
	} {
		if err := db.Create(&o).Error; err != nil {
			t.Fatal(err)
		}
	}
	paid := 1
	for _, keyword := range []string{"", "test-"} {
		for _, tc := range []struct {
			name   string
			filter OrderFilter
			want   int64
		}{
			{"enabled", OrderFilter{SelfUse: true}, 2},
			{"paid", OrderFilter{SelfUse: true, Status: &paid}, 1},
			{"date", OrderFilter{SelfUse: true, DateTo: "2000-01-01"}, 0},
			{"reset", OrderFilter{}, 3},
		} {
			t.Run(keyword+tc.name, func(t *testing.T) {
				repo := NewOrderRepository()
				var rows []model.Order
				var total int64
				var err error
				if keyword == "" {
					rows, total, err = repo.FindByShopPaged(1, tc.filter, 1, 1)
				} else {
					rows, total, err = repo.Search(1, keyword, tc.filter, 1, 1)
				}
				if err != nil {
					t.Fatal(err)
				}
				if total != tc.want || len(rows) > 1 || (tc.want > 0 && len(rows) != 1) {
					t.Fatalf("total=%d rows=%d, want total=%d", total, len(rows), tc.want)
				}
			})
		}
	}
}

func TestOrderSearchFeedingPets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	oldDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = oldDB; sqlDB, _ := db.DB(); _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}, &model.FeedingPlanPet{}); err != nil {
		t.Fatal(err)
	}
	create := func(value interface{}) {
		t.Helper()
		if err := db.Create(value).Error; err != nil {
			t.Fatal(err)
		}
	}
	pet := model.Pet{ShopID: 1, Name: "奈奈新名"}
	create(&pet)
	plan := model.FeedingPlan{ShopID: 1, CustomerID: 1, StartDate: "2026-09-02", EndDate: "2026-09-03"}
	create(&plan)
	for i := 0; i < 2; i++ {
		create(&model.FeedingPlanPet{FeedingPlanID: plan.ID, PetID: pet.ID, PetNameSnapshot: "奈奈旧名"})
	}
	order := model.Order{ShopID: 1, OrderNo: "feeding-search", FeedingPlanID: &plan.ID}
	create(&order)
	create(&model.OrderItem{OrderID: order.ID, Name: "上门喂养", ItemType: 1})
	create(&model.OrderItem{OrderID: order.ID, Name: "陪玩", ItemType: 3})
	create(&model.Order{ShopID: 2, OrderNo: "other-shop", FeedingPlanID: &plan.ID})
	paid := 1
	for _, tc := range []struct {
		keyword string
		shop    uint
		filter  OrderFilter
		want    int64
	}{
		{"奈奈旧名", 1, OrderFilter{}, 1},
		{"奈奈新名", 1, OrderFilter{}, 1},
		{"奈奈", 1, OrderFilter{OrderKind: "feeding"}, 1},
		{"奈奈", 1, OrderFilter{Status: &paid}, 0},
		{"奈奈", 1, OrderFilter{DateTo: "2000-01-01"}, 0},
		{"奈奈", 2, OrderFilter{}, 0},
		{"上门喂养", 1, OrderFilter{}, 1},
	} {
		t.Run(fmt.Sprintf("%s-%d-%v", tc.keyword, tc.shop, tc.filter), func(t *testing.T) {
			orders, total, err := NewOrderRepository().Search(tc.shop, tc.keyword, tc.filter, 1, 20)
			if err != nil {
				t.Fatal(err)
			}
			if total != tc.want || int64(len(orders)) != tc.want {
				t.Fatalf("got total=%d rows=%d, want %d", total, len(orders), tc.want)
			}
		})
	}
	if err := db.Where("feeding_plan_id = ?", plan.ID).Delete(&model.FeedingPlanPet{}).Error; err != nil {
		t.Fatal(err)
	}
	_, total, err := NewOrderRepository().Search(1, "奈奈", OrderFilter{}, 1, 20)
	if err != nil || total != 0 {
		t.Fatalf("deleted links: total=%d err=%v", total, err)
	}
}
