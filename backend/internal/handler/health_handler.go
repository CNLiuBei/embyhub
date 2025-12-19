// Package handler 系统健康监控处理器
package handler

import (
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康监控处理器
type HealthHandler struct {
	healthService *service.HealthService
}

// NewHealthHandler 创建健康监控处理器
func NewHealthHandler(healthService *service.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// GetHealth 获取健康状态
func (h *HealthHandler) GetHealth(c *gin.Context) {
	status := h.healthService.GetHealth()
	response.Success(c, status)
}

// GetStats 获取系统统计
func (h *HealthHandler) GetStats(c *gin.Context) {
	stats := h.healthService.GetStats()
	response.Success(c, stats)
}
