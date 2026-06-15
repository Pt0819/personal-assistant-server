package view

import (
	"time"

	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/utils"
)

type ViewApi struct{}

// DayView 日视图
func (a *ViewApi) DayView(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.FailWithMessage("无效的日期格式，请使用 YYYY-MM-DD", c)
		return
	}

	userID := utils.GetUserID(c)
	schedules, err := service.ServiceGroupApp.ViewService.DayView(c.Request.Context(), userID, date)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(schedules, c)
}

// WeekView 周视图
func (a *ViewApi) WeekView(c *gin.Context) {
	dateStr := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.FailWithMessage("无效的日期格式，请使用 YYYY-MM-DD", c)
		return
	}

	userID := utils.GetUserID(c)
	weekData, err := service.ServiceGroupApp.ViewService.WeekView(c.Request.Context(), userID, date)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(weekData, c)
}
