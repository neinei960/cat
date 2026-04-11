package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"github.com/neinei960/cat/server/pkg/utils"
	"gorm.io/gorm"
)

type BoardingPriceLine struct {
	Type      string  `json:"type"`
	Label     string  `json:"label"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Amount    float64 `json:"amount"`
}

type BoardingPricePreview struct {
	CheckInAt              string                         `json:"check_in_at"`
	CheckOutAt             string                         `json:"check_out_at"`
	Nights                 int                            `json:"nights"`
	PetCount               int                            `json:"pet_count"`
	RegularNights          int                            `json:"regular_nights"`
	HolidayNights          int                            `json:"holiday_nights"`
	BaseAmount             float64                        `json:"base_amount"`
	ExtraPetAmount         float64                        `json:"extra_pet_amount"`
	HolidaySurchargeAmount float64                        `json:"holiday_surcharge_amount"`
	DiscountAmount         float64                        `json:"discount_amount"`
	PayAmount              float64                        `json:"pay_amount"`
	Policies               []model.BoardingDiscountPolicy `json:"policies"`
	Lines                  []BoardingPriceLine            `json:"lines"`
}

type BoardingPreviewInput struct {
	CustomerID uint
	PetIDs     []uint
	PetCount   int
	CabinetID  uint
	CheckInAt  string
	CheckOutAt string
	PolicyIDs  []uint
}

type BoardingCreateInput struct {
	CustomerID   uint
	PetIDs       []uint
	CabinetID    uint
	CheckInAt    string
	CheckOutAt   string
	PolicyIDs    []uint
	HasDeworming *bool
	Remark       string
	OperatorID   uint
}

type BoardingCheckInInput struct {
	DiscountAmount float64
}

type BoardingDashboardGroup struct {
	CabinetID      uint                  `json:"cabinet_id"`
	CabinetType    string                `json:"cabinet_type"`
	RoomCount      int                   `json:"room_count"`
	Capacity       int                   `json:"capacity"`
	BasePrice      float64               `json:"base_price"`
	ExtraPetPrice  float64               `json:"extra_pet_price"`
	Status         string                `json:"status"`
	Remark         string                `json:"remark"`
	OccupiedRooms  int                   `json:"occupied_rooms"`
	ReservedRooms  int                   `json:"reserved_rooms"`
	RemainingRooms int                   `json:"remaining_rooms"`
	Orders         []model.BoardingOrder `json:"orders,omitempty"`
}

type stayRule struct {
	Stay int `json:"stay"`
	Free int `json:"free"`
}

type surchargeRule struct {
	Surcharge float64 `json:"surcharge"`
}

type BoardingService struct {
	repo         *repository.BoardingRepository
	orderRepo    *repository.OrderRepository
	customerRepo *repository.CustomerRepository
	petRepo      *repository.PetRepository
}

func NewBoardingService(repo *repository.BoardingRepository, orderRepo *repository.OrderRepository, customerRepo *repository.CustomerRepository, petRepo *repository.PetRepository) *BoardingService {
	return &BoardingService{repo: repo, orderRepo: orderRepo, customerRepo: customerRepo, petRepo: petRepo}
}

func applyMemberDiscountToBoardingPreview(customerID uint, preview *BoardingPricePreview) *BoardingPricePreview {
	if preview == nil || customerID == 0 {
		return preview
	}
	customerRef := customerID
	serviceDiscountRate, _ := resolveMemberDiscountRates(&customerRef)
	if serviceDiscountRate <= 0 || serviceDiscountRate >= 1 {
		return preview
	}

	discountedPay := roundMoney(preview.PayAmount * serviceDiscountRate)
	memberDiscountAmount := roundMoney(preview.PayAmount - discountedPay)
	if memberDiscountAmount <= 0 {
		return preview
	}

	adjusted := *preview
	adjusted.DiscountAmount = roundMoney(preview.DiscountAmount + memberDiscountAmount)
	adjusted.PayAmount = discountedPay
	adjusted.Lines = append(append([]BoardingPriceLine{}, preview.Lines...), BoardingPriceLine{
		Type:      "member_discount",
		Label:     "会员折扣",
		Quantity:  1,
		UnitPrice: -memberDiscountAmount,
		Amount:    -memberDiscountAmount,
	})
	return &adjusted
}

func applyManualDiscountToBoardingPreview(preview *BoardingPricePreview, amount float64) (*BoardingPricePreview, error) {
	if preview == nil {
		return nil, nil
	}
	amount = roundMoney(amount)
	if amount < 0 {
		return nil, errors.New("优惠金额不能小于 0")
	}
	if amount == 0 {
		return preview, nil
	}
	if amount > preview.PayAmount {
		return nil, errors.New("优惠金额不能大于当前应收金额")
	}

	adjusted := *preview
	adjusted.DiscountAmount = roundMoney(preview.DiscountAmount + amount)
	adjusted.PayAmount = roundMoney(preview.PayAmount - amount)
	adjusted.Lines = append(append([]BoardingPriceLine{}, preview.Lines...), BoardingPriceLine{
		Type:      "manual_discount",
		Label:     "入住优惠",
		Quantity:  1,
		UnitPrice: -amount,
		Amount:    -amount,
	})
	return &adjusted, nil
}

func (s *BoardingService) ListCabinets(shopID uint) ([]model.BoardingCabinet, error) {
	return s.repo.ListCabinets(shopID)
}

func (s *BoardingService) CreateCabinet(cabinet *model.BoardingCabinet) error {
	cabinet.CabinetType = strings.TrimSpace(cabinet.CabinetType)
	if cabinet.CabinetType == "" {
		return errors.New("请填写寄养房型")
	}
	cabinet.Code = cabinet.CabinetType
	cabinet.RoomCount = maxInt(cabinet.RoomCount, 1)
	if cabinet.Capacity < 1 {
		cabinet.Capacity = 1
	}
	if cabinet.Status == "" {
		cabinet.Status = model.BoardingCabinetStatusEnabled
	}
	if cabinet.ExtraPetPrice < 0 {
		cabinet.ExtraPetPrice = 0
	}
	return database.DB.Create(cabinet).Error
}

func (s *BoardingService) UpdateCabinet(shopID uint, cabinet *model.BoardingCabinet) error {
	existing, err := s.repo.FindCabinetByID(shopID, cabinet.ID)
	if err != nil {
		return err
	}
	existing.CabinetType = strings.TrimSpace(cabinet.CabinetType)
	if existing.CabinetType == "" {
		return errors.New("请填写寄养房型")
	}
	existing.Code = existing.CabinetType
	existing.RoomCount = maxInt(cabinet.RoomCount, 1)
	existing.Capacity = maxInt(cabinet.Capacity, 1)
	existing.BasePrice = cabinet.BasePrice
	existing.ExtraPetPrice = clampMinFloat(cabinet.ExtraPetPrice, 0)
	existing.Status = cabinet.Status
	existing.Remark = cabinet.Remark
	return database.DB.Save(existing).Error
}

func (s *BoardingService) ListHolidays(shopID uint) ([]model.BoardingHoliday, error) {
	return s.repo.ListHolidays(shopID)
}

func (s *BoardingService) CreateHoliday(holiday *model.BoardingHoliday) error {
	dateText, err := normalizeDate(holiday.HolidayDate)
	if err != nil {
		return err
	}
	holiday.HolidayDate = dateText
	if holiday.Name == "" {
		holiday.Name = "节假日"
	}
	return database.DB.Create(holiday).Error
}

func (s *BoardingService) DeleteHoliday(shopID, id uint) error {
	return database.DB.Where("shop_id = ?", shopID).Delete(&model.BoardingHoliday{}, id).Error
}

func (s *BoardingService) ListPolicies(shopID uint) ([]model.BoardingDiscountPolicy, error) {
	return s.repo.ListPolicies(shopID)
}

func (s *BoardingService) CreatePolicy(policy *model.BoardingDiscountPolicy) error {
	if err := validateBoardingPolicy(policy); err != nil {
		return err
	}
	return database.DB.Create(policy).Error
}

func (s *BoardingService) UpdatePolicy(shopID uint, policy *model.BoardingDiscountPolicy) error {
	var existing model.BoardingDiscountPolicy
	if err := database.DB.Where("shop_id = ?", shopID).First(&existing, policy.ID).Error; err != nil {
		return err
	}
	existing.Name = policy.Name
	existing.PolicyType = policy.PolicyType
	existing.RuleJSON = policy.RuleJSON
	existing.ValidFrom = policy.ValidFrom
	existing.ValidTo = policy.ValidTo
	existing.Priority = policy.Priority
	existing.Stackable = policy.Stackable
	existing.Status = policy.Status
	existing.Remark = policy.Remark
	if err := validateBoardingPolicy(&existing); err != nil {
		return err
	}
	return database.DB.Save(&existing).Error
}

func (s *BoardingService) GetAvailableCabinets(shopID uint, checkInAt, checkOutAt string, petCount int, excludeOrderID uint) ([]model.BoardingCabinet, error) {
	startDate, endDate, nights, err := normalizeStayRange(checkInAt, checkOutAt)
	if err != nil {
		return nil, err
	}
	_ = nights
	if petCount < 1 {
		petCount = 1
	}
	allCabinets, err := s.repo.ListCabinets(shopID)
	if err != nil {
		return nil, err
	}
	activeCounts, _, err := s.listOverlappingCabinetUsage(shopID, startDate, endDate, excludeOrderID)
	if err != nil {
		return nil, err
	}
	available := make([]model.BoardingCabinet, 0, len(allCabinets))
	for _, cabinet := range allCabinets {
		if cabinet.Status != model.BoardingCabinetStatusEnabled {
			continue
		}
		if cabinet.Capacity < petCount {
			continue
		}
		cabinet.RoomCount = maxInt(cabinet.RoomCount, 1)
		cabinet.OccupiedRooms = activeCounts[cabinet.ID]
		cabinet.RemainingRooms = maxInt(cabinet.RoomCount-cabinet.OccupiedRooms, 0)
		if cabinet.RemainingRooms < 1 {
			continue
		}
		available = append(available, cabinet)
	}
	return available, nil
}

func (s *BoardingService) listOverlappingCabinetUsage(shopID uint, startDate, endDate string, excludeOrderID uint) (map[uint]int, []model.BoardingOrder, error) {
	var activeOrders []model.BoardingOrder
	if err := database.DB.Preload("Customer").
		Preload("Pets.Pet").
		Where(
			"shop_id = ? AND status IN ? AND cabinet_id IS NOT NULL AND check_in_at < ? AND check_out_at > ? AND id <> ?",
			shopID,
			[]string{model.BoardingOrderStatusPendingCheckin, model.BoardingOrderStatusCheckedIn},
			endDate,
			startDate,
			excludeOrderID,
		).
		Order("check_in_at ASC, id ASC").
		Find(&activeOrders).Error; err != nil {
		return nil, nil, err
	}
	counts := make(map[uint]int, len(activeOrders))
	for _, order := range activeOrders {
		counts[order.CabinetID]++
	}
	return counts, activeOrders, nil
}

func (s *BoardingService) Preview(shopID uint, input BoardingPreviewInput) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	cabinet, err := s.repo.FindCabinetByID(shopID, input.CabinetID)
	if err != nil {
		return nil, nil, nil, errors.New("寄养房型不存在")
	}
	if cabinet.Status != model.BoardingCabinetStatusEnabled {
		return nil, nil, nil, errors.New("该寄养房型当前不可用")
	}
	cabinet.RoomCount = maxInt(cabinet.RoomCount, 1)
	petIDs, petCount, err := s.resolvePetSelection(shopID, input.CustomerID, input.PetIDs, input.PetCount)
	if err != nil {
		return nil, nil, nil, err
	}
	if cabinet.Capacity < petCount {
		return nil, nil, nil, errors.New("所选猫咪数量超出该房型单间容量")
	}
	availableCabinets, err := s.GetAvailableCabinets(shopID, input.CheckInAt, input.CheckOutAt, petCount, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	allowed := false
	for _, item := range availableCabinets {
		if item.ID == cabinet.ID {
			cabinet.OccupiedRooms = item.OccupiedRooms
			cabinet.RemainingRooms = item.RemainingRooms
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, nil, nil, errors.New("所选日期内该房型已经住满了")
	}

	selectedPolicies, err := s.resolvePolicies(shopID, input.PolicyIDs, input.CheckInAt, input.CheckOutAt)
	if err != nil {
		return nil, nil, nil, err
	}
	preview, err := s.computePreview(shopID, cabinet, input.CheckInAt, input.CheckOutAt, petCount, selectedPolicies)
	if err != nil {
		return nil, nil, nil, err
	}
	return preview, cabinet, petIDs, nil
}

func (s *BoardingService) CreateOrder(shopID uint, input BoardingCreateInput) (*model.BoardingOrder, error) {
	if input.CustomerID == 0 {
		return nil, errors.New("请选择客户")
	}
	preview, cabinet, petIDs, err := s.Preview(shopID, BoardingPreviewInput{
		CustomerID: input.CustomerID,
		PetIDs:     input.PetIDs,
		CabinetID:  input.CabinetID,
		CheckInAt:  input.CheckInAt,
		CheckOutAt: input.CheckOutAt,
		PolicyIDs:  input.PolicyIDs,
	})
	if err != nil {
		return nil, err
	}
	customer, err := s.customerRepo.FindByID(input.CustomerID)
	if err != nil {
		return nil, errors.New("客户不存在")
	}
	pets, err := s.loadPets(petIDs)
	if err != nil {
		return nil, err
	}
	adjustedPreview := applyMemberDiscountToBoardingPreview(input.CustomerID, preview)

	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(adjustedPreview)

	boardingOrder := &model.BoardingOrder{
		ShopID:                 shopID,
		CustomerID:             input.CustomerID,
		StaffID:                input.OperatorID,
		CabinetID:              cabinet.ID,
		CheckInAt:              adjustedPreview.CheckInAt,
		CheckOutAt:             adjustedPreview.CheckOutAt,
		Nights:                 adjustedPreview.Nights,
		BaseAmount:             adjustedPreview.BaseAmount,
		HolidaySurchargeAmount: adjustedPreview.HolidaySurchargeAmount,
		DiscountAmount:         adjustedPreview.DiscountAmount,
		PayAmount:              adjustedPreview.PayAmount,
		Status:                 model.BoardingOrderStatusPendingCheckin,
		HasDeworming:           input.HasDeworming,
		Remark:                 input.Remark,
		PolicySnapshotJSON:     string(policySnapshot),
		PriceSnapshotJSON:      string(priceSnapshot),
	}

	var createdID uint
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(boardingOrder).Error; err != nil {
			return err
		}
		var orderPetID *uint
		if len(petIDs) == 1 {
			orderPetID = &petIDs[0]
		}
		orderTotal := roundMoney(adjustedPreview.BaseAmount + adjustedPreview.HolidaySurchargeAmount)
		order := &model.Order{
			OrderNo:               utils.GenerateOrderNo(),
			ShopID:                shopID,
			CustomerID:            &input.CustomerID,
			PetID:                 orderPetID,
			StaffID:               uintPtr(input.OperatorID),
			TotalAmount:           orderTotal,
			ServiceTotal:          orderTotal,
			DiscountAmount:        adjustedPreview.DiscountAmount,
			ServiceDiscountAmount: adjustedPreview.DiscountAmount,
			DiscountRate:          calculateOrderDiscountRate(orderTotal, adjustedPreview.PayAmount),
			PayAmount:             adjustedPreview.PayAmount,
			PayStatus:             0,
			Status:                0,
			Remark:                fmt.Sprintf("寄养订单 · %s · %s", cabinet.CabinetType, customer.Nickname),
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		boardingOrder.OrderID = &order.ID
		if err := tx.Save(boardingOrder).Error; err != nil {
			return err
		}
		items := buildBoardingOrderItems(order.ID, cabinet, adjustedPreview)
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		orderPets := make([]model.BoardingOrderPet, 0, len(pets))
		for _, pet := range pets {
			orderPets = append(orderPets, model.BoardingOrderPet{
				BoardingOrderID: boardingOrder.ID,
				PetID:           pet.ID,
				PetNameSnapshot: pet.Name,
			})
		}
		if len(orderPets) > 0 {
			if err := tx.Create(&orderPets).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: boardingOrder.ID,
			OperatorID:      input.OperatorID,
			Action:          "create",
			Content:         fmt.Sprintf("创建寄养单，房型 %s，入住 %s，离店 %s", cabinet.CabinetType, preview.CheckInAt, preview.CheckOutAt),
		}).Error; err != nil {
			return err
		}
		createdID = boardingOrder.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindBoardingOrderByID(shopID, createdID)
}

func (s *BoardingService) ListOrders(shopID uint, status string, page, pageSize int) ([]model.BoardingOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.ListBoardingOrders(shopID, status, page, pageSize)
}

func (s *BoardingService) GetOrder(shopID, id uint) (*model.BoardingOrder, error) {
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) Dashboard(shopID uint) ([]BoardingDashboardGroup, error) {
	cabinets, err := s.repo.ListCabinets(shopID)
	if err != nil {
		return nil, err
	}
	activeOrders, err := s.repo.ListActiveOrders(shopID)
	if err != nil {
		return nil, err
	}
	ordersByCabinet := make(map[uint][]model.BoardingOrder)
	occupiedCount := make(map[uint]int)
	reservedCount := make(map[uint]int)
	today := time.Now().Format("2006-01-02")
	for _, order := range activeOrders {
		ordersByCabinet[order.CabinetID] = append(ordersByCabinet[order.CabinetID], order)
		if order.Status == model.BoardingOrderStatusCheckedIn {
			occupiedCount[order.CabinetID]++
		} else if order.CheckInAt >= today {
			reservedCount[order.CabinetID]++
		} else {
			occupiedCount[order.CabinetID]++
		}
	}
	groups := make([]BoardingDashboardGroup, 0, len(cabinets))
	for _, cabinet := range cabinets {
		cabinet.RoomCount = maxInt(cabinet.RoomCount, 1)
		activeCount := occupiedCount[cabinet.ID] + reservedCount[cabinet.ID]
		remaining := maxInt(cabinet.RoomCount-activeCount, 0)
		if cabinet.Status != model.BoardingCabinetStatusEnabled {
			remaining = 0
		}
		groups = append(groups, BoardingDashboardGroup{
			CabinetID:      cabinet.ID,
			CabinetType:    cabinet.CabinetType,
			RoomCount:      cabinet.RoomCount,
			Capacity:       cabinet.Capacity,
			BasePrice:      cabinet.BasePrice,
			ExtraPetPrice:  cabinet.ExtraPetPrice,
			Status:         cabinet.Status,
			Remark:         cabinet.Remark,
			OccupiedRooms:  occupiedCount[cabinet.ID],
			ReservedRooms:  reservedCount[cabinet.ID],
			RemainingRooms: remaining,
			Orders:         ordersByCabinet[cabinet.ID],
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].CabinetType < groups[j].CabinetType
	})
	return groups, nil
}

func (s *BoardingService) CheckIn(shopID, id, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if order.Status != model.BoardingOrderStatusPendingCheckin {
		return nil, errors.New("当前状态不可办理入住")
	}

	selectedPolicies := parsePolicySnapshot(order.PolicySnapshotJSON)
	preview, cabinet, petIDs, err := s.computePreviewFromExisting(shopID, order, order.CheckOutAt, selectedPolicies)
	if err != nil {
		return nil, err
	}
	adjustedPreview := applyMemberDiscountToBoardingPreview(order.CustomerID, preview)
	adjustedPreview, err = applyManualDiscountToBoardingPreview(adjustedPreview, input.DiscountAmount)
	if err != nil {
		return nil, err
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if order.OrderID != nil && *order.OrderID > 0 {
			var payOrder model.Order
			if err := tx.First(&payOrder, *order.OrderID).Error; err == nil && payOrder.PayStatus == 1 && input.DiscountAmount > 0 {
				return errors.New("已支付订单不可在入住时追加优惠")
			}
		}
		order.Status = model.BoardingOrderStatusCheckedIn
		order.Nights = adjustedPreview.Nights
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = roundMoney(input.DiscountAmount)
		order.PayAmount = adjustedPreview.PayAmount
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if err := s.syncOrder(tx, order, cabinet, adjustedPreview, petIDs, false); err != nil {
			if strings.Contains(err.Error(), "已支付订单不可修改") && input.DiscountAmount == 0 {
				return nil
			}
			return err
		}
		content := "办理入住"
		if input.DiscountAmount > 0 {
			content = fmt.Sprintf("办理入住，享受优惠 ¥%.2f", roundMoney(input.DiscountAmount))
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "check_in",
			Content:         content,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) CheckOut(shopID, id, operatorID uint, actualDate string) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if order.Status != model.BoardingOrderStatusCheckedIn {
		return nil, errors.New("当前状态不可办理离店")
	}
	actualDate, err = normalizeDate(actualDate)
	if err != nil {
		return nil, err
	}
	selectedPolicies := parsePolicySnapshot(order.PolicySnapshotJSON)
	preview, cabinet, petIDs, err := s.computePreviewFromExisting(shopID, order, actualDate, selectedPolicies)
	if err != nil {
		return nil, err
	}
	adjustedPreview := applyMemberDiscountToBoardingPreview(order.CustomerID, preview)
	appliedManualDiscount := roundMoney(minFloat(order.ManualDiscountAmount, adjustedPreview.PayAmount))
	adjustedPreview, err = applyManualDiscountToBoardingPreview(adjustedPreview, appliedManualDiscount)
	if err != nil {
		return nil, err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.ActualCheckOutAt = actualDate
		order.CheckOutAt = actualDate
		order.Nights = adjustedPreview.Nights
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = appliedManualDiscount
		order.PayAmount = adjustedPreview.PayAmount
		order.Status = model.BoardingOrderStatusCheckedOut
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return s.syncOrder(tx, order, cabinet, adjustedPreview, petIDs, true)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "check_out",
		Content:         fmt.Sprintf("办理离店，实际离店日期 %s", actualDate),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) Extend(shopID, id, operatorID uint, newCheckOutAt string) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureEditableOrder(order); err != nil {
		return nil, err
	}
	selectedPolicies := parsePolicySnapshot(order.PolicySnapshotJSON)
	preview, cabinet, petIDs, err := s.computePreviewFromExisting(shopID, order, newCheckOutAt, selectedPolicies)
	if err != nil {
		return nil, err
	}
	adjustedPreview := applyMemberDiscountToBoardingPreview(order.CustomerID, preview)
	appliedManualDiscount := roundMoney(minFloat(order.ManualDiscountAmount, adjustedPreview.PayAmount))
	adjustedPreview, err = applyManualDiscountToBoardingPreview(adjustedPreview, appliedManualDiscount)
	if err != nil {
		return nil, err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.CheckOutAt = adjustedPreview.CheckOutAt
		order.Nights = adjustedPreview.Nights
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = appliedManualDiscount
		order.PayAmount = adjustedPreview.PayAmount
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return s.syncOrder(tx, order, cabinet, adjustedPreview, petIDs, false)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "extend",
		Content:         fmt.Sprintf("续住至 %s", preview.CheckOutAt),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) ChangeCabinet(shopID, id, operatorID, cabinetID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if err := s.ensureEditableOrder(order); err != nil {
		return nil, err
	}
	selectedPolicies := parsePolicySnapshot(order.PolicySnapshotJSON)
	petIDs := collectBoardingPetIDs(order)
	preview, cabinet, _, err := s.Preview(shopID, BoardingPreviewInput{
		CustomerID: order.CustomerID,
		PetIDs:     petIDs,
		CabinetID:  cabinetID,
		CheckInAt:  order.CheckInAt,
		CheckOutAt: order.CheckOutAt,
		PolicyIDs:  policyIDsFromPolicies(selectedPolicies),
	})
	if err != nil {
		return nil, err
	}
	adjustedPreview := applyMemberDiscountToBoardingPreview(order.CustomerID, preview)
	appliedManualDiscount := roundMoney(minFloat(order.ManualDiscountAmount, adjustedPreview.PayAmount))
	adjustedPreview, err = applyManualDiscountToBoardingPreview(adjustedPreview, appliedManualDiscount)
	if err != nil {
		return nil, err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.CabinetID = cabinet.ID
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = appliedManualDiscount
		order.PayAmount = adjustedPreview.PayAmount
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return s.syncOrder(tx, order, cabinet, adjustedPreview, petIDs, false)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "change_cabinet",
		Content:         fmt.Sprintf("更换寄养房型为 %s", cabinet.CabinetType),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) Cancel(shopID, id, operatorID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if order.Status != model.BoardingOrderStatusPendingCheckin {
		return nil, errors.New("当前状态不可取消")
	}
	if order.Order != nil && order.Order.PayStatus == 1 {
		return nil, errors.New("已支付订单不可取消")
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.Status = model.BoardingOrderStatusCancelled
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if order.OrderID != nil && *order.OrderID > 0 {
			var payOrder model.Order
			if err := tx.First(&payOrder, *order.OrderID).Error; err == nil {
				payOrder.Status = 2
				payOrder.PayStatus = 0
				if err := tx.Save(&payOrder).Error; err != nil {
					return err
				}
			}
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "cancel",
			Content:         "取消寄养订单",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) computePreviewFromExisting(shopID uint, order *model.BoardingOrder, targetCheckOutAt string, policies []model.BoardingDiscountPolicy) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	petIDs := collectBoardingPetIDs(order)
	return s.Preview(shopID, BoardingPreviewInput{
		CustomerID: order.CustomerID,
		PetIDs:     petIDs,
		CabinetID:  order.CabinetID,
		CheckInAt:  order.CheckInAt,
		CheckOutAt: targetCheckOutAt,
		PolicyIDs:  policyIDsFromPolicies(policies),
	})
}

func (s *BoardingService) syncOrder(tx *gorm.DB, boardingOrder *model.BoardingOrder, cabinet *model.BoardingCabinet, preview *BoardingPricePreview, petIDs []uint, allowPaidCheckOut bool) error {
	if boardingOrder.OrderID == nil || *boardingOrder.OrderID == 0 {
		return nil
	}
	var payOrder model.Order
	if err := tx.First(&payOrder, *boardingOrder.OrderID).Error; err != nil {
		return err
	}
	if payOrder.PayStatus == 1 && !allowPaidCheckOut {
		return errors.New("已支付订单不可修改")
	}
	if payOrder.PayStatus == 1 && allowPaidCheckOut {
		return nil
	}
	payOrder.TotalAmount = roundMoney(preview.BaseAmount + preview.HolidaySurchargeAmount)
	payOrder.ServiceTotal = payOrder.TotalAmount
	payOrder.ProductTotal = 0
	payOrder.AddonTotal = 0
	payOrder.ServiceDiscountAmount = preview.DiscountAmount
	payOrder.ProductDiscountAmount = 0
	payOrder.DiscountAmount = preview.DiscountAmount
	payOrder.DiscountRate = calculateOrderDiscountRate(payOrder.TotalAmount, preview.PayAmount)
	payOrder.PayAmount = preview.PayAmount
	if err := tx.Save(&payOrder).Error; err != nil {
		return err
	}
	if err := tx.Where("order_id = ?", payOrder.ID).Delete(&model.OrderItem{}).Error; err != nil {
		return err
	}
	items := buildBoardingOrderItems(payOrder.ID, cabinet, preview)
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *BoardingService) ensureEditableOrder(order *model.BoardingOrder) error {
	if order.Status == model.BoardingOrderStatusCancelled || order.Status == model.BoardingOrderStatusCheckedOut {
		return errors.New("当前状态不可修改")
	}
	if order.Order != nil && order.Order.PayStatus == 1 {
		return errors.New("已支付订单不可修改")
	}
	return nil
}

func (s *BoardingService) resolvePetSelection(shopID, customerID uint, petIDs []uint, petCount int) ([]uint, int, error) {
	if len(petIDs) > 0 {
		pets, err := s.loadPets(petIDs)
		if err != nil {
			return nil, 0, err
		}
		normalizedIDs := make([]uint, 0, len(pets))
		for _, pet := range pets {
			if pet.CustomerID == nil || *pet.CustomerID == 0 {
				return nil, 0, errors.New("寄养猫咪必须关联客户")
			}
			if customerID > 0 && *pet.CustomerID != customerID {
				return nil, 0, errors.New("同柜多猫必须属于同一客户")
			}
			customerID = *pet.CustomerID
			normalizedIDs = append(normalizedIDs, pet.ID)
		}
		return normalizedIDs, len(normalizedIDs), nil
	}
	if petCount < 1 {
		return nil, 0, errors.New("请至少选择一只猫咪")
	}
	return nil, petCount, nil
}

func (s *BoardingService) loadPets(petIDs []uint) ([]model.Pet, error) {
	if len(petIDs) == 0 {
		return nil, nil
	}
	pets := make([]model.Pet, 0, len(petIDs))
	seen := map[uint]struct{}{}
	for _, petID := range petIDs {
		if _, ok := seen[petID]; ok {
			continue
		}
		pet, err := s.petRepo.FindByID(petID)
		if err != nil {
			return nil, errors.New("猫咪不存在")
		}
		pets = append(pets, *pet)
		seen[petID] = struct{}{}
	}
	return pets, nil
}

func (s *BoardingService) resolvePolicies(shopID uint, policyIDs []uint, checkInAt, checkOutAt string) ([]model.BoardingDiscountPolicy, error) {
	var policies []model.BoardingDiscountPolicy
	var err error
	if len(policyIDs) > 0 {
		policies, err = s.repo.FindPoliciesByIDs(shopID, policyIDs)
	} else {
		policies, err = s.repo.ListPolicies(shopID)
	}
	if err != nil {
		return nil, err
	}
	validPolicies := make([]model.BoardingDiscountPolicy, 0, len(policies))
	for _, policy := range policies {
		if policy.Status != 1 {
			continue
		}
		if !policyOverlapsStay(policy, checkInAt, checkOutAt) {
			continue
		}
		validPolicies = append(validPolicies, policy)
	}
	byType := map[string]model.BoardingDiscountPolicy{}
	for _, policy := range validPolicies {
		existing, ok := byType[policy.PolicyType]
		if !ok || policy.Priority > existing.Priority || (policy.Priority == existing.Priority && policy.ID > existing.ID) {
			byType[policy.PolicyType] = policy
		}
	}
	selected := make([]model.BoardingDiscountPolicy, 0, len(byType))
	for _, policyType := range []string{model.BoardingPolicyTypeHolidaySurcharge, model.BoardingPolicyTypeStayNFreeM} {
		if policy, ok := byType[policyType]; ok {
			selected = append(selected, policy)
		}
	}
	return selected, nil
}

func (s *BoardingService) computePreview(shopID uint, cabinet *model.BoardingCabinet, checkInAt, checkOutAt string, petCount int, policies []model.BoardingDiscountPolicy) (*BoardingPricePreview, error) {
	startDate, endDate, nights, err := normalizeStayRange(checkInAt, checkOutAt)
	if err != nil {
		return nil, err
	}
	holidays, err := s.repo.ListHolidaysInRange(shopID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	holidayMap := make(map[string]model.BoardingHoliday, len(holidays))
	for _, holiday := range holidays {
		holidayMap[holiday.HolidayDate] = holiday
	}
	regularNights := 0
	holidayNights := 0
	for cursor := startDate; cursor < endDate; {
		if _, ok := holidayMap[cursor]; ok {
			holidayNights++
		} else {
			regularNights++
		}
		cursor = addDays(cursor, 1)
	}
	baseStayAmount := roundMoney(float64(nights) * cabinet.BasePrice)
	extraPetAmount := 0.0
	if petCount > 1 && cabinet.ExtraPetPrice > 0 {
		extraPetAmount = roundMoney(float64(nights) * cabinet.ExtraPetPrice)
	}
	baseAmount := roundMoney(baseStayAmount + extraPetAmount)
	var holidaySurchargeAmount float64
	var discountAmount float64
	lines := []BoardingPriceLine{
		{Type: "base", Label: fmt.Sprintf("%s 寄养住宿", cabinet.CabinetType), Quantity: nights, UnitPrice: cabinet.BasePrice, Amount: baseStayAmount},
	}
	if extraPetAmount > 0 {
		lines = append(lines, BoardingPriceLine{
			Type:      "extra_pet",
			Label:     "第二只加价",
			Quantity:  nights,
			UnitPrice: cabinet.ExtraPetPrice,
			Amount:    extraPetAmount,
		})
	}
	for _, policy := range policies {
		switch policy.PolicyType {
		case model.BoardingPolicyTypeHolidaySurcharge:
			var rule surchargeRule
			if err := json.Unmarshal([]byte(policy.RuleJSON), &rule); err == nil && rule.Surcharge > 0 && holidayNights > 0 {
				holidaySurchargeAmount = roundMoney(float64(holidayNights) * rule.Surcharge)
				lines = append(lines, BoardingPriceLine{
					Type:      "holiday_surcharge",
					Label:     policy.Name,
					Quantity:  holidayNights,
					UnitPrice: rule.Surcharge,
					Amount:    holidaySurchargeAmount,
				})
			}
		case model.BoardingPolicyTypeStayNFreeM:
			var rule stayRule
			if err := json.Unmarshal([]byte(policy.RuleJSON), &rule); err == nil && rule.Stay > 0 && rule.Free > 0 && nights >= rule.Stay {
				freeNights := minInt(rule.Free, regularNights)
				if freeNights > 0 {
					discountAmount = roundMoney(float64(freeNights) * cabinet.BasePrice)
					lines = append(lines, BoardingPriceLine{
						Type:      "discount",
						Label:     policy.Name,
						Quantity:  freeNights,
						UnitPrice: -cabinet.BasePrice,
						Amount:    -discountAmount,
					})
				}
			}
		}
	}
	payAmount := roundMoney(baseAmount + holidaySurchargeAmount - discountAmount)
	return &BoardingPricePreview{
		CheckInAt:              startDate,
		CheckOutAt:             endDate,
		Nights:                 nights,
		PetCount:               petCount,
		RegularNights:          regularNights,
		HolidayNights:          holidayNights,
		BaseAmount:             baseAmount,
		ExtraPetAmount:         extraPetAmount,
		HolidaySurchargeAmount: holidaySurchargeAmount,
		DiscountAmount:         discountAmount,
		PayAmount:              payAmount,
		Policies:               policies,
		Lines:                  lines,
	}, nil
}

func validateBoardingPolicy(policy *model.BoardingDiscountPolicy) error {
	policy.Name = strings.TrimSpace(policy.Name)
	if policy.Name == "" {
		return errors.New("请填写优惠名称")
	}
	if policy.Status == 0 {
		policy.Status = 1
	}
	if policy.PolicyType != model.BoardingPolicyTypeStayNFreeM && policy.PolicyType != model.BoardingPolicyTypeHolidaySurcharge {
		return errors.New("不支持的优惠类型")
	}
	if policy.ValidFrom != "" {
		dateText, err := normalizeDate(policy.ValidFrom)
		if err != nil {
			return err
		}
		policy.ValidFrom = dateText
	}
	if policy.ValidTo != "" {
		dateText, err := normalizeDate(policy.ValidTo)
		if err != nil {
			return err
		}
		policy.ValidTo = dateText
	}
	switch policy.PolicyType {
	case model.BoardingPolicyTypeStayNFreeM:
		var rule stayRule
		if err := json.Unmarshal([]byte(policy.RuleJSON), &rule); err != nil || rule.Stay < 1 || rule.Free < 1 {
			return errors.New("住N免M规则无效")
		}
	case model.BoardingPolicyTypeHolidaySurcharge:
		var rule surchargeRule
		if err := json.Unmarshal([]byte(policy.RuleJSON), &rule); err != nil || rule.Surcharge <= 0 {
			return errors.New("节假日加收规则无效")
		}
	}
	return nil
}

func policyOverlapsStay(policy model.BoardingDiscountPolicy, checkInAt, checkOutAt string) bool {
	if policy.ValidFrom == "" && policy.ValidTo == "" {
		return true
	}
	start := checkInAt
	end := addDays(checkOutAt, -1)
	if policy.ValidTo != "" && start > policy.ValidTo {
		return false
	}
	if policy.ValidFrom != "" && end < policy.ValidFrom {
		return false
	}
	return true
}

func parsePolicySnapshot(snapshot string) []model.BoardingDiscountPolicy {
	var policies []model.BoardingDiscountPolicy
	if snapshot == "" {
		return policies
	}
	_ = json.Unmarshal([]byte(snapshot), &policies)
	return policies
}

func policyIDsFromPolicies(policies []model.BoardingDiscountPolicy) []uint {
	ids := make([]uint, 0, len(policies))
	for _, policy := range policies {
		if policy.ID > 0 {
			ids = append(ids, policy.ID)
		}
	}
	return ids
}

func buildBoardingOrderItems(orderID uint, cabinet *model.BoardingCabinet, preview *BoardingPricePreview) []model.OrderItem {
	items := make([]model.OrderItem, 0, len(preview.Lines))
	for _, line := range preview.Lines {
		itemType := 4
		switch line.Type {
		case "holiday_surcharge":
			itemType = 5
		case "discount", "member_discount", "manual_discount":
			itemType = 6
		}
		if line.Amount == 0 {
			continue
		}
		items = append(items, model.OrderItem{
			OrderID:   orderID,
			ItemType:  itemType,
			ItemID:    cabinet.ID,
			Name:      line.Label,
			Quantity:  maxInt(line.Quantity, 1),
			UnitPrice: line.UnitPrice,
			Amount:    line.Amount,
		})
	}
	return items
}

func collectBoardingPetIDs(order *model.BoardingOrder) []uint {
	ids := make([]uint, 0, len(order.Pets))
	for _, pet := range order.Pets {
		if pet.PetID > 0 {
			ids = append(ids, pet.PetID)
		}
	}
	return ids
}

func normalizeStayRange(checkInAt, checkOutAt string) (string, string, int, error) {
	startDate, err := normalizeDate(checkInAt)
	if err != nil {
		return "", "", 0, errors.New("入住日期格式错误")
	}
	endDate, err := normalizeDate(checkOutAt)
	if err != nil {
		return "", "", 0, errors.New("离店日期格式错误")
	}
	if endDate <= startDate {
		return "", "", 0, errors.New("离店日期必须晚于入住日期")
	}
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	nights := int(end.Sub(start).Hours() / 24)
	if nights < 1 {
		return "", "", 0, errors.New("至少需要寄养 1 天")
	}
	return startDate, endDate, nights, nil
}

func normalizeDate(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if len(text) >= 10 {
		text = text[:10]
	}
	t, err := time.Parse("2006-01-02", text)
	if err != nil {
		return "", errors.New("日期格式需为 YYYY-MM-DD")
	}
	return t.Format("2006-01-02"), nil
}

func addDays(dateText string, offset int) string {
	t, _ := time.Parse("2006-01-02", dateText)
	return t.AddDate(0, 0, offset).Format("2006-01-02")
}

func roundMoney(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func clampMinFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func uintPtr(v uint) *uint {
	if v == 0 {
		return nil
	}
	return &v
}
