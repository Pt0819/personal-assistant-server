package auth

import (
	"net/http"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/service/auth"

	"github.com/gin-gonic/gin"
)

// SendEmailCode sends an email verification code
// @Router /api/v1/auth/register/send-email-code [post]
func (a *AuthApi) SendEmailCode(c *gin.Context) {
	var req auth.SendEmailCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Code: 7, Msg: "请提供邮箱地址"})
		return
	}

	svc := auth.NewRegisterService()
	if err := svc.SendEmailCode(c.Request.Context(), &req); err != nil {
		c.JSON(httpCode(err), response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithMessage("验证码已发送", c)
}

// SendSMSCode sends an SMS verification code
// @Router /api/v1/auth/register/send-sms-code [post]
func (a *AuthApi) SendSMSCode(c *gin.Context) {
	var req auth.SendSMSCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Code: 7, Msg: "请提供手机号"})
		return
	}

	svc := auth.NewRegisterService()
	if err := svc.SendSMSCode(c.Request.Context(), &req); err != nil {
		c.JSON(httpCode(err), response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithMessage("验证码已发送", c)
}

// RegisterByEmail registers a user by email + verification code + password
// @Router /api/v1/auth/register/email [post]
func (a *AuthApi) RegisterByEmail(c *gin.Context) {
	var req auth.RegisterByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Code: 7, Msg: "请提供邮箱、验证码和密码"})
		return
	}

	svc := auth.NewRegisterService()
	resp, err := svc.RegisterByEmail(c.Request.Context(), &req)
	if err != nil {
		c.JSON(httpCode(err), response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithData(resp, c)
}

// RegisterByPhone registers a user by phone + verification code + password
// @Router /api/v1/auth/register/phone [post]
func (a *AuthApi) RegisterByPhone(c *gin.Context) {
	var req auth.RegisterByPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Code: 7, Msg: "请提供手机号、验证码和密码"})
		return
	}

	svc := auth.NewRegisterService()
	resp, err := svc.RegisterByPhone(c.Request.Context(), &req)
	if err != nil {
		c.JSON(httpCode(err), response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithData(resp, c)
}

// LoginByCredential handles email/phone + password login
// @Router /api/v1/auth/login/credential [post]
func (a *AuthApi) LoginByCredential(c *gin.Context) {
	var req auth.CredentialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Response{Code: 7, Msg: "请提供账号和密码"})
		return
	}

	svc := auth.NewRegisterService()
	resp, err := svc.LoginByCredential(c.Request.Context(), &req)
	if err != nil {
		c.JSON(httpCode(err), response.Response{Code: 7, Msg: err.Error()})
		return
	}

	response.OkWithData(resp, c)
}

// httpCode maps error messages to appropriate HTTP status codes
func httpCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	msg := err.Error()
	switch {
	case msg == "发送太频繁，请60秒后再试" || msg == "验证次数过多，请重新获取验证码":
		return http.StatusTooManyRequests
	case msg == "该邮箱已注册" || msg == "该手机号已注册":
		return http.StatusConflict
	case msg == "邮箱/手机号或密码错误":
		return http.StatusUnauthorized
	default:
		return http.StatusBadRequest
	}
}
