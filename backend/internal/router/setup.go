package router

import (
	"embyhub/internal/handler"

	"github.com/gin-gonic/gin"
)

// SetupSetupRoutes 注册安装程序路由
func SetupSetupRoutes(r *gin.Engine) {
	h := handler.NewSetupHandler()

	setup := r.Group("/api/setup")
	{
		setup.GET("/status", h.CheckStatus)
		setup.GET("/config", h.GetDefaultConfig)
		setup.POST("/verify-license", h.VerifyLicense)
		setup.POST("/test-database", h.TestDatabase)
		setup.POST("/test-emby", h.TestEmby)
		setup.POST("/test-email", h.TestEmail)
		setup.POST("/finish", h.FinishSetup)
	}
}
