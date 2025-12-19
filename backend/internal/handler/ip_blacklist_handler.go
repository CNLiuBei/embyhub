// Package handler IP黑名单处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// IPBlacklistHandler IP黑名单处理器
type IPBlacklistHandler struct {
	ipBlacklistService *service.IPBlacklistService
}

// NewIPBlacklistHandler 创建IP黑名单处理器
func NewIPBlacklistHandler(ipBlacklistService *service.IPBlacklistService) *IPBlacklistHandler {
	return &IPBlacklistHandler{ipBlacklistService: ipBlacklistService}
}

// GetList 获取黑名单列表
func (h *IPBlacklistHandler) GetList(c *gin.Context) {
	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	list, total, err := h.ipBlacklistService.GetList(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total, page, pageSize)
}

// Add 添加IP到黑名单
func (h *IPBlacklistHandler) Add(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)

	var req service.AddIPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入有效的IP地址")
		return
	}

	if err := h.ipBlacklistService.Add(adminID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "添加成功", nil)
}

// Remove 从黑名单移除IP
func (h *IPBlacklistHandler) Remove(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		response.BadRequest(c, "无效的IP地址")
		return
	}

	if err := h.ipBlacklistService.Remove(ip); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "移除成功", nil)
}

// CheckIP 检查IP是否被封禁
func (h *IPBlacklistHandler) CheckIP(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		ip = c.ClientIP()
	}

	blocked := h.ipBlacklistService.IsBlocked(ip)
	response.Success(c, map[string]interface{}{
		"ip":      ip,
		"blocked": blocked,
	})
}
