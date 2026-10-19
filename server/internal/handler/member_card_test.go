package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMemberCardHandlerTestDB(t *testing.T) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	database.DB = db
	if err := database.DB.AutoMigrate(
		&model.Customer{},
		&model.MemberCardTemplate{},
		&model.MemberCard{},
		&model.RechargeRecord{},
		&model.Staff{},
		&model.Order{},
		&model.OrderItem{},
		&model.ServiceCategory{},
		&model.Service{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

func performMemberCardRequest(handler gin.HandlerFunc, customerID uint) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/customers/%d/member-card", customerID), nil)
	c.Request = req
	c.Params = gin.Params{{Key: "id", Value: fmt.Sprintf("%d", customerID)}}
	c.Set("shop_id", uint(1))
	handler(c)
	return w
}

func TestDeleteCardPhysicallyDeletesMistakenCard(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138101", Nickname: "误建卡客户", MemberBalance: 0, DiscountRate: 0.8}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "误建卡", MinRecharge: 100, DiscountRate: 0.8, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "误建卡", Balance: 0, DiscountRate: 0.8, ProductDiscountRate: 1, Status: model.MemberCardStatusActive}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	if err := database.DB.Model(&customer).Updates(map[string]interface{}{"member_card_id": card.ID, "discount_rate": 0.8}).Error; err != nil {
		t.Fatalf("link card: %v", err)
	}

	h := NewMemberCardHandler()
	w := performMemberCardRequest(h.DeleteCard, customer.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("expected delete card status 200, got %d body %s", w.Code, w.Body.String())
	}

	var count int64
	if err := database.DB.Unscoped().Model(&model.MemberCard{}).Where("id = ?", card.ID).Count(&count).Error; err != nil {
		t.Fatalf("count card: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected mistaken card to be physically deleted, got count %d", count)
	}
	var saved model.Customer
	if err := database.DB.First(&saved, customer.ID).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	if saved.MemberCardID != nil || saved.MemberBalance != 0 || saved.DiscountRate != 1 {
		t.Fatalf("expected customer member fields cleared, got card %#v balance %.2f discount %.2f", saved.MemberCardID, saved.MemberBalance, saved.DiscountRate)
	}
}

func TestDeleteCardRejectsCardWithHistory(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138102", Nickname: "有流水客户", MemberBalance: 0, DiscountRate: 0.9}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "历史卡", MinRecharge: 100, DiscountRate: 0.9, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "历史卡", Balance: 0, DiscountRate: 0.9, ProductDiscountRate: 1, Status: model.MemberCardStatusActive}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	record := model.RechargeRecord{ShopID: 1, CustomerID: customer.ID, CardID: card.ID, Type: 4, Amount: -100, BalanceAfter: 0, Remark: "清零"}
	if err := database.DB.Create(&record).Error; err != nil {
		t.Fatalf("create record: %v", err)
	}

	h := NewMemberCardHandler()
	w := performMemberCardRequest(h.DeleteCard, customer.ID)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected delete card status 400, got %d body %s", w.Code, w.Body.String())
	}

	var count int64
	if err := database.DB.Model(&model.MemberCard{}).Where("id = ?", card.ID).Count(&count).Error; err != nil {
		t.Fatalf("count card: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected card with history to remain, got count %d", count)
	}
}

func TestDisableCardClearsCustomerMemberState(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138103", Nickname: "作废卡客户", MemberBalance: 0, DiscountRate: 0.85}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "作废卡", MinRecharge: 100, DiscountRate: 0.85, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "作废卡", Balance: 0, DiscountRate: 0.85, ProductDiscountRate: 1, Status: model.MemberCardStatusActive}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	if err := database.DB.Model(&customer).Updates(map[string]interface{}{"member_card_id": card.ID, "discount_rate": 0.85}).Error; err != nil {
		t.Fatalf("link card: %v", err)
	}

	h := NewMemberCardHandler()
	w := performMemberCardRequest(h.DisableCard, customer.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("expected disable card status 200, got %d body %s", w.Code, w.Body.String())
	}

	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Status != model.MemberCardStatusDisabled {
		t.Fatalf("expected card status disabled, got %d", savedCard.Status)
	}
	var savedCustomer model.Customer
	if err := database.DB.First(&savedCustomer, customer.ID).Error; err != nil {
		t.Fatalf("load customer: %v", err)
	}
	if savedCustomer.MemberCardID != nil || savedCustomer.MemberBalance != 0 || savedCustomer.DiscountRate != 1 {
		t.Fatalf("expected customer member fields cleared, got card %#v balance %.2f discount %.2f", savedCustomer.MemberCardID, savedCustomer.MemberBalance, savedCustomer.DiscountRate)
	}
}

