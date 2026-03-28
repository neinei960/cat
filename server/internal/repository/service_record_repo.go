package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type ServiceRecordRepository struct{}

func NewServiceRecordRepository() *ServiceRecordRepository {
	return &ServiceRecordRepository{}
}

func (r *ServiceRecordRepository) Create(record *model.ServiceRecord) error {
	return database.DB.Create(record).Error
}

func (r *ServiceRecordRepository) FindByAppointment(appointmentID uint) ([]model.ServiceRecord, error) {
	var records []model.ServiceRecord
	err := database.DB.Where("appointment_id = ?", appointmentID).
		Preload("Staff").Order("id ASC").Find(&records).Error
	return records, err
}

func (r *ServiceRecordRepository) FindByPet(petID uint, limit int) ([]model.ServiceRecord, error) {
	var records []model.ServiceRecord
	err := database.DB.Where("pet_id = ?", petID).
		Preload("Staff").Preload("Appointment").
		Order("id DESC").Limit(limit).Find(&records).Error
	return records, err
}

func (r *ServiceRecordRepository) Update(record *model.ServiceRecord) error {
	return database.DB.Save(record).Error
}
