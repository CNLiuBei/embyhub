// Package handler 用户邀请处理器
package handler

import (
	"strconv"

	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// InviteHandler 邀请处理器
type InviteHandler struct {
	inviteService *service.InviteService
}

// NewInviteHandler 创建邀请处理器
func NewInviteHandler(inviteService *service.InviteService) *InviteHandler {
	return &InviteHandler{inviteService: inviteService}
}

// GetMyInviteInfo 获取我的邀请信息（用户端）
func (h *InviteHandler) GetMyInviteInfo(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	info, err := h.inviteService.GetUserInviteInfo(userID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, info)
}

// GetMyInvites 获取我的邀请记录（用户端）
func (h *InviteHandler) GetMyInvites(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	list, total, err := h.inviteService.GetMyInvites(userID)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"list":  list,
		"total": total,
	})
}

// GetInviteRanking 获取邀请排行榜
func (h *InviteHandler) GetInviteRanking(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	list, err := h.inviteService.GetInviteRanking(limit)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, list)
}

// Admin: GetInviteStats 获取邀请统计
func (h *InviteHandler) GetInviteStats(c *gin.Context) {
	stats := h.inviteService.GetInviteStats()
	response.Success(c, stats)
}

// Admin: SetRewardDays 设置邀请奖励天数
func (h *InviteHandler) SetRewardDays(c *gin.Context) {
	var req struct {
		Days int `json:"days" binding:"required,min=0,max=365"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.inviteService.SetInviteRewardDays(req.Days); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "设置成功", nil)
}

// Admin: GetInviteRecords 获取邀请记录列表
func (h *InviteHandler) GetInviteRecords(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	if ps := c.Query("page_size"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}

	list, total, err := h.inviteService.GetInviteRecords(page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, list, total, page, pageSize)
}
