package database

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gosee/internal/model"
)

func TestMigrateAlertEventStateConvertsAckedAndDeduplicates(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:alert_migration?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层数据库失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&model.AlertEvent{}); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
	}

	old := time.Now().Add(-time.Minute)
	latest := time.Now()
	events := []model.AlertEvent{
		{ServerID: 1, AlertRuleID: 2, Metric: model.MetricCPUUsage, Level: model.AlertLevelWarning, Status: model.EventStatusFiring, FirstTriggeredAt: old, LastTriggeredAt: old},
		{ServerID: 1, AlertRuleID: 2, Metric: model.MetricCPUUsage, Level: model.AlertLevelWarning, Status: "acked", FirstTriggeredAt: old, LastTriggeredAt: latest},
		{ServerID: 3, AlertRuleID: 4, Metric: model.MetricMemoryUsage, Level: model.AlertLevelCritical, Status: model.EventStatusFiring, FirstTriggeredAt: old, LastTriggeredAt: latest},
	}
	if err := db.Create(&events).Error; err != nil {
		t.Fatalf("写入旧版事件失败: %v", err)
	}

	if err := migrateAlertEventState(db); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	if err := migrateAlertEventState(db); err != nil {
		t.Fatalf("重复执行迁移应保持幂等: %v", err)
	}

	var group []model.AlertEvent
	if err := db.Where("server_id = ? AND alert_rule_id = ?", 1, 2).Order("id ASC").Find(&group).Error; err != nil {
		t.Fatalf("读取迁移结果失败: %v", err)
	}
	if len(group) != 2 {
		t.Fatalf("迁移不应删除历史事件，实际 %d 条", len(group))
	}
	if group[0].Status != model.EventStatusClosed || group[0].ActiveKey != nil {
		t.Fatalf("较旧重复事件应关闭并释放活动键: status=%s key=%v", group[0].Status, group[0].ActiveKey)
	}
	if group[1].Status != model.EventStatusFiring || group[1].ActiveKey == nil || group[1].AckedAt == nil {
		t.Fatalf("最新 acked 事件应迁为已确认的活动事件: status=%s key=%v acked_at=%v", group[1].Status, group[1].ActiveKey, group[1].AckedAt)
	}

	var activeCount int64
	if err := db.Model(&model.AlertEvent{}).Where("active_key IS NOT NULL").Count(&activeCount).Error; err != nil {
		t.Fatalf("统计活动告警失败: %v", err)
	}
	if activeCount != 2 {
		t.Fatalf("两个服务器/规则组合应各保留一条活动告警，实际 %d", activeCount)
	}
}
