package service

import (
	"personal-assistant-server/service/auth"
	"personal-assistant-server/service/conversation"
	"personal-assistant-server/service/push"
	"personal-assistant-server/service/schedule"
	"personal-assistant-server/service/user"
	"personal-assistant-server/service/view"
)

var ServiceGroupApp = new(ServiceGroup)

type ServiceGroup struct {
	AuthService         auth.AuthService
	ScheduleService     schedule.ScheduleService
	ConversationService conversation.ConversationService
	PushService         push.PushService
	ViewService         view.ViewService
	UserService         user.UserService
}
