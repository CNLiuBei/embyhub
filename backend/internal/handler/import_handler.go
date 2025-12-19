// Package handler 导入处理器
package handler

import (
	"path/filepath"
	"strings"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// ImportHandler 导入处理器
type ImportHandler struct {
	importService *service.ImportService
}

// NewImportHandler 创建导入处理器
func NewImportHandler(importService *service.ImportService) *ImportHandler {
	return &ImportHandler{importService: importService}
}

// ImportUsers 批量导入用户
func (h *ImportHandler) ImportUsers(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要导入的文件")
		return
	}

	// 检查文件大小（最大10MB）
	if file.Size > 10*1024*1024 {
		response.BadRequest(c, "文件大小不能超过10MB")
		return
	}

	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
		response.BadRequest(c, "只支持 CSV 或 Excel 文件")
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.ServerError(c, "读取文件失败")
		return
	}
	defer src.Close()

	var result *service.ImportResult

	if ext == ".csv" {
		result, err = h.importService.ImportFromCSV(src, adminID)
	} else {
		result, err = h.importService.ImportFromExcel(src, adminID)
	}

	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetTemplate 下载导入模板
func (h *ImportHandler) GetTemplate(c *gin.Context) {
	data, err := h.importService.GetImportTemplate()
	if err != nil {
		response.ServerError(c, "生成模板失败")
		return
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=user_import_template.xlsx")
	c.Data(200, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
