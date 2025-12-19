// Package handler API处理层
package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/email"
	"feiniu-user-system/pkg/emby"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// createMediaAdapterFromSettings 从设置创建媒体适配器
func createMediaAdapterFromSettings(settings *service.EmbySettings) emby.MediaAdapter {
	if settings.IsEmbyMode() {
		return emby.NewEmbyAdapter(&emby.EmbyAdapterConfig{
			BaseURL: settings.BaseURL,
			APIKey:  settings.APIKey,
		})
	}
	return emby.NewFeiniuAdapter(&emby.FeiniuAdapterConfig{
		BaseURL:   settings.BaseURL,
		AdminUser: settings.AdminUser,
		AdminPass: settings.AdminPass,
	})
}

// SettingHandler 设置处理器
type SettingHandler struct {
	service *service.SettingService
}

// NewSettingHandler 创建设置处理器实例
func NewSettingHandler(service *service.SettingService) *SettingHandler {
	return &SettingHandler{service: service}
}

// ============= 邮件设置 =============

// GetEmailSettings 获取邮件设置
func (h *SettingHandler) GetEmailSettings(c *gin.Context) {
	settings, err := h.service.GetEmailSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	// 不返回密码
	settings.Password = ""
	response.Success(c, settings)
}

// SaveEmailSettings 保存邮件设置
func (h *SettingHandler) SaveEmailSettings(c *gin.Context) {
	var req service.EmailSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveEmailSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// TestEmailSettings 测试邮件设置
func (h *SettingHandler) TestEmailSettings(c *gin.Context) {
	var req struct {
		To string `json:"to" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取邮件设置
	settings, err := h.service.GetEmailSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	if !settings.Enabled {
		response.Error(c, 400, "邮件服务未启用")
		return
	}

	// 创建邮件服务并发送测试邮件
	emailService := email.NewServiceWithConfig(&email.Config{
		Host:     settings.Host,
		Port:     settings.Port,
		Username: settings.Username,
		Password: settings.Password,
		From:     settings.From,
		FromName: settings.FromName,
	})

	subject := "测试邮件"
	body := "这是一封测试邮件，用于验证邮件服务配置是否正确。如果您收到此邮件，说明配置成功！"

	if err := emailService.SendHTML(req.To, subject, body); err != nil {
		response.Error(c, 500, "发送失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 域名白名单设置 =============

// GetDomainSettings 获取域名白名单设置
func (h *SettingHandler) GetDomainSettings(c *gin.Context) {
	settings, err := h.service.GetDomainSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveDomainSettings 保存域名白名单设置
func (h *SettingHandler) SaveDomainSettings(c *gin.Context) {
	var req service.DomainSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证域名列表不为空
	if req.Enabled && len(req.Domains) == 0 {
		response.BadRequest(c, "启用域名白名单时，至少需要添加一个域名")
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveDomainSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 注册设置 =============

// GetRegisterSettings 获取注册设置
func (h *SettingHandler) GetRegisterSettings(c *gin.Context) {
	settings, err := h.service.GetRegisterSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveRegisterSettings 保存注册设置
func (h *SettingHandler) SaveRegisterSettings(c *gin.Context) {
	var req service.RegisterSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveRegisterSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}


// ============= Emby/媒体服务设置 =============

// GetEmbySettings 获取Emby设置
func (h *SettingHandler) GetEmbySettings(c *gin.Context) {
	settings, err := h.service.GetEmbySettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	// 不返回敏感信息
	settings.APIKey = maskString(settings.APIKey)
	settings.AdminPass = ""
	response.Success(c, settings)
}

// SaveEmbySettings 保存Emby设置
func (h *SettingHandler) SaveEmbySettings(c *gin.Context) {
	var req service.EmbySettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必填字段
	if req.Enabled {
		if req.BaseURL == "" {
			response.BadRequest(c, "请填写服务器地址")
			return
		}
		if req.Mode == "emby" || req.Mode == "" {
			// Emby模式需要API Key
			// 如果是掩码值，保留原来的值
			if req.APIKey == "" || req.APIKey == maskString(req.APIKey) {
				oldSettings, _ := h.service.GetEmbySettings()
				if oldSettings != nil && oldSettings.APIKey != "" {
					req.APIKey = oldSettings.APIKey
				}
			}
			if req.APIKey == "" {
				response.BadRequest(c, "Emby模式需要填写API密钥")
				return
			}
		} else if req.Mode == "feiniu" {
			// 飞牛模式需要用户名密码
			if req.AdminUser == "" {
				response.BadRequest(c, "飞牛模式需要填写管理员用户名")
				return
			}
			// 如果密码为空，保留原来的值
			if req.AdminPass == "" {
				oldSettings, _ := h.service.GetEmbySettings()
				if oldSettings != nil && oldSettings.AdminPass != "" {
					req.AdminPass = oldSettings.AdminPass
				}
			}
			if req.AdminPass == "" {
				response.BadRequest(c, "飞牛模式需要填写管理员密码")
				return
			}
		}
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveEmbySettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// TestEmbyConnection 测试Emby连接
func (h *SettingHandler) TestEmbyConnection(c *gin.Context) {
	settings, err := h.service.GetEmbySettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	if !settings.Enabled {
		response.Error(c, 400, "媒体服务未启用")
		return
	}

	// 创建适配器并测试连接
	adapter := createMediaAdapterFromSettings(settings)
	if adapter == nil {
		response.Error(c, 500, "创建媒体服务连接失败")
		return
	}

	// 尝试获取用户列表来测试连接
	users, err := adapter.GetUserList()
	if err != nil {
		response.Error(c, 500, "连接失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message":    "连接成功",
		"user_count": len(users),
	})
}

// maskString 掩码字符串
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

// ============= 客户端白名单设置 =============

// GetClientWhitelistSettings 获取客户端白名单设置
func (h *SettingHandler) GetClientWhitelistSettings(c *gin.Context) {
	settings, err := h.service.GetClientWhitelistSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveClientWhitelistSettings 保存客户端白名单设置
func (h *SettingHandler) SaveClientWhitelistSettings(c *gin.Context) {
	var req service.ClientWhitelistSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveClientWhitelistSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// AddClientToWhitelist 添加客户端到白名单
func (h *SettingHandler) AddClientToWhitelist(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.AddClientToWhitelist(req.Name, req.DisplayName, userID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveClientFromWhitelist 从白名单移除客户端
func (h *SettingHandler) RemoveClientFromWhitelist(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "请提供客户端名称")
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.RemoveClientFromWhitelist(name, userID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// UpdateClientStatus 更新客户端状态
func (h *SettingHandler) UpdateClientStatus(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		response.Error(c, 400, "请提供客户端名称")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.UpdateClientStatus(name, req.Enabled, userID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 会话限制设置 =============

// GetSessionLimitSettings 获取会话限制设置
func (h *SettingHandler) GetSessionLimitSettings(c *gin.Context) {
	settings, err := h.service.GetSessionLimitSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveSessionLimitSettings 保存会话限制设置
func (h *SettingHandler) SaveSessionLimitSettings(c *gin.Context) {
	var req service.SessionLimitSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveSessionLimitSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 网站设置 =============

// GetSiteSettings 获取网站设置
func (h *SettingHandler) GetSiteSettings(c *gin.Context) {
	settings, err := h.service.GetSiteSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveSiteSettings 保存网站设置
func (h *SettingHandler) SaveSiteSettings(c *gin.Context) {
	var req service.SiteSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveSiteSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetSiteSettingsPublic 获取网站设置（公开接口，无需登录）
func (h *SettingHandler) GetSiteSettingsPublic(c *gin.Context) {
	settings, err := h.service.GetSiteSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// UploadLogo 上传网站 Logo
func (h *SettingHandler) UploadLogo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 检查文件大小（最大 2MB）
	if file.Size > 2*1024*1024 {
		response.BadRequest(c, "文件大小不能超过 2MB")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true, ".webp": true, ".ico": true}
	if !allowedExts[ext] {
		response.BadRequest(c, "只支持 PNG、JPG、GIF、SVG、WebP、ICO 格式")
		return
	}

	// 确保上传目录存在
	uploadDir := "./uploads/logo"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.ServerError(c, "创建上传目录失败")
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("logo_%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		response.ServerError(c, "保存文件失败")
		return
	}

	// 返回访问 URL
	logoURL := "/uploads/logo/" + filename
	response.Success(c, gin.H{
		"url": logoURL,
	})
}

// ============= 播放限制设置 =============

// GetPlayLimitSettings 获取播放限制设置
func (h *SettingHandler) GetPlayLimitSettings(c *gin.Context) {
	settings, err := h.service.GetPlayLimitSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SavePlayLimitSettings 保存播放限制设置
func (h *SettingHandler) SavePlayLimitSettings(c *gin.Context) {
	var req service.PlayLimitSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SavePlayLimitSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 用户清理设置 =============

// GetUserCleanupSettings 获取用户清理设置
func (h *SettingHandler) GetUserCleanupSettings(c *gin.Context) {
	settings, err := h.service.GetUserCleanupSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveUserCleanupSettings 保存用户清理设置
func (h *SettingHandler) SaveUserCleanupSettings(c *gin.Context) {
	var req service.UserCleanupSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveUserCleanupSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// ============= 充值链接设置 =============

// GetRechargeLinksSettings 获取充值链接设置
func (h *SettingHandler) GetRechargeLinksSettings(c *gin.Context) {
	settings, err := h.service.GetRechargeLinksSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveRechargeLinksSettings 保存充值链接设置
func (h *SettingHandler) SaveRechargeLinksSettings(c *gin.Context) {
	var req service.RechargeLinksSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveRechargeLinksSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetRechargeLinksPublic 获取充值链接（公开接口，无需登录）
func (h *SettingHandler) GetRechargeLinksPublic(c *gin.Context) {
	settings, err := h.service.GetRechargeLinksSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	// 只返回启用的链接
	enabledLinks := make([]service.RechargeLink, 0)
	for _, link := range settings.Links {
		if link.Enabled && link.URL != "" {
			enabledLinks = append(enabledLinks, link)
		}
	}

	response.Success(c, gin.H{
		"links": enabledLinks,
	})
}


// ============= 积分卡购买链接设置 =============

// GetPointsRechargeLinksSettings 获取积分卡购买链接设置
func (h *SettingHandler) GetPointsRechargeLinksSettings(c *gin.Context) {
	settings, err := h.service.GetPointsRechargeLinksSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SavePointsRechargeLinksSettings 保存积分卡购买链接设置
func (h *SettingHandler) SavePointsRechargeLinksSettings(c *gin.Context) {
	var req service.PointsRechargeLinksSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SavePointsRechargeLinksSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetPointsRechargeLinksPublic 获取积分卡购买链接（公开接口，无需登录）
func (h *SettingHandler) GetPointsRechargeLinksPublic(c *gin.Context) {
	settings, err := h.service.GetPointsRechargeLinksSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	// 只返回启用的链接
	enabledLinks := make([]service.PointsRechargeLink, 0)
	for _, link := range settings.Links {
		if link.Enabled && link.URL != "" {
			enabledLinks = append(enabledLinks, link)
		}
	}

	response.Success(c, gin.H{
		"links": enabledLinks,
	})
}

// ============= 图床设置 =============

// GetImageHostSettings 获取图床设置
func (h *SettingHandler) GetImageHostSettings(c *gin.Context) {
	settings, err := h.service.GetImageHostSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}

// SaveImageHostSettings 保存图床设置
func (h *SettingHandler) SaveImageHostSettings(c *gin.Context) {
	var req service.ImageHostSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetUserID(c)
	if err := h.service.SaveImageHostSettings(&req, userID); err != nil {
		response.Error(c, 500, "保存设置失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// GetImageHostSettingsPublic 获取图床设置（公开接口，无需登录）
func (h *SettingHandler) GetImageHostSettingsPublic(c *gin.Context) {
	settings, err := h.service.GetImageHostSettings()
	if err != nil {
		response.Error(c, 500, "获取设置失败: "+err.Error())
		return
	}

	response.Success(c, settings)
}
