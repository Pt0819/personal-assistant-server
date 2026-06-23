package auth

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"personal-assistant-server/service/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSendEmailCode_InvalidEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(auth.SendEmailCodeRequest{Email: "not-an-email"})
	c.Request = httptest.NewRequest("POST", "/auth/register/send-email-code", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api := &AuthApi{}
	api.SendEmailCode(c)

	assert.Equal(t, 400, w.Code)
}

func TestSendSMSCode_InvalidPhone(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(auth.SendSMSCodeRequest{Phone: "123"})
	c.Request = httptest.NewRequest("POST", "/auth/register/send-sms-code", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api := &AuthApi{}
	api.SendSMSCode(c)

	assert.Equal(t, 400, w.Code)
}

func TestRegisterByEmail_WeakPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(auth.RegisterByEmailRequest{
		Email:    "test@example.com",
		Code:     "123456",
		Password: "short",
	})
	c.Request = httptest.NewRequest("POST", "/auth/register/email", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api := &AuthApi{}
	api.RegisterByEmail(c)

	assert.Equal(t, 400, w.Code)
}

func TestLoginByCredential_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/auth/login/credential", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	api := &AuthApi{}
	api.LoginByCredential(c)

	assert.Equal(t, 400, w.Code)
}
