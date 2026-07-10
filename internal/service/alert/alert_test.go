package alert

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gosee/internal/model"
	"gosee/internal/repository"
	"gosee/internal/utils"
)

func setupAlertTest(t *testing.T) (*gorm.DB, *Service, *repository.AlertEventRepository, *repository.ServerMetricRepository) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层数据库失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(
		&model.Server{},
		&model.ServerMetric{},
		&model.ServerDisk{},
		&model.AlertRule{},
		&model.AlertEvent{},
	); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
	}

	utils.Logger = zap.NewNop()
	eventRepo := repository.NewAlertEventRepository(db)
	metricRepo := repository.NewServerMetricRepository(db)
	svc := NewService(
		repository.NewAlertRuleRepository(db),
		eventRepo,
		metricRepo,
		repository.NewServerRepository(db),
	)
	return db, svc, eventRepo, metricRepo
}

func TestEvaluateRuleRequiresFullDurationWindow(t *testing.T) {
	db, svc, _, metricRepo := setupAlertTest(t)
	server := &model.Server{ID: 1, Name: "server-1"}
	rule := &model.AlertRule{
		ID:            1,
		Name:          "CPU high",
		Metric:        model.MetricCPUUsage,
		Operator:      ">",
		Threshold:     90,
		DurationTimes: 3,
		Level:         model.AlertLevelWarning,
	}

	for i := 1; i <= 3; i++ {
		metric := &model.ServerMetric{
			CPUUsage:    95,
			CollectedAt: time.Now().Add(time.Duration(i) * time.Second),
		}
		if err := metricRepo.CreateMetric(server.ID, metric, nil); err != nil {
			t.Fatalf("写入第 %d 条指标失败: %v", i, err)
		}
		svc.evaluateRule(server, rule, metric, nil, 0)

		var count int64
		if err := db.Model(&model.AlertEvent{}).Count(&count).Error; err != nil {
			t.Fatalf("统计告警失败: %v", err)
		}
		want := int64(0)
		if i == rule.DurationTimes {
			want = 1
		}
		if count != want {
			t.Fatalf("第 %d 次连续超限后告警数=%d，期望 %d", i, count, want)
		}
	}
}

func TestAcknowledgeKeepsEventActive(t *testing.T) {
	db, svc, eventRepo, _ := setupAlertTest(t)
	server := &model.Server{ID: 7, Name: "server-7"}
	rule := &model.AlertRule{ID: 9, Name: "CPU high", Metric: model.MetricCPUUsage, Level: model.AlertLevelWarning}

	svc.fireEvent(server, rule, 95, 90)
	event, err := eventRepo.FindFiring(rule.ID, server.ID)
	if err != nil {
		t.Fatalf("查询活动告警失败: %v", err)
	}
	if err := svc.AckEvent(event.ID, 42); err != nil {
		t.Fatalf("确认告警失败: %v", err)
	}

	var acked model.AlertEvent
	if err := db.First(&acked, event.ID).Error; err != nil {
		t.Fatalf("读取确认后的告警失败: %v", err)
	}
	if acked.Status != model.EventStatusFiring {
		t.Fatalf("确认不应终止告警，status=%s", acked.Status)
	}
	if acked.AckedAt == nil || acked.AckedBy == nil || *acked.AckedBy != 42 {
		t.Fatalf("确认信息未正确保存: acked_at=%v acked_by=%v", acked.AckedAt, acked.AckedBy)
	}
	if acked.ActiveKey == nil {
		t.Fatal("确认后活动唯一键不应释放")
	}

	svc.fireEvent(server, rule, 96, 90)
	var count int64
	if err := db.Model(&model.AlertEvent{}).
		Where("server_id = ? AND alert_rule_id = ?", server.ID, rule.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("统计告警失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("确认后持续超限应刷新原事件，实际共有 %d 条", count)
	}
	refreshed, err := eventRepo.FindFiring(rule.ID, server.ID)
	if err != nil {
		t.Fatalf("重新查询活动告警失败: %v", err)
	}
	if refreshed.ID != event.ID || refreshed.AckedAt == nil {
		t.Fatalf("持续告警未保留原事件及确认信息: id=%d acked_at=%v", refreshed.ID, refreshed.AckedAt)
	}
}

func TestUpsertFiringDeduplicatesConcurrentCreates(t *testing.T) {
	db, _, eventRepo, _ := setupAlertTest(t)
	const workers = 32
	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(value float64) {
			defer wg.Done()
			now := time.Now()
			threshold := 90.0
			_, err := eventRepo.UpsertFiring(&model.AlertEvent{
				ServerID:         11,
				AlertRuleID:      13,
				Metric:           model.MetricCPUUsage,
				CurrentValue:     &value,
				Threshold:        &threshold,
				Level:            model.AlertLevelWarning,
				FirstTriggeredAt: now,
				LastTriggeredAt:  now,
			})
			if err != nil {
				errCh <- err
			}
		}(float64(91 + i))
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("并发 upsert 失败: %v", err)
	}

	var total, active int64
	if err := db.Model(&model.AlertEvent{}).Count(&total).Error; err != nil {
		t.Fatalf("统计事件失败: %v", err)
	}
	if err := db.Model(&model.AlertEvent{}).
		Where("status = ? AND active_key IS NOT NULL", model.EventStatusFiring).
		Count(&active).Error; err != nil {
		t.Fatalf("统计活动事件失败: %v", err)
	}
	if total != 1 || active != 1 {
		t.Fatalf("并发触发应只有一条活动告警，total=%d active=%d", total, active)
	}
}
