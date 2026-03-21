package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type ProductCategoryHandler struct{}

func NewProductCategoryHandler() *ProductCategoryHandler {
	return &ProductCategoryHandler{}
}

type productCategoryReq struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status"`
}

func (h *ProductCategoryHandler) Create(c *gin.Context) {
	var req productCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	cat := &model.ProductCategory{
		ShopID:    c.GetUint("shop_id"),
		Name:      req.Name,
		SortOrder: req.SortOrder,
		Status:    1,
	}

	if err := database.DB.Create(cat).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败: "+err.Error())
		return
	}
	response.Success(c, cat)
}

func (h *ProductCategoryHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	var cats []model.ProductCategory
	if err := database.DB.Where("shop_id = ?", shopID).Order("sort_order ASC").Find(&cats).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, cats)
}

func (h *ProductCategoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var cat model.ProductCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "分类不存在")
		return
	}

	var req productCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	cat.Name = req.Name
	cat.SortOrder = req.SortOrder
	if req.Status > 0 {
		cat.Status = req.Status
	}

	if err := database.DB.Save(&cat).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	response.Success(c, cat)
}

func (h *ProductCategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// 检查是否有商品使用该分类
	var count int64
	database.DB.Model(&model.Product{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		response.Error(c, http.StatusBadRequest, "该分类下有商品，无法删除")
		return
	}

	if err := database.DB.Delete(&model.ProductCategory{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}
