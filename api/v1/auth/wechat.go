package auth

import (
	"github.com/gin-gonic/gin"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service"
	"personal-assistant-server/service/auth"
	"personal-assistant-server/utils"
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

// RefreshToken 刷新 access token
// @Router /api/v1/auth/refresh [post]
func (a *AuthApi) RefreshToken(c *gin.Context) {
	var req auth.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("缺少refresh_token参数", c)
		return
	}

	resp, err := service.ServiceGroupApp.AuthService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		c.JSON(401, response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithData(resp, c)
}

// Logout 登出
// @Router /api/v1/auth/logout [post]
func (a *AuthApi) Logout(c *gin.Context) {
	userID := utils.GetUserID(c)
	deviceID, _ := c.Get("device_id")

	deviceIDStr, ok := deviceID.(string)
	if !ok || deviceIDStr == "" {
		response.FailWithMessage("无法获取设备信息", c)
		return
	}

	claims, _ := c.Get("claims")
	jti := ""
	if cl, ok := claims.(*utils.WechatClaims); ok {
		jti = cl.ID
	}

	if err := service.ServiceGroupApp.AuthService.Logout(c.Request.Context(), userID, deviceIDStr, jti); err != nil {
		response.FailWithMessage("登出失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("已登出", c)
}
