package retention

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gosee/internal/config"
	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

// setupDB 建立内存 SQLite（单连接，与现网 SQLite 行为一致）并迁移所需表。
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 sqlDB 失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(1) // 内存库单连接，避免多连接各自独立的 :memory:
	if err := db.AutoMigrate(&model.ServerMetric{}, &model.ServerDisk{}, &model.AlertEvent{}, &model.AlertNotification{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	return db
}

func count(t *testing.T, db *gorm.DB, dst interface{}, where string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	if err := db.Model(dst).Where(where, args...).Count(&n).Error; err != nil {
		t.Fatalf("count 失败: %v", err)
	}
	return n
}

// TestRun_PurgesExpiredKeepsActive 验证：过期指标(含磁盘)被级联删、近期数据保留、
// 已结束告警事件被删但 firing 事件保留、过期通知记录被删。
func TestRun_PurgesExpiredKeepsActive(t *testing.T) {
	utils.Logger = zap.NewNop() // Run 内部用 utils.Logger 记日志，赋值避免 nil panic
	db := setupDB(t)
	metricRepo := repository.NewServerMetricRepository(db)
	eventRepo := repository.NewAlertEventRepository(db)
	notifyRepo := repository.NewAlertNotificationRepository(db)

	now := time.Now()
	old := now.AddDate(0, 0, -91)   // 91 天前
	day31 := now.AddDate(0, 0, -31) // 31 天前（超过 30 天保留期）
	day1 := now.AddDate(0, 0, -1)   // 1 天前（保留期内）

	// 旧指标(31天前) + 磁盘 —— 应删
	oldMetric := &model.ServerMetric{CPUUsage: 50, CollectedAt: day31}
	if err := metricRepo.CreateMetric(1, oldMetric, []model.ServerDisk{{Filesystem: "/dev/sda1", MountPoint: "/", UsagePercent: 40}}); err != nil {
		t.Fatalf("插入旧指标失败: %v", err)
	}
	oldMetricID := oldMetric.ID

	// 新指标(1天前) + 磁盘 —— 应留
	freshMetric := &model.ServerMetric{CPUUsage: 60, CollectedAt: day1}
	if err := metricRepo.CreateMetric(1, freshMetric, []model.ServerDisk{{Filesystem: "/dev/sda2", MountPoint: "/data", UsagePercent: 30}}); err != nil {
		t.Fatalf("插入新指标失败: %v", err)
	}
	freshMetricID := freshMetric.ID

	// 旧 recovered 事件(91天前 updated) + 通知 —— 应删
	oldEvent := &model.AlertEvent{
		ServerID: 1, AlertRuleID: 1, Metric: model.MetricCPUUsage, Level: model.AlertLevelWarning,
		Status: model.EventStatusRecovered, FirstTriggeredAt: old, LastTriggeredAt: old,
	}
	if err := eventRepo.Create(oldEvent); err != nil {
		t.Fatalf("插入旧事件失败: %v", err)
	}
	if err := db.Exec("UPDATE alert_events SET updated_at = ? WHERE id = ?", old, oldEvent.ID).Error; err != nil {
		t.Fatalf("回溯事件 updated_at 失败: %v", err)
	}
	oldNotify := &model.AlertNotification{AlertEventID: oldEvent.ID, ChannelID: 1, Status: model.NotifyStatusSuccess}
	if err := notifyRepo.Create(oldNotify); err != nil {
		t.Fatalf("插入旧通知失败: %v", err)
	}
	if err := db.Exec("UPDATE alert_notifications SET created_at = ? WHERE id = ?", old, oldNotify.ID).Error; err != nil {
		t.Fatalf("回溯通知 created_at 失败: %v", err)
	}

	// firing 事件(91天前 updated，但 status=firing) —— 必须保留
	firingEvent := &model.AlertEvent{
		ServerID: 1, AlertRuleID: 1, Metric: model.MetricMemoryUsage, Level: model.AlertLevelCritical,
		Status: model.EventStatusFiring, FirstTriggeredAt: old, LastTriggeredAt: old,
	}
	if err := eventRepo.Create(firingEvent); err != nil {
		t.Fatalf("插入 firing 事件失败: %v", err)
	}
	if err := db.Exec("UPDATE alert_events SET updated_at = ? WHERE id = ?", old, firingEvent.ID).Error; err != nil {
		t.Fatalf("回溯 firing 事件 updated_at 失败: %v", err)
	}

	// 执行一次清理：指标保留 30 天 / 事件+通知保留 90 天
	cfg := &config.RetentionConfig{Enabled: true, MetricsDays: 30, AlertEventsDays: 90, NotificationsDays: 90, BatchSize: 100}
	NewService(metricRepo, eventRepo, notifyRepo, cfg).Run()

	if n := count(t, db, &model.ServerMetric{}, "id = ?", oldMetricID); n != 0 {
		t.Errorf("旧指标(31天前)应被删除，仍剩 %d 条", n)
	}
	if n := count(t, db, &model.ServerDisk{}, "metric_id = ?", oldMetricID); n != 0 {
		t.Errorf("旧指标的磁盘明细应级联删除，仍剩 %d 条", n)
	}
	if n := count(t, db, &model.ServerMetric{}, "id = ?", freshMetricID); n != 1 {
		t.Errorf("新指标(1天前)应保留，剩 %d 条", n)
	}
	if n := count(t, db, &model.ServerDisk{}, "metric_id = ?", freshMetricID); n != 1 {
		t.Errorf("新指标的磁盘明细应保留，剩 %d 条", n)
	}
	if n := count(t, db, &model.AlertEvent{}, "id = ?", oldEvent.ID); n != 0 {
		t.Errorf("旧 recovered 事件(91天前)应被删除，仍剩 %d 条", n)
	}
	if n := count(t, db, &model.AlertEvent{}, "id = ?", firingEvent.ID); n != 1 {
		t.Errorf("firing 事件必须保留(即便超过保留期)，剩 %d 条", n)
	}
	if n := count(t, db, &model.AlertNotification{}, "id = ?", oldNotify.ID); n != 0 {
		t.Errorf("旧通知记录(91天前)应被删除，仍剩 %d 条", n)
	}
}

// TestRun_DisabledNoOp 验证 Enabled=false 时不动任何数据。
func TestRun_DisabledNoOp(t *testing.T) {
	utils.Logger = zap.NewNop()
	db := setupDB(t)
	metricRepo := repository.NewServerMetricRepository(db)
	eventRepo := repository.NewAlertEventRepository(db)
	notifyRepo := repository.NewAlertNotificationRepository(db)

	old := time.Now().AddDate(0, 0, -31)
	oldMetric := &model.ServerMetric{CPUUsage: 50, CollectedAt: old}
	if err := metricRepo.CreateMetric(1, oldMetric, []model.ServerDisk{{MountPoint: "/", UsagePercent: 1}}); err != nil {
		t.Fatalf("插入失败: %v", err)
	}

	cfg := &config.RetentionConfig{Enabled: false, MetricsDays: 30, BatchSize: 100}
	NewService(metricRepo, eventRepo, notifyRepo, cfg).Run()

	if n := count(t, db, &model.ServerMetric{}, "id = ?", oldMetric.ID); n != 1 {
		t.Errorf("清理关闭时不应删除任何指标，剩 %d 条", n)
	}
}

// TestRun_BatchLoop 验证分批删除能处理超过 batch_size 的数据量（循环到清空）。
func TestRun_BatchLoop(t *testing.T) {
	utils.Logger = zap.NewNop()
	db := setupDB(t)
	metricRepo := repository.NewServerMetricRepository(db)
	eventRepo := repository.NewAlertEventRepository(db)
	notifyRepo := repository.NewAlertNotificationRepository(db)

	old := time.Now().AddDate(0, 0, -31)
	// 插入 25 条过期指标，batch_size=10 需 3 轮才能删完
	for i := 0; i < 25; i++ {
		if err := metricRepo.CreateMetric(1, &model.ServerMetric{CPUUsage: float64(i), CollectedAt: old}, nil); err != nil {
			t.Fatalf("插入失败: %v", err)
		}
	}

	cfg := &config.RetentionConfig{Enabled: true, MetricsDays: 30, BatchSize: 10}
	NewService(metricRepo, eventRepo, notifyRepo, cfg).Run()

	if n := count(t, db, &model.ServerMetric{}, "collected_at < ?", old.Add(time.Second)); n != 0 {
		t.Errorf("分批删除后应全部清空过期指标，仍剩 %d 条", n)
	}
}
