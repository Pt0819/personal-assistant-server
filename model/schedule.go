package model

import (
	"personal-assistant-server/global"
	"time"
)

type Schedule struct {
	global.GVA_MODEL
	UserID      uint      `json:"user_id" gorm:"index;comment:用户ID"`
	Title       string    `json:"title" gorm:"size:255;comment:日程标题"`
	Description string    `json:"description" gorm:"type:text;comment:描述"`
	Location    string    `json:"location" gorm:"size:255;comment:地点"`
	StartTime   time.Time `json:"start_time" gorm:"index;comment:开始时间"`
	EndTime     time.Time `json:"end_time" gorm:"comment:结束时间"`
	IsAllDay    bool      `json:"is_all_day" gorm:"default:false;comment:全天事件"`
	Tags        string    `json:"tags" gorm:"size:255;comment:标签"`
	Status      string    `json:"status" gorm:"default:active;size:20;comment:active/cancelled"`
	Source      string    `json:"source" gorm:"default:ai;size:20;comment:ai/manual/import"`
	Reminded    bool      `json:"reminded" gorm:"default:false;comment:已推送提醒"`
}

func (Schedule) TableName() string {
	return "schedules"
}
