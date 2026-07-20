package service

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fogleman/gg"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
)

var careReportOrderFixtureSeq uint64

func TestBuildOrderCareReportDrawDataUsesDisplayOverrides(t *testing.T) {
	pet := &model.Pet{Name: "数据库名字", Breed: "数据库品种", Gender: 1}
	petName := "报告名字"
	breed := "报告品种"
	gender := "MM"
	age := "2岁1月"

	data := buildOrderCareReportDrawData(pet, CreateOrderCareReportInput{
		PetName: &petName,
		Breed:   &breed,
		Gender:  &gender,
		Age:     &age,
	}, time.Date(2026, 7, 16, 0, 0, 0, 0, time.Local))

	if data.PetName != petName || data.Breed != breed || data.Gender != gender || data.Age != age {
		t.Fatalf("unexpected display data: %+v", data)
	}
	if pet.Name != "数据库名字" || pet.Breed != "数据库品种" || pet.Gender != 1 {
		t.Fatalf("report overrides must not mutate pet: %+v", pet)
	}
}

func TestOrderCareReportDisplayDateUsesRealTemplateFormat(t *testing.T) {
	if got := formatOrderCareReportDisplayDate("2026-07-03"); got != "2026.7.3" {
		t.Fatalf("want 2026.7.3, got %q", got)
	}
	if got := formatOrderCareReportDisplayDate("not-a-date"); got != "not-a-date" {
		t.Fatalf("invalid date should remain readable, got %q", got)
	}
}

func TestOrderCareReportBoldFontLoads(t *testing.T) {
	if _, err := orderCareReportFontFace(28, orderCareReportFontBold); err != nil {
		t.Fatalf("load bold care report font: %v", err)
	}
}

func TestOrderCareReportPrimaryFieldWeightsMatchRealTemplate(t *testing.T) {
	if got := orderCareReportPrimaryFieldWeight("care_date"); got != orderCareReportFontBold {
		t.Fatalf("care date should be bold, got %v", got)
	}
	if got := orderCareReportPrimaryFieldWeight("next_care_date"); got != orderCareReportFontRegular {
		t.Fatalf("next care date should use the lighter real-template weight, got %v", got)
	}
}

func TestOrderCareReportTemplateOverridesCareContentLabel(t *testing.T) {
	baseImage := decodeOrderCareReportBaseImage(t)
	renderBase := decodeOrderCareReportBaseImage(t)
	dc := gg.NewContextForImage(renderBase)
	if err := drawOrderCareReportTemplateOverrides(dc); err != nil {
		t.Fatalf("draw care report template overrides: %v", err)
	}

	_, _, found := changedPixelXBounds(baseImage, dc.Image(), image.Rect(90, 560, 270, 650))
	if !found {
		t.Fatalf("expected care-content label area to differ from the old template")
	}
}

func TestOrderCareReportLongNoteWrapsToTwoLines(t *testing.T) {
	dc := gg.NewContext(1279, 1810)
	lines, size, err := layoutOrderCareReportNote(
		dc,
		"重度废毛 局部轻微打结 全身毛发黏腻感明显 较多跳蚤尸体 建议每月定期体内外驱虫",
		orderCareReportSectionLayouts["hair"].NoteBox,
		80,
	)
	if err != nil {
		t.Fatalf("layout long note: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected two note lines, got %d: %#v", len(lines), lines)
	}
	if size < 18 || size > 24 {
		t.Fatalf("unexpected note font size %.1f", size)
	}
}

