# Order Care Report Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a paid-order `生成报告` flow that lets staff pick a cat, upload and crop a portrait, fill a care report form, submit it to the backend, receive a server-rendered report image, save it on mobile, and automatically create a `pet_bath_reports` record.

**Architecture:** Keep report editing on the H5 side and move final image rendering to the backend. The frontend collects and validates structured report input, uploads the cropped portrait with the existing upload API, then submits a typed payload to a new order care-report endpoint. The backend validates order state and pet ownership, renders onto the fixed `base.jpg` template with a bundled font and absolute coordinates, saves the final JPG to `/uploads`, and persists one `PetBathReport` row.

**Tech Stack:** uni-app + Vue 3 + TypeScript, Gin, GORM, Go `embed`, `github.com/fogleman/gg`, existing upload API, existing receipt-style mobile image save flow.

---

## File Map

### Backend

- Create: `server/internal/service/order_care_report_service.go`
  - Defines the request/response contract, order/pet validation, rendering orchestration, upload-path write, and bath report persistence.
- Create: `server/internal/service/order_care_report_layout.go`
  - Stores the fixed template size, portrait frame bounds, text anchors, checkbox anchors, and note block limits in one place.
- Create: `server/internal/service/order_care_report_service_test.go`
  - Covers paid-order validation, pet ownership validation, render output dimensions, and bath report persistence.
- Create: `server/internal/handler/order_care_report_test.go`
  - Covers request binding, HTTP status mapping, and endpoint success/error contracts.
- Create: `server/internal/service/assets/order-care-report/base.jpg`
  - Bundled copy of the approved base template image.
- Create: `server/internal/service/assets/order-care-report/SourceHanSansSC-Regular.otf`
  - Bundled Chinese font used for deterministic text layout.
- Modify: `server/internal/handler/order.go`
  - Adds the `POST /b/orders/:id/care-report` request type and handler method.
- Modify: `server/internal/router/router.go`
  - Wires the new business route.
- Modify: `server/go.mod`
  - Adds `github.com/fogleman/gg`.
- Modify: `server/go.sum`
  - Captures the new dependency lock.

### Frontend

- Create: `web/src/api/order-care-report.ts`
  - Typed API wrapper for report generation.
- Create: `web/src/utils/order-care-report.ts`
  - Pure helpers for selecting reportable cats, building defaults from an order, and converting draft state into request payload.
- Create: `web/src/utils/web-image-save.ts`
  - Shared image-save helper that preserves current receipt behavior, including iPhone Safari preview-page fallback.
- Create: `web/src/components/order/OrderCareReportModal.vue`
  - Fullscreen modal for cat selection, portrait upload/crop, form fill, generate, preview, and save.
- Create: `web/scripts/test-order-care-report.ts`
  - Regression script for the pure order-report helpers.
- Modify: `web/src/pages/order/detail.vue`
  - Adds the `生成报告` action, opens the modal, and reuses the shared save helper for receipt/report image flows.
- Modify: `web/src/types/index.d.ts`
  - Adds frontend types for report request/response payloads if needed by page/component code.

## Concrete Payload Contract

Use this exact request shape end-to-end:

```ts
export interface OrderCareReportSectionInput {
  checks: string[]
  note: string
}

export interface CreateOrderCareReportReq {
  pet_id: number
  portrait_url: string
  weight: string
  care_date: string
  next_care_date: string
  care_content: string
  body_shape: string
  skin: OrderCareReportSectionInput
  hair: OrderCareReportSectionInput
  nails: OrderCareReportSectionInput
  eyes_face: OrderCareReportSectionInput
  ears: OrderCareReportSectionInput
  oral: OrderCareReportSectionInput
  anus: OrderCareReportSectionInput
}
```

Use these stable checkbox codes in both frontend and backend:

```text
body_shape: thin, skinny, standard, chubby, obese

skin: normal, dandruff, red, greasy, scab
hair: shedding, undercoat_many, dry, greasy, matting
nails: trimmed, dewclaw_abnormal, pads_dry, too_long, wound
eyes_face: cleaned, tear_many, eye_red, eye_abnormal, wound
ears: cleaned, touch_sensitive, inflamed, earwax, black_earwax
oral: normal, tartar, gum_red, gum_swollen
anus: normal, red, inflamed, anal_gland_swollen
```

Use this exact response shape:

```ts
export interface CreateOrderCareReportResp {
  image_url: string
  report_id: number
  bath_date: string
}
```

### Task 1: Backend Validation Contract

**Files:**
- Create: `server/internal/service/order_care_report_service.go`
- Create: `server/internal/service/order_care_report_service_test.go`
- Test: `server/internal/service/order_care_report_service_test.go`

- [ ] **Step 1: Write the failing test**

Add validation-first tests in `server/internal/service/order_care_report_service_test.go`:

