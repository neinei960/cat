package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type BoardingHandler struct {
	service *service.BoardingService
}

func NewBoardingHandler(service *service.BoardingService) *BoardingHandler {
	return &BoardingHandler{service: service}
}

type saveCabinetReq struct {
	Code          string  `json:"code"`
	CabinetType   string  `json:"cabinet_type" binding:"required"`
	RoomCount     int     `json:"room_count"`
	Capacity      int     `json:"capacity"`
	BasePrice     float64 `json:"base_price"`
	ExtraPetPrice float64 `json:"extra_pet_price"`
	Status        string  `json:"status"`
	Remark        string  `json:"remark"`
}

func (h *BoardingHandler) ListCabinets(c *gin.Context) {
	list, err := h.service.ListCabinets(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, list)
}

func (h *BoardingHandler) CreateCabinet(c *gin.Context) {
	var req saveCabinetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	cabinet := &model.BoardingCabinet{
		ShopID:        c.GetUint("shop_id"),
		Code:          req.Code,
		CabinetType:   req.CabinetType,
		RoomCount:     req.RoomCount,
		Capacity:      req.Capacity,
		BasePrice:     req.BasePrice,
		ExtraPetPrice: req.ExtraPetPrice,
		Status:        req.Status,
		Remark:        req.Remark,
	}
	if err := h.service.CreateCabinet(cabinet); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, cabinet)
}

func (h *BoardingHandler) UpdateCabinet(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req saveCabinetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	cabinet := &model.BoardingCabinet{
		Model:         model.BoardingCabinet{}.Model,
		Code:          req.Code,
		CabinetType:   req.CabinetType,
		RoomCount:     req.RoomCount,
		Capacity:      req.Capacity,
		BasePrice:     req.BasePrice,
		ExtraPetPrice: req.ExtraPetPrice,
		Status:        req.Status,
		Remark:        req.Remark,
	}
	cabinet.ID = uint(id)
	if err := h.service.UpdateCabinet(c.GetUint("shop_id"), cabinet); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, cabinet)
}

type saveHolidayReq struct {
	HolidayDate string `json:"holiday_date" binding:"required"`
	Name        string `json:"name"`
}

func (h *BoardingHandler) ListHolidays(c *gin.Context) {
	list, err := h.service.ListHolidays(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, list)
}

func (h *BoardingHandler) CreateHoliday(c *gin.Context) {
	var req saveHolidayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	holiday := &model.BoardingHoliday{
		ShopID:      c.GetUint("shop_id"),
		HolidayDate: req.HolidayDate,
		Name:        req.Name,
	}
	if err := h.service.CreateHoliday(holiday); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, holiday)
}

func (h *BoardingHandler) DeleteHoliday(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.DeleteHoliday(c.GetUint("shop_id"), uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

type savePolicyReq struct {
	Name       string `json:"name" binding:"required"`
	PolicyType string `json:"policy_type" binding:"required"`
	Rule       any    `json:"rule"`
	ValidFrom  string `json:"valid_from"`
	ValidTo    string `json:"valid_to"`
	Priority   int    `json:"priority"`
	Stackable  bool   `json:"stackable"`
	Status     int    `json:"status"`
	Remark     string `json:"remark"`
}

func (h *BoardingHandler) ListPolicies(c *gin.Context) {
	list, err := h.service.ListPolicies(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, list)
}

func buildPolicyModel(shopID uint, req savePolicyReq) (*model.BoardingDiscountPolicy, error) {
	ruleJSON, err := json.Marshal(req.Rule)
	if err != nil {
		return nil, err
	}
	return &model.BoardingDiscountPolicy{
		ShopID:     shopID,
		Name:       req.Name,
		PolicyType: req.PolicyType,
		RuleJSON:   string(ruleJSON),
		ValidFrom:  req.ValidFrom,
		ValidTo:    req.ValidTo,
		Priority:   req.Priority,
		Stackable:  req.Stackable,
		Status:     req.Status,
		Remark:     req.Remark,
	}, nil
}

func (h *BoardingHandler) CreatePolicy(c *gin.Context) {
	var req savePolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	policy, err := buildPolicyModel(c.GetUint("shop_id"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "规则格式错误")
		return
	}
	if err := h.service.CreatePolicy(policy); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, policy)
}

func (h *BoardingHandler) UpdatePolicy(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req savePolicyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	policy, err := buildPolicyModel(c.GetUint("shop_id"), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "规则格式错误")
		return
	}
	policy.ID = uint(id)
	if err := h.service.UpdatePolicy(c.GetUint("shop_id"), policy); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, policy)
}

