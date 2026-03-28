package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type ServiceHandler struct {
	serviceService *service.ServiceService
}

func NewServiceHandler(serviceService *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{serviceService: serviceService}
}

type createServiceReq struct {
	Name              string  `json:"name" binding:"required"`
	Category          string  `json:"category"`
	CategoryID        *uint   `json:"category_id"`
	Description       string  `json:"description"`
	BasePrice         float64 `json:"base_price" binding:"required"`
	Duration          int     `json:"duration"`
	PricingType       int     `json:"pricing_type"`
	HolidayPrice      float64 `json:"holiday_price"`
	ApplicableSpecies string  `json:"applicable_species"`
	ApplicableSizes   string  `json:"applicable_sizes"`
	SortOrder         int     `json:"sort_order"`
}

func (h *ServiceHandler) Create(c *gin.Context) {
	var req createServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	pricingType := req.PricingType
	if pricingType == 0 {
		pricingType = 1
	}
	svc := &model.Service{
		ShopID:            c.GetUint("shop_id"),
		Name:              req.Name,
		Category:          req.Category,
		CategoryID:        req.CategoryID,
		Description:       req.Description,
		BasePrice:         req.BasePrice,
		Duration:          req.Duration,
		PricingType:       pricingType,
		HolidayPrice:      req.HolidayPrice,
		ApplicableSpecies: req.ApplicableSpecies,
		ApplicableSizes:   req.ApplicableSizes,
		SortOrder:         req.SortOrder,
	}

	if err := h.serviceService.Create(svc); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, svc)
}

func (h *ServiceHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	svc, err := h.serviceService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "服务不存在")
		return
	}
	response.Success(c, svc)
}

func (h *ServiceHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.serviceService.List(shopID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	svc, err := h.serviceService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "服务不存在")
		return
	}

	var req createServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	svc.Name = req.Name
	svc.Category = req.Category
	svc.CategoryID = req.CategoryID
	svc.Description = req.Description
	svc.BasePrice = req.BasePrice
	svc.Duration = req.Duration
	if req.PricingType > 0 {
		svc.PricingType = req.PricingType
	}
	svc.HolidayPrice = req.HolidayPrice
	svc.ApplicableSpecies = req.ApplicableSpecies
	svc.ApplicableSizes = req.ApplicableSizes
	svc.SortOrder = req.SortOrder

	if err := h.serviceService.Update(svc); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, svc)
}

func (h *ServiceHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.serviceService.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

// Price rules

type createPriceRuleReq struct {
	Name     string  `json:"name"`
	FurLevel string  `json:"fur_level"`
	PetSize  string  `json:"pet_size"`
	Breed    string  `json:"breed"`
	Price    float64 `json:"price" binding:"required"`
	Duration int     `json:"duration"`
}

func (h *ServiceHandler) CreatePriceRule(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req createPriceRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	rule := &model.ServicePriceRule{
		ServiceID: uint(serviceID),
		Name:      req.Name,
		FurLevel:  req.FurLevel,
		PetSize:   req.PetSize,
		Breed:     req.Breed,
		Price:     req.Price,
		Duration:  req.Duration,
	}

	if err := h.serviceService.CreatePriceRule(rule); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, rule)
}

func (h *ServiceHandler) GetPriceRules(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	rules, err := h.serviceService.GetPriceRules(uint(serviceID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, rules)
}

func (h *ServiceHandler) DeletePriceRule(c *gin.Context) {
	ruleID, _ := strconv.ParseUint(c.Param("rule_id"), 10, 64)
	if err := h.serviceService.DeletePriceRule(uint(ruleID)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

// Discounts

type createDiscountReq struct {
	Type          int     `json:"type" binding:"required"`
	MinDays       int     `json:"min_days" binding:"required"`
	DiscountPrice float64 `json:"discount_price"`
	FreeDays      int     `json:"free_days"`
	IsHoliday     bool    `json:"is_holiday"`
}

func (h *ServiceHandler) CreateDiscount(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req createDiscountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	d := &model.ServiceDiscount{
		ServiceID:     uint(serviceID),
		Type:          req.Type,
		MinDays:       req.MinDays,
		DiscountPrice: req.DiscountPrice,
		FreeDays:      req.FreeDays,
		IsHoliday:     req.IsHoliday,
	}

	if err := h.serviceService.CreateDiscount(d); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, d)
}

func (h *ServiceHandler) GetDiscounts(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	discounts, err := h.serviceService.GetDiscounts(uint(serviceID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, discounts)
}

func (h *ServiceHandler) DeleteDiscount(c *gin.Context) {
	discountID, _ := strconv.ParseUint(c.Param("discount_id"), 10, 64)
	if err := h.serviceService.DeleteDiscount(uint(discountID)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}
