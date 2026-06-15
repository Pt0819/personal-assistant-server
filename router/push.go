package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitPushRouter(privateGroup *gin.RouterGroup) {
	pushApi := v1.ApiGroupApp.PushApi
	pushRouter := privateGroup.Group("/push")
	{
		pushRouter.POST("/subscribe", pushApi.Subscribe)
		pushRouter.DELETE("/unsubscribe", pushApi.Unsubscribe)
		pushRouter.GET("/subscriptions", pushApi.ListSubscriptions)
	}
}