func (h *BoardingHandler) GetAvailableCabinets(c *gin.Context) {
	petCount, _ := strconv.Atoi(c.DefaultQuery("pet_count", "1"))
	excludeOrderID, _ := strconv.ParseUint(c.DefaultQuery("exclude_order_id", "0"), 10, 64)
	excludeRoomID, _ := strconv.ParseUint(c.DefaultQuery("exclude_room_id", "0"), 10, 64)
	list, err := h.service.GetAvailableCabinets(
		c.GetUint("shop_id"),
		c.Query("check_in_at"),
		c.Query("check_out_at"),
		petCount,
		uint(excludeOrderID),
		uint(excludeRoomID),
	)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, list)
}

type boardingRoomGroupReq struct {
	PetIDs     []uint `json:"pet_ids"`
	PetCount   int    `json:"pet_count"`
	CabinetID  uint   `json:"cabinet_id" binding:"required"`
	CheckInAt  string `json:"check_in_at" binding:"required"`
	CheckOutAt string `json:"check_out_at" binding:"required"`
}

type previewReq struct {
	CustomerID uint                   `json:"customer_id"`
	PetIDs     []uint                 `json:"pet_ids"`
	PetCount   int                    `json:"pet_count"`
	CabinetID  uint                   `json:"cabinet_id"`
	CheckInAt  string                 `json:"check_in_at"`
	CheckOutAt string                 `json:"check_out_at"`
	PolicyIDs  []uint                 `json:"policy_ids"`
	RoomGroups []boardingRoomGroupReq `json:"room_groups"`
}