func TestOrderCareReportLayoutMatchesTemplateRows(t *testing.T) {
	wantFieldBaselines := map[string]float64{
		"pet_name":       257,
		"breed":          389,
		"gender":         511,
		"age":            511,
		"care_content":   625,
		"care_date":      730,
		"next_care_date": 730,
		"weight":         843,
	}
	for fieldName, wantBaseline := range wantFieldBaselines {
		if got := orderCareReportPrimaryFieldBoxes[fieldName].Baseline; got != wantBaseline {
			t.Fatalf("%s baseline should stay above its underline: want %.0f, got %.0f", fieldName, wantBaseline, got)
		}
	}
	if got := orderCareReportSectionLayouts["skin"].NoteBox.Left; got != 648 {
		t.Fatalf("skin note should start after the built-in label, got %.0f", got)
	}
	if got := orderCareReportSectionLayouts["ears"].NoteBox.Left; got != 648 {
		t.Fatalf("ears note should start after the built-in label, got %.0f", got)
	}
	if got := orderCareReportSectionLayouts["anus"].Checkboxes["normal"].Y; got != 1553 {
		t.Fatalf("anus checkbox should align with its option row, got %.0f", got)
	}
	if got := orderCareReportSectionLayouts["anus"].NoteBox.Baseline; got != 1600 {
		t.Fatalf("anus note should align with its note row, got %.0f", got)
	}
	wantNoteBaselines := map[string]float64{
		"skin":      975,
		"hair":      1071,
		"nails":     1168,
		"eyes_face": 1264,
		"ears":      1359,
		"oral":      1503,
		"anus":      1600,
	}
	for sectionName, wantBaseline := range wantNoteBaselines {
		if got := orderCareReportSectionLayouts[sectionName].NoteBox.Baseline; got != wantBaseline {
			t.Fatalf("%s note baseline should stay above its underline: want %.0f, got %.0f", sectionName, wantBaseline, got)
		}
	}

	wantColumnCenters := map[string]float64{
		"thin":     406,
		"skinny":   569,
		"standard": 732,
		"chubby":   895,
		"obese":    1058,
	}
	for key, wantX := range wantColumnCenters {
		if got := orderCareReportBodyShapeAnchors[key].X; got != wantX {
			t.Fatalf("%s checkbox center should match the template: want %.0f, got %.0f", key, wantX, got)
		}
	}

	wantSectionCheckCenters := map[string]map[string]orderCareReportPoint{
		"skin": {
			"normal": {X: 406, Y: 929}, "dandruff": {X: 569, Y: 929}, "red": {X: 732, Y: 929},
			"greasy": {X: 895, Y: 929}, "scab": {X: 1058, Y: 929}, "wound": {X: 406, Y: 977},
		},
		"hair": {
			"shedding": {X: 406, Y: 1025}, "undercoat_many": {X: 569, Y: 1025}, "dry": {X: 732, Y: 1025},
			"greasy": {X: 895, Y: 1025}, "matting": {X: 1058, Y: 1025},
		},
		"nails": {
			"trimmed": {X: 406, Y: 1121}, "dewclaw_abnormal": {X: 569, Y: 1121}, "pads_dry": {X: 732, Y: 1121},
			"too_long": {X: 895, Y: 1121}, "wound": {X: 1058, Y: 1121},
		},
		"eyes_face": {
			"cleaned": {X: 406, Y: 1217}, "tear_many": {X: 569, Y: 1217}, "eye_red": {X: 732, Y: 1217},
			"eye_abnormal": {X: 895, Y: 1217}, "wound": {X: 1058, Y: 1217},
		},
		"ears": {
			"cleaned": {X: 406, Y: 1313}, "touch_sensitive": {X: 569, Y: 1313}, "inflamed": {X: 732, Y: 1313},
			"earwax": {X: 895, Y: 1313}, "black_earwax": {X: 1058, Y: 1313}, "wound": {X: 406, Y: 1361},
		},
		"oral": {
			"normal": {X: 406, Y: 1410}, "touch_sensitive": {X: 569, Y: 1410}, "tartar": {X: 732, Y: 1410},
			"gum_red": {X: 895, Y: 1410}, "gum_swollen": {X: 1058, Y: 1410}, "oral_ulcer": {X: 406, Y: 1458},
			"bad_breath": {X: 569, Y: 1458}, "dental_abnormal": {X: 732, Y: 1458},
		},
		"anus": {
			"normal": {X: 406, Y: 1553}, "prolapse": {X: 569, Y: 1553}, "red": {X: 732, Y: 1553}, "inflamed": {X: 895, Y: 1553},
		},
	}
	for sectionName, wantChecks := range wantSectionCheckCenters {
		gotChecks := orderCareReportSectionLayouts[sectionName].Checkboxes
		if len(gotChecks) != len(wantChecks) {
			t.Fatalf("%s checkbox count should match template: want %d, got %d", sectionName, len(wantChecks), len(gotChecks))
		}
		for checkName, wantPoint := range wantChecks {
			if got := gotChecks[checkName]; got != wantPoint {
				t.Fatalf("%s.%s checkbox center should match template: want %+v, got %+v", sectionName, checkName, wantPoint, got)
			}
		}
	}
}

