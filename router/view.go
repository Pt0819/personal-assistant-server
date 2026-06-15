package router

import (
	"github.com/gin-gonic/gin"

	v1 "personal-assistant-server/api/v1"
)

func InitViewRouter(privateGroup *gin.RouterGroup) {
	viewApi := v1.ApiGroupApp.ViewApi
	viewRouter := privateGroup.Group("/view")
	{
		viewRouter.GET("/day", viewApi.DayView)
		viewRouter.GET("/week", viewApi.WeekView)
	}
}
