package handler

import (
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DomainHandler 域名验证处理器
type DomainHandler struct {
	settingService *service.SettingService
}

// NewDomainHandler 创建域名验证处理器
func NewDomainHandler(db *gorm.DB) *DomainHandler {
	return &DomainHandler{
		settingService: service.NewSettingService(db),
	}
}

// CheckDomainRequest 域名检查请求
type CheckDomainRequest struct {
	Domain string `json:"domain" binding:"required"`
}

// CheckDomain 检查域名是否在白名单中（公开接口，不需要认证）
func (h *DomainHandler) CheckDomain(c *gin.Context) {
	var req CheckDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数错误")
		return
	}

	// 检查域名是否允许访问
	allowed := h.settingService.IsDomainAllowed(req.Domain)

	if !allowed {
		response.Error(c, 403, "域名未授权访问")
		return
	}

	response.Success(c, gin.H{
		"allowed": true,
		"domain":  req.Domain,
	})
}
