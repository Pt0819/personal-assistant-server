package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitScheduleRouter(privateGroup *gin.RouterGroup) {
	scheduleApi := v1.ApiGroupApp.ScheduleApi
	scheduleRouter := privateGroup.Group("/schedules")
	{
		scheduleRouter.POST("", scheduleApi.Create)
		scheduleRouter.PUT("/:id", scheduleApi.Update)
		scheduleRouter.DELETE("/:id", scheduleApi.Delete)
		scheduleRouter.GET("/:id", scheduleApi.GetByID)
		scheduleRouter.GET("", scheduleApi.List)
		scheduleRouter.GET("/conflicts", scheduleApi.CheckConflict)
	}
}
