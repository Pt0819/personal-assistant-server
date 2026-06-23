package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"personal-assistant-server/global"
	"personal-assistant-server/utils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) sqlmock.Sqlmock {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	global.GVA_DB = gormDB
	return mock
}

func setupTestConfig() func() {
	origLog := global.GVA_LOG
	global.GVA_LOG = zap.NewNop()
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	return func() {
		global.GVA_LOG = origLog
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}
}

func TestJWTAuth_BlacklistedToken(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	mock := setupMockDB(t)

	gin.SetMode(gin.TestMode)

	j := utils.NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device-1")
	tokenStr, err := j.CreateToken(claims)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `jwt_blacklists`").
		WithArgs(claims.ID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("x-token", tokenStr)

	JWTAuth()(c)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_ValidToken(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	mock := setupMockDB(t)

	gin.SetMode(gin.TestMode)

	j := utils.NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device-1")
	tokenStr, err := j.CreateToken(claims)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `jwt_blacklists`").
		WithArgs(claims.ID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", nil)
	c.Request.Header.Set("x-token", tokenStr)

	JWTAuth()(c)

	assert.False(t, c.IsAborted())
	deviceID, exists := c.Get("device_id")
	assert.True(t, exists)
	assert.Equal(t, "device-1", deviceID)
}
