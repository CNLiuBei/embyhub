// Package handler 统计处理器
package handler

import (
	"strconv"

	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	statsService *service.StatsService
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(statsService *service.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

// GetUserStats 获取用户统计
func (h *StatsHandler) GetUserStats(c *gin.Context) {
	stats := h.statsService.GetUserStats()
	response.Success(c, stats)
}

// GetDailyStats 获取每日统计
func (h *StatsHandler) GetDailyStats(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		days, _ = strconv.Atoi(d)
	}
	stats := h.statsService.GetDailyStats(days)
	response.Success(c, stats)
}

// GetVisitRanking 获取访问排行
func (h *StatsHandler) GetVisitRanking(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	ranking := h.statsService.GetVisitRanking(limit)
	response.Success(c, ranking)
}
