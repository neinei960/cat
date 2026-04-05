package service

import (
	"errors"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/internal/repository"
)

type StaffService struct {
	staffRepo    *repository.StaffRepository
	scheduleRepo *repository.ScheduleRepository
	serviceRepo  *repository.ServiceRepository
}

func NewStaffService(staffRepo *repository.StaffRepository, scheduleRepo *repository.ScheduleRepository, serviceRepo *repository.ServiceRepository) *StaffService {
	return &StaffService{staffRepo: staffRepo, scheduleRepo: scheduleRepo, serviceRepo: serviceRepo}
}

func (s *StaffService) Create(staff *model.Staff) error {
	return s.CreateWithPassword(staff, "123456")
}

func (s *StaffService) CreateWithPassword(staff *model.Staff, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if staff.SortOrder <= 0 {
		nextSort, err := s.staffRepo.NextSortOrder(staff.ShopID)
		if err != nil {
			return err
		}
		staff.SortOrder = nextSort
	}
	staff.PasswordHash = hash
	return s.staffRepo.Create(staff)
}

func (s *StaffService) GetByID(id uint) (*model.Staff, error) {
	return s.staffRepo.FindByID(id)
}

func (s *StaffService) List(shopID uint, page, pageSize int) ([]model.Staff, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.staffRepo.FindByShopID(shopID, page, pageSize)
}

func (s *StaffService) Update(staff *model.Staff) error {
	return s.staffRepo.Update(staff)
}

func (s *StaffService) Delete(id uint) error {
	return s.staffRepo.Delete(id)
}

func (s *StaffService) Reorder(shopID uint, orderedIDs []uint) error {
	return s.staffRepo.BatchUpdateSortOrders(shopID, orderedIDs)
}

func (s *StaffService) ResetPassword(id uint, newPassword string) error {
	staff, err := s.staffRepo.FindByID(id)
	if err != nil {
		return errors.New("员工不存在")
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	staff.PasswordHash = hash
	return s.staffRepo.Update(staff)
}

// Schedule

func (s *StaffService) SetSchedule(schedule *model.StaffSchedule) error {
	return s.scheduleRepo.Upsert(schedule)
}

func (s *StaffService) BatchSetSchedule(schedules []model.StaffSchedule) error {
	return s.scheduleRepo.BatchUpsert(schedules)
}

func (s *StaffService) GetSchedule(staffID uint, startDate, endDate string) ([]model.StaffSchedule, error) {
	return s.scheduleRepo.FindByStaffAndDateRange(staffID, startDate, endDate)
}

// Staff services

func (s *StaffService) SetServices(staffID uint, serviceIDs []uint) error {
	return s.serviceRepo.SetStaffServices(staffID, serviceIDs)
}

func (s *StaffService) GetServices(staffID uint) ([]model.Service, error) {
	return s.serviceRepo.FindStaffServices(staffID)
}
