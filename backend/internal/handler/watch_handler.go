// Package handler 观影记录处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// WatchHandler 观影记录处理器
type WatchHandler struct {
	watchService *service.WatchService
}

// NewWatchHandler 创建观影记录处理器
func NewWatchHandler(watchService *service.WatchService) *WatchHandler {
	return &WatchHandler{watchService: watchService}
}

// RecordWatch 记录观影
func (h *WatchHandler) RecordWatch(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.RecordWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.watchService.RecordWatch(userID, &req); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "记录成功", nil)
}

// GetWatchHistory 获取观影历史
func (h *WatchHandler) GetWatchHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	histories, total, err := h.watchService.GetWatchHistory(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, histories, total, page, pageSize)
}

// DeleteWatchHistory 删除观影记录
func (h *WatchHandler) DeleteWatchHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.watchService.DeleteWatchHistory(userID, req.ID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ClearWatchHistory 清空观影记录
func (h *WatchHandler) ClearWatchHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	if err := h.watchService.ClearWatchHistory(userID); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "清空成功", nil)
}

// AddFavorite 添加收藏
func (h *WatchHandler) AddFavorite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.watchService.AddFavorite(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "收藏成功", nil)
}

// GetFavorites 获取收藏列表
func (h *WatchHandler) GetFavorites(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	page := 1
	pageSize := 20
	c.ShouldBindQuery(&struct {
		Page     *int `form:"page"`
		PageSize *int `form:"page_size"`
	}{&page, &pageSize})

	favorites, total, err := h.watchService.GetFavorites(userID, page, pageSize)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.SuccessPage(c, favorites, total, page, pageSize)
}

// RemoveFavorite 取消收藏
func (h *WatchHandler) RemoveFavorite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req struct {
		VideoID string `json:"video_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.watchService.RemoveFavorite(userID, req.VideoID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "取消收藏成功", nil)
}

// CheckFavorite 检查是否已收藏
func (h *WatchHandler) CheckFavorite(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	videoID := c.Query("video_id")

	if videoID == "" {
		response.BadRequest(c, "video_id不能为空")
		return
	}

	isFavorite := h.watchService.IsFavorite(userID, videoID)
	response.Success(c, gin.H{"is_favorite": isFavorite})
}
