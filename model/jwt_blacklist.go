package model

import "personal-assistant-server/global"

type JwtBlacklist struct {
	global.GVA_MODEL
	Jwt string `json:"jwt" gorm:"type:text;comment:黑名单Token"`
}

func (JwtBlacklist) TableName() string {
	return "jwt_blacklists"
}