```go
package service

import (
	"testing"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

func TestCreateOrderCareReportRejectsUnpaidOrder(t *testing.T) {
	setupOrderServiceTestDB(t)
	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:   1,
		OrderPaid: false,
	})

	svc := NewOrderCareReportService(
		repository.NewOrderRepository(),
		repository.NewPetBathReportRepository(),
	)

	_, err := svc.Create(1, state.order.ID, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  "/uploads/portrait.jpg",
		Weight:       "5.55",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-05-20",
		CareContent:  "Harmurry精致皮毛调理",
		BodyShape:    "standard",
	})
	if err == nil || err.Error() != "仅已支付订单可生成报告" {
		t.Fatalf("expected paid-order validation error, got %v", err)
	}
}

func TestCreateOrderCareReportRejectsPetOutsideOrder(t *testing.T) {
	setupOrderServiceTestDB(t)
	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:   1,
		OrderPaid: true,
	})
	otherPet := seedCareReportPet(t, 1, state.customer.ID, "无关猫咪")

	svc := NewOrderCareReportService(
		repository.NewOrderRepository(),
		repository.NewPetBathReportRepository(),
	)

	_, err := svc.Create(1, state.order.ID, CreateOrderCareReportInput{
		PetID:        otherPet.ID,
		PortraitURL:  "/uploads/portrait.jpg",
		Weight:       "5.55",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-05-20",
		CareContent:  "Harmurry精致皮毛调理",
		BodyShape:    "standard",
	})
	if err == nil || err.Error() != "所选猫咪不属于当前订单" {
		t.Fatalf("expected pet ownership validation error, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd server && go test ./internal/service -run 'TestCreateOrderCareReport(RejectsUnpaidOrder|RejectsPetOutsideOrder)$' -v
```

Expected:

```text
FAIL
undefined: NewOrderCareReportService
undefined: CreateOrderCareReportInput
```

- [ ] **Step 3: Write minimal implementation**

Create `server/internal/service/order_care_report_service.go` with a validation-first scaffold:

```go
package service

import (
	"errors"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type OrderCareReportSectionInput struct {
	Checks []string `json:"checks"`
	Note   string   `json:"note"`
}

type CreateOrderCareReportInput struct {
	PetID        uint                        `json:"pet_id"`
	PortraitURL  string                      `json:"portrait_url"`
	Weight       string                      `json:"weight"`
	CareDate     string                      `json:"care_date"`
	NextCareDate string                      `json:"next_care_date"`
	CareContent  string                      `json:"care_content"`
	BodyShape    string                      `json:"body_shape"`
	Skin         OrderCareReportSectionInput `json:"skin"`
	Hair         OrderCareReportSectionInput `json:"hair"`
	Nails        OrderCareReportSectionInput `json:"nails"`
	EyesFace     OrderCareReportSectionInput `json:"eyes_face"`
	Ears         OrderCareReportSectionInput `json:"ears"`
	Oral         OrderCareReportSectionInput `json:"oral"`
	Anus         OrderCareReportSectionInput `json:"anus"`
}

type OrderCareReportResult struct {
	ImageURL string `json:"image_url"`
	ReportID uint   `json:"report_id"`
	BathDate string `json:"bath_date"`
}

type OrderCareReportService struct {
	orderRepo         *repository.OrderRepository
	petBathReportRepo *repository.PetBathReportRepository
}

func NewOrderCareReportService(orderRepo *repository.OrderRepository, petBathReportRepo *repository.PetBathReportRepository) *OrderCareReportService {
	return &OrderCareReportService{
		orderRepo:         orderRepo,
		petBathReportRepo: petBathReportRepo,
	}
}

func (s *OrderCareReportService) Create(shopID, orderID uint, input CreateOrderCareReportInput) (*OrderCareReportResult, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil || order.ShopID != shopID {
		return nil, errors.New("订单不存在")
	}
	if order.PayStatus != 1 {
		return nil, errors.New("仅已支付订单可生成报告")
	}
	if !orderHasPet(order, input.PetID) {
		return nil, errors.New("所选猫咪不属于当前订单")
	}
	if input.PortraitURL == "" {
		return nil, errors.New("请先上传护理照片")
	}
	if input.NextCareDate == "" {
		return nil, errors.New("请填写建议下次护理日期")
	}
	return nil, errors.New("render not implemented")
}

func orderHasPet(order *model.Order, petID uint) bool {
	if order == nil || petID == 0 {
		return false
	}
	if order.PetID != nil && *order.PetID == petID {
		return true
	}
	if order.Appointment != nil {
		for _, item := range order.Appointment.Pets {
			if item.PetID == petID {
				return true
			}
		}
	}
	if order.FeedingPlan != nil {
		for _, item := range order.FeedingPlan.Pets {
			if item.PetID == petID {
				return true
			}
		}
	}
	for _, group := range order.PetGroups {
		if group.PetID != nil && *group.PetID == petID {
			return true
		}
	}
	return false
}

func parseCareReportDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}
```

Add local fixture helpers at the bottom of the test file:

