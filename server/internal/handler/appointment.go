package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	serviceIDs := parseServiceIDs(c.Query("service_ids"))
	duration, _ := strconv.Atoi(c.DefaultQuery("duration", "0"))
	serviceID, _ := strconv.ParseUint(c.Query("service_id"), 10, 64)
	excludeID, _ := strconv.ParseUint(c.Query("exclude_id"), 10, 64)
	if date == "" || (len(serviceIDs) == 0 && serviceID == 0) {
		response.Error(c, http.StatusBadRequest, "请提供date和service_ids")
		return
	}

	var (
		slots []service.StaffSlots
		err   error
	)
	if len(serviceIDs) > 0 {
		slots, err = h.apptService.GetAvailableSlotsByServicesExcluding(shopID, date, serviceIDs, duration, uint(excludeID))
	} else {
		slots, err = h.apptService.GetAvailableSlotsExcluding(shopID, date, uint(serviceID), uint(excludeID))
	}
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, slots)
}

type createApptReq struct {
	CustomerID uint `json:"customer_id"`
	PetID      uint `json:"pet_id"`
	Pets       []struct {
		PetID      uint   `json:"pet_id"`
		ServiceIDs []uint `json:"service_ids"`
	} `json:"pets"`
	StaffID    *uint  `json:"staff_id"`
	Date       string `json:"date" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time"`
	ServiceIDs []uint `json:"service_ids"`
	Source     int    `json:"source"`
	Notes      string `json:"notes"`
}

func buildPetSelections(req createApptReq) ([]service.AppointmentPetSelection, error) {
	petSelections := make([]service.AppointmentPetSelection, 0, len(req.Pets))
	if len(req.Pets) > 0 {
		for _, item := range req.Pets {
			petSelections = append(petSelections, service.AppointmentPetSelection{
				PetID:      item.PetID,
				ServiceIDs: item.ServiceIDs,
			})
		}
	} else if req.PetID > 0 && len(req.ServiceIDs) > 0 {
		petSelections = append(petSelections, service.AppointmentPetSelection{
			PetID:      req.PetID,
			ServiceIDs: req.ServiceIDs,
		})
	} else {
		return nil, errors.New("请至少选择一只宠物和服务")
	}
	return petSelections, nil
}

func (h *AppointmentHandler) Create(c *gin.Context) {
	var req createApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	petSelections, err := buildPetSelections(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	appt := &model.Appointment{
		ShopID:     c.GetUint("shop_id"),
		CustomerID: req.CustomerID,
		PetID:      petSelections[0].PetID,
		StaffID:    req.StaffID,
		Date:       req.Date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Status:     1,
		Source:     req.Source,
		Notes:      req.Notes,
	}
	if appt.Source == 0 {
		appt.Source = 2 // merchant created
	}

	if err := h.apptService.CreateMulti(appt, petSelections); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Reload with relations
	result, _ := h.apptService.GetByID(appt.ID)
	response.Success(c, result)
}

func (h *AppointmentHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var req createApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	petSelections, err := buildPetSelections(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	appt := &model.Appointment{
		ShopID:     c.GetUint("shop_id"),
		CustomerID: req.CustomerID,
		StaffID:    req.StaffID,
		Date:       req.Date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Source:     req.Source,
		Notes:      req.Notes,
	}
	if appt.Source == 0 {
		appt.Source = 2
	}

	if err := h.apptService.UpdateMulti(uint(id), c.GetUint("shop_id"), appt, petSelections); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, _ := h.apptService.GetByID(uint(id))
	response.Success(c, result)
}

func parseServiceIDs(raw string) []uint {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]uint, 0, len(parts))
	for _, part := range parts {
		id, err := strconv.ParseUint(strings.TrimSpace(part), 10, 64)
		if err == nil && id > 0 {
			result = append(result, uint(id))
		}
	}
	return result
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

func (h *AppointmentHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.apptService.Delete(uint(id), c.GetUint("shop_id")); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, gin.H{"id": id})
}

// GET /b/appointments?page=1&page_size=20&status=0
func (h *AppointmentHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	staffID, _ := strconv.ParseUint(c.Query("staff_id"), 10, 64)

	var status *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		status = &v
	}

	// Support customer_id filter
	customerID, _ := strconv.ParseUint(c.Query("customer_id"), 10, 64)
	if customerID > 0 {
		list, total, err := h.apptService.ListByCustomer(uint(customerID), page, pageSize)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "查询失败")
			return
		}
		response.Success(c, gin.H{"list": list, "total": total})
		return
	}

	list, total, err := h.apptService.ListPaged(shopID, status, dateFrom, dateTo, uint(staffID), page, pageSize)
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

func (h *AppointmentHandler) CalendarSummary(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	if startDate == "" || endDate == "" {
		response.Error(c, http.StatusBadRequest, "请提供start_date和end_date")
		return
	}

	summary, err := h.apptService.GetCalendarSummary(shopID, startDate, endDate)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, summary)
}

type setCalendarMarkReq struct {
	Marked bool `json:"marked"`
}

func (h *AppointmentHandler) SetCalendarMark(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	date := c.Param("date")
	if date == "" {
		response.Error(c, http.StatusBadRequest, "请提供日期")
		return
	}

	var req setCalendarMarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.apptService.SetCalendarMark(shopID, date, req.Marked); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"date": date, "marked": req.Marked})
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

type updateNotesReq struct {
	Notes string `json:"notes"`
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

func (h *AppointmentHandler) UpdateNotes(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req updateNotesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := h.apptService.UpdateNotes(uint(id), c.GetUint("shop_id"), req.Notes); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, _ := h.apptService.GetByID(uint(id))
	response.Success(c, result)
}
