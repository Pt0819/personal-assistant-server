package user

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/user"
	"personal-assistant-server/utils"
)

type UserApi struct{}

func (a *UserApi) GetProfile(c *gin.Context) {
	userID := utils.GetUserID(c)
	u, err := service.ServiceGroupApp.UserService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		response.FailWithMessage("获取用户信息失败: "+err.Error(), c)
		return
	}
	response.OkWithData(u, c)
}

func (a *UserApi) UpdateProfile(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req user.UpdateProfileRequest
	if err := c.ShouldBind(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if v := c.PostForm("default_reminder_minutes"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			req.DefaultReminderMinutes = &n
		}
	}
	if v := c.PostForm("onboarding_completed"); v != "" {
		b := v == "true"
		req.OnboardingCompleted = &b
	}

	u, err := service.ServiceGroupApp.UserService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithData(u, c)
}

// BindPhone 绑定/更新手机号
func (a *UserApi) BindPhone(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req user.BindPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("code不能为空", c)
		return
	}

	resp, err := service.ServiceGroupApp.UserService.BindPhone(c.Request.Context(), userID, &req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithDetailed(resp, "手机号绑定成功", c)
}

// GetSessions 获取用户的所有活跃会话
func (a *UserApi) GetSessions(c *gin.Context) {
	userID := utils.GetUserID(c)
	deviceID, _ := c.Get("device_id")
	currentDeviceID, _ := deviceID.(string)

	resp, err := service.ServiceGroupApp.UserService.GetSessions(c.Request.Context(), userID, currentDeviceID)
	if err != nil {
		response.FailWithMessage("获取设备列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(resp, c)
}

// KickSession 踢出指定设备
func (a *UserApi) KickSession(c *gin.Context) {
	userID := utils.GetUserID(c)
	deviceID, _ := c.Get("device_id")
	currentDeviceID, _ := deviceID.(string)
	targetDeviceID := c.Param("device_id")

	if err := service.ServiceGroupApp.UserService.KickSession(c.Request.Context(), userID, targetDeviceID, currentDeviceID); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	response.OkWithMessage("设备已下线", c)
}
