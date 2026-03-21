package service

import (
	"errors"
	"fmt"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
	"github.com/neinei960/cat/server/pkg/database"
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

// TimeSlot represents an available time slot
type TimeSlot struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
}

// StaffSlots represents available slots for a staff member
type StaffSlots struct {
	Staff *model.Staff `json:"staff"`
	Slots []TimeSlot   `json:"slots"`
}

// GetAvailableSlots calculates available time slots for a given date and service
// Algorithm:
// 1. Find qualified staff (via staff_services)
// 2. Get each staff's schedule for the date
// 3. Get existing appointments
// 4. Generate 30-min slots, remove occupied ones
// 5. Check remaining duration >= service duration
func (s *AppointmentService) GetAvailableSlots(shopID uint, date string, serviceID uint) ([]StaffSlots, error) {
	// Get service info for duration
	svc, err := s.serviceRepo.FindByID(serviceID)
	if err != nil {
		return nil, errors.New("服务不存在")
	}

	// Find staff who can do this service
	var staffServices []model.StaffService
	database.DB.Where("service_id = ?", serviceID).Find(&staffServices)

	// If no staff explicitly assigned, fallback to all active staff with role=staff
	var staffIDs []uint
	if len(staffServices) > 0 {
		for _, ss := range staffServices {
			staffIDs = append(staffIDs, ss.StaffID)
		}
	} else {
		var allStaff []model.Staff
		database.DB.Where("shop_id = ? AND status = 1 AND role = 'staff'", shopID).Find(&allStaff)
		for _, st := range allStaff {
			staffIDs = append(staffIDs, st.ID)
		}
	}

	var result []StaffSlots
	for _, staffID := range staffIDs {
		staff, err := s.staffRepo.FindByID(staffID)
		if err != nil || staff.Status != 1 {
			continue
		}

		// Get schedule for date, fallback to default hours if not set
		schedules, err := s.scheduleRepo.FindByStaffAndDateRange(staffID, date, date)
		var schedule model.StaffSchedule
		if err != nil || len(schedules) == 0 {
			// No schedule set — use default business hours
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

		// Get existing appointments
		appts, err := s.apptRepo.FindByStaffAndDate(staffID, date)
		if err != nil {
			continue
		}

		// Generate available slots
		slots := s.calculateSlots(schedule, appts, svc.Duration)
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

	// Build occupied intervals from existing appointments
	type interval struct{ start, end int }
	var occupied []interval
	for _, a := range appts {
		occupied = append(occupied, interval{timeToMinutes(a.StartTime), timeToMinutes(a.EndTime)})
	}

	var slots []TimeSlot
	// Iterate in 30-minute increments
	for t := startMin; t+serviceDuration <= endMin; t += 30 {
		slotEnd := t + serviceDuration

		// Skip if overlaps with break
		if breakStartMin > 0 && breakEndMin > 0 {
			if t < breakEndMin && slotEnd > breakStartMin {
				continue
			}
		}

		// Check conflict with existing appointments
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

// Create appointment with conflict detection
func (s *AppointmentService) Create(appt *model.Appointment, serviceIDs []uint) error {
	// Calculate total duration and amount
	var totalDuration int
	var totalAmount float64
	var apptServices []model.AppointmentService

	for _, sid := range serviceIDs {
		svc, err := s.serviceRepo.FindByID(sid)
		if err != nil {
			return fmt.Errorf("服务 %d 不存在", sid)
		}
		totalDuration += svc.Duration
		totalAmount += svc.BasePrice
		apptServices = append(apptServices, model.AppointmentService{
			ServiceID:   sid,
			ServiceName: svc.Name,
			Price:       svc.BasePrice,
			Duration:    svc.Duration,
		})
	}

	// Calculate end time
	startMin := timeToMinutes(appt.StartTime)
	appt.EndTime = minutesToTime(startMin + totalDuration)
	appt.TotalAmount = totalAmount

	// Check conflict if staff assigned
	if appt.StaffID != nil && *appt.StaffID > 0 {
		conflict, err := s.apptRepo.HasConflict(*appt.StaffID, appt.Date, appt.StartTime, appt.EndTime, 0)
		if err != nil {
			return err
		}
		if conflict {
			return errors.New("该时段技师已有预约，存在时间冲突")
		}
	}

	// Create in transaction
	tx := database.DB.Begin()
	if err := tx.Create(appt).Error; err != nil {
		tx.Rollback()
		return err
	}

	for i := range apptServices {
		apptServices[i].AppointmentID = appt.ID
	}
	if err := tx.Create(&apptServices).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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

// UpdateStatus handles appointment status transitions
func (s *AppointmentService) UpdateStatus(id uint, newStatus int, staffNotes, cancelReason, cancelledBy string) error {
	appt, err := s.apptRepo.FindByID(id)
	if err != nil {
		return errors.New("预约不存在")
	}

	// Validate status transition
	valid := false
	switch newStatus {
	case 1: // confirm
		valid = appt.Status == 0
	case 2: // in progress
		valid = appt.Status == 1
	case 3: // complete
		valid = appt.Status == 2
	case 4: // cancel
		valid = appt.Status == 0 || appt.Status == 1
	case 5: // no-show
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

// AssignStaff assigns a staff to an appointment
func (s *AppointmentService) AssignStaff(apptID, staffID uint) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil {
		return errors.New("预约不存在")
	}

	// Check conflict
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

// Reschedule changes the date/time of an appointment
func (s *AppointmentService) Reschedule(apptID uint, newDate, newStartTime string) error {
	appt, err := s.apptRepo.FindByID(apptID)
	if err != nil {
		return errors.New("预约不存在")
	}
	if appt.Status >= 3 {
		return errors.New("已完成/已取消的预约无法改期")
	}

	// Calculate new end time (keep same duration)
	oldDuration := timeToMinutes(appt.EndTime) - timeToMinutes(appt.StartTime)
	newEndTime := minutesToTime(timeToMinutes(newStartTime) + oldDuration)

	// Check conflict
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
