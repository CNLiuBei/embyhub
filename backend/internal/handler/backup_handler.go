// Package handler 数据备份处理器
package handler

import (
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// BackupHandler 备份处理器
type BackupHandler struct {
	backupService *service.BackupService
}

// NewBackupHandler 创建备份处理器
func NewBackupHandler(backupService *service.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// CreateBackup 创建备份
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	info, err := h.backupService.CreateBackup()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "备份创建成功", info)
}

// GetBackupList 获取备份列表
func (h *BackupHandler) GetBackupList(c *gin.Context) {
	list, err := h.backupService.GetBackupList()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// DeleteBackup 删除备份
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.BadRequest(c, "文件名不能为空")
		return
	}

	if err := h.backupService.DeleteBackup(filename); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// DownloadBackup 下载备份
func (h *BackupHandler) DownloadBackup(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.BadRequest(c, "文件名不能为空")
		return
	}

	path, err := h.backupService.GetBackupPath(filename)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.File(path)
}

// RestoreBackup 恢复备份
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		response.BadRequest(c, "文件名不能为空")
		return
	}

	if err := h.backupService.RestoreBackup(filename); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "恢复成功，请重启服务", nil)
}
