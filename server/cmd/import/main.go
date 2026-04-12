package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/utils"
	"github.com/xuri/excelize/v2"
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
		&model.Appointment{}, &model.AppointmentCalendarMark{}, &model.AppointmentService{},
		&model.AppointmentPet{}, &model.AppointmentPetService{},
		&model.Order{}, &model.OrderItem{},
		&model.FeedingSetting{}, &model.FeedingPlan{}, &model.FeedingPlanPet{}, &model.FeedingPlanRule{},
		&model.FeedingVisit{}, &model.FeedingVisitItem{}, &model.FeedingVisitLog{}, &model.FeedingVisitMedia{},
		&model.BoardingCabinet{}, &model.BoardingHoliday{}, &model.BoardingDiscountPolicy{},
		&model.BoardingOrder{}, &model.BoardingOrderRoom{}, &model.BoardingOrderPet{}, &model.BoardingOrderLog{},
		&model.NotificationLog{}, &model.DailyStats{},
	)

	// Get shop (must exist from seed)
	var shop model.Shop
	if err := database.DB.First(&shop).Error; err != nil {
		log.Fatalf("No shop found. Run seed first: %v", err)
	}

	// Open Excel file
	excelPath := "猫咪大全最新2.26.xlsx"
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		log.Fatalf("Failed to open Excel: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	fmt.Printf("Sheets: %v\n", sheets)

	// Build staff map
	staffMap := map[string]uint{}
	var staffList []model.Staff
	database.DB.Where("shop_id = ? AND role = 'staff'", shop.ID).Find(&staffList)
	for _, s := range staffList {
		staffMap[s.Name] = s.ID
	}

	// Build service map
	serviceMap := map[string]uint{}
	var serviceList []model.Service
	database.DB.Where("shop_id = ?", shop.ID).Find(&serviceList)
	for _, s := range serviceList {
		serviceMap[s.Name] = s.ID
	}

	// Import cat profiles from "洗浴猫咪档案" sheet
	importCatProfiles(f, shop.ID)

	// Import grooming records from groomer sheets (e.g., "乐乐记录", "小美记录", etc.)
	for _, sheet := range sheets {
		if strings.Contains(sheet, "记录") {
			importGroomingRecords(f, sheet, shop.ID, staffMap, serviceMap)
		}
	}

	// Print summary
	var petCount, orderCount int64
	database.DB.Model(&model.Pet{}).Where("shop_id = ?", shop.ID).Count(&petCount)
	database.DB.Model(&model.Order{}).Where("shop_id = ?", shop.ID).Count(&orderCount)
	fmt.Printf("\n=== Import Summary ===\n")
	fmt.Printf("Total cats: %d\n", petCount)
	fmt.Printf("Total orders: %d\n", orderCount)
}

func importCatProfiles(f *excelize.File, shopID uint) {
	sheetName := ""
	for _, s := range f.GetSheetList() {
		if strings.Contains(s, "档案") || strings.Contains(s, "猫咪") {
			sheetName = s
			break
		}
	}
	if sheetName == "" {
		fmt.Println("Warning: 未找到猫咪档案Sheet")
		return
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("Error reading sheet %s: %v\n", sheetName, err)
		return
	}

	if len(rows) < 2 {
		fmt.Println("猫咪档案Sheet数据为空")
		return
	}

	// Parse header to find column indices
	header := rows[0]
	colMap := map[string]int{}
	for i, h := range header {
		h = strings.TrimSpace(h)
		colMap[h] = i
	}

	imported := 0
	for _, row := range rows[1:] {
		name := getCell(row, colMap, "猫咪名", "名字", "猫名")
		if name == "" {
			continue
		}

		pet := model.Pet{
			ShopID:  shopID,
			Name:    name,
			Species: "猫",
			Status:  1,
		}

		if v := getCell(row, colMap, "品种"); v != "" {
			pet.Breed = v
		}
		if v := getCell(row, colMap, "性格", "性格特征"); v != "" {
			pet.Personality = v
		}
		if v := getCell(row, colMap, "攻击性"); v != "" {
			pet.Aggression = v
		}
		if v := getCell(row, colMap, "禁区"); v != "" {
			pet.ForbiddenZones = v
		}
		if v := getCell(row, colMap, "频率", "洗澡频率"); v != "" {
			pet.BathFrequency = v
		}
		if v := getCell(row, colMap, "毛发", "毛发类型", "毛发等级"); v != "" {
			pet.FurLevel = normalizeFurLevel(v)
		}
		if v := getCell(row, colMap, "体重"); v != "" {
			if w, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				pet.Weight = w
			}
		}
		if v := getCell(row, colMap, "注意事项", "备注"); v != "" {
			pet.CareNotes = v
		}

		database.DB.Where("name = ? AND shop_id = ?", pet.Name, shopID).FirstOrCreate(&pet)
		imported++
	}
	fmt.Printf("Imported %d cat profiles from '%s'\n", imported, sheetName)
}

