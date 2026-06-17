package middleware

import (
	"errors"

	"personal-assistant-server/model/common/response"
	"personal-assistant-server/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuthLax is a relaxed JWT middleware that accepts expired-but-signed tokens.
// Used for logout endpoint — users should always be able to log out.
func JWTAuthLax() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := utils.GetToken(c)
		if token == "" {
			response.NoAuth("未登录", c)
			c.Abort()
			return
		}

		j := utils.NewJWT()
		claims, err := j.ParseTokenLax(token)
		if err != nil {
			if errors.Is(err, utils.TokenMalformed) ||
				errors.Is(err, utils.TokenSignatureInvalid) {
				response.NoAuth("token无效", c)
				c.Abort()
				return
			}
			// TokenExpired 不拦截 — JWTAuthLax 的核心语义
		}

		if claims != nil {
			c.Set("claims", claims)
			c.Set("user_id", claims.UserID)
			c.Set("device_id", claims.DeviceID)
		}
		c.Next()
	}
}
