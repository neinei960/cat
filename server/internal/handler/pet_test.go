package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

func TestParseBirthDateSupportsSingleDigitMonthAndDay(t *testing.T) {
	birthDate := parseBirthDate("2019-11-4")
	if birthDate == nil {
		t.Fatalf("expected single-digit day birth date to parse")
	}
	if got := birthDate.Format("2006-01-02"); got != "2019-11-04" {
		t.Fatalf("expected normalized birth date 2019-11-04, got %s", got)
	}
}

func TestPetPersonalityFieldAllowsLongParsedText(t *testing.T) {
	field, ok := reflect.TypeOf(model.Pet{}).FieldByName("Personality")
	if !ok {
		t.Fatalf("Pet.Personality field not found")
	}
	if got := field.Tag.Get("gorm"); !strings.Contains(got, "size:100") {
		t.Fatalf("expected personality field gorm tag to include size:100, got %q", got)
	}
}

func TestPetCreateWithoutOwnerPhoneCreatesLinkedCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	router := newPetTestRouter()
	rec := performCustomerRequest(t, router, http.MethodPost, "/pets", map[string]any{
		"name": "奶糖",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		Code int       `json:"code"`
		Data model.Pet `json:"data"`
		Msg  string    `json:"msg"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode pet response: %v", err)
	}
	if resp.Data.CustomerID == nil || *resp.Data.CustomerID == 0 {
		t.Fatalf("expected created pet to link generated customer, got customer_id=%v", resp.Data.CustomerID)
	}

	var customer model.Customer
	if err := database.DB.First(&customer, *resp.Data.CustomerID).Error; err != nil {
		t.Fatalf("find generated customer: %v", err)
	}
	if customer.Phone == "" {
		t.Fatalf("expected generated customer phone to be filled")
	}
}

func TestPetUpdateWithBlankOwnerPhoneKeepsExistingCustomer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	customer := seedCustomer(t, 1, "13800138001", "原主人")
	pet := model.Pet{ShopID: 1, CustomerID: &customer.ID, Name: "奶糖", Species: "猫"}
	if err := database.DB.Create(&pet).Error; err != nil {
		t.Fatalf("seed pet: %v", err)
	}

	router := newPetTestRouter()
	rec := performCustomerRequest(t, router, http.MethodPut, fmt.Sprintf("/pets/%d", pet.ID), map[string]any{
		"name":        "奶糖",
		"owner_phone": "",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var updated model.Pet
	if err := database.DB.First(&updated, pet.ID).Error; err != nil {
		t.Fatalf("find updated pet: %v", err)
	}
	if updated.CustomerID == nil || *updated.CustomerID != customer.ID {
		t.Fatalf("expected customer link to stay %d, got %v", customer.ID, updated.CustomerID)
	}
}

func TestPetListShortNumericKeywordDoesNotMatchPhoneOrNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setupCustomerTestDB(t)

	customer := seedCustomer(t, 1, "15600000011", "黄大宝")
	numericNameCustomer := seedCustomer(t, 1, "15108328921", "AmyL-0119")
	pets := []model.Pet{
		{Model: gorm.Model{ID: 11}, ShopID: 1, CustomerID: &customer.ID, Name: "福福", Species: "猫"},
		{Model: gorm.Model{ID: 12}, ShopID: 1, CustomerID: &customer.ID, Name: "奈奈", Species: "猫"},
		{ShopID: 1, CustomerID: &numericNameCustomer.ID, Name: "牛奶", Species: "猫"},
		{ShopID: 1, Name: "肉粽", Species: "猫", CareNotes: "25.11月来剃毛有脾气"},
	}
	for _, pet := range pets {
		if err := database.DB.Create(&pet).Error; err != nil {
			t.Fatalf("seed pet %s: %v", pet.Name, err)
		}
	}

	router := newPetTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/pets?keyword=11", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			List []model.Pet `json:"list"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode pet list response: %v", err)
	}
	if len(resp.Data.List) != 1 {
		t.Fatalf("expected only exact numeric pet match, got %d pets: %+v", len(resp.Data.List), resp.Data.List)
	}
	if resp.Data.List[0].ID != 11 {
		t.Fatalf("expected pet ID 11, got ID=%d name=%s", resp.Data.List[0].ID, resp.Data.List[0].Name)
	}
}

func newPetTestRouter() *gin.Engine {
	customerRepo := repository.NewCustomerRepository()
	petRepo := repository.NewPetRepository()
	handler := NewPetHandler(service.NewPetService(petRepo), service.NewCustomerService(customerRepo))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("shop_id", uint(1))
		c.Next()
	})
	router.GET("/pets", handler.List)
	router.POST("/pets", handler.Create)
	router.PUT("/pets/:id", handler.Update)
	return router
}