func (h *BoardingHandler) PricePreview(c *gin.Context) {
	var req previewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	roomGroups := make([]service.BoardingRoomGroupInput, 0, len(req.RoomGroups))
	for _, group := range req.RoomGroups {
		roomGroups = append(roomGroups, service.BoardingRoomGroupInput{
			PetIDs:     group.PetIDs,
			PetCount:   group.PetCount,
			CabinetID:  group.CabinetID,
			CheckInAt:  group.CheckInAt,
			CheckOutAt: group.CheckOutAt,
		})
	}
	preview, err := h.service.PreviewOrder(c.GetUint("shop_id"), service.BoardingPreviewInput{
		CustomerID: req.CustomerID,
		PetIDs:     req.PetIDs,
		PetCount:   req.PetCount,
		CabinetID:  req.CabinetID,
		CheckInAt:  req.CheckInAt,
		CheckOutAt: req.CheckOutAt,
		PolicyIDs:  req.PolicyIDs,
		RoomGroups: roomGroups,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, preview)
}

type createBoardingOrderReq struct {
	CustomerID   uint                   `json:"customer_id" binding:"required"`
	PetIDs       []uint                 `json:"pet_ids"`
	CabinetID    uint                   `json:"cabinet_id"`
	CheckInAt    string                 `json:"check_in_at"`
	CheckOutAt   string                 `json:"check_out_at"`
	PolicyIDs    []uint                 `json:"policy_ids"`
	RoomGroups   []boardingRoomGroupReq `json:"room_groups"`
	HasDeworming *bool                  `json:"has_deworming"`
	Remark       string                 `json:"remark"`
}

func (h *BoardingHandler) CreateOrder(c *gin.Context) {
	var req createBoardingOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	roomGroups := make([]service.BoardingRoomGroupInput, 0, len(req.RoomGroups))
	for _, group := range req.RoomGroups {
		roomGroups = append(roomGroups, service.BoardingRoomGroupInput{
			PetIDs:     group.PetIDs,
			PetCount:   group.PetCount,
			CabinetID:  group.CabinetID,
			CheckInAt:  group.CheckInAt,
			CheckOutAt: group.CheckOutAt,
		})
	}
	order, err := h.service.CreateOrder(c.GetUint("shop_id"), service.BoardingCreateInput{
		CustomerID:   req.CustomerID,
		PetIDs:       req.PetIDs,
		CabinetID:    req.CabinetID,
		CheckInAt:    req.CheckInAt,
		CheckOutAt:   req.CheckOutAt,
		PolicyIDs:    req.PolicyIDs,
		RoomGroups:   roomGroups,
		HasDeworming: req.HasDeworming,
		Remark:       req.Remark,
		OperatorID:   c.GetUint("staff_id"),
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.service.ListOrders(c.GetUint("shop_id"), c.Query("status"), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *BoardingHandler) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.service.GetOrder(c.GetUint("shop_id"), uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "寄养订单不存在")
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) Dashboard(c *gin.Context) {
	data, err := h.service.Dashboard(c.GetUint("shop_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, data)
}

func (h *BoardingHandler) CheckIn(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		DiscountAmount float64 `json:"discount_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.CheckIn(c.GetUint("shop_id"), uint(id), c.GetUint("staff_id"), service.BoardingCheckInInput{
		DiscountAmount: req.DiscountAmount,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) CheckInRoom(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roomID, _ := strconv.ParseUint(c.Param("room_id"), 10, 64)
	var req struct {
		DiscountAmount float64 `json:"discount_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.CheckInRoom(c.GetUint("shop_id"), uint(id), uint(roomID), c.GetUint("staff_id"), service.BoardingCheckInInput{
		DiscountAmount: req.DiscountAmount,
	})
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) CheckOut(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		ActualCheckOutAt string `json:"actual_check_out_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.CheckOut(c.GetUint("shop_id"), uint(id), c.GetUint("staff_id"), req.ActualCheckOutAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) CheckOutRoom(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roomID, _ := strconv.ParseUint(c.Param("room_id"), 10, 64)
	var req struct {
		ActualCheckOutAt string `json:"actual_check_out_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.CheckOutRoom(c.GetUint("shop_id"), uint(id), uint(roomID), c.GetUint("staff_id"), req.ActualCheckOutAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) Extend(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		CheckOutAt string `json:"check_out_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.Extend(c.GetUint("shop_id"), uint(id), c.GetUint("staff_id"), req.CheckOutAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) ExtendRoom(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roomID, _ := strconv.ParseUint(c.Param("room_id"), 10, 64)
	var req struct {
		CheckOutAt string `json:"check_out_at" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.ExtendRoom(c.GetUint("shop_id"), uint(id), uint(roomID), c.GetUint("staff_id"), req.CheckOutAt)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) ChangeCabinet(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		CabinetID uint `json:"cabinet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.ChangeCabinet(c.GetUint("shop_id"), uint(id), c.GetUint("staff_id"), req.CabinetID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) ChangeRoomCabinet(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roomID, _ := strconv.ParseUint(c.Param("room_id"), 10, 64)
	var req struct {
		CabinetID uint `json:"cabinet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.service.ChangeRoomCabinet(c.GetUint("shop_id"), uint(id), uint(roomID), c.GetUint("staff_id"), req.CabinetID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.service.Cancel(c.GetUint("shop_id"), uint(id), c.GetUint("staff_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}

func (h *BoardingHandler) CancelRoom(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	roomID, _ := strconv.ParseUint(c.Param("room_id"), 10, 64)
	order, err := h.service.CancelRoom(c.GetUint("shop_id"), uint(id), uint(roomID), c.GetUint("staff_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, order)
}
