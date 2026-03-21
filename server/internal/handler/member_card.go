package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/response"
)

type MemberCardHandler struct{}

func NewMemberCardHandler() *MemberCardHandler {
	return &MemberCardHandler{}
}

// ========== Template CRUD ==========

type templateReq struct {
	Name         string  `json:"name" binding:"required"`
	MinRecharge  float64 `json:"min_recharge" binding:"required"`
	DiscountRate float64 `json:"discount_rate" binding:"required"`
	ValidDays    int     `json:"valid_days"`
	SortOrder    int     `json:"sort_order"`
	Status       int     `json:"status"`
	Color        string  `json:"color"`
}

func (h *MemberCardHandler) CreateTemplate(c *gin.Context) {
	var req templateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	color := req.Color
	if color == "" {
		color = "linear-gradient(135deg, #4F46E5, #7C3AED)"
	}
	tpl := &model.MemberCardTemplate{
		ShopID:       c.GetUint("shop_id"),
		Name:         req.Name,
		MinRecharge:  req.MinRecharge,
		DiscountRate: req.DiscountRate,
		ValidDays:    req.ValidDays,
		SortOrder:    req.SortOrder,
		Color:        color,
		Status:       1,
	}
	if err := database.DB.Create(tpl).Error; err != nil {
		response.Error(c, http.StatusInternalServerError, "创建失败")
		return
	}
	response.Success(c, tpl)
}

func (h *MemberCardHandler) ListTemplates(c *gin.Context) {
	shopID := c.GetUint("shop_id")
	var list []model.MemberCardTemplate
	database.DB.Preload("Discounts").Where("shop_id = ?", shopID).Order("sort_order ASC, id ASC").Find(&list)
	response.Success(c, list)
}

// ========== Template Discounts (per category) ==========

type setDiscountsReq struct {
	Discounts []struct {
		CategoryID   uint    `json:"category_id"`
		CategoryName string  `json:"category_name"`
		DiscountRate float64 `json:"discount_rate"`
	} `json:"discounts"`
}

func (h *MemberCardHandler) SetDiscounts(c *gin.Context) {
	templateID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req setDiscountsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	tx := database.DB.Begin()
	// Delete old discounts
	tx.Where("template_id = ?", templateID).Delete(&model.MemberCardDiscount{})
	// Create new
	for _, d := range req.Discounts {
		if d.DiscountRate > 0 && d.DiscountRate < 1 {
			tx.Create(&model.MemberCardDiscount{
				TemplateID:   uint(templateID),
				CategoryID:   d.CategoryID,
				CategoryName: d.CategoryName,
				DiscountRate: d.DiscountRate,
			})
		}
	}
	tx.Commit()

	// Return updated template with discounts
	var tpl model.MemberCardTemplate
	database.DB.Preload("Discounts").First(&tpl, templateID)
	response.Success(c, tpl)
}

func (h *MemberCardHandler) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var tpl model.MemberCardTemplate
	if err := database.DB.First(&tpl, id).Error; err != nil {
		response.Error(c, http.StatusNotFound, "模板不存在")
		return
	}
	var req templateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	tpl.Name = req.Name
	tpl.MinRecharge = req.MinRecharge
	tpl.DiscountRate = req.DiscountRate
	tpl.ValidDays = req.ValidDays
	tpl.SortOrder = req.SortOrder
	if req.Color != "" {
		tpl.Color = req.Color
	}
	if req.Status > 0 {
		tpl.Status = req.Status
	}
	database.DB.Save(&tpl)
	response.Success(c, tpl)
}

func (h *MemberCardHandler) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	// Check if any active cards use this template
	var count int64
	database.DB.Model(&model.MemberCard{}).Where("template_id = ? AND status = 1", id).Count(&count)
	if count > 0 {
		response.Error(c, http.StatusBadRequest, "有会员正在使用此卡，无法删除")
		return
	}
	database.DB.Delete(&model.MemberCardTemplate{}, id)
	response.Success(c, nil)
}

// ========== Open Card ==========

type openCardReq struct {
	TemplateID     uint    `json:"template_id" binding:"required"`
	RechargeAmount float64 `json:"recharge_amount" binding:"required"`
	Remark         string  `json:"remark"`
}

func (h *MemberCardHandler) OpenCard(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	shopID := c.GetUint("shop_id")
	staffID := c.GetUint("staff_id")

	var req openCardReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	// Check customer exists
	var customer model.Customer
	if err := database.DB.First(&customer, customerID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "客户不存在")
		return
	}

	// Check no existing card
	var existingCount int64
	database.DB.Model(&model.MemberCard{}).Where("customer_id = ? AND status = 1", customerID).Count(&existingCount)
	if existingCount > 0 {
		response.Error(c, http.StatusBadRequest, "该客户已有会员卡")
		return
	}

	// Get template
	var tpl model.MemberCardTemplate
	if err := database.DB.First(&tpl, req.TemplateID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "会员卡模板不存在")
		return
	}

	if req.RechargeAmount < tpl.MinRecharge {
		response.Error(c, http.StatusBadRequest, "充值金额不能低于门槛"+strconv.FormatFloat(tpl.MinRecharge, 'f', 0, 64)+"元")
		return
	}

	// Calculate expiry
	var expireAt *time.Time
	if tpl.ValidDays > 0 {
		t := time.Now().AddDate(0, 0, tpl.ValidDays)
		expireAt = &t
	}

	// Transaction
	tx := database.DB.Begin()

	card := &model.MemberCard{
		ShopID:        shopID,
		CustomerID:    uint(customerID),
		TemplateID:    req.TemplateID,
		CardName:      tpl.Name,
		Balance:       req.RechargeAmount,
		TotalRecharge: req.RechargeAmount,
		DiscountRate:        tpl.DiscountRate,
		ProductDiscountRate: tpl.ProductDiscountRate,
		ExpireAt:      expireAt,
		Status:        1,
	}
	if err := tx.Create(card).Error; err != nil {
		tx.Rollback()
		response.Error(c, http.StatusInternalServerError, "开卡失败")
		return
	}

	record := &model.RechargeRecord{
		ShopID:       shopID,
		CustomerID:   uint(customerID),
		CardID:       card.ID,
		Type:         1, // 充值
		Amount:       req.RechargeAmount,
		BalanceAfter: card.Balance,
		Remark:       "开卡充值",
		OperatorID:   &staffID,
	}
	tx.Create(record)

	// Update customer
	tx.Model(&customer).Updates(map[string]interface{}{
		"member_card_id":  card.ID,
		"member_balance":  card.Balance,
		"discount_rate":   tpl.DiscountRate,
	})

	tx.Commit()
	response.Success(c, card)
}

