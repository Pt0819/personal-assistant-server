package schedule

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/schedule"
	"personal-assistant-server/utils"
)

type ScheduleApi struct{}

// Create 创建日程
func (a *ScheduleApi) Create(c *gin.Context) {
	var req schedule.CreateScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	s, err := service.ServiceGroupApp.ScheduleService.Create(c.Request.Context(), userID, &req)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(s, c)
}

// Update 更新日程
func (a *ScheduleApi) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的日程ID", c)
		return
	}

	var req schedule.UpdateScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	s, err := service.ServiceGroupApp.ScheduleService.Update(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(s, c)
}

// Delete 删除日程（软删除）
func (a *ScheduleApi) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的日程ID", c)
		return
	}

	userID := utils.GetUserID(c)
	if err := service.ServiceGroupApp.ScheduleService.Delete(c.Request.Context(), userID, uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.Ok(c)
}

// GetByID 获取单个日程
func (a *ScheduleApi) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("无效的日程ID", c)
		return
	}

	userID := utils.GetUserID(c)
	s, err := service.ServiceGroupApp.ScheduleService.GetByID(c.Request.Context(), userID, uint(id))
	if err != nil {
		response.FailWithMessage("未找到该日程", c)
		return
	}
	response.OkWithData(s, c)
}

// List 获取日程列表
func (a *ScheduleApi) List(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		startDate = time.Now().AddDate(0, 0, -7)
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		endDate = time.Now().AddDate(0, 0, 7)
	}

	userID := utils.GetUserID(c)
	schedules, err := service.ServiceGroupApp.ScheduleService.List(c.Request.Context(), userID, startDate, endDate)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(schedules, c)
}

// CheckConflict 检测时间冲突
func (a *ScheduleApi) CheckConflict(c *gin.Context) {
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")
	excludeStr := c.Query("exclude_id")

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		response.FailWithMessage("无效的开始时间格式", c)
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		response.FailWithMessage("无效的结束时间格式", c)
		return
	}

	var excludeID uint
	if excludeStr != "" {
		id, _ := strconv.ParseUint(excludeStr, 10, 64)
		excludeID = uint(id)
	}

	userID := utils.GetUserID(c)
	conflicts, err := service.ServiceGroupApp.ScheduleService.CheckConflict(c.Request.Context(), userID, startTime, endTime, excludeID)
	if err != nil {
		response.FailWithMessage("冲突检测失败: "+err.Error(), c)
		return
	}
	response.OkWithData(conflicts, c)
}