func TestDisableCardUpdatesPendingOrdersAtomically(t *testing.T) {
	for _, failOrderUpdate := range []bool{false, true} {
		t.Run(fmt.Sprintf("rollback=%t", failOrderUpdate), func(t *testing.T) {
			setupMemberCardHandlerTestDB(t)
			customer := model.Customer{ShopID: 1, Phone: "13800138889", DiscountRate: 0.9}
			if err := database.DB.Create(&customer).Error; err != nil {
				t.Fatal(err)
			}
			card := model.MemberCard{ShopID: 1, CustomerID: customer.ID, Balance: 0, DiscountRate: 0.9, ProductDiscountRate: 0.98, Status: model.MemberCardStatusActive}
			if err := database.DB.Create(&card).Error; err != nil {
				t.Fatal(err)
			}
			if err := database.DB.Model(&customer).Update("member_card_id", card.ID).Error; err != nil {
				t.Fatal(err)
			}
			staff := model.Staff{ShopID: 1, Phone: "13800138890", PasswordHash: "hash", CommissionRate: 20}
			if err := database.DB.Create(&staff).Error; err != nil {
				t.Fatal(err)
			}
			otherCustomerID := customer.ID + 1
			pending := model.Order{ShopID: 1, OrderNo: "DISABLE-PENDING", CustomerID: &customer.ID, StaffID: &staff.ID,
				TotalAmount: 342.8, ServiceTotal: 308, ProductTotal: 39.8, AddonTotal: -5,
				ServiceDiscountAmount: 30.8, ProductDiscountAmount: 0.8, DiscountAmount: 31.6, DiscountRate: 0.91,
				PayAmount: 291.2, AppointmentDepositAmount: 20, AppointmentDepositDeductionAmount: 20, Commission: 55.44}
			orders := []model.Order{pending, pending, pending, pending, pending, pending}
			orders[1].OrderNo, orders[1].Status, orders[1].PayStatus = "DISABLE-PAID", 1, 1
			orders[2].OrderNo, orders[2].Status = "DISABLE-CANCELLED", 2
			orders[3].OrderNo, orders[3].ShopID = "DISABLE-OTHER-SHOP", 2
			orders[4].OrderNo, orders[4].CustomerID = "DISABLE-OTHER-CUSTOMER", &otherCustomerID
			orders[5].OrderNo = "DISABLE-MANUAL"
			orders[5].ServiceDiscountAmount, orders[5].ProductDiscountAmount = 0, 0
			for i := range orders {
				if err := database.DB.Create(&orders[i]).Error; err != nil {
					t.Fatal(err)
				}
				items := []model.OrderItem{
					{OrderID: orders[i].ID, ItemType: 1, ItemID: 1, Name: "Service", Quantity: 1, UnitPrice: 308, Amount: 308},
					{OrderID: orders[i].ID, ItemType: 2, ItemID: 2, Name: "Product", Quantity: 2, UnitPrice: 19.9, Amount: 39.8},
					{OrderID: orders[i].ID, ItemType: 3, Name: "Manual adjustment", Quantity: 1, UnitPrice: -5, Amount: -5},
				}
				if err := database.DB.Create(&items).Error; err != nil {
					t.Fatal(err)
				}
			}
			if failOrderUpdate {
				if err := database.DB.Callback().Update().Before("gorm:update").Register("test:fail_order_pricing", func(tx *gorm.DB) {
					if tx.Statement.Table == "orders" {
						tx.AddError(errors.New("order pricing update failed"))
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
			w := performMemberCardRequest(NewMemberCardHandler().DisableCard, customer.ID)
			wantStatus := http.StatusOK
			if failOrderUpdate {
				wantStatus = http.StatusInternalServerError
			}
			if w.Code != wantStatus {
				t.Fatalf("want status %d, got %d: %s", wantStatus, w.Code, w.Body.String())
			}
			var savedCard model.MemberCard
			var savedCustomer model.Customer
			if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
				t.Fatal(err)
			}
			if err := database.DB.First(&savedCustomer, customer.ID).Error; err != nil {
				t.Fatal(err)
			}
			if failOrderUpdate {
				if savedCard.Status != model.MemberCardStatusActive || savedCustomer.MemberCardID == nil || savedCustomer.DiscountRate != 0.9 {
					t.Fatal("failed order update must roll back card and customer changes")
				}
			} else if savedCard.Status != model.MemberCardStatusDisabled || savedCustomer.MemberCardID != nil || savedCustomer.DiscountRate != 1 {
				t.Fatal("card and customer were not disabled")
			}
			for i, before := range orders {
				var after model.Order
				if err := database.DB.First(&after, before.ID).Error; err != nil {
					t.Fatal(err)
				}
				if i == 0 && !failOrderUpdate {
					if after.PayAmount != 322.8 || after.DiscountAmount != 0 || after.ServiceDiscountAmount != 0 || after.ProductDiscountAmount != 0 || after.Commission != 61.6 {
						t.Fatalf("pending pricing was not persisted: %+v", after)
					}
					if after.AddonTotal != -5 || after.AppointmentDepositDeductionAmount != 20 || after.PayStatus != 0 {
						t.Fatal("manual adjustment, deposit, and payment state must remain unchanged")
					}
				} else if after.PayAmount != before.PayAmount || after.DiscountAmount != before.DiscountAmount || after.Commission != before.Commission || after.PayStatus != before.PayStatus {
					t.Fatalf("unrelated or rolled back order %s changed", before.OrderNo)
				}
			}
		})
	}
}

func TestBalancePaymentUsesServiceCategoryCommissionOverride(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{ShopID: 1, Phone: "13800138032", Nickname: "赛洗余额客户", MemberBalance: 100, DiscountRate: 1}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "赛洗卡", MinRecharge: 100, DiscountRate: 1, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "赛洗卡", Balance: 100, DiscountRate: 1, ProductDiscountRate: 1, Status: 1}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{ShopID: 1, Phone: "13800139032", PasswordHash: "hash", Name: "曼露", Role: model.StaffRoleStaff, Status: 1, CommissionRate: 20}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	overrideRate := 16.0
	root := model.ServiceCategory{ShopID: 1, Name: "赛级沙龙", CommissionRate: &overrideRate, Status: 1}
	if err := database.DB.Create(&root).Error; err != nil {
		t.Fatalf("create root category: %v", err)
	}
	child := model.ServiceCategory{ShopID: 1, ParentID: &root.ID, Name: "SO POSH", Status: 1}
	if err := database.DB.Create(&child).Error; err != nil {
		t.Fatalf("create child category: %v", err)
	}
	serviceItem := model.Service{ShopID: 1, CategoryID: &child.ID, Name: "赛级皮毛护理", BasePrice: 100, Duration: 60, Status: 1}
	if err := database.DB.Create(&serviceItem).Error; err != nil {
		t.Fatalf("create service: %v", err)
	}
	order := model.Order{OrderNo: "TEST-BALANCE-CATEGORY-COMMISSION", ShopID: 1, CustomerID: &customer.ID, StaffID: &staff.ID, ServiceTotal: 100, PayAmount: 100, PayStatus: 0, TotalAmount: 100, Commission: 20}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	item := model.OrderItem{OrderID: order.ID, ItemType: 1, ItemID: serviceItem.ID, Name: serviceItem.Name, Quantity: 1, UnitPrice: 100, Amount: 100}
	if err := database.DB.Create(&item).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 100, staff.ID); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if got := saved.Commission; got != 16 {
		t.Fatalf("expected category commission override 16.00 after balance payment, got %.2f", got)
	}
}

func TestBalancePaymentRecalculatesDiscountWhenCardOpenedAfterOrder(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138033",
		Nickname:      "开单后办卡客户",
		MemberBalance: 1000,
		DiscountRate:  0.9,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "I", MinRecharge: 1000, DiscountRate: 0.9, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "I",
		Balance:             1000,
		DiscountRate:        0.9,
		ProductDiscountRate: 0.98,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{ShopID: 1, Phone: "13800139033", PasswordHash: "hash", Name: "曼露", Role: model.StaffRoleStaff, Status: 1, CommissionRate: 20}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:      "TEST-CARD-OPENED-AFTER-ORDER",
		ShopID:       1,
		CustomerID:   &customer.ID,
		StaffID:      &staff.ID,
		ServiceTotal: 580,
		TotalAmount:  580,
		PayAmount:    580,
		DiscountRate: 1,
		PayStatus:    0,
		Status:       0,
		Commission:   116,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := database.DB.Create(&model.OrderItem{
		OrderID: order.ID, ItemType: 1, Name: "基础洗护", Quantity: 1, UnitPrice: 580, Amount: 580,
	}).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, order.PayAmount, staff.ID); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var savedOrder model.Order
	if err := database.DB.First(&savedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if savedOrder.PayAmount != 522 || savedOrder.ServiceDiscountAmount != 58 || savedOrder.DiscountAmount != 58 || savedOrder.DiscountRate != 0.9 {
		t.Fatalf("expected order repriced to 522.00 with 58.00 discount, got pay %.2f service discount %.2f total discount %.2f rate %.2f", savedOrder.PayAmount, savedOrder.ServiceDiscountAmount, savedOrder.DiscountAmount, savedOrder.DiscountRate)
	}
	if savedOrder.MemberBalanceBefore == nil || *savedOrder.MemberBalanceBefore != 1000 || savedOrder.MemberBalanceAfter == nil || *savedOrder.MemberBalanceAfter != 478 {
		t.Fatalf("expected balance snapshot 1000.00 -> 478.00, got before %#v after %#v", savedOrder.MemberBalanceBefore, savedOrder.MemberBalanceAfter)
	}

	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Balance != 478 || savedCard.TotalSpent != 522 {
		t.Fatalf("expected card balance 478.00 and spent 522.00, got balance %.2f spent %.2f", savedCard.Balance, savedCard.TotalSpent)
	}
}

func TestBalancePaymentRemovesStaleDiscountWhenCurrentRateIsFullPrice(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138034",
		Nickname:      "折扣已取消客户",
		MemberBalance: 100,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "无折扣卡", MinRecharge: 100, DiscountRate: 1, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "无折扣卡",
		Balance:             100,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              model.MemberCardStatusActive,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	order := model.Order{
		OrderNo:               "TEST-BALANCE-PAYMENT-REMOVE-STALE-DISCOUNT",
		ShopID:                1,
		CustomerID:            &customer.ID,
		ServiceTotal:          100,
		TotalAmount:           100,
		ServiceDiscountAmount: 10,
		DiscountAmount:        10,
		DiscountRate:          0.9,
		PayAmount:             90,
		PayStatus:             0,
		Status:                0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, order.PayAmount, 1); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var savedOrder model.Order
	if err := database.DB.First(&savedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if savedOrder.PayAmount != 100 || savedOrder.DiscountAmount != 0 || savedOrder.ServiceDiscountAmount != 0 || savedOrder.DiscountRate != 1 {
		t.Fatalf("expected stale discount removed and pay amount 100, got pay %.2f discount %.2f service discount %.2f rate %.2f", savedOrder.PayAmount, savedOrder.DiscountAmount, savedOrder.ServiceDiscountAmount, savedOrder.DiscountRate)
	}
	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Balance != 0 {
		t.Fatalf("expected full-price 100 deduction, got balance %.2f", savedCard.Balance)
	}
}

func TestBalancePaymentSnapshotsOrderMemberBalance(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138002",
		Nickname:      "余额客户",
		MemberBalance: 100,
		DiscountRate:  0.9,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "金卡",
		MinRecharge:  100,
		DiscountRate: 0.9,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "金卡",
		Balance:             100,
		TotalSpent:          0,
		DiscountRate:        0.9,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139002",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:               "TEST-BALANCE-PAYMENT-SNAPSHOT",
		ShopID:                1,
		CustomerID:            &customer.ID,
		StaffID:               &staff.ID,
		ServiceTotal:          100,
		ServiceDiscountAmount: 10,
		PayAmount:             90,
		PayStatus:             0,
		PayMethod:             "",
		TotalAmount:           100,
		Commission:            20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 90, 1); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.MemberBalanceBefore == nil || *saved.MemberBalanceBefore != 100 {
		t.Fatalf("expected member balance before snapshot 100.00, got %#v", saved.MemberBalanceBefore)
	}
	if saved.MemberBalanceAfter == nil || *saved.MemberBalanceAfter != 10 {
		t.Fatalf("expected member balance after snapshot 10.00, got %#v", saved.MemberBalanceAfter)
	}
	if saved.Commission != 18 {
		t.Fatalf("expected commission recalculated from discounted service amount 18.00, got %.2f", saved.Commission)
	}
}

func TestBalancePaymentDoesNotCalculateCommissionForBoardingOrder(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138022",
		Nickname:      "寄养余额客户",
		MemberBalance: 500,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "寄养卡",
		MinRecharge:  100,
		DiscountRate: 1,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "寄养卡",
		Balance:             500,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139022",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:      "TEST-BALANCE-BOARDING-NO-COMMISSION",
		ShopID:       1,
		CustomerID:   &customer.ID,
		StaffID:      &staff.ID,
		ServiceTotal: 300,
		PayAmount:    300,
		PayStatus:    0,
		Status:       0,
		TotalAmount:  300,
		Commission:   0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := database.DB.Create(&model.OrderItem{
		OrderID:   order.ID,
		ItemType:  4,
		Name:      "寄养住宿",
		Quantity:  3,
		UnitPrice: 100,
		Amount:    300,
	}).Error; err != nil {
		t.Fatalf("create boarding order item: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 300, staff.ID); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.Commission != 0 {
		t.Fatalf("expected boarding order commission 0, got %.2f", saved.Commission)
	}
}

func TestBalancePaymentDoesNotApplyMemberDiscountToFeedingOrder(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138042",
		Nickname:      "上门喂养余额客户",
		MemberBalance: 400,
		DiscountRate:  0.9,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "I", MinRecharge: 1000, DiscountRate: 0.9, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "I",
		Balance: 400, DiscountRate: 0.9, ProductDiscountRate: 0.98, Status: 1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	feedingPlanID := uint(25)
	order := model.Order{
		OrderNo: "TEST-BALANCE-FEEDING-NO-DISCOUNT", ShopID: 1, CustomerID: &customer.ID,
		FeedingPlanID: &feedingPlanID, ServiceTotal: 272, AddonTotal: 20,
		TotalAmount: 292, PayAmount: 292, DiscountRate: 1, PayStatus: 0, Status: 0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, order.PayAmount, 1); err != nil {
		t.Fatalf("balance payment: %v", err)
	}

	var savedOrder model.Order
	if err := database.DB.First(&savedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if savedOrder.PayAmount != 292 || savedOrder.ServiceDiscountAmount != 0 || savedOrder.DiscountAmount != 0 || savedOrder.DiscountRate != 1 {
		t.Fatalf("expected feeding order to remain undiscounted at 292.00, got pay %.2f service discount %.2f total discount %.2f rate %.2f", savedOrder.PayAmount, savedOrder.ServiceDiscountAmount, savedOrder.DiscountAmount, savedOrder.DiscountRate)
	}
	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Balance != 108 || savedCard.TotalSpent != 292 {
		t.Fatalf("expected feeding order to deduct full 292.00, got balance %.2f spent %.2f", savedCard.Balance, savedCard.TotalSpent)
	}
}

func TestMixedBalancePaymentDoesNotApplyMemberDiscountToFeedingOrder(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138043",
		Nickname:      "上门喂养补差客户",
		MemberBalance: 100,
		DiscountRate:  0.9,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{ShopID: 1, Name: "I", MinRecharge: 1000, DiscountRate: 0.9, Status: 1}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID: 1, CustomerID: customer.ID, TemplateID: template.ID, CardName: "I",
		Balance: 100, DiscountRate: 0.9, ProductDiscountRate: 0.98, Status: 1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	feedingPlanID := uint(26)
	order := model.Order{
		OrderNo: "TEST-MIXED-BALANCE-FEEDING-NO-DISCOUNT", ShopID: 1, CustomerID: &customer.ID,
		FeedingPlanID: &feedingPlanID, ServiceTotal: 272, AddonTotal: 20,
		TotalAmount: 292, PayAmount: 292, DiscountRate: 1, PayStatus: 0, Status: 0,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	result, err := MixedBalancePayment(1, customer.ID, order.ID, "qrcode", 1)
	if err != nil {
		t.Fatalf("mixed balance payment: %v", err)
	}
	if result.FinalPayAmount != 292 || result.BalanceUsed != 100 || result.CashPayAmount != 192 {
		t.Fatalf("expected feeding mixed payment 292.00 = 100.00 balance + 192.00 cash, got final %.2f balance %.2f cash %.2f", result.FinalPayAmount, result.BalanceUsed, result.CashPayAmount)
	}
	var savedOrder model.Order
	if err := database.DB.First(&savedOrder, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if savedOrder.ServiceDiscountAmount != 0 || savedOrder.DiscountAmount != 0 || savedOrder.DiscountRate != 1 {
		t.Fatalf("expected feeding mixed payment without discount, got service discount %.2f total discount %.2f rate %.2f", savedOrder.ServiceDiscountAmount, savedOrder.DiscountAmount, savedOrder.DiscountRate)
	}
}

func TestBalancePaymentRejectsAlreadyPaidOrderBeforeDeductingBalance(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138004",
		Nickname:      "已支付客户",
		MemberBalance: 100,
		DiscountRate:  1,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "金卡",
		MinRecharge:  100,
		DiscountRate: 1,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "金卡",
		Balance:             100,
		TotalSpent:          20,
		DiscountRate:        1,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139004",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:      "TEST-BALANCE-ALREADY-PAID",
		ShopID:       1,
		CustomerID:   &customer.ID,
		StaffID:      &staff.ID,
		ServiceTotal: 100,
		PayAmount:    100,
		PayStatus:    1,
		Status:       1,
		PayMethod:    "balance",
		TotalAmount:  100,
		Commission:   20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	if err := BalancePayment(1, customer.ID, order.ID, 100, 1); err == nil {
		t.Fatalf("expected already paid error")
	}

	var savedCard model.MemberCard
	if err := database.DB.First(&savedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if savedCard.Balance != 100 || savedCard.TotalSpent != 20 {
		t.Fatalf("expected card unchanged, got balance %.2f spent %.2f", savedCard.Balance, savedCard.TotalSpent)
	}
}

func TestMixedBalancePaymentUsesBalanceDiscountOnlyForCoveredAmount(t *testing.T) {
	setupMemberCardHandlerTestDB(t)

	customer := model.Customer{
		ShopID:        1,
		Phone:         "13800138003",
		Nickname:      "余额不足客户",
		MemberBalance: 60,
		DiscountRate:  0.8,
	}
	if err := database.DB.Create(&customer).Error; err != nil {
		t.Fatalf("create customer: %v", err)
	}
	template := model.MemberCardTemplate{
		ShopID:       1,
		Name:         "八折卡",
		MinRecharge:  100,
		DiscountRate: 0.8,
		Status:       1,
	}
	if err := database.DB.Create(&template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}
	card := model.MemberCard{
		ShopID:              1,
		CustomerID:          customer.ID,
		TemplateID:          template.ID,
		CardName:            "八折卡",
		Balance:             60,
		DiscountRate:        0.8,
		ProductDiscountRate: 1,
		Status:              1,
	}
	if err := database.DB.Create(&card).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}
	staff := model.Staff{
		ShopID:         1,
		Phone:          "13800139003",
		PasswordHash:   "hash",
		Name:           "乐乐",
		Role:           model.StaffRoleStaff,
		Status:         1,
		CommissionRate: 20,
	}
	if err := database.DB.Create(&staff).Error; err != nil {
		t.Fatalf("create staff: %v", err)
	}
	order := model.Order{
		OrderNo:               "TEST-MIXED-BALANCE-PAYMENT",
		ShopID:                1,
		CustomerID:            &customer.ID,
		StaffID:               &staff.ID,
		TotalAmount:           100,
		ServiceTotal:          100,
		ServiceDiscountAmount: 20,
		DiscountAmount:        20,
		DiscountRate:          0.8,
		PayAmount:             80,
		PayStatus:             0,
		Status:                0,
		Commission:            20,
	}
	if err := database.DB.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}

	result, err := MixedBalancePayment(1, customer.ID, order.ID, "qrcode", 1)
	if err != nil {
		t.Fatalf("mixed balance payment: %v", err)
	}

	if result.BalanceUsed != 60 {
		t.Fatalf("expected balance used 60.00, got %.2f", result.BalanceUsed)
	}
	if result.CashPayAmount != 25 {
		t.Fatalf("expected cash pay amount 25.00, got %.2f", result.CashPayAmount)
	}
	if result.FinalPayAmount != 85 {
		t.Fatalf("expected final pay amount 85.00, got %.2f", result.FinalPayAmount)
	}

	var saved model.Order
	if err := database.DB.First(&saved, order.ID).Error; err != nil {
		t.Fatalf("load order: %v", err)
	}
	if saved.PayMethod != "mixed_balance" {
		t.Fatalf("expected pay method mixed_balance, got %q", saved.PayMethod)
	}
	if saved.CashPayMethod != "qrcode" {
		t.Fatalf("expected cash pay method qrcode, got %q", saved.CashPayMethod)
	}
	if saved.MemberBalanceUsed != 60 {
		t.Fatalf("expected saved balance used 60.00, got %.2f", saved.MemberBalanceUsed)
	}
	if saved.CashPayAmount != 25 {
		t.Fatalf("expected saved cash pay amount 25.00, got %.2f", saved.CashPayAmount)
	}
	if saved.PayAmount != 85 {
		t.Fatalf("expected saved pay amount 85.00, got %.2f", saved.PayAmount)
	}
	if saved.ServiceDiscountAmount != 15 || saved.DiscountAmount != 15 {
		t.Fatalf("expected partial discount 15.00, got service %.2f total %.2f", saved.ServiceDiscountAmount, saved.DiscountAmount)
	}
	if saved.Commission != 17 {
		t.Fatalf("expected commission recalculated from discounted service amount 17.00, got %.2f", saved.Commission)
	}
	if saved.MemberBalanceBefore == nil || *saved.MemberBalanceBefore != 60 {
		t.Fatalf("expected member balance before 60.00, got %#v", saved.MemberBalanceBefore)
	}
	if saved.MemberBalanceAfter == nil || *saved.MemberBalanceAfter != 0 {
		t.Fatalf("expected member balance after 0.00, got %#v", saved.MemberBalanceAfter)
	}

	var reloadedCard model.MemberCard
	if err := database.DB.First(&reloadedCard, card.ID).Error; err != nil {
		t.Fatalf("load card: %v", err)
	}
	if reloadedCard.Balance != 0 || reloadedCard.TotalSpent != 60 {
		t.Fatalf("expected card balance 0 and spent 60, got balance %.2f spent %.2f", reloadedCard.Balance, reloadedCard.TotalSpent)
	}
}
