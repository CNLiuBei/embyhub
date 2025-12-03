package handler

import (
	"runtime"
	"time"

	"embyhub/pkg/database"
	"embyhub/pkg/redis"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	startTime time.Time
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`    // healthy / unhealthy
	Uptime    string            `json:"uptime"`    // 运行时长
	Timestamp string            `json:"timestamp"` // 当前时间
	Services  map[string]string `json:"services"`  // 服务状态
	System    SystemInfo        `json:"system"`    // 系统信息
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
	MemoryMB     uint64 `json:"memory_mb"`
}

// Health 健康检查接口
// @Summary 健康检查
// @Tags 系统
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	services := make(map[string]string)
	overallStatus := "healthy"

	// 检查数据库
	if err := h.checkDatabase(); err != nil {
		services["database"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		services["database"] = "healthy"
	}

	// 检查Redis
	if err := h.checkRedis(); err != nil {
		services["redis"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		services["redis"] = "healthy"
	}

	// 获取系统信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := HealthResponse{
		Status:    overallStatus,
		Uptime:    time.Since(h.startTime).Round(time.Second).String(),
		Timestamp: time.Now().Format(time.RFC3339),
		Services:  services,
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			MemoryMB:     memStats.Alloc / 1024 / 1024,
		},
	}

	if overallStatus == "healthy" {
		c.JSON(200, response)
	} else {
		c.JSON(503, response)
	}
}

// Ready 就绪检查
// @Summary 就绪检查
// @Tags 系统
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// 检查核心依赖是否就绪
	if err := h.checkDatabase(); err != nil {
		c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
		return
	}
	if err := h.checkRedis(); err != nil {
		c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"status": "ready"})
}

// Live 存活检查
// @Summary 存活检查
// @Tags 系统
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(200, gin.H{"status": "alive"})
}

// checkDatabase 检查数据库连接
func (h *HealthHandler) checkDatabase() error {
	sqlDB, err := database.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// checkRedis 检查Redis连接
func (h *HealthHandler) checkRedis() error {
	_, err := redis.Get("health_check_ping")
	// 如果是key不存在的错误，也算正常
	if err != nil && err.Error() != "redis: nil" {
		return err
	}
	return nil
}
