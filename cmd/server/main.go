package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gosee/internal/api"
	"gosee/internal/api/handler"
	"gosee/internal/config"
	"gosee/internal/database"
	"gosee/internal/repository"
	"gosee/internal/scheduler"
	"gosee/internal/service/alert"
	"gosee/internal/service/auth"
	"gosee/internal/service/collector"
	"gosee/internal/service/dashboard"
	"gosee/internal/service/notifier"
	"gosee/internal/service/retention"
	"gosee/internal/service/server"
	"gosee/internal/service/server_group"
	"gosee/internal/utils"
)

func main() {
	// 1. 配置
	configPath := os.Getenv("GOSEE_CONFIG")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 日志
	if err := utils.InitLogger(cfg.Log.Level, cfg.Log.Dir); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = utils.Logger.Sync() }()

	// 3. 数据库 + 迁移 + 默认管理员
	db, err := database.Init(&cfg.Database)
	if err != nil {
		utils.Logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	if err := database.SeedAdmin(db, &cfg.Admin); err != nil {
		utils.Logger.Fatal("初始化管理员失败", zap.Error(err))
	}
	if err := database.SeedAlertRules(db); err != nil {
		utils.Logger.Fatal("初始化告警规则失败", zap.Error(err))
	}

	// 4. 装配依赖：repository → service → handler
	userRepo := repository.NewUserRepository(db)
	serverRepo := repository.NewServerRepository(db)
	groupRepo := repository.NewServerGroupRepository(db)
	metricRepo := repository.NewServerMetricRepository(db)

	// 采集超时（从配置解析，异常时回退默认值）
	connectTimeout, _ := time.ParseDuration(cfg.Collector.SSHConnectTimeout)
	if connectTimeout <= 0 {
		connectTimeout = 5 * time.Second
	}
	commandTimeout, _ := time.ParseDuration(cfg.Collector.SSHCommandTimeout)
	if commandTimeout <= 0 {
		commandTimeout = 15 * time.Second
	}
	authSvc := auth.NewService(userRepo, &cfg.JWT)
	serverSvc := server.NewService(serverRepo, groupRepo, &cfg.Security, connectTimeout)
	collectorSvc := collector.NewService(serverSvc, metricRepo, commandTimeout, cfg.Collector.MaxRetries)
	groupSvc := server_group.NewService(groupRepo)

	// 告警：规则 + 事件 + 评估器
	alertRuleRepo := repository.NewAlertRuleRepository(db)
	alertEventRepo := repository.NewAlertEventRepository(db)
	alertSvc := alert.NewService(alertRuleRepo, alertEventRepo, metricRepo, serverRepo)
	collectorSvc.SetHook(alertSvc) // 采集完成后自动触发告警评估

	// 通知：飞书等通道，接入告警事件
	notifyChannelRepo := repository.NewNotificationChannelRepository(db)
	notifyRepo := repository.NewAlertNotificationRepository(db)
	notifierSvc := notifier.NewService(notifyChannelRepo, notifyRepo, alertEventRepo, serverRepo, alertRuleRepo)
	alertSvc.SetNotifierHook(notifierSvc) // 告警触发/恢复后发送通知

	// 仪表盘聚合
	dashboardSvc := dashboard.NewService(serverRepo, metricRepo, alertEventRepo)

	// 定时采集调度器（cron + Worker Pool）
	schedulerSvc := scheduler.New(collectorSvc, serverRepo, &cfg.Collector)
	if err := schedulerSvc.Start(cfg.Collector.Interval); err != nil {
		utils.Logger.Fatal("启动定时采集失败", zap.Error(err))
	}

	// 数据保留清理任务（复用采集调度器的 cron 实例）
	retentionSvc := retention.NewService(metricRepo, alertEventRepo, notifyRepo, &cfg.Retention)
	if cfg.Retention.Enabled {
		if err := schedulerSvc.AddJob(cfg.Retention.Schedule, retentionSvc.Run); err != nil {
			utils.Logger.Fatal("注册数据清理任务失败", zap.Error(err))
		}
		utils.Logger.Info("数据清理任务已注册",
			zap.String("schedule", cfg.Retention.Schedule),
			zap.Int("metrics_days", cfg.Retention.MetricsDays),
			zap.Int("alert_events_days", cfg.Retention.AlertEventsDays),
		)
	}

	handlers := api.Handlers{
		Auth:                handler.NewAuthHandler(authSvc),
		Server:              handler.NewServerHandler(serverSvc, collectorSvc),
		ServerGroup:         handler.NewServerGroupHandler(groupSvc),
		Collector:           handler.NewCollectorHandler(schedulerSvc),
		AlertRule:           handler.NewAlertRuleHandler(alertSvc),
		AlertEvent:          handler.NewAlertEventHandler(alertSvc),
		NotificationChannel: handler.NewNotificationChannelHandler(notifierSvc),
		Dashboard:           handler.NewDashboardHandler(dashboardSvc),
	}

	// 5. 路由
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := api.SetupRouter(handlers, cfg.JWT.Secret)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	// 6. 启动
	go func() {
		utils.Logger.Info("服务启动",
			zap.String("addr", addr),
			zap.String("env", cfg.App.Env),
			zap.String("db", cfg.Database.Driver),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			utils.Logger.Fatal("服务启动失败", zap.Error(err))
		}
	}()

	// 7. 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	utils.Logger.Info("收到退出信号，正在关闭服务...")

	schedulerSvc.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		utils.Logger.Error("服务关闭异常", zap.Error(err))
	}
	utils.Logger.Info("服务已退出")
}
