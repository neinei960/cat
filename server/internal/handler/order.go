package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
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
type fromApptReq struct {
	AppointmentID uint `json:"appointment_id" binding:"required"`
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

func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	shopID := c.GetUint("shop_id")

	// Build order items
	var items []model.OrderItem
	var serviceAmount float64

	// If service_id + pet_id provided, do auto-pricing from fur_level
	if req.ServiceID > 0 && req.PetID > 0 {
		pet, err := h.petService.GetByID(req.PetID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "猫咪不存在")
			return
		}

		svc, err := h.serviceService.GetByID(req.ServiceID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "服务不存在")
			return
		}

		// Look up price by fur_level
		price := svc.BasePrice
		if pet.FurLevel != "" {
			rules, _ := h.serviceService.GetPriceRules(svc.ID)
			for _, r := range rules {
				if r.FurLevel == pet.FurLevel {
					price = r.Price
					break
				}
			}
		}

		serviceAmount = price
		items = append(items, model.OrderItem{
			ItemType:  1,
			ItemID:    svc.ID,
			Name:      svc.Name,
			Quantity:  1,
			UnitPrice: price,
			Amount:    price,
		})
	}

	// Legacy: support raw items array for backward compatibility
	for _, it := range req.Items {
		qty := it.Quantity
		if qty < 1 {
			qty = 1
		}
		amount := it.UnitPrice * float64(qty)
		if it.ItemType == 1 {
			serviceAmount += amount
		}
		items = append(items, model.OrderItem{
			ItemType:  it.ItemType,
			ItemID:    it.ItemID,
			Name:      it.Name,
			Quantity:  qty,
			UnitPrice: it.UnitPrice,
			Amount:    amount,
		})
	}

	// Add addon items (type=3)
	var addonTotal float64
	for _, addon := range req.Addons {
		if addon.Amount > 0 {
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
	}

	if len(items) == 0 {
		response.Error(c, http.StatusBadRequest, "请添加服务项目")
		return
	}

	// Calculate totals
	var totalAmount float64
	for _, it := range items {
		totalAmount += it.Amount
	}

	// Determine member discount
	discountRate := 1.0
	var customerID *uint
	if req.CustomerID != nil && *req.CustomerID > 0 {
		customerID = req.CustomerID
		cust, err := h.customerService.GetByID(*req.CustomerID)
		if err == nil && cust.DiscountRate > 0 && cust.DiscountRate < 1 {
			discountRate = cust.DiscountRate
		}
	} else if req.PetID > 0 {
		pet, _ := h.petService.GetByID(req.PetID)
		if pet != nil && pet.CustomerID != nil {
			customerID = pet.CustomerID
			cust, err := h.customerService.GetByID(*pet.CustomerID)
			if err == nil && cust.DiscountRate > 0 && cust.DiscountRate < 1 {
				discountRate = cust.DiscountRate
			}
		}
	}

	payAmount := math.Round(totalAmount*discountRate*100) / 100
	discountAmount := totalAmount - payAmount

	// Calculate staff commission
	var commission float64
	if req.StaffID != nil && *req.StaffID > 0 {
		var staff model.Staff
		if err := database.DB.First(&staff, *req.StaffID).Error; err == nil {
			// Commission based on service amount (not addons), rate is percentage like 20/30/35
			commission = math.Round(serviceAmount*staff.CommissionRate) / 100
		}
	}

	petID := &req.PetID
	if req.PetID == 0 {
		petID = nil
	}

	order := &model.Order{
		ShopID:         shopID,
		CustomerID:     customerID,
		PetID:          petID,
		StaffID:        req.StaffID,
		TotalAmount:    totalAmount,
		DiscountRate:   discountRate,
		DiscountAmount: discountAmount,
		PayAmount:      payAmount,
		Commission:     commission,
		Remark:         req.Remark,
	}

	if err := h.orderService.CreateDirect(order, items); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建订单失败")
		return
	}

	result, _ := h.orderService.GetByID(order.ID)
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
	var status *int
	if s := c.Query("status"); s != "" {
		v, _ := strconv.Atoi(s)
		status = &v
	}

	var list []model.Order
	var total int64
	var err error

	if keyword != "" {
		list, total, err = h.orderService.Search(shopID, keyword, status, page, pageSize)
	} else {
		list, total, err = h.orderService.ListPaged(shopID, status, page, pageSize)
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

	if err := h.orderService.MarkPaid(uint(id), req.PayMethod, req.TransactionID); err != nil {
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
