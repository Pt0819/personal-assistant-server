package model

import "personal-assistant-server/global"

type Conversation struct {
	global.GVA_MODEL
	UserID uint   `json:"user_id" gorm:"index;comment:用户ID"`
	Title  string `json:"title" gorm:"size:255;comment:会话标题"`
	Status string `json:"status" gorm:"default:active;size:20;comment:active/archived"`
}

func (Conversation) TableName() string {
	return "conversations"
}
