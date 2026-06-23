package utils

import (
	"errors"
	"fmt"
	"time"

	"personal-assistant-server/global"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// UserClaims JWT claims for all user types (WeChat, email, phone)
type UserClaims struct {
	UserID     uint   `json:"user_id"`
	Username   string `json:"username"`
	OpenID     string `json:"open_id,omitempty"`
	DeviceID   string `json:"device_id"`    // 设备/会话标识
	BufferTime int64  `json:"buffer_time"`
	jwt.RegisteredClaims
}

type JWT struct {
	SigningKey []byte
}

var (
	TokenValid            = errors.New("未知错误")
	TokenExpired          = errors.New("token已过期")
	TokenNotValidYet      = errors.New("token尚未激活")
	TokenMalformed        = errors.New("这不是一个token")
	TokenSignatureInvalid = errors.New("无效签名")
	TokenInvalid          = errors.New("无法处理此token")
)

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.GVA_CONFIG.JWT.SigningKey),
	}
}

// CreateClaims creates UserClaims with configured expiration
func (j *JWT) CreateClaims(userID uint, username string, openID string, deviceID string) UserClaims {
	bf, _ := ParseDuration(global.GVA_CONFIG.JWT.BufferTime)
	ep, _ := ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	claims := UserClaims{
		UserID:     userID,
		Username:   username,
		OpenID:     openID,
		DeviceID:   deviceID,
		BufferTime: int64(bf / time.Second),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),                       // jti — 用于黑名单精确匹配
			Audience:  jwt.ClaimStrings{"PA"},                    // Personal Assistant
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),    // 过期时间
			Issuer:    global.GVA_CONFIG.JWT.Issuer,              // 签发者
			IssuedAt:  jwt.NewNumericDate(time.Now()),            // 签发时间
			Subject:   fmt.Sprintf("%d", userID),                 // 用户标识
		},
	}
	// 缓冲时间用于token自动刷新
	_ = bf
	return claims
}

// CreateToken creates a signed JWT token
func (j *JWT) CreateToken(claims UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken parses and validates a JWT token string
func (j *JWT) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, TokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, TokenMalformed
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, TokenSignatureInvalid
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, TokenNotValidYet
		default:
			return nil, TokenInvalid
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, TokenValid
}

// ParseTokenLax parses a JWT without validating exp/nbf — used for logout endpoint.
func (j *JWT) ParseTokenLax(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.SigningKey, nil
		},
		jwt.WithoutClaimsValidation(),
	)
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, TokenMalformed
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, TokenSignatureInvalid
		default:
			return nil, TokenInvalid
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*UserClaims); ok {
			return claims, nil
		}
	}
	return nil, TokenInvalid
}
