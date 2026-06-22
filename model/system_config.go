package model

import "time"

// SystemConfig 系统配置表 — 存储敏感配置（密钥等），作为 config 文件为空时的兜底
type SystemConfig struct {
	ID        uint      `gorm:"primarykey"`
	Key       string    `gorm:"size:128;uniqueIndex;not null;comment:配置键"`
	Value     string    `gorm:"type:text;not null;comment:配置值"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
