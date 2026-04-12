package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/utils"
	"gorm.io/gorm"
)

type OrderService struct {
	orderRepo *repository.OrderRepository
	apptRepo  *repository.AppointmentRepository
}

func NewOrderService(orderRepo *repository.OrderRepository, apptRepo *repository.AppointmentRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo, apptRepo: apptRepo}
}

func resolveMemberDiscountRates(customerID *uint) (float64, float64) {
	serviceDiscountRate := 1.0
	productDiscountRate := 1.0
	if customerID == nil || *customerID == 0 {
		return serviceDiscountRate, productDiscountRate
	}

	var customer model.Customer
	if err := database.DB.Select("id", "discount_rate").First(&customer, *customerID).Error; err == nil {
		if customer.DiscountRate > 0 && customer.DiscountRate < 1 {
			serviceDiscountRate = customer.DiscountRate
		}
	}

	var card model.MemberCard
	if err := database.DB.Where("customer_id = ? AND status = 1", *customerID).First(&card).Error; err == nil {
		if serviceDiscountRate == 1 && card.DiscountRate > 0 && card.DiscountRate < 1 {
			serviceDiscountRate = card.DiscountRate
		}
		if card.ProductDiscountRate > 0 && card.ProductDiscountRate < 1 {
			productDiscountRate = card.ProductDiscountRate
		}
	}

	return serviceDiscountRate, productDiscountRate
}

func roundOrderAmount(value float64) float64 {
	return math.Round(value*100) / 100
}

func calculateOrderDiscountRate(totalAmount, payAmount float64) float64 {
	if totalAmount <= 0 {
		return 1
	}
	return roundOrderAmount(payAmount / totalAmount)
}

func applyMemberDiscountToOrder(order *model.Order, customerID *uint, serviceTotal, productTotal, addonTotal float64) {
	serviceDiscountRate, productDiscountRate := resolveMemberDiscountRates(customerID)

	serviceTotal = roundOrderAmount(serviceTotal)
	productTotal = roundOrderAmount(productTotal)
	addonTotal = roundOrderAmount(addonTotal)

	servicePayAmount := roundOrderAmount(serviceTotal * serviceDiscountRate)
	productPayAmount := roundOrderAmount(productTotal * productDiscountRate)
	serviceDiscountAmount := roundOrderAmount(serviceTotal - servicePayAmount)
	productDiscountAmount := roundOrderAmount(productTotal - productPayAmount)
	totalAmount := roundOrderAmount(serviceTotal + productTotal + addonTotal)
	payAmount := roundOrderAmount(servicePayAmount + productPayAmount + addonTotal)
	discountAmount := roundOrderAmount(serviceDiscountAmount + productDiscountAmount)

	order.TotalAmount = totalAmount
	order.ServiceTotal = serviceTotal
	order.ProductTotal = productTotal
	order.AddonTotal = addonTotal
	order.ServiceDiscountAmount = serviceDiscountAmount
	order.ProductDiscountAmount = productDiscountAmount
	order.DiscountAmount = discountAmount
	order.PayAmount = payAmount
	order.DiscountRate = calculateOrderDiscountRate(totalAmount, payAmount)
}