func importGroomingRecords(f *excelize.File, sheetName string, shopID uint, staffMap map[string]uint, serviceMap map[string]uint) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		fmt.Printf("Error reading sheet %s: %v\n", sheetName, err)
		return
	}

	if len(rows) < 2 {
		return
	}

	header := rows[0]
	colMap := map[string]int{}
	for i, h := range header {
		h = strings.TrimSpace(h)
		colMap[h] = i
	}

	// Determine groomer from sheet name
	groomerName := strings.TrimSuffix(sheetName, "记录")
	groomerName = strings.TrimSpace(groomerName)
	staffID := staffMap[groomerName]

	imported := 0
	for _, row := range rows[1:] {
		catName := getCell(row, colMap, "猫咪", "猫咪名", "名字")
		if catName == "" {
			continue
		}

		// Find or create pet
		var pet model.Pet
		if err := database.DB.Where("name = ? AND shop_id = ?", catName, shopID).First(&pet).Error; err != nil {
			pet = model.Pet{ShopID: shopID, Name: catName, Species: "猫", Status: 1}
			database.DB.Create(&pet)
		}

		// Parse date
		dateStr := getCell(row, colMap, "日期")
		orderDate := parseDate(dateStr)

		// Parse service
		serviceName := getCell(row, colMap, "项目", "洗浴项目")
		serviceID := serviceMap[serviceName]

		// Parse amounts
		basePrice := parseFloat(getCell(row, colMap, "基础价", "基础", "价格"))
		overweight := parseFloat(getCell(row, colMap, "超重", "超重费"))
		degrease := parseFloat(getCell(row, colMap, "去油", "去油费"))
		medBath := parseFloat(getCell(row, colMap, "药浴"))
		brush := parseFloat(getCell(row, colMap, "刷牙"))
		detangle := parseFloat(getCell(row, colMap, "开结", "开结费"))
		actual := parseFloat(getCell(row, colMap, "实付", "实付金额", "实收"))

		totalAmount := basePrice + overweight + degrease + medBath + brush + detangle
		if totalAmount == 0 && actual > 0 {
			totalAmount = actual
		}
		payAmount := actual
		if payAmount == 0 {
			payAmount = totalAmount
		}

		var staffIDPtr *uint
		if staffID > 0 {
			staffIDPtr = &staffID
		}
		petID := pet.ID

		order := model.Order{
			OrderNo:        utils.GenerateOrderNo(),
			ShopID:         shopID,
			PetID:          &petID,
			CustomerID:     pet.CustomerID,
			StaffID:        staffIDPtr,
			TotalAmount:    totalAmount,
			DiscountAmount: totalAmount - payAmount,
			PayAmount:      payAmount,
			PayMethod:      "cash",
			PayStatus:      1,
			Status:         1,
		}

		if !orderDate.IsZero() {
			order.PayTime = &orderDate
		}

		tx := database.DB.Begin()
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			continue
		}

		// Create items
		var items []model.OrderItem
		if basePrice > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 1, ItemID: serviceID,
				Name: serviceName, Quantity: 1, UnitPrice: basePrice, Amount: basePrice,
			})
		}
		if overweight > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 3, Name: "超重费",
				Quantity: 1, UnitPrice: overweight, Amount: overweight,
			})
		}
		if degrease > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 3, Name: "去油费",
				Quantity: 1, UnitPrice: degrease, Amount: degrease,
			})
		}
		if medBath > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 3, Name: "药浴",
				Quantity: 1, UnitPrice: medBath, Amount: medBath,
			})
		}
		if brush > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 3, Name: "刷牙",
				Quantity: 1, UnitPrice: brush, Amount: brush,
			})
		}
		if detangle > 0 {
			items = append(items, model.OrderItem{
				OrderID: order.ID, ItemType: 3, Name: "开结",
				Quantity: 1, UnitPrice: detangle, Amount: detangle,
			})
		}

		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				tx.Rollback()
				continue
			}
		}

		// Set created_at to order date
		if !orderDate.IsZero() {
			tx.Model(&order).UpdateColumn("created_at", orderDate)
		}

		tx.Commit()
		imported++
	}
	fmt.Printf("Imported %d records from '%s' (groomer: %s)\n", imported, sheetName, groomerName)
}

// getCell finds first matching column header and returns the value
func getCell(row []string, colMap map[string]int, names ...string) string {
	for _, name := range names {
		if idx, ok := colMap[name]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
	}
	return ""
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "¥", "")
	s = strings.ReplaceAll(s, ",", "")
	if s == "" || s == "-" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	// Try common formats
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006.01.02",
		"01-02-06",
		"1/2/2006",
		"2006-1-2",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	// Try Excel serial date (days since 1900-01-01)
	if v, err := strconv.ParseFloat(s, 64); err == nil && v > 40000 && v < 50000 {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		return base.AddDate(0, 0, int(v))
	}
	return time.Time{}
}

func normalizeFurLevel(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "短毛"):
		return "短毛猫"
	case strings.Contains(s, "长毛"):
		return "长毛猫"
	case strings.HasPrefix(s, "A") || s == "a":
		return "A"
	case strings.HasPrefix(s, "B") || s == "b":
		return "B"
	case strings.HasPrefix(s, "C") || s == "c":
		return "C"
	case strings.HasPrefix(s, "D") || s == "d":
		return "D"
	default:
		return s
	}
}
