package repository

import (
	"gorm.io/gorm"

	"gosee/internal/model"
)

type NotificationChannelRepository struct {
	db *gorm.DB
}

func NewNotificationChannelRepository(db *gorm.DB) *NotificationChannelRepository {
	return &NotificationChannelRepository{db: db}
}

func (r *NotificationChannelRepository) List() ([]model.NotificationChannel, error) {
	var chs []model.NotificationChannel
	err := r.db.Order("id ASC").Find(&chs).Error
	return chs, err
}

// ListEnabled 启用的通道（告警发送时用）
func (r *NotificationChannelRepository) ListEnabled() ([]model.NotificationChannel, error) {
	var chs []model.NotificationChannel
	err := r.db.Where("enabled = 1").Order("id ASC").Find(&chs).Error
	return chs, err
}

func (r *NotificationChannelRepository) FindByID(id int64) (*model.NotificationChannel, error) {
	var ch model.NotificationChannel
	err := r.db.First(&ch, id).Error
	return &ch, err
}

func (r *NotificationChannelRepository) Create(ch *model.NotificationChannel) error {
	return r.db.Create(ch).Error
}

func (r *NotificationChannelRepository) Update(ch *model.NotificationChannel) error {
	return r.db.Save(ch).Error
}

func (r *NotificationChannelRepository) Delete(id int64) error {
	return r.db.Delete(&model.NotificationChannel{}, id).Error
}
