package repository

import (
	"fmt"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
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
	err := r.withRelations().First(&appt, id).Error
	if err == nil {
		r.normalizeAppointment(&appt)
	}
	return &appt, err
}

func (r *AppointmentRepository) FindByShopAndDate(shopID uint, date string) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := r.withRelations().
		Where("shop_id = ? AND date = ? AND status IN (0,1,2,3)", shopID, date).
		Order("start_time ASC").Find(&appts).Error
	if err == nil {
		r.normalizeAppointments(appts)
	}
	return appts, err
}

func (r *AppointmentRepository) FindByShopAndDateRange(shopID uint, startDate, endDate string) ([]model.Appointment, error) {
	var appts []model.Appointment
	err := r.withRelations().
		Where("shop_id = ? AND date >= ? AND date <= ? AND status IN (0,1,2,3)", shopID, startDate, endDate).
		Order("date ASC, start_time ASC").Find(&appts).Error
	if err == nil {
		r.normalizeAppointments(appts)
	}
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
	err := r.withRelations().
		Where("customer_id = ?", customerID).
		Order("date DESC, start_time DESC").Offset(offset).Limit(pageSize).Find(&appts).Error
	if err == nil {
		r.normalizeAppointments(appts)
	}
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
	query := r.withRelations().Where("shop_id = ?", shopID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	err := query.Order("date DESC, start_time DESC").Offset(offset).Limit(pageSize).Find(&appts).Error
	if err == nil {
		r.normalizeAppointments(appts)
	}
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

func (r *AppointmentRepository) CreatePets(pets []model.AppointmentPet) error {
	return database.DB.Create(&pets).Error
}

func (r *AppointmentRepository) CreatePetServices(services []model.AppointmentPetService) error {
	return database.DB.Create(&services).Error
}

func (r *AppointmentRepository) DeleteServicesByAppointment(appointmentID uint) error {
	return database.DB.Where("appointment_id = ?", appointmentID).Delete(&model.AppointmentService{}).Error
}

func (r *AppointmentRepository) withRelations() *gorm.DB {
	return database.DB.Preload("Customer").
		Preload("Pet").
		Preload("Staff").
		Preload("Services").
		Preload("Pets", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Pets.Pet").
		Preload("Pets.Services")
}

func (r *AppointmentRepository) normalizeAppointments(appts []model.Appointment) {
	for i := range appts {
		r.normalizeAppointment(&appts[i])
	}
}

func (r *AppointmentRepository) normalizeAppointment(appt *model.Appointment) {
	if appt == nil {
		return
	}

	if len(appt.Pets) == 0 && appt.PetID > 0 {
		fallback := model.AppointmentPet{
			AppointmentID: appt.ID,
			PetID:         appt.PetID,
			SortOrder:     1,
			TotalAmount:   appt.TotalAmount,
			Pet:           appt.Pet,
		}
		for _, svc := range appt.Services {
			fallback.TotalDuration += svc.Duration
			fallback.Services = append(fallback.Services, model.AppointmentPetService{
				ServiceID:   svc.ServiceID,
				ServiceName: svc.ServiceName,
				Price:       svc.Price,
				Duration:    svc.Duration,
			})
		}
		appt.Pets = []model.AppointmentPet{fallback}
	}

	if appt.Pet == nil && len(appt.Pets) > 0 {
		appt.Pet = appt.Pets[0].Pet
	}

	names := make([]string, 0, len(appt.Pets))
	for _, petItem := range appt.Pets {
		if petItem.Pet != nil && petItem.Pet.Name != "" {
			names = append(names, petItem.Pet.Name)
		}
	}
	appt.PetCount = len(appt.Pets)
	switch len(names) {
	case 0:
		if appt.PetCount > 1 {
			appt.PetSummary = "共" + itoa(appt.PetCount) + "只"
		} else {
			appt.PetSummary = "-"
		}
	case 1:
		appt.PetSummary = names[0]
	default:
		appt.PetSummary = names[0] + "等" + itoa(len(names)) + "只"
	}
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
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
