package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type DashboardHandler struct {
	dashService *service.DashboardService
}

func NewDashboardHandler(dashService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashService: dashService}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var stats interface{}
	var err error
	if startDate != "" && endDate != "" {
		stats, err = h.dashService.GetOverviewByRange(shopID, startDate, endDate)
	} else {
		stats, err = h.dashService.GetOverview(shopID)
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, stats)
}

func (h *DashboardHandler) Revenue(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}
	data, err := h.dashService.GetRevenueChart(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *DashboardHandler) ServiceRanking(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}
	data, err := h.dashService.GetServiceRanking(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *DashboardHandler) StaffPerformance(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}
	data, err := h.dashService.GetStaffPerformance(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, data)
}

// GET /b/dashboard/category — 洗浴分类统计（项目×毛发等级）
func (h *DashboardHandler) CategoryStats(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}
	data, err := h.dashService.GetCategoryStats(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *DashboardHandler) Aggregate(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	date := c.Query("date")
	if date == "" {
		response.Error(c, http.StatusBadRequest, "请提供date")
		return
	}
	if err := h.dashService.AggregateDaily(shopID, date); err != nil {
		response.Error(c, http.StatusInternalServerError, "聚合失败")
		return
	}
	response.Success(c, nil)
}
