package utils

import (
	"errors"
	"time"

	"personal-assistant-server/global"

	jwt "github.com/golang-jwt/jwt/v5"
)

// WechatClaims JWT claims for WeChat mini-program users
type WechatClaims struct {
	UserID     uint   `json:"user_id"`
	OpenID     string `json:"open_id"`
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

// CreateClaims creates WechatClaims with configured expiration
func (j *JWT) CreateClaims(userID uint, openID string) WechatClaims {
	bf, _ := ParseDuration(global.GVA_CONFIG.JWT.BufferTime)
	ep, _ := ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
	claims := WechatClaims{
		UserID:     userID,
		OpenID:     openID,
		BufferTime: int64(bf / time.Second),
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"PA"},                    // Personal Assistant
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),    // 过期时间
			Issuer:    global.GVA_CONFIG.JWT.Issuer,              // 签发者
		},
	}
	// 缓冲时间用于token自动刷新
	_ = bf
	return claims
}

// CreateToken creates a signed JWT token
func (j *JWT) CreateToken(claims WechatClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// ParseToken parses and validates a JWT token string
func (j *JWT) ParseToken(tokenString string) (*WechatClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &WechatClaims{}, func(token *jwt.Token) (i interface{}, e error) {
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
		if claims, ok := token.Claims.(*WechatClaims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, TokenValid
}
