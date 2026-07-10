package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gosee/internal/config"
	"gosee/internal/model"
	"gosee/internal/utils"
)

// Init 初始化数据库连接并执行 AutoMigrate
func Init(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	switch cfg.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), gormCfg)
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), gormCfg)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.Driver == "sqlite" {
		// SQLite 单写多读，限制连接数避免锁冲突
		sqlDB.SetMaxOpenConns(1)
	} else {
		if cfg.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		}
		if cfg.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		}
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.ServerGroup{},
		&model.Server{},
		&model.ServerMetric{},
		&model.ServerDisk{},
		&model.AlertRule{},
		&model.AlertEvent{},
		&model.NotificationChannel{},
		&model.AlertNotification{},
	); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}
	if err := migrateAlertEventState(db); err != nil {
		return nil, fmt.Errorf("告警事件状态迁移失败: %w", err)
	}
	if utils.Logger != nil {
		utils.Logger.Info("数据库迁移完成", zap.String("driver", cfg.Driver))
	}
	return db, nil
}

// SeedAdmin 首次启动时若无管理员则创建默认账号
func SeedAdmin(db *gorm.DB, adminCfg *config.AdminConfig) error {
	var count int64
	if err := db.Model(&model.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hashed, err := utils.HashPassword(adminCfg.Password)
	if err != nil {
		return fmt.Errorf("默认管理员密码哈希失败: %w", err)
	}
	user := model.User{
		Username: adminCfg.Username,
		Password: hashed,
		Nickname: "Administrator",
		Status:   1,
	}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("创建默认管理员失败: %w", err)
	}
	if utils.Logger != nil {
		utils.Logger.Info("已创建默认管理员账号", zap.String("username", adminCfg.Username))
	}
	return nil
}

// SeedAlertRules 首次启动时若无告警规则则插入默认规则（PRD 6.4）
func SeedAlertRules(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.AlertRule{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	global := model.ScopeTypeGlobal
	rules := []model.AlertRule{
		{Name: "CPU 使用率过高", ScopeType: global, Metric: model.MetricCPUUsage, Operator: ">", Threshold: 90, DurationTimes: 3, Level: model.AlertLevelWarning, Enabled: 1, NotifyIntervalMinutes: 60},
		{Name: "内存使用率过高", ScopeType: global, Metric: model.MetricMemoryUsage, Operator: ">", Threshold: 90, DurationTimes: 3, Level: model.AlertLevelWarning, Enabled: 1, NotifyIntervalMinutes: 60},
		{Name: "磁盘使用率告急", ScopeType: global, Metric: model.MetricDiskUsage, Operator: ">", Threshold: 85, DurationTimes: 1, Level: model.AlertLevelCritical, Enabled: 1, NotifyIntervalMinutes: 60},
		// load5 阈值=0 表示动态阈值（cpu_cores*2），评估时计算
		{Name: "系统负载过高", ScopeType: global, Metric: model.MetricLoad5, Operator: ">", Threshold: 0, DurationTimes: 3, Level: model.AlertLevelWarning, Enabled: 1, NotifyIntervalMinutes: 60},
		{Name: "SSH 连接失败", ScopeType: global, Metric: model.MetricSSHFail, Operator: ">=", Threshold: 3, DurationTimes: 3, Level: model.AlertLevelCritical, Enabled: 1, NotifyIntervalMinutes: 60},
	}
	if err := db.Create(&rules).Error; err != nil {
		return fmt.Errorf("创建默认告警规则失败: %w", err)
	}
	if utils.Logger != nil {
		utils.Logger.Info("已创建默认告警规则", zap.Int("count", len(rules)))
	}
	return nil
}
