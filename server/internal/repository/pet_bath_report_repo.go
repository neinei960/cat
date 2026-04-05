package repository

import (
	"time"

	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
	"gorm.io/gorm"
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
		Order("sort_order DESC, COALESCE(bath_date, created_at) DESC, created_at DESC, id DESC").
		Find(&reports).Error
	return reports, err
}

func (r *PetBathReportRepository) GetNextSortOrder(shopID, petID uint) (int, error) {
	var maxSort int
	err := database.DB.Model(&model.PetBathReport{}).
		Where("shop_id = ? AND pet_id = ?", shopID, petID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxSort).Error
	if err != nil {
		return 0, err
	}
	return maxSort + 1, nil
}

func (r *PetBathReportRepository) UpdateBathDate(shopID, petID, reportID uint, bathDate *time.Time) error {
	return database.DB.Model(&model.PetBathReport{}).
		Where("shop_id = ? AND pet_id = ? AND id = ?", shopID, petID, reportID).
		Update("bath_date", bathDate).Error
}

func (r *PetBathReportRepository) Reorder(shopID, petID uint, reportIDs []uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.PetBathReport{}).
			Where("shop_id = ? AND pet_id = ? AND id IN ?", shopID, petID, reportIDs).
			Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(reportIDs)) {
			return gorm.ErrRecordNotFound
		}

		total := len(reportIDs)
		for index, reportID := range reportIDs {
			if err := tx.Model(&model.PetBathReport{}).
				Where("shop_id = ? AND pet_id = ? AND id = ?", shopID, petID, reportID).
				Update("sort_order", total-index).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PetBathReportRepository) Delete(shopID, petID, reportID uint) error {
	return database.DB.Where("shop_id = ? AND pet_id = ? AND id = ?", shopID, petID, reportID).
		Delete(&model.PetBathReport{}).Error
}
