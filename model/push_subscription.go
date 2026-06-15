package model

import "personal-assistant-server/global"

type PushSubscription struct {
	global.GVA_MODEL
	UserID     uint   `json:"user_id" gorm:"index;comment:用户ID"`
	OpenID     string `json:"openid" gorm:"size:64;comment:微信OpenID"`
	TemplateID string `json:"template_id" gorm:"size:64;comment:订阅消息模板ID"`
	IsActive   bool   `json:"is_active" gorm:"default:true;comment:是否启用"`
}

func (PushSubscription) TableName() string {
	return "push_subscriptions"
}
