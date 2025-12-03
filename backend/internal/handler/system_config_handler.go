package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type SystemConfigHandler struct {
	configService *service.SystemConfigService
}

func NewSystemConfigHandler() *SystemConfigHandler {
	return &SystemConfigHandler{
		configService: service.NewSystemConfigService(),
	}
}

// List 获取系统配置列表
func (h *SystemConfigHandler) List(c *gin.Context) {
	resp, err := h.configService.List()
	if err != nil {
		util.InternalErrorResponse(c, "获取系统配置失败")
		return
	}

	util.SuccessResponse(c, resp)
}

// Update 更新系统配置
func (h *SystemConfigHandler) Update(c *gin.Context) {
	configKey := c.Param("key")

	var req model.SystemConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.configService.Update(configKey, req.ConfigValue); err != nil {
		util.BadRequestResponse(c, err.Error())
		return
	}

	util.SuccessWithMessage(c, "更新配置成功", nil)
}
