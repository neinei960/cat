package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/handler"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/router"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var orderCareReportTestSeq uint64

func TestCreateOrderCareReportReturns400ForUnpaidOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOrderCareReportHandlerTestDB(t)

	state := seedOrderCareReportHandlerFixture(t, orderCareReportHandlerFixtureInput{
		ShopID:           1,
		OrderNo:          "TEST-HANDLER-CARE-REPORT-UNPAID",
		CustomerPhone:    "13800138301",
		CustomerNickname: "未支付护理客户",
		PetName:          "未支付猫咪",
		Paid:             false,
	})

	handler := newOrderCareReportTestHandler(nil)
	router := newOrderCareReportTestRouter(handler)

	rec := performOrderCareReportRequest(t, router, http.MethodPost, fmt.Sprintf("/b/orders/%d/care-report", state.order.ID), map[string]any{
		"pet_id":         state.pet.ID,
		"portrait_url":   "/uploads/test-portrait.jpg",
		"weight":         "4.2kg",
		"care_date":      "2026-04-20",
		"next_care_date": "2026-04-25",
		"care_content":   "护理记录",
		"body_shape":     "standard",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	resp := decodeOrderCareReportResponse(t, rec)
	if resp.Msg != "仅已支付订单可生成报告" {
		t.Fatalf("expected unpaid-order message, got %q", resp.Msg)
	}
}

func TestCreateOrderCareReportReturns400ForBindingFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOrderCareReportHandlerTestDB(t)

	handler := newOrderCareReportTestHandler(nil)
	router := newOrderCareReportTestRouter(handler)

	rec := performOrderCareReportRequest(t, router, http.MethodPost, "/b/orders/1/care-report", map[string]any{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	resp := decodeOrderCareReportResponse(t, rec)
	if resp.Msg != "参数错误" {
		t.Fatalf("expected binding error message, got %q", resp.Msg)
	}
}

func TestCreateOrderCareReportReturns500ForUnexpectedServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOrderCareReportHandlerTestDB(t)

	handler := newOrderCareReportTestHandler(failingCareReportService{err: fmt.Errorf("boom")})
	router := newOrderCareReportTestRouter(handler)

	rec := performOrderCareReportRequest(t, router, http.MethodPost, "/b/orders/1/care-report", map[string]any{
		"pet_id":         1,
		"portrait_url":   "/uploads/test-portrait.jpg",
		"weight":         "4.2kg",
		"care_date":      "2026-04-20",
		"next_care_date": "2026-04-25",
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}

	resp := decodeOrderCareReportResponse(t, rec)
	if resp.Msg != "生成护理报告失败" {
		t.Fatalf("expected internal error message, got %q", resp.Msg)
	}
}

func TestCreateOrderCareReportPassesDisplayOverridesToService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOrderCareReportHandlerTestDB(t)

	capturingService := &capturingCareReportService{}
	handler := newOrderCareReportTestHandler(capturingService)
	router := newOrderCareReportTestRouter(handler)

	rec := performOrderCareReportRequest(t, router, http.MethodPost, "/b/orders/7/care-report", map[string]any{
		"pet_id":         12,
		"pet_name":       "报告猫咪",
		"breed":          "金吉拉",
		"gender":         "妹妹",
		"age":            "2岁1个月",
		"portrait_url":   "/uploads/test-portrait.jpg",
		"weight":         "4.2kg",
		"care_date":      "2026-07-16",
		"next_care_date": "2026-08-16",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	input := capturingService.input
	if input.PetName == nil || *input.PetName != "报告猫咪" {
		t.Fatalf("unexpected pet name override: %#v", input.PetName)
	}
	if input.Breed == nil || *input.Breed != "金吉拉" {
		t.Fatalf("unexpected breed override: %#v", input.Breed)
	}
	if input.Gender == nil || *input.Gender != "妹妹" {
		t.Fatalf("unexpected gender override: %#v", input.Gender)
	}
	if input.Age == nil || *input.Age != "2岁1个月" {
		t.Fatalf("unexpected age override: %#v", input.Age)
	}
}

func TestRouterRegistersOrderCareReportRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := router.Setup(gin.TestMode)
	for _, route := range r.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/b/orders/:id/care-report" {
			return
		}
	}

	t.Fatalf("expected POST /api/v1/b/orders/:id/care-report to be registered")
}

