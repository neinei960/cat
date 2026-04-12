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
	Rooms                  []BoardingRoomPreview          `json:"rooms,omitempty"`
}

type BoardingPreviewInput struct {
	CustomerID     uint
	PetIDs         []uint
	PetCount       int
	CabinetID      uint
	CheckInAt      string
	CheckOutAt     string
	PolicyIDs      []uint
	RoomGroups     []BoardingRoomGroupInput
	ExcludeOrderID uint
	ExcludeRoomID  uint
}

type BoardingCreateInput struct {
	CustomerID   uint
	PetIDs       []uint
	CabinetID    uint
	CheckInAt    string
	CheckOutAt   string
	PolicyIDs    []uint
	RoomGroups   []BoardingRoomGroupInput
	HasDeworming *bool
	Remark       string
	OperatorID   uint
}

type BoardingRoomGroupInput struct {
	PetIDs     []uint
	PetCount   int
	CabinetID  uint
	CheckInAt  string
	CheckOutAt string
}

type BoardingCheckInInput struct {
	DiscountAmount float64
}

type BoardingRoomPreview struct {
	RoomIndex              int                 `json:"room_index"`
	CabinetID              uint                `json:"cabinet_id"`
	CabinetType            string              `json:"cabinet_type"`
	PetIDs                 []uint              `json:"pet_ids,omitempty"`
	PetCount               int                 `json:"pet_count"`
	CheckInAt              string              `json:"check_in_at"`
	CheckOutAt             string              `json:"check_out_at"`
	Nights                 int                 `json:"nights"`
	RegularNights          int                 `json:"regular_nights"`
	HolidayNights          int                 `json:"holiday_nights"`
	BaseAmount             float64             `json:"base_amount"`
	ExtraPetAmount         float64             `json:"extra_pet_amount"`
	HolidaySurchargeAmount float64             `json:"holiday_surcharge_amount"`
	DiscountAmount         float64             `json:"discount_amount"`
	ManualDiscountAmount   float64             `json:"manual_discount_amount"`
	PayAmount              float64             `json:"pay_amount"`
	Lines                  []BoardingPriceLine `json:"lines"`
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

func (s *BoardingService) GetAvailableCabinets(shopID uint, checkInAt, checkOutAt string, petCount int, excludeOrderID, excludeRoomID uint) ([]model.BoardingCabinet, error) {
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
	activeCounts, _, err := s.listOverlappingCabinetUsage(shopID, startDate, endDate, excludeOrderID, excludeRoomID)
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

func (s *BoardingService) listOverlappingCabinetUsage(shopID uint, startDate, endDate string, excludeOrderID, excludeRoomID uint) (map[uint]int, []model.BoardingOrderRoom, error) {
	activeOrders, err := s.repo.ListActiveOrders(shopID)
	if err != nil {
		return nil, nil, err
	}
	counts := make(map[uint]int)
	rooms := make([]model.BoardingOrderRoom, 0)
	for _, order := range activeOrders {
		if len(order.Rooms) == 0 {
			if excludeOrderID > 0 && order.ID == excludeOrderID {
				continue
			}
			if order.CheckInAt < endDate && order.CheckOutAt > startDate && activeBoardingRoomStatus(order.Status) {
				counts[order.CabinetID]++
				rooms = append(rooms, legacyBoardingRoom(&order))
			}
			continue
		}
		for _, room := range order.Rooms {
			if excludeRoomID > 0 && room.ID == excludeRoomID {
				continue
			}
			if !activeBoardingRoomStatus(room.Status) {
				continue
			}
			if room.CheckInAt < endDate && room.CheckOutAt > startDate {
				counts[room.CabinetID]++
				rooms = append(rooms, room)
			}
		}
	}
	return counts, rooms, nil
}

func (s *BoardingService) Preview(shopID uint, input BoardingPreviewInput) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	if len(input.RoomGroups) > 0 {
		groups := normalizeBoardingRoomGroups(input.RoomGroups, input)
		if len(groups) != 1 {
			return nil, nil, nil, errors.New("分房预览请使用多房预览流程")
		}
		input.PetIDs = groups[0].PetIDs
		input.PetCount = groups[0].PetCount
		input.CabinetID = groups[0].CabinetID
		input.CheckInAt = groups[0].CheckInAt
		input.CheckOutAt = groups[0].CheckOutAt
	}
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
	availableCabinets, err := s.GetAvailableCabinets(shopID, input.CheckInAt, input.CheckOutAt, petCount, input.ExcludeOrderID, input.ExcludeRoomID)
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

func normalizeBoardingRoomGroups(groups []BoardingRoomGroupInput, legacy BoardingPreviewInput) []BoardingRoomGroupInput {
	if len(groups) > 0 {
		normalized := make([]BoardingRoomGroupInput, 0, len(groups))
		for _, group := range groups {
			if group.CabinetID == 0 {
				continue
			}
			normalized = append(normalized, group)
		}
		return normalized
	}
	if legacy.CabinetID == 0 {
		return nil
	}
	return []BoardingRoomGroupInput{{
		PetIDs:     append([]uint(nil), legacy.PetIDs...),
		PetCount:   legacy.PetCount,
		CabinetID:  legacy.CabinetID,
		CheckInAt:  legacy.CheckInAt,
		CheckOutAt: legacy.CheckOutAt,
	}}
}

func (s *BoardingService) resolveRoomModels(shopID, customerID uint, groups []BoardingRoomGroupInput, policyIDs []uint, excludeOrderID, excludeRoomID uint) ([]model.BoardingOrderRoom, error) {
	if len(groups) == 0 {
		return nil, errors.New("请至少选择一个房间分组")
	}
	resolved := make([]model.BoardingOrderRoom, 0, len(groups))
	seenPets := map[uint]struct{}{}
	requestedBySlot := map[string]int{}
	remainingBySlot := map[string]int{}

	for index, group := range groups {
		preview, cabinet, petIDs, err := s.Preview(shopID, BoardingPreviewInput{
			CustomerID:     customerID,
			PetIDs:         group.PetIDs,
			PetCount:       group.PetCount,
			CabinetID:      group.CabinetID,
			CheckInAt:      group.CheckInAt,
			CheckOutAt:     group.CheckOutAt,
			PolicyIDs:      policyIDs,
			ExcludeOrderID: excludeOrderID,
			ExcludeRoomID:  excludeRoomID,
		})
		if err != nil {
			return nil, err
		}
		for _, petID := range petIDs {
			if _, ok := seenPets[petID]; ok {
				return nil, errors.New("同一只猫不能重复分配到多个房间")
			}
			seenPets[petID] = struct{}{}
		}
		slotKey := fmt.Sprintf("%d|%s|%s", cabinet.ID, preview.CheckInAt, preview.CheckOutAt)
		if _, ok := remainingBySlot[slotKey]; !ok {
			activeCounts, _, err := s.listOverlappingCabinetUsage(shopID, preview.CheckInAt, preview.CheckOutAt, excludeOrderID, excludeRoomID)
			if err != nil {
				return nil, err
			}
			remainingBySlot[slotKey] = maxInt(cabinet.RoomCount-activeCounts[cabinet.ID], 0)
		}
		requestedBySlot[slotKey]++
		if requestedBySlot[slotKey] > remainingBySlot[slotKey] {
			return nil, fmt.Errorf("%s 在所选日期内房间不足", cabinet.CabinetType)
		}

		pets, err := s.loadPets(petIDs)
		if err != nil {
			return nil, err
		}
		policySnapshot, _ := json.Marshal(preview.Policies)
		priceSnapshot, _ := json.Marshal(preview)
		room := model.BoardingOrderRoom{
			CabinetID:              cabinet.ID,
			RoomIndex:              index + 1,
			CheckInAt:              preview.CheckInAt,
			CheckOutAt:             preview.CheckOutAt,
			Nights:                 preview.Nights,
			BaseAmount:             preview.BaseAmount,
			HolidaySurchargeAmount: preview.HolidaySurchargeAmount,
			DiscountAmount:         preview.DiscountAmount,
			ManualDiscountAmount:   0,
			PayAmount:              preview.PayAmount,
			Status:                 model.BoardingOrderStatusPendingCheckin,
			PolicySnapshotJSON:     string(policySnapshot),
			PriceSnapshotJSON:      string(priceSnapshot),
			Cabinet:                cabinet,
		}
		if len(pets) > 0 {
			room.Pets = make([]model.BoardingOrderPet, 0, len(pets))
			for _, pet := range pets {
				room.Pets = append(room.Pets, model.BoardingOrderPet{
					PetID:           pet.ID,
					PetNameSnapshot: pet.Name,
				})
			}
		}
		resolved = append(resolved, room)
	}
	return resolved, nil
}

func (s *BoardingService) PreviewOrder(shopID uint, input BoardingPreviewInput) (*BoardingPricePreview, error) {
	roomGroups := normalizeBoardingRoomGroups(input.RoomGroups, input)
	rooms, err := s.resolveRoomModels(shopID, input.CustomerID, roomGroups, input.PolicyIDs, input.ExcludeOrderID, input.ExcludeRoomID)
	if err != nil {
		return nil, err
	}
	preview := buildAggregatePreviewFromRooms(input.CustomerID, rooms)
	if preview == nil {
		return nil, errors.New("无法生成寄养预览")
	}
	return preview, nil
}

func (s *BoardingService) CreateOrder(shopID uint, input BoardingCreateInput) (*model.BoardingOrder, error) {
	if input.CustomerID == 0 {
		return nil, errors.New("请选择客户")
	}
	roomGroups := normalizeBoardingRoomGroups(input.RoomGroups, BoardingPreviewInput{
		PetIDs:     input.PetIDs,
		CabinetID:  input.CabinetID,
		CheckInAt:  input.CheckInAt,
		CheckOutAt: input.CheckOutAt,
	})
	rooms, err := s.resolveRoomModels(shopID, input.CustomerID, roomGroups, input.PolicyIDs, 0, 0)
	if err != nil {
		return nil, err
	}
	preview := buildAggregatePreviewFromRooms(input.CustomerID, rooms)
	if preview == nil {
		return nil, errors.New("无法生成寄养预览")
	}
	if _, err := s.customerRepo.FindByID(input.CustomerID); err != nil {
		return nil, errors.New("客户不存在")
	}
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	boardingOrder := &model.BoardingOrder{
		ShopID:                 shopID,
		CustomerID:             input.CustomerID,
		StaffID:                input.OperatorID,
		CabinetID:              rooms[0].CabinetID,
		CheckInAt:              preview.CheckInAt,
		CheckOutAt:             preview.CheckOutAt,
		Nights:                 preview.Nights,
		BaseAmount:             preview.BaseAmount,
		HolidaySurchargeAmount: preview.HolidaySurchargeAmount,
		DiscountAmount:         preview.DiscountAmount,
		PayAmount:              preview.PayAmount,
		Status:                 model.BoardingOrderStatusPendingCheckin,
		HasDeworming:           input.HasDeworming,
		Remark:                 strings.TrimSpace(input.Remark),
		PolicySnapshotJSON:     string(policySnapshot),
		PriceSnapshotJSON:      string(priceSnapshot),
	}

	var createdID uint
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(boardingOrder).Error; err != nil {
			return err
		}
		allPetIDs := make([]uint, 0)
		for _, room := range rooms {
			for _, pet := range room.Pets {
				if pet.PetID > 0 {
					allPetIDs = append(allPetIDs, pet.PetID)
				}
			}
		}
		var orderPetID *uint
		if len(allPetIDs) == 1 {
			orderPetID = &allPetIDs[0]
		}
		orderTotal := roundMoney(preview.BaseAmount + preview.HolidaySurchargeAmount)
		order := &model.Order{
			OrderNo:               utils.GenerateOrderNo(),
			ShopID:                shopID,
			CustomerID:            &input.CustomerID,
			PetID:                 orderPetID,
			StaffID:               uintPtr(input.OperatorID),
			TotalAmount:           orderTotal,
			ServiceTotal:          orderTotal,
			DiscountAmount:        preview.DiscountAmount,
			ServiceDiscountAmount: preview.DiscountAmount,
			DiscountRate:          calculateOrderDiscountRate(orderTotal, preview.PayAmount),
			PayAmount:             preview.PayAmount,
			PayStatus:             0,
			Status:                0,
			Remark:                strings.TrimSpace(input.Remark),
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		boardingOrder.OrderID = &order.ID
		if err := tx.Save(boardingOrder).Error; err != nil {
			return err
		}
		items := buildBoardingOrderItemsFromAggregate(order.ID, preview)
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		for _, draftRoom := range rooms {
			roomRecord := draftRoom
			roomRecord.BoardingOrderID = boardingOrder.ID
			roomPets := append([]model.BoardingOrderPet(nil), roomRecord.Pets...)
			roomRecord.Pets = nil
			roomRecord.Cabinet = nil
			if err := tx.Create(&roomRecord).Error; err != nil {
				return err
			}
			if len(roomPets) > 0 {
				for i := range roomPets {
					roomPets[i].BoardingOrderID = boardingOrder.ID
					roomPets[i].BoardingOrderRoomID = &roomRecord.ID
				}
				if err := tx.Create(&roomPets).Error; err != nil {
					return err
				}
			}
		}
		if err := tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: boardingOrder.ID,
			OperatorID:      input.OperatorID,
			Action:          "create",
			Content:         fmt.Sprintf("创建寄养单，共 %d 个房间分组，入住 %s，离店 %s", len(rooms), preview.CheckInAt, preview.CheckOutAt),
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
	list, total, err := s.repo.ListBoardingOrders(shopID, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range list {
		normalizeLoadedBoardingOrder(&list[i])
	}
	return list, total, nil
}

func (s *BoardingService) GetOrder(shopID, id uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	normalizeLoadedBoardingOrder(order)
	return order, nil
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
		normalizeLoadedBoardingOrder(&order)
		for _, room := range order.Rooms {
			if !activeBoardingRoomStatus(room.Status) {
				continue
			}
			entry := order
			entry.CabinetID = room.CabinetID
			entry.CheckInAt = room.CheckInAt
			entry.CheckOutAt = room.CheckOutAt
			entry.ActualCheckOutAt = room.ActualCheckOutAt
			entry.Nights = room.Nights
			entry.BaseAmount = room.BaseAmount
			entry.HolidaySurchargeAmount = room.HolidaySurchargeAmount
			entry.ManualDiscountAmount = room.ManualDiscountAmount
			entry.PayAmount = roundMoney(maxBoardingFloat(room.PayAmount-room.ManualDiscountAmount, 0))
			entry.Status = room.Status
			entry.RoomIndex = room.RoomIndex
			entry.Cabinet = room.Cabinet
			entry.Pets = room.Pets
			ordersByCabinet[room.CabinetID] = append(ordersByCabinet[room.CabinetID], entry)
			if room.Status == model.BoardingOrderStatusCheckedIn {
				occupiedCount[room.CabinetID]++
			} else if room.CheckInAt >= today {
				reservedCount[room.CabinetID]++
			} else {
				occupiedCount[room.CabinetID]++
			}
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
	if len(order.Rooms) == 0 {
		return s.checkInLegacy(shopID, order, operatorID, input)
	}
	if len(order.Rooms) != 1 {
		return nil, errors.New("多房寄养请在房间分组中操作")
	}
	return s.CheckInRoom(shopID, id, order.Rooms[0].ID, operatorID, input)
}

func (s *BoardingService) CheckInRoom(shopID, id, roomID, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.checkInLegacy(shopID, order, operatorID, input)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != model.BoardingOrderStatusPendingCheckin {
		return nil, errors.New("当前房间状态不可办理入住")
	}
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, room.CheckOutAt)
	if err != nil {
		return nil, err
	}
	if _, err := applyManualDiscountToBoardingPreview(preview, input.DiscountAmount); err != nil {
		return nil, err
	}
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)
	manualDiscount := roundMoney(input.DiscountAmount)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if order.OrderID != nil && *order.OrderID > 0 {
			var payOrder model.Order
			if err := tx.First(&payOrder, *order.OrderID).Error; err == nil && payOrder.PayStatus == 1 && manualDiscount > 0 {
				return errors.New("已支付订单不可在入住时追加优惠")
			}
		}
		room.Status = model.BoardingOrderStatusCheckedIn
		room.Nights = preview.Nights
		room.BaseAmount = preview.BaseAmount
		room.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
		room.DiscountAmount = preview.DiscountAmount
		room.ManualDiscountAmount = manualDiscount
		room.PayAmount = preview.PayAmount
		room.PolicySnapshotJSON = string(policySnapshot)
		room.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		if err := syncBoardingPayOrder(tx, order, aggregatePreview, false); err != nil {
			if strings.Contains(err.Error(), "已支付订单不可修改") && manualDiscount == 0 {
				return nil
			}
			return err
		}
		content := fmt.Sprintf("%s 办理入住", roomGroupLabel(room.RoomIndex))
		if manualDiscount > 0 {
			content = fmt.Sprintf("%s 办理入住，享受优惠 ¥%.2f", roomGroupLabel(room.RoomIndex), manualDiscount)
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
	if len(order.Rooms) == 0 {
		return s.checkOutLegacy(shopID, order, operatorID, actualDate)
	}
	if len(order.Rooms) != 1 {
		return nil, errors.New("多房寄养请在房间分组中操作")
	}
	return s.CheckOutRoom(shopID, id, order.Rooms[0].ID, operatorID, actualDate)
}

func (s *BoardingService) CheckOutRoom(shopID, id, roomID, operatorID uint, actualDate string) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.checkOutLegacy(shopID, order, operatorID, actualDate)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != model.BoardingOrderStatusCheckedIn {
		return nil, errors.New("当前房间状态不可办理离店")
	}
	actualDate, err = normalizeDate(actualDate)
	if err != nil {
		return nil, err
	}
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, actualDate)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.ActualCheckOutAt = actualDate
		room.CheckOutAt = actualDate
		room.Nights = preview.Nights
		room.BaseAmount = preview.BaseAmount
		room.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
		room.DiscountAmount = preview.DiscountAmount
		room.ManualDiscountAmount = manualDiscount
		room.PayAmount = preview.PayAmount
		room.Status = model.BoardingOrderStatusCheckedOut
		room.PolicySnapshotJSON = string(policySnapshot)
		room.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		return syncBoardingPayOrder(tx, order, aggregatePreview, true)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "check_out",
		Content:         fmt.Sprintf("%s 办理离店，实际离店日期 %s", roomGroupLabel(room.RoomIndex), actualDate),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) Extend(shopID, id, operatorID uint, newCheckOutAt string) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 {
		return s.extendLegacy(shopID, order, operatorID, newCheckOutAt)
	}
	if len(order.Rooms) != 1 {
		return nil, errors.New("多房寄养请在房间分组中操作")
	}
	return s.ExtendRoom(shopID, id, order.Rooms[0].ID, operatorID, newCheckOutAt)
}

func (s *BoardingService) ExtendRoom(shopID, id, roomID, operatorID uint, newCheckOutAt string) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.extendLegacy(shopID, order, operatorID, newCheckOutAt)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureEditableRoom(order, room); err != nil {
		return nil, err
	}
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, newCheckOutAt)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.CheckOutAt = preview.CheckOutAt
		room.Nights = preview.Nights
		room.BaseAmount = preview.BaseAmount
		room.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
		room.DiscountAmount = preview.DiscountAmount
		room.ManualDiscountAmount = manualDiscount
		room.PayAmount = preview.PayAmount
		room.PolicySnapshotJSON = string(policySnapshot)
		room.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		return syncBoardingPayOrder(tx, order, aggregatePreview, false)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "extend",
		Content:         fmt.Sprintf("%s 续住至 %s", roomGroupLabel(room.RoomIndex), preview.CheckOutAt),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) ChangeCabinet(shopID, id, operatorID, cabinetID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 {
		return s.changeCabinetLegacy(shopID, order, operatorID, cabinetID)
	}
	if len(order.Rooms) != 1 {
		return nil, errors.New("多房寄养请在房间分组中操作")
	}
	return s.ChangeRoomCabinet(shopID, id, order.Rooms[0].ID, operatorID, cabinetID)
}

