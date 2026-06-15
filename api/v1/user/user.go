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
