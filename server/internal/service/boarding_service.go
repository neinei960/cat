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
	Type          string  `json:"type"`
	Label         string  `json:"label"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	Amount        float64 `json:"amount"`
	SpecialItemID uint    `json:"special_item_id,omitempty"`
}

const (
	fixedBoardingDepositAmount      = 200.0
	defaultHolidaySurchargeAmount   = 30.0
	defaultHolidaySurchargeName     = "节假日加收"
	defaultHolidaySurchargeRemark   = "创建寄养节假日时自动生成，可在寄养优惠中修改金额或停用"
	defaultHolidaySurchargePriority = 20
)

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
	SpecialItemID          *uint                          `json:"special_item_id,omitempty"`
	SpecialItemName        string                         `json:"special_item_name,omitempty"`
	SpecialItemDailyPrice  float64                        `json:"special_item_daily_price,omitempty"`
	SpecialItemDays        int                            `json:"special_item_days,omitempty"`
	SpecialItemAmount      float64                        `json:"special_item_amount"`
	DiscountAmount         float64                        `json:"discount_amount"`
	PayAmount              float64                        `json:"pay_amount"`
	Policies               []model.BoardingDiscountPolicy `json:"policies"`
	Lines                  []BoardingPriceLine            `json:"lines"`
	Rooms                  []BoardingRoomPreview          `json:"rooms,omitempty"`
}

type BoardingPreviewInput struct {
	CustomerID             uint
	PetIDs                 []uint
	PetCount               int
	CabinetID              uint
	CheckInAt              string
	CheckOutAt             string
	AvailabilityCheckInAt  string
	AvailabilityCheckOutAt string
	DepositEnabled         bool
	PolicyIDs              []uint
	RoomGroups             []BoardingRoomGroupInput
	SpecialItemID          uint
	SpecialItemDailyPrice  float64
	SpecialItemDays        int
	SpecialItems           []BoardingSpecialItemSelection
	ExcludeOrderID         uint
	ExcludeRoomID          uint
}

type BoardingCreateInput struct {
	CustomerID            uint
	PetIDs                []uint
	CabinetID             uint
	CheckInAt             string
	CheckOutAt            string
	DepositEnabled        bool
	SpecialItemID         uint
	SpecialItemDailyPrice float64
	SpecialItemDays       int
	SpecialItems          []BoardingSpecialItemSelection
	PolicyIDs             []uint
	RoomGroups            []BoardingRoomGroupInput
	HasDeworming          *bool
	Remark                string
	OperatorID            uint
}

type BoardingRoomGroupInput struct {
	PetIDs                []uint
	PetCount              int
	CabinetID             uint
	CheckInAt             string
	CheckOutAt            string
	SpecialItemID         uint
	SpecialItemDailyPrice float64
	SpecialItemDays       int
	SpecialItems          []BoardingSpecialItemSelection
}

type BoardingCheckInInput struct {
	DiscountAmount        float64
	SpecialItemID         *uint
	SpecialItemDailyPrice *float64
	SpecialItemDays       *int
	SpecialItems          *[]BoardingSpecialItemSelection
}

type BoardingRoomPreview struct {
	RoomIndex              int                 `json:"room_index"`
	CabinetID              uint                `json:"cabinet_id"`
	CabinetType            string              `json:"cabinet_type"`
	PetIDs                 []uint              `json:"pet_ids,omitempty"`
	PetCount               int                 `json:"pet_count"`
	SpecialItemID          *uint               `json:"special_item_id,omitempty"`
	SpecialItemName        string              `json:"special_item_name,omitempty"`
	SpecialItemDailyPrice  float64             `json:"special_item_daily_price,omitempty"`
	SpecialItemDays        int                 `json:"special_item_days,omitempty"`
	CheckInAt              string              `json:"check_in_at"`
	CheckOutAt             string              `json:"check_out_at"`
	Nights                 int                 `json:"nights"`
	RegularNights          int                 `json:"regular_nights"`
	HolidayNights          int                 `json:"holiday_nights"`
	BaseAmount             float64             `json:"base_amount"`
	ExtraPetAmount         float64             `json:"extra_pet_amount"`
	HolidaySurchargeAmount float64             `json:"holiday_surcharge_amount"`
	SpecialItemAmount      float64             `json:"special_item_amount"`
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

type resolvedBoardingSpecialItem struct {
	ID         *uint
	Name       string
	DailyPrice float64
	Days       int
	Amount     float64
	Lines      []BoardingPriceLine
}

type BoardingSpecialItemSelection struct {
	ID         uint
	DailyPrice float64
	Days       int
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

	discountableAmount := roundMoney(maxFloat(preview.PayAmount-preview.SpecialItemAmount, 0))
	if discountableAmount <= 0 {
		return preview
	}

	discountedPay := roundMoney(discountableAmount * serviceDiscountRate)
	memberDiscountAmount := roundMoney(discountableAmount - discountedPay)
	if memberDiscountAmount <= 0 {
		return preview
	}

	adjusted := *preview
	adjusted.DiscountAmount = roundMoney(preview.DiscountAmount + memberDiscountAmount)
	adjusted.PayAmount = roundMoney(discountedPay + preview.SpecialItemAmount)
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

func applyBoardingDepositToPreview(preview *BoardingPricePreview, deposit float64) (*BoardingPricePreview, float64) {
	if preview == nil {
		return nil, 0
	}

	deposit = roundMoney(deposit)
	if deposit <= 0 {
		return preview, 0
	}

	deduction := roundMoney(minFloat(deposit, preview.PayAmount))
	if deduction <= 0 {
		return preview, 0
	}

	adjusted := *preview
	adjusted.PayAmount = roundMoney(preview.PayAmount - deduction)
	adjusted.Lines = append(append([]BoardingPriceLine{}, preview.Lines...), BoardingPriceLine{
		Type:      "boarding_deposit",
		Label:     "定金抵扣",
		Quantity:  1,
		UnitPrice: -deduction,
		Amount:    -deduction,
	})
	return &adjusted, deduction
}

func applyBoardingDepositFields(order *model.Order, deposit, deduction float64) {
	if order == nil {
		return
	}

	order.AppointmentDepositAmount = roundMoney(maxFloat(deposit, 0))
	order.AppointmentDepositDeductionAmount = roundMoney(maxFloat(deduction, 0))
}

func boardingDepositAmountForOrder(order *model.BoardingOrder) float64 {
	if order == nil || order.Order == nil {
		return 0
	}
	return roundMoney(maxFloat(order.Order.AppointmentDepositAmount, 0))
}

func boardingDepositDeductionFromPreview(preview *BoardingPricePreview) float64 {
	if preview == nil {
		return 0
	}
	for _, line := range preview.Lines {
		if line.Type == "boarding_deposit" {
			return roundMoney(maxFloat(-line.Amount, 0))
		}
	}
	return 0
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
	holiday.SurchargeAmount = roundMoney(holiday.SurchargeAmount)
	if holiday.SurchargeAmount < 0 {
		return errors.New("节假日加收金额不能小于0")
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(holiday).Error; err != nil {
			return err
		}
		return s.ensureDefaultHolidaySurchargePolicy(tx, holiday.ShopID)
	})
}

func (s *BoardingService) CreateHolidayRange(shopID uint, startDate, endDate, name string, surchargeAmount float64) ([]model.BoardingHoliday, error) {
	startText, err := normalizeDate(startDate)
	if err != nil {
		return nil, errors.New("开始日期格式错误")
	}
	endText, err := normalizeDate(endDate)
	if err != nil {
		return nil, errors.New("结束日期格式错误")
	}
	if endText < startText {
		return nil, errors.New("结束日期不能早于开始日期")
	}
	if name == "" {
		name = "节假日"
	}
	surchargeAmount = roundMoney(surchargeAmount)
	if surchargeAmount < 0 {
		return nil, errors.New("节假日加收金额不能小于0")
	}

	existing, err := s.repo.ListHolidaysInRange(shopID, startText, addDays(endText, 1))
	if err != nil {
		return nil, err
	}
	existingDates := make(map[string]struct{}, len(existing))
	for _, holiday := range existing {
		existingDates[holiday.HolidayDate] = struct{}{}
	}

	created := make([]model.BoardingHoliday, 0)
	for cursor := startText; cursor <= endText; cursor = addDays(cursor, 1) {
		if _, ok := existingDates[cursor]; ok {
			continue
		}
		created = append(created, model.BoardingHoliday{
			ShopID:          shopID,
			HolidayDate:     cursor,
			Name:            name,
			SurchargeAmount: surchargeAmount,
		})
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if len(created) > 0 {
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
		}
		return s.ensureDefaultHolidaySurchargePolicy(tx, shopID)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *BoardingService) UpdateHolidayRange(shopID uint, ids []uint, startDate, endDate, name string, surchargeAmount float64) ([]model.BoardingHoliday, error) {
	if len(ids) == 0 {
		return nil, errors.New("请选择要修改的节假日范围")
	}
	startText, err := normalizeDate(startDate)
	if err != nil {
		return nil, errors.New("开始日期格式错误")
	}
	endText, err := normalizeDate(endDate)
	if err != nil {
		return nil, errors.New("结束日期格式错误")
	}
	if endText < startText {
		return nil, errors.New("结束日期不能早于开始日期")
	}
	if name == "" {
		name = "节假日"
	}
	surchargeAmount = roundMoney(surchargeAmount)
	if surchargeAmount < 0 {
		return nil, errors.New("节假日加收金额不能小于0")
	}

	idSet := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id > 0 {
			idSet[id] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return nil, errors.New("请选择要修改的节假日范围")
	}

	created := make([]model.BoardingHoliday, 0)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var ownedCount int64
		if err := tx.Model(&model.BoardingHoliday{}).Where("shop_id = ? AND id IN ?", shopID, ids).Count(&ownedCount).Error; err != nil {
			return err
		}
		if int(ownedCount) != len(idSet) {
			return errors.New("节假日范围不存在")
		}

		var conflicts []model.BoardingHoliday
		if err := tx.Where("shop_id = ? AND holiday_date >= ? AND holiday_date <= ?", shopID, startText, endText).
			Find(&conflicts).Error; err != nil {
			return err
		}
		for _, holiday := range conflicts {
			if _, ok := idSet[holiday.ID]; !ok {
				return errors.New("所选日期已存在其他节假日配置")
			}
		}

		if err := tx.Where("shop_id = ? AND id IN ?", shopID, ids).Delete(&model.BoardingHoliday{}).Error; err != nil {
			return err
		}
		for cursor := startText; cursor <= endText; cursor = addDays(cursor, 1) {
			created = append(created, model.BoardingHoliday{
				ShopID:          shopID,
				HolidayDate:     cursor,
				Name:            name,
				SurchargeAmount: surchargeAmount,
			})
		}
		if len(created) > 0 {
			if err := tx.Create(&created).Error; err != nil {
				return err
			}
		}
		return s.ensureDefaultHolidaySurchargePolicy(tx, shopID)
	})
	if err != nil {
		return nil, err
	}
	return created, nil
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

func (s *BoardingService) ensureDefaultHolidaySurchargePolicy(tx *gorm.DB, shopID uint) error {
	var existing model.BoardingDiscountPolicy
	err := tx.Where("shop_id = ? AND policy_type = ?", shopID, model.BoardingPolicyTypeHolidaySurcharge).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	ruleJSON, _ := json.Marshal(surchargeRule{Surcharge: defaultHolidaySurchargeAmount})
	policy := model.BoardingDiscountPolicy{
		ShopID:     shopID,
		Name:       defaultHolidaySurchargeName,
		PolicyType: model.BoardingPolicyTypeHolidaySurcharge,
		RuleJSON:   string(ruleJSON),
		Priority:   defaultHolidaySurchargePriority,
		Stackable:  true,
		Status:     1,
		Remark:     defaultHolidaySurchargeRemark,
	}
	if err := validateBoardingPolicy(&policy); err != nil {
		return err
	}
	return tx.Create(&policy).Error
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

func (s *BoardingService) ListSpecialItems(shopID uint, onlyActive bool) ([]model.BoardingSpecialItem, error) {
	return s.repo.ListSpecialItems(shopID, onlyActive)
}

func (s *BoardingService) CreateSpecialItem(item *model.BoardingSpecialItem) error {
	if err := validateBoardingSpecialItem(item); err != nil {
		return err
	}
	return database.DB.Create(item).Error
}

func (s *BoardingService) UpdateSpecialItem(shopID uint, item *model.BoardingSpecialItem) error {
	existing, err := s.repo.FindSpecialItemByID(shopID, item.ID)
	if err != nil {
		return err
	}
	existing.Name = item.Name
	existing.DefaultDailyPrice = item.DefaultDailyPrice
	existing.SortOrder = item.SortOrder
	existing.Status = item.Status
	existing.Remark = item.Remark
	if err := validateBoardingSpecialItem(existing); err != nil {
		return err
	}
	return database.DB.Save(existing).Error
}

func (s *BoardingService) DeleteSpecialItem(shopID, id uint) error {
	return database.DB.Where("shop_id = ?", shopID).Delete(&model.BoardingSpecialItem{}, id).Error
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
	dailyCounts := make(map[uint]map[string]int)
	rooms := make([]model.BoardingOrderRoom, 0)
	addUsage := func(cabinetID uint, checkInAt, checkOutAt string) {
		overlapStart := maxDateText(startDate, checkInAt)
		overlapEnd := minDateText(endDate, checkOutAt)
		for cursor := overlapStart; cursor < overlapEnd; cursor = addDays(cursor, 1) {
			if _, ok := dailyCounts[cabinetID]; !ok {
				dailyCounts[cabinetID] = make(map[string]int)
			}
			dailyCounts[cabinetID][cursor]++
			if dailyCounts[cabinetID][cursor] > counts[cabinetID] {
				counts[cabinetID] = dailyCounts[cabinetID][cursor]
			}
		}
	}
	for _, order := range activeOrders {
		if len(order.Rooms) == 0 {
			if excludeOrderID > 0 && order.ID == excludeOrderID {
				continue
			}
			if order.CheckInAt < endDate && order.CheckOutAt > startDate && activeBoardingRoomStatus(order.Status) {
				addUsage(order.CabinetID, order.CheckInAt, order.CheckOutAt)
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
				addUsage(room.CabinetID, room.CheckInAt, room.CheckOutAt)
				rooms = append(rooms, room)
			}
		}
	}
	return counts, rooms, nil
}

func maxDateText(a, b string) string {
	if a > b {
		return a
	}
	return b
}

func minDateText(a, b string) string {
	if a < b {
		return a
	}
	return b
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
		input.SpecialItemID = groups[0].SpecialItemID
		input.SpecialItemDailyPrice = groups[0].SpecialItemDailyPrice
		input.SpecialItemDays = groups[0].SpecialItemDays
		input.SpecialItems = groups[0].SpecialItems
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
	availabilityCheckInAt := input.CheckInAt
	availabilityCheckOutAt := input.CheckOutAt
	if strings.TrimSpace(input.AvailabilityCheckInAt) != "" && strings.TrimSpace(input.AvailabilityCheckOutAt) != "" && input.AvailabilityCheckInAt < input.AvailabilityCheckOutAt {
		availabilityCheckInAt = input.AvailabilityCheckInAt
		availabilityCheckOutAt = input.AvailabilityCheckOutAt
	}
	availableCabinets, err := s.GetAvailableCabinets(shopID, availabilityCheckInAt, availabilityCheckOutAt, petCount, input.ExcludeOrderID, input.ExcludeRoomID)
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
	specialItem, err := s.resolveSpecialItemSelection(shopID, input)
	if err != nil {
		return nil, nil, nil, err
	}
	preview, err := s.computePreview(shopID, cabinet, input.CheckInAt, input.CheckOutAt, petCount, selectedPolicies, specialItem)
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
		PetIDs:                append([]uint(nil), legacy.PetIDs...),
		PetCount:              legacy.PetCount,
		CabinetID:             legacy.CabinetID,
		CheckInAt:             legacy.CheckInAt,
		CheckOutAt:            legacy.CheckOutAt,
		SpecialItemID:         legacy.SpecialItemID,
		SpecialItemDailyPrice: legacy.SpecialItemDailyPrice,
		SpecialItemDays:       legacy.SpecialItemDays,
		SpecialItems:          append([]BoardingSpecialItemSelection(nil), legacy.SpecialItems...),
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
			CustomerID:            customerID,
			PetIDs:                group.PetIDs,
			PetCount:              group.PetCount,
			CabinetID:             group.CabinetID,
			CheckInAt:             group.CheckInAt,
			CheckOutAt:            group.CheckOutAt,
			SpecialItemID:         group.SpecialItemID,
			SpecialItemDailyPrice: group.SpecialItemDailyPrice,
			SpecialItemDays:       group.SpecialItemDays,
			SpecialItems:          append([]BoardingSpecialItemSelection(nil), group.SpecialItems...),
			PolicyIDs:             policyIDs,
			ExcludeOrderID:        excludeOrderID,
			ExcludeRoomID:         excludeRoomID,
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
			SpecialItemID:          cloneUint(preview.SpecialItemID),
			SpecialItemName:        strings.TrimSpace(preview.SpecialItemName),
			SpecialItemDailyPrice:  roundMoney(preview.SpecialItemDailyPrice),
			SpecialItemDays:        preview.SpecialItemDays,
			SpecialItemAmount:      roundMoney(preview.SpecialItemAmount),
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
	if input.DepositEnabled {
		preview, _ = applyBoardingDepositToPreview(preview, fixedBoardingDepositAmount)
	}
	return preview, nil
}

func (s *BoardingService) CreateOrder(shopID uint, input BoardingCreateInput) (*model.BoardingOrder, error) {
	if input.CustomerID == 0 {
		return nil, errors.New("请选择客户")
	}
	roomGroups := normalizeBoardingRoomGroups(input.RoomGroups, BoardingPreviewInput{
		PetIDs:                input.PetIDs,
		CabinetID:             input.CabinetID,
		CheckInAt:             input.CheckInAt,
		CheckOutAt:            input.CheckOutAt,
		SpecialItemID:         input.SpecialItemID,
		SpecialItemDailyPrice: input.SpecialItemDailyPrice,
		SpecialItemDays:       input.SpecialItemDays,
		SpecialItems:          append([]BoardingSpecialItemSelection(nil), input.SpecialItems...),
	})
	rooms, err := s.resolveRoomModels(shopID, input.CustomerID, roomGroups, input.PolicyIDs, 0, 0)
	if err != nil {
		return nil, err
	}
	preview := buildAggregatePreviewFromRooms(input.CustomerID, rooms)
	if preview == nil {
		return nil, errors.New("无法生成寄养预览")
	}
	depositAmount := 0.0
	if input.DepositEnabled {
		depositAmount = fixedBoardingDepositAmount
	}
	preview, depositDeduction := applyBoardingDepositToPreview(preview, depositAmount)
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
		SpecialItemAmount:      preview.SpecialItemAmount,
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
		serviceTotal := boardingServiceAmount(preview)
		orderTotal := roundMoney(serviceTotal + preview.SpecialItemAmount)
		order := &model.Order{
			OrderNo:               utils.GenerateOrderNo(),
			ShopID:                shopID,
			CustomerID:            &input.CustomerID,
			PetID:                 orderPetID,
			StaffID:               uintPtr(input.OperatorID),
			TotalAmount:           orderTotal,
			ServiceTotal:          serviceTotal,
			AddonTotal:            roundMoney(preview.SpecialItemAmount),
			DiscountAmount:        preview.DiscountAmount,
			ServiceDiscountAmount: preview.DiscountAmount,
			DiscountRate:          calculateOrderDiscountRate(orderTotal, preview.PayAmount),
			PayAmount:             preview.PayAmount,
			PayStatus:             0,
			Status:                0,
			Remark:                strings.TrimSpace(input.Remark),
		}
		applyBoardingDepositFields(order, depositAmount, depositDeduction)
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

func (s *BoardingService) ListOrders(shopID uint, status, dateFrom, dateTo string, cabinetID uint, page, pageSize int) ([]model.BoardingOrder, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.repo.ListBoardingOrders(shopID, status, dateFrom, dateTo, cabinetID, page, pageSize)
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
	s.attachBoardingPaymentLogs(order)
	return order, nil
}

func (s *BoardingService) attachBoardingPaymentLogs(order *model.BoardingOrder) {
	if order == nil {
		return
	}
	paymentLogs := make([]model.BoardingPaymentLog, 0, 2)
	seenOrderIDs := map[uint]bool{}
	appendPaidOrder := func(payOrder *model.Order) {
		if payOrder == nil || payOrder.PayStatus != 1 || seenOrderIDs[payOrder.ID] {
			return
		}
		amount := paidBoardingAmountFromPayOrder(payOrder)
		if amount <= 0 {
			amount = roundMoney(payOrder.PayAmount)
		}
		paymentLogs = append(paymentLogs, model.BoardingPaymentLog{
			OrderID:   payOrder.ID,
			OrderNo:   payOrder.OrderNo,
			PayAmount: amount,
			PayMethod: payOrder.PayMethod,
			PayTime:   payOrder.PayTime,
		})
		seenOrderIDs[payOrder.ID] = true
	}

	appendPaidOrder(order.Order)
	if sourceOrderNo := sourceBoardingPaidOrderNo(order.Order); sourceOrderNo != "" {
		var sourceOrder model.Order
		if err := database.DB.Preload("Items").
			Where("shop_id = ? AND order_no = ?", order.ShopID, sourceOrderNo).
			First(&sourceOrder).Error; err == nil {
			appendPaidOrder(&sourceOrder)
		}
	}

	sort.Slice(paymentLogs, func(i, j int) bool {
		if paymentLogs[i].PayTime == nil {
			return false
		}
		if paymentLogs[j].PayTime == nil {
			return true
		}
		return paymentLogs[i].PayTime.Before(*paymentLogs[j].PayTime)
	})
	order.PaymentLogs = paymentLogs
}

func sourceBoardingPaidOrderNo(payOrder *model.Order) string {
	if payOrder == nil || payOrder.Remark == "" {
		return ""
	}
	const prefix = "原订单 "
	start := strings.Index(payOrder.Remark, prefix)
	if start < 0 {
		return ""
	}
	rest := payOrder.Remark[start+len(prefix):]
	end := strings.Index(rest, " ")
	if end < 0 {
		return strings.TrimSpace(rest)
	}
	return strings.TrimSpace(rest[:end])
}

func (s *BoardingService) DeleteOrder(shopID, id uint, role string) error {
	if !model.HasStaffRoleAtLeast(role, model.StaffRoleManager) {
		return errors.New("仅店长可删除历史寄养订单")
	}
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return errors.New("寄养订单不存在")
	}
	if !isHistoricalBoardingOrder(order) {
		return errors.New("仅历史寄养订单可删除")
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := deleteLinkedPayOrders(tx, shopID, order.OrderID, nil); err != nil {
			return err
		}
		if err := tx.Where("boarding_order_id = ?", order.ID).Delete(&model.BoardingOrderLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("boarding_order_id = ?", order.ID).Delete(&model.BoardingOrderPet{}).Error; err != nil {
			return err
		}
		if err := tx.Where("boarding_order_id = ?", order.ID).Delete(&model.BoardingOrderRoom{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.BoardingOrder{}, order.ID).Error
	})
}

func (s *BoardingService) UpdateDeworming(shopID, id, operatorID uint, hasDeworming *bool) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if boardingDewormingEqual(order.HasDeworming, hasDeworming) {
		normalizeLoadedBoardingOrder(order)
		return order, nil
	}

	previousLabel := boardingDewormingLabel(order.HasDeworming)
	currentLabel := boardingDewormingLabel(hasDeworming)
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.BoardingOrder{}).
			Where("id = ? AND shop_id = ?", order.ID, shopID).
			Update("has_deworming", hasDeworming).Error; err != nil {
			return err
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "update_deworming",
			Content:         fmt.Sprintf("更新驱虫状态：%s -> %s", previousLabel, currentLabel),
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetOrder(shopID, id)
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
			if boardingRoomHistoryDate(room) < today {
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
		if groups[i].BasePrice != groups[j].BasePrice {
			return groups[i].BasePrice < groups[j].BasePrice
		}
		if groups[i].CabinetType != groups[j].CabinetType {
			return groups[i].CabinetType < groups[j].CabinetType
		}
		return groups[i].CabinetID < groups[j].CabinetID
	})
	return groups, nil
}

func boardingRoomHistoryDate(room model.BoardingOrderRoom) string {
	if room.ActualCheckOutAt != "" {
		return room.ActualCheckOutAt
	}
	return room.CheckOutAt
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

func (s *BoardingService) AdjustPrice(shopID, id, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 {
		return s.adjustPriceLegacy(shopID, order, operatorID, input)
	}
	if len(order.Rooms) != 1 {
		return nil, errors.New("多房寄养请在房间分组中操作")
	}
	return s.AdjustRoomPrice(shopID, id, order.Rooms[0].ID, operatorID, input)
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
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, room.CheckOutAt, &input)
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
		applyPreviewToBoardingRoom(room, preview, manualDiscount, string(policySnapshot), string(priceSnapshot))
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

func (s *BoardingService) AdjustRoomPrice(shopID, id, roomID, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
	order, err := s.repo.FindBoardingOrderByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if len(order.Rooms) == 0 || roomID == 0 {
		return s.adjustPriceLegacy(shopID, order, operatorID, input)
	}
	room, err := findBoardingRoom(order.Rooms, roomID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureEditableRoom(order, room); err != nil {
		return nil, err
	}
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, room.CheckOutAt, &input)
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
		applyPreviewToBoardingRoom(room, preview, manualDiscount, string(policySnapshot), string(priceSnapshot))
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
		content := fmt.Sprintf("%s 调整价格", roomGroupLabel(room.RoomIndex))
		if manualDiscount > 0 {
			content = fmt.Sprintf("%s 调整入住优惠为 ¥%.2f", roomGroupLabel(room.RoomIndex), manualDiscount)
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "adjust_price",
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
	preview, _, _, err := s.previewExistingRoom(shopID, order, room, room.CabinetID, actualDate, nil)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.ActualCheckOutAt = actualDate
		room.CheckOutAt = actualDate
		applyPreviewToBoardingRoom(room, preview, manualDiscount, string(policySnapshot), string(priceSnapshot))
		room.Status = model.BoardingOrderStatusCheckedOut
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
	if err := s.ensureExtendableRoom(order, room, newCheckOutAt); err != nil {
		return nil, err
	}
	preview, _, _, err := s.previewExistingRoomForExtend(shopID, order, room, room.CabinetID, newCheckOutAt)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.CheckOutAt = preview.CheckOutAt
		room.ActualCheckOutAt = ""
		if room.Status == model.BoardingOrderStatusCheckedOut {
			room.Status = model.BoardingOrderStatusCheckedIn
		}
		applyPreviewToBoardingRoom(room, preview, manualDiscount, string(policySnapshot), string(priceSnapshot))
		if err := tx.Save(room).Error; err != nil {
			return err
		}
		_, aggregatePreview, err := s.refreshBoardingOrderAggregate(tx, order)
		if err != nil {
			return err
		}
		return s.syncBoardingPayOrderAfterExtend(tx, order, aggregatePreview, operatorID)
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
	preview, cabinet, _, err := s.previewExistingRoom(shopID, order, room, cabinetID, room.CheckOutAt, nil)
	if err != nil {
		return nil, err
	}
	manualDiscount := roundMoney(minFloat(room.ManualDiscountAmount, preview.PayAmount))
	policySnapshot, _ := json.Marshal(preview.Policies)
	priceSnapshot, _ := json.Marshal(preview)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		room.CabinetID = cabinet.ID
		applyPreviewToBoardingRoom(room, preview, manualDiscount, string(policySnapshot), string(priceSnapshot))
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
	adjustedPreview, _ = applyBoardingDepositToPreview(adjustedPreview, boardingDepositAmountForOrder(order))

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

func (s *BoardingService) adjustPriceLegacy(shopID uint, order *model.BoardingOrder, operatorID uint, input BoardingCheckInInput) (*model.BoardingOrder, error) {
	if err := s.ensureEditableOrder(order); err != nil {
		return nil, err
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
	adjustedPreview, _ = applyBoardingDepositToPreview(adjustedPreview, boardingDepositAmountForOrder(order))
	manualDiscount := roundMoney(input.DiscountAmount)

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.Nights = adjustedPreview.Nights
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = manualDiscount
		order.PayAmount = adjustedPreview.PayAmount
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if err := s.syncOrder(tx, order, cabinet, adjustedPreview, petIDs, false); err != nil {
			return err
		}
		content := "调整价格"
		if manualDiscount > 0 {
			content = fmt.Sprintf("调整入住优惠为 ¥%.2f", manualDiscount)
		}
		return tx.Create(&model.BoardingOrderLog{
			BoardingOrderID: order.ID,
			OperatorID:      operatorID,
			Action:          "adjust_price",
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
	adjustedPreview, _ = applyBoardingDepositToPreview(adjustedPreview, boardingDepositAmountForOrder(order))
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
	if err := s.ensureExtendableOrder(order, newCheckOutAt); err != nil {
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
	adjustedPreview, _ = applyBoardingDepositToPreview(adjustedPreview, boardingDepositAmountForOrder(order))
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		order.CheckOutAt = adjustedPreview.CheckOutAt
		order.ActualCheckOutAt = ""
		order.Nights = adjustedPreview.Nights
		order.BaseAmount = adjustedPreview.BaseAmount
		order.HolidaySurchargeAmount = adjustedPreview.HolidaySurchargeAmount
		order.DiscountAmount = adjustedPreview.DiscountAmount
		order.ManualDiscountAmount = appliedManualDiscount
		order.PayAmount = adjustedPreview.PayAmount
		if order.Status == model.BoardingOrderStatusCheckedOut {
			order.Status = model.BoardingOrderStatusCheckedIn
		}
		priceSnapshot, _ := json.Marshal(adjustedPreview)
		order.PriceSnapshotJSON = string(priceSnapshot)
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return s.syncLegacyPayOrderAfterExtend(tx, order, cabinet, adjustedPreview, petIDs, operatorID)
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
		PolicyIDs:  s.policyIDsForExistingReprice(shopID, selectedPolicies),
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
	adjustedPreview, _ = applyBoardingDepositToPreview(adjustedPreview, boardingDepositAmountForOrder(order))
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
		CustomerID:             order.CustomerID,
		PetIDs:                 petIDs,
		CabinetID:              order.CabinetID,
		CheckInAt:              order.CheckInAt,
		CheckOutAt:             targetCheckOutAt,
		AvailabilityCheckInAt:  order.CheckOutAt,
		AvailabilityCheckOutAt: targetCheckOutAt,
		PolicyIDs:              s.policyIDsForExistingReprice(shopID, policies),
		ExcludeOrderID:         order.ID,
	})
}

func previewInputSpecialSelection(room *model.BoardingOrderRoom, input *BoardingCheckInInput) []BoardingSpecialItemSelection {
	var specialItemID uint
	if room != nil && room.SpecialItemID != nil {
		specialItemID = *room.SpecialItemID
	}
	specialItemDailyPrice := 0.0
	if room != nil {
		specialItemDailyPrice = roundMoney(room.SpecialItemDailyPrice)
	}
	specialItemDays := 0
	if room != nil {
		specialItemDays = room.SpecialItemDays
	}
	if input != nil {
		if input.SpecialItems != nil {
			return append([]BoardingSpecialItemSelection(nil), (*input.SpecialItems)...)
		}
		if input.SpecialItemID != nil {
			specialItemID = *input.SpecialItemID
			if specialItemID == 0 {
				specialItemDailyPrice = 0
				specialItemDays = 0
			}
		}
		if input.SpecialItemDailyPrice != nil {
			specialItemDailyPrice = roundMoney(*input.SpecialItemDailyPrice)
		}
		if input.SpecialItemDays != nil {
			specialItemDays = *input.SpecialItemDays
		}
	}
	if specialItemID == 0 {
		return nil
	}
	return []BoardingSpecialItemSelection{{
		ID:         specialItemID,
		DailyPrice: specialItemDailyPrice,
		Days:       specialItemDays,
	}}
}

func applyPreviewToBoardingRoom(room *model.BoardingOrderRoom, preview *BoardingPricePreview, manualDiscount float64, policySnapshot string, priceSnapshot string) {
	if room == nil || preview == nil {
		return
	}
	room.SpecialItemID = cloneUint(preview.SpecialItemID)
	room.SpecialItemName = strings.TrimSpace(preview.SpecialItemName)
	room.SpecialItemDailyPrice = roundMoney(preview.SpecialItemDailyPrice)
	room.SpecialItemDays = preview.SpecialItemDays
	room.SpecialItemAmount = roundMoney(preview.SpecialItemAmount)
	room.Nights = preview.Nights
	room.BaseAmount = preview.BaseAmount
	room.HolidaySurchargeAmount = preview.HolidaySurchargeAmount
	room.DiscountAmount = preview.DiscountAmount
	room.ManualDiscountAmount = roundMoney(manualDiscount)
	room.PayAmount = preview.PayAmount
	room.PolicySnapshotJSON = policySnapshot
	room.PriceSnapshotJSON = priceSnapshot
}

func (s *BoardingService) previewExistingRoom(shopID uint, order *model.BoardingOrder, room *model.BoardingOrderRoom, cabinetID uint, targetCheckOutAt string, input *BoardingCheckInInput) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	return s.previewExistingRoomWithAvailability(shopID, order, room, cabinetID, targetCheckOutAt, "", "", input)
}

func (s *BoardingService) previewExistingRoomForExtend(shopID uint, order *model.BoardingOrder, room *model.BoardingOrderRoom, cabinetID uint, targetCheckOutAt string) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	return s.previewExistingRoomWithAvailability(shopID, order, room, cabinetID, targetCheckOutAt, room.CheckOutAt, targetCheckOutAt, nil)
}

func (s *BoardingService) previewExistingRoomWithAvailability(shopID uint, order *model.BoardingOrder, room *model.BoardingOrderRoom, cabinetID uint, targetCheckOutAt string, availabilityCheckInAt string, availabilityCheckOutAt string, input *BoardingCheckInInput) (*BoardingPricePreview, *model.BoardingCabinet, []uint, error) {
	if cabinetID == 0 {
		cabinetID = room.CabinetID
	}
	if targetCheckOutAt == "" {
		targetCheckOutAt = room.CheckOutAt
	}
	petIDs := petIDsFromRoom(*room)
	specialItems := previewInputSpecialSelection(room, input)
	policies := parsePolicySnapshot(room.PolicySnapshotJSON)
	return s.Preview(shopID, BoardingPreviewInput{
		CustomerID:             order.CustomerID,
		PetIDs:                 petIDs,
		PetCount:               maxInt(len(petIDs), 1),
		CabinetID:              cabinetID,
		CheckInAt:              room.CheckInAt,
		CheckOutAt:             targetCheckOutAt,
		AvailabilityCheckInAt:  availabilityCheckInAt,
		AvailabilityCheckOutAt: availabilityCheckOutAt,
		SpecialItems:           specialItems,
		PolicyIDs:              s.policyIDsForExistingReprice(shopID, policies),
		ExcludeOrderID:         order.ID,
		ExcludeRoomID:          room.ID,
	})
}

func (s *BoardingService) syncOrder(tx *gorm.DB, boardingOrder *model.BoardingOrder, cabinet *model.BoardingCabinet, preview *BoardingPricePreview, petIDs []uint, allowPaidCheckOut bool) error {
	if boardingOrder.OrderID == nil || *boardingOrder.OrderID == 0 {
		return nil
	}
	payOrder, ok, err := loadLinkedBoardingPayOrder(tx, boardingOrder)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if payOrder.PayStatus == 1 && !allowPaidCheckOut {
		return errors.New("已支付订单不可修改")
	}
	if payOrder.PayStatus == 1 && allowPaidCheckOut {
		return nil
	}
	preservedProductItems := clonePayOrderProductItems(payOrder.ID, payOrder.Items)
	paidCredit := boardingPaidCreditFromItems(payOrder.Items)
	syncBoardingPreviewToPayOrder(payOrder, preview, preservedProductItems)
	paidCredit = applyBoardingPaidCreditToPayOrder(payOrder, paidCredit)
	if err := tx.Save(&payOrder).Error; err != nil {
		return err
	}
	if err := tx.Where("order_id = ?", payOrder.ID).Delete(&model.OrderItem{}).Error; err != nil {
		return err
	}
	items := buildBoardingOrderItems(payOrder.ID, cabinet, preview)
	items = appendBoardingPaidCreditItem(items, payOrder.ID, paidCredit)
	items = append(items, preservedProductItems...)
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *BoardingService) syncBoardingPayOrderAfterExtend(tx *gorm.DB, order *model.BoardingOrder, preview *BoardingPricePreview, operatorID uint) error {
	if order == nil || order.OrderID == nil || *order.OrderID == 0 {
		return nil
	}
	payOrder, ok, err := loadLinkedBoardingPayOrder(tx, order)
	if err != nil || !ok {
		return err
	}
	if payOrder.PayStatus != 1 {
		return syncBoardingPayOrder(tx, order, preview, false)
	}
	paidCredit := paidBoardingAmountFromPayOrder(payOrder)
	if paidCredit <= 0 {
		return syncBoardingPayOrder(tx, order, preview, false)
	}
	if paidCredit >= roundMoney(preview.PayAmount) {
		return nil
	}
	return s.createBoardingExtensionBalanceOrder(tx, order, nil, preview, paidCredit, operatorID, payOrder.OrderNo)
}

func (s *BoardingService) syncLegacyPayOrderAfterExtend(tx *gorm.DB, order *model.BoardingOrder, cabinet *model.BoardingCabinet, preview *BoardingPricePreview, petIDs []uint, operatorID uint) error {
	if order == nil || order.OrderID == nil || *order.OrderID == 0 {
		return nil
	}
	payOrder, ok, err := loadLinkedBoardingPayOrder(tx, order)
	if err != nil || !ok {
		return err
	}
	if payOrder.PayStatus != 1 {
		return s.syncOrder(tx, order, cabinet, preview, petIDs, false)
	}
	paidCredit := paidBoardingAmountFromPayOrder(payOrder)
	if paidCredit <= 0 {
		return s.syncOrder(tx, order, cabinet, preview, petIDs, false)
	}
	if paidCredit >= roundMoney(preview.PayAmount) {
		return nil
	}
	return s.createBoardingExtensionBalanceOrder(tx, order, cabinet, preview, paidCredit, operatorID, payOrder.OrderNo)
}

func (s *BoardingService) createBoardingExtensionBalanceOrder(tx *gorm.DB, boardingOrder *model.BoardingOrder, cabinet *model.BoardingCabinet, preview *BoardingPricePreview, paidCredit float64, operatorID uint, sourceOrderNo string) error {
	if boardingOrder == nil || preview == nil {
		return nil
	}
	petIDs := collectBoardingPetIDs(boardingOrder)
	var orderPetID *uint
	if len(petIDs) == 1 {
		orderPetID = &petIDs[0]
	}
	payOrder := &model.Order{
		OrderNo:      utils.GenerateOrderNo(),
		ShopID:       boardingOrder.ShopID,
		CustomerID:   &boardingOrder.CustomerID,
		PetID:        orderPetID,
		StaffID:      uintPtr(operatorID),
		PayStatus:    0,
		Status:       0,
		Remark:       strings.TrimSpace(fmt.Sprintf("寄养续住差额，原订单 %s 已支付 ¥%.2f", sourceOrderNo, paidCredit)),
		DiscountRate: 1,
	}
	applyBoardingDepositFields(payOrder, boardingDepositAmountForOrder(boardingOrder), boardingDepositDeductionFromPreview(preview))
	syncBoardingPreviewToPayOrder(payOrder, preview, nil)
	appliedCredit := applyBoardingPaidCreditToPayOrder(payOrder, paidCredit)
	if appliedCredit <= 0 || payOrder.PayAmount <= 0 {
		return nil
	}
	if err := tx.Create(payOrder).Error; err != nil {
		return err
	}
	boardingOrder.OrderID = &payOrder.ID
	boardingOrder.Order = payOrder
	if err := tx.Model(boardingOrder).Update("order_id", payOrder.ID).Error; err != nil {
		return err
	}
	var items []model.OrderItem
	if len(boardingOrder.Rooms) > 0 {
		items = buildBoardingOrderItemsFromAggregate(payOrder.ID, preview)
	} else {
		items = buildBoardingOrderItems(payOrder.ID, cabinet, preview)
	}
	items = appendBoardingPaidCreditItem(items, payOrder.ID, appliedCredit)
	if len(items) > 0 {
		return tx.Create(&items).Error
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

func (s *BoardingService) ensureExtendableOrder(order *model.BoardingOrder, newCheckOutAt string) error {
	if order.Status == model.BoardingOrderStatusCancelled {
		return errors.New("当前状态不可续住")
	}
	if order.Status != model.BoardingOrderStatusPendingCheckin && order.Status != model.BoardingOrderStatusCheckedIn && order.Status != model.BoardingOrderStatusCheckedOut {
		return errors.New("当前状态不可续住")
	}
	if strings.TrimSpace(newCheckOutAt) <= order.CheckOutAt {
		return errors.New("续住日期需晚于当前离店日期")
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

func (s *BoardingService) ensureExtendableRoom(order *model.BoardingOrder, room *model.BoardingOrderRoom, newCheckOutAt string) error {
	if room.Status == model.BoardingOrderStatusCancelled {
		return errors.New("当前房间状态不可续住")
	}
	if room.Status != model.BoardingOrderStatusPendingCheckin && room.Status != model.BoardingOrderStatusCheckedIn && room.Status != model.BoardingOrderStatusCheckedOut {
		return errors.New("当前房间状态不可续住")
	}
	if strings.TrimSpace(newCheckOutAt) <= room.CheckOutAt {
		return errors.New("续住日期需晚于当前离店日期")
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

func cloneUint(value *uint) *uint {
	if value == nil || *value == 0 {
		return nil
	}
	cloned := *value
	return &cloned
}

func (s *BoardingService) resolveSpecialItemSelection(shopID uint, input BoardingPreviewInput) (*resolvedBoardingSpecialItem, error) {
	selections := append([]BoardingSpecialItemSelection(nil), input.SpecialItems...)
	if len(selections) == 0 && input.SpecialItemID > 0 {
		selections = append(selections, BoardingSpecialItemSelection{
			ID:         input.SpecialItemID,
			DailyPrice: input.SpecialItemDailyPrice,
			Days:       input.SpecialItemDays,
		})
	}

	if len(selections) == 0 {
		if input.SpecialItemDailyPrice != 0 || input.SpecialItemDays != 0 {
			return nil, errors.New("请选择特殊寄养项目")
		}
		return nil, nil
	}

	_, _, nights, err := normalizeStayRange(input.CheckInAt, input.CheckOutAt)
	if err != nil {
		return nil, err
	}
	seen := map[uint]struct{}{}
	seenNames := map[string]struct{}{}
	names := make([]string, 0, len(selections))
	lines := make([]BoardingPriceLine, 0, len(selections))
	var firstID *uint
	totalAmount := 0.0
	totalDailyPrice := 0.0
	maxDays := 0
	for _, selection := range selections {
		if selection.ID == 0 {
			continue
		}
		if _, ok := seen[selection.ID]; ok {
			return nil, errors.New("特殊寄养项目不能重复选择")
		}
		seen[selection.ID] = struct{}{}
		item, err := s.repo.FindSpecialItemByID(shopID, selection.ID)
		if err != nil {
			return nil, errors.New("特殊寄养项目不存在")
		}
		itemName := strings.TrimSpace(item.Name)
		if itemName != "" {
			if _, ok := seenNames[itemName]; ok {
				return nil, errors.New("同名特殊寄养项目不能重复选择")
			}
			seenNames[itemName] = struct{}{}
		}
		if selection.Days < 1 {
			return nil, errors.New("请填写特殊寄养天数")
		}
		if selection.Days > nights {
			return nil, errors.New("特殊寄养天数不能超过寄养晚数")
		}
		dailyPrice := roundMoney(selection.DailyPrice)
		if dailyPrice <= 0 {
			dailyPrice = roundMoney(item.DefaultDailyPrice)
		}
		if dailyPrice <= 0 {
			return nil, errors.New("请填写特殊寄养日价")
		}
		itemID := item.ID
		if firstID == nil {
			firstID = &itemID
		}
		amount := roundMoney(float64(selection.Days) * dailyPrice)
		names = append(names, itemName)
		totalAmount = roundMoney(totalAmount + amount)
		totalDailyPrice = roundMoney(totalDailyPrice + dailyPrice)
		maxDays = maxInt(maxDays, selection.Days)
		lines = append(lines, BoardingPriceLine{
			Type:          "special_item",
			Label:         itemName,
			Quantity:      selection.Days,
			UnitPrice:     dailyPrice,
			Amount:        amount,
			SpecialItemID: item.ID,
		})
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return &resolvedBoardingSpecialItem{
		ID:         firstID,
		Name:       strings.Join(names, "、"),
		DailyPrice: totalDailyPrice,
		Days:       maxDays,
		Amount:     totalAmount,
		Lines:      lines,
	}, nil
}

func (s *BoardingService) computePreview(shopID uint, cabinet *model.BoardingCabinet, checkInAt, checkOutAt string, petCount int, policies []model.BoardingDiscountPolicy, specialItem *resolvedBoardingSpecialItem) (*BoardingPricePreview, error) {
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
	var specialItemAmount float64
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
				holidaySurchargeAmount = roundMoney(resolveHolidaySurchargeAmount(startDate, endDate, holidayMap, rule.Surcharge))
				lines = append(lines, BoardingPriceLine{
					Type:      "holiday_surcharge",
					Label:     policy.Name,
					Quantity:  holidayNights,
					UnitPrice: resolveHolidaySurchargeUnitPrice(holidaySurchargeAmount, holidayNights, rule.Surcharge),
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
	if specialItem != nil && specialItem.Amount > 0 {
		specialItemAmount = roundMoney(specialItem.Amount)
		if len(specialItem.Lines) > 0 {
			lines = append(lines, specialItem.Lines...)
		} else {
			lines = append(lines, BoardingPriceLine{
				Type:      "special_item",
				Label:     specialItem.Name,
				Quantity:  specialItem.Days,
				UnitPrice: specialItem.DailyPrice,
				Amount:    specialItemAmount,
			})
		}
	}
	payAmount := roundMoney(baseAmount + holidaySurchargeAmount + specialItemAmount - discountAmount)
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
		SpecialItemID: cloneUint(func() *uint {
			if specialItem == nil {
				return nil
			}
			return specialItem.ID
		}()),
		SpecialItemName: func() string {
			if specialItem == nil {
				return ""
			}
			return specialItem.Name
		}(),
		SpecialItemDailyPrice: func() float64 {
			if specialItem == nil {
				return 0
			}
			return specialItem.DailyPrice
		}(),
		SpecialItemDays: func() int {
			if specialItem == nil {
				return 0
			}
			return specialItem.Days
		}(),
		SpecialItemAmount: specialItemAmount,
		DiscountAmount:    discountAmount,
		PayAmount:         payAmount,
		Policies:          policies,
		Lines:             lines,
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

func validateBoardingSpecialItem(item *model.BoardingSpecialItem) error {
	item.Name = strings.TrimSpace(item.Name)
	if item.Name == "" {
		return errors.New("请填写项目名称")
	}
	item.DefaultDailyPrice = roundMoney(item.DefaultDailyPrice)
	if item.DefaultDailyPrice <= 0 {
		return errors.New("请填写默认日价")
	}
	if item.Status == 0 {
		item.Status = 1
	}
	item.Remark = strings.TrimSpace(item.Remark)
	return nil
}

func resolveHolidaySurchargeAmount(startDate, endDate string, holidayMap map[string]model.BoardingHoliday, defaultSurcharge float64) float64 {
	var amount float64
	for cursor := startDate; cursor < endDate; cursor = addDays(cursor, 1) {
		holiday, ok := holidayMap[cursor]
		if !ok {
			continue
		}
		surcharge := roundMoney(holiday.SurchargeAmount)
		if surcharge <= 0 {
			surcharge = defaultSurcharge
		}
		amount += surcharge
	}
	return roundMoney(amount)
}

func resolveHolidaySurchargeUnitPrice(total float64, holidayNights int, defaultSurcharge float64) float64 {
	if holidayNights <= 0 {
		return 0
	}
	average := roundMoney(total / float64(holidayNights))
	if average > 0 {
		return average
	}
	return roundMoney(defaultSurcharge)
}

func boardingServiceAmount(preview *BoardingPricePreview) float64 {
	if preview == nil {
		return 0
	}
	return roundMoney(preview.BaseAmount + preview.HolidaySurchargeAmount)
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

func (s *BoardingService) policyIDsForExistingReprice(shopID uint, policies []model.BoardingDiscountPolicy) []uint {
	ids := make([]uint, 0, len(policies)+1)
	seen := map[uint]struct{}{}
	hasHolidaySurcharge := false
	for _, policy := range policies {
		if policy.PolicyType == model.BoardingPolicyTypeHolidaySurcharge {
			hasHolidaySurcharge = true
		}
		if policy.ID == 0 {
			continue
		}
		if _, ok := seen[policy.ID]; ok {
			continue
		}
		ids = append(ids, policy.ID)
		seen[policy.ID] = struct{}{}
	}
	if hasHolidaySurcharge {
		return ids
	}
	currentPolicies, err := s.repo.ListPolicies(shopID)
	if err != nil {
		return ids
	}
	for _, policy := range currentPolicies {
		if policy.ID == 0 || policy.Status != 1 || policy.PolicyType != model.BoardingPolicyTypeHolidaySurcharge {
			continue
		}
		if _, ok := seen[policy.ID]; ok {
			continue
		}
		ids = append(ids, policy.ID)
		seen[policy.ID] = struct{}{}
	}
	return ids
}

func buildBoardingOrderItems(orderID uint, cabinet *model.BoardingCabinet, preview *BoardingPricePreview) []model.OrderItem {
	items := make([]model.OrderItem, 0, len(preview.Lines))
	for _, line := range preview.Lines {
		itemType := 4
		switch line.Type {
		case "special_item":
			itemType = 3
		case "holiday_surcharge":
			itemType = 5
		case "discount", "member_discount", "manual_discount", "boarding_deposit":
			itemType = 6
		}
		if line.Amount == 0 {
			continue
		}
		itemID := cabinet.ID
		if line.Type == "special_item" && line.SpecialItemID > 0 {
			itemID = line.SpecialItemID
		}
		items = append(items, model.OrderItem{
			OrderID:   orderID,
			ItemType:  itemType,
			ItemID:    itemID,
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
	return roundOrderAmount(v)
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

func isHistoricalBoardingOrder(order *model.BoardingOrder) bool {
	if order == nil {
		return false
	}
	if order.Status == model.BoardingOrderStatusCheckedOut || order.Status == model.BoardingOrderStatusCancelled {
		return true
	}
	checkOutAt := strings.TrimSpace(order.ActualCheckOutAt)
	if checkOutAt == "" {
		checkOutAt = strings.TrimSpace(order.CheckOutAt)
	}
	date, err := time.Parse("2006-01-02", checkOutAt)
	if err != nil {
		return false
	}
	today, _ := time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	return date.Before(today)
}

func boardingDewormingEqual(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func boardingDewormingLabel(value *bool) string {
	if value == nil {
		return "未填写"
	}
	if *value {
		return "已驱虫"
	}
	return "未驱虫"
}
