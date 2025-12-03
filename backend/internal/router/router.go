package router

import (
	"os"

	"embyhub/internal/handler"
	"embyhub/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 跨域中间件
	r.Use(middleware.CORSMiddleware())

	// 健康检查（运维接口）
	healthHandler := handler.NewHealthHandler()
	r.GET("/health", healthHandler.Health)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/live", healthHandler.Live)

	// 安装向导路由（无需认证）
	SetupSetupRoutes(r)

	// 初始化处理器
	authHandler := handler.NewAuthHandler()
	registerHandler := handler.NewRegisterHandler()
	userHandler := handler.NewUserHandler()
	roleHandler := handler.NewRoleHandler()
	permissionHandler := handler.NewPermissionHandler()
	accessRecordHandler := handler.NewAccessRecordHandler()
	systemConfigHandler := handler.NewSystemConfigHandler()
	embyHandler := handler.NewEmbyHandler()
	cardKeyHandler := handler.NewCardKeyHandler()

	// 初始化邮件处理器
	emailHandler := handler.NewEmailHandler()

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关（无需JWT，但有频率限制）
		auth := api.Group("/auth")
		{
			auth.POST("/login", middleware.LoginRateLimitMiddleware(), authHandler.Login)
			auth.POST("/register", middleware.LoginRateLimitMiddleware(), registerHandler.Register)
		}

		// 邮件相关（无需JWT，但有频率限制）
		email := api.Group("/email")
		{
			email.POST("/send-code", middleware.LoginRateLimitMiddleware(), emailHandler.SendCode)
			email.POST("/reset-code", middleware.LoginRateLimitMiddleware(), emailHandler.SendResetCode)
			email.POST("/reset-password", middleware.LoginRateLimitMiddleware(), emailHandler.ResetPassword)
		}

		// 需要认证的路由
		authorized := api.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 认证相关
			authorized.POST("/auth/logout", authHandler.Logout)
			authorized.GET("/auth/current", authHandler.GetCurrentUser)
			authorized.PUT("/auth/password", authHandler.ChangePassword) // 用户修改自己的密码
			authorized.POST("/auth/refresh", authHandler.RefreshToken)   // 刷新Token

			// 用户管理
			users := authorized.Group("/users")
			{
				users.GET("", userHandler.List)
				users.POST("", middleware.PermissionMiddleware("user:create"), userHandler.Create)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", middleware.PermissionMiddleware("user:edit"), userHandler.Update)
				users.DELETE("/:id", middleware.PermissionMiddleware("user:delete"), userHandler.Delete)
				users.PUT("/:id/password", middleware.PermissionMiddleware("user:edit"), userHandler.ResetPassword)
				users.PUT("/:id/vip", middleware.PermissionMiddleware("user:edit"), userHandler.SetVip)
				users.PUT("/batch/status", middleware.PermissionMiddleware("user:edit"), userHandler.BatchUpdateStatus)
			}

			// 角色管理
			roles := authorized.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.POST("", middleware.PermissionMiddleware("role:create"), roleHandler.Create)
				roles.GET("/:id", roleHandler.GetByID)
				roles.PUT("/:id", middleware.PermissionMiddleware("role:edit"), roleHandler.Update)
				roles.DELETE("/:id", middleware.PermissionMiddleware("role:delete"), roleHandler.Delete)
				roles.POST("/:id/permissions", middleware.PermissionMiddleware("permission:assign"), roleHandler.AssignPermissions)
			}

			// 权限管理
			permissions := authorized.Group("/permissions")
			{
				permissions.GET("", permissionHandler.List)
			}

			// 访问记录
			accessRecords := authorized.Group("/access-records")
			{
				accessRecords.GET("", middleware.PermissionMiddleware("stats:view"), accessRecordHandler.List)
				accessRecords.POST("", accessRecordHandler.Create)
			}

			// 统计数据
			authorized.GET("/statistics", middleware.PermissionMiddleware("stats:view"), accessRecordHandler.GetStatistics)

			// 系统配置
			configs := authorized.Group("/configs")
			{
				configs.GET("", middleware.PermissionMiddleware("system:view"), systemConfigHandler.List)
				configs.PUT("/:key", middleware.PermissionMiddleware("system:edit"), systemConfigHandler.Update)
			}

			// 邮件测试（需要系统配置权限）
			authorized.POST("/email/test", middleware.PermissionMiddleware("system:edit"), emailHandler.TestConfig)

			// Emby同步
			emby := authorized.Group("/emby")
			{
				emby.POST("/test", middleware.PermissionMiddleware("emby:config"), embyHandler.TestConnection)
				emby.POST("/sync", middleware.PermissionMiddleware("emby:sync"), embyHandler.SyncUsers)
				emby.GET("/users", middleware.PermissionMiddleware("emby:view"), embyHandler.GetUsers)
			}

			// 媒体库（所有登录用户可访问）
			media := authorized.Group("/media")
			{
				media.GET("/server-url", embyHandler.GetServerURL) // 获取服务器URL
				media.GET("/libraries", embyHandler.GetLibraries)
				media.GET("/items", embyHandler.GetItems)
				media.GET("/items/:id", embyHandler.GetItem)
				media.GET("/latest", embyHandler.GetLatestItems)
				media.GET("/image/:id", embyHandler.GetImageURL)
			}

			// 卡密管理
			cardKeys := authorized.Group("/card-keys")
			{
				cardKeys.GET("", cardKeyHandler.List)
				cardKeys.POST("", middleware.PermissionMiddleware("cardkey:create"), cardKeyHandler.Create)
				cardKeys.GET("/statistics", cardKeyHandler.GetStatistics)
				cardKeys.POST("/use-vip", cardKeyHandler.UseVipCard) // 使用VIP升级码
				cardKeys.GET("/:id", cardKeyHandler.GetByID)
				cardKeys.PUT("/:id/disable", middleware.PermissionMiddleware("cardkey:edit"), cardKeyHandler.Disable)
				cardKeys.PUT("/:id/enable", middleware.PermissionMiddleware("cardkey:edit"), cardKeyHandler.Enable)
				cardKeys.DELETE("/:id", middleware.PermissionMiddleware("cardkey:delete"), cardKeyHandler.Delete)
			}
		}

		// 公开接口（无需认证）
		api.POST("/card-keys/validate", cardKeyHandler.Validate)
	}

	// 静态文件服务（支持多种运行环境）
	frontendDirs := []string{"frontend", "../frontend/dist", "./frontend"}
	var frontendDir string
	for _, dir := range frontendDirs {
		if _, err := os.Stat(dir); err == nil {
			frontendDir = dir
			break
		}
	}
	if frontendDir != "" {
		r.Static("/assets", frontendDir+"/assets")
		r.StaticFile("/favicon.ico", frontendDir+"/favicon.ico")
		r.NoRoute(func(c *gin.Context) {
			c.File(frontendDir + "/index.html")
		})
	}

	return r
}