// CreateFromAppointment generates an order from a completed appointment
func (s *OrderService) CreateFromAppointment(appointmentID uint) (*model.Order, error) {
	existingCount, err := s.orderRepo.CountByAppointment(appointmentID)
	if err != nil {
		return nil, err
	}
	if existingCount > 0 {
		return nil, errors.New("该预约已开单")
	}

	appt, err := s.apptRepo.FindByID(appointmentID)
	if err != nil {
		return nil, errors.New("预约不存在")
	}

	custID := appt.CustomerID
	order := &model.Order{
		OrderNo:       utils.GenerateOrderNo(),
		ShopID:        appt.ShopID,
		CustomerID:    &custID,
		AppointmentID: &appt.ID,
		StaffID:       appt.StaffID,
	}

	var items []model.OrderItem
	var serviceTotal float64
	if len(appt.Pets) > 0 {
		for _, apptPet := range appt.Pets {
			petName := "宠物"
			if apptPet.Pet != nil && apptPet.Pet.Name != "" {
				petName = apptPet.Pet.Name
			}
			for _, svc := range apptPet.Services {
				items = append(items, model.OrderItem{
					ItemType:  1, // service
					ItemID:    svc.ServiceID,
					Name:      petName + " · " + svc.ServiceName,
					Quantity:  1,
					UnitPrice: svc.Price,
					Amount:    svc.Price,
				})
				serviceTotal += svc.Price
			}
		}
	} else {
		for _, svc := range appt.Services {
			items = append(items, model.OrderItem{
				ItemType:  1, // service
				ItemID:    svc.ServiceID,
				Name:      svc.ServiceName,
				Quantity:  1,
				UnitPrice: svc.Price,
				Amount:    svc.Price,
			})
			serviceTotal += svc.Price
		}
	}

	applyMemberDiscountToOrder(order, &custID, serviceTotal, 0, 0)

	tx := database.DB.Begin()
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(items) > 0 {
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	_ = s.syncAppointmentSettlement(appointmentID)
	return order, nil
}

type ServiceOverride struct {
	ServiceID   uint
	ServiceName string
	Price       float64
	Duration    int
}

type ProductOverride struct {
	ProductID uint
	SKUID     uint
	Name      string
	Price     float64
	Quantity  int
}

type PetOverrideData struct {
	Services []ServiceOverride
	Products []ProductOverride
}

func (s *OrderService) CreateSplitFromAppointment(appointmentID uint, overrides ...map[uint]PetOverrideData) ([]model.Order, error) {
	var overrideMap map[uint]PetOverrideData
	if len(overrides) > 0 && overrides[0] != nil {
		overrideMap = overrides[0]
	}

	appt, err := s.apptRepo.FindByID(appointmentID)
	if err != nil {
		return nil, errors.New("预约不存在")
	}

	validPets := filterBillableAppointmentPets(appt)

	existingOrders, err := s.orderRepo.FindByAppointment(appointmentID)
	if err != nil {
		return nil, err
	}
	if hasConflictingOrders(existingOrders, validPets) {
		return nil, errors.New("该预约已开单")
	}

	if len(validPets) == 0 {
		order, err := s.CreateFromAppointment(appointmentID)
		if err != nil {
			return nil, err
		}
		return []model.Order{*order}, nil
	}

	tx := database.DB.Begin()
	custID := appt.CustomerID
	order := model.Order{
		OrderNo:       utils.GenerateOrderNo(),
		ShopID:        appt.ShopID,
		CustomerID:    &custID,
		PetID:         nil,
		AppointmentID: &appt.ID,
		StaffID:       appt.StaffID,
		Remark:        appt.Notes,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	items := make([]model.OrderItem, 0)
	var serviceTotal float64
	var productTotal float64
	for _, apptPet := range validPets {
		petName := "宠物"
		if apptPet.Pet != nil && apptPet.Pet.Name != "" {
			petName = apptPet.Pet.Name
		}

		serviceList := buildOrderServices(apptPet, overrideMap)
		if len(serviceList) == 0 {
			tx.Rollback()
			return nil, errors.New("请至少保留一项服务后再开单")
		}

		for _, svc := range serviceList {
			name := svc.ServiceName
			if name == "" {
				name = "服务项目"
			}
			items = append(items, model.OrderItem{
				OrderID:   order.ID,
				ItemType:  1,
				ItemID:    svc.ServiceID,
				Name:      fmt.Sprintf("%s · %s", petName, name),
				Quantity:  1,
				UnitPrice: svc.Price,
				Amount:    svc.Price,
			})
			serviceTotal += svc.Price
		}

		// 商品项
		if overrideMap != nil {
			if petData, ok := overrideMap[apptPet.PetID]; ok {
				for _, prod := range petData.Products {
					qty := prod.Quantity
					if qty < 1 {
						qty = 1
					}
					amount := prod.Price * float64(qty)
					items = append(items, model.OrderItem{
						OrderID:   order.ID,
						ItemType:  2,
						ItemID:    prod.ProductID,
						Name:      fmt.Sprintf("%s · %s", petName, prod.Name),
						Quantity:  qty,
						UnitPrice: prod.Price,
						Amount:    amount,
					})
					productTotal += amount
				}
			}
		}
	}
	applyMemberDiscountToOrder(&order, &custID, serviceTotal, productTotal, 0)
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	_ = s.syncAppointmentSettlement(appointmentID)
	result, err := s.GetByID(order.ID)
	if err != nil {
		return nil, err
	}
	return []model.Order{*result}, nil
}

func filterBillableAppointmentPets(appt *model.Appointment) []model.AppointmentPet {
	if appt == nil || len(appt.Pets) == 0 {
		return nil
	}

	filtered := make([]model.AppointmentPet, 0, len(appt.Pets))
	seen := make(map[uint]struct{}, len(appt.Pets))
	for _, petItem := range appt.Pets {
		if petItem.PetID == 0 {
			continue
		}
		if _, exists := seen[petItem.PetID]; exists {
			continue
		}
		if petItem.Pet != nil && petItem.Pet.CustomerID != nil && *petItem.Pet.CustomerID != appt.CustomerID {
			continue
		}
		seen[petItem.PetID] = struct{}{}
		filtered = append(filtered, petItem)
	}
	return filtered
}

func hasConflictingOrders(existingOrders []model.Order, apptPets []model.AppointmentPet) bool {
	activeOrders := filterActiveAppointmentOrders(existingOrders)
	if len(activeOrders) == 0 {
		return false
	}

	validPetIDs := make(map[uint]struct{}, len(apptPets))
	for _, petItem := range apptPets {
		if petItem.PetID > 0 {
			validPetIDs[petItem.PetID] = struct{}{}
		}
	}

	for _, order := range activeOrders {
		if order.PetID == nil || *order.PetID == 0 {
			return true
		}
		if _, exists := validPetIDs[*order.PetID]; exists {
			return true
		}
	}
	return false
}

func filterActiveAppointmentOrders(orders []model.Order) []model.Order {
	if len(orders) == 0 {
		return nil
	}

	filtered := make([]model.Order, 0, len(orders))
	for _, order := range orders {
		if order.Status != 0 && order.Status != 1 {
			continue
		}
		filtered = append(filtered, order)
	}
	return filtered
}

func buildOrderServices(apptPet model.AppointmentPet, overrideMap map[uint]PetOverrideData) []ServiceOverride {
	if overrideMap != nil {
		if petData, ok := overrideMap[apptPet.PetID]; ok {
			services := make([]ServiceOverride, 0, len(petData.Services))
			for _, override := range petData.Services {
				services = append(services, override)
			}
			return services
		}
	}

	services := make([]ServiceOverride, 0, len(apptPet.Services))
	for _, svc := range apptPet.Services {
		services = append(services, ServiceOverride{
			ServiceID:   svc.ServiceID,
			ServiceName: svc.ServiceName,
			Price:       svc.Price,
			Duration:    svc.Duration,
		})
	}
	return services
}

// CreateDirect creates a standalone order (walk-in / direct billing)
func (s *OrderService) CreateDirect(order *model.Order, items []model.OrderItem) error {
	order.OrderNo = utils.GenerateOrderNo()

	var total, serviceTotal, productTotal, addonTotal float64
	for i := range items {
		items[i].Amount = items[i].UnitPrice * float64(items[i].Quantity)
		total += items[i].Amount
		switch items[i].ItemType {
		case 1:
			serviceTotal += items[i].Amount
		case 2:
			productTotal += items[i].Amount
		case 3:
			addonTotal += items[i].Amount
		}
	}
	order.TotalAmount = total
	order.ServiceTotal = serviceTotal
	order.ProductTotal = productTotal
	order.AddonTotal = addonTotal
	order.PayAmount = total - order.DiscountAmount

	tx := database.DB.Begin()
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return err
	}
	for i := range items {
		items[i].OrderID = order.ID
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (s *OrderService) UpdateDraft(shopID, id uint, patch *model.Order, items []model.OrderItem) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.ShopID != shopID {
		return errors.New("订单不存在")
	}
	if order.PayStatus != 0 {
		return errors.New("已支付订单不可修改价格")
	}
	if order.Status == 2 || order.Status == 3 {
		return errors.New("当前订单状态不可修改价格")
	}
	if len(items) == 0 {
		return errors.New("请添加商品或服务")
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.CustomerID = patch.CustomerID
		order.PetID = patch.PetID
		order.StaffID = patch.StaffID
		order.TotalAmount = patch.TotalAmount
		order.ServiceTotal = patch.ServiceTotal
		order.ProductTotal = patch.ProductTotal
		order.AddonTotal = patch.AddonTotal
		order.DiscountRate = patch.DiscountRate
		order.DiscountAmount = patch.DiscountAmount
		order.ServiceDiscountAmount = patch.ServiceDiscountAmount
		order.ProductDiscountAmount = patch.ProductDiscountAmount
		order.PayAmount = patch.PayAmount
		order.Commission = patch.Commission
		order.Remark = patch.Remark

		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if err := tx.Where("order_id = ?", order.ID).Delete(&model.OrderItem{}).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].OrderID = order.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	_ = s.syncAppointmentSettlementPtr(order.AppointmentID)
	return nil
}

func (s *OrderService) GetByID(id uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	s.decorateOrder(order)
	return order, nil
}

func (s *OrderService) GetByIDIncludeDeleted(id uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByIDUnscoped(id)
	if err != nil {
		return nil, err
	}
	s.decorateOrder(order)
	return order, nil
}

func (s *OrderService) ListPaged(shopID uint, f repository.OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.orderRepo.FindByShopPaged(shopID, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.decorateOrders(list)
	return list, total, nil
}

func (s *OrderService) Search(shopID uint, keyword string, f repository.OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.orderRepo.Search(shopID, keyword, f, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	s.decorateOrders(list)
	return list, total, nil
}

// MarkPaid marks an order as paid
func (s *OrderService) MarkPaid(id uint, payMethod, transactionID string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.PayStatus == 1 {
		return errors.New("订单已支付")
	}

	now := time.Now()
	order.PayStatus = 1
	order.PayMethod = payMethod
	order.PayTime = &now
	order.TransactionID = transactionID
	order.Status = 1 // completed

	// 美团订单提成打9折
	if payMethod == "meituan" && order.Commission > 0 {
		order.Commission = math.Round(order.Commission*90) / 100
	}

	if err := s.orderRepo.Update(order); err != nil {
		return err
	}

	// 更新客户最近到店时间和到店次数
	if order.CustomerID != nil {
		database.DB.Model(&model.Customer{}).Where("id = ?", *order.CustomerID).
			Updates(map[string]interface{}{
				"last_visit_at": now,
				"visit_count":   gorm.Expr("visit_count + 1"),
			})
	}

	_ = s.syncAppointmentSettlementPtr(order.AppointmentID)

	return nil
}

// Refund processes a refund
func (s *OrderService) Refund(id uint, remark string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.PayStatus != 1 {
		return errors.New("订单未支付，无法退款")
	}

	order.PayStatus = 2
	order.Status = 3
	order.Remark = remark
	if err := s.orderRepo.Update(order); err != nil {
		return err
	}
	_ = s.syncAppointmentSettlementPtr(order.AppointmentID)
	return nil
}

func (s *OrderService) UpdateRemark(id uint, remark string) error {
	if _, err := s.orderRepo.FindByID(id); err != nil {
		return errors.New("订单不存在")
	}
	return s.orderRepo.UpdateRemark(id, remark)
}

// Cancel cancels an unpaid order
func (s *OrderService) Cancel(id uint) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.Status != 0 {
		return errors.New("仅待付款订单可取消")
	}

	order.Status = 2
	if err := s.orderRepo.Update(order); err != nil {
		return err
	}
	_ = s.syncAppointmentSettlementPtr(order.AppointmentID)
	return nil
}

func (s *OrderService) Delete(shopID, id uint, role ...string) error {
	order, err := s.orderRepo.FindByID(id)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.ShopID != shopID {
		return errors.New("订单不存在")
	}
	callerRole := ""
	if len(role) > 0 {
		callerRole = role[0]
	}
	isManager := model.HasStaffRoleAtLeast(callerRole, model.StaffRoleManager)
	if !isManager && order.Status != 0 && order.Status != 1 && order.Status != 2 && order.Status != 3 {
		return errors.New("仅待付款、待结算、已取消或已退款订单可删除")
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", order.ID).Delete(&model.OrderItem{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.Order{}, order.ID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	_ = s.syncAppointmentSettlementPtr(order.AppointmentID)
	return nil
}

func (s *OrderService) ListDeleted(shopID uint, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	since := time.Now().Add(-48 * time.Hour)
	list, total, err := s.orderRepo.FindDeleted(shopID, page, pageSize, since)
	if err != nil {
		return nil, 0, err
	}
	s.decorateOrders(list)
	return list, total, nil
}

func (s *OrderService) Restore(shopID, id uint) error {
	var deletedOrder model.Order
	if err := database.DB.Unscoped().Where("id = ? AND shop_id = ? AND deleted_at IS NOT NULL", id, shopID).First(&deletedOrder).Error; err != nil {
		return errors.New("订单不存在")
	}
	if deletedOrder.DeletedAt.Time.Before(time.Now().Add(-48 * time.Hour)) {
		return errors.New("订单已超过 2 天恢复期限")
	}
	if err := s.orderRepo.Restore(shopID, id); err != nil {
		return err
	}
	_ = s.syncAppointmentSettlementPtr(deletedOrder.AppointmentID)
	return nil
}

func (s *OrderService) decorateOrders(list []model.Order) {
	for i := range list {
		s.decorateOrder(&list[i])
	}
}

func (s *OrderService) decorateOrder(order *model.Order) {
	if order == nil {
		return
	}
	decorateOrderTotals(order)
	order.PetGroups = buildOrderPetGroups(order)
	order.PetSummary = buildOrderPetSummary(order)
	order.OrderKind = buildOrderKind(order)
}

func buildOrderPetGroups(order *model.Order) []model.OrderPetGroup {
	if order == nil {
		return nil
	}
	appointmentPetIDs := appointmentPetIDsByName(order)
	feedingPetIDs := feedingPetIDsByName(order)
	singleFeedingPetName, singleFeedingPetID := singleFeedingPet(order)
	groupMap := make(map[string]*model.OrderPetGroup)
	groupOrder := make([]string, 0)
	for _, item := range order.Items {
		groupKey := ""
		itemName := item.Name
		if item.ItemType == 2 {
			groupKey = "零售商品"
		} else {
			petName, strippedName := splitOrderItemPetName(item.Name)
			groupKey = petName
			if strippedName != "" {
				itemName = strippedName
			}
			if groupKey == "" {
				if order.Pet != nil && order.Pet.Name != "" {
					groupKey = order.Pet.Name
				} else if singleFeedingPetName != "" {
					groupKey = singleFeedingPetName
				} else {
					groupKey = "未分组"
				}
			}
		}
		group, exists := groupMap[groupKey]
		if !exists {
			group = &model.OrderPetGroup{PetName: groupKey}
			if order.Pet != nil && order.Pet.Name == groupKey && order.PetID != nil && *order.PetID > 0 {
				group.PetID = order.PetID
			} else if petID, ok := appointmentPetIDs[groupKey]; ok && petID > 0 {
				group.PetID = uintPtr(petID)
			} else if petID, ok := feedingPetIDs[groupKey]; ok && petID > 0 {
				group.PetID = uintPtr(petID)
			} else if singleFeedingPetName == groupKey && singleFeedingPetID > 0 {
				group.PetID = uintPtr(singleFeedingPetID)
			}
			groupMap[groupKey] = group
			groupOrder = append(groupOrder, groupKey)
		}
		itemCopy := item
		itemCopy.Name = itemName
		group.Items = append(group.Items, itemCopy)
	}
	if len(groupOrder) == 0 && order.Pet != nil && order.Pet.Name != "" {
		return []model.OrderPetGroup{{
			PetID:   order.PetID,
			PetName: order.Pet.Name,
			Items:   append([]model.OrderItem(nil), order.Items...),
		}}
	}
	if len(groupOrder) == 0 && singleFeedingPetName != "" {
		return []model.OrderPetGroup{{
			PetID:   uintPtr(singleFeedingPetID),
			PetName: singleFeedingPetName,
			Items:   append([]model.OrderItem(nil), order.Items...),
		}}
	}
	groups := make([]model.OrderPetGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		groups = append(groups, *groupMap[key])
	}
	return groups
}

func appointmentPetIDsByName(order *model.Order) map[string]uint {
	result := make(map[string]uint)
	if order == nil || order.Appointment == nil {
		return result
	}
	for _, apptPet := range order.Appointment.Pets {
		if apptPet.PetID == 0 || apptPet.Pet == nil || apptPet.Pet.Name == "" {
			continue
		}
		if _, exists := result[apptPet.Pet.Name]; exists {
			continue
		}
		result[apptPet.Pet.Name] = apptPet.PetID
	}
	return result
}

func feedingPetIDsByName(order *model.Order) map[string]uint {
	result := make(map[string]uint)
	if order == nil || order.FeedingPlan == nil {
		return result
	}
	for _, planPet := range order.FeedingPlan.Pets {
		name := ""
		if planPet.Pet != nil && strings.TrimSpace(planPet.Pet.Name) != "" {
			name = strings.TrimSpace(planPet.Pet.Name)
		} else {
			name = strings.TrimSpace(planPet.PetNameSnapshot)
		}
		if name == "" || planPet.PetID == 0 {
			continue
		}
		if _, exists := result[name]; exists {
			continue
		}
		result[name] = planPet.PetID
	}
	return result
}

func singleFeedingPet(order *model.Order) (string, uint) {
	if order == nil || order.FeedingPlan == nil || len(order.FeedingPlan.Pets) != 1 {
		return "", 0
	}
	planPet := order.FeedingPlan.Pets[0]
	if planPet.Pet != nil && strings.TrimSpace(planPet.Pet.Name) != "" {
		return strings.TrimSpace(planPet.Pet.Name), planPet.PetID
	}
	return strings.TrimSpace(planPet.PetNameSnapshot), planPet.PetID
}

func buildOrderPetSummary(order *model.Order) string {
	groups := buildOrderPetGroups(order)
	names := make([]string, 0, len(groups))
	seen := make(map[string]struct{})
	for _, group := range groups {
		if group.PetName == "" || group.PetName == "未分组" || group.PetName == "零售商品" {
			continue
		}
		if _, ok := seen[group.PetName]; ok {
			continue
		}
		seen[group.PetName] = struct{}{}
		names = append(names, group.PetName)
	}
	if len(names) == 0 && order.Pet != nil && order.Pet.Name != "" {
		names = append(names, order.Pet.Name)
	}
	if len(names) == 0 && order.FeedingPlan != nil {
		seenFeeding := make(map[string]struct{})
		for _, planPet := range order.FeedingPlan.Pets {
			name := strings.TrimSpace(planPet.PetNameSnapshot)
			if planPet.Pet != nil && strings.TrimSpace(planPet.Pet.Name) != "" {
				name = strings.TrimSpace(planPet.Pet.Name)
			}
			if name == "" {
				continue
			}
			if _, ok := seenFeeding[name]; ok {
				continue
			}
			seenFeeding[name] = struct{}{}
			names = append(names, name)
		}
	}
	return strings.Join(names, " / ")
}

func decorateOrderTotals(order *model.Order) {
	if order == nil {
		return
	}

	var serviceTotal, productTotal, addonTotal float64
	for _, item := range order.Items {
		switch item.ItemType {
		case 1:
			serviceTotal += item.Amount
		case 2:
			productTotal += item.Amount
		case 3:
			addonTotal += item.Amount
		}
	}

	if serviceTotal > 0 || order.ServiceTotal == 0 {
		order.ServiceTotal = serviceTotal
	}
	if productTotal > 0 || order.ProductTotal == 0 {
		order.ProductTotal = productTotal
	}
	if addonTotal > 0 || order.AddonTotal == 0 {
		order.AddonTotal = addonTotal
	}

	if order.TotalAmount == 0 {
		order.TotalAmount = order.ServiceTotal + order.ProductTotal + order.AddonTotal
	}

	if order.ServiceDiscountAmount == 0 && order.ProductDiscountAmount == 0 && order.DiscountAmount > 0 {
		switch {
		case order.ServiceTotal > 0 && order.ProductTotal == 0:
			order.ServiceDiscountAmount = order.DiscountAmount
		case order.ProductTotal > 0 && order.ServiceTotal == 0:
			order.ProductDiscountAmount = order.DiscountAmount
		}
	}
}

func buildOrderKind(order *model.Order) string {
	if order == nil {
		return "service"
	}
	if order.FeedingPlanID != nil && *order.FeedingPlanID > 0 {
		return "feeding"
	}
	hasService := order.ServiceTotal > 0
	hasProduct := order.ProductTotal > 0
	switch {
	case hasService && hasProduct:
		return "mixed"
	case hasProduct:
		return "product"
	default:
		return "service"
	}
}

func splitOrderItemPetName(name string) (string, string) {
	parts := strings.SplitN(name, " · ", 2)
	if len(parts) < 2 {
		return "", name
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func (s *OrderService) syncAppointmentSettlementPtr(appointmentID *uint) error {
	if appointmentID == nil || *appointmentID == 0 {
		return nil
	}
	return s.syncAppointmentSettlement(*appointmentID)
}

func (s *OrderService) syncAppointmentSettlement(appointmentID uint) error {
	appt, err := s.apptRepo.FindByID(appointmentID)
	if err != nil {
		return err
	}

	orders, err := s.orderRepo.FindByAppointment(appointmentID)
	if err != nil {
		return err
	}
	activeOrders := filterActiveAppointmentOrders(orders)

	var paidAmount float64
	for _, order := range orders {
		if order.Status == 1 && order.PayStatus == 1 {
			paidAmount += order.PayAmount
		}
	}

	updates := map[string]interface{}{
		"paid_amount": paidAmount,
	}
	if len(activeOrders) > 0 {
		updates["status"] = 7
	} else if appt.Status == 7 {
		updates["status"] = 3
	}

	return database.DB.Model(&model.Appointment{}).
		Where("id = ?", appointmentID).
		Updates(updates).Error
}
