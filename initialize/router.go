package initialize

import (
	"net/http"

	"personal-assistant-server/global"
	"personal-assistant-server/middleware"
	"personal-assistant-server/router"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.New()
	Router.Use(middleware.GinRecovery(true))
	Router.Use(middleware.Cors())

	if gin.Mode() == gin.DebugMode {
		Router.Use(gin.Logger())
	}

	PublicGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	{
		PublicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, "ok")
		})
	}

	// 公开路由（无需鉴权）
	router.InitAuthRouter(PublicGroup)

	// 私有路由（需要 JWT 鉴权）
	PrivateGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	PrivateGroup.Use(middleware.JWTAuth())
	{
		router.InitScheduleRouter(PrivateGroup)
		router.InitConversationRouter(PrivateGroup)
		router.InitViewRouter(PrivateGroup)
		router.InitPushRouter(PrivateGroup)
	}

	global.GVA_LOG.Info("router register success")
	return Router
}
