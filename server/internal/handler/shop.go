package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type ShopHandler struct {
	shopService *service.ShopService
}

func NewShopHandler(shopService *service.ShopService) *ShopHandler {
	return &ShopHandler{shopService: shopService}
}

func (h *ShopHandler) Get(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	shop, err := h.shopService.GetByID(shopID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "店铺不存在")
		return
	}
	response.Success(c, shop)
}

type updateShopReq struct {
	Name          string `json:"name"`
	Logo          string `json:"logo"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	BusinessHours interface{} `json:"business_hours"`
}

func (h *ShopHandler) Update(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	shop, err := h.shopService.GetByID(shopID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "店铺不存在")
		return
	}

	var req updateShopReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Name != "" {
		shop.Name = req.Name
	}
	if req.Logo != "" {
		shop.Logo = req.Logo
	}
	if req.Phone != "" {
		shop.Phone = req.Phone
	}
	if req.Address != "" {
		shop.Address = req.Address
	}
	if req.Latitude != 0 {
		shop.Latitude = req.Latitude
	}
	if req.Longitude != 0 {
		shop.Longitude = req.Longitude
	}

	if err := h.shopService.Update(shop); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, shop)
}
