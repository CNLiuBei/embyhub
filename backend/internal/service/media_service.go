// Package service 媒体服务
package service

import (
	"context"
	"errors"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/pkg/emby"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MediaService 媒体服务
type MediaService struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewMediaService 创建媒体服务
func NewMediaService(db *gorm.DB, cfg *config.Config) *MediaService {
	return &MediaService{
		db:  db,
		cfg: cfg,
	}
}

// getAdapter 获取媒体适配器（从数据库配置）
func (s *MediaService) getAdapter() emby.MediaAdapter {
	return GetMediaAdapterFromDB(s.db)
}

// isEmbyEnabled 检查媒体服务是否启用
func (s *MediaService) isEmbyEnabled() bool {
	return IsEmbyEnabledFromDB(s.db)
}

// isEmbyMode 检查是否为Emby模式
func (s *MediaService) isEmbyMode() bool {
	return IsEmbyModeFromDB(s.db)
}

// MediaDBItem 媒体库项
type MediaDBItem struct {
	GUID     string   `json:"guid"`
	Title    string   `json:"title"`
	Posters  []string `json:"posters"`
	Category string   `json:"category"`
	ImageURL string   `json:"image_url"` // 完整图片URL前缀
}

// GetMediaDBList 获取媒体库列表
func (s *MediaService) GetMediaDBList(userID uuid.UUID) ([]MediaDBItem, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, err := s.getUserTokenInfo(userID)
	if err != nil {
		return nil, err
	}

	// 获取媒体库列表（使用用户token获取用户可见的媒体库）
	mediaDBs, err := adapter.GetMediaDBList(tokenInfo)
	if err != nil {
		return nil, err
	}

	// 转换为返回格式 - 使用本地代理
	imageBaseURL := "/api/v1/image/"
	items := make([]MediaDBItem, len(mediaDBs))
	for i, db := range mediaDBs {
		items[i] = MediaDBItem{
			GUID:     db.GUID,
			Title:    db.Title,
			Posters:  db.Posters,
			Category: db.Category,
			ImageURL: imageBaseURL,
		}
	}

	return items, nil
}

// getUserTokenInfo 获取用户token信息（包含token和Emby用户ID）
func (s *MediaService) getUserTokenInfo(userID uuid.UUID) (*emby.UserTokenInfo, error) {
	ctx := context.Background()
	embyTokenKey := "emby:token:" + userID.String()
	token, _ := database.GetCache(ctx, embyTokenKey)

	// 如果用户token不存在
	if token == "" {
		// 对于Emby模式，可以使用API Key，不需要用户token
		if s.isEmbyMode() {
			return nil, nil
		}
		// 飞牛模式需要token
		return nil, errors.New("无法获取媒体服务token")
	}

	// 获取Emby用户ID
	embyUserIDKey := "emby:userid:" + userID.String()
	embyUserID, _ := database.GetCache(ctx, embyUserIDKey)

	return &emby.UserTokenInfo{
		Token:      token,
		EmbyUserID: embyUserID,
	}, nil
}

// GetMediaDBItemsRequest 获取媒体列表请求
type GetMediaDBItemsRequest struct {
	MediaDBGUID string `json:"mediadb_guid" binding:"required"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

// GetMediaDBItemsResponse 获取媒体列表响应
type GetMediaDBItemsResponse struct {
	Items    []emby.MediaItem `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	ImageURL string             `json:"image_url"`
}

// GetMediaDBItems 获取媒体库中的媒体列表
func (s *MediaService) GetMediaDBItems(userID uuid.UUID, req *GetMediaDBItemsRequest) (*GetMediaDBItemsResponse, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)

	items, total, err := adapter.GetMediaDBItems(tokenInfo, req.MediaDBGUID, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &GetMediaDBItemsResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		ImageURL: "/api/v1/image/",
	}, nil
}

// GetMediaDetail 获取媒体详情
func (s *MediaService) GetMediaDetail(userID uuid.UUID, mediaGUID string) (*emby.MediaDetail, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)
	return adapter.GetMediaDetail(tokenInfo, mediaGUID)
}

// SearchMediaRequest 搜索媒体请求
type SearchMediaRequest struct {
	Keyword  string `json:"keyword" binding:"required"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// SearchMediaResponse 搜索媒体响应
type SearchMediaResponse struct {
	Items    []emby.MediaItem `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	ImageURL string             `json:"image_url"`
}

// SearchMedia 搜索媒体
func (s *MediaService) SearchMedia(userID uuid.UUID, req *SearchMediaRequest) (*SearchMediaResponse, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)

	items, total, err := adapter.SearchMedia(tokenInfo, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &SearchMediaResponse{
		Items:    items,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		ImageURL: "/api/v1/image/",
	}, nil
}

// GetMediaDBSum 获取媒体库统计
func (s *MediaService) GetMediaDBSum(userID uuid.UUID, mediaDBGUID string) (*emby.MediaDBSum, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)
	return adapter.GetMediaDBSum(tokenInfo, mediaDBGUID)
}

// GetMediaSeasons 获取媒体的季列表
func (s *MediaService) GetMediaSeasons(userID uuid.UUID, mediaGUID string) ([]emby.Season, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)
	return adapter.GetMediaSeasons(tokenInfo, mediaGUID)
}

// GetSeasonEpisodes 获取季的剧集列表
func (s *MediaService) GetSeasonEpisodes(userID uuid.UUID, seasonGUID string) ([]emby.Episode, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)
	return adapter.GetSeasonEpisodes(tokenInfo, seasonGUID)
}

// GetAllEpisodes 获取剧集的所有集数（直接通过剧集ID获取）
func (s *MediaService) GetAllEpisodes(userID uuid.UUID, seriesGUID string) ([]emby.Episode, error) {
	adapter := s.getAdapter()
	if adapter == nil {
		return nil, errors.New("媒体服务未启用")
	}

	tokenInfo, _ := s.getUserTokenInfo(userID)
	// 直接使用 GetSeasonEpisodes，它会自动判断类型并获取所有集数
	return adapter.GetSeasonEpisodes(tokenInfo, seriesGUID)
}
