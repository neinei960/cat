package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

var careReportOrderFixtureSeq uint64

func TestCreateOrderCareReportRejectsUnpaidOrder(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-CARE-REPORT-UNPAID",
		CustomerPhone:    "13800138201",
		CustomerNickname: "未支付护理客户",
		PetName:          "未支付猫咪",
		Paid:             false,
	})

	svc := NewOrderCareReportService(repository.NewOrderRepository(), repository.NewPetBathReportRepository())
	_, err := svc.Create(state.shopID, state.order.ID, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  "https://example.com/portrait.jpg",
		NextCareDate: "2026-04-25",
	})
	if err == nil {
		t.Fatalf("expected unpaid order to be rejected")
	}
	if err.Error() != "仅已支付订单可生成报告" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateOrderCareReportRejectsPetOutsideOrder(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-CARE-REPORT-PET",
		CustomerPhone:    "13800138202",
		CustomerNickname: "护理客户",
		PetName:          "订单内猫咪",
		Paid:             true,
	})
	outsidePet := seedCareReportPet(t, state.shopID, state.customer.ID, "订单外猫咪")

	svc := NewOrderCareReportService(repository.NewOrderRepository(), repository.NewPetBathReportRepository())
	_, err := svc.Create(state.shopID, state.order.ID, CreateOrderCareReportInput{
		PetID:        outsidePet.ID,
		PortraitURL:  "https://example.com/portrait.jpg",
		NextCareDate: "2026-04-25",
	})
	if err == nil {
		t.Fatalf("expected pet outside order to be rejected")
	}
	if err.Error() != "所选猫咪不属于当前订单" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateOrderCareReportCreatesBathReportAndRenderedImage(t *testing.T) {
	setupOrderServiceTestDB(t)
	if err := database.DB.AutoMigrate(&model.PetBathReport{}); err != nil {
		t.Fatalf("auto migrate pet bath report: %v", err)
	}

	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-CARE-REPORT-HAPPY",
		CustomerPhone:    "13800138203",
		CustomerNickname: "有效护理客户",
		PetName:          "有效猫咪",
		Paid:             true,
	})

	uploadDir := t.TempDir()
	originalUploadPath := config.AppConfig.Upload.Path
	config.AppConfig.Upload.Path = uploadDir
	t.Cleanup(func() {
		config.AppConfig.Upload.Path = originalUploadPath
	})

	portraitURL := writeTestPortraitUpload(t, uploadDir)

	svc := NewOrderCareReportService(repository.NewOrderRepository(), repository.NewPetBathReportRepository())
	result, err := svc.Create(state.shopID, state.order.ID, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  portraitURL,
		Weight:       "4.2kg",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-04-25",
		CareContent:  "护理记录",
		BodyShape:    "standard",
		Skin:         OrderCareReportSectionInput{Checks: []string{"normal", "scab"}, Note: "skin normal"},
		Hair:         OrderCareReportSectionInput{Checks: []string{"undercoat_many", "dry"}, Note: "hair dry"},
		Nails:        OrderCareReportSectionInput{Checks: []string{"trimmed", "pads_dry", "too_long"}, Note: "nails trimmed"},
		EyesFace:     OrderCareReportSectionInput{Checks: []string{"cleaned", "eye_red"}, Note: "eyes_face cleaned"},
		Ears:         OrderCareReportSectionInput{Checks: []string{"cleaned", "black_earwax"}, Note: "ears cleaned"},
		Oral:         OrderCareReportSectionInput{Checks: []string{"normal", "gum_red"}, Note: "oral normal"},
		Anus:         OrderCareReportSectionInput{Checks: []string{"normal", "anal_gland_swollen"}, Note: "anus normal"},
	})
	if err != nil {
		t.Fatalf("create care report: %v", err)
	}
	if result.ImageURL == "" {
		t.Fatalf("expected image url")
	}
	if result.ReportID == 0 {
		t.Fatalf("expected report id")
	}

	renderedPath := filepath.Join(uploadDir, filepath.Base(result.ImageURL))
	file, err := os.Open(renderedPath)
	if err != nil {
		t.Fatalf("open rendered image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode rendered image: %v", err)
	}
	if got := img.Bounds().Dx(); got != 1279 {
		t.Fatalf("expected width 1279, got %d", got)
	}
	if got := img.Bounds().Dy(); got != 1810 {
		t.Fatalf("expected height 1810, got %d", got)
	}
	baseImage := decodeOrderCareReportBaseImage(t)
	assertImageDiffersAt(t, baseImage, img, int(orderCareReportPortraitFrame.CenterX), int(orderCareReportPortraitFrame.CenterY), "portrait center")

	assertCanonicalContractCheck(t, "skin", "scab", "scab")
	assertCanonicalContractCheck(t, "hair", "undercoat_many", "undercoat_many")
	assertCanonicalContractCheck(t, "nails", "pads_dry", "pads_dry")
	assertCanonicalContractCheck(t, "nails", "too_long", "too_long")
	assertCanonicalContractCheck(t, "eyes_face", "eye_red", "eye_red")
	assertCanonicalContractCheck(t, "ears", "black_earwax", "black_earwax")
	assertCanonicalContractCheck(t, "oral", "gum_red", "gum_red")
	assertCanonicalContractCheck(t, "anus", "anal_gland_swollen", "anal_gland_swollen")

	var reports []model.PetBathReport
	if err := database.DB.Where("shop_id = ? AND pet_id = ?", state.shopID, state.pet.ID).Find(&reports).Error; err != nil {
		t.Fatalf("query bath reports: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 bath report, got %d", len(reports))
	}
	if reports[0].PetID != state.pet.ID {
		t.Fatalf("expected pet id %d, got %d", state.pet.ID, reports[0].PetID)
	}
	if reports[0].ImageURL != result.ImageURL {
		t.Fatalf("expected image url %q, got %q", result.ImageURL, reports[0].ImageURL)
	}
}

