package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type ServiceCategoryHandler struct{}

func NewServiceCategoryHandler() *ServiceCategoryHandler {
	return &ServiceCategoryHandler{}
}

type serviceCategoryReq struct {
	Name      string `json:"name" binding:"required"`
	ParentID  *uint  `json:"parent_id"`
	SortOrder int    `json:"sort_order"`
}

func (h *ServiceCategoryHandler) Create(c *gin.Context) {
	var req serviceCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// Validate parent exists and is a top-level category (no 3rd level)
	if req.ParentID != nil {
		var parent model.ServiceCategory
		if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
			response.Error(c, http.StatusBadRequest, "父分类不存在")
			return
		}
		if parent.ParentID != nil {
			response.Error(c, http.StatusBadRequest, "最多支持二级分类")
			return
		}
	}

	cat := &model.ServiceCategory{
		ShopID:    c.GetUint("shop_id"),
		ParentID:  req.ParentID,
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

func (h *ServiceCategoryHandler) Tree(c *gin.Context) {
	shopID := c.GetUint("shop_id")

	var cats []model.ServiceCategory
	if err := database.DB.
		Where("shop_id = ? AND parent_id IS NULL AND status = 1", shopID).
		Preload("Children", "status = 1").
		Order("sort_order ASC").
		Find(&cats).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, cats)
}

func (h *ServiceCategoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var cat model.ServiceCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "分类不存在")
		return
	}

	var req serviceCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	cat.Name = req.Name
	cat.SortOrder = req.SortOrder

	if err := database.DB.Save(&cat).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败: "+err.Error())
		return
	}
	response.Success(c, cat)
}

func (h *ServiceCategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// Check for child categories
	var childCount int64
	database.DB.Model(&model.ServiceCategory{}).Where("parent_id = ?", id).Count(&childCount)
	if childCount > 0 {
		response.Error(c, http.StatusBadRequest, "该分类下有子分类，请先删除子分类")
		return
	}

	// Check for services using this category
	var svcCount int64
	database.DB.Model(&model.Service{}).Where("category_id = ?", id).Count(&svcCount)
	if svcCount > 0 {
		response.Error(c, http.StatusBadRequest, "该分类下有服务项目，请先移除或修改相关服务")
		return
	}

	if err := database.DB.Delete(&model.ServiceCategory{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}
