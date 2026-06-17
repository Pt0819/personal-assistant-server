package model

import (
	"time"
)

// UserSession 用户会话表 — 双Token + 多设备管理
type UserSession struct {
	ID                    uint      `gorm:"primarykey"`
	UserID                uint      `gorm:"not null;comment:用户ID"`
	RefreshTokenHash      string    `gorm:"size:128;not null;comment:Refresh Token SHA-256哈希"`
	DeviceID              string    `gorm:"size:128;not null;comment:设备唯一标识"`
	DeviceInfo            string    `gorm:"size:255;comment:设备信息"`
	SessionKeyEncrypted   []byte    `gorm:"type:varbinary(512);comment:微信session_key AES-256-GCM加密"`
	AccessExpiresAt       time.Time `gorm:"not null;comment:Access Token过期时间"`
	RefreshExpiresAt      time.Time `gorm:"not null;comment:Refresh Token过期时间"`
	LastUsedAt            time.Time `gorm:"not null;default:CURRENT_TIMESTAMP;comment:最后活跃时间"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (UserSession) TableName() string {
	return "user_sessions"
}
