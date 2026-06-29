package model

import "time"

// ServerStatus 服务器状态枚举
const (
	ServerStatusNormal   = "normal"   // 正常
	ServerStatusWarning  = "warning"  // 有预警
	ServerStatusCritical = "critical" // 严重异常
	ServerStatusOffline  = "offline"  // SSH 连接失败
	ServerStatusDisabled = "disabled" // 已禁用
	ServerStatusUnknown  = "unknown"  // 未采集
)

// AuthType 认证方式
const (
	AuthTypePrivateKey = "private_key" // 私钥（推荐）
	AuthTypePassword   = "password"    // 密码（非推荐）
)

// Server 被监控的目标服务器。私钥/密码字段加密存储，JSON 不输出明文。
type Server struct {
	ID                  int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name                string     `gorm:"size:100;not null" json:"name"`
	GroupID             *int64     `gorm:"column:group_id;index" json:"group_id"`
	Host                string     `gorm:"size:255;not null" json:"host"`
	Port                int        `gorm:"not null;default:22" json:"port"`
	Username            string     `gorm:"size:100;not null" json:"username"`
	AuthType            string     `gorm:"size:20;not null;default:private_key" json:"auth_type"`
	PrivateKeyEncrypted *string    `gorm:"type:text" json:"-"` // AES-GCM 加密后的私钥
	PasswordEncrypted   *string    `gorm:"type:text" json:"-"` // AES-GCM 加密后的密码
	Remark              string     `gorm:"size:500" json:"remark"`
	Status              string     `gorm:"size:20;not null;default:unknown;index" json:"status"`
	Enabled             int8       `gorm:"not null;default:1;index" json:"enabled"` // 1=启用 0=禁用
	LastCheckedAt       *time.Time `json:"last_checked_at"`
	LastError           *string    `gorm:"type:text" json:"last_error"`
	SSHFailureCount     int        `gorm:"not null;default:0" json:"ssh_failure_count"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// HasPrivateKey 是否已配置私钥
func (s Server) HasPrivateKey() bool {
	return s.PrivateKeyEncrypted != nil && *s.PrivateKeyEncrypted != ""
}

// HasPassword 是否已配置密码
func (s Server) HasPassword() bool { return s.PasswordEncrypted != nil && *s.PasswordEncrypted != "" }
