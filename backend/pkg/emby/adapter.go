// Package emby 媒体服务适配器
// 统一封装Emby官方API和飞牛影视API
package emby

import (
	"fmt"

	"feiniu-user-system/internal/config"
)

// EmbyUserInfo Emby用户详细信息
type EmbyUserInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	IsAdmin          bool   `json:"is_admin"`
	IsDisabled       bool   `json:"is_disabled"`
	HasPassword      bool   `json:"has_password"`
	LastLoginDate    string `json:"last_login_date"`
	LastActivityDate string `json:"last_activity_date"`
	// 权限信息
	EnableMediaPlayback            bool `json:"enable_media_playback"`
	EnableRemoteAccess             bool `json:"enable_remote_access"`
	EnableAllFolders               bool `json:"enable_all_folders"`
	EnableLiveTvAccess             bool `json:"enable_live_tv_access"`
	EnableLiveTvManagement         bool `json:"enable_live_tv_management"`
	EnableContentDeletion          bool `json:"enable_content_deletion"`
	EnableContentDownloading       bool `json:"enable_content_downloading"`
	EnableSubtitleDownloading      bool `json:"enable_subtitle_downloading"`
	EnableSubtitleManagement       bool `json:"enable_subtitle_management"`
	EnablePlaybackRemuxing         bool `json:"enable_playback_remuxing"`
	EnableVideoPlaybackTranscoding bool `json:"enable_video_playback_transcoding"`
	EnableAudioPlaybackTranscoding bool `json:"enable_audio_playback_transcoding"`
	EnablePublicSharing            bool `json:"enable_public_sharing"`
	EnableAllDevices               bool `json:"enable_all_devices"`
	EnableAllChannels              bool `json:"enable_all_channels"`
	SimultaneousStreamLimit        int  `json:"simultaneous_stream_limit"`
	RemoteClientBitrateLimit       int  `json:"remote_client_bitrate_limit"`
	InvalidLoginAttemptCount       int  `json:"invalid_login_attempt_count"`
}

// UserTokenInfo 用户token信息（包含token和Emby用户ID）
type UserTokenInfo struct {
	Token      string // 用户的Emby访问token
	EmbyUserID string // 用户在Emby中的ID
}

// MediaAdapter 媒体服务适配器接口
type MediaAdapter interface {
	// 用户管理
	CreateUser(username, password string) (string, error)
	CreateUserWithTemplate(username, password, templateUsername string) (string, error)
	DeleteUser(username string) error
	UpdateUserPassword(username, password string) error
	UpdateUserStatus(username string, enabled bool) error
	GetUserList() ([]UserInfo, error)
	GetEmbyUserList() ([]EmbyUserInfo, error)
	LoginUser(username, password string) (*UserTokenInfo, error) // 返回token和用户ID

	// 媒体库 - 使用UserTokenInfo来获取用户特定的媒体
	GetMediaDBList(tokenInfo *UserTokenInfo) ([]MediaDB, error)
	GetMediaDBItems(tokenInfo *UserTokenInfo, mediaDBGUID string, page, pageSize int) ([]MediaItem, int, error)
	GetMediaDetail(tokenInfo *UserTokenInfo, mediaGUID string) (*MediaDetail, error)
	GetMediaSeasons(tokenInfo *UserTokenInfo, mediaGUID string) ([]Season, error)
	GetSeasonEpisodes(tokenInfo *UserTokenInfo, seasonGUID string) ([]Episode, error)
	SearchMedia(tokenInfo *UserTokenInfo, keyword string, page, pageSize int) ([]MediaItem, int, error)
	GetMediaDBSum(tokenInfo *UserTokenInfo, mediaDBGUID string) (*MediaDBSum, error)

	// 设备和会话管理
	GetSessions() ([]SessionInfo, error)
	GetSessionsByUserID(userID string) ([]SessionInfo, error)
	GetDevices() ([]DeviceInfo, error)
	GetDevicesByUserID(userID string) ([]DeviceInfo, error)
	DeleteDevice(deviceID string) error
	KillSession(sessionID string) error
	StopPlayback(sessionID string) error
	SetUserStreamLimit(userID string, limit int) error
	GetUserStreamLimit(userID string) (int, error)

	// 用户设备策略管理（使用Emby的EnableAllDevices/EnabledDevices）
	GetUserDevicePolicy(userID string) (enableAll bool, enabledDevices []string, err error)
	SetUserDevicePolicy(userID string, enableAll bool, deviceIDs []string) error
	AddDeviceToUserWhitelist(userID, deviceID string) error
	RemoveDeviceFromUserWhitelist(userID, deviceID string) error

	// 图片
	GetImageURL(path string) string

	// 获取基础URL
	GetBaseURL() string
}

