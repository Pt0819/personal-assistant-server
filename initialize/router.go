package initialize

import (
	"net/http"
	"time"

	v1 "personal-assistant-server/api/v1"
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

	// 网页端微信 OAuth 登录（无需 JWT）
	PublicGroup.GET("/auth/wechat/web/qrcode", v1.ApiGroupApp.AuthApi.WebQrcode)
	PublicGroup.GET("/auth/wechat/web/callback", v1.ApiGroupApp.AuthApi.WebCallback)
	PublicGroup.GET("/auth/wechat/web/status", v1.ApiGroupApp.AuthApi.WebStatus)

	// 创建限速器: 每分钟最多 5 次（auth 端点）
	authRateLimiter := middleware.NewRateLimiter(5, time.Minute)
	// 对登录和刷新端点额外应用限速
	PublicGroup.POST("/auth/wechat/login", authRateLimiter.RateLimit(), v1.ApiGroupApp.AuthApi.Login)
	PublicGroup.POST("/auth/refresh", authRateLimiter.RateLimit(), v1.ApiGroupApp.AuthApi.RefreshToken)

	// 登出路由 — 使用宽松 JWT 中间件（接受过期token以允许随时登出）
	logoutGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	logoutGroup.Use(middleware.JWTAuthLax())
	{
		logoutGroup.POST("/auth/logout", v1.ApiGroupApp.AuthApi.Logout)
	}

	// 私有路由（需要 JWT 鉴权）
	PrivateGroup := Router.Group(global.GVA_CONFIG.System.RouterPrefix)
	PrivateGroup.Use(middleware.JWTAuth())
	{
		router.InitScheduleRouter(PrivateGroup)
		router.InitConversationRouter(PrivateGroup)
		router.InitViewRouter(PrivateGroup)
		router.InitPushRouter(PrivateGroup)
		router.InitUserRouter(PrivateGroup)
	}

	global.GVA_LOG.Info("router register success")
	return Router
}