```go
type seedCareReportOrderFixtureInput struct {
	ShopID    uint
	OrderPaid bool
}

type careReportOrderFixture struct {
	customer model.Customer
	pet      model.Pet
	order    model.Order
}

func seedCareReportOrderFixture(t *testing.T, input seedCareReportOrderFixtureInput) careReportOrderFixture {
	t.Helper()

	customer := model.Customer{ShopID: input.ShopID, Name: "报告客户", Phone: "13800138001"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}

	pet := model.Pet{ShopID: input.ShopID, CustomerID: &customer.ID, Name: "福福", Breed: "挪威森林猫", Gender: 1}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}

	order := model.Order{
		OrderNo:    "TEST-CARE-REPORT",
		ShopID:     input.ShopID,
		CustomerID: &customer.ID,
		PetID:      &pet.ID,
		PayStatus:  0,
		Status:     0,
	}
	if input.OrderPaid {
		order.PayStatus = 1
		order.Status = 1
		now := time.Date(2026, 4, 20, 11, 14, 17, 0, time.Local)
		order.PayTime = &now
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	item := model.OrderItem{OrderID: order.ID, ItemType: 1, ItemID: 10, Name: "福福 · Harmurry精致皮毛调理", Quantity: 1, UnitPrice: 168, Amount: 168}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}

	return careReportOrderFixture{customer: customer, pet: pet, order: order}
}

func seedCareReportPet(t *testing.T, shopID uint, customerID uint, name string) model.Pet {
	t.Helper()

	pet := model.Pet{ShopID: shopID, CustomerID: &customerID, Name: name}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create extra pet: %v", err)
	}
	return pet
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd server && go test ./internal/service -run 'TestCreateOrderCareReport(RejectsUnpaidOrder|RejectsPetOutsideOrder)$' -v
```

Expected:

```text
PASS
ok  	github.com/neinei960/cat/server/internal/service
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/service/order_care_report_service.go server/internal/service/order_care_report_service_test.go
git commit -m "feat: add order care report validation scaffold"
```

### Task 2: Backend Rendering, Upload Write, and Bath Report Persistence

**Files:**
- Create: `server/internal/service/order_care_report_layout.go`
- Create: `server/internal/service/assets/order-care-report/base.jpg`
- Create: `server/internal/service/assets/order-care-report/SourceHanSansSC-Regular.otf`
- Modify: `server/internal/service/order_care_report_service.go`
- Modify: `server/internal/service/order_care_report_service_test.go`
- Modify: `server/go.mod`
- Modify: `server/go.sum`
- Test: `server/internal/service/order_care_report_service_test.go`

- [ ] **Step 1: Write the failing test**

Extend `server/internal/service/order_care_report_service_test.go` with a happy-path render test:

```go
func TestCreateOrderCareReportCreatesBathReportAndRenderedImage(t *testing.T) {
	setupOrderServiceTestDB(t)
	state := seedCareReportOrderFixture(t, seedCareReportOrderFixtureInput{
		ShopID:    1,
		OrderPaid: true,
	})

	uploadDir := t.TempDir()
	config.AppConfig.Upload.Path = uploadDir
	portraitURL := writeTestPortraitUpload(t, uploadDir, "portrait-source.jpg")

	svc := NewOrderCareReportService(
		repository.NewOrderRepository(),
		repository.NewPetBathReportRepository(),
	)

	result, err := svc.Create(1, state.order.ID, CreateOrderCareReportInput{
		PetID:        state.pet.ID,
		PortraitURL:  portraitURL,
		Weight:       "5.55",
		CareDate:     "2026-04-20",
		NextCareDate: "2026-05-20",
		CareContent:  "Harmurry精致皮毛调理",
		BodyShape:    "standard",
		Skin:         OrderCareReportSectionInput{Checks: []string{"normal"}, Note: "皮肤状态稳定"},
	})
	if err != nil {
		t.Fatalf("create care report: %v", err)
	}

	if result.ImageURL == "" || result.ReportID == 0 {
		t.Fatalf("expected persisted report result, got %+v", result)
	}

	imgPath := filepath.Join(uploadDir, filepath.Base(result.ImageURL))
	file, err := os.Open(imgPath)
	if err != nil {
		t.Fatalf("open rendered image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decode rendered image: %v", err)
	}
	if img.Bounds().Dx() != 1279 || img.Bounds().Dy() != 1810 {
		t.Fatalf("expected 1279x1810 image, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}

	var count int64
	if err := database.DB.Model(&model.PetBathReport{}).
		Where("pet_id = ? AND image_url = ?", state.pet.ID, result.ImageURL).
		Count(&count).Error; err != nil {
		t.Fatalf("count pet bath report rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 persisted bath report row, got %d", count)
	}
}

func writeTestPortraitUpload(t *testing.T, uploadDir string, fileName string) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 480, 480))
	buf := bytes.NewBuffer(nil)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode portrait: %v", err)
	}
	fullPath := filepath.Join(uploadDir, fileName)
	if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write portrait: %v", err)
	}
	return "/uploads/" + fileName
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd server && go test ./internal/service -run TestCreateOrderCareReportCreatesBathReportAndRenderedImage -v
```

Expected:

```text
FAIL
render not implemented
```

- [ ] **Step 3: Write minimal implementation**

Add `github.com/fogleman/gg`:

```bash
cd server && go get github.com/fogleman/gg@v1.3.0
```

Create `server/internal/service/order_care_report_layout.go`:

```go
package service

type orderCareReportLayout struct {
	Width  int
	Height int
}

var defaultOrderCareReportLayout = orderCareReportLayout{
	Width:  1279,
	Height: 1810,
}
```

Extend `server/internal/service/order_care_report_service.go`:

