package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServicePriceRuleAllowsZeroPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupServiceHandlerTestDB(t)

	router := gin.New()
	serviceHandler := NewServiceHandler(service.NewServiceService(repository.NewServiceRepository()))
	router.POST("/services/:id/prices", serviceHandler.CreatePriceRule)

	rec := performServiceRequest(t, router, http.MethodPost, "/services/1/prices", map[string]any{
		"name":     "已购口腔年卡",
		"price":    0,
		"duration": 5,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected zero price rule to be accepted, got status=%d body=%s", rec.Code, rec.Body.String())
	}

	var rule model.ServicePriceRule
	if err := database.DB.Where("service_id = ? AND name = ?", 1, "已购口腔年卡").First(&rule).Error; err != nil {
		t.Fatalf("find created zero price rule: %v", err)
	}
	if rule.Price != 0 {
		t.Fatalf("expected stored price 0, got %.2f", rule.Price)
	}
}

func setupServiceHandlerTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	database.DB = db
	if err := database.DB.AutoMigrate(&model.Service{}, &model.ServicePriceRule{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
}

func performServiceRequest(t *testing.T, router *gin.Engine, method, path string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
