package initialize

import (
	"fmt"

	"personal-assistant-server/global"
	"personal-assistant-server/service"

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
	}()
}
