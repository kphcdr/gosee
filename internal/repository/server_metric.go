package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"gosee/internal/model"
)

type ServerMetricRepository struct {
	db *gorm.DB
}

func NewServerMetricRepository(db *gorm.DB) *ServerMetricRepository {
	return &ServerMetricRepository{db: db}
}

// CreateMetric 在事务中保存指标汇总与磁盘明细
func (r *ServerMetricRepository) CreateMetric(serverID int64, m *model.ServerMetric, disks []model.ServerDisk) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		m.ServerID = serverID
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		for i := range disks {
			disks[i].MetricID = m.ID
			disks[i].ServerID = serverID
		}
		if len(disks) > 0 {
			if err := tx.CreateInBatches(disks, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Latest 服务器最新一条指标
func (r *ServerMetricRepository) Latest(serverID int64) (*model.ServerMetric, error) {
	var m model.ServerMetric
	err := r.db.Where("server_id = ?", serverID).Order("collected_at DESC").First(&m).Error
	return &m, err
}

// Trend 趋势查询：取 collected_at >= since 的最近 limit 条，按时间正序返回便于绘图
func (r *ServerMetricRepository) Trend(serverID int64, since time.Time, limit int) ([]model.ServerMetric, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	var list []model.ServerMetric
	err := r.db.Where("server_id = ? AND collected_at >= ?", serverID, since).
		Order("collected_at DESC").Limit(limit).Find(&list).Error
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list, nil
}

// RecentN 最近 n 条采集记录（按时间倒序），供告警连续触发判断
func (r *ServerMetricRepository) RecentN(serverID int64, n int) ([]model.ServerMetric, error) {
	if n <= 0 {
		n = 1
	}
	var list []model.ServerMetric
	err := r.db.Where("server_id = ?", serverID).Order("collected_at DESC").Limit(n).Find(&list).Error
	return list, err
}

// DisksByMetric 某次采集的磁盘明细，按使用率降序
func (r *ServerMetricRepository) DisksByMetric(metricID int64) ([]model.ServerDisk, error) {
	var disks []model.ServerDisk
	err := r.db.Where("metric_id = ?", metricID).Order("usage_percent DESC").Find(&disks).Error
	return disks, err
}

// LatestDisks 最新一次采集的磁盘明细。若无采集记录，返回空切片而非错误。
func (r *ServerMetricRepository) LatestDisks(serverID int64) ([]model.ServerDisk, error) {
	latest, err := r.Latest(serverID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.ServerDisk{}, nil
		}
		return nil, err
	}
	return r.DisksByMetric(latest.ID)
}

// LatestMetricView 每台服务器最新指标 + 名称/主机（仪表盘 Top5 用）
type LatestMetricView struct {
	ServerID     int64     `gorm:"column:server_id"`
	Name         string    `gorm:"column:name"`
	Host         string    `gorm:"column:host"`
	CPUUsage     float64   `gorm:"column:cpu_usage"`
	MemoryUsage  float64   `gorm:"column:memory_usage"`
	DiskMaxUsage float64   `gorm:"column:disk_max_usage"`
	CollectedAt  time.Time `gorm:"column:collected_at"`
}

// LatestMetricsOfEnabled 取所有启用服务器的最新一条指标（含 name/host）
func (r *ServerMetricRepository) LatestMetricsOfEnabled() ([]LatestMetricView, error) {
	return r.LatestMetricsOfEnabledByGroup(nil)
}

// LatestMetricsOfEnabledByGroup 取指定分组中启用服务器的最新指标；groupID 为空时查询全部。
func (r *ServerMetricRepository) LatestMetricsOfEnabledByGroup(groupID *int64) ([]LatestMetricView, error) {
	var rows []LatestMetricView
	tx := r.db.Table("servers").
		Select(`servers.id AS server_id, servers.name, servers.host,
			m.cpu_usage, m.memory_usage, m.disk_max_usage, m.collected_at`).
		Joins("LEFT JOIN server_metrics m ON m.id = (SELECT id FROM server_metrics WHERE server_id = servers.id ORDER BY collected_at DESC LIMIT 1)").
		Where("servers.enabled = 1")
	if groupID != nil {
		tx = tx.Where("servers.group_id = ?", *groupID)
	}
	err := tx.Scan(&rows).Error
	return rows, err
}

// DeleteMetricsBefore 删除 collected_at 早于 t 的指标记录，并在同一事务中级联删除关联的磁盘明细。
// 分批处理（每批 batch 条），避免单事务过大长时间持锁阻塞采集。返回累计删除的指标条数。
func (r *ServerMetricRepository) DeleteMetricsBefore(t time.Time, batch int) (int64, error) {
	if batch <= 0 {
		batch = 1000
	}
	var total int64
	for {
		var ids []int64
		if err := r.db.Model(&model.ServerMetric{}).
			Where("collected_at < ?", t).
			Limit(batch).Pluck("id", &ids).Error; err != nil {
			return total, err
		}
		if len(ids) == 0 {
			break
		}
		if err := r.db.Transaction(func(tx *gorm.DB) error {
			// 先删子表 server_disks，避免孤儿记录
			if err := tx.Where("metric_id IN ?", ids).Delete(&model.ServerDisk{}).Error; err != nil {
				return err
			}
			// 再删父表 server_metrics
			return tx.Where("id IN ?", ids).Delete(&model.ServerMetric{}).Error
		}); err != nil {
			return total, err
		}
		total += int64(len(ids))
		if len(ids) < batch {
			break // 已无更多数据
		}
	}
	return total, nil
}
