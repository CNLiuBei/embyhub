// Package handler 媒体处理器
package handler

import (
	"feiniu-user-system/internal/middleware"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
)

// MediaHandler 媒体处理器
type MediaHandler struct {
	mediaService *service.MediaService
}

// NewMediaHandler 创建媒体处理器
func NewMediaHandler(mediaService *service.MediaService) *MediaHandler {
	return &MediaHandler{mediaService: mediaService}
}

// GetMediaDBList 获取媒体库列表
// @Summary 获取用户的媒体库列表
// @Tags 媒体
// @Security Bearer
// @Success 200 {object} response.Response{data=[]service.MediaDBItem}
// @Router /api/v1/media/list [get]
func (h *MediaHandler) GetMediaDBList(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	// 调用媒体服务获取媒体库
	list, err := h.mediaService.GetMediaDBList(userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, list)
}

// GetMediaDBItems 获取媒体库中的媒体列表
// @Summary 获取媒体库中的媒体列表
// @Tags 媒体
// @Security Bearer
// @Param mediadb_guid path string true "媒体库GUID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=service.GetMediaDBItemsResponse}
// @Router /api/v1/media/db/{mediadb_guid}/items [get]
func (h *MediaHandler) GetMediaDBItems(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	mediaDBGUID := c.Param("guid")

	var req service.GetMediaDBItemsRequest
	req.MediaDBGUID = mediaDBGUID
	req.Page = c.GetInt("page")
	req.PageSize = c.GetInt("page_size")

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	resp, err := h.mediaService.GetMediaDBItems(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetMediaDetail 获取媒体详情
// @Summary 获取媒体详情
// @Tags 媒体
// @Security Bearer
// @Param media_guid path string true "媒体GUID"
// @Success 200 {object} response.Response{data=emby.MediaDetail}
// @Router /api/v1/media/{media_guid} [get]
func (h *MediaHandler) GetMediaDetail(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	mediaGUID := c.Param("guid")

	detail, err := h.mediaService.GetMediaDetail(userID, mediaGUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, detail)
}

// SearchMedia 搜索媒体
// @Summary 搜索媒体
// @Tags 媒体
// @Security Bearer
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=service.SearchMediaResponse}
// @Router /api/v1/media/search [get]
func (h *MediaHandler) SearchMedia(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req service.SearchMediaRequest
	req.Keyword = c.Query("keyword")
	if req.Keyword == "" {
		response.BadRequest(c, "请输入搜索关键词")
		return
	}

	req.Page = c.GetInt("page")
	req.PageSize = c.GetInt("page_size")

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	resp, err := h.mediaService.SearchMedia(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetMediaDBSum 获取媒体库统计
// @Summary 获取媒体库统计
// @Tags 媒体
// @Security Bearer
// @Param mediadb_guid path string true "媒体库GUID"
// @Success 200 {object} response.Response{data=emby.MediaDBSum}
// @Router /api/v1/media/db/{mediadb_guid}/sum [get]
func (h *MediaHandler) GetMediaDBSum(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	mediaDBGUID := c.Param("guid")

	sum, err := h.mediaService.GetMediaDBSum(userID, mediaDBGUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, sum)
}

// GetMediaSeasons 获取媒体的季列表
// @Summary 获取电视剧的季列表
// @Tags 媒体
// @Security Bearer
// @Param media_guid path string true "媒体GUID"
// @Success 200 {object} response.Response{data=[]emby.Season}
// @Router /api/v1/media/{media_guid}/seasons [get]
func (h *MediaHandler) GetMediaSeasons(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	mediaGUID := c.Param("guid")

	seasons, err := h.mediaService.GetMediaSeasons(userID, mediaGUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, seasons)
}

// GetSeasonEpisodes 获取季的剧集列表
// @Summary 获取季的剧集列表
// @Tags 媒体
// @Security Bearer
// @Param season_guid path string true "季GUID"
// @Success 200 {object} response.Response{data=[]emby.Episode}
// @Router /api/v1/media/season/{season_guid}/episodes [get]
func (h *MediaHandler) GetSeasonEpisodes(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	seasonGUID := c.Param("guid")

	episodes, err := h.mediaService.GetSeasonEpisodes(userID, seasonGUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, episodes)
}

// GetAllEpisodes 获取剧集的所有集数
// @Summary 获取剧集的所有集数（不分季）
// @Tags 媒体
// @Security Bearer
// @Param series_guid path string true "剧集GUID"
// @Success 200 {object} response.Response{data=[]emby.Episode}
// @Router /api/v1/media/series/{series_guid}/episodes [get]
func (h *MediaHandler) GetAllEpisodes(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	seriesGUID := c.Param("guid")

	episodes, err := h.mediaService.GetAllEpisodes(userID, seriesGUID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, episodes)
}
