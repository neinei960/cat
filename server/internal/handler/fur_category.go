package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type FurCategoryHandler struct{}

func NewFurCategoryHandler() *FurCategoryHandler {
	return &FurCategoryHandler{}
}

type furCategoryReq struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Status    int    `json:"status"`
}

func (h *FurCategoryHandler) Create(c *gin.Context) {
	var req furCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	cat := &model.FurCategory{
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

func (h *FurCategoryHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	var cats []model.FurCategory
	if err := database.DB.Where("shop_id = ? AND status = 1", shopID).Order("sort_order ASC").Find(&cats).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败: "+err.Error())
		return
	}
	response.Success(c, cats)
}

func (h *FurCategoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var cat model.FurCategory
	if err := database.DB.First(&cat, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "类别不存在")
		return
	}

	var req furCategoryReq
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

func (h *FurCategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := database.DB.Delete(&model.FurCategory{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}
