package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type AppointmentRepository struct{}

func NewAppointmentRepository() *AppointmentRepository {
	return &AppointmentRepository{}
}

func (r *AppointmentRepository) Create(appt *model.Appointment) error {
	return database.DB.Create(appt).Error
}

func (r *AppointmentRepository) FindByID(id uint) (*model.Appointment, error) {
	var appt model.Appointment
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Services").
		First(&appt, id).Error
	return &appt, err
}

func (r *AppointmentRepository) FindByShopAndDate(shopID uint, date string) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Services").
		Where("shop_id = ? AND date = ? AND status IN (0,1,2,3)", shopID, date).
		Order("start_time ASC").Find(&appts).Error
	return appts, err
}

func (r *AppointmentRepository) FindByShopAndDateRange(shopID uint, startDate, endDate string) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := database.DB.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Services").
		Where("shop_id = ? AND date >= ? AND date <= ? AND status IN (0,1,2,3)", shopID, startDate, endDate).
		Order("date ASC, start_time ASC").Find(&appts).Error
	return appts, err
}

func (r *AppointmentRepository) FindByStaffAndDate(staffID uint, date string) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := database.DB.Where("staff_id = ? AND date = ? AND status IN (0,1,2)", staffID, date).
		Order("start_time ASC").Find(&appts).Error
	return appts, err
}

func (r *AppointmentRepository) FindByCustomer(customerID uint, page, pageSize int) ([]model.Appointment, int64, error) {
	var appts []model.Appointment
	var total int64
	db := database.DB.Model(&model.Appointment{}).Where("customer_id = ?", customerID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Pet").Preload("Staff").Preload("Services").
		Order("date DESC, start_time DESC").Offset(offset).Limit(pageSize).Find(&appts).Error
	return appts, total, err
}

func (r *AppointmentRepository) FindByShopPaged(shopID uint, status *int, page, pageSize int) ([]model.Appointment, int64, error) {
	var appts []model.Appointment
	var total int64
	db := database.DB.Model(&model.Appointment{}).Where("shop_id = ?", shopID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Preload("Customer").Preload("Pet").Preload("Staff").Preload("Services").
		Order("date DESC, start_time DESC").Offset(offset).Limit(pageSize).Find(&appts).Error
	return appts, total, err
}

func (r *AppointmentRepository) Update(appt *model.Appointment) error {
	return database.DB.Save(appt).Error
}

func (r *AppointmentRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Appointment{}, id).Error
}

// Appointment services

func (r *AppointmentRepository) CreateServices(services []model.AppointmentService) error {
	return database.DB.Create(&services).Error
}

func (r *AppointmentRepository) DeleteServicesByAppointment(appointmentID uint) error {
	return database.DB.Where("appointment_id = ?", appointmentID).Delete(&model.AppointmentService{}).Error
}

// Check conflict: does the staff have overlapping appointments?
func (r *AppointmentRepository) HasConflict(staffID uint, date, startTime, endTime string, excludeID uint) (bool, error) {
	var count int64
	db := database.DB.Model(&model.Appointment{}).
		Where("staff_id = ? AND date = ? AND status IN (0,1,2)", staffID, date).
		Where("start_time < ? AND end_time > ?", endTime, startTime)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}
