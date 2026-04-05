package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
)

type AppointmentService struct {
	apptRepo     *repository.AppointmentRepository
	scheduleRepo *repository.ScheduleRepository
	serviceRepo  *repository.ServiceRepository
	staffRepo    *repository.StaffRepository
	orderRepo    *repository.OrderRepository
}

func NewAppointmentService(
	apptRepo *repository.AppointmentRepository,
	scheduleRepo *repository.ScheduleRepository,
	serviceRepo *repository.ServiceRepository,
	staffRepo *repository.StaffRepository,
	orderRepo *repository.OrderRepository,
) *AppointmentService {
	return &AppointmentService{
		apptRepo:     apptRepo,
		scheduleRepo: scheduleRepo,
		serviceRepo:  serviceRepo,
		staffRepo:    staffRepo,
		orderRepo:    orderRepo,
	}
}

// TimeSlot represents an available time slot.
type TimeSlot struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
}

// StaffSlots represents available slots for a staff member.
type StaffSlots struct {
	Staff *model.Staff `json:"staff"`
	Slots []TimeSlot   `json:"slots"`
}

type timeInterval struct {
	start int
	end   int
}

type AppointmentPetSelection struct {
	PetID      uint   `json:"pet_id"`
	ServiceIDs []uint `json:"service_ids"`
}

type CalendarDaySummary struct {
	Date            string `json:"date"`
	HasAppointments bool   `json:"has_appointments"`
	IsFull          bool   `json:"is_full"`
}

// GetAvailableSlots keeps the legacy single-service path for C-end callers.
func (s *AppointmentService) GetAvailableSlots(shopID uint, date string, serviceID uint) ([]StaffSlots, error) {
	return s.GetAvailableSlotsExcluding(shopID, date, serviceID, 0)
}

func (s *AppointmentService) GetAvailableSlotsExcluding(shopID uint, date string, serviceID uint, excludeAppointmentID uint) ([]StaffSlots, error) {
	svc, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return nil, errors.New("服务不存在")
	}
	return s.GetAvailableSlotsByServicesExcluding(shopID, date, []uint{serviceID}, svc.Duration, excludeAppointmentID)
}

// GetAvailableSlotsByServices calculates available time slots for a service bundle.
func (s *AppointmentService) GetAvailableSlotsByServices(shopID uint, date string, serviceIDs []uint, totalDuration int) ([]StaffSlots, error) {
	return s.GetAvailableSlotsByServicesExcluding(shopID, date, serviceIDs, totalDuration, 0)
}

// GetAvailableSlotsByServicesExcluding calculates available time slots while excluding one appointment from conflicts.
func (s *AppointmentService) GetAvailableSlotsByServicesExcluding(shopID uint, date string, serviceIDs []uint, totalDuration int, excludeAppointmentID uint) ([]StaffSlots, error) {
	serviceIDs = uniqueUintSlice(serviceIDs)
	if len(serviceIDs) == 0 {
		return nil, errors.New("请至少选择一个服务")
	}

	if totalDuration <= 0 {
		for _, serviceID := range serviceIDs {
			svc, err := s.serviceRepo.FindByID(serviceID)
			if err != nil {
				return nil, fmt.Errorf("服务 %d 不存在", serviceID)
			}
			totalDuration += svc.Duration
		}
	}

	staffIDs, err := s.getQualifiedStaffIDs(shopID, serviceIDs)
	if err != nil {
		return nil, err
	}

	var result []StaffSlots
	for _, staffID := range staffIDs {
		staff, err := s.staffRepo.FindByID(staffID)
		if err != nil || staff.Status != 1 {
			continue
		}

		schedules, err := s.scheduleRepo.FindByStaffAndDateRange(staffID, date, date)
		var schedule model.StaffSchedule
		if err != nil || len(schedules) == 0 {
			schedule = model.StaffSchedule{
				StaffID:   staffID,
				Date:      date,
				StartTime: "10:00",
				EndTime:   "22:00",
			}
		} else {
			schedule = schedules[0]
		}
		if schedule.IsDayOff {
			continue
		}

		appts, err := s.apptRepo.FindByStaffAndDate(staffID, date)
		if err != nil {
			continue
		}
		appts = filterAppointmentsByID(appts, excludeAppointmentID)

		slots := s.calculateSlots(schedule, appts, totalDuration)
		if len(slots) > 0 {
			result = append(result, StaffSlots{Staff: staff, Slots: slots})
		}
	}

	return result, nil
}

