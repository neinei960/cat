package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// AutoMigrate
	database.AutoMigrate(
		&model.Shop{}, &model.Staff{}, &model.Customer{}, &model.CustomerTag{}, &model.CustomerTagRelation{}, &model.Pet{},
		&model.Service{}, &model.ServicePriceRule{}, &model.ServiceAddon{}, &model.FurCategory{},
		&model.StaffService{}, &model.StaffSchedule{},
		&model.Appointment{}, &model.AppointmentService{},
		&model.AppointmentPet{}, &model.AppointmentPetService{},
		&model.Order{}, &model.OrderItem{}, &model.NotificationLog{}, &model.DailyStats{},
	)

	// 1. 创建店铺
	shop := model.Shop{Name: "猫咪洗护工作室", Phone: "010-88886666", Address: "北京市朝阳区猫咪路1号", Status: 1}
	database.DB.FirstOrCreate(&shop, model.Shop{Name: "猫咪洗护工作室"})
	fmt.Printf("Shop ID: %d\n", shop.ID)

	// 2. 创建管理员
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	admin := model.Staff{
		ShopID: shop.ID, Phone: "13800138000", PasswordHash: string(hash),
		Name: "店长", Role: "admin", Status: 1, CommissionRate: 0,
	}
	database.DB.Where("phone = ?", "13800138000").FirstOrCreate(&admin)
	fmt.Printf("Admin Staff ID: %d\n", admin.ID)

	// 3. 创建3位洗护师（不同提成比例）
	groomer1 := model.Staff{
		ShopID: shop.ID, Phone: "13800138001", PasswordHash: string(hash),
		Name: "乐乐", Role: "staff", Status: 1, CommissionRate: 30, // 0.30
	}
	database.DB.Where("phone = ?", "13800138001").FirstOrCreate(&groomer1)

	groomer2 := model.Staff{
		ShopID: shop.ID, Phone: "13800138002", PasswordHash: string(hash),
		Name: "小美", Role: "staff", Status: 1, CommissionRate: 25, // 0.25
	}
	database.DB.Where("phone = ?", "13800138002").FirstOrCreate(&groomer2)

	groomer3 := model.Staff{
		ShopID: shop.ID, Phone: "13800138003", PasswordHash: string(hash),
		Name: "阿花", Role: "staff", Status: 1, CommissionRate: 20, // 0.20
	}
	database.DB.Where("phone = ?", "13800138003").FirstOrCreate(&groomer3)
	fmt.Printf("Groomers: %d, %d, %d\n", groomer1.ID, groomer2.ID, groomer3.ID)

	// 4. 创建洗浴服务项目
	svc1 := model.Service{ShopID: shop.ID, Name: "日常皮毛护理", Category: "洗澡", BasePrice: 88, Duration: 60, Status: 1, SortOrder: 1, Description: "适用于短毛猫/长毛猫的日常清洁"}
	svc2 := model.Service{ShopID: shop.ID, Name: "日常皮毛调理", Category: "洗澡", BasePrice: 107, Duration: 75, Status: 1, SortOrder: 2, Description: "适用于A/B/C/D类猫的日常调理"}
	svc3 := model.Service{ShopID: shop.ID, Name: "深层清洁护理", Category: "洗澡", BasePrice: 108, Duration: 90, Status: 1, SortOrder: 3, Description: "适用于A/B/C/D类猫的深层清洁"}
	svc4 := model.Service{ShopID: shop.ID, Name: "精致皮毛调理", Category: "SPA", BasePrice: 198, Duration: 120, Status: 1, SortOrder: 4, Description: "高端皮毛护理，适用于A/B/C/D类"}
	database.DB.Where("name = ? AND shop_id = ?", svc1.Name, shop.ID).FirstOrCreate(&svc1)
	database.DB.Where("name = ? AND shop_id = ?", svc2.Name, shop.ID).FirstOrCreate(&svc2)
	database.DB.Where("name = ? AND shop_id = ?", svc3.Name, shop.ID).FirstOrCreate(&svc3)
	database.DB.Where("name = ? AND shop_id = ?", svc4.Name, shop.ID).FirstOrCreate(&svc4)
	fmt.Printf("Services: %d, %d, %d, %d\n", svc1.ID, svc2.ID, svc3.ID, svc4.ID)

	// 5. 创建定价规则（项目 × 毛发等级矩阵）
	priceRules := []model.ServicePriceRule{
		// 日常皮毛护理 - 短毛猫/长毛猫
		{ServiceID: svc1.ID, FurLevel: "短毛猫", Price: 88, Duration: 60},
		{ServiceID: svc1.ID, FurLevel: "长毛猫", Price: 108, Duration: 75},
		// 日常皮毛调理 - A/B/C/D
		{ServiceID: svc2.ID, FurLevel: "A", Price: 107, Duration: 60},
		{ServiceID: svc2.ID, FurLevel: "B", Price: 138, Duration: 75},
		{ServiceID: svc2.ID, FurLevel: "C", Price: 168, Duration: 90},
		{ServiceID: svc2.ID, FurLevel: "D", Price: 188, Duration: 105},
		// 深层清洁护理 - A/B/C/D
		{ServiceID: svc3.ID, FurLevel: "A", Price: 108, Duration: 75},
		{ServiceID: svc3.ID, FurLevel: "B", Price: 158, Duration: 90},
		{ServiceID: svc3.ID, FurLevel: "C", Price: 188, Duration: 105},
		{ServiceID: svc3.ID, FurLevel: "D", Price: 208, Duration: 120},
		// 精致皮毛调理 - A/B/C/D
		{ServiceID: svc4.ID, FurLevel: "A", Price: 198, Duration: 90},
		{ServiceID: svc4.ID, FurLevel: "B", Price: 238, Duration: 105},
		{ServiceID: svc4.ID, FurLevel: "C", Price: 288, Duration: 120},
		{ServiceID: svc4.ID, FurLevel: "D", Price: 328, Duration: 150},
	}
	for _, rule := range priceRules {
		database.DB.Where("service_id = ? AND fur_level = ?", rule.ServiceID, rule.FurLevel).FirstOrCreate(&rule)
	}
	fmt.Println("Price rules created")

	// 6. 创建毛发类别
	furCategories := []model.FurCategory{
		{ShopID: shop.ID, Name: "短毛猫", SortOrder: 1, Status: 1},
		{ShopID: shop.ID, Name: "长毛猫", SortOrder: 2, Status: 1},
		{ShopID: shop.ID, Name: "A", SortOrder: 3, Status: 1},
		{ShopID: shop.ID, Name: "B", SortOrder: 4, Status: 1},
		{ShopID: shop.ID, Name: "C", SortOrder: 5, Status: 1},
		{ShopID: shop.ID, Name: "D", SortOrder: 6, Status: 1},
	}
	for _, fc := range furCategories {
		database.DB.Where("name = ? AND shop_id = ?", fc.Name, fc.ShopID).FirstOrCreate(&fc)
	}
	fmt.Println("Fur categories created")

	// 7. 创建附加费项目
	addons := []model.ServiceAddon{
		{ShopID: shop.ID, Name: "超重费", DefaultPrice: 10, IsVariable: true, SortOrder: 1, Status: 1},
		{ShopID: shop.ID, Name: "去油费", DefaultPrice: 20, IsVariable: true, SortOrder: 2, Status: 1},
		{ShopID: shop.ID, Name: "药浴", DefaultPrice: 30, IsVariable: true, SortOrder: 3, Status: 1},
		{ShopID: shop.ID, Name: "刷牙", DefaultPrice: 20, IsVariable: false, SortOrder: 4, Status: 1},
		{ShopID: shop.ID, Name: "开结", DefaultPrice: 0, IsVariable: true, SortOrder: 5, Status: 1},
		{ShopID: shop.ID, Name: "攻击费", DefaultPrice: 20, IsVariable: true, SortOrder: 6, Status: 1},
		{ShopID: shop.ID, Name: "春节加收", DefaultPrice: 50, IsVariable: true, SortOrder: 7, Status: 1},
	}
	for _, addon := range addons {
		database.DB.Where("name = ? AND shop_id = ?", addon.Name, addon.ShopID).FirstOrCreate(&addon)
	}
	fmt.Println("Service addons created")

	// 7. 分配服务给洗护师
	for _, staffID := range []uint{groomer1.ID, groomer2.ID, groomer3.ID} {
		for _, svcID := range []uint{svc1.ID, svc2.ID, svc3.ID, svc4.ID} {
			ss := model.StaffService{StaffID: staffID, ServiceID: svcID}
			database.DB.Where("staff_id = ? AND service_id = ?", staffID, svcID).FirstOrCreate(&ss)
		}
	}

	// 8. 创建洗护师排班（未来7天）
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, i).Format("2006-01-02")
		for _, staffID := range []uint{groomer1.ID, groomer2.ID, groomer3.ID} {
			sched := model.StaffSchedule{
				StaffID: staffID, ShopID: shop.ID, Date: date,
				StartTime: "12:00", EndTime: "22:00",
				BreakStart: "17:00", BreakEnd: "18:00",
				MaxCapacity: 1, IsDayOff: false,
			}
			database.DB.Where("staff_id = ? AND date = ?", staffID, date).FirstOrCreate(&sched)
		}
	}
	fmt.Println("Schedules created for 7 days")

	// 9. 创建测试会员客户
	member1 := model.Customer{
		ShopID: shop.ID, Nickname: "王小姐", Phone: "13900139000",
		Gender: 2, Tags: "会员", MemberBalance: 1000, DiscountRate: 0.9,
	}
	database.DB.Where("phone = ? AND shop_id = ?", "13900139000", shop.ID).FirstOrCreate(&member1)

	member2 := model.Customer{
		ShopID: shop.ID, Nickname: "李先生", Phone: "13900139001",
		Gender: 1, Tags: "VIP会员", MemberBalance: 3000, DiscountRate: 0.86,
	}
	database.DB.Where("phone = ? AND shop_id = ?", "13900139001", shop.ID).FirstOrCreate(&member2)
	fmt.Printf("Members: %d, %d\n", member1.ID, member2.ID)

	// 10. 创建测试猫咪
	custID1 := member1.ID
	custID2 := member2.ID
	cat1 := model.Pet{
		ShopID: shop.ID, CustomerID: &custID1,
		Name: "月光", Species: "猫", Breed: "英短蓝猫", Gender: 1,
		Weight: 5.5, FurLevel: "短毛猫", Personality: "胆小敏感",
		Aggression: "无", BathFrequency: "两月",
		CareNotes: "胆子很小，需要轻声安抚", Status: 1,
	}
	database.DB.Where("name = ? AND shop_id = ?", "月光", shop.ID).FirstOrCreate(&cat1)

	cat2 := model.Pet{
		ShopID: shop.ID, CustomerID: &custID2,
		Name: "大橘", Species: "猫", Breed: "橘猫", Gender: 1,
		Weight: 7.2, FurLevel: "B", Personality: "神仙宝贝",
		Aggression: "无", BathFrequency: "每月",
		CareNotes: "超重，洗完要吹很久", Status: 1,
	}
	database.DB.Where("name = ? AND shop_id = ?", "大橘", shop.ID).FirstOrCreate(&cat2)

	cat3 := model.Pet{
		ShopID: shop.ID, CustomerID: &custID1,
		Name: "小白", Species: "猫", Breed: "布偶", Gender: 2,
		Weight: 4.8, FurLevel: "C", Personality: "笑里藏刀",
		Aggression: "可能", ForbiddenZones: "肚子、后腿",
		BathFrequency: "两月",
		CareNotes:     "看起来乖但可能突然咬人，注意禁区", Status: 1,
	}
	database.DB.Where("name = ? AND shop_id = ?", "小白", shop.ID).FirstOrCreate(&cat3)
	fmt.Printf("Cats: %d, %d, %d\n", cat1.ID, cat2.ID, cat3.ID)

	fmt.Println("\n=== Seed completed ===")
	fmt.Println("Admin login: 13800138000 / 123456")
	fmt.Println("\n定价矩阵:")
	fmt.Println("日常皮毛护理: 短毛猫¥88, 长毛猫¥108")
	fmt.Println("日常皮毛调理: A¥107, B¥138, C¥168, D¥188")
	fmt.Println("深层清洁护理: A¥108, B¥158, C¥188, D¥208")
	fmt.Println("精致皮毛调理: A¥198, B¥238, C¥288, D¥328")
	fmt.Println("\n会员折扣:")
	fmt.Println("充值¥1000 → 9折 (王小姐)")
	fmt.Println("充值¥3000 → 8.6折 (李先生)")
}