```go
package service

import (
	"bytes"
	"embed"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
	"github.com/google/uuid"
	"github.com/neinei960/cat/server/config"
	"github.com/neinei960/cat/server/internal/model"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/order-care-report/base.jpg assets/order-care-report/SourceHanSansSC-Regular.otf
var orderCareReportAssets embed.FS

func (s *OrderCareReportService) Create(shopID, orderID uint, input CreateOrderCareReportInput) (*OrderCareReportResult, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil || order.ShopID != shopID {
		return nil, errors.New("订单不存在")
	}
	if order.PayStatus != 1 {
		return nil, errors.New("仅已支付订单可生成报告")
	}
	if !orderHasPet(order, input.PetID) {
		return nil, errors.New("所选猫咪不属于当前订单")
	}
	if input.PortraitURL == "" {
		return nil, errors.New("请先上传护理照片")
	}
	if input.NextCareDate == "" {
		return nil, errors.New("请填写建议下次护理日期")
	}

	bathDate, err := parseCareReportDate(input.CareDate)
	if err != nil {
		return nil, errors.New("护理日期格式错误")
	}

	rendered, err := renderOrderCareReport(input)
	if err != nil {
		return nil, err
	}

	uploadPath := config.AppConfig.Upload.Path
	if uploadPath == "" {
		uploadPath = "./uploads"
	}
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, err
	}

	fileName := "care-report-" + uuid.NewString() + ".jpg"
	fullPath := filepath.Join(uploadPath, fileName)
	if err := os.WriteFile(fullPath, rendered, 0644); err != nil {
		return nil, err
	}

	report := &model.PetBathReport{
		ShopID:   shopID,
		PetID:    input.PetID,
		ImageURL: "/uploads/" + fileName,
		BathDate: &bathDate,
	}
	if err := s.petBathReportRepo.Create(report); err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}

	return &OrderCareReportResult{
		ImageURL: report.ImageURL,
		ReportID: report.ID,
		BathDate: bathDate.Format("2006-01-02"),
	}, nil
}

func renderOrderCareReport(input CreateOrderCareReportInput) ([]byte, error) {
	baseBytes, err := orderCareReportAssets.ReadFile("assets/order-care-report/base.jpg")
	if err != nil {
		return nil, err
	}
	baseImg, _, err := image.Decode(bytes.NewReader(baseBytes))
	if err != nil {
		return nil, err
	}
	fontBytes, err := orderCareReportAssets.ReadFile("assets/order-care-report/SourceHanSansSC-Regular.otf")
	if err != nil {
		return nil, err
	}
	fontFile, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(fontFile, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(defaultOrderCareReportLayout.Width, defaultOrderCareReportLayout.Height)
	dc.DrawImage(baseImg, 0, 0)
	dc.SetFontFace(face)

	dc.DrawStringAnchored(input.CareContent, 385, 585, 0, 0.5)
	dc.DrawStringAnchored(input.Weight+" KG", 165, 780, 0, 0.5)
	dc.DrawStringAnchored(input.NextCareDate, 1000, 470, 0.5, 0.5)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dc.Image(), &jpeg.Options{Quality: 92}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd server && go test ./internal/service -run TestCreateOrderCareReportCreatesBathReportAndRenderedImage -v
```

Expected:

```text
PASS
ok  	github.com/neinei960/cat/server/internal/service
```

- [ ] **Step 5: Commit**

```bash
git add server/go.mod server/go.sum server/internal/service/order_care_report_service.go server/internal/service/order_care_report_layout.go server/internal/service/order_care_report_service_test.go server/internal/service/assets/order-care-report/base.jpg server/internal/service/assets/order-care-report/SourceHanSansSC-Regular.otf
git commit -m "feat: render and persist order care reports"
```

### Task 3: HTTP Endpoint Wiring

**Files:**
- Modify: `server/internal/handler/order.go`
- Modify: `server/internal/router/router.go`
- Create: `server/internal/handler/order_care_report_test.go`
- Test: `server/internal/handler/order_care_report_test.go`

- [ ] **Step 1: Write the failing test**

Create `server/internal/handler/order_care_report_test.go`:

```go
package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateOrderCareReportReturns400ForUnpaidOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupHandlerOrderCareReportTestDB(t)
	state := seedHandlerCareReportFixture(t, false)

	orderHandler := NewOrderHandler(
		service.NewOrderService(repository.NewOrderRepository(), repository.NewAppointmentRepository()),
		nil,
		nil,
		nil,
	).WithCareReportService(service.NewOrderCareReportService(
		repository.NewOrderRepository(),
		repository.NewPetBathReportRepository(),
	))

	router := gin.New()
	router.POST("/b/orders/:id/care-report", func(c *gin.Context) {
		c.Set("shop_id", uint(1))
		orderHandler.CreateCareReport(c)
	})

	body := []byte(`{"pet_id":1,"portrait_url":"/uploads/portrait.jpg","weight":"5.55","care_date":"2026-04-20","next_care_date":"2026-05-20","care_content":"Harmurry精致皮毛调理","body_shape":"standard"}`)
	req := httptest.NewRequest(http.MethodPost, "/b/orders/"+strconv.Itoa(int(state.order.ID))+"/care-report", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
}

type handlerCareReportFixture struct {
	order model.Order
}

func setupHandlerOrderCareReportTestDB(t *testing.T) {
	t.Helper()

	dsn := "file:handler-care-report?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	database.DB = db

	if err := database.DB.AutoMigrate(
		&model.Customer{},
		&model.Pet{},
		&model.Appointment{},
		&model.AppointmentPet{},
		&model.Order{},
		&model.OrderItem{},
		&model.PetBathReport{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func seedHandlerCareReportFixture(t *testing.T, paid bool) handlerCareReportFixture {
	t.Helper()

	customer := model.Customer{ShopID: 1, Name: "报告客户", Phone: "13800138000"}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "福福"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create pet: %v", err)
	}
	order := model.Order{
		OrderNo:    "TEST-CARE-REPORT",
		ShopID:     1,
		CustomerID: &customer.ID,
		PetID:      &pet.ID,
		PayStatus:  0,
		Status:     0,
	}
	if paid {
		order.PayStatus = 1
		order.Status = 1
		now := time.Date(2026, 4, 20, 11, 14, 17, 0, time.Local)
		order.PayTime = &now
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	item := model.OrderItem{OrderID: order.ID, ItemType: 1, ItemID: 10, Name: "福福 · Harmurry精致皮毛调理", Quantity: 1, UnitPrice: 168, Amount: 168}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}
	return handlerCareReportFixture{order: order}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd server && go test ./internal/handler -run TestCreateOrderCareReportReturns400ForUnpaidOrder -v
```

