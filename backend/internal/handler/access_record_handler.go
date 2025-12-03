package handler

import (
	"embyhub/internal/model"
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type AccessRecordHandler struct {
	recordService *service.AccessRecordService
}

func NewAccessRecordHandler() *AccessRecordHandler {
	return &AccessRecordHandler{
		recordService: service.NewAccessRecordService(),
	}
}

// Create 创建访问记录
func (h *AccessRecordHandler) Create(c *gin.Context) {
	var req model.AccessRecordCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.recordService.Create(&req); err != nil {
		util.InternalErrorResponse(c, "创建访问记录失败")
		return
	}

	util.SuccessWithMessage(c, "创建访问记录成功", nil)
}

// List 获取访问记录列表
func (h *AccessRecordHandler) List(c *gin.Context) {
	var req model.AccessRecordListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		util.BadRequestResponse(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.recordService.List(&req)
	if err != nil {
		util.InternalErrorResponse(c, "获取访问记录失败")
		return
	}

	util.SuccessResponse(c, resp)
}

// GetStatistics 获取统计数据
func (h *AccessRecordHandler) GetStatistics(c *gin.Context) {
	resp, err := h.recordService.GetStatistics()
	if err != nil {
		util.InternalErrorResponse(c, "获取统计数据失败")
		return
	}

	util.SuccessResponse(c, resp)
}
