package utils

import (
	"net"
	"time"

	"personal-assistant-server/global"

	"github.com/gin-gonic/gin"
)

func ClearToken(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	if net.ParseIP(host) != nil {
		c.SetCookie("x-token", "", -1, "/", "", false, false)
	} else {
		c.SetCookie("x-token", "", -1, "/", host, false, false)
	}
}

func SetToken(c *gin.Context, token string, maxAge int) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	if net.ParseIP(host) != nil {
		c.SetCookie("x-token", token, maxAge, "/", "", false, false)
	} else {
		c.SetCookie("x-token", token, maxAge, "/", host, false, false)
	}
}

func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get("x-token")
	if token == "" {
		// Also check Authorization: Bearer <token>
		authHeader := c.Request.Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}
	if token == "" {
		j := NewJWT()
		token, _ = c.Cookie("x-token")
		claims, err := j.ParseToken(token)
		if err != nil {
			if global.GVA_CONFIG.System.Env != "local" {
			global.GVA_LOG.Warn("重新写入cookie token失败,未能成功解析token,请检查请求头是否存在x-token且claims是否为规定结构")
		}
			return token
		}
		SetToken(c, token, int(claims.ExpiresAt.Unix()-time.Now().Unix()))
	}
	return token
}

// GetClaims extracts WechatClaims from the Gin context
func GetClaims(c *gin.Context) (*WechatClaims, error) {
	token := GetToken(c)
	j := NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		global.GVA_LOG.Error("从Gin的Context中获取从jwt解析信息失败, 请检查请求头是否存在x-token且claims是否为规定结构")
	}
	return claims, err
}

// GetUserID extracts user ID from JWT claims in Gin context
func GetUserID(c *gin.Context) uint {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return 0
		} else {
			return cl.UserID
		}
	} else {
		waitUse := claims.(*WechatClaims)
		return waitUse.UserID
	}
}

// GetOpenID extracts WeChat OpenID from JWT claims in Gin context
func GetOpenID(c *gin.Context) string {
	if claims, exists := c.Get("claims"); !exists {
		if cl, err := GetClaims(c); err != nil {
			return ""
		} else {
			return cl.OpenID
		}
	} else {
		waitUse := claims.(*WechatClaims)
		return waitUse.OpenID
	}
}
