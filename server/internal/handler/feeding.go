package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type FeedingHandler struct {
	service *service.FeedingService
}

func NewFeedingHandler(service *service.FeedingService) *FeedingHandler {
	return &FeedingHandler{service: service}
}

func (h *FeedingHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, settings)
}

func (h *FeedingHandler) UpdatePricing(c *gin.Context) {
	var req service.FeedingPricingSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	settings, err := h.service.UpdatePricing(c.GetUint("shop_id"), c.GetUint("staff_id"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, settings)
}

func (h *FeedingHandler) UpdateItems(c *gin.Context) {
	var req []service.FeedingItemTemplate
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	settings, err := h.service.UpdateItems(c.GetUint("shop_id"), c.GetUint("staff_id"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, settings)
}

func (h *FeedingHandler) CreatePlan(c *gin.Context) {
	var req service.FeedingPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	plan, err := h.service.CreatePlan(c.GetUint("shop_id"), c.GetUint("staff_id"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) ListPlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	customerID, _ := strconv.ParseUint(c.DefaultQuery("customer_id", "0"), 10, 64)
	result, err := h.service.ListPlans(c.GetUint("shop_id"), page, pageSize, c.Query("status"), uint(customerID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, result)
}

func (h *FeedingHandler) GetPlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	plan, err := h.service.GetPlan(c.GetUint("shop_id"), uint(id), roleText, c.GetUint("staff_id"))
	if err != nil {
		response.Error(c, http.StatusNotFound, "计划不存在")
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) UpdatePlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.FeedingPlanInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	plan, err := h.service.UpdatePlan(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) PausePlan(c *gin.Context) {
	h.changePlanStatus(c, "pause")
}

func (h *FeedingHandler) ResumePlan(c *gin.Context) {
	h.changePlanStatus(c, "resume")
}

func (h *FeedingHandler) CancelPlan(c *gin.Context) {
	h.changePlanStatus(c, "cancel")
}

func (h *FeedingHandler) changePlanStatus(c *gin.Context, action string) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var (
		plan *model.FeedingPlan
		err  error
	)
	switch action {
	case "pause":
		plan, err = h.service.PausePlan(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id))
	case "resume":
		plan, err = h.service.ResumePlan(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id))
	case "cancel":
		plan, err = h.service.CancelPlan(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id))
	}
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) GenerateOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.service.GenerateOrder(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *FeedingHandler) UpdateDeposit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input struct {
		Deposit float64 `json:"deposit"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	plan, err := h.service.UpdateDeposit(c.GetUint("shop_id"), uint(id), input.Deposit)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) UpdatePlayDates(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.FeedingUpdatePlayDatesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	plan, err := h.service.UpdatePlayDates(c.GetUint("shop_id"), uint(id), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, plan)
}

func (h *FeedingHandler) Dashboard(c *gin.Context) {
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	staffID, _ := strconv.ParseUint(c.DefaultQuery("staff_id", "0"), 10, 64)
	data, err := h.service.GetDashboard(
		c.GetUint("shop_id"),
		roleText,
		c.GetUint("staff_id"),
		c.Query("date"),
		uint(staffID),
		c.Query("window_code"),
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, data)
}

func (h *FeedingHandler) ListVisits(c *gin.Context) {
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	id, _ := strconv.ParseUint(c.DefaultQuery("id", "0"), 10, 64)
	planID, _ := strconv.ParseUint(c.DefaultQuery("plan_id", "0"), 10, 64)
	staffID, _ := strconv.ParseUint(c.DefaultQuery("staff_id", "0"), 10, 64)
	visits, err := h.service.ListVisits(c.GetUint("shop_id"), roleText, c.GetUint("staff_id"), service.FeedingVisitFilterInput{
		ID:            uint(id),
		PlanID:        uint(planID),
		ScheduledDate: c.Query("scheduled_date"),
		Status:        c.Query("status"),
		StaffID:       uint(staffID),
		WindowCode:    c.Query("window_code"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visits)
}

func (h *FeedingHandler) AssignVisit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req service.FeedingAssignVisitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	visit, err := h.service.AssignVisit(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visit)
}

func (h *FeedingHandler) UpdateVisitNote(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	var req service.FeedingUpdateVisitNoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	visit, err := h.service.UpdateVisitNote(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), roleText, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visit)
}

func (h *FeedingHandler) StartVisit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	visit, err := h.service.StartVisit(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), roleText)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visit)
}

func (h *FeedingHandler) CompleteVisit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	var req service.FeedingCompleteVisitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	visit, err := h.service.CompleteVisit(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), roleText, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visit)
}

func (h *FeedingHandler) ExceptionVisit(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	var req service.FeedingExceptionVisitInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	visit, err := h.service.SetVisitException(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), roleText, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, visit)
}

func (h *FeedingHandler) AddVisitMedia(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	role, _ := c.Get("role")
	roleText, _ := role.(string)
	var req service.FeedingVisitMediaInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	media, err := h.service.AddVisitMedia(c.GetUint("shop_id"), c.GetUint("staff_id"), uint(id), roleText, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, media)
}
