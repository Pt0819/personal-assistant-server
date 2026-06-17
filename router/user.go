package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitUserRouter(privateGroup *gin.RouterGroup) {
	userApi := v1.ApiGroupApp.UserApi
	userRouter := privateGroup.Group("/user")
	{
		userRouter.GET("/profile", userApi.GetProfile)
		userRouter.PUT("/profile", userApi.UpdateProfile)
		userRouter.PUT("/phone", userApi.BindPhone)
		userRouter.GET("/sessions", userApi.GetSessions)
		userRouter.DELETE("/sessions/:device_id", userApi.KickSession)
	}
}
