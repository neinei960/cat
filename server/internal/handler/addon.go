package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type AddonHandler struct{}

func NewAddonHandler() *AddonHandler {
	return &AddonHandler{}
}

type createAddonReq struct {
	Name         string  `json:"name" binding:"required"`
	DefaultPrice float64 `json:"default_price"`
	IsVariable   bool    `json:"is_variable"`
	SortOrder    int     `json:"sort_order"`
}

func (h *AddonHandler) Create(c *gin.Context) {
	var req createAddonReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	addon := &model.ServiceAddon{
		ShopID:       c.GetUint("shop_id"),
		Name:         req.Name,
		DefaultPrice: req.DefaultPrice,
		IsVariable:   req.IsVariable,
		SortOrder:    req.SortOrder,
		Status:       1,
	}

	if err := database.DB.Create(addon).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, addon)
}

func (h *AddonHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	var addons []model.ServiceAddon
	if err := database.DB.Where("shop_id = ? AND status = 1", shopID).Order("sort_order ASC").Find(&addons).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, addons)
}

func (h *AddonHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var addon model.ServiceAddon
	if err := database.DB.First(&addon, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "附加项不存在")
		return
	}

	var req createAddonReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	addon.Name = req.Name
	addon.DefaultPrice = req.DefaultPrice
	addon.IsVariable = req.IsVariable
	addon.SortOrder = req.SortOrder

	if err := database.DB.Save(&addon).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, addon)
}

func (h *AddonHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := database.DB.Delete(&model.ServiceAddon{}, id).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}
