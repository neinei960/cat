package service

import (
	"errors"
	"fmt"
	"sort"

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
}

func NewAppointmentService(
	apptRepo *repository.AppointmentRepository,
	scheduleRepo *repository.ScheduleRepository,
	serviceRepo *repository.ServiceRepository,
	staffRepo *repository.StaffRepository,
) *AppointmentService {
	return &AppointmentService{
		apptRepo:     apptRepo,
		scheduleRepo: scheduleRepo,
		serviceRepo:  serviceRepo,
		staffRepo:    staffRepo,
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

type AppointmentPetSelection struct {
	PetID      uint   `json:"pet_id"`
	ServiceIDs []uint `json:"service_ids"`
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

	type interval struct{ start, end int }
	var occupied []interval
	for _, a := range appts {
		occupied = append(occupied, interval{timeToMinutes(a.StartTime), timeToMinutes(a.EndTime)})
	}

	var slots []TimeSlot
	for t := startMin; t+serviceDuration <= endMin; t += 30 {
		slotEnd := t + serviceDuration

		if breakStartMin > 0 && breakEndMin > 0 && t < breakEndMin && slotEnd > breakStartMin {
			continue
		}

		conflict := false
		for _, o := range occupied {
			if t < o.end && slotEnd > o.start {
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}

		slots = append(slots, TimeSlot{
			StartTime: minutesToTime(t),
			EndTime:   minutesToTime(slotEnd),
		})
	}

	return slots
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
		var petServices []model.AppointmentPetService
		for _, sid := range serviceIDs {
			svc, err := s.serviceRepo.FindByID(sid)
			if err != nil {
				return nil, fmt.Errorf("服务 %d 不存在", sid)
			}

			petRow.TotalDuration += svc.Duration
			petRow.TotalAmount += svc.BasePrice
			totalDuration += svc.Duration
			totalAmount += svc.BasePrice

			petServices = append(petServices, model.AppointmentPetService{
				ServiceID:   sid,
				ServiceName: svc.Name,
				Price:       svc.BasePrice,
				Duration:    svc.Duration,
			})
			apptServices = append(apptServices, model.AppointmentService{
				ServiceID:   sid,
				ServiceName: svc.Name,
				Price:       svc.BasePrice,
				Duration:    svc.Duration,
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
		conflict, err := s.apptRepo.HasConflict(*appt.StaffID, appt.Date, appt.StartTime, appt.EndTime, excludeAppointmentID)
		if err != nil {
			return nil, err
		}
		if conflict {
			return nil, errors.New("该时段技师已有预约，存在时间冲突")
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
	if appt.Status > 1 {
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

func (s *AppointmentService) getQualifiedStaffIDs(shopID uint, serviceIDs []uint) ([]uint, error) {
	var allStaff []model.Staff
	if err := database.DB.Where("shop_id = ? AND status = 1 AND role = 'staff'", shopID).Find(&allStaff).Error; err != nil {
		return nil, err
	}
	if len(allStaff) == 0 {
		return nil, nil
	}

	activeStaff := make(map[uint]struct{}, len(allStaff))
	for _, st := range allStaff {
		activeStaff[st.ID] = struct{}{}
	}

	var candidate map[uint]struct{}
	for _, serviceID := range serviceIDs {
		var staffServices []model.StaffService
		if err := database.DB.Where("service_id = ?", serviceID).Find(&staffServices).Error; err != nil {
			return nil, err
		}

		current := make(map[uint]struct{})
		if len(staffServices) == 0 {
			for id := range activeStaff {
				current[id] = struct{}{}
			}
		} else {
			for _, ss := range staffServices {
				if _, ok := activeStaff[ss.StaffID]; ok {
					current[ss.StaffID] = struct{}{}
				}
			}
		}

		if candidate == nil {
			candidate = current
			continue
		}

		next := make(map[uint]struct{})
		for id := range candidate {
			if _, ok := current[id]; ok {
				next[id] = struct{}{}
			}
		}
		candidate = next
	}

	ids := make([]uint, 0, len(candidate))
	for id := range candidate {
		ids = append(ids, id)
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

func (s *AppointmentService) ListPaged(shopID uint, status *int, page, pageSize int) ([]model.Appointment, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.apptRepo.FindByShopPaged(shopID, status, page, pageSize)
}

// UpdateStatus handles appointment status transitions.
func (s *AppointmentService) UpdateStatus(id uint, newStatus int, staffNotes, cancelReason, cancelledBy string) error {
	appt, err := s.apptRepo.FindByID(id)
	if err != nil {
		return errors.New("预约不存在")
	}

	valid := false
	switch newStatus {
	case 1:
		valid = appt.Status == 0
	case 2:
		valid = appt.Status == 1
	case 3:
		valid = appt.Status == 2
	case 4:
		valid = appt.Status == 0 || appt.Status == 1
	case 5:
		valid = appt.Status == 1
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

	conflict, err := s.apptRepo.HasConflict(staffID, appt.Date, appt.StartTime, appt.EndTime, apptID)
	if err != nil {
		return err
	}
	if conflict {
		return errors.New("该时段技师已有预约，存在时间冲突")
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
		conflict, err := s.apptRepo.HasConflict(*appt.StaffID, newDate, newStartTime, newEndTime, apptID)
		if err != nil {
			return err
		}
		if conflict {
			return errors.New("新时段存在冲突")
		}
	}

	appt.Date = newDate
	appt.StartTime = newStartTime
	appt.EndTime = newEndTime
	return s.apptRepo.Update(appt)
}
