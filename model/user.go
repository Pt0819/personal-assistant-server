package model

import "personal-assistant-server/global"

type User struct {
	global.GVA_MODEL
	OpenID                 string `json:"openid" gorm:"column:openid;uniqueIndex;size:64;comment:微信OpenID"`
	UnionID                string `json:"unionid" gorm:"column:unionid;index;size:64;comment:微信UnionID"`
	Nickname               string `json:"nickname" gorm:"column:nickname;size:128;default:'';comment:微信昵称"`
	AvatarURL              string `json:"avatar_url" gorm:"column:avatar_url;size:512;default:'';comment:头像URL"`
	Phone                  string `json:"phone" gorm:"column:phone;size:20;comment:手机号"`
	DefaultReminderMinutes int    `json:"default_reminder_minutes" gorm:"column:default_reminder_minutes;default:30;comment:默认提醒分钟数"`
	WeekStartDay           int    `json:"week_start_day" gorm:"column:week_start_day;default:1;comment:周起始日 1=周一"`
	OnboardingCompleted    bool   `json:"onboarding_completed" gorm:"column:onboarding_completed;default:false;comment:引导是否完成"`
	Status                 int    `json:"status" gorm:"column:status;default:1;comment:1=正常 0=禁用"`
}

func (User) TableName() string {
	return "users"
}
