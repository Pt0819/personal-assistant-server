package auth

import (
	"testing"

	"personal-assistant-server/global"
	"personal-assistant-server/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupAuthServiceTest(t *testing.T) (sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	global.GVA_DB = gormDB

	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origRefreshExpiresTime := global.GVA_CONFIG.JWT.RefreshExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	origMaxDevices := global.GVA_CONFIG.Security.MaxDevices
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.RefreshExpiresTime = "720h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	global.GVA_CONFIG.Security.MaxDevices = 3

	encKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
	}
	global.GVA_ENCRYPTION_KEY = encKey

	cleanup := func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.RefreshExpiresTime = origRefreshExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
		global.GVA_CONFIG.Security.MaxDevices = origMaxDevices
		global.GVA_ENCRYPTION_KEY = nil
	}
	return mock, cleanup
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := utils.GenerateRefreshToken()
	require.NoError(t, err)
	assert.Len(t, token, 64)
}

func TestHashRefreshToken(t *testing.T) {
	raw := "test-refresh-token-12345678"
	hash := utils.HashRefreshToken(raw)
	assert.Len(t, hash, 64)
	assert.Equal(t, hash, utils.HashRefreshToken(raw))
}

func TestLoginRequestJSON(t *testing.T) {
	req := LoginRequest{
		Code:       "test-code",
		Nickname:   "测试用户",
		AvatarURL:  "https://example.com/avatar.png",
		DeviceID:   "device-uuid-001",
		DeviceInfo: "Test Device",
	}
	assert.Equal(t, "test-code", req.Code)
	assert.Equal(t, "device-uuid-001", req.DeviceID)
}

func TestLoginResponseFields(t *testing.T) {
	resp := LoginResponse{
		AccessToken:  "access-token-xxx",
		RefreshToken: "refresh-token-yyy",
		ExpiresIn:    7200,
	}
	assert.Equal(t, "access-token-xxx", resp.AccessToken)
	assert.Equal(t, "refresh-token-yyy", resp.RefreshToken)
	assert.Equal(t, int64(7200), resp.ExpiresIn)
}
