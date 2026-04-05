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

type FeedingPricingSetting struct {
	BaseVisitPrice   float64 `json:"base_visit_price"`
	ExtraPetPrice    float64 `json:"extra_pet_price"`
	HolidaySurcharge float64 `json:"holiday_surcharge"`
}

type FeedingItemTemplate struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	ExtraPrice float64 `json:"extra_price"`
}

type FeedingSettings struct {
	Pricing FeedingPricingSetting `json:"pricing"`
	Items   []FeedingItemTemplate `json:"items"`
}

type FeedingAddressSnapshot struct {
	Address  string `json:"address"`
	Detail   string `json:"detail"`
	DoorCode string `json:"door_code"`
}

type FeedingPlanPetInput struct {
	PetID  uint   `json:"pet_id"`
	Remark string `json:"remark"`
}

type FeedingPlanRuleInput struct {
	Weekday    int    `json:"weekday"`
	WindowCode string `json:"window_code"`
	VisitCount int    `json:"visit_count"`
}

type FeedingPlanInput struct {
	CustomerID      uint                   `json:"customer_id"`
	AddressSnapshot FeedingAddressSnapshot `json:"address_snapshot"`
	ContactName     string                 `json:"contact_name"`
	ContactPhone    string                 `json:"contact_phone"`
	StartDate       string                 `json:"start_date"`
	EndDate         string                 `json:"end_date"`
	Remark          string                 `json:"remark"`
	Pets            []FeedingPlanPetInput  `json:"pets"`
	Rules           []FeedingPlanRuleInput `json:"rules"`
	ItemCodes       []string               `json:"item_codes"`
}

type FeedingPlanListResult struct {
	List  []model.FeedingPlan `json:"list"`
	Total int64               `json:"total"`
}

type FeedingDashboardGroup struct {
	Status string               `json:"status"`
	Label  string               `json:"label"`
	Count  int                  `json:"count"`
	Visits []model.FeedingVisit `json:"visits"`
}

type FeedingDashboardResponse struct {
	Date   string                  `json:"date"`
	Groups []FeedingDashboardGroup `json:"groups"`
}

type FeedingVisitFilterInput struct {
	ID            uint
	PlanID        uint
	ScheduledDate string
	Status        string
	StaffID       uint
	WindowCode    string
}

type FeedingAssignVisitInput struct {
	StaffID uint `json:"staff_id"`
}

type FeedingVisitItemCheck struct {
	ID      uint `json:"id"`
	Checked bool `json:"checked"`
}

type FeedingCompleteVisitInput struct {
	ItemChecks   []FeedingVisitItemCheck `json:"item_checks"`
	CustomerNote string                  `json:"customer_note"`
	InternalNote string                  `json:"internal_note"`
}

type FeedingExceptionVisitInput struct {
	ExceptionType string `json:"exception_type"`
	CustomerNote  string `json:"customer_note"`
	InternalNote  string `json:"internal_note"`
}

type FeedingVisitMediaInput struct {
	MediaType string `json:"media_type"`
	URL       string `json:"url"`
}

type FeedingService struct {
	repo         *repository.FeedingRepository
	orderRepo    *repository.OrderRepository
	customerRepo *repository.CustomerRepository
	petRepo      *repository.PetRepository
}

func NewFeedingService(repo *repository.FeedingRepository, orderRepo *repository.OrderRepository, customerRepo *repository.CustomerRepository, petRepo *repository.PetRepository) *FeedingService {
	return &FeedingService{
		repo:         repo,
		orderRepo:    orderRepo,
		customerRepo: customerRepo,
		petRepo:      petRepo,
	}
}

func (s *FeedingService) GetSettings(shopID uint) (*FeedingSettings, error) {
	setting, err := s.repo.GetSetting(shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaults := defaultFeedingSettings()
			return &defaults, nil
		}
		return nil, err
	}
	return decodeFeedingSettings(setting)
}

