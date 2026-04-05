package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type OrderHandler struct {
	orderService    *service.OrderService
	petService      *service.PetService
	customerService *service.CustomerService
	serviceService  *service.ServiceService
}

func NewOrderHandler(orderService *service.OrderService, petService *service.PetService, customerService *service.CustomerService, serviceService *service.ServiceService) *OrderHandler {
	return &OrderHandler{
		orderService:    orderService,
		petService:      petService,
		customerService: customerService,
		serviceService:  serviceService,
	}
}

// POST /b/orders/from-appointment
type serviceOverride struct {
	ServiceID   uint    `json:"service_id"`
	ServiceName string  `json:"service_name"`
	Price       float64 `json:"price"`
	Duration    int     `json:"duration"`
}

type petOverride struct {
	PetID    uint              `json:"pet_id"`
	Services []serviceOverride `json:"services"`
}

type fromApptReq struct {
	AppointmentID uint          `json:"appointment_id" binding:"required"`
	Overrides     []petOverride `json:"overrides"`
}

func (h *OrderHandler) CreateFromAppointment(c *gin.Context) {
	var req fromApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	order, err := h.orderService.CreateFromAppointment(req.AppointmentID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, _ := h.orderService.GetByID(order.ID)
	response.Success(c, result)
}

func (h *OrderHandler) CreateBatchFromAppointment(c *gin.Context) {
	var req fromApptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// 构建服务覆盖 map: petId -> []services
	overrideMap := make(map[uint][]service.ServiceOverride)
	for _, po := range req.Overrides {
		svcList := make([]service.ServiceOverride, 0, len(po.Services))
		for _, so := range po.Services {
			svcList = append(svcList, service.ServiceOverride{
				ServiceID:   so.ServiceID,
				ServiceName: so.ServiceName,
				Price:       so.Price,
				Duration:    so.Duration,
			})
		}
		overrideMap[po.PetID] = svcList
	}

	orders, err := h.orderService.CreateSplitFromAppointment(req.AppointmentID, overrideMap)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if len(orders) == 0 {
		response.Error(c, http.StatusInternalServerError, "开单失败")
		return
	}
	result, err := h.orderService.GetByID(orders[0].ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}
	response.Success(c, result)
}

// POST /b/orders — 猫咪洗护快速开单
type createOrderReq struct {
	PetID      uint             `json:"pet_id"`
	CustomerID *uint            `json:"customer_id"`
	StaffID    *uint            `json:"staff_id"`
	ServiceID  uint             `json:"service_id"`
	Remark     string           `json:"remark"`
	Addons     []addonInput     `json:"addons"`
	Items      []orderItemInput `json:"items"`
}

type addonInput struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type orderItemInput struct {
	ItemType  int     `json:"item_type"`
	ItemID    uint    `json:"item_id"`
	Name      string  `json:"name" binding:"required"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
}

var (
	errDraftPetNotFound    = errors.New("draft_pet_not_found")
	errDraftServiceMissing = errors.New("draft_service_missing")
	errDraftItemsMissing   = errors.New("draft_items_missing")
	errDraftPetRequired    = errors.New("draft_pet_required_for_service")
	errDraftStaffRequired  = errors.New("draft_staff_required_for_service")
)

func (h *OrderHandler) buildOrderDraft(shopID uint, req createOrderReq, existing *model.Order) (*model.Order, []model.OrderItem, error) {
	var items []model.OrderItem
	var selectedPet *model.Pet
	hasServiceItems := false
	hasProductItems := false

	if req.PetID > 0 {
		pet, err := h.petService.GetByID(req.PetID)
		if err != nil {
			return nil, nil, errDraftPetNotFound
		}
		selectedPet = pet
	}

	if req.ServiceID > 0 {
		if req.PetID == 0 {
			return nil, nil, errDraftPetRequired
		}

		svc, err := h.serviceService.GetByID(req.ServiceID)
		if err != nil {
			return nil, nil, errDraftServiceMissing
		}

		price := svc.BasePrice
		if selectedPet != nil && selectedPet.FurLevel != "" {
			rules, _ := h.serviceService.GetPriceRules(svc.ID)
			for _, r := range rules {
				if r.FurLevel == selectedPet.FurLevel {
					price = r.Price
					break
				}
			}
		}

		hasServiceItems = true
		items = append(items, model.OrderItem{
			ItemType:  1,
			ItemID:    svc.ID,
			Name:      svc.Name,
			Quantity:  1,
			UnitPrice: price,
			Amount:    price,
		})
	}

	for _, it := range req.Items {
		if it.ItemType == 1 {
			if req.PetID == 0 {
				return nil, nil, errDraftPetRequired
			}
			hasServiceItems = true
		}
		if it.ItemType == 2 {
			hasProductItems = true
		}
		qty := it.Quantity
		if qty < 1 {
			qty = 1
		}
		amount := it.UnitPrice * float64(qty)
		items = append(items, model.OrderItem{
			ItemType:  it.ItemType,
			ItemID:    it.ItemID,
			Name:      it.Name,
			Quantity:  qty,
			UnitPrice: it.UnitPrice,
			Amount:    amount,
		})
	}

	var addonTotal float64
	for _, addon := range req.Addons {
		if addon.Amount <= 0 {
			continue
		}
		addonTotal += addon.Amount
		items = append(items, model.OrderItem{
			ItemType:  3,
			ItemID:    0,
			Name:      addon.Name,
			Quantity:  1,
			UnitPrice: addon.Amount,
			Amount:    addon.Amount,
		})
	}

	if !hasServiceItems && !hasProductItems {
		return nil, nil, errDraftItemsMissing
	}

	var customerID *uint
	if existing != nil {
		customerID = existing.CustomerID
	}
	if req.CustomerID != nil && *req.CustomerID > 0 {
		customerID = req.CustomerID
	}
	if selectedPet != nil && selectedPet.CustomerID != nil {
		customerID = selectedPet.CustomerID
	}

	serviceDiscountRate := 1.0
	productDiscountRate := 1.0
	if customerID != nil && *customerID > 0 {
		cust, err := h.customerService.GetByID(*customerID)
		if err == nil && cust.DiscountRate > 0 && cust.DiscountRate < 1 {
			serviceDiscountRate = cust.DiscountRate
		}

		var card model.MemberCard
		if err := database.DB.Where("customer_id = ? AND status = 1", *customerID).First(&card).Error; err == nil {
			if card.ProductDiscountRate > 0 && card.ProductDiscountRate < 1 {
				productDiscountRate = card.ProductDiscountRate
			}
		}
	}

	var serviceTotal, productTotal float64
	for _, it := range items {
		switch it.ItemType {
		case 1:
			serviceTotal += it.Amount
		case 2:
			productTotal += it.Amount
		}
	}
	totalAmount := serviceTotal + productTotal + addonTotal
	servicePayAmount := math.Round(serviceTotal*serviceDiscountRate*100) / 100
	productPayAmount := math.Round(productTotal*productDiscountRate*100) / 100
	serviceDiscountAmount := serviceTotal - servicePayAmount
	productDiscountAmount := productTotal - productPayAmount
	payAmount := math.Round((servicePayAmount+productPayAmount+addonTotal)*100) / 100
	discountAmount := math.Round((serviceDiscountAmount+productDiscountAmount)*100) / 100

	var staffID *uint
	if existing != nil {
		staffID = existing.StaffID
	}
	if req.StaffID != nil && *req.StaffID > 0 {
		staffID = req.StaffID
	}
	if hasServiceItems && (staffID == nil || *staffID == 0) {
		return nil, nil, errDraftStaffRequired
	}

	var commission float64
	if staffID != nil && *staffID > 0 {
		var staff model.Staff
		if err := database.DB.First(&staff, *staffID).Error; err == nil {
			commission += math.Round(serviceTotal*staff.CommissionRate) / 100
			commission += math.Round(productTotal*staff.ProductCommissionRate) / 100
		}
	}

	var petID *uint
	if existing != nil {
		petID = existing.PetID
	}
	if req.PetID > 0 {
		petID = &req.PetID
	} else if existing == nil {
		petID = nil
	}

	order := &model.Order{
		ShopID:                shopID,
		CustomerID:            customerID,
		PetID:                 petID,
		StaffID:               staffID,
		TotalAmount:           totalAmount,
		ServiceTotal:          serviceTotal,
		ProductTotal:          productTotal,
		AddonTotal:            addonTotal,
		DiscountRate:          calculateEffectiveDiscountRate(totalAmount, payAmount),
		DiscountAmount:        discountAmount,
		ServiceDiscountAmount: serviceDiscountAmount,
		ProductDiscountAmount: productDiscountAmount,
		PayAmount:             payAmount,
		Commission:            commission,
		Remark:                req.Remark,
	}
	if existing != nil {
		order.AppointmentID = existing.AppointmentID
	}
	return order, items, nil
}

func calculateEffectiveDiscountRate(totalAmount, payAmount float64) float64 {
	if totalAmount <= 0 {
		return 1
	}
	return math.Round((payAmount/totalAmount)*100) / 100
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	shopID := c.GetUint("shop_id")
	order, items, err := h.buildOrderDraft(shopID, req, nil)
	if err != nil {
		switch err {
		case errDraftPetNotFound:
			response.Error(c, http.StatusBadRequest, "猫咪不存在")
			return
		case errDraftServiceMissing:
			response.Error(c, http.StatusBadRequest, "服务不存在")
			return
		case errDraftItemsMissing:
			response.Error(c, http.StatusBadRequest, "请添加商品或服务")
			return
		case errDraftPetRequired:
			response.Error(c, http.StatusBadRequest, "请选择猫咪后再添加服务")
			return
		case errDraftStaffRequired:
			response.Error(c, http.StatusBadRequest, "请选择洗护师")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.orderService.CreateDirect(order, items); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	result, _ := h.orderService.GetByID(order.ID)
	response.Success(c, result)
}

func (h *OrderHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	existing, err := h.orderService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "订单不存在")
		return
	}

	order, items, err := h.buildOrderDraft(c.GetUint("shop_id"), req, existing)
	if err != nil {
		switch err {
		case errDraftPetNotFound:
			response.Error(c, http.StatusBadRequest, "猫咪不存在")
			return
		case errDraftServiceMissing:
			response.Error(c, http.StatusBadRequest, "服务不存在")
			return
		case errDraftItemsMissing:
			response.Error(c, http.StatusBadRequest, "请添加商品或服务")
			return
		case errDraftPetRequired:
			response.Error(c, http.StatusBadRequest, "请选择猫咪后再添加服务")
			return
		case errDraftStaffRequired:
			response.Error(c, http.StatusBadRequest, "请选择洗护师")
			return
		}
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.orderService.UpdateDraft(c.GetUint("shop_id"), uint(id), order, items); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.orderService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询订单失败")
		return
	}
	response.Success(c, result)
}

// GET /b/orders/price-lookup?service_id=X&fur_level=Y — 定价查询
func (h *OrderHandler) PriceLookup(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Query("service_id"), 10, 64)
	furLevel := c.Query("fur_level")

	if serviceID == 0 {
		response.Error(c, http.StatusBadRequest, "请提供service_id")
		return
	}

	svc, err := h.serviceService.GetByID(uint(serviceID))
	if err != nil {
		response.Error(c, http.StatusNotFound, "服务不存在")
		return
	}

	price := svc.BasePrice
	if furLevel != "" {
		rules, _ := h.serviceService.GetPriceRules(svc.ID)
		for _, r := range rules {
			if r.FurLevel == furLevel {
				price = r.Price
				break
			}
		}
	}

	response.Success(c, gin.H{"price": price, "service_name": svc.Name, "fur_level": furLevel})
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	order, err := h.orderService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "订单不存在")
		return
	}
	response.Success(c, order)
}

func (h *OrderHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	staffID, _ := strconv.ParseUint(c.Query("staff_id"), 10, 64)

	f := repository.OrderFilter{
		DateFrom:       c.Query("date_from"),
		DateTo:         c.Query("date_to"),
		StaffID:        uint(staffID),
		PayMethod:      c.Query("pay_method"),
		ProductKeyword: c.Query("product_keyword"),
	}
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		f.Status = &v
	}

	var list []model.Order
	var total int64
	var err error

	if keyword != "" {
		list, total, err = h.orderService.Search(shopID, keyword, f, page, pageSize)
	} else {
		list, total, err = h.orderService.ListPaged(shopID, f, page, pageSize)
	}
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// PUT /b/orders/:id/pay
type payReq struct {
	PayMethod     string `json:"pay_method" binding:"required"`
	TransactionID string `json:"transaction_id"`
	Remark        string `json:"remark"`
}

func (h *OrderHandler) Pay(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req payReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// If paying with balance, deduct from member card first
	if req.PayMethod == "balance" {
		order, err := h.orderService.GetByID(uint(id))
		if err != nil {
			response.Error(c, http.StatusNotFound, "订单不存在")
			return
		}
		if order.CustomerID == nil {
			response.Error(c, http.StatusBadRequest, "该订单无关联客户")
			return
		}
		staffID := c.GetUint("staff_id")
		if err := BalancePayment(order.ShopID, *order.CustomerID, order.ID, order.PayAmount, staffID); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	if req.Remark != "" {
		if err := h.orderService.UpdateRemark(uint(id), req.Remark); err != nil {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := h.orderService.MarkPaid(uint(id), req.PayMethod, req.TransactionID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

type updateOrderRemarkReq struct {
	Remark string `json:"remark"`
}

func (h *OrderHandler) UpdateRemark(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req updateOrderRemarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	if err := h.orderService.UpdateRemark(uint(id), req.Remark); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// PUT /b/orders/:id/refund
type refundReq struct {
	Remark string `json:"remark"`
}

func (h *OrderHandler) Refund(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req refundReq
	c.ShouldBindJSON(&req)

	if err := h.orderService.Refund(uint(id), req.Remark); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

// PUT /b/orders/:id/cancel
func (h *OrderHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.orderService.Cancel(uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *OrderHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.orderService.Delete(c.GetUint("shop_id"), uint(id)); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}
