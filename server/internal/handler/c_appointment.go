package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type CAppointmentHandler struct {
	apptService    *service.AppointmentService
	serviceService *service.ServiceService
	staffService   *service.StaffService
	petService     *service.PetService
}

func NewCAppointmentHandler(
	apptService *service.AppointmentService,
	serviceService *service.ServiceService,
	staffService *service.StaffService,
	petService *service.PetService,
) *CAppointmentHandler {
	return &CAppointmentHandler{
		apptService:    apptService,
		serviceService: serviceService,
		staffService:   staffService,
		petService:     petService,
	}
}

// GET /c/services — list active services for the shop
func (h *CAppointmentHandler) ListServices(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	services, err := h.serviceService.ListActive(shopID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, services)
}

// GET /c/staffs?service_id=1 — list staff who can do a service
func (h *CAppointmentHandler) ListStaffs(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Query("service_id"), 10, 64)
	if serviceID == 0 {
		response.Error(c, http.StatusBadRequest, "请提供service_id")
		return
	}
	// Use service repo to find staff who can do this service
	shopID := c.GetUint("shop_id")
	staffs, _, err := h.staffService.List(shopID, 1, 100)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	// Filter active staff only
	var active []model.Staff
	for _, s := range staffs {
		if s.Status == 1 {
			active = append(active, s)
		}
	}
	response.Success(c, active)
}

// GET /c/slots?date=2026-03-20&service_id=1
func (h *CAppointmentHandler) GetSlots(c *gin.Context) {
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

// POST /c/appointments
type cCreateApptReq struct {
	PetID      uint   `json:"pet_id" binding:"required"`
	StaffID    *uint  `json:"staff_id"`
	Date       string `json:"date" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"`
	ServiceIDs []uint `json:"service_ids" binding:"required"`
	Notes      string `json:"notes"`
}

func (h *CAppointmentHandler) CreateAppointment(c *gin.Context) {
	var req cCreateApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	customerID, _ := c.Get("customer_id")
	shopID := c.GetUint("shop_id")

	appt := &model.Appointment{
		ShopID:     shopID,
		CustomerID: customerID.(uint),
		PetID:      req.PetID,
		StaffID:    req.StaffID,
		Date:       req.Date,
		StartTime:  req.StartTime,
		Source:     1, // mini-program
		Notes:      req.Notes,
	}

	if err := h.apptService.Create(appt, req.ServiceIDs); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, _ := h.apptService.GetByID(appt.ID)
	response.Success(c, result)
}

// GET /c/appointments
func (h *CAppointmentHandler) ListAppointments(c *gin.Context) {
	customerID, _ := c.Get("customer_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	repo := h.apptService // access via service
	_ = repo
	// Use the appointment repo through service's exposed method
	// For simplicity, query directly
	shopID := c.GetUint("shop_id")
	var status *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		status = &v
	}
	_ = shopID
	_ = status
	_ = page
	_ = pageSize

	// Customer's appointments via paged query
	list, total, err := h.apptService.ListByCustomer(customerID.(uint), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GET /c/appointments/:id
func (h *CAppointmentHandler) GetAppointment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	appt, err := h.apptService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "预约不存在")
		return
	}
	response.Success(c, appt)
}

// PUT /c/appointments/:id/cancel
func (h *CAppointmentHandler) CancelAppointment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.apptService.UpdateStatus(uint(id), 4, "", req.Reason, "customer"); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// C-end pet CRUD
func (h *CAppointmentHandler) ListPets(c *gin.Context) {
	customerID, _ := c.Get("customer_id")
	pets, err := h.petService.FindByCustomer(customerID.(uint))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, pets)
}

func (h *CAppointmentHandler) CreatePet(c *gin.Context) {
	customerID, _ := c.Get("customer_id")
	shopID := c.GetUint("shop_id")

	var req createPetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	custID := customerID.(uint)
	species := req.Species
	if species == "" {
		species = "猫"
	}
	pet := &model.Pet{
		ShopID:         shopID,
		CustomerID:     &custID,
		Name:           req.Name,
		Species:        species,
		Breed:          req.Breed,
		Gender:         req.Gender,
		Weight:         req.Weight,
		CoatType:       req.CoatType,
		CoatColor:      req.CoatColor,
		FurLevel:       req.FurLevel,
		Personality:    req.Personality,
		Aggression:     req.Aggression,
		ForbiddenZones: req.ForbiddenZones,
		BathFrequency:  req.BathFrequency,
		Neutered:       req.Neutered,
		CareNotes:      req.CareNotes,
		BehaviorNotes:  req.BehaviorNotes,
	}

	if err := h.petService.Create(pet); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, pet)
}

func (h *CAppointmentHandler) UpdatePet(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	pet, err := h.petService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "宠物不存在")
		return
	}

	var req createPetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	pet.Name = req.Name
	if req.Species != "" {
		pet.Species = req.Species
	}
	pet.Breed = req.Breed
	pet.Gender = req.Gender
	pet.Weight = req.Weight
	pet.CoatType = req.CoatType
	pet.CoatColor = req.CoatColor
	pet.FurLevel = req.FurLevel
	pet.Personality = req.Personality
	pet.Aggression = req.Aggression
	pet.ForbiddenZones = req.ForbiddenZones
	pet.BathFrequency = req.BathFrequency
	pet.Neutered = req.Neutered
	pet.CareNotes = req.CareNotes
	pet.BehaviorNotes = req.BehaviorNotes

	if err := h.petService.Update(pet); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, pet)
}
