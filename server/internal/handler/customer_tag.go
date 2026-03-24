package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type CustomerTagHandler struct {
	service *service.CustomerTagService
}

func NewCustomerTagHandler(service *service.CustomerTagService) *CustomerTagHandler {
	return &CustomerTagHandler{service: service}
}

type customerTagReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
	Status      int    `json:"status"`
}

func (h *CustomerTagHandler) Create(c *gin.Context) {
	var req customerTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	tag := &model.CustomerTag{
		ShopID:      c.GetUint("shop_id"),
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}
	if tag.Color == "" {
		tag.Color = "#4F46E5"
	}
	if tag.Status == 0 {
		tag.Status = 1
	}

	if err := h.service.Create(tag); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, tag)
}

func (h *CustomerTagHandler) List(c *gin.Context) {
	list, err := h.service.List(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, list)
}

func (h *CustomerTagHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	tag, err := h.service.GetByID(uint(id), c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "标签不存在")
		return
	}

	var req customerTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	tag.Name = req.Name
	tag.Description = req.Description
	if req.Color != "" {
		tag.Color = req.Color
	}
	tag.SortOrder = req.SortOrder
	if req.Status == 0 {
		tag.Status = 0
	} else {
		tag.Status = 1
	}

	if err := h.service.Update(tag); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, tag)
}

func (h *CustomerTagHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(uint(id), c.GetUint("shop_id")); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}
