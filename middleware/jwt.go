package middleware

import (
	"errors"
	"strconv"
	"time"

	"personal-assistant-server/global"
	"personal-assistant-server/model"
	"personal-assistant-server/model/common/response"
	"personal-assistant-server/utils"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth("未登录或非法访问，请登录", c)
			c.Abort()
			return
		}

		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if errors.Is(err, utils.TokenExpired) {
				response.NoAuth("登录已过期，请重新登录", c)
				utils.ClearToken(c)
				c.Abort()
				return
			}
			response.NoAuth(err.Error(), c)
			utils.ClearToken(c)
			c.Abort()
			return
		}

		// 黑名单检查
		if claims.ID != "" {
			var count int64
			global.GVA_DB.Model(&model.JwtBlacklist{}).
				Where("jwt = ?", claims.ID).Count(&count)
			if count > 0 {
				response.NoAuth("token已被注销", c)
				utils.ClearToken(c)
				c.Abort()
				return
			}
		}

		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Set("device_id", claims.DeviceID)

		// Auto-refresh token if close to expiry
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr, _ := utils.ParseDuration(global.GVA_CONFIG.JWT.ExpiresTime)
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
			newToken, _ := j.CreateToken(*claims)
			newClaims, _ := j.ParseToken(newToken)
			c.Header("new-token", newToken)
			c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			utils.SetToken(c, newToken, int(dr.Seconds()/60))
		}

		c.Next()
	}
}