func TestRenderOrderCareReportDrawsFieldsOutsidePortraitArea(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-CARE-REPORT-RENDER",
		CustomerPhone:    "13800138205",
		CustomerNickname: "渲染护理客户",
		PetName:          "渲染猫咪",
		Paid:             true,
	})

	uploadDir := t.TempDir()
	originalUploadPath := config.AppConfig.Upload.Path
	config.AppConfig.Upload.Path = uploadDir
	t.Cleanup(func() {
		config.AppConfig.Upload.Path = originalUploadPath
	})

	portraitURL := writeTestPortraitUpload(t, uploadDir)
	bathDate := time.Date(2026, 4, 20, 11, 14, 17, 0, time.Local)

	rendered, err := renderOrderCareReport(&state.order, &state.pet, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  portraitURL,
		Weight:       "4.2kg",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-04-25",
		CareContent:  "护理记录",
		BodyShape:    "standard",
		Skin:         OrderCareReportSectionInput{Checks: []string{"normal"}, Note: "skin normal"},
	}, bathDate)
	if err != nil {
		t.Fatalf("render care report: %v", err)
	}

	baseImage := decodeOrderCareReportBaseImage(t)
	assertImageDiffersAt(t, baseImage, rendered, int(orderCareReportPortraitFrame.CenterX), int(orderCareReportPortraitFrame.CenterY), "portrait center")
	assertImageDiffersNear(t, baseImage, rendered, int(orderCareReportPrimaryFieldBoxes["care_date"].Left)+120, int(orderCareReportPrimaryFieldBoxes["care_date"].Baseline)-12, 32, "care date text")
	assertImageDiffersNear(t, baseImage, rendered, int(orderCareReportBodyShapeAnchors["standard"].X), int(orderCareReportBodyShapeAnchors["standard"].Y), 18, "body shape checkmark")
	assertImageDiffersNear(t, baseImage, rendered, int(orderCareReportSectionLayouts["skin"].Checkboxes["normal"].X), int(orderCareReportSectionLayouts["skin"].Checkboxes["normal"].Y), 18, "skin normal checkmark")
}

func TestCreateOrderCareReportDeletesRenderedImageWhenBathReportPersistenceFails(t *testing.T) {
	setupOrderServiceTestDB(t)

	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-CARE-REPORT-CLEANUP",
		CustomerPhone:    "13800138204",
		CustomerNickname: "清理护理客户",
		PetName:          "清理猫咪",
		Paid:             true,
	})

	uploadDir := t.TempDir()
	originalUploadPath := config.AppConfig.Upload.Path
	config.AppConfig.Upload.Path = uploadDir
	t.Cleanup(func() {
		config.AppConfig.Upload.Path = originalUploadPath
	})

	portraitURL := writeTestPortraitUpload(t, uploadDir)

	svc := NewOrderCareReportService(repository.NewOrderRepository(), repository.NewPetBathReportRepository())
	_, err := svc.Create(state.shopID, state.order.ID, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  portraitURL,
		Weight:       "4.2kg",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-04-25",
		CareContent:  "护理记录",
		BodyShape:    "standard",
	})
	if err == nil {
		t.Fatalf("expected persistence error when pet bath reports table is absent")
	}

	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		t.Fatalf("read upload dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected only portrait upload to remain, got %d files", len(entries))
	}
	if entries[0].Name() != filepath.Base(portraitURL) {
		t.Fatalf("expected remaining file %q, got %q", filepath.Base(portraitURL), entries[0].Name())
	}
}

type seedCareReportOrderFixtureInput struct {
	ShopID           uint
	OrderNo          string
	CustomerPhone    string
	CustomerNickname string
	PetName          string
	Paid             bool
}

type careReportOrderFixture struct {
	shopID   uint
	customer model.Customer
	pet      model.Pet
	order    model.Order
}