func (s *AppointmentService) calculateSlots(schedule model.StaffSchedule, appts []model.Appointment, serviceDuration int) []TimeSlot {
	startMin := timeToMinutes(schedule.StartTime)
	endMin := timeToMinutes(schedule.EndTime)
	breakStartMin := timeToMinutes(schedule.BreakStart)
	breakEndMin := timeToMinutes(schedule.BreakEnd)
	maxCapacity := schedule.MaxCapacity
	if maxCapacity <= 0 {
		maxCapacity = 1
	}

	var occupied []timeInterval
	for _, a := range appts {
		occupied = append(occupied, timeInterval{timeToMinutes(a.StartTime), timeToMinutes(a.EndTime)})
	}

	var slots []TimeSlot
	for t := startMin; t+serviceDuration <= endMin; t += 30 {
		slotEnd := t + serviceDuration

		if breakStartMin > 0 && breakEndMin > 0 && t < breakEndMin && slotEnd > breakStartMin {
			continue
		}

		if hasIntervalCapacityConflict(occupied, t, slotEnd, maxCapacity) {
			continue
		}

		slots = append(slots, TimeSlot{
			StartTime: minutesToTime(t),
			EndTime:   minutesToTime(slotEnd),
		})
	}

	return slots
}

func hasIntervalCapacityConflict(occupied []timeInterval, startMin, endMin, maxCapacity int) bool {
	if maxCapacity <= 0 {
		maxCapacity = 1
	}
	for segmentStart := startMin; segmentStart < endMin; segmentStart += 30 {
		segmentEnd := segmentStart + 30
		if segmentEnd > endMin {
			segmentEnd = endMin
		}
		active := 0
		for _, o := range occupied {
			if segmentStart < o.end && segmentEnd > o.start {
				active++
			}
		}
		if active >= maxCapacity {
			return true
		}
	}
	return false
}

func timeToMinutes(t string) int {
	if t == "" {
		return 0
	}
	var h, m int
	fmt.Sscanf(t, "%d:%d", &h, &m)
	return h*60 + m
}

func minutesToTime(m int) string {
	return fmt.Sprintf("%02d:%02d", m/60, m%60)
}

func filterAppointmentsByID(appts []model.Appointment, excludeAppointmentID uint) []model.Appointment {
	if excludeAppointmentID == 0 {
		return appts
	}
	filtered := make([]model.Appointment, 0, len(appts))
	for _, appt := range appts {
		if appt.ID == excludeAppointmentID {
			continue
		}
		filtered = append(filtered, appt)
	}
	return filtered
}

func (s *AppointmentService) GetCalendarSummary(shopID uint, startDate, endDate string) ([]CalendarDaySummary, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("开始日期格式错误")
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("结束日期格式错误")
	}
	if end.Before(start) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	appts, err := s.apptRepo.FindByShopAndDateRange(shopID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	marks, err := s.apptRepo.FindCalendarMarks(shopID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	hasAppointments := make(map[string]bool)
	for _, appt := range appts {
		hasAppointments[appt.Date] = true
	}
	markedDates := make(map[string]bool, len(marks))
	for _, mark := range marks {
		markedDates[mark.Date] = true
	}

	result := make([]CalendarDaySummary, 0)
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		date := current.Format("2006-01-02")
		result = append(result, CalendarDaySummary{
			Date:            date,
			HasAppointments: hasAppointments[date],
			IsFull:          markedDates[date],
		})
	}

	return result, nil
}

