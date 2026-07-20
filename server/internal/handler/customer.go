package handler

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/service"
	"github.com/neinei960/cat/server/pkg/response"
	"gorm.io/gorm"
)

type CustomerHandler struct {
	customerService *service.CustomerService
	petService      *service.PetService
}

var errDuplicateCustomerPhone = errors.New("该手机号客户已存在")

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
	Address        string  `json:"address"`
	AddressDetail  string  `json:"address_detail"`
	DoorCode       string  `json:"door_code"`
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req createCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	shopID := c.GetUint("shop_id")
	req.Phone = ensureCustomerPhone(h.customerService, shopID, req.Phone)
	if err := h.ensurePhoneUnique(shopID, req.Phone, 0); err != nil {
		if errors.Is(err, errDuplicateCustomerPhone) {
			response.Error(c, http.StatusConflict, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, "校验手机号失败")
		}
		return
	}

	discountRate := req.DiscountRate
	if discountRate <= 0 {
		discountRate = 1
	}
	customer := &model.Customer{
		ShopID:        shopID,
		Phone:         req.Phone,
		Nickname:      req.Nickname,
		Gender:        req.Gender,
		Remark:        req.Remark,
		Tags:          req.Tags,
		MemberBalance: req.MemberBalance,
		DiscountRate:  discountRate,
		Address:       req.Address,
		AddressDetail: req.AddressDetail,
		DoorCode:      req.DoorCode,
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

	req.Phone = ensureCustomerPhone(h.customerService, customer.ShopID, req.Phone)
	if err := h.ensurePhoneUnique(customer.ShopID, req.Phone, customer.ID); err != nil {
		if errors.Is(err, errDuplicateCustomerPhone) {
			response.Error(c, http.StatusConflict, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, "校验手机号失败")
		}
		return
	}

	customer.Nickname = req.Nickname
	customer.Phone = req.Phone
	customer.Gender = req.Gender
	customer.Remark = req.Remark
	customer.Tags = req.Tags
	customer.MemberBalance = req.MemberBalance
	customer.Address = req.Address
	customer.AddressDetail = req.AddressDetail
	customer.DoorCode = req.DoorCode
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

func ensureCustomerPhone(customerService *service.CustomerService, shopID uint, phone string) string {
	if phone != "" {
		return phone
	}
	for i := 0; i < 10; i++ {
		candidate := generatedCustomerPhone(i)
		if _, err := customerService.GetByPhone(candidate, shopID); errors.Is(err, gorm.ErrRecordNotFound) {
			return candidate
		}
	}
	return generatedCustomerPhone(99)
}

func generatedCustomerPhone(offset int) string {
	value := randomTenDigitValue()
	if value < 0 {
		value = (time.Now().UnixNano() + int64(offset)) % 10000000000
	}
	if value < 0 {
		value = -value
	}
	return "9" + strconv.FormatInt(value+10000000000, 10)[1:]
}

func randomTenDigitValue() int64 {
	max := big.NewInt(10000000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return -1
	}
	return n.Int64()
}

func (h *CustomerHandler) ensurePhoneUnique(shopID uint, phone string, currentCustomerID uint) error {
	if phone == "" {
		return nil
	}

	customer, err := h.customerService.GetByPhone(phone, shopID)
	if err == nil && customer.ID != currentCustomerID {
		return errDuplicateCustomerPhone
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
