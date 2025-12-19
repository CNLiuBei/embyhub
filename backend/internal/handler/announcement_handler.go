// Package handler 公告处理器
package handler

import (
	"strconv"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// AnnouncementHandler 公告处理器
type AnnouncementHandler struct {
	announcementService *service.AnnouncementService
}

// NewAnnouncementHandler 创建公告处理器
func NewAnnouncementHandler(announcementService *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{announcementService: announcementService}
}

// Create 创建公告
func (h *AnnouncementHandler) Create(c *gin.Context) {
	adminID, _ := middleware.GetUserID(c)

	var req service.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	announcement, err := h.announcementService.Create(adminID, &req)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, announcement)
}

// Update 更新公告
func (h *AnnouncementHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	var req service.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.announcementService.Update(id, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// Delete 删除公告
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	if err := h.announcementService.Delete(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetList 获取公告列表（管理员）
func (h *AnnouncementHandler) GetList(c *gin.Context) {
	page := 1
	pageSize := 20
	var status *int8

	c.ShouldBindQuery(&struct {
		Page     *int  `form:"page"`
		PageSize *int  `form:"page_size"`
		Status   *int8 `form:"status"`
	}{&page, &pageSize, status})

	list, total, err := h.announcementService.GetList(page, pageSize, status)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total, page, pageSize)
}

// GetPublished 获取已发布公告（用户端）
func (h *AnnouncementHandler) GetPublished(c *gin.Context) {
	limit := 10
	c.ShouldBindQuery(&struct {
		Limit *int `form:"limit"`
	}{&limit})

	list, err := h.announcementService.GetPublished(limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetByID 获取公告详情
func (h *AnnouncementHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	announcement, err := h.announcementService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, announcement)
}

// Publish 发布公告
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	if err := h.announcementService.Publish(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "发布成功", nil)
}

// Offline 下线公告
func (h *AnnouncementHandler) Offline(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的公告ID")
		return
	}

	if err := h.announcementService.Offline(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "下线成功", nil)
}
