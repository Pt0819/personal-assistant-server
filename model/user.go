package model

import "personal-assistant-server/global"

type User struct {
	global.GVA_MODEL
	OpenID                 string `json:"openid" gorm:"uniqueIndex;size:64;comment:微信OpenID"`
	UnionID                string `json:"unionid" gorm:"index;size:64;comment:微信UnionID"`
	Nickname               string `json:"nickname" gorm:"size:128;default:'';comment:微信昵称"`
	AvatarURL              string `json:"avatar_url" gorm:"size:512;default:'';comment:头像URL"`
	Phone                  string `json:"phone" gorm:"size:20;default:null;comment:手机号"`
	DefaultReminderMinutes int    `json:"default_reminder_minutes" gorm:"default:30;comment:默认提醒分钟数"`
	WeekStartDay           int    `json:"week_start_day" gorm:"default:1;comment:周起始日 1=周一"`
	OnboardingCompleted    bool   `json:"onboarding_completed" gorm:"default:false;comment:引导是否完成"`
	Status                 int    `json:"status" gorm:"default:1;comment:1=正常 0=禁用"`
}

func (User) TableName() string {
	return "users"
}
