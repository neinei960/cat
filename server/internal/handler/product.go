package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

type skuInput struct {
	SpecName string  `json:"spec_name"`
	Price    float64 `json:"price"`
	Weight   float64 `json:"weight"`
	Sellable bool    `json:"sellable"`
}

type productReq struct {
	Name        string     `json:"name" binding:"required"`
	CategoryID  uint       `json:"category_id"`
	Brand       string     `json:"brand"`
	Description string     `json:"description"`
	MultiSpec   bool       `json:"multi_spec"`
	SKUs        []skuInput `json:"skus"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	product := &model.Product{
		ShopID:      c.GetUint("shop_id"),
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Brand:       req.Brand,
		Description: req.Description,
		MultiSpec:   req.MultiSpec,
	}

	for _, s := range req.SKUs {
		product.SKUs = append(product.SKUs, model.ProductSKU{
			SpecName: s.SpecName,
			Price:    s.Price,
			Weight:   s.Weight,
			Sellable: s.Sellable,
		})
	}

	if err := h.productService.Create(product); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, product)
}

func (h *ProductHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "商品不存在")
		return
	}
	response.Success(c, product)
}

func (h *ProductHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	keyword := c.Query("keyword")

	list, total, err := h.productService.List(shopID, uint(categoryID), keyword, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "商品不存在")
		return
	}

	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	product.Name = req.Name
	product.CategoryID = req.CategoryID
	product.Brand = req.Brand
	product.Description = req.Description
	product.MultiSpec = req.MultiSpec

	if err := h.productService.Update(product); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}

	// Replace SKUs
	if req.SKUs != nil {
		var skus []model.ProductSKU
		for _, s := range req.SKUs {
			skus = append(skus, model.ProductSKU{
				SpecName: s.SpecName,
				Price:    s.Price,
				Weight:   s.Weight,
				Sellable: s.Sellable,
			})
		}
		if err := h.productService.ReplaceSKUs(product.ID, skus); err != nil {
			response.Error(c, http.StatusInternalServerError, "更新规格失败")
			return
		}
	}

	// Reload
	product, _ = h.productService.GetByID(product.ID)
	response.Success(c, product)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.productService.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

func (h *ProductHandler) GetBrands(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	brands, err := h.productService.GetBrands(shopID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, brands)
}
