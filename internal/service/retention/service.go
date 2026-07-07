package retention

import (
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"gosee/internal/config"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

// Service 数据保留与自动清理编排：按保留天数分批删除过期时序数据，防止数据库无限增长。
type Service struct {
	metricRepo *repository.ServerMetricRepository
	eventRepo  *repository.AlertEventRepository
	notifyRepo *repository.AlertNotificationRepository
	cfg        *config.RetentionConfig
	running    int32 // atomic，防止清理任务重叠
}

func NewService(
	metricRepo *repository.ServerMetricRepository,
	eventRepo *repository.AlertEventRepository,
	notifyRepo *repository.AlertNotificationRepository,
	cfg *config.RetentionConfig,
) *Service {
	return &Service{
		metricRepo: metricRepo,
		eventRepo:  eventRepo,
		notifyRepo: notifyRepo,
		cfg:        cfg,
	}
}

// Run 执行一次完整清理：metrics → notifications → events（先子后父，避免孤儿记录）。
// 自带防重叠保护，被 cron 定时调用；各表 days<=0 时跳过该表。
func (s *Service) Run() {
	if s.cfg == nil || !s.cfg.Enabled {
		return
	}
	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		utils.Logger.Warn("上一轮清理仍在进行，跳过本次触发")
		return
	}
	defer atomic.StoreInt32(&s.running, 0)

	batch := s.cfg.BatchSize
	now := time.Now()

	var metricsDeleted, eventsDeleted, notificationsDeleted int64

	if s.cfg.MetricsDays > 0 {
		cut := now.AddDate(0, 0, -s.cfg.MetricsDays)
		n, err := s.metricRepo.DeleteMetricsBefore(cut, batch)
		metricsDeleted = n
		if err != nil {
			utils.Logger.Error("清理监控指标失败", zap.Error(err), zap.Time("before", cut))
		}
	}
	// 先删 notifications 再删 events，避免 alert_notifications 残留指向已删事件的孤儿记录
	if s.cfg.NotificationsDays > 0 {
		cut := now.AddDate(0, 0, -s.cfg.NotificationsDays)
		n, err := s.notifyRepo.DeleteBefore(cut, batch)
		notificationsDeleted = n
		if err != nil {
			utils.Logger.Error("清理通知记录失败", zap.Error(err), zap.Time("before", cut))
		}
	}
	if s.cfg.AlertEventsDays > 0 {
		cut := now.AddDate(0, 0, -s.cfg.AlertEventsDays)
		n, err := s.eventRepo.DeleteEndedBefore(cut, batch)
		eventsDeleted = n
		if err != nil {
			utils.Logger.Error("清理告警事件失败", zap.Error(err), zap.Time("before", cut))
		}
	}

	utils.Logger.Info("数据清理完成",
		zap.Int64("metrics", metricsDeleted),
		zap.Int64("events", eventsDeleted),
		zap.Int64("notifications", notificationsDeleted),
	)
}
