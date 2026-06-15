package auth

import (
	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
)

type AuthApi struct{}

// Login 微信小程序登录
// @Router /api/v1/auth/wechat/login [post]
func (a *AuthApi) Login(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请提供微信登录code", c)
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.Login(c.Request.Context(), req.Code)
	if err != nil {
		response.FailWithMessage("登录失败: "+err.Error(), c)
		return
	}

	response.OkWithData(resp, c)
}
