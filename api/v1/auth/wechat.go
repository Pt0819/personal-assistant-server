package auth

import (
	"net/http"

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

// WebQrcode 获取网页登录二维码
// @Router /api/v1/auth/wechat/web/qrcode [get]
func (a *AuthApi) WebQrcode(c *gin.Context) {
	var req auth.WebQrcodeRequest
	resp, err := service.ServiceGroupApp.AuthService.WebQrcode(c.Request.Context(), req)
	if err != nil {
		c.JSON(500, response.Response{Code: 7, Msg: "获取二维码失败: " + err.Error()})
		return
	}
	response.OkWithData(resp, c)
}

// WebCallback 微信 OAuth 回调（302 重定向到 SPA）
// @Router /api/v1/auth/wechat/web/callback [get]
func (a *AuthApi) WebCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.FailWithMessage("缺少code或state参数", c)
		return
	}

	redirectURL, err := service.ServiceGroupApp.AuthService.WebCallback(c.Request.Context(), code, state)
	if err != nil {
		c.JSON(400, response.Response{Code: 7, Msg: "登录回调失败: " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, redirectURL)
}

// WebStatus 轮询网页登录状态
// @Router /api/v1/auth/wechat/web/status [get]
func (a *AuthApi) WebStatus(c *gin.Context) {
	tempToken := c.Query("temp_token")
	resp, err := service.ServiceGroupApp.AuthService.WebStatus(c.Request.Context(), tempToken)
	if err != nil {
		c.JSON(410, response.Response{Code: 7, Msg: err.Error()})
		return
	}
	response.OkWithData(resp, c)
}

// DevLogin 开发环境免扫码登录
// @Router /api/v1/auth/dev/login [post]
func (a *AuthApi) DevLogin(c *gin.Context) {
	resp, err := service.ServiceGroupApp.AuthService.DevLogin(c.Request.Context())
	if err != nil {
		c.JSON(403, response.Response{Code: 7, Msg: err.Error()})
		return
	}
	response.OkWithData(resp, c)
}
