package initialize

import (
	"fmt"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/service"
	"personal-assistant-server/service/auth"

	"github.com/robfig/cron/v3"
)

func Timer() {
	go func() {
		var option []cron.Option
		option = append(option, cron.WithSeconds())

		// 推送提醒定时任务：每60秒扫描一次
		_, err := global.GVA_Timer.AddTaskByFunc("PushReminder", "*/60 * * * * *", func() {
			if err := service.ServiceGroupApp.PushService.ScanAndPushReminders(); err != nil {
				fmt.Println("push timer error:", err)
			}
		}, "每分钟扫描待推送日程提醒", option...)
		if err != nil {
			fmt.Println("add timer error:", err)
		}

		// 每天凌晨 3:07 执行过期会话清理
		_, err = global.GVA_Timer.AddTaskByFunc("cleanExpiredSessions", "7 3 * * *", cleanExpiredSessions, "每天凌晨3:07清理过期会话和黑名单")
		if err != nil {
			global.GVA_LOG.Warn("注册过期会话清理任务失败: " + err.Error())
		}

		// 每小时清理过期验证码
		_, err = global.GVA_Timer.AddTaskByFunc("cleanExpiredVerifications", "7 * * * *", auth.CleanExpiredVerifications, "每小时第7分钟清理过期验证码")
		if err != nil {
			global.GVA_LOG.Warn("注册验证码清理任务失败: " + err.Error())
		}
	}()
}

// cleanExpiredSessions 清理过期的会话和黑名单记录
func cleanExpiredSessions() {
	if global.GVA_DB == nil {
		return
	}
	// 清理过期黑名单
	global.GVA_DB.Where("expires_at < ?", time.Now()).Delete(&model.JwtBlacklist{})
	// 清理过期会话
	global.GVA_DB.Where("refresh_expires_at < ?", time.Now()).Delete(&model.UserSession{})
}
