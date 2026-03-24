package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
)

type CustomerHandler struct {
	customerService *service.CustomerService
	petService      *service.PetService
}

func NewCustomerHandler(customerService *service.CustomerService, petService *service.PetService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService, petService: petService}
}

type createCustomerReq struct {
	Phone          string  `json:"phone"`
	Nickname       string  `json:"nickname" binding:"required"`
	Gender         int     `json:"gender"`
	Remark         string  `json:"remark"`
	Tags           string  `json:"tags"`
	CustomerTagIDs []uint  `json:"customer_tag_ids"`
	MemberBalance  float64 `json:"member_balance"`
	DiscountRate   float64 `json:"discount_rate"`
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req createCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	discountRate := req.DiscountRate
	if discountRate <= 0 {
		discountRate = 1
	}
	customer := &model.Customer{
		ShopID:        c.GetUint("shop_id"),
		Phone:         req.Phone,
		Nickname:      req.Nickname,
		Gender:        req.Gender,
		Remark:        req.Remark,
		Tags:          req.Tags,
		MemberBalance: req.MemberBalance,
		DiscountRate:  discountRate,
	}

	if err := h.customerService.CreateWithTags(customer, req.CustomerTagIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, customer)
}

func (h *CustomerHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	customer, err := h.customerService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "客户不存在")
		return
	}
	response.Success(c, customer)
}

func (h *CustomerHandler) List(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	memberCardTemplateID, _ := strconv.ParseUint(c.DefaultQuery("member_card_template_id", "0"), 10, 64)
	customerTagID, _ := strconv.ParseUint(c.DefaultQuery("customer_tag_id", "0"), 10, 64)

	var list []model.Customer
	var total int64
	var err error

	if keyword != "" {
		list, total, err = h.customerService.Search(shopID, keyword, page, pageSize, uint(memberCardTemplateID), uint(customerTagID))
	} else {
		list, total, err = h.customerService.List(shopID, page, pageSize, uint(memberCardTemplateID), uint(customerTagID))
	}

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	customer, err := h.customerService.GetByID(uint(id))
	if err != nil {
		response.Error(c, http.StatusNotFound, "客户不存在")
		return
	}

	var req createCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	customer.Nickname = req.Nickname
	customer.Phone = req.Phone
	customer.Gender = req.Gender
	customer.Remark = req.Remark
	customer.Tags = req.Tags
	customer.MemberBalance = req.MemberBalance
	if req.DiscountRate > 0 {
		customer.DiscountRate = req.DiscountRate
	}

	if err := h.customerService.UpdateWithTags(customer, req.CustomerTagIDs); err != nil {
		response.Error(c, http.StatusInternalServerError, "更新失败")
		return
	}
	response.Success(c, customer)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.customerService.Delete(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "删除失败")
		return
	}
	response.Success(c, nil)
}

func (h *CustomerHandler) ListDeleted(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	list, total, err := h.customerService.ListDeleted(shopID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

func (h *CustomerHandler) Restore(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.customerService.Restore(uint(id)); err != nil {
		response.Error(c, http.StatusInternalServerError, "恢复失败")
		return
	}
	response.Success(c, nil)
}

func (h *CustomerHandler) GetPets(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	pets, err := h.petService.FindByCustomer(uint(id))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	response.Success(c, pets)
}
