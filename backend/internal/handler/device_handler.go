// Package handler 设备管理处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// DeviceHandler 设备管理处理器
type DeviceHandler struct {
	deviceService *service.DeviceService
}

// NewDeviceHandler 创建设备管理处理器
func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

// GetDevices 获取设备列表
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	// 从请求头获取当前设备ID（如果有的话）
	currentDeviceID := c.GetHeader("X-Device-ID")

	devices, err := h.deviceService.GetUserDevices(userID, currentDeviceID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, devices)
}

// RemoveDevice 移除设备
func (h *DeviceHandler) RemoveDevice(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	deviceID := c.Param("device_id")

	if deviceID == "" {
		response.BadRequest(c, "设备ID不能为空")
		return
	}

	if err := h.deviceService.RemoveDevice(userID, deviceID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设备已移除", nil)
}

// RemoveAllDevices 移除所有其他设备
func (h *DeviceHandler) RemoveAllDevices(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	currentDeviceID := c.GetHeader("X-Device-ID")

	count, err := h.deviceService.RemoveAllDevices(userID, currentDeviceID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "已移除其他设备", map[string]int64{"removed_count": count})
}