// ========== Recharge ==========

type rechargeReq struct {
	Amount float64 `json:"amount" binding:"required"`
	Remark string  `json:"remark"`
}

func (h *MemberCardHandler) Recharge(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	shopID := c.GetUint("shop_id")
	staffID := c.GetUint("staff_id")

	var req rechargeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	var card model.MemberCard
	if err := database.DB.Where("customer_id = ? AND status = 1", customerID).First(&card).Error; err != nil {
		response.Error(c, http.StatusNotFound, "该客户没有会员卡")
		return
	}

	tx := database.DB.Begin()

	card.Balance += req.Amount
	card.TotalRecharge += req.Amount
	tx.Save(&card)

	record := &model.RechargeRecord{
		ShopID:       shopID,
		CustomerID:   uint(customerID),
		CardID:       card.ID,
		Type:         1,
		Amount:       req.Amount,
		BalanceAfter: card.Balance,
		Remark:       req.Remark,
		OperatorID:   &staffID,
	}
	tx.Create(record)

	tx.Model(&model.Customer{}).Where("id = ?", customerID).Update("member_balance", card.Balance)

	tx.Commit()
	response.Success(c, card)
}

// ========== Get Card ==========

func (h *MemberCardHandler) GetCard(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var card model.MemberCard
	if err := database.DB.Preload("Template").Where("customer_id = ? AND status = 1", customerID).First(&card).Error; err != nil {
		response.Success(c, nil) // No card, not an error
		return
	}
	response.Success(c, card)
}

// ========== Recharge Records ==========

func (h *MemberCardHandler) GetRecords(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var records []model.RechargeRecord
	database.DB.Where("customer_id = ?", customerID).Order("created_at DESC").Limit(50).Find(&records)
	response.Success(c, records)
}

// ========== Adjust Balance (admin/manager only) ==========

type adjustBalanceReq struct {
	Amount float64 `json:"amount" binding:"required"` // 正数=增加 负数=减少
	Remark string  `json:"remark" binding:"required"`
}

func (h *MemberCardHandler) AdjustBalance(c *gin.Context) {
	customerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	shopID := c.GetUint("shop_id")
	staffID := c.GetUint("staff_id")

	var req adjustBalanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请填写金额和原因")
		return
	}

	var card model.MemberCard
	if err := database.DB.Where("customer_id = ? AND status = 1", customerID).First(&card).Error; err != nil {
		response.Error(c, http.StatusNotFound, "该客户没有会员卡")
		return
	}

	newBalance := card.Balance + req.Amount
	if newBalance < 0 {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("余额不足，当前余额:%.2f", card.Balance))
		return
	}

	tx := database.DB.Begin()

	card.Balance = newBalance
	if req.Amount > 0 {
		card.TotalRecharge += req.Amount
	}
	tx.Save(&card)

	// 记录类型: 4=调整
	record := &model.RechargeRecord{
		ShopID:       shopID,
		CustomerID:   uint(customerID),
		CardID:       card.ID,
		Type:         4,
		Amount:       req.Amount,
		BalanceAfter: card.Balance,
		Remark:       "余额调整: " + req.Remark,
		OperatorID:   &staffID,
	}
	tx.Create(record)

	tx.Model(&model.Customer{}).Where("id = ?", customerID).Update("member_balance", card.Balance)

	tx.Commit()
	response.Success(c, card)
}

// BalancePayment deducts from member card balance for order payment
func BalancePayment(shopID, customerID, orderID uint, amount float64, staffID uint) error {
	var card model.MemberCard
	if err := database.DB.Where("customer_id = ? AND status = 1", customerID).First(&card).Error; err != nil {
		return fmt.Errorf("该客户没有会员卡")
	}

	if card.Balance < amount {
		return fmt.Errorf("会员余额不足（余额:%.2f 需付:%.2f）", card.Balance, amount)
	}

	card.Balance -= amount
	card.TotalSpent += amount
	database.DB.Save(&card)

	record := &model.RechargeRecord{
		ShopID:       shopID,
		CustomerID:   customerID,
		CardID:       card.ID,
		Type:         2,
		Amount:       amount,
		BalanceAfter: card.Balance,
		OrderID:      &orderID,
		Remark:       "订单消费",
		OperatorID:   &staffID,
	}
	database.DB.Create(record)

	database.DB.Model(&model.Customer{}).Where("id = ?", customerID).Update("member_balance", card.Balance)
	return nil
}