func (s *BoardingService) ChangeRoomCabinet(shopID, id, roomID, operatorID, cabinetID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.changeCabinetLegacy(shopID, order, operatorID, cabinetID)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureEditableRoom(order, room); err != nil {
		return nil, err
	}
	preview, cabinet, _, err := s.previewExistingRoom(shopID, order, room, cabinetID, room.CheckOutAt)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.CabinetID = cabinet.ID
		room.Nights = preview.Nights
		room.BaseAmount = preview.BaseAmount
		room.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
		room.DiscountAmount = preview.DiscountAmount
		room.ManualDiscountAmount = manualDiscount
		room.PayAmount = preview.PayAmount
		room.PolicySnapshotJSON = string(policySnapshot)
		room.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		return syncBoardingPayOrder(tx, order, aggregatePreview, false)
	})
	if err != nil {
		return nil, err
	}
	_ = database.DB.Create(&model.BoardingOrderLog{
		BoardingOrderID: order.ID,
		OperatorID:      operatorID,
		Action:          "change_cabinet",
		Content:         fmt.Sprintf("%s 更换寄养房型为 %s", roomGroupLabel(room.RoomIndex), cabinet.CabinetType),
	}).Error
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) Cancel(shopID, id, operatorID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 {
		return s.cancelLegacy(shopID, order, operatorID)
	}
	if order.Order != nil && order.Order.PayStatus == 1 {
		return nil, errors.New("已支付订单不可取消")
	}
	for _, room := range order.Rooms {
		if room.Status == model.BoardingOrderStatusCheckedIn || room.Status == model.BoardingOrderStatusCheckedOut {
			return nil, errors.New("已有房间开始寄养，无法整单取消")
		}
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		for _, room := range order.Rooms {
			if room.Status == model.BoardingOrderStatusCancelled {
				continue
			}
			room.Status = model.BoardingOrderStatusCancelled
			room.ManualDiscountAmount = 0
			room.PayAmount = 0
			if err := tx.Save(&room).Error; err != nil {
				return err
			}
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		if err := syncBoardingPayOrder(tx, order, aggregatePreview, false); err != nil {
			return err
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "cancel",
			Content:         "整单取消寄养订单",
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) CancelRoom(shopID, id, roomID, operatorID uint) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.cancelLegacy(shopID, order, operatorID)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if room.Status != model.BoardingOrderStatusPendingCheckin {
		return nil, errors.New("当前房间状态不可取消")
	}
	if order.Order != nil && order.Order.PayStatus == 1 {
		return nil, errors.New("已支付订单不可取消")
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.Status = model.BoardingOrderStatusCancelled
		room.ManualDiscountAmount = 0
		room.PayAmount = 0
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		if err := syncBoardingPayOrder(tx, order, aggregatePreview, false); err != nil {
			return err
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "cancel",
			Content:         fmt.Sprintf("%s 已取消", roomGroupLabel(room.RoomIndex)),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.repo.FindBoardingOrderByID(shopID, id)
}

func (s *BoardingService) checkInLegacy(shopID uint, order *model.BoardingOrder, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
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
	return s.repo.FindBoardingOrderByID(shopID, order.ID)
}

func (s *BoardingService) checkOutLegacy(shopID uint, order *model.BoardingOrder, operatorID uint, actualDate string) (*model.BoardingOrder, error) {
	if order.Status != model.BoardingOrderStatusCheckedIn {
		return nil, errors.New("当前状态不可办理离店")
	}
	actualDate, err := normalizeDate(actualDate)
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
	return s.repo.FindBoardingOrderByID(shopID, order.ID)
}

func (s *BoardingService) extendLegacy(shopID uint, order *model.BoardingOrder, operatorID uint, newCheckOutAt string) (*model.BoardingOrder, error) {
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
	return s.repo.FindBoardingOrderByID(shopID, order.ID)
}

func (s *BoardingService) changeCabinetLegacy(shopID uint, order *model.BoardingOrder, operatorID, cabinetID uint) (*model.BoardingOrder, error) {
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
	return s.repo.FindBoardingOrderByID(shopID, order.ID)
}

func (s *BoardingService) cancelLegacy(shopID uint, order *model.BoardingOrder, operatorID uint) (*model.BoardingOrder, error) {
	if order.Status != model.BoardingOrderStatusPendingCheckin {
		return nil, errors.New("当前状态不可取消")
	}
	if order.Order != nil && order.Order.PayStatus == 1 {
		return nil, errors.New("已支付订单不可取消")
	}
	err := database.DB.Transaction(func(tx *gorm.DB) error {
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
	return s.repo.FindBoardingOrderByID(shopID, order.ID)
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

func (s *BoardingService) previewExistingRoom(shopID uint, order *model.BoardingOrder, room *model.BoardingOrderRoom, cabinetID uint, targetCheckOutAt string) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	if cabinetID == 0 {
		cabinetID = room.CabinetID
	}
	if targetCheckOutAt == "" {
		targetCheckOutAt = room.CheckOutAt
	}
	petIDs := petIDsFromRoom(*room)
	return s.Preview(shopID, BoardingPreviewInput{
		CustomerID:     order.CustomerID,
		PetIDs:         petIDs,
		PetCount:       maxInt(len(petIDs), 1),
		CabinetID:      cabinetID,
		CheckInAt:      room.CheckInAt,
		CheckOutAt:     targetCheckOutAt,
		PolicyIDs:      policyIDsFromPolicies(parsePolicySnapshot(room.PolicySnapshotJSON)),
		ExcludeOrderID: order.ID,
		ExcludeRoomID:  room.ID,
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

func (s *BoardingService) ensureEditableRoom(order *model.BoardingOrder, room *model.BoardingOrderRoom) error {
	if room.Status == model.BoardingOrderStatusCancelled || room.Status == model.BoardingOrderStatusCheckedOut {
		return errors.New("当前房间状态不可修改")
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
	if len(ids) == 0 {
		for _, room := range order.Rooms {
			ids = append(ids, petIDsFromRoom(room)...)
		}
	}
	return ids
}

func findBoardingRoom(rooms []model.BoardingOrderRoom, roomID uint) (*model.BoardingOrderRoom, error) {
	for i := range rooms {
		if rooms[i].ID == roomID {
			return &rooms[i], nil
		}
	}
	return nil, errors.New("房间分组不存在")
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