func (s *FeedingService) UpdatePricing(shopID, operatorID uint, pricing FeedingPricingSetting) (*FeedingSettings, error) {
	settings, err := s.GetSettings(shopID)
	if err != nil {
		return nil, err
	}
	settings.Pricing = normalizePricing(pricing)
	if err := s.saveSettings(shopID, operatorID, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *FeedingService) UpdateItems(shopID, operatorID uint, items []FeedingItemTemplate) (*FeedingSettings, error) {
	settings, err := s.GetSettings(shopID)
	if err != nil {
		return nil, err
	}
	settings.Items = normalizeItemTemplates(items)
	if err := s.saveSettings(shopID, operatorID, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *FeedingService) CreatePlan(shopID, operatorID uint, input FeedingPlanInput) (*model.FeedingPlan, error) {
	normalized, settings, err := s.normalizePlanInput(shopID, input)
	if err != nil {
		return nil, err
	}
	addressJSON, _ := json.Marshal(normalized.AddressSnapshot)
	pricingSnapshot, visitsPlan, err := s.buildPlanPricing(shopID, normalized, settings)
	if err != nil {
		return nil, err
	}
	pricingJSON, _ := json.Marshal(pricingSnapshot)
	selectedItemsJSON, _ := json.Marshal(resolveSelectedTemplates(settings.Items, normalized.ItemCodes))

	plan := &model.FeedingPlan{
		ShopID:              shopID,
		CustomerID:          normalized.CustomerID,
		AddressSnapshotJSON: string(addressJSON),
		ContactName:         normalized.ContactName,
		ContactPhone:        normalized.ContactPhone,
		StartDate:           normalized.StartDate,
		EndDate:             normalized.EndDate,
		TimeGranularity:     "window",
		Status:              model.FeedingPlanStatusActive,
		Remark:              normalized.Remark,
		PricingSnapshotJSON: string(pricingJSON),
		SelectedItemsJSON:   string(selectedItemsJSON),
		TotalAmount:         pricingSnapshot.TotalAmount,
		UnpaidAmount:        pricingSnapshot.TotalAmount,
	}

	tx := database.DB.Begin()
	if err := tx.Create(plan).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.replacePlanPets(tx, plan.ID, normalized.Pets); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.replacePlanRules(tx, plan.ID, normalized.Rules); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.createVisitsForPlan(tx, plan, visitsPlan, settings.Items, normalized.ItemCodes); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, 0, operatorID, "create_plan", fmt.Sprintf("创建上门喂养计划，共 %d 次上门", len(visitsPlan))); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindPlanByID(shopID, plan.ID)
}

func (s *FeedingService) ListPlans(shopID uint, page, pageSize int, status string) (*FeedingPlanListResult, error) {
	list, total, err := s.repo.ListPlans(shopID, repository.FeedingPlanFilter{
		Status: status,
		Page:   page,
		Size:   pageSize,
	})
	if err != nil {
		return nil, err
	}
	return &FeedingPlanListResult{List: list, Total: total}, nil
}

func (s *FeedingService) GetPlan(shopID, id uint, role string, staffID uint) (*model.FeedingPlan, error) {
	plan, err := s.repo.FindPlanByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if model.HasStaffRoleAtLeast(role, model.StaffRoleManager) {
		return plan, nil
	}
	visibleVisits := make([]model.FeedingVisit, 0, len(plan.Visits))
	for _, visit := range plan.Visits {
		if visit.StaffID == nil || *visit.StaffID == staffID {
			visibleVisits = append(visibleVisits, visit)
		}
	}
	plan.Visits = visibleVisits
	return plan, nil
}

func (s *FeedingService) UpdatePlan(shopID, operatorID, id uint, input FeedingPlanInput) (*model.FeedingPlan, error) {
	plan, err := s.repo.FindPlanByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if plan.Status == model.FeedingPlanStatusCancelled || plan.Status == model.FeedingPlanStatusCompleted {
		return nil, errors.New("当前状态不可修改计划")
	}
	normalized, settings, err := s.normalizePlanInput(shopID, input)
	if err != nil {
		return nil, err
	}
	addressJSON, _ := json.Marshal(normalized.AddressSnapshot)
	pricingSnapshot, visitsPlan, err := s.buildPlanPricing(shopID, normalized, settings)
	if err != nil {
		return nil, err
	}
	pricingJSON, _ := json.Marshal(pricingSnapshot)
	selectedItemsJSON, _ := json.Marshal(resolveSelectedTemplates(settings.Items, normalized.ItemCodes))

	tx := database.DB.Begin()
	plan.CustomerID = normalized.CustomerID
	plan.AddressSnapshotJSON = string(addressJSON)
	plan.ContactName = normalized.ContactName
	plan.ContactPhone = normalized.ContactPhone
	plan.StartDate = normalized.StartDate
	plan.EndDate = normalized.EndDate
	plan.Remark = normalized.Remark
	plan.PricingSnapshotJSON = string(pricingJSON)
	plan.SelectedItemsJSON = string(selectedItemsJSON)
	plan.TotalAmount = pricingSnapshot.TotalAmount
	plan.UnpaidAmount = pricingSnapshot.TotalAmount
	if err := tx.Save(plan).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.replacePlanPets(tx, plan.ID, normalized.Pets); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.replacePlanRules(tx, plan.ID, normalized.Rules); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.regeneratePendingVisits(tx, plan, visitsPlan, settings.Items, normalized.ItemCodes); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, 0, operatorID, "update_plan", "更新了上门喂养计划"); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindPlanByID(shopID, plan.ID)
}

func (s *FeedingService) PausePlan(shopID, operatorID, id uint) (*model.FeedingPlan, error) {
	return s.updatePlanStatus(shopID, operatorID, id, model.FeedingPlanStatusPaused, "pause_plan", "暂停喂养计划")
}

func (s *FeedingService) ResumePlan(shopID, operatorID, id uint) (*model.FeedingPlan, error) {
	return s.updatePlanStatus(shopID, operatorID, id, model.FeedingPlanStatusActive, "resume_plan", "恢复喂养计划")
}

