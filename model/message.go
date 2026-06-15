package model

import "personal-assistant-server/global"

type Message struct {
	global.GVA_MODEL
	ConversationID uint   `json:"conversation_id" gorm:"index;comment:会话ID"`
	UserID         uint   `json:"user_id" gorm:"index;comment:用户ID"`
	Role           string `json:"role" gorm:"size:20;comment:user/assistant/system"`
	Content        string `json:"content" gorm:"type:longtext;comment:消息内容"`
	Intent         string `json:"intent" gorm:"size:50;comment:意图分类"`
	ParsedJSON     string `json:"parsed_json" gorm:"type:text;comment:AI解析的结构化JSON"`
	ModelUsed      string `json:"model_used" gorm:"size:50;comment:使用的模型"`
	LatencyMs      int    `json:"latency_ms" gorm:"comment:延迟毫秒"`
}

func (Message) TableName() string {
	return "messages"
}
