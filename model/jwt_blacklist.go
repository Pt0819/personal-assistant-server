package model

import (
	"time"
)

type JwtBlacklist struct {
	ID        uint      `gorm:"primarykey"`
	Jwt       string    `gorm:"type:varchar(36);comment:JWT ID(jti),UUID格式,36字符"`
	ExpiresAt time.Time `gorm:"comment:黑名单过期时间(=JWT原过期时间)"`
	CreatedAt time.Time
}

func (JwtBlacklist) TableName() string {
	return "jwt_blacklists"
}
