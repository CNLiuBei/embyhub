// Package handler 通知处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler 创建通知处理器
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// GetNotifications 获取通知列表
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	page := 1
	pageSize := 20
	onlyUnread := false
	c.ShouldBindQuery(&struct {
		Page       *int  `form:"page"`
		PageSize   *int  `form:"page_size"`
		OnlyUnread *bool `form:"only_unread"`
	}{&page, &pageSize, &onlyUnread})

	notifications, total, err := h.notificationService.GetNotifications(userID, page, pageSize, onlyUnread)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, notifications, total, page, pageSize)
}

// GetUnreadCount 获取未读数量
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	count := h.notificationService.GetUnreadCount(userID)
	response.Success(c, gin.H{"count": count})
}

// MarkAsRead 标记为已读
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.notificationService.MarkAsRead(userID, req.ID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// MarkAllAsRead 全部标记为已读
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	if err := h.notificationService.MarkAllAsRead(userID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "全部已读", nil)
}

// DeleteNotification 删除通知
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.notificationService.DeleteNotification(userID, req.ID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
