package service

import (
	"fmt"
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateFeedingPlanAllowsBaseServiceWithoutAddons(t *testing.T) {
	setupFeedingServiceTestDB(t)
	customer, pet := seedFeedingCustomerAndPet(t, 1)
	svc := NewFeedingService(
		repository.NewFeedingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)

	plan, err := svc.CreatePlan(1, 1, FeedingPlanInput{
		CustomerID: customer.ID,
		AddressSnapshot: FeedingAddressSnapshot{
			Address:  "魔方公寓702",
			Detail:   "停车、楼栋等",
			DoorCode: "临时密码",
		},
		ContactName:  "zwaang",
		ContactPhone: "13776833249",
		StartDate:    "2026-05-03",
		EndDate:      "2026-05-03",
		SelectedDates: []string{
			"2026-05-03",
		},
		Pets: []FeedingPlanPetInput{
			{PetID: pet.ID},
		},
		ItemCodes: []string{},
	})
	if err != nil {
		t.Fatalf("create feeding plan without addon items: %v", err)
	}
	if got := len(plan.Visits); got != 1 {
		t.Fatalf("expected 1 visit, got %d", got)
	}
	if got := len(plan.Visits[0].Items); got != 0 {
		t.Fatalf("expected no addon visit items, got %d", got)
	}
	if got := roundMoney(plan.TotalAmount); got != 85 {
		t.Fatalf("expected base service amount 85, got %.2f", got)
	}
}

func TestCreateFeedingPlanDefaultsDepositForSelectedHoliday(t *testing.T) {
	setupFeedingServiceTestDB(t)
	customer, pet := seedFeedingCustomerAndPet(t, 1)
	seedFeedingHoliday(t, 1, "2026-05-03")
	svc := newFeedingServiceForTest()

	plan, err := svc.CreatePlan(1, 1, FeedingPlanInput{
		CustomerID: customer.ID,
		AddressSnapshot: FeedingAddressSnapshot{
			Address: "魔方公寓702",
		},
		ContactName:   "zwaang",
		ContactPhone:  "13776833249",
		StartDate:     "2026-05-03",
		EndDate:       "2026-05-03",
		SelectedDates: []string{"2026-05-03"},
		Pets:          []FeedingPlanPetInput{{PetID: pet.ID}},
	})
	if err != nil {
		t.Fatalf("create holiday feeding plan: %v", err)
	}
	if got := roundMoney(plan.Deposit); got != 200 {
		t.Fatalf("expected default holiday deposit 200, got %.2f", got)
	}
	if got := roundMoney(plan.UnpaidAmount); got != 0 {
		t.Fatalf("expected unpaid amount clamped to 0, got %.2f", got)
	}
}

func TestCreateFeedingPlanUsesCustomHolidayDeposit(t *testing.T) {
	setupFeedingServiceTestDB(t)
	customer, pet := seedFeedingCustomerAndPet(t, 1)
	seedFeedingHoliday(t, 1, "2026-05-03")
	svc := newFeedingServiceForTest()
	deposit := 50.0

	plan, err := svc.CreatePlan(1, 1, FeedingPlanInput{
		CustomerID: customer.ID,
		AddressSnapshot: FeedingAddressSnapshot{
			Address: "魔方公寓702",
		},
		ContactName:   "zwaang",
		ContactPhone:  "13776833249",
		StartDate:     "2026-05-03",
		EndDate:       "2026-05-03",
		SelectedDates: []string{"2026-05-03"},
		Pets:          []FeedingPlanPetInput{{PetID: pet.ID}},
		Deposit:       &deposit,
	})
	if err != nil {
		t.Fatalf("create holiday feeding plan with custom deposit: %v", err)
	}
	if got := roundMoney(plan.Deposit); got != 50 {
		t.Fatalf("expected custom holiday deposit 50, got %.2f", got)
	}
	if got := roundMoney(plan.UnpaidAmount); got != 45 {
		t.Fatalf("expected unpaid amount 45, got %.2f", got)
	}
}

func TestCreateFeedingPlanDailyPlayDefaultsAllSelectedDates(t *testing.T) {
	setupFeedingServiceTestDB(t)
	customer, pet := seedFeedingCustomerAndPet(t, 1)
	svc := newFeedingServiceForTest()

	plan, err := svc.CreatePlan(1, 1, FeedingPlanInput{
		CustomerID: customer.ID,
		AddressSnapshot: FeedingAddressSnapshot{
			Address: "魔方公寓702",
		},
		ContactName:   "zwaang",
		ContactPhone:  "13776833249",
		StartDate:     "2026-05-03",
		EndDate:       "2026-05-05",
		SelectedDates: []string{"2026-05-03", "2026-05-04", "2026-05-05"},
		Pets:          []FeedingPlanPetInput{{PetID: pet.ID}},
		ItemCodes:     []string{"play"},
		PlayMode:      "daily",
	})
	if err != nil {
		t.Fatalf("create daily play feeding plan: %v", err)
	}
	playDates := parseJSONStringDates(plan.PlayDatesJSON)
	if got := len(playDates); got != 3 {
		t.Fatalf("expected all 3 selected dates as play dates, got %d: %v", got, playDates)
	}
	for i, want := range []string{"2026-05-03", "2026-05-04", "2026-05-05"} {
		if playDates[i] != want {
			t.Fatalf("expected play date %d to be %s, got %s", i, want, playDates[i])
		}
	}
}

func TestCreateFeedingPlanCountPlayDoesNotDefaultPlayDates(t *testing.T) {
	setupFeedingServiceTestDB(t)
	customer, pet := seedFeedingCustomerAndPet(t, 1)
	svc := newFeedingServiceForTest()

	plan, err := svc.CreatePlan(1, 1, FeedingPlanInput{
		CustomerID: customer.ID,
		AddressSnapshot: FeedingAddressSnapshot{
			Address: "魔方公寓702",
		},
		ContactName:   "zwaang",
		ContactPhone:  "13776833249",
		StartDate:     "2026-05-03",
		EndDate:       "2026-05-05",
		SelectedDates: []string{"2026-05-03", "2026-05-04", "2026-05-05"},
		Pets:          []FeedingPlanPetInput{{PetID: pet.ID}},
		ItemCodes:     []string{"play"},
		PlayMode:      "count",
		PlayCount:     2,
	})
	if err != nil {
		t.Fatalf("create count play feeding plan: %v", err)
	}
	if got := len(parseJSONStringDates(plan.PlayDatesJSON)); got != 0 {
		t.Fatalf("expected no default play dates for count mode, got %d", got)
	}
}

func setupFeedingServiceTestDB(t *testing.T) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	database.DB = db
	if err := database.DB.AutoMigrate(
		&model.Customer{},
		&model.Pet{},
		&model.Staff{},
		&model.MemberCardTemplate{},
		&model.MemberCard{},
		&model.CustomerTag{},
		&model.FeedingSetting{},
		&model.FeedingPlan{},
		&model.FeedingPlanPet{},
		&model.FeedingPlanRule{},
		&model.FeedingVisit{},
		&model.FeedingVisitItem{},
		&model.FeedingVisitLog{},
		&model.FeedingVisitMedia{},
		&model.BoardingHoliday{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func newFeedingServiceForTest() *FeedingService {
	return NewFeedingService(
		repository.NewFeedingRepository(),
		repository.NewOrderRepository(),
		repository.NewCustomerRepository(),
		repository.NewPetRepository(),
	)
}

func seedFeedingCustomerAndPet(t *testing.T, shopID uint) (model.Customer, model.Pet) {
	t.Helper()

	customer := model.Customer{
		ShopID:        shopID,
		Phone:         "13776833249",
		Nickname:      "zwaang",
		Address:       "魔方公寓702",
		AddressDetail: "停车、楼栋等",
		DoorCode:      "临时密码",
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	pet := model.Pet{
		ShopID:     shopID,
		CustomerID: &customer.ID,
		Name:       "测试猫",
		Species:    "猫",
	}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	return customer, pet
}

func seedFeedingHoliday(t *testing.T, shopID uint, date string) {
	t.Helper()

	holiday := model.BoardingHoliday{
		ShopID:      shopID,
		HolidayDate: date,
		Name:        "节假日",
	}
	if err := database.DB.Create(&holiday).Error; err != nil {
		t.Fatalf("create holiday: %v", err)
	}
}