func (s *FeedingService) CancelPlan(shopID, operatorID, id uint) (*model.FeedingPlan, error) {
	plan, err := s.repo.FindPlanByID(shopID, id)
	if err != nil {
		return nil, err
	}
	if plan.Status == model.FeedingPlanStatusCancelled {
		return plan, nil
	}
	for _, visit := range plan.Visits {
		if visit.Status == model.FeedingVisitStatusInProgress {
			return nil, errors.New("存在进行中的上门任务，无法取消计划")
		}
	}
	tx := database.DB.Begin()
	if err := tx.Model(&model.FeedingPlan{}).
		Where("id = ? AND shop_id = ?", id, shopID).
		Updates(map[string]any{
			"status":        model.FeedingPlanStatusCancelled,
			"unpaid_amount": 0,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Model(&model.FeedingVisit{}).
		Where("feeding_plan_id = ? AND status IN ?", id, []string{model.FeedingVisitStatusPending, model.FeedingVisitStatusAssigned}).
		Update("status", model.FeedingVisitStatusCancelled).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, 0, operatorID, "cancel_plan", "取消喂养计划"); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindPlanByID(shopID, id)
}

func (s *FeedingService) GetDashboard(shopID uint, role string, staffID uint, date string, filterStaffID uint, windowCode string) (*FeedingDashboardResponse, error) {
	dateText := date
	if dateText == "" {
		dateText = time.Now().Format("2006-01-02")
	}
	dateText, err := normalizeDate(dateText)
	if err != nil {
		return nil, err
	}

	visitFilter := repository.FeedingVisitFilter{
		ScheduledDate: dateText,
		WindowCode:    windowCode,
	}
	if model.HasStaffRoleAtLeast(role, model.StaffRoleManager) {
		if filterStaffID > 0 {
			visitFilter.StaffID = filterStaffID
		}
	} else {
		visitFilter.StaffID = staffID
	}
	visits, err := s.repo.ListVisits(shopID, visitFilter, true)
	if err != nil {
		return nil, err
	}
	groups := []FeedingDashboardGroup{
		{Status: model.FeedingVisitStatusPending, Label: "待上门"},
		{Status: model.FeedingVisitStatusAssigned, Label: "已分配"},
		{Status: model.FeedingVisitStatusInProgress, Label: "进行中"},
		{Status: model.FeedingVisitStatusException, Label: "异常"},
		{Status: model.FeedingVisitStatusDone, Label: "已完成"},
	}
	groupMap := make(map[string]*FeedingDashboardGroup, len(groups))
	for i := range groups {
		groupMap[groups[i].Status] = &groups[i]
	}
	for _, visit := range visits {
		group, ok := groupMap[visit.Status]
		if !ok {
			continue
		}
		group.Visits = append(group.Visits, visit)
		group.Count++
	}
	return &FeedingDashboardResponse{Date: dateText, Groups: groups}, nil
}

func (s *FeedingService) ListVisits(shopID uint, role string, staffID uint, filter FeedingVisitFilterInput) ([]model.FeedingVisit, error) {
	repoFilter := repository.FeedingVisitFilter{
		ID:            filter.ID,
		PlanID:        filter.PlanID,
		ScheduledDate: filter.ScheduledDate,
		Status:        filter.Status,
		WindowCode:    filter.WindowCode,
	}
	if model.HasStaffRoleAtLeast(role, model.StaffRoleManager) {
		repoFilter.StaffID = filter.StaffID
	} else {
		repoFilter.StaffID = staffID
	}
	return s.repo.ListVisits(shopID, repoFilter, true)
}

func (s *FeedingService) AssignVisit(shopID, operatorID, visitID uint, input FeedingAssignVisitInput) (*model.FeedingVisit, error) {
	visit, err := s.repo.FindVisitByID(shopID, visitID)
	if err != nil {
		return nil, err
	}
	if input.StaffID == 0 {
		return nil, errors.New("请选择执行员工")
	}
	tx := database.DB.Begin()
	if err := tx.Model(&model.FeedingVisit{}).
		Where("id = ? AND shop_id = ?", visit.ID, shopID).
		Updates(map[string]any{
			"staff_id": input.StaffID,
			"status":   model.FeedingVisitStatusAssigned,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, visit.ID, operatorID, "assign_visit", fmt.Sprintf("分配给员工 #%d", input.StaffID)); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindVisitByID(shopID, visitID)
}

func (s *FeedingService) StartVisit(shopID, operatorID, visitID uint, role string) (*model.FeedingVisit, error) {
	visit, err := s.repo.FindVisitByID(shopID, visitID)
	if err != nil {
		return nil, err
	}
	if err := ensureVisitPermission(visit, role, operatorID); err != nil {
		return nil, err
	}
	if visit.Plan != nil && visit.Plan.Status == model.FeedingPlanStatusPaused {
		return nil, errors.New("计划已暂停，无法开始执行")
	}
	if visit.Status != model.FeedingVisitStatusPending && visit.Status != model.FeedingVisitStatusAssigned {
		return nil, errors.New("当前状态不可开始执行")
	}
	now := time.Now()
	updates := map[string]any{
		"status":     model.FeedingVisitStatusInProgress,
		"arrived_at": &now,
	}
	if visit.StaffID == nil || *visit.StaffID == 0 {
		updates["staff_id"] = operatorID
	}
	tx := database.DB.Begin()
	if err := tx.Model(&model.FeedingVisit{}).Where("id = ?", visitID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, visitID, operatorID, "start_visit", "开始执行上门喂养"); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindVisitByID(shopID, visitID)
}

func (s *FeedingService) AddVisitMedia(shopID, operatorID, visitID uint, role string, input FeedingVisitMediaInput) (*model.FeedingVisitMedia, error) {
	visit, err := s.repo.FindVisitByID(shopID, visitID)
	if err != nil {
		return nil, err
	}
	if err := ensureVisitPermission(visit, role, operatorID); err != nil {
		return nil, err
	}
	if visit.Status != model.FeedingVisitStatusInProgress {
		return nil, errors.New("请先开始执行，再上传履约图片")
	}
	if strings.TrimSpace(input.URL) == "" {
		return nil, errors.New("请先上传图片")
	}
	mediaType := strings.TrimSpace(input.MediaType)
	if mediaType == "" {
		mediaType = "image"
	}
	record := &model.FeedingVisitMedia{
		FeedingVisitID: visitID,
		MediaType:      mediaType,
		URL:            strings.TrimSpace(input.URL),
	}
	tx := database.DB.Begin()
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, visitID, operatorID, "add_media", "上传了履约图片"); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (s *FeedingService) CompleteVisit(shopID, operatorID, visitID uint, role string, input FeedingCompleteVisitInput) (*model.FeedingVisit, error) {
	visit, err := s.repo.FindVisitByID(shopID, visitID)
	if err != nil {
		return nil, err
	}
	if err := ensureVisitPermission(visit, role, operatorID); err != nil {
		return nil, err
	}
	if visit.Status != model.FeedingVisitStatusInProgress {
		return nil, errors.New("请先开始执行，再完成任务")
	}
	if visit.ArrivedAt == nil {
		return nil, errors.New("缺少到达时间，请重新开始执行")
	}
	if len(visit.Media) == 0 {
		return nil, errors.New("至少上传 1 张图片后才能完成")
	}
	tx := database.DB.Begin()
	if err := applyVisitItemChecks(tx, visit.Items, input.ItemChecks); err != nil {
		tx.Rollback()
		return nil, err
	}
	var freshItems []model.FeedingVisitItem
	if err := tx.Where("feeding_visit_id = ?", visitID).Order("id ASC").Find(&freshItems).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, item := range freshItems {
		if !item.Checked {
			tx.Rollback()
			return nil, errors.New("请先勾选全部完成项")
		}
	}
	now := time.Now()
	if err := tx.Model(&model.FeedingVisit{}).
		Where("id = ?", visitID).
		Updates(map[string]any{
			"status":        model.FeedingVisitStatusDone,
			"completed_at":  &now,
			"customer_note": strings.TrimSpace(input.CustomerNote),
			"internal_note": strings.TrimSpace(input.InternalNote),
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, visitID, operatorID, "complete_visit", "完成一次上门喂养"); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := syncFeedingPlanStatusTx(tx, visit.FeedingPlanID); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindVisitByID(shopID, visitID)
}

func (s *FeedingService) SetVisitException(shopID, operatorID, visitID uint, role string, input FeedingExceptionVisitInput) (*model.FeedingVisit, error) {
	visit, err := s.repo.FindVisitByID(shopID, visitID)
	if err != nil {
		return nil, err
	}
	if err := ensureVisitPermission(visit, role, operatorID); err != nil {
		return nil, err
	}
	if visit.Status != model.FeedingVisitStatusAssigned && visit.Status != model.FeedingVisitStatusInProgress {
		return nil, errors.New("当前状态不可标记异常")
	}
	if strings.TrimSpace(input.ExceptionType) == "" {
		return nil, errors.New("请填写异常类型")
	}
	tx := database.DB.Begin()
	if err := tx.Model(&model.FeedingVisit{}).
		Where("id = ?", visitID).
		Updates(map[string]any{
			"status":         model.FeedingVisitStatusException,
			"exception_type": strings.TrimSpace(input.ExceptionType),
			"customer_note":  strings.TrimSpace(input.CustomerNote),
			"internal_note":  strings.TrimSpace(input.InternalNote),
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, visitID, operatorID, "exception_visit", strings.TrimSpace(input.ExceptionType)); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := syncFeedingPlanStatusTx(tx, visit.FeedingPlanID); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return s.repo.FindVisitByID(shopID, visitID)
}

func (s *FeedingService) GenerateOrder(shopID, operatorID, planID uint) (*model.Order, error) {
	plan, err := s.repo.FindPlanByID(shopID, planID)
	if err != nil {
		return nil, err
	}
	if plan.OrderID != nil && *plan.OrderID > 0 {
		savedOrder, err := s.orderRepo.FindByID(*plan.OrderID)
		if err != nil {
			return nil, err
		}
		decorateOrderTotals(savedOrder)
		savedOrder.OrderKind = buildOrderKind(savedOrder)
		return savedOrder, nil
	}
	for _, visit := range plan.Visits {
		if visit.Status == model.FeedingVisitStatusPending || visit.Status == model.FeedingVisitStatusAssigned || visit.Status == model.FeedingVisitStatusInProgress {
			return nil, errors.New("仍有未结束的上门任务，暂不能生成订单")
		}
	}
	doneVisits := make([]model.FeedingVisit, 0, len(plan.Visits))
	for _, visit := range plan.Visits {
		if visit.Status == model.FeedingVisitStatusDone {
			doneVisits = append(doneVisits, visit)
		}
	}
	if len(doneVisits) == 0 {
		return nil, errors.New("没有已完成的上门任务可结算")
	}

	snapshot, err := decodePricingSnapshot(plan.PricingSnapshotJSON)
	if err != nil {
		return nil, err
	}
	selectedItems, err := decodeSelectedTemplates(plan.SelectedItemsJSON)
	if err != nil {
		return nil, err
	}
	itemMap := make(map[string]FeedingItemTemplate, len(selectedItems))
	for _, item := range selectedItems {
		itemMap[item.Code] = item
	}
	holidayDates := make(map[string]struct{}, len(snapshot.HolidayDates))
	for _, date := range snapshot.HolidayDates {
		holidayDates[date] = struct{}{}
	}

	customerID := plan.CustomerID
	order := &model.Order{
		OrderNo:       utils.GenerateOrderNo(),
		ShopID:        shopID,
		CustomerID:    &customerID,
		FeedingPlanID: &plan.ID,
		PayStatus:     0,
		Status:        0,
		Remark:        fmt.Sprintf("上门喂养计划 #%d", plan.ID),
	}
	if len(plan.Pets) == 1 && plan.Pets[0].PetID > 0 {
		petID := plan.Pets[0].PetID
		order.PetID = &petID
	}
	orderItems := make([]model.OrderItem, 0)
	var totalAmount float64
	petCount := len(plan.Pets)
	extraPetAmount := roundMoney(float64(maxInt(petCount-1, 0)) * snapshot.Pricing.ExtraPetPrice)
	petLabel := ""
	if len(plan.Pets) == 1 {
		petLabel = strings.TrimSpace(plan.Pets[0].PetNameSnapshot)
		if plan.Pets[0].Pet != nil && strings.TrimSpace(plan.Pets[0].Pet.Name) != "" {
			petLabel = strings.TrimSpace(plan.Pets[0].Pet.Name)
		}
	}

	tx := database.DB.Begin()
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	for _, visit := range doneVisits {
		label := windowLabel(visit.WindowCode)
		baseAmount := roundMoney(snapshot.Pricing.BaseVisitPrice + extraPetAmount)
		baseName := fmt.Sprintf("上门喂养（%s %s）", formatMonthDay(visit.ScheduledDate), label)
		if petLabel != "" {
			baseName = fmt.Sprintf("%s · %s", petLabel, baseName)
		}
		orderItems = append(orderItems, model.OrderItem{
			OrderID:   order.ID,
			ItemType:  1,
			ItemID:    0,
			Name:      baseName,
			Quantity:  1,
			UnitPrice: baseAmount,
			Amount:    baseAmount,
		})
		totalAmount += baseAmount
		for _, visitItem := range visit.Items {
			template, ok := itemMap[visitItem.ItemCode]
			if !ok || template.ExtraPrice <= 0 {
				continue
			}
			extra := roundMoney(template.ExtraPrice)
			extraName := fmt.Sprintf("%s加收（%s %s）", template.Name, formatMonthDay(visit.ScheduledDate), label)
			if petLabel != "" {
				extraName = fmt.Sprintf("%s · %s", petLabel, extraName)
			}
			orderItems = append(orderItems, model.OrderItem{
				OrderID:   order.ID,
				ItemType:  3,
				ItemID:    0,
				Name:      extraName,
				Quantity:  1,
				UnitPrice: extra,
				Amount:    extra,
			})
			totalAmount += extra
		}
		if _, ok := holidayDates[visit.ScheduledDate]; ok && snapshot.Pricing.HolidaySurcharge > 0 {
			extra := roundMoney(snapshot.Pricing.HolidaySurcharge)
			holidayName := fmt.Sprintf("节假日加收（%s %s）", formatMonthDay(visit.ScheduledDate), label)
			if petLabel != "" {
				holidayName = fmt.Sprintf("%s · %s", petLabel, holidayName)
			}
			orderItems = append(orderItems, model.OrderItem{
				OrderID:   order.ID,
				ItemType:  3,
				ItemID:    0,
				Name:      holidayName,
				Quantity:  1,
				UnitPrice: extra,
				Amount:    extra,
			})
			totalAmount += extra
		}
	}
	totalAmount = roundMoney(totalAmount)
	order.TotalAmount = totalAmount
	order.ServiceTotal = totalAmount
	order.PayAmount = totalAmount
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(orderItems) > 0 {
		if err := tx.Create(&orderItems).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Model(&model.FeedingPlan{}).
		Where("id = ?", plan.ID).
		Updates(map[string]any{
			"order_id":      order.ID,
			"status":        model.FeedingPlanStatusCompleted,
			"unpaid_amount": totalAmount,
			"total_amount":  totalAmount,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, 0, operatorID, "generate_order", fmt.Sprintf("生成订单 %s", order.OrderNo)); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	savedOrder, err := s.orderRepo.FindByID(order.ID)
	if err != nil {
		return nil, err
	}
	decorateOrderTotals(savedOrder)
	savedOrder.OrderKind = buildOrderKind(savedOrder)
	return savedOrder, nil
}

type feedingPricingSnapshot struct {
	Pricing      FeedingPricingSetting `json:"pricing"`
	ItemCodes    []string              `json:"item_codes"`
	HolidayDates []string              `json:"holiday_dates"`
	VisitCount   int                   `json:"visit_count"`
	TotalAmount  float64               `json:"total_amount"`
}

type plannedFeedingVisit struct {
	ScheduledDate string
	WindowCode    string
	VisitPrice    float64
}

func (s *FeedingService) buildPlanPricing(shopID uint, input FeedingPlanInput, settings *FeedingSettings) (*feedingPricingSnapshot, []plannedFeedingVisit, error) {
	start, _ := time.Parse("2006-01-02", input.StartDate)
	end, _ := time.Parse("2006-01-02", input.EndDate)
	ruleMap := make(map[int]map[string]int)
	for _, rule := range input.Rules {
		if _, ok := ruleMap[rule.Weekday]; !ok {
			ruleMap[rule.Weekday] = make(map[string]int)
		}
		ruleMap[rule.Weekday][rule.WindowCode] += maxInt(rule.VisitCount, 1)
	}
	selectedItems := resolveSelectedTemplates(settings.Items, input.ItemCodes)
	selectedExtra := 0.0
	for _, item := range selectedItems {
		selectedExtra += item.ExtraPrice
	}
	selectedExtra = roundMoney(selectedExtra)
	holidayDates, err := listHolidayDates(shopID, input.StartDate, input.EndDate)
	if err != nil {
		return nil, nil, err
	}
	holidaySet := make(map[string]struct{}, len(holidayDates))
	for _, date := range holidayDates {
		holidaySet[date] = struct{}{}
	}
	visits := make([]plannedFeedingVisit, 0)
	totalAmount := 0.0
	petExtraAmount := roundMoney(float64(maxInt(len(input.Pets)-1, 0)) * settings.Pricing.ExtraPetPrice)
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		weekday := int(current.Weekday())
		windowCounts := ruleMap[weekday]
		if len(windowCounts) == 0 {
			continue
		}
		dateText := current.Format("2006-01-02")
		windows := []string{model.FeedingWindowMorning, model.FeedingWindowAfternoon, model.FeedingWindowEvening}
		for _, window := range windows {
			count := windowCounts[window]
			for i := 0; i < count; i++ {
				price := settings.Pricing.BaseVisitPrice + petExtraAmount + selectedExtra
				if _, ok := holidaySet[dateText]; ok {
					price += settings.Pricing.HolidaySurcharge
				}
				price = roundMoney(price)
				visits = append(visits, plannedFeedingVisit{
					ScheduledDate: dateText,
					WindowCode:    window,
					VisitPrice:    price,
				})
				totalAmount += price
			}
		}
	}
	if len(visits) == 0 {
		return nil, nil, errors.New("至少需要生成 1 条上门任务")
	}
	sort.Slice(visits, func(i, j int) bool {
		if visits[i].ScheduledDate == visits[j].ScheduledDate {
			return visits[i].WindowCode < visits[j].WindowCode
		}
		return visits[i].ScheduledDate < visits[j].ScheduledDate
	})
	snapshot := &feedingPricingSnapshot{
		Pricing:      settings.Pricing,
		ItemCodes:    append([]string(nil), input.ItemCodes...),
		HolidayDates: holidayDates,
		VisitCount:   len(visits),
		TotalAmount:  roundMoney(totalAmount),
	}
	return snapshot, visits, nil
}

func (s *FeedingService) normalizePlanInput(shopID uint, input FeedingPlanInput) (FeedingPlanInput, *FeedingSettings, error) {
	input.ContactName = strings.TrimSpace(input.ContactName)
	input.ContactPhone = strings.TrimSpace(input.ContactPhone)
	input.Remark = strings.TrimSpace(input.Remark)
	if input.CustomerID == 0 {
		return input, nil, errors.New("请选择客户")
	}
	if _, err := s.customerRepo.FindByID(input.CustomerID); err != nil {
		return input, nil, errors.New("客户不存在")
	}
	startDate, err := normalizeDate(input.StartDate)
	if err != nil {
		return input, nil, err
	}
	endDate, err := normalizeDate(input.EndDate)
	if err != nil {
		return input, nil, err
	}
	if endDate < startDate {
		return input, nil, errors.New("结束日期不能早于开始日期")
	}
	input.StartDate = startDate
	input.EndDate = endDate
	if strings.TrimSpace(input.AddressSnapshot.Address) == "" {
		return input, nil, errors.New("请填写服务地址")
	}
	if len(input.Pets) == 0 {
		return input, nil, errors.New("请至少选择 1 只猫咪")
	}
	petSeen := map[uint]struct{}{}
	validPets := make([]FeedingPlanPetInput, 0, len(input.Pets))
	for _, petInput := range input.Pets {
		if petInput.PetID == 0 {
			continue
		}
		if _, ok := petSeen[petInput.PetID]; ok {
			continue
		}
		pet, err := s.petRepo.FindByID(petInput.PetID)
		if err != nil || pet.CustomerID == nil || *pet.CustomerID != input.CustomerID {
			return input, nil, errors.New("所选猫咪与客户不匹配")
		}
		petSeen[petInput.PetID] = struct{}{}
		validPets = append(validPets, FeedingPlanPetInput{
			PetID:  pet.ID,
			Remark: strings.TrimSpace(petInput.Remark),
		})
	}
	if len(validPets) == 0 {
		return input, nil, errors.New("请至少选择 1 只猫咪")
	}
	input.Pets = validPets
	validRules := make([]FeedingPlanRuleInput, 0, len(input.Rules))
	for _, rule := range input.Rules {
		if rule.Weekday < 0 || rule.Weekday > 6 {
			continue
		}
		rule.WindowCode = normalizeWindowCode(rule.WindowCode)
		if rule.WindowCode == "" {
			continue
		}
		rule.VisitCount = maxInt(rule.VisitCount, 1)
		validRules = append(validRules, rule)
	}
	if len(validRules) == 0 {
		return input, nil, errors.New("请至少配置 1 个上门时间窗")
	}
	input.Rules = validRules
	settings, err := s.GetSettings(shopID)
	if err != nil {
		return input, nil, err
	}
	validCodes := make([]string, 0, len(input.ItemCodes))
	templateMap := make(map[string]FeedingItemTemplate, len(settings.Items))
	for _, item := range settings.Items {
		templateMap[item.Code] = item
	}
	seenCode := map[string]struct{}{}
	for _, code := range input.ItemCodes {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := templateMap[code]; !ok {
			continue
		}
		if _, ok := seenCode[code]; ok {
			continue
		}
		seenCode[code] = struct{}{}
		validCodes = append(validCodes, code)
	}
	if len(validCodes) == 0 {
		return input, nil, errors.New("请至少选择 1 项服务内容")
	}
	input.ItemCodes = validCodes
	return input, settings, nil
}

func (s *FeedingService) saveSettings(shopID, operatorID uint, settings *FeedingSettings) error {
	setting, err := s.repo.GetSetting(shopID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		setting = &model.FeedingSetting{ShopID: shopID}
	}
	pricingJSON, _ := json.Marshal(settings.Pricing)
	itemsJSON, _ := json.Marshal(settings.Items)
	setting.PricingJSON = string(pricingJSON)
	setting.ItemsJSON = string(itemsJSON)
	setting.UpdatedBy = operatorID
	return s.repo.SaveSetting(setting)
}

func (s *FeedingService) replacePlanPets(tx *gorm.DB, planID uint, pets []FeedingPlanPetInput) error {
	if err := tx.Where("feeding_plan_id = ?", planID).Delete(&model.FeedingPlanPet{}).Error; err != nil {
		return err
	}
	records := make([]model.FeedingPlanPet, 0, len(pets))
	for _, petInput := range pets {
		pet, err := s.petRepo.FindByID(petInput.PetID)
		if err != nil {
			return err
		}
		records = append(records, model.FeedingPlanPet{
			FeedingPlanID:   planID,
			PetID:           petInput.PetID,
			PetNameSnapshot: pet.Name,
			Remark:          petInput.Remark,
		})
	}
	if len(records) == 0 {
		return nil
	}
	return tx.Create(&records).Error
}

func (s *FeedingService) replacePlanRules(tx *gorm.DB, planID uint, rules []FeedingPlanRuleInput) error {
	if err := tx.Where("feeding_plan_id = ?", planID).Delete(&model.FeedingPlanRule{}).Error; err != nil {
		return err
	}
	records := make([]model.FeedingPlanRule, 0, len(rules))
	for _, rule := range rules {
		records = append(records, model.FeedingPlanRule{
			FeedingPlanID: planID,
			Weekday:       rule.Weekday,
			WindowCode:    rule.WindowCode,
			VisitCount:    maxInt(rule.VisitCount, 1),
		})
	}
	if len(records) == 0 {
		return nil
	}
	return tx.Create(&records).Error
}

func (s *FeedingService) createVisitsForPlan(tx *gorm.DB, plan *model.FeedingPlan, visits []plannedFeedingVisit, templates []FeedingItemTemplate, itemCodes []string) error {
	selectedTemplates := resolveSelectedTemplates(templates, itemCodes)
	for _, item := range visits {
		visit := &model.FeedingVisit{
			ShopID:        plan.ShopID,
			FeedingPlanID: plan.ID,
			ScheduledDate: item.ScheduledDate,
			WindowCode:    item.WindowCode,
			Status:        model.FeedingVisitStatusPending,
			VisitPrice:    item.VisitPrice,
		}
		if err := tx.Create(visit).Error; err != nil {
			return err
		}
		if len(selectedTemplates) > 0 {
			items := make([]model.FeedingVisitItem, 0, len(selectedTemplates))
			for _, template := range selectedTemplates {
				items = append(items, model.FeedingVisitItem{
					FeedingVisitID:   visit.ID,
					ItemCode:         template.Code,
					ItemNameSnapshot: template.Name,
					ExtraPrice:       template.ExtraPrice,
				})
			}
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *FeedingService) regeneratePendingVisits(tx *gorm.DB, plan *model.FeedingPlan, visits []plannedFeedingVisit, templates []FeedingItemTemplate, itemCodes []string) error {
	var locked []model.FeedingVisit
	if err := tx.Where("feeding_plan_id = ? AND status NOT IN ?", plan.ID, []string{model.FeedingVisitStatusPending, model.FeedingVisitStatusAssigned}).
		Find(&locked).Error; err != nil {
		return err
	}
	lockedCounts := map[string]int{}
	for _, visit := range locked {
		key := fmt.Sprintf("%s|%s", visit.ScheduledDate, visit.WindowCode)
		lockedCounts[key]++
	}
	if err := tx.Where("feeding_plan_id = ? AND status IN ?", plan.ID, []string{model.FeedingVisitStatusPending, model.FeedingVisitStatusAssigned}).
		Delete(&model.FeedingVisit{}).Error; err != nil {
		return err
	}
	filtered := make([]plannedFeedingVisit, 0, len(visits))
	for _, visit := range visits {
		key := fmt.Sprintf("%s|%s", visit.ScheduledDate, visit.WindowCode)
		if lockedCounts[key] > 0 {
			lockedCounts[key]--
			continue
		}
		filtered = append(filtered, visit)
	}
	return s.createVisitsForPlan(tx, plan, filtered, templates, itemCodes)
}

func (s *FeedingService) appendPlanLog(tx *gorm.DB, visitID, operatorID uint, action, content string) error {
	if visitID > 0 {
		logRecord := &model.FeedingVisitLog{
			FeedingVisitID: visitID,
			OperatorID:     operatorID,
			Action:         action,
			Content:        content,
		}
		return tx.Create(logRecord).Error
	}
	return nil
}

func (s *FeedingService) updatePlanStatus(shopID, operatorID, id uint, status, action, content string) (*model.FeedingPlan, error) {
	plan, err := s.repo.FindPlanByID(shopID, id)
	if err != nil {
		return nil, err
	}
	tx := database.DB.Begin()
	if err := tx.Model(&model.FeedingPlan{}).
		Where("id = ? AND shop_id = ?", id, shopID).
		Update("status", status).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.appendPlanLog(tx, 0, operatorID, action, content); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	plan.Status = status
	return s.repo.FindPlanByID(shopID, id)
}

func decodeFeedingSettings(setting *model.FeedingSetting) (*FeedingSettings, error) {
	defaults := defaultFeedingSettings()
	if setting == nil {
		return &defaults, nil
	}
	if strings.TrimSpace(setting.PricingJSON) != "" {
		if err := json.Unmarshal([]byte(setting.PricingJSON), &defaults.Pricing); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(setting.ItemsJSON) != "" {
		if err := json.Unmarshal([]byte(setting.ItemsJSON), &defaults.Items); err != nil {
			return nil, err
		}
	}
	defaults.Pricing = normalizePricing(defaults.Pricing)
	defaults.Items = normalizeItemTemplates(defaults.Items)
	return &defaults, nil
}

func defaultFeedingSettings() FeedingSettings {
	return FeedingSettings{
		Pricing: FeedingPricingSetting{
			BaseVisitPrice:   50,
			ExtraPetPrice:    20,
			HolidaySurcharge: 20,
		},
		Items: []FeedingItemTemplate{
			{Code: "feed", Name: "添粮", ExtraPrice: 0},
			{Code: "water", Name: "换水", ExtraPrice: 0},
			{Code: "litter", Name: "铲屎", ExtraPrice: 0},
			{Code: "play", Name: "陪玩", ExtraPrice: 0},
			{Code: "medicate", Name: "喂药", ExtraPrice: 20},
			{Code: "ventilate", Name: "开窗通风", ExtraPrice: 0},
			{Code: "video", Name: "视频回访", ExtraPrice: 10},
		},
	}
}

func normalizePricing(pricing FeedingPricingSetting) FeedingPricingSetting {
	pricing.BaseVisitPrice = roundMoney(maxFloat(pricing.BaseVisitPrice, 0))
	pricing.ExtraPetPrice = roundMoney(maxFloat(pricing.ExtraPetPrice, 0))
	pricing.HolidaySurcharge = roundMoney(maxFloat(pricing.HolidaySurcharge, 0))
	return pricing
}

func normalizeItemTemplates(items []FeedingItemTemplate) []FeedingItemTemplate {
	normalized := make([]FeedingItemTemplate, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		code := strings.TrimSpace(item.Code)
		name := strings.TrimSpace(item.Name)
		if code == "" || name == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		normalized = append(normalized, FeedingItemTemplate{
			Code:       code,
			Name:       name,
			ExtraPrice: roundMoney(maxFloat(item.ExtraPrice, 0)),
		})
	}
	if len(normalized) == 0 {
		defaults := defaultFeedingSettings()
		return defaults.Items
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].Code < normalized[j].Code
	})
	return normalized
}

func resolveSelectedTemplates(templates []FeedingItemTemplate, codes []string) []FeedingItemTemplate {
	templateMap := make(map[string]FeedingItemTemplate, len(templates))
	for _, template := range templates {
		templateMap[template.Code] = template
	}
	selected := make([]FeedingItemTemplate, 0, len(codes))
	for _, code := range codes {
		if template, ok := templateMap[code]; ok {
			selected = append(selected, template)
		}
	}
	return selected
}

func normalizeWindowCode(code string) string {
	switch strings.TrimSpace(code) {
	case model.FeedingWindowMorning:
		return model.FeedingWindowMorning
	case model.FeedingWindowAfternoon:
		return model.FeedingWindowAfternoon
	case model.FeedingWindowEvening:
		return model.FeedingWindowEvening
	default:
		return ""
	}
}

func ensureVisitPermission(visit *model.FeedingVisit, role string, operatorID uint) error {
	if visit == nil {
		return errors.New("上门任务不存在")
	}
	if model.HasStaffRoleAtLeast(role, model.StaffRoleManager) {
		return nil
	}
	if visit.StaffID == nil || *visit.StaffID == 0 {
		return errors.New("请先由店长分配执行员工")
	}
	if *visit.StaffID != operatorID {
		return errors.New("只能操作分配给自己的任务")
	}
	return nil
}

func applyVisitItemChecks(tx *gorm.DB, existing []model.FeedingVisitItem, checks []FeedingVisitItemCheck) error {
	checkMap := make(map[uint]bool, len(checks))
	for _, check := range checks {
		checkMap[check.ID] = check.Checked
	}
	for _, item := range existing {
		checked, ok := checkMap[item.ID]
		if !ok {
			continue
		}
		if err := tx.Model(&model.FeedingVisitItem{}).Where("id = ?", item.ID).Update("checked", checked).Error; err != nil {
			return err
		}
	}
	return nil
}

func decodePricingSnapshot(raw string) (*feedingPricingSnapshot, error) {
	var snapshot feedingPricingSnapshot
	if strings.TrimSpace(raw) == "" {
		defaults := defaultFeedingSettings()
		snapshot.Pricing = defaults.Pricing
		return &snapshot, nil
	}
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func decodeSelectedTemplates(raw string) ([]FeedingItemTemplate, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var items []FeedingItemTemplate
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func syncFeedingPlanStatusTx(tx *gorm.DB, planID uint) error {
	var visits []model.FeedingVisit
	if err := tx.Where("feeding_plan_id = ?", planID).Find(&visits).Error; err != nil {
		return err
	}
	if len(visits) == 0 {
		return nil
	}
	status := model.FeedingPlanStatusActive
	doneCount := 0
	cancelledCount := 0
	activeCount := 0
	for _, visit := range visits {
		switch visit.Status {
		case model.FeedingVisitStatusDone:
			doneCount++
		case model.FeedingVisitStatusCancelled:
			cancelledCount++
		case model.FeedingVisitStatusPending, model.FeedingVisitStatusAssigned, model.FeedingVisitStatusInProgress:
			activeCount++
		}
	}
	switch {
	case activeCount > 0:
		status = model.FeedingPlanStatusActive
	case doneCount > 0:
		status = model.FeedingPlanStatusCompleted
	case cancelledCount == len(visits):
		status = model.FeedingPlanStatusCancelled
	default:
		status = model.FeedingPlanStatusPaused
	}
	return tx.Model(&model.FeedingPlan{}).Where("id = ? AND status <> ?", planID, model.FeedingPlanStatusCancelled).Update("status", status).Error
}

func listHolidayDates(shopID uint, startDate, endDate string) ([]string, error) {
	var holidays []model.BoardingHoliday
	if err := database.DB.Where("shop_id = ? AND holiday_date >= ? AND holiday_date <= ?", shopID, startDate, endDate).
		Order("holiday_date ASC").
		Find(&holidays).Error; err != nil {
		return nil, err
	}
	dates := make([]string, 0, len(holidays))
	for _, holiday := range holidays {
		dates = append(dates, holiday.HolidayDate)
	}
	return dates, nil
}

func formatMonthDay(dateText string) string {
	t, err := time.Parse("2006-01-02", dateText)
	if err != nil {
		return dateText
	}
	return fmt.Sprintf("%d月%d日", int(t.Month()), t.Day())
}

func windowLabel(code string) string {
	switch code {
	case model.FeedingWindowMorning:
		return "早间"
	case model.FeedingWindowAfternoon:
		return "午后"
	case model.FeedingWindowEvening:
		return "晚间"
	default:
		return code
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
