package utils

import (
	"testing"
	"time"

	"personal-assistant-server/global"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateClaimsWithDeviceIDAndJTI(t *testing.T) {
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	defer func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}()

	j := NewJWT()
	claims := j.CreateClaims(42, "testuser", "openid-abc", "device-xyz")

	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "openid-abc", claims.OpenID)
	assert.Equal(t, "device-xyz", claims.DeviceID)
	assert.NotEmpty(t, claims.ID, "jti should be set")
	assert.Equal(t, "test-issuer", claims.Issuer)
	assert.Equal(t, "42", claims.Subject)
}

func TestCreateTokenAndParseWithDeviceID(t *testing.T) {
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	defer func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}()

	j := NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid-test", "device-uuid")
	tokenStr, err := j.CreateToken(claims)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	parsed, err := j.ParseToken(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, uint(1), parsed.UserID)
	assert.Equal(t, "device-uuid", parsed.DeviceID)
	assert.Equal(t, claims.ID, parsed.ID)
}

func TestParseTokenLax_ValidToken(t *testing.T) {
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	defer func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}()

	j := NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device")
	tokenStr, _ := j.CreateToken(claims)

	parsed, err := j.ParseTokenLax(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, uint(1), parsed.UserID)
}

func TestParseTokenLax_ExpiredToken(t *testing.T) {
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	defer func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}()

	j := NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device")
	claims.ExpiresAt = jwtlib.NewNumericDate(time.Now().Add(-1 * time.Hour))
	tokenStr, _ := j.CreateToken(claims)

	// ParseToken should reject expired token
	_, err := j.ParseToken(tokenStr)
	assert.ErrorIs(t, err, TokenExpired)

	// ParseTokenLax should accept expired token
	parsed, err := j.ParseTokenLax(tokenStr)
	require.NoError(t, err)
	assert.Equal(t, uint(1), parsed.UserID)
}

func TestParseTokenLax_InvalidSignature(t *testing.T) {
	origSigningKey := global.GVA_CONFIG.JWT.SigningKey
	origExpiresTime := global.GVA_CONFIG.JWT.ExpiresTime
	origBufferTime := global.GVA_CONFIG.JWT.BufferTime
	origIssuer := global.GVA_CONFIG.JWT.Issuer
	global.GVA_CONFIG.JWT.SigningKey = "test-key-at-least-32-characters-long!!"
	global.GVA_CONFIG.JWT.ExpiresTime = "2h"
	global.GVA_CONFIG.JWT.BufferTime = "5m"
	global.GVA_CONFIG.JWT.Issuer = "test-issuer"
	defer func() {
		global.GVA_CONFIG.JWT.SigningKey = origSigningKey
		global.GVA_CONFIG.JWT.ExpiresTime = origExpiresTime
		global.GVA_CONFIG.JWT.BufferTime = origBufferTime
		global.GVA_CONFIG.JWT.Issuer = origIssuer
	}()

	j := NewJWT()
	claims := j.CreateClaims(1, "testuser", "openid", "device")
	tokenStr, _ := j.CreateToken(claims)

	tampered := tokenStr + "tampered"

	_, err := j.ParseTokenLax(tampered)
	assert.Error(t, err)
}
