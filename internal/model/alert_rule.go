package model

import "time"

// scope_type 取值
const (
	ScopeTypeGlobal = "global"
	ScopeTypeGroup  = "group"
	ScopeTypeServer = "server"
)

// 告警等级
const (
	AlertLevelInfo     = "info"
	AlertLevelWarning  = "warning"
	AlertLevelCritical = "critical"
)

// 告警指标类型（metric 列的取值）
const (
	MetricCPUUsage    = "cpu_usage"
	MetricMemoryUsage = "memory_usage"
	MetricDiskUsage   = "disk_usage"
	MetricLoad1       = "load_1"
	MetricLoad5       = "load_5"
	MetricLoad15      = "load_15"
	MetricSSHFail     = "ssh_fail"
)

// AlertRule 告警规则（PRD 8.5）。
// DB 列保持 PRD（metric / duration_times），JSON tag 对齐前端（metric_type / duration_count）。
type AlertRule struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                  string    `gorm:"size:100;not null" json:"name"`
	ScopeType             string    `gorm:"size:20;not null;default:global;index:idx_scope,priority:1" json:"scope_type"`
	ScopeID               *int64    `gorm:"column:scope_id;index:idx_scope,priority:2" json:"scope_id"`
	Metric                string    `gorm:"size:50;not null" json:"metric_type"`
	Operator              string    `gorm:"size:10;not null" json:"operator"`
	Threshold             float64   `gorm:"type:decimal(10,2);not null" json:"threshold"`
	DurationTimes         int       `gorm:"column:duration_times;not null;default:1" json:"duration_count"`
	Level                 string    `gorm:"size:20;not null;default:warning" json:"level"`
	Enabled               int8      `gorm:"not null;default:1;index" json:"enabled"`
	NotifyIntervalMinutes int       `gorm:"not null;default:60" json:"-"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (AlertRule) TableName() string { return "alert_rules" }