Expected:

```text
FAIL
orderHandler.CreateCareReport undefined
```

- [ ] **Step 3: Write minimal implementation**

Modify `server/internal/handler/order.go`:

```go
type createCareReportReq struct {
	PetID        uint                               `json:"pet_id" binding:"required"`
	PortraitURL  string                             `json:"portrait_url" binding:"required"`
	Weight       string                             `json:"weight"`
	CareDate     string                             `json:"care_date" binding:"required"`
	NextCareDate string                             `json:"next_care_date" binding:"required"`
	CareContent  string                             `json:"care_content"`
	BodyShape    string                             `json:"body_shape"`
	Skin         service.OrderCareReportSectionInput `json:"skin"`
	Hair         service.OrderCareReportSectionInput `json:"hair"`
	Nails        service.OrderCareReportSectionInput `json:"nails"`
	EyesFace     service.OrderCareReportSectionInput `json:"eyes_face"`
	Ears         service.OrderCareReportSectionInput `json:"ears"`
	Oral         service.OrderCareReportSectionInput `json:"oral"`
	Anus         service.OrderCareReportSectionInput `json:"anus"`
}

func (h *OrderHandler) WithCareReportService(careReportService *service.OrderCareReportService) *OrderHandler {
	h.careReportService = careReportService
	return h
}

func (h *OrderHandler) CreateCareReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "订单ID错误")
		return
	}

	var req createCareReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := h.careReportService.Create(c.GetUint("shop_id"), uint(id), service.CreateOrderCareReportInput{
		PetID:        req.PetID,
		PortraitURL:  req.PortraitURL,
		Weight:       req.Weight,
		CareDate:     req.CareDate,
		NextCareDate: req.NextCareDate,
		CareContent:  req.CareContent,
		BodyShape:    req.BodyShape,
		Skin:         req.Skin,
		Hair:         req.Hair,
		Nails:        req.Nails,
		EyesFace:     req.EyesFace,
		Ears:         req.Ears,
		Oral:         req.Oral,
		Anus:         req.Anus,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}
```

Extend the handler struct and router:

```go
type OrderHandler struct {
	orderService      *service.OrderService
	careReportService *service.OrderCareReportService
	petService        *service.PetService
	customerService   *service.CustomerService
	serviceService    *service.ServiceService
}
```

```go
b.POST("/orders/:id/care-report", orderHandler.CreateCareReport)
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd server && go test ./internal/handler -run TestCreateOrderCareReportReturns400ForUnpaidOrder -v
```

Expected:

```text
PASS
ok  	github.com/neinei960/cat/server/internal/handler
```

- [ ] **Step 5: Commit**

```bash
git add server/internal/handler/order.go server/internal/router/router.go server/internal/handler/order_care_report_test.go
git commit -m "feat: add order care report endpoint"
```

### Task 4: Frontend Types, API, and Pure Mapping Helpers

**Files:**
- Create: `web/src/api/order-care-report.ts`
- Create: `web/src/utils/order-care-report.ts`
- Create: `web/scripts/test-order-care-report.ts`
- Modify: `web/src/types/index.d.ts`
- Test: `web/scripts/test-order-care-report.ts`

- [ ] **Step 1: Write the failing test**

Create `web/scripts/test-order-care-report.ts`:

```ts
import assert from 'node:assert/strict'
import {
  buildOrderCareReportDraft,
  buildOrderCareReportPetOptions,
  buildOrderCareReportPayload,
  canGenerateOrderCareReport,
} from '../src/utils/order-care-report'

const order = {
  ID: 167,
  status: 1,
  pay_time: '2026-04-20 11:14:17',
  pet: { ID: 12, name: '福福', breed: '挪威森林猫', gender: 1, birth_date: '2025-02-01' },
  items: [
    { item_type: 1, item_id: 10, name: '福福 · Harmurry精致皮毛调理', quantity: 1, unit_price: 168, amount: 168 },
  ],
  pet_groups: [
    { pet_id: 12, pet_name: '福福', items: [{ item_type: 1, item_id: 10, name: '福福 · Harmurry精致皮毛调理', quantity: 1, unit_price: 168, amount: 168 }] },
  ],
} as any

assert.equal(canGenerateOrderCareReport(order), true)

const options = buildOrderCareReportPetOptions(order)
assert.equal(options.length, 1)
assert.equal(options[0].petId, 12)

const draft = buildOrderCareReportDraft(order, 12)
assert.equal(draft.petName, '福福')
assert.equal(draft.breed, '挪威森林猫')
assert.equal(draft.careContent, 'Harmurry精致皮毛调理')

const payload = buildOrderCareReportPayload({
  ...draft,
  portraitUrl: '/uploads/portrait.jpg',
  weight: '5.55',
  nextCareDate: '2026-05-20',
})
assert.equal(payload.pet_id, 12)
assert.equal(payload.portrait_url, '/uploads/portrait.jpg')
assert.deepEqual(payload.skin.checks, [])
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd web && npx tsc ./src/utils/order-care-report.ts ./scripts/test-order-care-report.ts --module commonjs --target ES2020 --lib ES2020,DOM --outDir ./.tmp/order-care-report-test && node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
```

