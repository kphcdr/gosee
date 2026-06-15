package repository

import (
	"gorm.io/gorm"

	"gosee/internal/model"
)

type AlertNotificationRepository struct {
	db *gorm.DB
}

func NewAlertNotificationRepository(db *gorm.DB) *AlertNotificationRepository {
	return &AlertNotificationRepository{db: db}
}

// Create 记录一次通知发送结果
func (r *AlertNotificationRepository) Create(n *model.AlertNotification) error {
	return r.db.Create(n).Error
}
