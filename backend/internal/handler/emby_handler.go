package handler

import (
	"embyhub/internal/service"
	"embyhub/internal/util"

	"github.com/gin-gonic/gin"
)

type EmbyHandler struct {
	embyService *service.EmbyService
	authService *service.AuthService
}

func NewEmbyHandler() *EmbyHandler {
	return &EmbyHandler{
		embyService: service.NewEmbyService(),
		authService: service.NewAuthService(),
	}
}

// TestConnection 测试Emby连接
func (h *EmbyHandler) TestConnection(c *gin.Context) {
	if err := h.embyService.TestConnection(); err != nil {
		util.BadRequestResponse(c, "Emby连接失败: "+err.Error())
		return
	}

	util.SuccessWithMessage(c, "Emby连接成功", nil)
}

// SyncUsers 同步Emby用户
func (h *EmbyHandler) SyncUsers(c *gin.Context) {
	count, err := h.embyService.SyncUsers()
	if err != nil {
		util.BadRequestResponse(c, "同步失败: "+err.Error())
		return
	}

	util.SuccessWithMessage(c, "同步成功", map[string]interface{}{
		"sync_count": count,
	})
}

// GetUsers 获取Emby用户列表
func (h *EmbyHandler) GetUsers(c *gin.Context) {
	users, err := h.embyService.GetUsers()
	if err != nil {
		util.BadRequestResponse(c, "获取Emby用户失败: "+err.Error())
		return
	}

	util.SuccessResponse(c, users)
}

// ========== 媒体库相关接口 ==========

// GetLibraries 获取媒体库列表（使用用户视图API保持与Emby一致的顺序）
func (h *EmbyHandler) GetLibraries(c *gin.Context) {
	// 获取当前登录用户的 EmbyUserID
	var embyUserId string
	userID, exists := c.Get("user_id")
	if exists {
		user, err := h.authService.GetUserByID(userID.(int))
		if err == nil && user != nil && user.EmbyUserID != "" {
			embyUserId = user.EmbyUserID
		}
	}

	libraries, err := h.embyService.GetLibraries(embyUserId)
	if err != nil {
		util.BadRequestResponse(c, "获取媒体库失败: "+err.Error())
		return
	}

	util.SuccessResponse(c, libraries)
}

// GetItems 获取媒体项目列表
func (h *EmbyHandler) GetItems(c *gin.Context) {
	parentId := c.Query("parent_id")
	itemType := c.Query("type")
	sortBy := c.DefaultQuery("sort_by", "DateCreated")
	sortOrder := c.DefaultQuery("sort_order", "Descending")
	searchTerm := c.Query("search") // 搜索关键词

	page := util.GetQueryInt(c, "page", 1)
	pageSize := util.GetQueryInt(c, "page_size", 20)
	startIndex := (page - 1) * pageSize

	result, err := h.embyService.GetItems(parentId, itemType, startIndex, pageSize, sortBy, sortOrder, searchTerm)
	if err != nil {
		util.BadRequestResponse(c, "获取媒体列表失败: "+err.Error())
		return
	}

	util.SuccessResponse(c, map[string]interface{}{
		"list":      result.Items,
		"total":     result.TotalRecordCount,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetItem 获取单个媒体详情
func (h *EmbyHandler) GetItem(c *gin.Context) {
	itemId := c.Param("id")
	if itemId == "" {
		util.BadRequestResponse(c, "缺少媒体ID")
		return
	}

	item, err := h.embyService.GetItem(itemId)
	if err != nil {
		util.BadRequestResponse(c, "获取媒体详情失败: "+err.Error())
		return
	}

	util.SuccessResponse(c, item)
}

// GetLatestItems 获取最新媒体
func (h *EmbyHandler) GetLatestItems(c *gin.Context) {
	// 获取当前登录用户的 EmbyUserID
	userID, exists := c.Get("user_id")
	if !exists {
		util.UnauthorizedResponse(c, "用户未登录")
		return
	}

	user, err := h.authService.GetUserByID(userID.(int))
	if err != nil || user == nil {
		util.BadRequestResponse(c, "获取用户信息失败")
		return
	}

	if user.EmbyUserID == "" {
		util.BadRequestResponse(c, "当前用户未绑定Emby账号")
		return
	}

	parentId := c.Query("parent_id")
	limit := util.GetQueryInt(c, "limit", 20)

	items, err := h.embyService.GetLatestItems(user.EmbyUserID, parentId, limit)
	if err != nil {
		util.BadRequestResponse(c, "获取最新媒体失败: "+err.Error())
		return
	}

	util.SuccessResponse(c, items)
}

// GetImageURL 获取媒体图片URL
func (h *EmbyHandler) GetImageURL(c *gin.Context) {
	itemId := c.Param("id")
	imageType := c.DefaultQuery("type", "Primary")
	tag := c.Query("tag")

	url := h.embyService.GetImageURL(itemId, imageType, tag)
	util.SuccessResponse(c, map[string]string{"url": url})
}

// GetServerURL 获取Emby服务器URL（公开接口，用于前端显示图片）
func (h *EmbyHandler) GetServerURL(c *gin.Context) {
	url := h.embyService.GetServerURL()
	util.SuccessResponse(c, map[string]string{"server_url": url})
}