func TestCreateOrderCareReportRejectsInvalidOrderID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupOrderCareReportHandlerTestDB(t)

	handler := newOrderCareReportTestHandler(nil)
	router := newOrderCareReportTestRouter(handler)

	rec := performOrderCareReportRequest(t, router, http.MethodPost, "/b/orders/0/care-report", map[string]any{
		"pet_id":         1,
		"portrait_url":   "/uploads/test-portrait.jpg",
		"weight":         "4.2kg",
		"care_date":      "2026-04-20",
		"next_care_date": "2026-04-25",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}

	resp := decodeOrderCareReportResponse(t, rec)
	if resp.Msg != "订单ID错误" {
		t.Fatalf("expected invalid-order-id message, got %q", resp.Msg)
	}
}

func setupOrderCareReportHandlerTestDB(t *testing.T) {
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
		&model.Order{},
		&model.OrderItem{},
		&model.PetBathReport{},
	); err != nil {
		t.Fatalf("auto migrate order care report handler tables: %v", err)
	}
}

type orderCareReportHandlerFixtureInput struct {
	ShopID           uint
	OrderNo          string
	CustomerPhone    string
	CustomerNickname string
	PetName          string
	Paid             bool
}

type orderCareReportHandlerFixture struct {
	shopID   uint
	customer model.Customer
	pet      model.Pet
	order    model.Order
}

func seedOrderCareReportHandlerFixture(t *testing.T, input orderCareReportHandlerFixtureInput) orderCareReportHandlerFixture {
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

	pet := model.Pet{
		ShopID:     input.ShopID,
		CustomerID: &customer.ID,
		Name:       input.PetName,
		Species:    "猫",
		Breed:      "布偶",
		Gender:     2,
	}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("create care report pet: %v", err)
	}

	order := model.Order{
		OrderNo:      uniqueOrderCareReportHandlerOrderNo(input.OrderNo),
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

	return orderCareReportHandlerFixture{
		shopID:   input.ShopID,
		customer: customer,
		pet:      pet,
		order:    order,
	}
}

func uniqueOrderCareReportHandlerOrderNo(base string) string {
	seq := atomic.AddUint64(&orderCareReportTestSeq, 1)
	return fmt.Sprintf("%s-%d", base, seq)
}

func newOrderCareReportTestHandler(careReportService interface {
	Create(shopID, orderID uint, input service.CreateOrderCareReportInput) (*service.OrderCareReportResult, error)
}) *handler.OrderHandler {
	orderRepo := repository.NewOrderRepository()
	petRepo := repository.NewPetRepository()
	customerRepo := repository.NewCustomerRepository()
	serviceRepo := repository.NewServiceRepository()
	if careReportService == nil {
		careReportService = service.NewOrderCareReportService(orderRepo, repository.NewPetBathReportRepository())
	}
	return handler.NewOrderHandler(
		service.NewOrderService(orderRepo, repository.NewAppointmentRepository()),
		service.NewPetService(petRepo),
		service.NewCustomerService(customerRepo),
		service.NewServiceService(serviceRepo),
		careReportService,
	)
}

type failingCareReportService struct {
	err error
}

func (s failingCareReportService) Create(shopID, orderID uint, input service.CreateOrderCareReportInput) (*service.OrderCareReportResult, error) {
	return nil, s.err
}

type capturingCareReportService struct {
	input service.CreateOrderCareReportInput
}

func (s *capturingCareReportService) Create(shopID, orderID uint, input service.CreateOrderCareReportInput) (*service.OrderCareReportResult, error) {
	s.input = input
	return &service.OrderCareReportResult{}, nil
}

func newOrderCareReportTestRouter(orderHandler *handler.OrderHandler) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("shop_id", uint(1))
		c.Next()
	})
	router.POST("/b/orders/:id/care-report", orderHandler.CreateCareReport)
	return router
}

func performOrderCareReportRequest(t *testing.T, router *gin.Engine, method, path string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeOrderCareReportResponse(t *testing.T, rec *httptest.ResponseRecorder) response.Response {
	t.Helper()

	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}
