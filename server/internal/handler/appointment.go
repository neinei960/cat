package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type AppointmentHandler struct {
	apptService *service.AppointmentService
}

func NewAppointmentHandler(apptService *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{apptService: apptService}
}

// GET /b/appointments/slots?date=2026-03-20&service_id=1
func (h *AppointmentHandler) GetSlots(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	date := c.Query("date")
	serviceID, _ := strconv.ParseUint(c.Query("service_id"), 10, 64)
	if date == "" || serviceID == 0 {
		response.Error(c, http.StatusBadRequest, "请提供date和service_id")
		return
	}

	slots, err := h.apptService.GetAvailableSlots(shopID, date, uint(serviceID))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, slots)
}

type createApptReq struct {
	CustomerID uint   `json:"customer_id"`
	PetID      uint   `json:"pet_id"`
	StaffID    *uint  `json:"staff_id"`
	Date       string `json:"date" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"`
	ServiceIDs []uint `json:"service_ids" binding:"required"`
	Source     int    `json:"source"`
	Notes      string `json:"notes"`
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	var req createApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	appt := &model.Appointment{
		ShopID:     c.GetUint("shop_id"),
		CustomerID: req.CustomerID,
		PetID:      req.PetID,
		StaffID:    req.StaffID,
		Date:       req.Date,
		StartTime:  req.StartTime,
		Source:     req.Source,
		Notes:      req.Notes,
	}
	if appt.Source == 0 {
		appt.Source = 2 // merchant created
	}

	if err := h.apptService.Create(appt, req.ServiceIDs); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Reload with relations
	result, _ := h.apptService.GetByID(appt.ID)
	response.Success(c, result)
}

func (h *AppointmentHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	appt, err := h.apptService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "预约不存在")
		return
	}
	response.Success(c, appt)
}

// GET /b/appointments?page=1&page_size=20&status=0
func (h *AppointmentHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		status = &v
	}

	list, total, err := h.apptService.ListPaged(shopID, status, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GET /b/appointments/calendar?start_date=2026-03-20&end_date=2026-03-26
func (h *AppointmentHandler) Calendar(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}

	appts, err := h.apptService.ListByDateRange(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, appts)
}

type updateStatusReq struct {
	Status       int    `json:"status" binding:"required"`
	StaffNotes   string `json:"staff_notes"`
	CancelReason string `json:"cancel_reason"`
	CancelledBy  string `json:"cancelled_by"`
}

func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.apptService.UpdateStatus(uint(id), req.Status, req.StaffNotes, req.CancelReason, req.CancelledBy); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

type assignStaffReq struct {
	StaffID uint `json:"staff_id" binding:"required"`
}

func (h *AppointmentHandler) AssignStaff(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req assignStaffReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.apptService.AssignStaff(uint(id), req.StaffID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

type rescheduleReq struct {
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
}

func (h *AppointmentHandler) Reschedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req rescheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.apptService.Reschedule(uint(id), req.Date, req.StartTime); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}
