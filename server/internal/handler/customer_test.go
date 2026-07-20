package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCustomerCreateRejectsDuplicatePhoneInSameShop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	seedCustomer(t, 1, "13800138000", "已有客户")

	handler := newCustomerTestHandler()
	router := newCustomerTestRouter(handler)

	body := map[string]any{
		"phone":    "13800138000",
		"nickname": "重复客户",
	}
	rec := performCustomerRequest(t, router, http.MethodPost, "/customers", body)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rec.Code, rec.Body.String())
	}

	resp := decodeCustomerResponse(t, rec)
	if resp.Msg != "该手机号客户已存在" {
		t.Fatalf("expected duplicate phone message, got %q", resp.Msg)
	}
}

func TestCustomerUpdateRejectsDuplicatePhoneInSameShop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	existing := seedCustomer(t, 1, "13800138000", "已有客户")
	target := seedCustomer(t, 1, "13900139000", "待修改客户")

	handler := newCustomerTestHandler()
	router := newCustomerTestRouter(handler)

	body := map[string]any{
		"phone":    existing.Phone,
		"nickname": target.Nickname,
	}
	rec := performCustomerRequest(t, router, http.MethodPut, fmt.Sprintf("/customers/%d", target.ID), body)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusConflict, rec.Code, rec.Body.String())
	}

	resp := decodeCustomerResponse(t, rec)
	if resp.Msg != "该手机号客户已存在" {
		t.Fatalf("expected duplicate phone message, got %q", resp.Msg)
	}
}

func TestCustomerUpdateAllowsKeepingOwnPhone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	target := seedCustomer(t, 1, "13800138000", "待修改客户")

	handler := newCustomerTestHandler()
	router := newCustomerTestRouter(handler)

	body := map[string]any{
		"phone":    target.Phone,
		"nickname": "改后昵称",
	}
	rec := performCustomerRequest(t, router, http.MethodPut, fmt.Sprintf("/customers/%d", target.ID), body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	resp := decodeCustomerResponse(t, rec)
	if resp.Code != 0 {
		t.Fatalf("expected success response, got code=%d msg=%q", resp.Code, resp.Msg)
	}
}

func TestCustomerCreateGeneratesPhoneWhenBlank(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	handler := newCustomerTestHandler()
	router := newCustomerTestRouter(handler)

	rec := performCustomerRequest(t, router, http.MethodPost, "/customers", map[string]any{
		"phone":    "",
		"nickname": "无手机号客户",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var customer model.Customer
	if err := database.DB.Where("nickname = ?", "无手机号客户").First(&customer).Error; err != nil {
		t.Fatalf("find created customer: %v", err)
	}
	if customer.Phone == "" {
		t.Fatalf("expected generated phone")
	}
}

func newCustomerTestHandler() *CustomerHandler {
	customerRepo := repository.NewCustomerRepository()
	petRepo := repository.NewPetRepository()
	return NewCustomerHandler(service.NewCustomerService(customerRepo), service.NewPetService(petRepo))
}

func newCustomerTestRouter(handler *CustomerHandler) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("shop_id", uint(1))
		c.Next()
	})
	router.POST("/customers", handler.Create)
	router.PUT("/customers/:id", handler.Update)
	return router
}

func performCustomerRequest(t *testing.T, router *gin.Engine, method, path string, body map[string]any) *httptest.ResponseRecorder {
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

func decodeCustomerResponse(t *testing.T, rec *httptest.ResponseRecorder) response.Response {
	t.Helper()

	var resp response.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func setupCustomerTestDB(t *testing.T) {
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
		&model.MemberCard{},
		&model.MemberCardTemplate{},
		&model.CustomerTag{},
		&model.CustomerTagRelation{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func seedCustomer(t *testing.T, shopID uint, phone, nickname string) *model.Customer {
	t.Helper()

	customer := &model.Customer{
		ShopID:       shopID,
		Phone:        phone,
		Nickname:     nickname,
		DiscountRate: 1,
	}
	if err := database.DB.Create(customer).Error; err != nil {
		t.Fatalf("seed customer: %v", err)
	}
	return customer
}