// SessionInfo 会话信息（统一格式）
type SessionInfo struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"user_id"`
	UserName           string          `json:"user_name"`
	Client             string          `json:"client"`
	DeviceID           string          `json:"device_id"`
	DeviceName         string          `json:"device_name"`
	DeviceType         string          `json:"device_type"`
	AppVersion         string          `json:"app_version"`
	LastActivityDate   string          `json:"last_activity_date"`
	RemoteEndPoint     string          `json:"remote_end_point"`
	IsPlaying          bool            `json:"is_playing"`
	NowPlayingItem     *NowPlayingInfo `json:"now_playing_item,omitempty"`
	PlayState          *PlayStateInfo  `json:"play_state,omitempty"`
	TranscodingInfo    *TranscodeInfo  `json:"transcoding_info,omitempty"`
}

// NowPlayingInfo 正在播放信息
type NowPlayingInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	SeriesName     string `json:"series_name,omitempty"`
	SeasonName     string `json:"season_name,omitempty"`
	EpisodeNumber  int    `json:"episode_number,omitempty"`
	SeasonNumber   int    `json:"season_number,omitempty"`
	RunTimeTicks   int64  `json:"run_time_ticks,omitempty"`
	ProductionYear int    `json:"production_year,omitempty"`
}

// PlayStateInfo 播放状态信息
type PlayStateInfo struct {
	PositionTicks int64  `json:"position_ticks"`
	IsPaused      bool   `json:"is_paused"`
	IsMuted       bool   `json:"is_muted"`
	PlayMethod    string `json:"play_method,omitempty"`
}

// TranscodeInfo 转码信息
type TranscodeInfo struct {
	VideoCodec    string  `json:"video_codec,omitempty"`
	AudioCodec    string  `json:"audio_codec,omitempty"`
	IsVideoDirect bool    `json:"is_video_direct"`
	IsAudioDirect bool    `json:"is_audio_direct"`
	Bitrate       int     `json:"bitrate,omitempty"`
	Width         int     `json:"width,omitempty"`
	Height        int     `json:"height,omitempty"`
	Completion    float64 `json:"completion,omitempty"`
}

// DeviceInfo 设备信息（统一格式）
type DeviceInfo struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	AppName          string `json:"app_name"`
	AppVersion       string `json:"app_version"`
	LastUserID       string `json:"last_user_id,omitempty"`
	LastUserName     string `json:"last_user_name,omitempty"`
	LastActivityDate string `json:"last_activity_date"`
}

// EmbyAdapterConfig Emby适配器配置（用于从设置创建）
type EmbyAdapterConfig struct {
	BaseURL string
	APIKey  string
}

// FeiniuAdapterConfig 飞牛适配器配置（用于从设置创建）
type FeiniuAdapterConfig struct {
	BaseURL   string
	AdminUser string
	AdminPass string
}

// NewMediaAdapter 创建媒体服务适配器（从配置文件）
func NewMediaAdapter(cfg *config.EmbyConfig) MediaAdapter {
	if cfg.IsEmbyMode() {
		return NewEmbyAdapterFromConfig(cfg)
	}
	return NewFeiniuAdapterFromConfig(cfg)
}

// ========== Emby官方适配器 ==========

// EmbyAdapter Emby官方API适配器
type EmbyAdapter struct {
	client  *EmbyClient
	baseURL string
}

// NewEmbyAdapterFromConfig 从配置文件创建Emby适配器
func NewEmbyAdapterFromConfig(cfg *config.EmbyConfig) *EmbyAdapter {
	return &EmbyAdapter{
		client: NewEmbyClient(&EmbyConfig{
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		}),
		baseURL: cfg.BaseURL,
	}
}

// NewEmbyAdapter 从设置创建Emby适配器
func NewEmbyAdapter(cfg *EmbyAdapterConfig) *EmbyAdapter {
	return &EmbyAdapter{
		client: NewEmbyClient(&EmbyConfig{
			BaseURL: cfg.BaseURL,
			APIKey:  cfg.APIKey,
		}),
		baseURL: cfg.BaseURL,
	}
}

// CreateUser 创建用户
func (a *EmbyAdapter) CreateUser(username, password string) (string, error) {
	user, err := a.client.CreateUserWithPassword(username, password)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

// CreateUserWithTemplate 使用模板用户权限创建用户
func (a *EmbyAdapter) CreateUserWithTemplate(username, password, templateUsername string) (string, error) {
	user, err := a.client.CreateUserWithTemplate(username, password, templateUsername)
	if err != nil {
		return "", err
	}
	return user.ID, nil
}

// DeleteUser 删除用户
func (a *EmbyAdapter) DeleteUser(username string) error {
	return a.client.DeleteUserByName(username)
}

// UpdateUserPassword 更新用户密码
func (a *EmbyAdapter) UpdateUserPassword(username, password string) error {
	return a.client.UpdateUserPasswordByName(username, password)
}

// UpdateUserStatus 更新用户状态
func (a *EmbyAdapter) UpdateUserStatus(username string, enabled bool) error {
	if enabled {
		return a.client.EnableUserByName(username)
	}
	return a.client.DisableUserByName(username)
}

// GetUserList 获取用户列表
func (a *EmbyAdapter) GetUserList() ([]UserInfo, error) {
	users, err := a.client.GetUsers()
	if err != nil {
		return nil, err
	}

	result := make([]UserInfo, len(users))
	for i, u := range users {
		isAdmin := 0
		if u.Policy != nil && u.Policy.IsAdministrator {
			isAdmin = 1
		}
		result[i] = UserInfo{
			GUID:     u.ID,
			Username: u.Name,
			IsAdmin:  isAdmin,
		}
	}
	return result, nil
}

// GetEmbyUserList 获取详细的Emby用户列表
func (a *EmbyAdapter) GetEmbyUserList() ([]EmbyUserInfo, error) {
	users, err := a.client.GetUsers()
	if err != nil {
		return nil, err
	}

	result := make([]EmbyUserInfo, len(users))
	for i, u := range users {
		info := EmbyUserInfo{
			ID:               u.ID,
			Name:             u.Name,
			HasPassword:      u.HasPassword,
			LastLoginDate:    u.LastLoginDate,
			LastActivityDate: u.LastActivityDate,
		}
		if u.Policy != nil {
			info.IsAdmin = u.Policy.IsAdministrator
			info.IsDisabled = u.Policy.IsDisabled
			info.EnableMediaPlayback = u.Policy.EnableMediaPlayback
			info.EnableRemoteAccess = u.Policy.EnableRemoteAccess
			info.EnableAllFolders = u.Policy.EnableAllFolders
			info.EnableLiveTvAccess = u.Policy.EnableLiveTvAccess
			info.EnableLiveTvManagement = u.Policy.EnableLiveTvManagement
			info.EnableContentDeletion = u.Policy.EnableContentDeletion
			info.EnableContentDownloading = u.Policy.EnableContentDownloading
			info.EnableSubtitleDownloading = u.Policy.EnableSubtitleDownloading
			info.EnableSubtitleManagement = u.Policy.EnableSubtitleManagement
			info.EnablePlaybackRemuxing = u.Policy.EnablePlaybackRemuxing
			info.EnableVideoPlaybackTranscoding = u.Policy.EnableVideoPlaybackTranscoding
			info.EnableAudioPlaybackTranscoding = u.Policy.EnableAudioPlaybackTranscoding
			info.EnablePublicSharing = u.Policy.EnablePublicSharing
			info.EnableAllDevices = u.Policy.EnableAllDevices
			info.EnableAllChannels = u.Policy.EnableAllChannels
			info.SimultaneousStreamLimit = u.Policy.SimultaneousStreamLimit
			info.RemoteClientBitrateLimit = u.Policy.RemoteClientBitrateLimit
			info.InvalidLoginAttemptCount = u.Policy.InvalidLoginAttemptCount
		}
		result[i] = info
	}
	return result, nil
}

// LoginUser 用户登录，返回token和Emby用户ID
func (a *EmbyAdapter) LoginUser(username, password string) (*UserTokenInfo, error) {
	result, err := a.client.AuthenticateUser(username, password)
	if err != nil {
		return nil, err
	}
	return &UserTokenInfo{
		Token:      result.AccessToken,
		EmbyUserID: result.User.ID,
	}, nil
}

// GetMediaDBList 获取媒体库列表（使用用户token获取用户可见的媒体库）
func (a *EmbyAdapter) GetMediaDBList(tokenInfo *UserTokenInfo) ([]MediaDB, error) {
	var folders []MediaFolder
	var err error

	// 如果有用户token，使用用户token获取用户可见的媒体库
	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		folders, err = a.client.GetUserViewsWithToken(tokenInfo.EmbyUserID, tokenInfo.Token)
	} else {
		// 回退到API Key获取所有媒体库
		folders, err = a.client.GetMediaFolders()
	}

	if err != nil {
		return nil, err
	}

	result := make([]MediaDB, len(folders))
	for i, f := range folders {
		// 直接生成海报URL
		posters := []string{a.client.GetPrimaryImageURL(f.ID, 400)}

		result[i] = MediaDB{
			GUID:     f.ID,
			Title:    f.Name,
			Posters:  posters,
			Category: f.CollectionType,
			ViewType: 0,
		}
	}
	return result, nil
}

// GetMediaDBItems 获取媒体库中的媒体列表（使用用户token）
func (a *EmbyAdapter) GetMediaDBItems(tokenInfo *UserTokenInfo, mediaDBGUID string, page, pageSize int) ([]MediaItem, int, error) {
	startIndex := (page - 1) * pageSize

	params := &GetItemsParams{
		ParentID:         mediaDBGUID,
		IncludeItemTypes: "Movie,Series",
		SortBy:           "DateCreated,SortName",
		SortOrder:        "Descending",
		StartIndex:       startIndex,
		Limit:            pageSize,
		Recursive:        true,
		Fields:           "Overview,Genres,CommunityRating,ProductionYear,PrimaryImageAspectRatio,ChildCount,RecursiveItemCount,SeriesId",
	}

	var resp *EmbyItemsResponse
	var err error

	// 使用用户token获取媒体
	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		params.UserID = tokenInfo.EmbyUserID
		resp, err = a.client.GetItemsWithToken(params, tokenInfo.Token)
	} else {
		resp, err = a.client.GetItems(params)
	}

	if err != nil {
		return nil, 0, err
	}

	items := make([]MediaItem, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = a.convertEmbyToMediaItem(&item)
	}

	return items, resp.TotalRecordCount, nil
}

// GetMediaDetail 获取媒体详情（使用用户token）
func (a *EmbyAdapter) GetMediaDetail(tokenInfo *UserTokenInfo, mediaGUID string) (*MediaDetail, error) {
	var item *EmbyMediaItem
	var err error

	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		item, err = a.client.GetItemByIDWithToken(tokenInfo.EmbyUserID, mediaGUID, tokenInfo.Token)
	} else {
		// 回退：获取第一个用户ID用于查询
		users, err := a.client.GetUsers()
		if err != nil || len(users) == 0 {
			return nil, err
		}
		item, err = a.client.GetItemByID(users[0].ID, mediaGUID)
	}

	if err != nil {
		return nil, err
	}

	return a.convertEmbyToMediaDetail(item), nil
}

// GetMediaSeasons 获取媒体的季列表（使用用户token）
func (a *EmbyAdapter) GetMediaSeasons(tokenInfo *UserTokenInfo, mediaGUID string) ([]Season, error) {
	var items []EmbyMediaItem
	var err error

	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		items, err = a.client.GetSeasonsWithToken(tokenInfo.EmbyUserID, mediaGUID, tokenInfo.Token)
	} else {
		users, err := a.client.GetUsers()
		if err != nil || len(users) == 0 {
			return nil, err
		}
		items, err = a.client.GetSeasons(users[0].ID, mediaGUID)
	}

	if err != nil {
		return nil, err
	}

	seasons := make([]Season, len(items))
	for i, item := range items {
		seasons[i] = Season{
			GUID:              item.ID,
			Title:             item.Name,
			SeasonNumber:      item.IndexNumber,
			EpisodeCount:      item.ChildCount,
			LocalEpisodeCount: item.ChildCount,
			Poster:            a.client.GetPrimaryImageURL(item.ID, 300),
		}
	}
	return seasons, nil
}

// GetSeasonEpisodes 获取季的剧集列表（使用用户token）
// 注意：seasonGUID 可以是季ID或剧集ID，会自动获取该季或整个剧集的所有集数
func (a *EmbyAdapter) GetSeasonEpisodes(tokenInfo *UserTokenInfo, seasonGUID string) ([]Episode, error) {
	var userID string
	var userToken string

	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		userID = tokenInfo.EmbyUserID
		userToken = tokenInfo.Token
	} else {
		users, err := a.client.GetUsers()
		if err != nil || len(users) == 0 {
			return nil, err
		}
		userID = users[0].ID
	}

	// 先获取项目信息以确定类型和SeriesId
	var item *EmbyMediaItem
	var err error
	if userToken != "" {
		item, err = a.client.GetItemByIDWithToken(userID, seasonGUID, userToken)
	} else {
		item, err = a.client.GetItemByID(userID, seasonGUID)
	}
	if err != nil {
		return nil, err
	}

	var items []EmbyMediaItem
	
	// 根据类型决定如何获取剧集
	if item.Type == "Season" {
		// 如果是季，获取该季的剧集
		if userToken != "" {
			items, err = a.client.GetEpisodesWithToken(userID, item.SeriesID, seasonGUID, userToken)
		} else {
			items, err = a.client.GetEpisodes(userID, item.SeriesID, seasonGUID)
		}
	} else if item.Type == "Series" {
		// 如果是剧集，获取所有集数
		if userToken != "" {
			items, err = a.client.GetAllEpisodesWithToken(userID, seasonGUID, userToken)
		} else {
			items, err = a.client.GetAllEpisodes(userID, seasonGUID)
		}
	} else {
		// 其他类型，尝试使用SeriesID获取
		seriesID := item.SeriesID
		if seriesID == "" {
			seriesID = seasonGUID
		}
		if userToken != "" {
			items, err = a.client.GetAllEpisodesWithToken(userID, seriesID, userToken)
		} else {
			items, err = a.client.GetAllEpisodes(userID, seriesID)
		}
	}
	
	if err != nil {
		return nil, err
	}

	episodes := make([]Episode, len(items))
	for i, item := range items {
		isWatched := 0
		playPosition := 0
		if item.UserData != nil {
			if item.UserData.Played {
				isWatched = 1
			}
			playPosition = int(item.UserData.PlaybackPositionTicks / 10000000) // 转换为秒
		}

		episodes[i] = Episode{
			GUID:          item.ID,
			Title:         item.Name,
			EpisodeNumber: item.IndexNumber,
			SeasonNumber:  item.ParentIndexNumber,
			Overview:      item.Overview,
			StillPath:     a.client.GetPrimaryImageURL(item.ID, 400),
			Runtime:       int(item.RunTimeTicks / 600000000), // 转换为分钟
			IsWatched:     isWatched,
			PlayPosition:  playPosition,
		}
	}
	return episodes, nil
}

// SearchMedia 搜索媒体（使用用户token）
func (a *EmbyAdapter) SearchMedia(tokenInfo *UserTokenInfo, keyword string, page, pageSize int) ([]MediaItem, int, error) {
	startIndex := (page - 1) * pageSize

	params := &GetItemsParams{
		SearchTerm:       keyword,
		IncludeItemTypes: "Movie,Series",
		Recursive:        true,
		StartIndex:       startIndex,
		Limit:            pageSize,
		Fields:           "Overview,Genres,CommunityRating,ProductionYear,ChildCount,RecursiveItemCount,SeriesId",
	}

	var resp *EmbyItemsResponse
	var err error

	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		params.UserID = tokenInfo.EmbyUserID
		resp, err = a.client.GetItemsWithToken(params, tokenInfo.Token)
	} else {
		users, usersErr := a.client.GetUsers()
		if usersErr != nil || len(users) == 0 {
			return nil, 0, usersErr
		}
		params.UserID = users[0].ID
		resp, err = a.client.GetItems(params)
	}

	if err != nil {
		return nil, 0, err
	}

	items := make([]MediaItem, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = a.convertEmbyToMediaItem(&item)
	}

	// 如果 TotalRecordCount 为 0 但有结果，使用结果数量
	// 这是因为某些 Emby 版本在搜索时不返回正确的 TotalRecordCount
	total := resp.TotalRecordCount
	if total == 0 && len(items) > 0 {
		total = len(items)
	}

	return items, total, nil
}

// GetMediaDBSum 获取媒体库统计（使用用户token）
func (a *EmbyAdapter) GetMediaDBSum(tokenInfo *UserTokenInfo, mediaDBGUID string) (*MediaDBSum, error) {
	movieParams := &GetItemsParams{
		ParentID:         mediaDBGUID,
		IncludeItemTypes: "Movie",
		Recursive:        true,
		Limit:            0,
	}
	tvParams := &GetItemsParams{
		ParentID:         mediaDBGUID,
		IncludeItemTypes: "Series",
		Recursive:        true,
		Limit:            0,
	}

	var movieResp, tvResp *EmbyItemsResponse

	if tokenInfo != nil && tokenInfo.Token != "" && tokenInfo.EmbyUserID != "" {
		movieParams.UserID = tokenInfo.EmbyUserID
		tvParams.UserID = tokenInfo.EmbyUserID
		movieResp, _ = a.client.GetItemsWithToken(movieParams, tokenInfo.Token)
		tvResp, _ = a.client.GetItemsWithToken(tvParams, tokenInfo.Token)
	} else {
		movieResp, _ = a.client.GetItems(movieParams)
		tvResp, _ = a.client.GetItems(tvParams)
	}

	movieCount := 0
	tvCount := 0
	if movieResp != nil {
		movieCount = movieResp.TotalRecordCount
	}
	if tvResp != nil {
		tvCount = tvResp.TotalRecordCount
	}

	return &MediaDBSum{
		Total:    movieCount + tvCount,
		Movie:    movieCount,
		TV:       tvCount,
		Video:    0,
		Favorite: 0,
	}, nil
}

// GetImageURL 获取图片URL
func (a *EmbyAdapter) GetImageURL(path string) string {
	return a.baseURL + path
}

// GetBaseURL 获取基础URL
func (a *EmbyAdapter) GetBaseURL() string {
	return a.baseURL
}

// GetSessions 获取所有活动会话
func (a *EmbyAdapter) GetSessions() ([]SessionInfo, error) {
	sessions, err := a.client.GetSessions()
	if err != nil {
		return nil, err
	}

	result := make([]SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = a.convertEmbySession(&s)
	}
	return result, nil
}

// GetSessionsByUserID 获取指定用户的活动会话
func (a *EmbyAdapter) GetSessionsByUserID(userID string) ([]SessionInfo, error) {
	sessions, err := a.client.GetSessionsByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]SessionInfo, len(sessions))
	for i, s := range sessions {
		result[i] = a.convertEmbySession(&s)
	}
	return result, nil
}

// GetDevices 获取所有注册设备
func (a *EmbyAdapter) GetDevices() ([]DeviceInfo, error) {
	devices, err := a.client.GetDevices()
	if err != nil {
		return nil, err
	}

	result := make([]DeviceInfo, len(devices))
	for i, d := range devices {
		// 使用 ReportedDeviceID 作为设备 ID（这是 EnabledDevices 策略需要的）
		// 如果没有 ReportedDeviceID，则使用内部 ID
		deviceID := d.ReportedDeviceID
		if deviceID == "" {
			deviceID = d.ID
		}
		result[i] = DeviceInfo{
			ID:               deviceID,
			Name:             d.Name,
			AppName:          d.AppName,
			AppVersion:       d.AppVersion,
			LastUserID:       d.LastUserID,
			LastUserName:     d.LastUserName,
			LastActivityDate: d.DateLastActivity,
		}
	}
	return result, nil
}

// GetDevicesByUserID 获取指定用户的设备
func (a *EmbyAdapter) GetDevicesByUserID(userID string) ([]DeviceInfo, error) {
	devices, err := a.client.GetDevicesByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := make([]DeviceInfo, len(devices))
	for i, d := range devices {
		// 使用 ReportedDeviceID 作为设备 ID（这是 EnabledDevices 策略需要的）
		deviceID := d.ReportedDeviceID
		if deviceID == "" {
			deviceID = d.ID
		}
		result[i] = DeviceInfo{
			ID:               deviceID,
			Name:             d.Name,
			AppName:          d.AppName,
			AppVersion:       d.AppVersion,
			LastUserID:       d.LastUserID,
			LastUserName:     d.LastUserName,
			LastActivityDate: d.DateLastActivity,
		}
	}
	return result, nil
}

// DeleteDevice 删除设备
func (a *EmbyAdapter) DeleteDevice(deviceID string) error {
	return a.client.DeleteDevice(deviceID)
}

// KillSession 终止会话
func (a *EmbyAdapter) KillSession(sessionID string) error {
	return a.client.KillSession(sessionID)
}

// StopPlayback 停止播放
func (a *EmbyAdapter) StopPlayback(sessionID string) error {
	return a.client.StopPlayback(sessionID)
}

// SetUserStreamLimit 设置用户同时在线流数限制
func (a *EmbyAdapter) SetUserStreamLimit(userID string, limit int) error {
	return a.client.SetUserSimultaneousStreamLimit(userID, limit)
}

// GetUserStreamLimit 获取用户同时在线流数限制
func (a *EmbyAdapter) GetUserStreamLimit(userID string) (int, error) {
	return a.client.GetUserSimultaneousStreamLimit(userID)
}

// GetUserDevicePolicy 获取用户设备策略
func (a *EmbyAdapter) GetUserDevicePolicy(userID string) (bool, []string, error) {
	enabledDevices, enableAll, err := a.client.GetUserEnabledDevices(userID)
	return enableAll, enabledDevices, err
}

// SetUserDevicePolicy 设置用户设备策略
func (a *EmbyAdapter) SetUserDevicePolicy(userID string, enableAll bool, deviceIDs []string) error {
	return a.client.SetUserDevicePolicy(userID, enableAll, deviceIDs)
}

// AddDeviceToUserWhitelist 添加设备到用户白名单
func (a *EmbyAdapter) AddDeviceToUserWhitelist(userID, deviceID string) error {
	return a.client.AddDeviceToUserWhitelist(userID, deviceID)
}

// RemoveDeviceFromUserWhitelist 从用户白名单移除设备
func (a *EmbyAdapter) RemoveDeviceFromUserWhitelist(userID, deviceID string) error {
	return a.client.RemoveDeviceFromUserWhitelist(userID, deviceID)
}

// convertEmbySession 转换Emby会话为统一格式
func (a *EmbyAdapter) convertEmbySession(s *EmbySession) SessionInfo {
	info := SessionInfo{
		ID:               s.ID,
		UserID:           s.UserID,
		UserName:         s.UserName,
		Client:           s.Client,
		DeviceID:         s.DeviceID,
		DeviceName:       s.DeviceName,
		DeviceType:       s.DeviceType,
		AppVersion:       s.ApplicationVersion,
		LastActivityDate: s.LastActivityDate,
		RemoteEndPoint:   s.RemoteEndPoint,
		IsPlaying:        s.NowPlayingItem != nil,
	}

	if s.NowPlayingItem != nil {
		info.NowPlayingItem = &NowPlayingInfo{
			ID:             s.NowPlayingItem.ID,
			Name:           s.NowPlayingItem.Name,
			Type:           s.NowPlayingItem.Type,
			SeriesName:     s.NowPlayingItem.SeriesName,
			SeasonName:     s.NowPlayingItem.SeasonName,
			EpisodeNumber:  s.NowPlayingItem.IndexNumber,
			SeasonNumber:   s.NowPlayingItem.ParentIndexNumber,
			RunTimeTicks:   s.NowPlayingItem.RunTimeTicks,
			ProductionYear: s.NowPlayingItem.ProductionYear,
		}
	}

	if s.PlayState != nil {
		info.PlayState = &PlayStateInfo{
			PositionTicks: s.PlayState.PositionTicks,
			IsPaused:      s.PlayState.IsPaused,
			IsMuted:       s.PlayState.IsMuted,
			PlayMethod:    s.PlayState.PlayMethod,
		}
	}

	if s.TranscodingInfo != nil {
		info.TranscodingInfo = &TranscodeInfo{
			VideoCodec:    s.TranscodingInfo.VideoCodec,
			AudioCodec:    s.TranscodingInfo.AudioCodec,
			IsVideoDirect: s.TranscodingInfo.IsVideoDirect,
			IsAudioDirect: s.TranscodingInfo.IsAudioDirect,
			Bitrate:       s.TranscodingInfo.Bitrate,
			Width:         s.TranscodingInfo.Width,
			Height:        s.TranscodingInfo.Height,
			Completion:    s.TranscodingInfo.CompletionPercentage,
		}
	}

	return info
}

// convertEmbyToMediaItem 转换Emby格式为统一的MediaItem格式
func (a *EmbyAdapter) convertEmbyToMediaItem(item *EmbyMediaItem) MediaItem {
	// 生成图片URL
	// 对于Episode类型，使用其所属Series的封面
	// 对于Season类型，也使用其所属Series的封面
	var poster string
	if item.Type == "Episode" && item.SeriesID != "" {
		poster = a.client.GetPrimaryImageURL(item.SeriesID, 300)
	} else if item.Type == "Season" && item.SeriesID != "" {
		poster = a.client.GetPrimaryImageURL(item.SeriesID, 300)
	} else {
		poster = a.client.GetPrimaryImageURL(item.ID, 300)
	}

	isFavorite := 0
	watched := 0
	if item.UserData != nil {
		if item.UserData.IsFavorite {
			isFavorite = 1
		}
		if item.UserData.Played {
			watched = 1
		}
	}

	// 剧集数量：优先使用ChildCount，其次RecursiveItemCount
	episodeCount := item.ChildCount
	if episodeCount == 0 {
		episodeCount = item.RecursiveItemCount
	}

	return MediaItem{
		GUID:                  item.ID,
		Title:                 item.Name,
		OriginalTitle:         item.OriginalTitle,
		Type:                  item.Type,
		Poster:                poster,
		VoteAverage:           formatRating(item.CommunityRating),
		ReleaseDate:           item.PremiereDate,
		Year:                  item.ProductionYear,
		Overview:              item.Overview,
		NumberOfEpisodes:      episodeCount,
		LocalNumberOfEpisodes: episodeCount,
		IsFavorite:            isFavorite,
		Watched:               watched,
		Duration:              int(item.RunTimeTicks / 600000000),
	}
}

// convertEmbyToMediaDetail 转换Emby格式为统一的MediaDetail格式
func (a *EmbyAdapter) convertEmbyToMediaDetail(item *EmbyMediaItem) *MediaDetail {
	// 生成图片URL
	// 对于Episode/Season类型，使用其所属Series的封面
	var posterID, backdropID string
	if (item.Type == "Episode" || item.Type == "Season") && item.SeriesID != "" {
		posterID = item.SeriesID
		backdropID = item.SeriesID
	} else {
		posterID = item.ID
		backdropID = item.ID
	}
	posters := []string{a.client.GetPrimaryImageURL(posterID, 500)}
	backdrops := []string{a.client.GetBackdropImageURL(backdropID, 1920)}

	return &MediaDetail{
		GUID:          item.ID,
		Title:         item.Name,
		OriginalTitle: item.OriginalTitle,
		Year:          item.ProductionYear,
		Rating:        item.CommunityRating,
		Posters:       posters,
		Backdrops:     backdrops,
		Genres:        item.Genres,
		Overview:      item.Overview,
		Duration:      int(item.RunTimeTicks / 600000000),
		Category:      item.Type,
	}
}

func formatRating(rating float64) string {
	if rating == 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", rating)
}

// ========== 飞牛影视适配器 ==========

// FeiniuAdapter 飞牛影视API适配器
type FeiniuAdapter struct {
	client  *Client
	baseURL string
}

// NewFeiniuAdapterFromConfig 从配置文件创建飞牛适配器
func NewFeiniuAdapterFromConfig(cfg *config.EmbyConfig) *FeiniuAdapter {
	return &FeiniuAdapter{
		client: NewClient(&Config{
			BaseURL:   cfg.BaseURL,
			AdminUser: cfg.AdminUser,
			AdminPass: cfg.AdminPass,
		}),
		baseURL: cfg.BaseURL,
	}
}

// NewFeiniuAdapter 从设置创建飞牛适配器
func NewFeiniuAdapter(cfg *FeiniuAdapterConfig) *FeiniuAdapter {
	return &FeiniuAdapter{
		client: NewClient(&Config{
			BaseURL:   cfg.BaseURL,
			AdminUser: cfg.AdminUser,
			AdminPass: cfg.AdminPass,
		}),
		baseURL: cfg.BaseURL,
	}
}

// CreateUser 创建用户
func (a *FeiniuAdapter) CreateUser(username, password string) (string, error) {
	resp, err := a.client.CreateUser(username, password)
	if err != nil {
		return "", err
	}
	return resp.GUID, nil
}

// CreateUserWithTemplate 使用模板用户权限创建用户（飞牛模式不支持，直接创建）
func (a *FeiniuAdapter) CreateUserWithTemplate(username, password, templateUsername string) (string, error) {
	return a.CreateUser(username, password)
}

// DeleteUser 删除用户
func (a *FeiniuAdapter) DeleteUser(username string) error {
	return a.client.DeleteUserByUsername(username)
}

// UpdateUserPassword 更新用户密码
func (a *FeiniuAdapter) UpdateUserPassword(username, password string) error {
	return a.client.UpdateUserPassword(username, password)
}

// UpdateUserStatus 更新用户状态
func (a *FeiniuAdapter) UpdateUserStatus(username string, enabled bool) error {
	return a.client.UpdateUserStatus(username, enabled)
}

// GetUserList 获取用户列表
func (a *FeiniuAdapter) GetUserList() ([]UserInfo, error) {
	return a.client.GetUserList()
}

// GetEmbyUserList 获取详细的Emby用户列表（飞牛模式返回简单列表）
func (a *FeiniuAdapter) GetEmbyUserList() ([]EmbyUserInfo, error) {
	users, err := a.client.GetUserList()
	if err != nil {
		return nil, err
	}

	result := make([]EmbyUserInfo, len(users))
	for i, u := range users {
		result[i] = EmbyUserInfo{
			ID:      u.GUID,
			Name:    u.Username,
			IsAdmin: u.IsAdmin == 1,
		}
	}
	return result, nil
}

// LoginUser 用户登录
func (a *FeiniuAdapter) LoginUser(username, password string) (*UserTokenInfo, error) {
	token, err := a.client.LoginUser(username, password)
	if err != nil {
		return nil, err
	}
	return &UserTokenInfo{
		Token:      token,
		EmbyUserID: "", // 飞牛模式不需要用户ID
	}, nil
}

// GetMediaDBList 获取媒体库列表
func (a *FeiniuAdapter) GetMediaDBList(tokenInfo *UserTokenInfo) ([]MediaDB, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetMediaDBList(token)
}

// GetMediaDBItems 获取媒体库中的媒体列表
func (a *FeiniuAdapter) GetMediaDBItems(tokenInfo *UserTokenInfo, mediaDBGUID string, page, pageSize int) ([]MediaItem, int, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetMediaDBItems(token, mediaDBGUID, page, pageSize)
}

// GetMediaDetail 获取媒体详情
func (a *FeiniuAdapter) GetMediaDetail(tokenInfo *UserTokenInfo, mediaGUID string) (*MediaDetail, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetMediaDetail(token, mediaGUID)
}

// GetMediaSeasons 获取媒体的季列表
func (a *FeiniuAdapter) GetMediaSeasons(tokenInfo *UserTokenInfo, mediaGUID string) ([]Season, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetMediaSeasons(token, mediaGUID)
}

// GetSeasonEpisodes 获取季的剧集列表
func (a *FeiniuAdapter) GetSeasonEpisodes(tokenInfo *UserTokenInfo, seasonGUID string) ([]Episode, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetSeasonEpisodes(token, seasonGUID)
}

// SearchMedia 搜索媒体
func (a *FeiniuAdapter) SearchMedia(tokenInfo *UserTokenInfo, keyword string, page, pageSize int) ([]MediaItem, int, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.SearchMedia(token, keyword, page, pageSize)
}

// GetMediaDBSum 获取媒体库统计
func (a *FeiniuAdapter) GetMediaDBSum(tokenInfo *UserTokenInfo, mediaDBGUID string) (*MediaDBSum, error) {
	token := ""
	if tokenInfo != nil {
		token = tokenInfo.Token
	}
	return a.client.GetMediaDBSum(token, mediaDBGUID)
}

// GetImageURL 获取图片URL
func (a *FeiniuAdapter) GetImageURL(path string) string {
	return a.baseURL + "/sys/img" + path
}

// GetBaseURL 获取基础URL
func (a *FeiniuAdapter) GetBaseURL() string {
	return a.baseURL
}

// GetSessions 获取所有活动会话（飞牛模式不支持）
func (a *FeiniuAdapter) GetSessions() ([]SessionInfo, error) {
	return nil, fmt.Errorf("飞牛模式不支持会话管理")
}

// GetSessionsByUserID 获取指定用户的活动会话（飞牛模式不支持）
func (a *FeiniuAdapter) GetSessionsByUserID(userID string) ([]SessionInfo, error) {
	return nil, fmt.Errorf("飞牛模式不支持会话管理")
}

// GetDevices 获取所有注册设备（飞牛模式不支持）
func (a *FeiniuAdapter) GetDevices() ([]DeviceInfo, error) {
	return nil, fmt.Errorf("飞牛模式不支持设备管理")
}

// GetDevicesByUserID 获取指定用户的设备（飞牛模式不支持）
func (a *FeiniuAdapter) GetDevicesByUserID(userID string) ([]DeviceInfo, error) {
	return nil, fmt.Errorf("飞牛模式不支持设备管理")
}

// DeleteDevice 删除设备（飞牛模式不支持）
func (a *FeiniuAdapter) DeleteDevice(deviceID string) error {
	return fmt.Errorf("飞牛模式不支持设备管理")
}

// KillSession 终止会话（飞牛模式不支持）
func (a *FeiniuAdapter) KillSession(sessionID string) error {
	return fmt.Errorf("飞牛模式不支持会话管理")
}

// StopPlayback 停止播放（飞牛模式不支持）
func (a *FeiniuAdapter) StopPlayback(sessionID string) error {
	return fmt.Errorf("飞牛模式不支持播放控制")
}

// SetUserStreamLimit 设置用户同时在线流数限制（飞牛模式不支持）
func (a *FeiniuAdapter) SetUserStreamLimit(userID string, limit int) error {
	return fmt.Errorf("飞牛模式不支持流数限制")
}

// GetUserStreamLimit 获取用户同时在线流数限制（飞牛模式不支持）
func (a *FeiniuAdapter) GetUserStreamLimit(userID string) (int, error) {
	return 0, fmt.Errorf("飞牛模式不支持流数限制")
}

// GetUserDevicePolicy 获取用户设备策略（飞牛模式不支持）
func (a *FeiniuAdapter) GetUserDevicePolicy(userID string) (bool, []string, error) {
	return true, nil, fmt.Errorf("飞牛模式不支持设备策略管理")
}

// SetUserDevicePolicy 设置用户设备策略（飞牛模式不支持）
func (a *FeiniuAdapter) SetUserDevicePolicy(userID string, enableAll bool, deviceIDs []string) error {
	return fmt.Errorf("飞牛模式不支持设备策略管理")
}

// AddDeviceToUserWhitelist 添加设备到用户白名单（飞牛模式不支持）
func (a *FeiniuAdapter) AddDeviceToUserWhitelist(userID, deviceID string) error {
	return fmt.Errorf("飞牛模式不支持设备策略管理")
}

// RemoveDeviceFromUserWhitelist 从用户白名单移除设备（飞牛模式不支持）
func (a *FeiniuAdapter) RemoveDeviceFromUserWhitelist(userID, deviceID string) error {
	return fmt.Errorf("飞牛模式不支持设备策略管理")
}