func (s *AppointmentService) SetCalendarMark(shopID uint, date string, marked bool) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return errors.New("日期格式错误")
	}
	return s.apptRepo.SetCalendarMark(shopID, date, marked)
}

// Create keeps the legacy single-pet entrypoint for C-end callers.
func (s *AppointmentService) Create(appt *model.Appointment, serviceIDs []uint) error {
	return s.CreateMulti(appt, []AppointmentPetSelection{{
		PetID:      appt.PetID,
		ServiceIDs: serviceIDs,
	}})
}

// CreateMulti creates a multi-pet appointment while preserving legacy flat service snapshots.
func (s *AppointmentService) CreateMulti(appt *model.Appointment, petSelections []AppointmentPetSelection) error {
	payload, err := s.buildMultiPayload(appt, petSelections, 0)
	if err != nil {
		return err
	}

	tx := database.DB.Begin()
	if err := tx.Create(appt).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := s.persistAppointmentRelations(tx, appt.ID, payload); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

type appointmentPayload struct {
	appointmentServices []model.AppointmentService
	appointmentPets     []model.AppointmentPet
	petServiceGroups    [][]model.AppointmentPetService
}

func (s *AppointmentService) buildMultiPayload(appt *model.Appointment, petSelections []AppointmentPetSelection, excludeAppointmentID uint) (*appointmentPayload, error) {
	if len(petSelections) == 0 {
		return nil, errors.New("请至少选择一只宠物")
	}

	var totalDuration int
	var totalAmount float64
	var apptServices []model.AppointmentService
	var apptPets []model.AppointmentPet
	var petServiceGroups [][]model.AppointmentPetService
	seenPets := make(map[uint]struct{}, len(petSelections))

	for idx, selection := range petSelections {
		if selection.PetID == 0 {
			return nil, errors.New("存在未选择的宠物")
		}
		if _, exists := seenPets[selection.PetID]; exists {
			return nil, errors.New("同一只宠物不能重复选择")
		}
		seenPets[selection.PetID] = struct{}{}

		serviceIDs := uniqueUintSlice(selection.ServiceIDs)
		if len(serviceIDs) == 0 {
			return nil, errors.New("每只宠物至少需要选择一个服务")
		}

		petRow := model.AppointmentPet{
			PetID:     selection.PetID,
			SortOrder: idx + 1,
		}
		// 查宠物毛发等级，用于匹配价格规则
		var petFurLevel string
		var pet model.Pet
		if err := database.DB.Select("fur_level").First(&pet, selection.PetID).Error; err == nil {
			petFurLevel = pet.FurLevel
		}

		var petServices []model.AppointmentPetService
		for _, sid := range serviceIDs {
			svc, err := s.serviceRepo.FindActiveByIDAndShop(sid, appt.ShopID)
			if err != nil {
				return nil, fmt.Errorf("服务(ID=%d)不存在或已下架，请刷新页面重新选择", sid)
			}

			// 根据毛发等级匹配价格规则，否则用基础价
			price := svc.BasePrice
			duration := svc.Duration
			if petFurLevel != "" {
				rules, _ := s.serviceRepo.FindPriceRules(sid)
				for _, r := range rules {
					if r.FurLevel == petFurLevel {
						price = r.Price
						if r.Duration > 0 {
							duration = r.Duration
						}
						break
					}
				}
			}

			petRow.TotalDuration += duration
			petRow.TotalAmount += price
			totalDuration += duration
			totalAmount += price

			petServices = append(petServices, model.AppointmentPetService{
				ServiceID:   sid,
				ServiceName: svc.Name,
				Price:       price,
				Duration:    duration,
			})
			apptServices = append(apptServices, model.AppointmentService{
				ServiceID:   sid,
				ServiceName: svc.Name,
				Price:       price,
				Duration:    duration,
			})
		}

		apptPets = append(apptPets, petRow)
		petServiceGroups = append(petServiceGroups, petServices)
	}

	startMin := timeToMinutes(appt.StartTime)
	endMin := startMin + totalDuration
	if appt.EndTime != "" {
		manualEndMin := timeToMinutes(appt.EndTime)
		if manualEndMin <= startMin {
			return nil, errors.New("结束时间必须晚于开始时间")
		}
		if (manualEndMin-startMin)%30 != 0 {
			return nil, errors.New("预约时间必须按30分钟粒度选择")
		}
		endMin = manualEndMin
	}
	appt.EndTime = minutesToTime(endMin)
	appt.TotalAmount = totalAmount
	appt.PetID = petSelections[0].PetID

	if appt.StaffID != nil && *appt.StaffID > 0 {
		if err := s.ensureStaffAvailability(*appt.StaffID, appt.Date, appt.StartTime, appt.EndTime, excludeAppointmentID); err != nil {
			return nil, err
		}
	}

	return &appointmentPayload{
		appointmentServices: apptServices,
		appointmentPets:     apptPets,
		petServiceGroups:    petServiceGroups,
	}, nil
}

func (s *AppointmentService) persistAppointmentRelations(tx *gorm.DB, appointmentID uint, payload *appointmentPayload) error {
	for i := range payload.appointmentPets {
		payload.appointmentPets[i].AppointmentID = appointmentID
		if err := tx.Create(&payload.appointmentPets[i]).Error; err != nil {
			return err
		}

		for j := range payload.petServiceGroups[i] {
			payload.petServiceGroups[i][j].AppointmentPetID = payload.appointmentPets[i].ID
		}
		if len(payload.petServiceGroups[i]) > 0 {
			if err := tx.Create(&payload.petServiceGroups[i]).Error; err != nil {
				return err
			}
		}
	}

	for i := range payload.appointmentServices {
		payload.appointmentServices[i].AppointmentID = appointmentID
	}
	if len(payload.appointmentServices) > 0 {
		if err := tx.Create(&payload.appointmentServices).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *AppointmentService) UpdateMulti(apptID, shopID uint, updates *model.Appointment, petSelections []AppointmentPetSelection) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil || appt.ShopID != shopID {
		return errors.New("预约不存在")
	}
	if appt.Status > 3 {
		return errors.New("当前状态不允许修改预约")
	}

	appt.CustomerID = updates.CustomerID
	appt.StaffID = updates.StaffID
	appt.Date = updates.Date
	appt.StartTime = updates.StartTime
	appt.EndTime = updates.EndTime
	appt.Notes = updates.Notes
	if updates.Source > 0 {
		appt.Source = updates.Source
	}

	payload, err := s.buildMultiPayload(appt, petSelections, appt.ID)
	if err != nil {
		return err
	}

	tx := database.DB.Begin()
	if err := tx.Model(&model.Appointment{}).Where("id = ?", appt.ID).Updates(map[string]interface{}{
		"customer_id":  appt.CustomerID,
		"pet_id":       appt.PetID,
		"staff_id":     appt.StaffID,
		"date":         appt.Date,
		"start_time":   appt.StartTime,
		"end_time":     appt.EndTime,
		"source":       appt.Source,
		"notes":        appt.Notes,
		"total_amount": appt.TotalAmount,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	petSubQuery := tx.Model(&model.AppointmentPet{}).Select("id").Where("appointment_id = ?", appt.ID)
	if err := tx.Where("appointment_pet_id IN (?)", petSubQuery).Delete(&model.AppointmentPetService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("appointment_id = ?", appt.ID).Delete(&model.AppointmentPet{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("appointment_id = ?", appt.ID).Delete(&model.AppointmentService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := s.persistAppointmentRelations(tx, appt.ID, payload); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *AppointmentService) Delete(apptID, shopID uint) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil || appt.ShopID != shopID {
		return errors.New("预约不存在")
	}
	if appt.Status == 2 || appt.Status == 3 || appt.Status == 7 {
		return errors.New("当前状态不允许删除预约")
	}

	orderCount, err := s.orderRepo.CountByAppointment(appt.ID)
	if err != nil {
		return err
	}
	if orderCount > 0 {
		return errors.New("该预约已关联订单，不能删除")
	}

	tx := database.DB.Begin()
	petSubQuery := tx.Model(&model.AppointmentPet{}).Select("id").Where("appointment_id = ?", appt.ID)
	if err := tx.Where("appointment_pet_id IN (?)", petSubQuery).Delete(&model.AppointmentPetService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("appointment_id = ?", appt.ID).Delete(&model.AppointmentPet{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("appointment_id = ?", appt.ID).Delete(&model.AppointmentService{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&model.Appointment{}, appt.ID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *AppointmentService) getQualifiedStaffIDs(shopID uint, serviceIDs []uint) ([]uint, error) {
	var allStaff []model.Staff
	if err := database.DB.Where("shop_id = ? AND status = 1 AND role IN ?", shopID, []string{
		model.StaffRoleStaff,
		model.StaffRoleManager,
		model.StaffRoleAdmin,
	}).Find(&allStaff).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(allStaff))
	for _, st := range allStaff {
		ids = append(ids, st.ID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

func uniqueUintSlice(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (s *AppointmentService) GetByID(id uint) (*model.Appointment, error) {
	return s.apptRepo.FindByID(id)
}

func (s *AppointmentService) UpdateNotes(apptID, shopID uint, notes string) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil || appt.ShopID != shopID {
		return errors.New("预约不存在")
	}
	if appt.Status > 3 {
		return errors.New("当前状态不允许修改预约备注")
	}
	appt.Notes = notes
	return database.DB.Model(&model.Appointment{}).
		Where("id = ? AND shop_id = ?", appt.ID, shopID).
		Update("notes", notes).Error
}

func (s *AppointmentService) ListByDate(shopID uint, date string) ([]model.Appointment, error) {
	return s.apptRepo.FindByShopAndDate(shopID, date)
}

func (s *AppointmentService) ListByDateRange(shopID uint, startDate, endDate string) ([]model.Appointment, error) {
	return s.apptRepo.FindByShopAndDateRange(shopID, startDate, endDate)
}

func (s *AppointmentService) ListByCustomer(customerID uint, page, pageSize int) ([]model.Appointment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.apptRepo.FindByCustomer(customerID, page, pageSize)
}

func (s *AppointmentService) ListPaged(shopID uint, status *int, dateFrom, dateTo string, staffID uint, page, pageSize int) ([]model.Appointment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.apptRepo.FindByShopPaged(shopID, status, dateFrom, dateTo, staffID, page, pageSize)
}

// UpdateStatus handles appointment status transitions.
func (s *AppointmentService) UpdateStatus(id uint, newStatus int, staffNotes, cancelReason, cancelledBy string) error {
	appt, err := s.apptRepo.FindByID(id)
	if err != nil {
		return errors.New("预约不存在")
	}

	valid := false
	switch newStatus {
	case 1: // 已确认 ← 待确认
		valid = appt.Status == 0
	case 2: // 服务中 ← 已确认（兼容历史已到店）
		valid = appt.Status == 1 || appt.Status == 6
	case 3: // 待结算 ← 服务中
		valid = appt.Status == 2
	case 4: // 已取消 ← 待确认/已确认（兼容历史已到店）
		valid = appt.Status == 0 || appt.Status == 1 || appt.Status == 6
	case 5: // 未到店 ← 允许从任意未终止状态标记
		valid = appt.Status != 4 && appt.Status != 5
	}
	if !valid {
		return fmt.Errorf("无法从状态 %d 变更为 %d", appt.Status, newStatus)
	}

	appt.Status = newStatus
	if staffNotes != "" {
		appt.StaffNotes = staffNotes
	}
	if newStatus == 4 {
		appt.CancelReason = cancelReason
		appt.CancelledBy = cancelledBy
	}

	return s.apptRepo.Update(appt)
}

// AssignStaff assigns a staff to an appointment.
func (s *AppointmentService) AssignStaff(apptID, staffID uint) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil {
		return errors.New("预约不存在")
	}

	if err := s.ensureStaffAvailability(staffID, appt.Date, appt.StartTime, appt.EndTime, apptID); err != nil {
		return err
	}

	appt.StaffID = &staffID
	return s.apptRepo.Update(appt)
}

// Reschedule changes the date/time of an appointment.
func (s *AppointmentService) Reschedule(apptID uint, newDate, newStartTime string) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil {
		return errors.New("预约不存在")
	}
	if appt.Status >= 3 {
		return errors.New("已完成/已取消的预约无法改期")
	}

	oldDuration := timeToMinutes(appt.EndTime) - timeToMinutes(appt.StartTime)
	newEndTime := minutesToTime(timeToMinutes(newStartTime) + oldDuration)

	if appt.StaffID != nil && *appt.StaffID > 0 {
		if err := s.ensureStaffAvailability(*appt.StaffID, newDate, newStartTime, newEndTime, apptID); err != nil {
			return err
		}
	}

	appt.Date = newDate
	appt.StartTime = newStartTime
	appt.EndTime = newEndTime
	return s.apptRepo.Update(appt)
}

func (s *AppointmentService) ensureStaffAvailability(staffID uint, date, startTime, endTime string, excludeAppointmentID uint) error {
	schedule, err := s.getScheduleForStaffDate(staffID, date)
	if err != nil {
		return err
	}
	if schedule.IsDayOff {
		return errors.New("该员工当天休息，无法安排预约")
	}

	startMin := timeToMinutes(startTime)
	endMin := timeToMinutes(endTime)
	if endMin <= startMin {
		return errors.New("结束时间必须晚于开始时间")
	}
	if (endMin-startMin)%30 != 0 {
		return errors.New("预约时间必须按30分钟粒度选择")
	}

	scheduleStart := timeToMinutes(schedule.StartTime)
	scheduleEnd := timeToMinutes(schedule.EndTime)
	if startMin < scheduleStart || endMin > scheduleEnd {
		return errors.New("预约时间超出员工排班范围")
	}

	breakStart := timeToMinutes(schedule.BreakStart)
	breakEnd := timeToMinutes(schedule.BreakEnd)
	if breakStart > 0 && breakEnd > 0 && startMin < breakEnd && endMin > breakStart {
		return errors.New("预约时间与员工休息时间冲突")
	}

	appts, err := s.apptRepo.FindByStaffAndDate(staffID, date)
	if err != nil {
		return err
	}
	appts = filterAppointmentsByID(appts, excludeAppointmentID)

	occupied := make([]timeInterval, 0, len(appts))
	for _, appt := range appts {
		occupied = append(occupied, timeInterval{
			start: timeToMinutes(appt.StartTime),
			end:   timeToMinutes(appt.EndTime),
		})
	}

	maxCapacity := schedule.MaxCapacity
	if maxCapacity <= 0 {
		maxCapacity = 1
	}
	if hasIntervalCapacityConflict(occupied, startMin, endMin, maxCapacity) {
		return errors.New("该时段已达到员工接单上限")
	}

	return nil
}

func (s *AppointmentService) getScheduleForStaffDate(staffID uint, date string) (model.StaffSchedule, error) {
	schedules, err := s.scheduleRepo.FindByStaffAndDateRange(staffID, date, date)
	if err == nil && len(schedules) > 0 {
		schedule := schedules[0]
		if schedule.MaxCapacity <= 0 {
			schedule.MaxCapacity = 1
		}
		return schedule, nil
	}

	return model.StaffSchedule{
		StaffID:     staffID,
		Date:        date,
		StartTime:   "10:00",
		EndTime:     "22:00",
		MaxCapacity: 1,
	}, nil
}
