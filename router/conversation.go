package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitConversationRouter(privateGroup *gin.RouterGroup) {
	convApi := v1.ApiGroupApp.ConversationApi
	convRouter := privateGroup.Group("/conversations")
	{
		convRouter.POST("", convApi.Create)
		convRouter.GET("", convApi.List)
		convRouter.GET("/:id", convApi.Get)
		convRouter.POST("/:id/messages", convApi.SendMessage)
		convRouter.GET("/:id/messages", convApi.ListMessages)
	}
}
