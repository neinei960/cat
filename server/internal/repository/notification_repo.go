package repository

import (
	"github.com/neinei960/cat/server/internal/model"
	"github.com/neinei960/cat/server/pkg/database"
)

type NotificationRepository struct{}

func NewNotificationRepository() *NotificationRepository {
	return &NotificationRepository{}
}

func (r *NotificationRepository) Create(log *model.NotificationLog) error {
	return database.DB.Create(log).Error
}

func (r *NotificationRepository) FindByShop(shopID uint, page, pageSize int) ([]model.NotificationLog, int64, error) {
	var logs []model.NotificationLog
	var total int64
	db := database.DB.Model(&model.NotificationLog{}).Where("shop_id = ?", shopID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