func TestOrderCareReportCheckmarkGeometryFitsCheckbox(t *testing.T) {
	if orderCareReportCheckmarkStrokeWidth < 5 || orderCareReportCheckmarkStrokeWidth > 6.5 {
		t.Fatalf("reference checkmark stroke should be bold, got %.1f", orderCareReportCheckmarkStrokeWidth)
	}
	if orderCareReportCheckmarkStartOffset.X > -9 || orderCareReportCheckmarkEndOffset.X < 12 {
		t.Fatalf("reference checkmark should extend past the box horizontally: start=%+v end=%+v", orderCareReportCheckmarkStartOffset, orderCareReportCheckmarkEndOffset)
	}
	if orderCareReportCheckmarkKneeOffset.Y < 6 || orderCareReportCheckmarkEndOffset.Y > -11 {
		t.Fatalf("reference checkmark should use the authentic tall path: knee=%+v end=%+v", orderCareReportCheckmarkKneeOffset, orderCareReportCheckmarkEndOffset)
	}
}

func TestOrderCareReportSectionNotesAreCenteredOnTheirLines(t *testing.T) {
	baseImage := decodeOrderCareReportBaseImage(t)
	renderBase := decodeOrderCareReportBaseImage(t)
	dc := gg.NewContextForImage(renderBase)
	dc.SetRGB(0, 0, 0)
	drawOrderCareReportSections(dc, CreateOrderCareReportInput{
		Skin: OrderCareReportSectionInput{Note: "居中"},
	})
	rendered := dc.Image()

	box := orderCareReportSectionLayouts["skin"].NoteBox
	minX, maxX, found := changedPixelXBounds(baseImage, rendered, image.Rect(
		int(box.Left),
		int(box.Baseline)-40,
		int(box.Right),
		int(box.Baseline)+6,
	))
	if !found {
		t.Fatalf("expected centered skin note to be rendered")
	}
	wantCenter := (box.Left + box.Right) / 2
	gotCenter := float64(minX+maxX) / 2
	if math.Abs(gotCenter-wantCenter) > 12 {
		t.Fatalf("skin note should be centered on its line: want center %.1f, got %.1f", wantCenter, gotCenter)
	}
}

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
	assertCanonicalContractCheck(t, "skin", "wound", "wound")
	assertCanonicalContractCheck(t, "ears", "wound", "wound")
	assertCanonicalContractCheck(t, "oral", "bad_breath", "bad_breath")
	assertCanonicalContractCheck(t, "anus", "prolapse", "prolapse")
	assertCanonicalContractCheck(t, "anus", "anal_gland_swollen", "inflamed")

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
	bathDate := time.Date(2026, 7, 3, 11, 14, 17, 0, time.Local)
	petName := "money"
	breed := "海双布偶"
	gender := "GG"
	age := "2岁"

	rendered, err := renderOrderCareReport(&state.order, &state.pet, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PetName:      &petName,
		Breed:        &breed,
		Gender:       &gender,
		Age:          &age,
		PortraitURL:  portraitURL,
		Weight:       "5.58kg",
		CareDate:     "2026-07-03",
		NextCareDate: "2026-09-01",
		CareContent:  "伊珊娜深层清洁护理",
		BodyShape:    "standard",
		Skin:         OrderCareReportSectionInput{Checks: []string{"normal", "dandruff", "red", "greasy"}, Note: "皮脂腺重度油脂堆积 后背尾巴有较多油脂分泌物"},
		Hair:         OrderCareReportSectionInput{Checks: []string{"shedding", "undercoat_many", "dry", "greasy", "matting"}, Note: "重度废毛 局部轻微打结 全身毛发黏腻感明显 较多跳蚤尸体 建议每月定期体内外驱虫"},
		Nails:        OrderCareReportSectionInput{Checks: []string{"trimmed"}, Note: "指甲已剪"},
		EyesFace:     OrderCareReportSectionInput{Checks: []string{"cleaned"}, Note: "眼睛已清洁"},
		Ears:         OrderCareReportSectionInput{Checks: []string{"cleaned"}, Note: "耳朵已清洁"},
		Oral:         OrderCareReportSectionInput{Checks: []string{"normal"}, Note: "尽早培养刷牙习惯 较多牙渍牙菌斑 长期间不干预会逐渐形成牙结石"},
		Anus:         OrderCareReportSectionInput{Checks: []string{"normal"}, Note: "肛周正常"},
	}, bathDate)
	if err != nil {
		t.Fatalf("render care report: %v", err)
	}
	if artifactPath := os.Getenv("CARE_REPORT_ARTIFACT_PATH"); artifactPath != "" {
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
			t.Fatalf("create artifact directory: %v", err)
		}
		artifact, err := os.Create(artifactPath)
		if err != nil {
			t.Fatalf("create rendered artifact: %v", err)
		}
		if err := jpeg.Encode(artifact, rendered, &jpeg.Options{Quality: 92}); err != nil {
			_ = artifact.Close()
			t.Fatalf("encode rendered artifact: %v", err)
		}
		if err := artifact.Close(); err != nil {
			t.Fatalf("close rendered artifact: %v", err)
		}
	}

	baseImage := decodeOrderCareReportBaseImage(t)
	assertImageDiffersAt(t, baseImage, rendered, int(orderCareReportPortraitFrame.CenterX), int(orderCareReportPortraitFrame.CenterY), "portrait center")
	for _, fieldName := range []string{"pet_name", "breed", "gender", "age", "care_content", "care_date", "next_care_date", "weight"} {
		box := orderCareReportPrimaryFieldBoxes[fieldName]
		assertImageDiffersNear(t, baseImage, rendered, int((box.Left+box.Right)/2), int(box.Baseline)-12, 60, fieldName+" text")
	}
	assertImageDiffersNear(t, baseImage, rendered, int(orderCareReportBodyShapeAnchors["standard"].X), int(orderCareReportBodyShapeAnchors["standard"].Y), 18, "body shape checkmark")

	sectionChecks := map[string]string{
		"skin":      "normal",
		"hair":      "dry",
		"nails":     "trimmed",
		"eyes_face": "cleaned",
		"ears":      "cleaned",
		"oral":      "normal",
		"anus":      "normal",
	}
	for sectionName, checkName := range sectionChecks {
		layout := orderCareReportSectionLayouts[sectionName]
		anchor := layout.Checkboxes[checkName]
		assertImageDiffersNear(t, baseImage, rendered, int(anchor.X), int(anchor.Y), 18, sectionName+" checkmark")
		assertImageDiffersNear(t, baseImage, rendered, int((layout.NoteBox.Left+layout.NoteBox.Right)/2), int(layout.NoteBox.Baseline)-8, 90, sectionName+" note")
	}
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
	if sourcePath := os.Getenv("CARE_REPORT_PORTRAIT_SOURCE"); sourcePath != "" {
		portrait, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read portrait source: %v", err)
		}
		if err := os.WriteFile(path, portrait, 0644); err != nil {
			t.Fatalf("write portrait upload: %v", err)
		}
		return "/uploads/" + filename
	}

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

func changedPixelXBounds(base, rendered image.Image, region image.Rectangle) (int, int, bool) {
	minX := region.Max.X
	maxX := region.Min.X
	found := false
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			br, bg, bb, ba := base.At(x, y).RGBA()
			rr, rg, rb, ra := rendered.At(x, y).RGBA()
			if br == rr && bg == rg && bb == rb && ba == ra {
				continue
			}
			found = true
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
		}
	}
	return minX, maxX, found
}
