// Package handler Cloudflare Tunnel 处理器
package handler

import (
	"net/http"

	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// CloudflareTunnelHandler 隧道处理器
type CloudflareTunnelHandler struct {
	tunnelService *service.CloudflareTunnelService
}

// NewCloudflareTunnelHandler 创建处理器
func NewCloudflareTunnelHandler(tunnelService *service.CloudflareTunnelService) *CloudflareTunnelHandler {
	return &CloudflareTunnelHandler{
		tunnelService: tunnelService,
	}
}

// GetStatus 获取隧道状态
func (h *CloudflareTunnelHandler) GetStatus(c *gin.Context) {
	status, err := h.tunnelService.GetStatus()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, status)
}

// GetConfig 获取隧道配置
func (h *CloudflareTunnelHandler) GetConfig(c *gin.Context) {
	config, err := h.tunnelService.GetConfig()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, config)
}

// DownloadCloudflared 下载 cloudflared 二进制文件
func (h *CloudflareTunnelHandler) DownloadCloudflared(c *gin.Context) {
	// 获取下载信息
	downloadInfo := h.tunnelService.GetDownloadInfo()

	// 检查是否已下载
	if h.tunnelService.IsCloudflaredDownloaded() {
		version, _ := h.tunnelService.GetCloudflaredVersion()
		response.SuccessWithMessage(c, "cloudflared 已存在", gin.H{
			"path":     h.tunnelService.GetCloudflaredBinPath(),
			"version":  version,
			"os":       downloadInfo.OS,
			"arch":     downloadInfo.Arch,
			"fileName": downloadInfo.FileName,
		})
		return
	}

	// 下载 cloudflared（同步下载，前端可以显示加载状态）
	err := h.tunnelService.DownloadCloudflared(nil)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "下载失败: "+err.Error())
		return
	}

	version, _ := h.tunnelService.GetCloudflaredVersion()
	response.SuccessWithMessage(c, "cloudflared 下载成功", gin.H{
		"path":     h.tunnelService.GetCloudflaredBinPath(),
		"version":  version,
		"os":       downloadInfo.OS,
		"arch":     downloadInfo.Arch,
		"fileName": downloadInfo.FileName,
	})
}

// CreateTunnel 创建隧道
func (h *CloudflareTunnelHandler) CreateTunnel(c *gin.Context) {
	var req service.CreateTunnelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := h.tunnelService.CreateTunnel(&req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 如果需要授权，返回授权URL
	if result.NeedAuth {
		response.Success(c, gin.H{
			"need_auth": true,
			"auth_url":  result.AuthURL,
			"message":   result.Message,
		})
		return
	}

	// 创建成功，返回配置
	response.Success(c, gin.H{
		"need_auth": false,
		"config":    result.Config,
		"message":   result.Message,
	})
}

// StartTunnel 启动隧道
func (h *CloudflareTunnelHandler) StartTunnel(c *gin.Context) {
	if err := h.tunnelService.StartTunnel(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(c, "隧道已启动", nil)
}

// StopTunnel 停止隧道
func (h *CloudflareTunnelHandler) StopTunnel(c *gin.Context) {
	if err := h.tunnelService.StopTunnel(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(c, "隧道已停止", nil)
}

// RestartTunnel 重启隧道
func (h *CloudflareTunnelHandler) RestartTunnel(c *gin.Context) {
	if err := h.tunnelService.RestartTunnel(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(c, "隧道已重启", nil)
}

// DeleteTunnel 删除隧道
func (h *CloudflareTunnelHandler) DeleteTunnel(c *gin.Context) {
	if err := h.tunnelService.DeleteTunnel(); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.SuccessWithMessage(c, "隧道已删除", nil)
}

// Login 获取 Cloudflare 授权URL
func (h *CloudflareTunnelHandler) Login(c *gin.Context) {
	result, err := h.tunnelService.Login()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if result.Success {
		response.SuccessWithMessage(c, result.Message, gin.H{
			"logged_in": true,
		})
		return
	}

	// 返回授权URL
	response.Success(c, gin.H{
		"logged_in": false,
		"auth_url":  result.AuthURL,
		"message":   result.Message,
	})
}
