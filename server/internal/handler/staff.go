package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type StaffHandler struct {
	staffService *service.StaffService
}

func NewStaffHandler(staffService *service.StaffService) *StaffHandler {
	return &StaffHandler{staffService: staffService}
}

type createStaffReq struct {
	Phone                 string  `json:"phone" binding:"required"`
	Name                  string  `json:"name" binding:"required"`
	Password              string  `json:"password"`
	Role                  string  `json:"role"`
	CommissionRate        float64 `json:"commission_rate"`
	ProductCommissionRate float64 `json:"product_commission_rate"`
	FeedingCommissionRate float64 `json:"feeding_commission_rate"`
}

func (h *StaffHandler) Create(c *gin.Context) {
	var req createStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	staff := &model.Staff{
		ShopID:                c.GetUint("shop_id"),
		Phone:                 req.Phone,
		Name:                  req.Name,
		Role:                  model.NormalizeStaffRole(req.Role),
		CommissionRate:        req.CommissionRate,
		ProductCommissionRate: req.ProductCommissionRate,
		FeedingCommissionRate: req.FeedingCommissionRate,
	}

	password := req.Password
	if password == "" {
		password = "123456"
	}

	if err := h.staffService.CreateWithPassword(staff, password); err != nil {
		response.Error(c, http.StatusBadRequest, "创建失败: "+err.Error())
		return
	}
	response.Success(c, staff)
}

func (h *StaffHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	staff, err := h.staffService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "员工不存在")
		return
	}
	response.Success(c, staff)
}

func (h *StaffHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.staffService.List(shopID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

type updateStaffReq struct {
	Name                  string   `json:"name"`
	Phone                 string   `json:"phone"`
	Role                  string   `json:"role"`
	Status                int      `json:"status"`
	SortOrder             *int     `json:"sort_order"`
	CommissionRate        *float64 `json:"commission_rate"`
	ProductCommissionRate *float64 `json:"product_commission_rate"`
	FeedingCommissionRate *float64 `json:"feeding_commission_rate"`
	Avatar                string   `json:"avatar"`
}

func (h *StaffHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	staff, err := h.staffService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "员工不存在")
		return
	}

	var req updateStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if req.Name != "" {
		staff.Name = req.Name
	}
	if req.Phone != "" {
		staff.Phone = req.Phone
	}
	if req.Role != "" {
		if !model.IsValidStaffRole(req.Role) {
			response.Error(c, http.StatusBadRequest, "员工角色无效")
			return
		}
		staff.Role = req.Role
	}
	if req.Status != 0 {
		staff.Status = req.Status
	}
	if req.SortOrder != nil {
		staff.SortOrder = *req.SortOrder
	}
	if req.CommissionRate != nil {
		staff.CommissionRate = *req.CommissionRate
	}
	if req.ProductCommissionRate != nil {
		staff.ProductCommissionRate = *req.ProductCommissionRate
	}
	if req.FeedingCommissionRate != nil {
		staff.FeedingCommissionRate = *req.FeedingCommissionRate
	}
	if req.Avatar != "" {
		staff.Avatar = req.Avatar
	}

	if err := h.staffService.Update(staff); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, staff)
}

type reorderStaffReq struct {
	StaffIDs []uint `json:"staff_ids" binding:"required"`
}

func (h *StaffHandler) Reorder(c *gin.Context) {
	var req reorderStaffReq
	if err := c.ShouldBindJSON(&req); err != nil || len(req.StaffIDs) == 0 {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	seen := make(map[uint]struct{}, len(req.StaffIDs))
	for _, id := range req.StaffIDs {
		if id == 0 {
			response.Error(c, http.StatusBadRequest, "员工ID错误")
			return
		}
		if _, exists := seen[id]; exists {
			response.Error(c, http.StatusBadRequest, "员工ID重复")
			return
		}
		seen[id] = struct{}{}
	}

	if err := h.staffService.Reorder(c.GetUint("shop_id"), req.StaffIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "保存排序失败")
		return
	}
	response.Success(c, nil)
}

func (h *StaffHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.staffService.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

// Password

type resetPasswordReq struct {
	Password string `json:"password" binding:"required"`
}

func (h *StaffHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req resetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请输入新密码")
		return
	}
	if len(req.Password) < 6 {
		response.Error(c, http.StatusBadRequest, "密码至少6位")
		return
	}
	if err := h.staffService.ResetPassword(uint(id), req.Password); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// Schedule

type setScheduleReq struct {
	Date        string `json:"date" binding:"required"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	BreakStart  string `json:"break_start"`
	BreakEnd    string `json:"break_end"`
	MaxCapacity int    `json:"max_capacity"`
	IsDayOff    bool   `json:"is_day_off"`
}

func (h *StaffHandler) SetSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req setScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	schedule := &model.StaffSchedule{
		StaffID:     uint(id),
		ShopID:      c.GetUint("shop_id"),
		Date:        req.Date,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		BreakStart:  req.BreakStart,
		BreakEnd:    req.BreakEnd,
		MaxCapacity: req.MaxCapacity,
		IsDayOff:    req.IsDayOff,
	}
	if schedule.MaxCapacity == 0 {
		schedule.MaxCapacity = 1
	}

	if err := h.staffService.SetSchedule(schedule); err != nil {
		response.Error(c, http.StatusInternalServerError, "设置排班失败")
		return
	}
	response.Success(c, schedule)
}

type batchScheduleReq struct {
	Schedules []setScheduleReq `json:"schedules" binding:"required"`
}

func (h *StaffHandler) BatchSetSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req batchScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	shopID := c.GetUint("shop_id")
	schedules := make([]model.StaffSchedule, len(req.Schedules))
	for i, s := range req.Schedules {
		cap := s.MaxCapacity
		if cap == 0 {
			cap = 1
		}
		schedules[i] = model.StaffSchedule{
			StaffID:     uint(id),
			ShopID:      shopID,
			Date:        s.Date,
			StartTime:   s.StartTime,
			EndTime:     s.EndTime,
			BreakStart:  s.BreakStart,
			BreakEnd:    s.BreakEnd,
			MaxCapacity: cap,
			IsDayOff:    s.IsDayOff,
		}
	}

	if err := h.staffService.BatchSetSchedule(schedules); err != nil {
		response.Error(c, http.StatusInternalServerError, "批量设置排班失败")
		return
	}
	response.Success(c, nil)
}

func (h *StaffHandler) GetSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}

	schedules, err := h.staffService.GetSchedule(uint(id), startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, schedules)
}

// Staff services

type setServicesReq struct {
	ServiceIDs []uint `json:"service_ids" binding:"required"`
}

func (h *StaffHandler) SetServices(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req setServicesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.staffService.SetServices(uint(id), req.ServiceIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "设置技能失败")
		return
	}
	response.Success(c, nil)
}

func (h *StaffHandler) GetServices(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	services, err := h.staffService.GetServices(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, services)
}
