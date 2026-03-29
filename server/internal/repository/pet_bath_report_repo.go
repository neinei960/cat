package repository

import (
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type PetBathReportRepository struct{}

func NewPetBathReportRepository() *PetBathReportRepository {
	return &PetBathReportRepository{}
}

func (r *PetBathReportRepository) Create(report *model.PetBathReport) error {
	return database.DB.Create(report).Error
}

func (r *PetBathReportRepository) FindByPet(shopID, petID uint) ([]model.PetBathReport, error) {
	var reports []model.PetBathReport
	err := database.DB.Where("shop_id = ? AND pet_id = ?", shopID, petID).
		Order("COALESCE(bath_date, created_at) DESC, created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *PetBathReportRepository) UpdateBathDate(shopID, petID, reportID uint, bathDate *time.Time) error {
	return database.DB.Model(&model.PetBathReport{}).
		Where("shop_id = ? AND pet_id = ? AND id = ?", shopID, petID, reportID).
		Update("bath_date", bathDate).Error
}

func (r *PetBathReportRepository) Delete(shopID, petID, reportID uint) error {
	return database.DB.Where("shop_id = ? AND pet_id = ? AND id = ?", shopID, petID, reportID).
		Delete(&model.PetBathReport{}).Error
}
