package auth

import (
	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/auth"
)

type AuthApi struct{}

// Login 微信小程序登录
// @Router /api/v1/auth/wechat/login [post]
func (a *AuthApi) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供微信登录code", c)
		return
	}

	if req.Code == "" {
		response.FailWithMessage("code不能为空", c)
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.LoginWithProfile(c.Request.Context(), req)
	if err != nil {
		response.FailWithMessage("登录失败: "+err.Error(), c)
		return
	}

	response.OkWithData(resp, c)
}
