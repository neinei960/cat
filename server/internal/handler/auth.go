package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type staffLoginReq struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) StaffLogin(c *gin.Context) {
	var req staffLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := h.authService.StaffLogin(req.Phone, req.Password)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	response.Success(c, result)
}

type wxLoginReq struct {
	Code   string `json:"code" binding:"required"`
	ShopID uint   `json:"shop_id" binding:"required"`
}

func (h *AuthHandler) WxLogin(c *gin.Context) {
	var req wxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := h.authService.WxLogin(req.Code, req.ShopID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}

type wxBindPhoneReq struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *AuthHandler) WxBindPhone(c *gin.Context) {
	var req wxBindPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	customerID, _ := c.Get("customer_id")
	if err := h.authService.WxBindPhone(customerID.(uint), req.Phone); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}
