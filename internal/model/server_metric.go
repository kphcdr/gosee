package model

import "time"

// ServerMetric 单次采集的服务器指标汇总（对齐 PRD 8.3）
type ServerMetric struct {
	ID                int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ServerID          int64     `gorm:"not null;index:idx_server_collected,priority:1" json:"server_id"`
	Hostname          string    `gorm:"size:255" json:"hostname"`
	OS                string    `gorm:"size:255" json:"os"`
	CPUUsage          float64   `gorm:"type:decimal(5,2)" json:"cpu_usage"`
	CPUCores          int       `json:"cpu_cores"`
	MemoryTotalMB     int64     `json:"memory_total_mb"`
	MemoryUsedMB      int64     `json:"memory_used_mb"`
	MemoryAvailableMB int64     `json:"memory_available_mb"`
	MemoryUsage       float64   `gorm:"type:decimal(5,2)" json:"memory_usage"`
	Load1             float64   `gorm:"type:decimal(10,2)" json:"load_1"`
	Load5             float64   `gorm:"type:decimal(10,2)" json:"load_5"`
	Load15            float64   `gorm:"type:decimal(10,2)" json:"load_15"`
	DiskMaxUsage      float64   `gorm:"type:decimal(5,2)" json:"disk_max_usage"`
	UptimeSeconds     int64     `json:"uptime_seconds"`
	RawJSON           string    `gorm:"type:text" json:"-"` // 原始采集 JSON，默认不返回
	CollectedAt       time.Time `gorm:"not null;index:idx_server_collected,priority:2;index:idx_collected_at" json:"collected_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// ServerDisk 单次采集的分区明细（对齐 PRD 8.4）
type ServerDisk struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	MetricID       int64     `gorm:"not null;index" json:"metric_id"`
	ServerID       int64     `gorm:"not null;index" json:"server_id"`
	Filesystem     string    `gorm:"size:255" json:"filesystem"`
	MountPoint     string    `gorm:"size:255" json:"mount_point"`
	SizeBytes      int64     `json:"size_bytes"`
	UsedBytes      int64     `json:"used_bytes"`
	AvailableBytes int64     `json:"available_bytes"`
	UsagePercent   float64   `gorm:"type:decimal(5,2)" json:"usage_percent"`
	CreatedAt      time.Time `json:"created_at"`
}
