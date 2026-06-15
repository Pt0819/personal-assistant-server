package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitAuthRouter(publicGroup *gin.RouterGroup) {
	authApi := v1.ApiGroupApp.AuthApi
	publicGroup.POST("/auth/wechat/login", authApi.Login)
}
