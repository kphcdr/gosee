package model

import "time"

// User 后台管理员账号。第一版单用户，预留多用户字段。
type User struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:100;not null" json:"-"` // bcrypt 哈希，不返回前端
	Nickname    string     `gorm:"size:50" json:"nickname"`
	Email       string     `gorm:"size:100" json:"email"`
	Status      int8       `gorm:"not null;default:1" json:"status"` // 1=启用 0=禁用
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
