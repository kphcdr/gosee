package repository

import (
	"time"

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

// DeleteBefore 删除 created_at 早于 t 的通知发送记录。分批处理，返回累计删除条数。
func (r *AlertNotificationRepository) DeleteBefore(t time.Time, batch int) (int64, error) {
	if batch <= 0 {
		batch = 1000
	}
	var total int64
	for {
		var ids []int64
		if err := r.db.Model(&model.AlertNotification{}).
			Where("created_at < ?", t).
			Limit(batch).Pluck("id", &ids).Error; err != nil {
			return total, err
		}
		if len(ids) == 0 {
			break
		}
		res := r.db.Where("id IN ?", ids).Delete(&model.AlertNotification{})
		if res.Error != nil {
			return total, res.Error
		}
		total += res.RowsAffected
		if len(ids) < batch {
			break
		}
	}
	return total, nil
}
