package v1

import (
	"personal-assistant-server/api/v1/auth"
	"personal-assistant-server/api/v1/conversation"
	"personal-assistant-server/api/v1/push"
	"personal-assistant-server/api/v1/schedule"
	"personal-assistant-server/api/v1/view"
)

var ApiGroupApp = new(ApiGroup)

type ApiGroup struct {
	AuthApi         auth.AuthApi
	ScheduleApi     schedule.ScheduleApi
	ConversationApi conversation.ConversationApi
	ViewApi         view.ViewApi
	PushApi         push.PushApi
}
