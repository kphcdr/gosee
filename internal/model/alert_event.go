package model

import "time"

// 告警事件状态
const (
	EventStatusFiring    = "firing"
	EventStatusRecovered = "recovered"
	EventStatusAcked     = "acked"
	EventStatusClosed    = "closed"
)

// AlertEvent 告警事件（PRD 8.6）。
// server_name / rule_name 非数据库字段，由 service 层 JOIN 填充供前端展示。
type AlertEvent struct {
	ID               int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ServerID         int64      `gorm:"not null;index:idx_server_status,priority:1" json:"server_id"`
	AlertRuleID      int64      `gorm:"not null;index:idx_rule_status,priority:1" json:"alert_rule_id"`
	Metric           string     `gorm:"size:50;not null" json:"metric"`
	CurrentValue     *float64   `gorm:"type:decimal(10,2)" json:"current_value"`
	Threshold        *float64   `gorm:"type:decimal(10,2)" json:"threshold"`
	Level            string     `gorm:"size:20;not null" json:"level"`
	Status           string     `gorm:"size:20;not null;default:firing;index:idx_server_status,priority:2;index:idx_rule_status,priority:2" json:"status"`
	FirstTriggeredAt time.Time  `gorm:"not null" json:"first_triggered_at"`
	LastTriggeredAt  time.Time  `gorm:"not null" json:"last_triggered_at"`
	RecoveredAt      *time.Time `json:"recovered_at"`
	LastNotifiedAt   *time.Time `json:"-"` // 防重复通知用，阶段六 Notifier 读取
	NotifyCount      int        `gorm:"not null;default:0" json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

func (AlertEvent) TableName() string { return "alert_events" }
