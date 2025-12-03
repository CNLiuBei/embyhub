package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"embyhub/config"
	"embyhub/internal/handler"
	"embyhub/internal/router"
	"embyhub/internal/task"
	"embyhub/internal/util"
	"embyhub/pkg/database"
	"embyhub/pkg/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// 检查配置文件是否存在
	configPath := "config/config.yaml"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，进入安装模式
		log.Println("配置文件不存在，启动安装向导模式...")
		startSetupMode()
		return
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化日志
	if err := util.InitLogger(cfg.Log.Level, cfg.Log.Filename); err != nil {
		log.Fatal("初始化日志失败:", err)
	}
	defer util.Logger.Sync()

	util.Info("Emby用户管理系统启动中...")

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		util.Fatal("初始化数据库失败")
	}
	defer database.Close()
	util.Info("数据库连接成功")

	// 初始化Redis
	if err := redis.InitRedis(&cfg.Redis); err != nil {
		util.Fatal("初始化Redis失败")
	}
	defer redis.Close()
	util.Info("Redis连接成功")

	// 设置Gin模式
	// gin.SetMode(cfg.Server.Mode)

	// 启动后台同步任务（每5分钟同步一次Emby用户）
	syncTask := task.NewSyncTask(5 * time.Minute)
	syncTask.Start()
	defer syncTask.Stop()
	util.Info("后台同步任务已启动")

	// 启动VIP到期检查任务（每小时检查一次）
	vipTask := task.NewVipTask(1 * time.Hour)
	vipTask.Start()
	defer vipTask.Stop()
	util.Info("VIP到期检查任务已启动")

	// 启动数据清理任务（每天凌晨执行）
	cleanupTask := task.NewCleanupTask(24 * time.Hour)
	cleanupTask.Start()
	defer cleanupTask.Stop()
	util.Info("数据清理任务已启动")

	// 设置路由
	r := router.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	util.Info(fmt.Sprintf("服务器启动在 %s", addr))

	if err := r.Run(addr); err != nil {
		util.Fatal("服务器启动失败")
	}
}

// startSetupMode 安装模式 - 只提供安装向导 API
func startSetupMode() {
	r := gin.Default()

	// 简单跨域中间件（不依赖配置）
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 安装向导路由
	setupHandler := handler.NewSetupHandler()
	setup := r.Group("/api/setup")
	{
		setup.GET("/status", setupHandler.CheckStatus)
		setup.GET("/config", setupHandler.GetDefaultConfig)
		setup.POST("/verify-license", setupHandler.VerifyLicense)
		setup.POST("/test-database", setupHandler.TestDatabase)
		setup.POST("/test-emby", setupHandler.TestEmby)
		setup.POST("/test-email", setupHandler.TestEmail)
		setup.POST("/finish", setupHandler.FinishSetup)
	}

	// 静态文件服务（支持多个可能的前端目录）
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
		log.Printf("静态文件目录: %s", frontendDir)
	}

	log.Println("安装向导服务启动在 :8080")
	log.Println("请访问 http://localhost:8080/setup 进行系统初始化")

	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
