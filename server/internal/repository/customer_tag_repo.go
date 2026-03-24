package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type CustomerTagRepository struct{}

func NewCustomerTagRepository() *CustomerTagRepository {
	return &CustomerTagRepository{}
}

func (r *CustomerTagRepository) Create(tag *model.CustomerTag) error {
	return database.DB.Create(tag).Error
}

func (r *CustomerTagRepository) List(shopID uint) ([]model.CustomerTag, error) {
	var tags []model.CustomerTag
	err := database.DB.Model(&model.CustomerTag{}).
		Select("customer_tags.*, COUNT(DISTINCT customer_tag_relations.customer_id) as relation_count").
		Joins("LEFT JOIN customer_tag_relations ON customer_tag_relations.tag_id = customer_tags.id").
		Joins("LEFT JOIN customers ON customers.id = customer_tag_relations.customer_id AND customers.deleted_at IS NULL").
		Where("customer_tags.shop_id = ?", shopID).
		Group("customer_tags.id").
		Order("sort_order ASC, id ASC").
		Find(&tags).Error
	return tags, err
}

func (r *CustomerTagRepository) FindByID(id uint, shopID uint) (*model.CustomerTag, error) {
	var tag model.CustomerTag
	err := database.DB.Where("id = ? AND shop_id = ?", id, shopID).First(&tag).Error
	return &tag, err
}

func (r *CustomerTagRepository) Update(tag *model.CustomerTag) error {
	return database.DB.Save(tag).Error
}

func (r *CustomerTagRepository) Delete(id uint, shopID uint) error {
	return database.DB.Where("id = ? AND shop_id = ?", id, shopID).Delete(&model.CustomerTag{}).Error
}