Expected:

```text
error TS6053: File 'src/utils/order-care-report.ts' not found.
```

- [ ] **Step 3: Write minimal implementation**

Create `web/src/api/order-care-report.ts`:

```ts
import { request } from './request'

export interface OrderCareReportSectionInput {
  checks: string[]
  note: string
}

export interface CreateOrderCareReportReq {
  pet_id: number
  portrait_url: string
  weight: string
  care_date: string
  next_care_date: string
  care_content: string
  body_shape: string
  skin: OrderCareReportSectionInput
  hair: OrderCareReportSectionInput
  nails: OrderCareReportSectionInput
  eyes_face: OrderCareReportSectionInput
  ears: OrderCareReportSectionInput
  oral: OrderCareReportSectionInput
  anus: OrderCareReportSectionInput
}

export interface CreateOrderCareReportResp {
  image_url: string
  report_id: number
  bath_date: string
}

export function createOrderCareReport(orderId: number, data: CreateOrderCareReportReq) {
  return request<CreateOrderCareReportResp>({ url: `/b/orders/${orderId}/care-report`, method: 'POST', data })
}
```

Create `web/src/utils/order-care-report.ts`:

```ts
import type { CreateOrderCareReportReq } from '@/api/order-care-report'

export interface OrderCareReportPetOption {
  petId: number
  petName: string
}

export interface OrderCareReportDraft {
  petId: number
  petName: string
  breed: string
  gender: string
  age: string
  careDate: string
  careContent: string
  portraitUrl: string
  weight: string
  nextCareDate: string
  bodyShape: string
  skin: { checks: string[]; note: string }
  hair: { checks: string[]; note: string }
  nails: { checks: string[]; note: string }
  eyesFace: { checks: string[]; note: string }
  ears: { checks: string[]; note: string }
  oral: { checks: string[]; note: string }
  anus: { checks: string[]; note: string }
}

export function canGenerateOrderCareReport(order: Order | null | undefined) {
  if (!order) return false
  if (Number(order.status || 0) !== 1) return false
  return buildOrderCareReportPetOptions(order).length > 0
}

export function buildOrderCareReportPetOptions(order: Order): OrderCareReportPetOption[] {
  const seen = new Set<number>()
  const result: OrderCareReportPetOption[] = []
  const pushPet = (petId: number, petName: string) => {
    if (!petId || seen.has(petId)) return
    seen.add(petId)
    result.push({ petId, petName })
  }
  if (order.pet?.ID) pushPet(order.pet.ID, order.pet.name || '当前猫咪')
  for (const group of order.pet_groups || []) {
    pushPet(Number(group.pet_id || 0), group.pet_name || '当前猫咪')
  }
  return result
}

export function buildOrderCareReportDraft(order: Order, petId: number): OrderCareReportDraft {
  const appointmentPet = ((order as any).appointment?.pets || []).find((item: any) => Number(item?.pet_id || item?.pet?.ID || 0) === petId)?.pet || null
  const pet = order.pet?.ID === petId ? order.pet : appointmentPet
  const itemNames = (order.items || [])
    .filter(item => Number(item.item_type) === 1)
    .map(item => String(item.name || '').split(' · ').pop() || '')
    .filter(Boolean)
  const age = formatAgeFromBirthDate(String(pet?.birth_date || ''))

  return {
    petId,
    petName: pet?.name || buildOrderCareReportPetOptions(order).find(item => item.petId === petId)?.petName || '当前猫咪',
    breed: pet?.breed || '',
    gender: pet?.gender === 1 ? 'GG' : pet?.gender === 2 ? 'MM' : '',
    age,
    careDate: String(order.pay_time || order.CreatedAt || '').slice(0, 10).replace(/-/g, '.'),
    careContent: itemNames.join('、'),
    portraitUrl: '',
    weight: '',
    nextCareDate: '',
    bodyShape: '',
    skin: { checks: [], note: '' },
    hair: { checks: [], note: '' },
    nails: { checks: [], note: '' },
    eyesFace: { checks: [], note: '' },
    ears: { checks: [], note: '' },
    oral: { checks: [], note: '' },
    anus: { checks: [], note: '' },
  }
}

function formatAgeFromBirthDate(value: string) {
  if (!value) return ''
  const birth = new Date(value)
  if (Number.isNaN(birth.getTime())) return ''
  const now = new Date()
  const months = (now.getFullYear() - birth.getFullYear()) * 12 + (now.getMonth() - birth.getMonth())
  if (months < 1) return '不到1个月'
  if (months < 12) return `${months}个月`
  const years = Math.floor(months / 12)
  const remain = months % 12
  return remain > 0 ? `${years}岁${remain}个月` : `${years}岁`
}

export function buildOrderCareReportPayload(draft: OrderCareReportDraft): CreateOrderCareReportReq {
  return {
    pet_id: draft.petId,
    portrait_url: draft.portraitUrl,
    weight: draft.weight,
    care_date: draft.careDate.replace(/\./g, '-'),
    next_care_date: draft.nextCareDate.replace(/\./g, '-'),
    care_content: draft.careContent,
    body_shape: draft.bodyShape,
    skin: draft.skin,
    hair: draft.hair,
    nails: draft.nails,
    eyes_face: draft.eyesFace,
    ears: draft.ears,
    oral: draft.oral,
    anus: draft.anus,
  }
}
```

