package model

import "time"

// 通知发送状态
const (
	NotifyStatusSuccess = "success"
	NotifyStatusFailed  = "failed"
)

// AlertNotification 通知发送记录（PRD 8.8）
type AlertNotification struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AlertEventID int64      `gorm:"not null;index" json:"alert_event_id"`
	ChannelID    int64      `gorm:"not null" json:"channel_id"`
	Status       string     `gorm:"size:20;not null" json:"status"`
	Response     string     `gorm:"type:text" json:"response"`
	SentAt       *time.Time `json:"sent_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (AlertNotification) TableName() string { return "alert_notifications" }
