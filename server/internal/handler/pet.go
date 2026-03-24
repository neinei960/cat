package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type PetHandler struct {
	petService      *service.PetService
	customerService *service.CustomerService
}

func NewPetHandler(petService *service.PetService, customerService *service.CustomerService) *PetHandler {
	return &PetHandler{petService: petService, customerService: customerService}
}

type createPetReq struct {
	CustomerID     *uint   `json:"customer_id"`
	OwnerPhone     string  `json:"owner_phone"`
	Name           string  `json:"name" binding:"required"`
	Species        string  `json:"species"`
	Breed          string  `json:"breed"`
	Gender         int     `json:"gender"`
	BirthDate      string  `json:"birth_date"`
	Weight         float64 `json:"weight"`
	CoatType       string  `json:"coat_type"`
	CoatColor      string  `json:"coat_color"`
	FurLevel       string  `json:"fur_level"`
	Personality    string  `json:"personality"`
	Aggression     string  `json:"aggression"`
	ForbiddenZones string  `json:"forbidden_zones"`
	BathFrequency  string  `json:"bath_frequency"`
	Neutered       bool    `json:"neutered"`
	Avatar         string  `json:"avatar"`
	CareNotes      string  `json:"care_notes"`
	BehaviorNotes  string  `json:"behavior_notes"`
}

func (h *PetHandler) Create(c *gin.Context) {
	var req createPetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	species := req.Species
	if species == "" {
		species = "猫"
	}

	shopID := c.GetUint("shop_id")

	// 根据手机号查找客户
	customerID := req.CustomerID
	if req.OwnerPhone != "" && customerID == nil {
		cust, err := h.customerService.GetByPhone(req.OwnerPhone, shopID)
		if err != nil {
			// 自动创建客户
			cust = &model.Customer{
				ShopID: shopID,
				Phone:  req.OwnerPhone,
			}
			if err := h.customerService.Create(cust); err != nil {
				response.Error(c, http.StatusInternalServerError, "创建客户失败")
				return
			}
		}
		customerID = &cust.ID
	}

	pet := &model.Pet{
		ShopID:         shopID,
		CustomerID:     customerID,
		Name:           req.Name,
		Species:        species,
		Breed:          req.Breed,
		Gender:         req.Gender,
		BirthDate:      parseBirthDate(req.BirthDate),
		Weight:         req.Weight,
		CoatType:       req.CoatType,
		CoatColor:      req.CoatColor,
		FurLevel:       req.FurLevel,
		Personality:    req.Personality,
		Aggression:     req.Aggression,
		ForbiddenZones: req.ForbiddenZones,
		BathFrequency:  req.BathFrequency,
		Neutered:       req.Neutered,
		Avatar:         req.Avatar,
		CareNotes:      req.CareNotes,
		BehaviorNotes:  req.BehaviorNotes,
	}

	if err := h.petService.Create(pet); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, pet)
}

func (h *PetHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	pet, err := h.petService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "宠物不存在")
		return
	}
	response.Success(c, pet)
}

func (h *PetHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	var list []model.Pet
	var total int64
	var err error

	if keyword != "" {
		list, total, err = h.petService.Search(shopID, keyword, page, pageSize)
	} else {
		list, total, err = h.petService.List(shopID, page, pageSize)
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *PetHandler) Update(c *gin.Context) {
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

	// 根据手机号查找/创建客户
	if req.OwnerPhone != "" {
		shopID := c.GetUint("shop_id")
		cust, err := h.customerService.GetByPhone(req.OwnerPhone, shopID)
		if err != nil {
			cust = &model.Customer{
				ShopID: shopID,
				Phone:  req.OwnerPhone,
			}
			if err := h.customerService.Create(cust); err != nil {
				response.Error(c, http.StatusInternalServerError, "创建客户失败")
				return
			}
		}
		pet.CustomerID = &cust.ID
	} else if req.CustomerID != nil {
		pet.CustomerID = req.CustomerID
	} else {
		pet.CustomerID = nil
	}
	pet.Name = req.Name
	if req.Species != "" {
		pet.Species = req.Species
	}
	pet.Breed = req.Breed
	pet.Gender = req.Gender
	pet.BirthDate = parseBirthDate(req.BirthDate)
	pet.Weight = req.Weight
	pet.CoatType = req.CoatType
	pet.CoatColor = req.CoatColor
	pet.FurLevel = req.FurLevel
	pet.Personality = req.Personality
	pet.Aggression = req.Aggression
	pet.ForbiddenZones = req.ForbiddenZones
	pet.BathFrequency = req.BathFrequency
	pet.Neutered = req.Neutered
	pet.Avatar = req.Avatar
	pet.CareNotes = req.CareNotes
	pet.BehaviorNotes = req.BehaviorNotes

	if err := h.petService.Update(pet); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, pet)
}

func parseBirthDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "2006-01-02T15:04:05Z", "2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func (h *PetHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.petService.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}