Modify `web/src/types/index.d.ts` so order detail code can read appointment pets and keep the report API typed:

```ts
interface Order {
  appointment?: {
    pets?: Array<{
      pet_id?: number
      pet?: Pet
    }>
  }
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd web && npx tsc ./src/utils/order-care-report.ts ./scripts/test-order-care-report.ts --module commonjs --target ES2020 --lib ES2020,DOM --outDir ./.tmp/order-care-report-test && node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
```

Expected:

```text
(no output)
exit 0
```

- [ ] **Step 5: Commit**

```bash
git add web/src/api/order-care-report.ts web/src/utils/order-care-report.ts web/scripts/test-order-care-report.ts web/src/types/index.d.ts
git commit -m "feat: add frontend order care report helpers"
```

### Task 5: Order Detail Modal, Portrait Crop, Preview, and Save

**Files:**
- Create: `web/src/utils/web-image-save.ts`
- Create: `web/src/components/order/OrderCareReportModal.vue`
- Modify: `web/src/pages/order/detail.vue`
- Modify: `web/scripts/test-order-care-report.ts`
- Test: `web/scripts/test-order-care-report.ts`

- [ ] **Step 1: Write the failing test**

Extend `web/scripts/test-order-care-report.ts` with entry-visibility and filename coverage:

```ts
import { buildOrderCareReportFileName } from '../src/utils/web-image-save'

assert.equal(canGenerateOrderCareReport({ status: 0, pet_groups: [] } as any), false)
assert.equal(buildOrderCareReportFileName('NO167', '福福'), '护理报告_NO167_福福.png')
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd web && npx tsc ./src/utils/order-care-report.ts ./src/utils/web-image-save.ts ./scripts/test-order-care-report.ts --module commonjs --target ES2020 --lib ES2020,DOM --outDir ./.tmp/order-care-report-test && node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
```

Expected:

```text
error TS6053: File 'src/utils/web-image-save.ts' not found.
```

- [ ] **Step 3: Write minimal implementation**

Create `web/src/utils/web-image-save.ts`:

```ts
export function isAppleSafariBrowser() {
  if (typeof navigator === 'undefined') return false
  const userAgent = navigator.userAgent || ''
  const vendor = navigator.vendor || ''
  const isAppleMobile = /iP(hone|od|ad)/i.test(userAgent)
  const isSafari = /Safari/i.test(userAgent) && !/CriOS|FxiOS|EdgiOS|OPiOS|Android/i.test(userAgent) && /Apple/i.test(vendor)
  return isAppleMobile && isSafari
}

export function buildOrderCareReportFileName(orderNo: string, petName: string) {
  return `护理报告_${orderNo}_${petName}.png`
}

export function openImagePreviewWindow(src: string, title = '护理报告图片') {
  if (!src || typeof window === 'undefined') return false
  const previewWindow = window.open('', '_blank')
  if (!previewWindow) return false
  previewWindow.document.write(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${title}</title></head><body style="margin:0;background:#111;padding:16px"><img src="${src}" style="display:block;max-width:100%;height:auto;margin:0 auto" alt="${title}" /></body></html>`)
  previewWindow.document.close()
  return true
}

