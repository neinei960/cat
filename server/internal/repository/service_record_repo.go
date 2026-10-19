package repository

import (
	"strconv"
	"strings"

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
		Preload("Pet").Preload("Staff").Order("id ASC").Find(&records).Error
	if err != nil {
		return records, err
	}
	attachOrderItemSummaries(appointmentID, records)
	return records, nil
}

func attachOrderItemSummaries(appointmentID uint, records []model.ServiceRecord) {
	if len(records) == 0 {
		return
	}

	var orders []model.Order
	if err := database.DB.Where("appointment_id = ?", appointmentID).
		Preload("Pet").Preload("Items").Order("id ASC").Find(&orders).Error; err != nil {
		return
	}
	if len(orders) == 0 {
		return
	}

	for i := range records {
		records[i].OrderItemSummary = summarizeOrderItemsForRecord(records[i], orders)
	}
}

func summarizeOrderItemsForRecord(record model.ServiceRecord, orders []model.Order) string {
	var petName string
	if record.Pet != nil {
		petName = strings.TrimSpace(record.Pet.Name)
	}

	var matched []string
	var fallback []string
	for _, order := range orders {
		for _, item := range order.Items {
			if item.ItemType == 1 || item.ItemType == 3 {
				label := formatOrderItemLabel(item)
				fallback = append(fallback, label)
				if order.PetID != nil && *order.PetID == record.PetID {
					matched = append(matched, label)
					continue
				}
				if petName != "" && strings.Contains(item.Name, petName) {
					matched = append(matched, label)
				}
			}
		}
	}

	if len(matched) > 0 {
		return strings.Join(uniqueStrings(matched), "、")
	}
	return strings.Join(uniqueStrings(fallback), "、")
}

func formatOrderItemLabel(item model.OrderItem) string {
	name := strings.TrimSpace(item.Name)
	if name == "" {
		return ""
	}
	if item.Quantity > 1 {
		return name + " x" + strconv.Itoa(item.Quantity)
	}
	return name
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	uniq := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		uniq = append(uniq, item)
	}
	return uniq
}

func (r *ServiceRecordRepository) FindByPet(petID uint, limit int) ([]model.ServiceRecord, error) {
	var records []model.ServiceRecord
	err := database.DB.Where("pet_id = ?", petID).
		Preload("Staff").Preload("Appointment").
		Order("id DESC").Limit(limit).Find(&records).Error
	return records, err
}

func (r *ServiceRecordRepository) FindByID(shopID uint, id uint) (*model.ServiceRecord, error) {
	var record model.ServiceRecord
	err := database.DB.Where("id = ? AND shop_id = ?", id, shopID).
		Preload("Pet").Preload("Staff").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *ServiceRecordRepository) Update(record *model.ServiceRecord) error {
	return database.DB.Save(record).Error
}
