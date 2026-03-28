package service

import (
	"errors"
	"math"
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
		TotalAmount:   appt.TotalAmount,
		PayAmount:     appt.TotalAmount,
	}

	tx := database.DB.Begin()
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var items []model.OrderItem
	if len(appt.Pets) > 0 {
		for _, apptPet := range appt.Pets {
			petName := "宠物"
			if apptPet.Pet != nil && apptPet.Pet.Name != "" {
				petName = apptPet.Pet.Name
			}
			for _, svc := range apptPet.Services {
				items = append(items, model.OrderItem{
					OrderID:   order.ID,
					ItemType:  1, // service
					ItemID:    svc.ServiceID,
					Name:      petName + " · " + svc.ServiceName,
					Quantity:  1,
					UnitPrice: svc.Price,
					Amount:    svc.Price,
				})
			}
		}
	} else {
		for _, svc := range appt.Services {
			items = append(items, model.OrderItem{
				OrderID:   order.ID,
				ItemType:  1, // service
				ItemID:    svc.ServiceID,
				Name:      svc.ServiceName,
				Quantity:  1,
				UnitPrice: svc.Price,
				Amount:    svc.Price,
			})
		}
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	return order, tx.Commit().Error
}

type ServiceOverride struct {
	Price    float64
	Duration int
}

func (s *OrderService) CreateSplitFromAppointment(appointmentID uint, overrides ...map[uint]map[uint]ServiceOverride) ([]model.Order, error) {
	// overrideMap: petId -> serviceId -> {Price, Duration}
	var overrideMap map[uint]map[uint]ServiceOverride
	if len(overrides) > 0 && overrides[0] != nil {
		overrideMap = overrides[0]
	}
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

	if len(appt.Pets) == 0 {
		order, err := s.CreateFromAppointment(appointmentID)
		if err != nil {
			return nil, err
		}
		return []model.Order{*order}, nil
	}

	orders := make([]model.Order, 0, len(appt.Pets))
	tx := database.DB.Begin()

	for _, apptPet := range appt.Pets {
		custID := appt.CustomerID
		petID := apptPet.PetID
		order := model.Order{
			OrderNo:       utils.GenerateOrderNo(),
			ShopID:        appt.ShopID,
			CustomerID:    &custID,
			PetID:         &petID,
			AppointmentID: &appt.ID,
			StaffID:       appt.StaffID,
			TotalAmount:   apptPet.TotalAmount,
			PayAmount:     apptPet.TotalAmount,
			Remark:        appt.Notes,
		}

		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		petName := "宠物"
		if apptPet.Pet != nil && apptPet.Pet.Name != "" {
			petName = apptPet.Pet.Name
		}

		// 应用价格覆盖
		petSvcOverrides := make(map[uint]ServiceOverride)
		if overrideMap != nil {
			if m, ok := overrideMap[apptPet.PetID]; ok {
				petSvcOverrides = m
			}
		}

		items := make([]model.OrderItem, 0, len(apptPet.Services))
		var orderTotal float64
		for _, svc := range apptPet.Services {
			price := svc.Price
			if ov, ok := petSvcOverrides[svc.ServiceID]; ok {
				price = ov.Price
			}
			items = append(items, model.OrderItem{
				OrderID:   order.ID,
				ItemType:  1,
				ItemID:    svc.ServiceID,
				Name:      petName + " · " + svc.ServiceName,
				Quantity:  1,
				UnitPrice: price,
				Amount:    price,
			})
			orderTotal += price
		}
		// 如果有覆盖，更新订单总额
		if len(petSvcOverrides) > 0 {
			order.TotalAmount = orderTotal
			order.PayAmount = orderTotal
			tx.Save(&order)
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		orders = append(orders, order)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// CreateDirect creates a standalone order (walk-in / direct billing)
func (s *OrderService) CreateDirect(order *model.Order, items []model.OrderItem) error {
	order.OrderNo = utils.GenerateOrderNo()

	var total float64
	for i := range items {
		items[i].Amount = items[i].UnitPrice * float64(items[i].Quantity)
		total += items[i].Amount
	}
	order.TotalAmount = total
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

func (s *OrderService) GetByID(id uint) (*model.Order, error) {
	return s.orderRepo.FindByID(id)
}

func (s *OrderService) ListPaged(shopID uint, f repository.OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.FindByShopPaged(shopID, f, page, pageSize)
}

func (s *OrderService) Search(shopID uint, keyword string, f repository.OrderFilter, page, pageSize int) ([]model.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.orderRepo.Search(shopID, keyword, f, page, pageSize)
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
	return s.orderRepo.Update(order)
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
	return s.orderRepo.Update(order)
}