export async function saveImageByUrl(src: string, fileName: string) {
  if (!src || typeof window === 'undefined' || typeof document === 'undefined') return
  if (isAppleSafariBrowser()) {
    if (openImagePreviewWindow(src)) return
  }
  const a = document.createElement('a')
  a.href = src
  a.download = fileName
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
```

Create `web/src/components/order/OrderCareReportModal.vue`:

```vue
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ImageCropper from '@/components/ImageCropper.vue'
import { uploadFile } from '@/api/upload'
import { createOrderCareReport } from '@/api/order-care-report'
import { buildOrderCareReportDraft, buildOrderCareReportPayload, buildOrderCareReportPetOptions } from '@/utils/order-care-report'
import { createCropperPreviewUrl } from '@/utils/image-cropper'
import { buildOrderCareReportFileName, saveImageByUrl } from '@/utils/web-image-save'

const props = defineProps<{ visible: boolean; order: Order | null }>()
const emit = defineEmits<{ close: [] }>()

const selectedPetId = ref(0)
const previewUrl = ref('')
const submitting = ref(false)
const draft = ref<any>(null)
const cropperVisible = ref(false)
const cropperSrc = ref('')

watch(() => props.visible, (visible) => {
  if (!visible || !props.order) return
  const petOptions = buildOrderCareReportPetOptions(props.order)
  selectedPetId.value = petOptions[0]?.petId || 0
  draft.value = selectedPetId.value ? buildOrderCareReportDraft(props.order, selectedPetId.value) : null
  previewUrl.value = ''
}, { immediate: true })

async function choosePortrait() {
  const chooseResult = await new Promise<any>((resolve, reject) => {
    uni.chooseImage({ count: 1, sizeType: ['compressed'], sourceType: ['album', 'camera'], success: resolve, fail: reject })
  })
  const rawPath = chooseResult?.tempFilePaths?.[0]
  if (!rawPath) return

  if (typeof window !== 'undefined') {
    const response = await fetch(rawPath)
    const blob = await response.blob()
    const file = new File([blob], rawPath.split('/').pop() || 'portrait.jpg', { type: blob.type || 'image/jpeg' })
    cropperSrc.value = await createCropperPreviewUrl(file)
  } else {
    cropperSrc.value = rawPath
  }
  cropperVisible.value = true
}

async function onCropConfirm(croppedPath: string) {
  if (!draft.value) return
  draft.value.portraitUrl = await uploadFile(croppedPath)
  cropperVisible.value = false
}

async function submit() {
  if (!props.order || !draft.value) return
  submitting.value = true
  try {
    const res = await createOrderCareReport(props.order.ID, buildOrderCareReportPayload(draft.value))
    previewUrl.value = res.data.image_url
  } finally {
    submitting.value = false
  }
}

async function save() {
  if (!props.order || !draft.value || !previewUrl.value) return
  await saveImageByUrl(previewUrl.value, buildOrderCareReportFileName(props.order.order_no || `NO${props.order.ID}`, draft.value.petName))
}
</script>
```

Modify `web/src/pages/order/detail.vue`:

```vue
<button v-if="canGenerateCareReport" class="btn receipt" @click="showCareReport = true">生成报告</button>
<OrderCareReportModal
  v-if="order"
  :visible="showCareReport"
  :order="order"
  @close="showCareReport = false"
/>
```

```ts
import OrderCareReportModal from '@/components/order/OrderCareReportModal.vue'
import { canGenerateOrderCareReport } from '@/utils/order-care-report'

const showCareReport = ref(false)
const canGenerateCareReport = computed(() => canGenerateOrderCareReport(order.value))
```

Move the existing receipt save helpers into `web-image-save.ts` and switch both receipt preview and care report preview to call the same exported helpers. Do not keep two separate Safari-save implementations.

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd web && npx tsc ./src/utils/order-care-report.ts ./src/utils/web-image-save.ts ./scripts/test-order-care-report.ts --module commonjs --target ES2020 --lib ES2020,DOM --outDir ./.tmp/order-care-report-test && node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
cd web && pnpm build:h5
```

Expected:

```text
(no output from node script)
DONE  Build complete.
```

- [ ] **Step 5: Commit**

```bash
git add web/src/utils/web-image-save.ts web/src/components/order/OrderCareReportModal.vue web/src/pages/order/detail.vue web/scripts/test-order-care-report.ts
git commit -m "feat: add paid-order care report modal"
```

### Task 6: Full Verification, Deployment, and Browser Check

**Files:**
- Modify: none expected
- Verify: `server/internal/service/order_care_report_service_test.go`
- Verify: `server/internal/handler/order_care_report_test.go`
- Verify: `web/scripts/test-order-care-report.ts`

- [ ] **Step 1: Run backend verification**

Run:

```bash
cd server && go test ./internal/service ./internal/handler
```

Expected:

```text
PASS
ok  	github.com/neinei960/cat/server/internal/service
ok  	github.com/neinei960/cat/server/internal/handler
```

- [ ] **Step 2: Run frontend verification**

Run:

```bash
cd web && npx tsc ./src/utils/order-care-report.ts ./src/utils/web-image-save.ts ./scripts/test-order-care-report.ts --module commonjs --target ES2020 --lib ES2020,DOM --outDir ./.tmp/order-care-report-test && node ./.tmp/order-care-report-test/scripts/test-order-care-report.js
cd web && pnpm build:h5
```

Expected:

```text
exit 0
DONE  Build complete.
```

- [ ] **Step 3: Deploy both server and web**

Run:

```bash
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/server/internal/service/order_care_report_service.go"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
printf '{"tool_input":{"file_path":"/Users/genglsh/workstation/cat/cat/web/src/components/order/OrderCareReportModal.vue"}}' | /Users/genglsh/workstation/cat/cat/.codex/hooks/deploy.sh
```

Expected:

```text
server deploy exits 0
web deploy exits 0
```

- [ ] **Step 4: Verify rendered behavior in the browser**

Use a fresh URL and verify this exact checklist on a paid order:

```text
1. `生成报告` appears only on paid orders with at least one cat.
2. Multi-cat orders force cat selection before form fill.
3. Portrait upload and crop succeed.
4. Submitting returns a preview image.
5. `保存图片` opens the correct save flow on Safari and normal browsers.
6. The selected cat’s `洗浴报告管理` page shows the new report image.
```

- [ ] **Step 5: Record final evidence**

Capture and keep:

```text
- backend test command output
- frontend script/build output
- deploy command output
- one screenshot of the report preview
- one screenshot of the pet report page showing the generated image
```
