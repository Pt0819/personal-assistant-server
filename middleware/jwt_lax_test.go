package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"personal-assistant-server/utils"

	"github.com/gin-gonic/gin"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthLax_ValidToken(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	gin.SetMode(gin.TestMode)

	j := utils.NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device-1")
	tokenStr, _ := j.CreateToken(claims)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/logout", nil)
	c.Request.Header.Set("x-token", tokenStr)

	JWTAuthLax()(c)

	assert.False(t, c.IsAborted())
	deviceID, _ := c.Get("device_id")
	assert.Equal(t, "device-1", deviceID)
}

func TestJWTAuthLax_ExpiredToken(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	gin.SetMode(gin.TestMode)

	j := utils.NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device-1")
	claims.ExpiresAt = jwtlib.NewNumericDate(time.Now().Add(-1 * time.Hour))
	tokenStr, _ := j.CreateToken(claims)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/logout", nil)
	c.Request.Header.Set("x-token", tokenStr)

	JWTAuthLax()(c)

	assert.False(t, c.IsAborted())
}

func TestJWTAuthLax_NoToken(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/logout", nil)

	JWTAuthLax()(c)

	assert.True(t, c.IsAborted())
}

func TestJWTAuthLax_InvalidSignature(t *testing.T) {
	cleanup := setupTestConfig()
	defer cleanup()
	gin.SetMode(gin.TestMode)

	j := utils.NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device-1")
	tokenStr, _ := j.CreateToken(claims)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/auth/logout", nil)
	c.Request.Header.Set("x-token", tokenStr+"tampered")

	JWTAuthLax()(c)

	assert.True(t, c.IsAborted())
}
