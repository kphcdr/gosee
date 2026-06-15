package model

import (
	"encoding/json"
	"time"
)

// 通知通道类型
const (
	ChannelTypeFeishu   = "feishu"
	ChannelTypeTelegram = "telegram"
	ChannelTypeEmail    = "email"
)

// NotificationChannel 通知通道（PRD 8.7）。
// Config 用 json.RawMessage 存 JSON 对象，前端传/取均为对象，无需手动序列化。
// 飞书：{"webhook":"https://...", "secret":"可选签名密钥"}
type NotificationChannel struct {
	ID        int64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string          `gorm:"size:100;not null" json:"name"`
	Type      string          `gorm:"size:50;not null" json:"type"`
	Config    json.RawMessage `gorm:"type:text" json:"config"`
	Enabled   int8            `gorm:"not null;default:1" json:"enabled"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (NotificationChannel) TableName() string { return "notification_channels" }
