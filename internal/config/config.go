package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置，对应 configs/config.yaml
type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Admin     AdminConfig     `mapstructure:"admin"`
	Security  SecurityConfig  `mapstructure:"security"`
	Log       LogConfig       `mapstructure:"log"`
	Collector CollectorConfig `mapstructure:"collector"`
	Retention RetentionConfig `mapstructure:"retention"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"` // sqlite | mysql
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	Issuer      string `mapstructure:"issuer"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type SecurityConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	Dir   string `mapstructure:"dir"`
}

// CollectorConfig 采集调度配置
type CollectorConfig struct {
	Interval          string `mapstructure:"interval"`            // cron @every 间隔，如 "10m"
	WorkerCount       int    `mapstructure:"worker_count"`        // 并发 worker 数
	SSHConnectTimeout string `mapstructure:"ssh_connect_timeout"` // SSH 连接超时
	SSHCommandTimeout string `mapstructure:"ssh_command_timeout"` // 脚本执行超时
	MaxRetries        int    `mapstructure:"max_retries"`         // 单台采集失败后的重试次数
}

// RetentionConfig 数据保留与自动清理配置。
// SQLite 的 DELETE 不缩小 db 文件，删除的页进入 freelist 被后续插入复用，
// 增长会趋于稳态；如需真正回收磁盘，低峰期手动 `sqlite3 gosee.db "VACUUM"`。
type RetentionConfig struct {
	Enabled           bool   `mapstructure:"enabled"`            // 总开关
	Schedule          string `mapstructure:"schedule"`           // 5 字段标准 cron 表达式，默认 "0 3 * * *"（每天 03:00）
	MetricsDays       int    `mapstructure:"metrics_days"`       // server_metrics + server_disks 保留天数；<=0 表示不清理
	AlertEventsDays   int    `mapstructure:"alert_events_days"`  // 已结束告警事件(status<>firing)保留天数；<=0 表示不清理
	NotificationsDays int    `mapstructure:"notifications_days"` // 通知发送记录保留天数；<=0 表示不清理
	BatchSize         int    `mapstructure:"batch_size"`         // 分批删除每批上限，控制单事务时长避免锁库
}

// Load 读取并解析配置文件。configPath 为空时默认 configs/config.yaml。
// 支持环境变量覆盖：前缀 GOSEE，键中的 "." 替换为 "_"，如 GOSEE_DATABASE_DSN。
func Load(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	if configPath == "" {
		configPath = "configs/config.yaml"
	}
	v.SetConfigFile(configPath)

	v.SetEnvPrefix("GOSEE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置失败 %s: %w", configPath, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return &cfg, nil
}