func seedCareReportOrderFixture(t *testing.T, input seedCareReportOrderFixtureInput) careReportOrderFixture {
	t.Helper()

	customer := model.Customer{
		ShopID:       input.ShopID,
		Phone:        input.CustomerPhone,
		Nickname:     input.CustomerNickname,
		DiscountRate: 1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create care report customer: %v", err)
	}

	pet := seedCareReportPet(t, input.ShopID, customer.ID, input.PetName)

	order := model.Order{
		OrderNo:      uniqueCareReportOrderNo(input.OrderNo),
		ShopID:       input.ShopID,
		CustomerID:   &customer.ID,
		PetID:        &pet.ID,
		TotalAmount:  88,
		ServiceTotal: 88,
		PayAmount:    88,
		PayMethod:    "wechat",
		Status:       0,
		PayStatus:    0,
	}
	if input.Paid {
		paidAt := time.Date(2026, 4, 20, 11, 14, 17, 0, time.Local)
		order.Status = 1
		order.PayStatus = 1
		order.PayTime = &paidAt
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create care report order: %v", err)
	}

	item := model.OrderItem{
		OrderID:   order.ID,
		ItemType:  1,
		ItemID:    101,
		Name:      "护理服务",
		Quantity:  1,
		UnitPrice: 88,
		Amount:    88,
	}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("create care report order item: %v", err)
	}

	return careReportOrderFixture{
		shopID:   input.ShopID,
		customer: customer,
		pet:      pet,
		order:    order,
	}
}

func uniqueCareReportOrderNo(base string) string {
	seq := atomic.AddUint64(&careReportOrderFixtureSeq, 1)
	return fmt.Sprintf("%s-%d", base, seq)
}

func seedCareReportPet(t *testing.T, shopID, customerID uint, name string) model.Pet {
	t.Helper()

	birthDate := time.Date(2024, 4, 20, 0, 0, 0, 0, time.Local)
	pet := model.Pet{
		ShopID:     shopID,
		CustomerID: &customerID,
		Name:       name,
		Species:    "猫",
		Breed:      "布偶",
		Gender:     2,
		BirthDate:  &birthDate,
	}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create care report pet: %v", err)
	}
	return pet
}

func writeTestPortraitUpload(t *testing.T, uploadDir string) string {
	t.Helper()

	filename := "test-portrait.jpg"
	path := filepath.Join(uploadDir, filename)
	img := image.NewRGBA(image.Rect(0, 0, 720, 960))
	for y := 0; y < 960; y++ {
		for x := 0; x < 720; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(40 + x%160),
				G: uint8(80 + y%120),
				B: 180,
				A: 255,
			})
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create portrait upload: %v", err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode portrait upload: %v", err)
	}

	return "/uploads/" + filename
}

func assertCanonicalContractCheck(t *testing.T, section, raw, want string) {
	t.Helper()

	got, ok := canonicalOrderCareReportSectionCheck(section, raw)
	if !ok {
		t.Fatalf("expected %s %q to be accepted", section, raw)
	}
	if got != want {
		t.Fatalf("expected %s %q => %q, got %q", section, raw, want, got)
	}
	point, exists := orderCareReportSectionLayouts[section].Checkboxes[want]
	if !exists {
		t.Fatalf("expected anchor for %s %q", section, want)
	}
	if point.X == 0 && point.Y == 0 {
		t.Fatalf("expected non-zero anchor for %s %q", section, want)
	}
}

func decodeOrderCareReportBaseImage(t *testing.T) image.Image {
	t.Helper()

	img, _, err := image.Decode(bytes.NewReader(orderCareReportBaseImage))
	if err != nil {
		t.Fatalf("decode base image: %v", err)
	}
	return img
}

func assertImageDiffersAt(t *testing.T, base image.Image, rendered image.Image, x, y int, label string) {
	t.Helper()

	baseBounds := base.Bounds()
	renderedBounds := rendered.Bounds()
	if !image.Pt(x, y).In(baseBounds) || !image.Pt(x, y).In(renderedBounds) {
		t.Fatalf("point %s out of bounds at %d,%d", label, x, y)
	}

	br, bg, bb, ba := base.At(x, y).RGBA()
	rr, rg, rb, ra := rendered.At(x, y).RGBA()
	if br == rr && bg == rg && bb == rb && ba == ra {
		t.Fatalf("expected rendered image to differ from base at %s (%d,%d)", label, x, y)
	}
}

func assertImageDiffersNear(t *testing.T, base image.Image, rendered image.Image, centerX, centerY, radius int, label string) {
	t.Helper()

	baseBounds := base.Bounds()
	renderedBounds := rendered.Bounds()
	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			if !image.Pt(x, y).In(baseBounds) || !image.Pt(x, y).In(renderedBounds) {
				continue
			}
			br, bg, bb, ba := base.At(x, y).RGBA()
			rr, rg, rb, ra := rendered.At(x, y).RGBA()
			if br != rr || bg != rg || bb != rb || ba != ra {
				return
			}
		}
	}

	t.Fatalf("expected rendered image to differ from base near %s around (%d,%d) within radius %d", label, centerX, centerY, radius)
}
