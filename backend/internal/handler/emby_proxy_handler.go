// Package handler Emby反向代理处理器
package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EmbyProxyHandler Emby代理处理器
type EmbyProxyHandler struct {
	db             *gorm.DB
	settingService *service.SettingService
}

// NewEmbyProxyHandler 创建Emby代理处理器
func NewEmbyProxyHandler(db *gorm.DB) *EmbyProxyHandler {
	return &EmbyProxyHandler{
		db:             db,
		settingService: service.NewSettingService(db),
	}
}

// ProxyEmby 代理Emby请求
func (h *EmbyProxyHandler) ProxyEmby(c *gin.Context) {
	// 获取Emby设置
	embySettings, err := h.settingService.GetEmbySettings()
	if err != nil || !embySettings.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Emby服务未启用"})
		return
	}

	// 获取客户端白名单设置
	clientSettings, err := h.settingService.GetClientWhitelistSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取客户端设置失败"})
		return
	}

	// 获取原始路径
	originalPath := c.Param("path")
	if originalPath == "" {
		originalPath = "/"
	}

	// 调试日志：记录所有请求
	log.Printf("代理请求: method=%s, path=%s", c.Request.Method, originalPath)

	// 解析客户端名称和用户ID
	authHeader := c.GetHeader("X-Emby-Authorization")
	clientName := c.GetHeader("X-Emby-Client")
	if clientName == "" && authHeader != "" {
		clientName = parseClientFromAuth(authHeader)
	}

	// 检查客户端白名单
	if clientSettings.Enabled && clientName != "" && !h.isClientAllowed(clientSettings, clientName) {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("客户端 %s 未授权访问", clientName)})
		return
	}

	// 检查播放限制（仅对播放请求生效）
	if h.isPlaybackRequest(originalPath) {
		userID := parseUserIDFromAuth(authHeader)
		log.Printf("检测到播放请求: path=%s, userID=%s", originalPath, userID)
		if userID != "" {
			blocked, msg := h.checkPlaybackLimit(userID)
			if blocked {
				log.Printf("播放请求被阻止: 用户ID=%s, 路径=%s, 原因=%s", userID, originalPath, msg)
				c.JSON(http.StatusForbidden, gin.H{"error": msg})
				return
			}
		}
	}

	// 解析目标URL
	targetURL, err := url.Parse(embySettings.BaseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Emby服务器地址配置错误"})
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = originalPath
		req.URL.RawQuery = c.Request.URL.RawQuery
		req.Host = targetURL.Host
		if req.URL.Query().Get("api_key") == "" && req.Header.Get("X-Emby-Token") == "" {
			q := req.URL.Query()
			q.Set("api_key", embySettings.APIKey)
			req.URL.RawQuery = q.Encode()
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		c.JSON(http.StatusBadGateway, gin.H{"error": "无法连接到Emby服务器"})
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("X-Powered-By")
		return nil
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// isClientAllowed 检查客户端是否允许
func (h *EmbyProxyHandler) isClientAllowed(settings *service.ClientWhitelistSettings, clientName string) bool {
	for _, client := range settings.Clients {
		if !client.Enabled {
			continue
		}
		if client.Name == clientName || strings.HasPrefix(clientName, client.Name) {
			return true
		}
	}
	return false
}

// parseClientFromAuth 从Authorization头解析客户端名称
func parseClientFromAuth(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Client=") || strings.HasPrefix(part, "MediaBrowser Client=") {
			idx := strings.Index(part, "Client=")
			if idx >= 0 {
				value := part[idx+7:]
				return strings.Trim(value, "\"")
			}
		}
	}
	return ""
}

// parseUserIDFromAuth 从Authorization头解析用户ID
func parseUserIDFromAuth(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "UserId=") {
			idx := strings.Index(part, "UserId=")
			if idx >= 0 {
				value := part[idx+7:]
				return strings.Trim(value, "\"")
			}
		}
	}
	return ""
}

// isPlaybackRequest 检查是否是播放请求
func (h *EmbyProxyHandler) isPlaybackRequest(path string) bool {
	playbackPatterns := []string{
		`^/Videos/[^/]+/stream`,
		`^/Audio/[^/]+/stream`,
		`^/Items/[^/]+/PlaybackInfo`,
		`^/Sessions/Playing$`,
		`^/emby/Videos/[^/]+/stream`,
		`^/emby/Audio/[^/]+/stream`,
	}
	for _, pattern := range playbackPatterns {
		if matched, _ := regexp.MatchString(pattern, path); matched {
			return true
		}
	}
	return false
}

// checkPlaybackLimit 检查用户播放限制
func (h *EmbyProxyHandler) checkPlaybackLimit(userID string) (bool, string) {
	limitSettings, err := h.settingService.GetPlayLimitSettings()
	if err != nil {
		log.Printf("获取播放限制设置失败: %v", err)
		return false, ""
	}
	if !limitSettings.Enabled || limitSettings.MaxPlaying <= 0 {
		return false, ""
	}

	adapter := service.GetMediaAdapterFromDB(h.db)
	if adapter == nil {
		log.Printf("无法获取媒体适配器")
		return false, ""
	}

	sessions, err := adapter.GetSessionsByUserID(userID)
	if err != nil {
		log.Printf("获取用户会话失败: %v", err)
		return false, ""
	}

	playingCount := 0
	for _, session := range sessions {
		if session.IsPlaying {
			playingCount++
		}
	}
	log.Printf("用户 %s 当前播放数: %d, 限制: %d", userID, playingCount, limitSettings.MaxPlaying)

	if playingCount >= limitSettings.MaxPlaying {
		return true, fmt.Sprintf("已达到最大同时播放数量限制（%d），请先停止其他设备的播放", limitSettings.MaxPlaying)
	}
	return false, ""
}

// GetEmbyProxySettings 获取Emby代理设置
func (h *EmbyProxyHandler) GetEmbyProxySettings(c *gin.Context) {
	embySettings, err := h.settingService.GetEmbySettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取设置失败"})
		return
	}
	clientSettings, err := h.settingService.GetClientWhitelistSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取客户端设置失败"})
		return
	}
	var names []string
	for _, client := range clientSettings.Clients {
		if client.Enabled {
			names = append(names, client.Name)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"enabled":          embySettings.Enabled,
			"proxy_url":        "/emby",
			"client_whitelist": clientSettings.Enabled,
			"allowed_clients":  names,
		},
	})
}

// StreamProxy 流媒体代理
func (h *EmbyProxyHandler) StreamProxy(c *gin.Context) {
	embySettings, err := h.settingService.GetEmbySettings()
	if err != nil || !embySettings.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Emby服务未启用"})
		return
	}

	originalPath := c.Param("path")
	targetURL := fmt.Sprintf("%s%s", embySettings.BaseURL, originalPath)
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}
	if !strings.Contains(targetURL, "api_key=") {
		if strings.Contains(targetURL, "?") {
			targetURL += "&api_key=" + embySettings.APIKey
		} else {
			targetURL += "?api_key=" + embySettings.APIKey
		}
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
		return
	}
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "连接Emby服务器失败"})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}
