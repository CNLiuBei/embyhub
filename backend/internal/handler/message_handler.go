// Package handler 私信处理器
package handler

import (
	"net/http"
	"strconv"

	"feiniu-user-system/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PrivateMessageHandler 私信处理器
type PrivateMessageHandler struct {
	messageService *service.MessageService
}

// NewPrivateMessageHandler 创建私信处理器
func NewPrivateMessageHandler(messageService *service.MessageService) *PrivateMessageHandler {
	return &PrivateMessageHandler{messageService: messageService}
}

// SendMessage 发送私信
func (h *PrivateMessageHandler) SendMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		ToUserID string   `json:"to_user_id" binding:"required"`
		Content  string   `json:"content" binding:"required"`
		Images   []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	msg, err := h.messageService.SendMessage(userID, toUserID, req.Content, req.Images)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "发送成功", "data": msg})
}

// GetConversations 获取会话列表
func (h *PrivateMessageHandler) GetConversations(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	conversations, total, err := h.messageService.GetConversations(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  conversations,
			"total": total,
		},
	})
}

// GetMessages 获取与某用户的消息列表
func (h *PrivateMessageHandler) GetMessages(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	otherUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))

	messages, total, err := h.messageService.GetMessages(userID, otherUserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  messages,
			"total": total,
		},
	})
}

// GetUnreadCount 获取未读消息数
func (h *PrivateMessageHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	count, err := h.messageService.GetUnreadCount(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取未读数失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"count": count}})
}

// MarkAsRead 标记消息为已读
func (h *PrivateMessageHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	otherUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	if err := h.messageService.MarkAsRead(userID, otherUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "标记失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "标记成功"})
}

// DeleteMessage 删除消息
func (h *PrivateMessageHandler) DeleteMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.messageService.DeleteMessage(userID, messageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "删除成功"})
}

// FollowUser 关注用户
func (h *PrivateMessageHandler) FollowUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	followID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	followed, err := h.messageService.FollowUser(userID, followID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	msg := "关注成功"
	if !followed {
		msg = "取消关注"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"followed": followed}})
}

// GetFollowings 获取关注列表
func (h *PrivateMessageHandler) GetFollowings(c *gin.Context) {
	userIDStr := c.Param("user_id")
	var userID uuid.UUID
	var err error

	if userIDStr == "" || userIDStr == "me" {
		userID = c.MustGet("user_id").(uuid.UUID)
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.messageService.GetFollowings(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取关注列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  users,
			"total": total,
		},
	})
}

// GetFollowers 获取粉丝列表
func (h *PrivateMessageHandler) GetFollowers(c *gin.Context) {
	userIDStr := c.Param("user_id")
	var userID uuid.UUID
	var err error

	if userIDStr == "" || userIDStr == "me" {
		userID = c.MustGet("user_id").(uuid.UUID)
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
			return
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.messageService.GetFollowers(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取粉丝列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  users,
			"total": total,
		},
	})
}

// GetFollowStats 获取关注统计
func (h *PrivateMessageHandler) GetFollowStats(c *gin.Context) {
	userIDStr := c.Param("user_id")
	var userID uuid.UUID
	var err error

	if userIDStr == "" || userIDStr == "me" {
		userID = c.MustGet("user_id").(uuid.UUID)
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
			return
		}
	}

	followings, followers := h.messageService.GetFollowStats(userID)

	// 检查当前用户是否关注了该用户
	isFollowing := false
	if currentUserID, exists := c.Get("user_id"); exists {
		currentID := currentUserID.(uuid.UUID)
		if currentID != userID {
			isFollowing = h.messageService.IsFollowing(currentID, userID)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"followings":   followings,
			"followers":    followers,
			"is_following": isFollowing,
		},
	})
}

// SearchUsers 搜索用户（用于发起私信）
func (h *PrivateMessageHandler) SearchUsers(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	keyword := c.Query("keyword")
	limit := 10

	users, err := h.messageService.SearchUsers(userID, keyword, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "搜索失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": users})
}

// CanSendMessage 检查是否可以发送私信
func (h *PrivateMessageHandler) CanSendMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	toUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	canSend, reason := h.messageService.CanSendMessage(userID, toUserID)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"can_send": canSend,
			"reason":   reason,
		},
	})
}

// RecallMessage 撤回消息
func (h *PrivateMessageHandler) RecallMessage(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	messageID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := h.messageService.RecallMessage(userID, messageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "撤回成功"})
}

// BlockUser 拉黑用户
func (h *PrivateMessageHandler) BlockUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	blockedID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.messageService.BlockUser(userID, blockedID, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "拉黑成功"})
}

// UnblockUser 取消拉黑
func (h *PrivateMessageHandler) UnblockUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	blockedID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	if err := h.messageService.UnblockUser(userID, blockedID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "取消拉黑成功"})
}

// GetBlacklist 获取黑名单列表
func (h *PrivateMessageHandler) GetBlacklist(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	blacklist, total, err := h.messageService.GetBlacklist(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取黑名单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":  blacklist,
			"total": total,
		},
	})
}

// IsBlocked 检查是否被拉黑
func (h *PrivateMessageHandler) IsBlocked(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	otherUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	blocked, reason := h.messageService.IsBlocked(userID, otherUserID)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"blocked": blocked,
			"reason":  reason,
		},
	})
}

// MuteConversation 静音/取消静音会话
func (h *PrivateMessageHandler) MuteConversation(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	otherUserID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID格式错误"})
		return
	}

	muted, err := h.messageService.MuteConversation(userID, otherUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	msg := "已静音"
	if !muted {
		msg = "已取消静音"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "data": gin.H{"muted": muted}})
}
