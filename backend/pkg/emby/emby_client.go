// Package emby Emby官方API客户端
package emby

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// EmbyClient Emby官方API客户端
type EmbyClient struct {
	baseURL    string
	apiKey     string
	adminToken string
	client     *http.Client
}

// EmbyConfig Emby配置
type EmbyConfig struct {
	BaseURL  string // Emby服务器地址，如 http://localhost:8096
	APIKey   string // Emby API密钥
	Username string // 管理员用户名（可选，用于获取管理员token）
	Password string // 管理员密码（可选）
}

// NewEmbyClient 创建Emby客户端
func NewEmbyClient(cfg *EmbyConfig) *EmbyClient {
	return &EmbyClient{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ========== 用户管理 API ==========

// EmbyUser Emby用户信息
type EmbyUser struct {
	ID                      string          `json:"Id"`
	Name                    string          `json:"Name"`
	ServerID                string          `json:"ServerId"`
	HasPassword             bool            `json:"HasPassword"`
	HasConfiguredPassword   bool            `json:"HasConfiguredPassword"`
	HasConfiguredEasyPassword bool          `json:"HasConfiguredEasyPassword"`
	EnableAutoLogin         bool            `json:"EnableAutoLogin"`
	LastLoginDate           string          `json:"LastLoginDate,omitempty"`
	LastActivityDate        string          `json:"LastActivityDate,omitempty"`
	Configuration           *UserConfig     `json:"Configuration,omitempty"`
	Policy                  *UserPolicy     `json:"Policy,omitempty"`
}

// UserConfig 用户配置
type UserConfig struct {
	PlayDefaultAudioTrack      bool     `json:"PlayDefaultAudioTrack"`
	SubtitleLanguagePreference string   `json:"SubtitleLanguagePreference"`
	DisplayMissingEpisodes     bool     `json:"DisplayMissingEpisodes"`
	SubtitleMode               string   `json:"SubtitleMode"`
	EnableLocalPassword        bool     `json:"EnableLocalPassword"`
	OrderedViews               []string `json:"OrderedViews"`
	LatestItemsExcludes        []string `json:"LatestItemsExcludes"`
	MyMediaExcludes            []string `json:"MyMediaExcludes"`
	HidePlayedInLatest         bool     `json:"HidePlayedInLatest"`
	RememberAudioSelections    bool     `json:"RememberAudioSelections"`
	RememberSubtitleSelections bool     `json:"RememberSubtitleSelections"`
	EnableNextEpisodeAutoPlay  bool     `json:"EnableNextEpisodeAutoPlay"`
}

// UserPolicy 用户策略/权限
type UserPolicy struct {
	IsAdministrator                  bool     `json:"IsAdministrator"`
	IsHidden                         bool     `json:"IsHidden"`
	IsHiddenRemotely                 bool     `json:"IsHiddenRemotely"`
	IsHiddenFromUnusedDevices        bool     `json:"IsHiddenFromUnusedDevices"`
	IsDisabled                       bool     `json:"IsDisabled"`
	MaxParentalRating                *int     `json:"MaxParentalRating,omitempty"`
	BlockedTags                      []string `json:"BlockedTags"`
	IsTagBlockingModeInclusive       bool     `json:"IsTagBlockingModeInclusive"`
	EnableUserPreferenceAccess       bool     `json:"EnableUserPreferenceAccess"`
	AccessSchedules                  []interface{} `json:"AccessSchedules"`
	BlockUnratedItems                []string `json:"BlockUnratedItems"`
	EnableRemoteControlOfOtherUsers  bool     `json:"EnableRemoteControlOfOtherUsers"`
	EnableSharedDeviceControl        bool     `json:"EnableSharedDeviceControl"`
	EnableRemoteAccess               bool     `json:"EnableRemoteAccess"`
	EnableLiveTvManagement           bool     `json:"EnableLiveTvManagement"`
	EnableLiveTvAccess               bool     `json:"EnableLiveTvAccess"`
	EnableMediaPlayback              bool     `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding   bool     `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding   bool     `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing           bool     `json:"EnablePlaybackRemuxing"`
	EnableContentDeletion            bool     `json:"EnableContentDeletion"`
	EnableContentDeletionFromFolders []string `json:"EnableContentDeletionFromFolders"`
	EnableContentDownloading         bool     `json:"EnableContentDownloading"`
	EnableSubtitleDownloading        bool     `json:"EnableSubtitleDownloading"`
	EnableSubtitleManagement         bool     `json:"EnableSubtitleManagement"`
	EnableSyncTranscoding            bool     `json:"EnableSyncTranscoding"`
	EnableMediaConversion            bool     `json:"EnableMediaConversion"`
	EnabledDevices                   []string `json:"EnabledDevices"`
	EnableAllDevices                 bool     `json:"EnableAllDevices"`
	EnabledChannels                  []string `json:"EnabledChannels"`
	EnableAllChannels                bool     `json:"EnableAllChannels"`
	EnabledFolders                   []string `json:"EnabledFolders"`
	EnableAllFolders                 bool     `json:"EnableAllFolders"`
	InvalidLoginAttemptCount         int      `json:"InvalidLoginAttemptCount"`
	EnablePublicSharing              bool     `json:"EnablePublicSharing"`
	RemoteClientBitrateLimit         int      `json:"RemoteClientBitrateLimit"`
	AuthenticationProviderId         string   `json:"AuthenticationProviderId"`
	ExcludedSubFolders               []string `json:"ExcludedSubFolders"`
	SimultaneousStreamLimit          int      `json:"SimultaneousStreamLimit"`
	EnabledChannelsForAllFolders     bool     `json:"EnabledChannelsForAllFolders"`
}

// AuthResult 认证结果
type AuthResult struct {
	User        *EmbyUser `json:"User"`
	AccessToken string    `json:"AccessToken"`
	ServerID    string    `json:"ServerId"`
}

// doRequestWithToken 使用用户token执行HTTP请求
func (c *EmbyClient) doRequestWithToken(method, path string, body interface{}, userToken string) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	// 使用用户token而不是API Key
	if userToken != "" {
		req.Header.Set("X-Emby-Token", userToken)
	} else {
		req.Header.Set("X-Emby-Token", c.apiKey)
	}
	req.Header.Set("X-Emby-Client", "EmbyUserSystem")
	req.Header.Set("X-Emby-Device-Name", "UserManagementServer")
	req.Header.Set("X-Emby-Device-Id", "emby-user-system-server")
	req.Header.Set("X-Emby-Client-Version", "1.0.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Emby API错误 (HTTP %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// GetUserViewsWithToken 使用用户token获取用户可见的媒体库视图
func (c *EmbyClient) GetUserViewsWithToken(userID, userToken string) ([]MediaFolder, error) {
	data, err := c.doRequestWithToken("GET", fmt.Sprintf("/Users/%s/Views", userID), nil, userToken)
	if err != nil {
		return nil, err
	}

	var resp MediaFoldersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetItemsWithToken 使用用户token获取媒体项列表
func (c *EmbyClient) GetItemsWithToken(params *GetItemsParams, userToken string) (*EmbyItemsResponse, error) {
	query := url.Values{}

	if params.UserID != "" {
		query.Set("UserId", params.UserID)
	}
	if params.ParentID != "" {
		query.Set("ParentId", params.ParentID)
	}
	if params.IncludeItemTypes != "" {
		query.Set("IncludeItemTypes", params.IncludeItemTypes)
	}
	if params.SortBy != "" {
		query.Set("SortBy", params.SortBy)
	}
	if params.SortOrder != "" {
		query.Set("SortOrder", params.SortOrder)
	}
	if params.StartIndex > 0 {
		query.Set("StartIndex", fmt.Sprintf("%d", params.StartIndex))
	}
	if params.Limit > 0 {
		query.Set("Limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Recursive {
		query.Set("Recursive", "true")
	}
	if params.Fields != "" {
		query.Set("Fields", params.Fields)
	}
	if params.SearchTerm != "" {
		query.Set("SearchTerm", params.SearchTerm)
	}
	if params.Filters != "" {
		query.Set("Filters", params.Filters)
	}
	if params.Genres != "" {
		query.Set("Genres", params.Genres)
	}
	if params.Years != "" {
		query.Set("Years", params.Years)
	}

	path := "/Items"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	data, err := c.doRequestWithToken("GET", path, nil, userToken)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetItemByIDWithToken 使用用户token获取单个媒体项详情
func (c *EmbyClient) GetItemByIDWithToken(userID, itemID, userToken string) (*EmbyMediaItem, error) {
	path := fmt.Sprintf("/Users/%s/Items/%s", userID, itemID)
	data, err := c.doRequestWithToken("GET", path, nil, userToken)
	if err != nil {
		return nil, err
	}

	var item EmbyMediaItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// GetSeasonsWithToken 使用用户token获取剧集的季列表
func (c *EmbyClient) GetSeasonsWithToken(userID, seriesID, userToken string) ([]EmbyMediaItem, error) {
	path := fmt.Sprintf("/Shows/%s/Seasons?UserId=%s", seriesID, userID)
	data, err := c.doRequestWithToken("GET", path, nil, userToken)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetEpisodesWithToken 使用用户token获取季的剧集列表
func (c *EmbyClient) GetEpisodesWithToken(userID, seriesID, seasonID, userToken string) ([]EmbyMediaItem, error) {
	path := fmt.Sprintf("/Shows/%s/Episodes?UserId=%s&SeasonId=%s", seriesID, userID, seasonID)
	data, err := c.doRequestWithToken("GET", path, nil, userToken)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetAllEpisodesWithToken 使用用户token获取剧集的所有集数（不分季）
func (c *EmbyClient) GetAllEpisodesWithToken(userID, seriesID, userToken string) ([]EmbyMediaItem, error) {
	// 使用 Items API 获取所有 Episode 类型的子项
	path := fmt.Sprintf("/Shows/%s/Episodes?UserId=%s&Fields=Overview,UserData", seriesID, userID)
	data, err := c.doRequestWithToken("GET", path, nil, userToken)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetAllEpisodes 获取剧集的所有集数（不分季）
func (c *EmbyClient) GetAllEpisodes(userID, seriesID string) ([]EmbyMediaItem, error) {
	path := fmt.Sprintf("/Shows/%s/Episodes?UserId=%s&Fields=Overview,UserData", seriesID, userID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// doRequest 执行HTTP请求
func (c *EmbyClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	reqURL := c.baseURL + path
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Token", c.apiKey)
	// Emby客户端标识
	req.Header.Set("X-Emby-Client", "EmbyUserSystem")
	req.Header.Set("X-Emby-Device-Name", "UserManagementServer")
	req.Header.Set("X-Emby-Device-Id", "emby-user-system-server")
	req.Header.Set("X-Emby-Client-Version", "1.0.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Emby API错误 (HTTP %d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}


// GetUsers 获取所有用户列表
func (c *EmbyClient) GetUsers() ([]EmbyUser, error) {
	data, err := c.doRequest("GET", "/Users", nil)
	if err != nil {
		return nil, err
	}

	var users []EmbyUser
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("解析用户列表失败: %w", err)
	}

	return users, nil
}

// GetUserByID 根据ID获取用户
func (c *EmbyClient) GetUserByID(userID string) (*EmbyUser, error) {
	data, err := c.doRequest("GET", "/Users/"+userID, nil)
	if err != nil {
		return nil, err
	}

	var user EmbyUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	return &user, nil
}

// GetUserByName 根据用户名获取用户
func (c *EmbyClient) GetUserByName(username string) (*EmbyUser, error) {
	users, err := c.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.Name == username {
			return &u, nil
		}
	}

	return nil, errors.New("用户不存在")
}

// EmbyCreateUserRequest 创建用户请求
type EmbyCreateUserRequest struct {
	Name string `json:"Name"`
}

// CreateUser 创建新用户
func (c *EmbyClient) CreateUser(username string) (*EmbyUser, error) {
	req := &EmbyCreateUserRequest{Name: username}
	data, err := c.doRequest("POST", "/Users/New", req)
	if err != nil {
		return nil, err
	}

	var user EmbyUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("解析创建用户响应失败: %w", err)
	}

	return &user, nil
}

// CreateUserWithPassword 创建用户并设置密码
func (c *EmbyClient) CreateUserWithPassword(username, password string) (*EmbyUser, error) {
	// 先创建用户
	user, err := c.CreateUser(username)
	if err != nil {
		return nil, err
	}

	// 设置密码
	if err := c.UpdateUserPassword(user.ID, "", password); err != nil {
		// 如果设置密码失败，删除用户
		c.DeleteUser(user.ID)
		return nil, fmt.Errorf("设置密码失败: %w", err)
	}

	return user, nil
}

// CreateUserWithTemplate 创建用户并使用模板用户的权限
func (c *EmbyClient) CreateUserWithTemplate(username, password, templateUsername string) (*EmbyUser, error) {
	// 先创建用户
	user, err := c.CreateUser(username)
	if err != nil {
		return nil, err
	}

	// 设置密码
	if err := c.UpdateUserPassword(user.ID, "", password); err != nil {
		c.DeleteUser(user.ID)
		return nil, fmt.Errorf("设置密码失败: %w", err)
	}

	// 如果指定了模板用户，复制其权限
	if templateUsername != "" {
		templateUser, err := c.GetUserByName(templateUsername)
		if err == nil && templateUser.Policy != nil {
			// 复制模板用户的权限，但保持非管理员
			policy := *templateUser.Policy
			policy.IsAdministrator = false
			policy.IsDisabled = false
			if err := c.UpdateUserPolicy(user.ID, &policy); err != nil {
				// 权限复制失败不影响用户创建
				fmt.Printf("复制模板用户权限失败: %v\n", err)
			}
		}
	}

	return user, nil
}

// UpdateUserPassword 更新用户密码
func (c *EmbyClient) UpdateUserPassword(userID, currentPassword, newPassword string) error {
	req := map[string]string{
		"CurrentPw": currentPassword,
		"NewPw":     newPassword,
	}

	_, err := c.doRequest("POST", fmt.Sprintf("/Users/%s/Password", userID), req)
	return err
}

// UpdateUserPasswordByName 根据用户名更新密码
func (c *EmbyClient) UpdateUserPasswordByName(username, newPassword string) error {
	user, err := c.GetUserByName(username)
	if err != nil {
		return err
	}

	return c.UpdateUserPassword(user.ID, "", newPassword)
}

// DeleteUser 删除用户
func (c *EmbyClient) DeleteUser(userID string) error {
	_, err := c.doRequest("DELETE", "/Users/"+userID, nil)
	return err
}

// DeleteUserByName 根据用户名删除用户
func (c *EmbyClient) DeleteUserByName(username string) error {
	user, err := c.GetUserByName(username)
	if err != nil {
		return nil // 用户不存在，无需删除
	}

	return c.DeleteUser(user.ID)
}

// UpdateUserPolicy 更新用户策略/权限
func (c *EmbyClient) UpdateUserPolicy(userID string, policy *UserPolicy) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/Users/%s/Policy", userID), policy)
	return err
}

// EnableUser 启用用户
func (c *EmbyClient) EnableUser(userID string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.IsDisabled = false

	return c.UpdateUserPolicy(userID, user.Policy)
}

// DisableUser 禁用用户
func (c *EmbyClient) DisableUser(userID string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.IsDisabled = true

	return c.UpdateUserPolicy(userID, user.Policy)
}

// EnableUserByName 根据用户名启用用户
func (c *EmbyClient) EnableUserByName(username string) error {
	user, err := c.GetUserByName(username)
	if err != nil {
		return err
	}
	return c.EnableUser(user.ID)
}

// DisableUserByName 根据用户名禁用用户
func (c *EmbyClient) DisableUserByName(username string) error {
	user, err := c.GetUserByName(username)
	if err != nil {
		return err
	}
	return c.DisableUser(user.ID)
}

// AuthenticateUser 用户认证（登录）
func (c *EmbyClient) AuthenticateUser(username, password string) (*AuthResult, error) {
	// 构建认证URL
	authURL := fmt.Sprintf("%s/Users/AuthenticateByName", c.baseURL)

	reqBody := map[string]string{
		"Username": username,
		"Pw":       password,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Client", "EmbyUserSystem")
	req.Header.Set("X-Emby-Device-Name", "UserManagementServer")
	req.Header.Set("X-Emby-Device-Id", "emby-user-system-server")
	req.Header.Set("X-Emby-Client-Version", "1.0.0")
	// 认证请求需要Authorization头
	req.Header.Set("X-Emby-Authorization", `MediaBrowser Client="EmbyUserSystem", Device="Server", DeviceId="emby-user-system", Version="1.0.0"`)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 401 {
		return nil, errors.New("用户名或密码错误")
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("认证失败 (HTTP %d): %s", resp.StatusCode, string(data))
	}

	var result AuthResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}


// ========== 媒体库 API ==========

// MediaFolder 媒体文件夹/库
type MediaFolder struct {
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	Type             string `json:"Type"`
	CollectionType   string `json:"CollectionType"`
	ImageTags        map[string]string `json:"ImageTags"`
	BackdropImageTags []string `json:"BackdropImageTags"`
	ChildCount       int    `json:"ChildCount"`
}

// MediaFoldersResponse 媒体库列表响应
type MediaFoldersResponse struct {
	Items            []MediaFolder `json:"Items"`
	TotalRecordCount int           `json:"TotalRecordCount"`
}

// GetMediaFolders 获取媒体库列表
func (c *EmbyClient) GetMediaFolders() ([]MediaFolder, error) {
	data, err := c.doRequest("GET", "/Library/MediaFolders", nil)
	if err != nil {
		return nil, err
	}

	var resp MediaFoldersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetUserViews 获取用户可见的媒体库视图
func (c *EmbyClient) GetUserViews(userID string) ([]MediaFolder, error) {
	data, err := c.doRequest("GET", fmt.Sprintf("/Users/%s/Views", userID), nil)
	if err != nil {
		return nil, err
	}

	var resp MediaFoldersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ========== 媒体项 API ==========

// EmbyMediaItem Emby媒体项
type EmbyMediaItem struct {
	ID                  string            `json:"Id"`
	Name                string            `json:"Name"`
	OriginalTitle       string            `json:"OriginalTitle,omitempty"`
	Type                string            `json:"Type"` // Movie, Series, Episode, Season, etc.
	Overview            string            `json:"Overview,omitempty"`
	ProductionYear      int               `json:"ProductionYear,omitempty"`
	PremiereDate        string            `json:"PremiereDate,omitempty"`
	CommunityRating     float64           `json:"CommunityRating,omitempty"`
	OfficialRating      string            `json:"OfficialRating,omitempty"`
	RunTimeTicks        int64             `json:"RunTimeTicks,omitempty"`
	ImageTags           map[string]string `json:"ImageTags,omitempty"`
	BackdropImageTags   []string          `json:"BackdropImageTags,omitempty"`
	ParentID            string            `json:"ParentId,omitempty"`
	SeriesID            string            `json:"SeriesId,omitempty"`
	SeriesName          string            `json:"SeriesName,omitempty"`
	SeasonID            string            `json:"SeasonId,omitempty"`
	SeasonName          string            `json:"SeasonName,omitempty"`
	IndexNumber         int               `json:"IndexNumber,omitempty"`
	ParentIndexNumber   int               `json:"ParentIndexNumber,omitempty"`
	Genres              []string          `json:"Genres,omitempty"`
	Studios             []NamedItem       `json:"Studios,omitempty"`
	People              []PersonInfo      `json:"People,omitempty"`
	MediaSources        []MediaSource     `json:"MediaSources,omitempty"`
	UserData            *UserItemData     `json:"UserData,omitempty"`
	ChildCount          int               `json:"ChildCount,omitempty"`
	RecursiveItemCount  int               `json:"RecursiveItemCount,omitempty"`
	Container           string            `json:"Container,omitempty"`
	Path                string            `json:"Path,omitempty"`
}

// NamedItem 命名项
type NamedItem struct {
	Name string `json:"Name"`
	ID   string `json:"Id,omitempty"`
}

// PersonInfo 人员信息
type PersonInfo struct {
	Name            string `json:"Name"`
	ID              string `json:"Id"`
	Role            string `json:"Role,omitempty"`
	Type            string `json:"Type"` // Actor, Director, Writer, etc.
	PrimaryImageTag string `json:"PrimaryImageTag,omitempty"`
}

// MediaSource 媒体源
type MediaSource struct {
	ID               string        `json:"Id"`
	Path             string        `json:"Path"`
	Container        string        `json:"Container"`
	Size             int64         `json:"Size"`
	Bitrate          int           `json:"Bitrate"`
	RunTimeTicks     int64         `json:"RunTimeTicks"`
	MediaStreams     []MediaStream `json:"MediaStreams"`
	VideoType        string        `json:"VideoType"`
}

// MediaStream 媒体流
type MediaStream struct {
	Codec            string `json:"Codec"`
	Type             string `json:"Type"` // Video, Audio, Subtitle
	Language         string `json:"Language,omitempty"`
	Title            string `json:"Title,omitempty"`
	DisplayTitle     string `json:"DisplayTitle,omitempty"`
	Width            int    `json:"Width,omitempty"`
	Height           int    `json:"Height,omitempty"`
	BitRate          int    `json:"BitRate,omitempty"`
	Channels         int    `json:"Channels,omitempty"`
	SampleRate       int    `json:"SampleRate,omitempty"`
	IsDefault        bool   `json:"IsDefault"`
	IsForced         bool   `json:"IsForced"`
	IsExternal       bool   `json:"IsExternal"`
	Index            int    `json:"Index"`
}

// UserItemData 用户媒体数据
type UserItemData struct {
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	PlayCount             int    `json:"PlayCount"`
	IsFavorite            bool   `json:"IsFavorite"`
	Played                bool   `json:"Played"`
	Key                   string `json:"Key"`
}

// EmbyItemsResponse 媒体项列表响应
type EmbyItemsResponse struct {
	Items            []EmbyMediaItem `json:"Items"`
	TotalRecordCount int             `json:"TotalRecordCount"`
}

// GetItemsParams 获取媒体项参数
type GetItemsParams struct {
	UserID           string
	ParentID         string
	IncludeItemTypes string // Movie,Series,Episode,etc.
	SortBy           string
	SortOrder        string
	StartIndex       int
	Limit            int
	Recursive        bool
	Fields           string
	SearchTerm       string
	Filters          string
	Genres           string
	Years            string
}

// GetItems 获取媒体项列表
func (c *EmbyClient) GetItems(params *GetItemsParams) (*EmbyItemsResponse, error) {
	query := url.Values{}
	
	if params.UserID != "" {
		query.Set("UserId", params.UserID)
	}
	if params.ParentID != "" {
		query.Set("ParentId", params.ParentID)
	}
	if params.IncludeItemTypes != "" {
		query.Set("IncludeItemTypes", params.IncludeItemTypes)
	}
	if params.SortBy != "" {
		query.Set("SortBy", params.SortBy)
	}
	if params.SortOrder != "" {
		query.Set("SortOrder", params.SortOrder)
	}
	if params.StartIndex > 0 {
		query.Set("StartIndex", fmt.Sprintf("%d", params.StartIndex))
	}
	if params.Limit > 0 {
		query.Set("Limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Recursive {
		query.Set("Recursive", "true")
	}
	if params.Fields != "" {
		query.Set("Fields", params.Fields)
	}
	if params.SearchTerm != "" {
		query.Set("SearchTerm", params.SearchTerm)
	}
	if params.Filters != "" {
		query.Set("Filters", params.Filters)
	}
	if params.Genres != "" {
		query.Set("Genres", params.Genres)
	}
	if params.Years != "" {
		query.Set("Years", params.Years)
	}

	path := "/Items"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetItemByID 获取单个媒体项详情
func (c *EmbyClient) GetItemByID(userID, itemID string) (*EmbyMediaItem, error) {
	path := fmt.Sprintf("/Users/%s/Items/%s", userID, itemID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var item EmbyMediaItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// GetSeasons 获取剧集的季列表
func (c *EmbyClient) GetSeasons(userID, seriesID string) ([]EmbyMediaItem, error) {
	path := fmt.Sprintf("/Shows/%s/Seasons?UserId=%s", seriesID, userID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// GetEpisodes 获取季的剧集列表
func (c *EmbyClient) GetEpisodes(userID, seriesID, seasonID string) ([]EmbyMediaItem, error) {
	path := fmt.Sprintf("/Shows/%s/Episodes?UserId=%s&SeasonId=%s", seriesID, userID, seasonID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp EmbyItemsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// SearchItems 搜索媒体
func (c *EmbyClient) SearchItems(userID, searchTerm string, limit int) ([]EmbyMediaItem, error) {
	params := &GetItemsParams{
		UserID:           userID,
		SearchTerm:       searchTerm,
		IncludeItemTypes: "Movie,Series",
		Recursive:        true,
		Limit:            limit,
		Fields:           "Overview,Genres,CommunityRating,ProductionYear",
	}

	resp, err := c.GetItems(params)
	if err != nil {
		return nil, err
	}

	return resp.Items, nil
}

// ========== 图片 API ==========

// GetImageURL 获取图片URL
func (c *EmbyClient) GetImageURL(itemID, imageType string, maxWidth, maxHeight int) string {
	// 返回相对路径（不带前导斜杠），由前端通过代理访问
	// Emby官方API格式: /Items/{itemId}/Images/{imageType}
	url := fmt.Sprintf("Items/%s/Images/%s", itemID, imageType)
	if maxWidth > 0 {
		url += fmt.Sprintf("?maxWidth=%d", maxWidth)
		if maxHeight > 0 {
			url += fmt.Sprintf("&maxHeight=%d", maxHeight)
		}
	} else if maxHeight > 0 {
		url += fmt.Sprintf("?maxHeight=%d", maxHeight)
	}
	return url
}

// GetPrimaryImageURL 获取主图URL
func (c *EmbyClient) GetPrimaryImageURL(itemID string, maxWidth int) string {
	return c.GetImageURL(itemID, "Primary", maxWidth, 0)
}

// GetBackdropImageURL 获取背景图URL
func (c *EmbyClient) GetBackdropImageURL(itemID string, maxWidth int) string {
	return c.GetImageURL(itemID, "Backdrop", maxWidth, 0)
}

// ========== 系统信息 API ==========

// SystemInfo 系统信息
type SystemInfo struct {
	ServerName                 string `json:"ServerName"`
	Version                    string `json:"Version"`
	ID                         string `json:"Id"`
	OperatingSystem            string `json:"OperatingSystem"`
	OperatingSystemDisplayName string `json:"OperatingSystemDisplayName"`
	HasPendingRestart          bool   `json:"HasPendingRestart"`
	IsShuttingDown             bool   `json:"IsShuttingDown"`
	LocalAddress               string `json:"LocalAddress"`
	WanAddress                 string `json:"WanAddress"`
}

// GetSystemInfo 获取系统信息
func (c *EmbyClient) GetSystemInfo() (*SystemInfo, error) {
	data, err := c.doRequest("GET", "/System/Info", nil)
	if err != nil {
		return nil, err
	}

	var info SystemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Ping 检查服务器连接
func (c *EmbyClient) Ping() error {
	_, err := c.doRequest("GET", "/System/Ping", nil)
	return err
}

// ========== 播放信息 API ==========

// PlaybackInfo 播放信息
type PlaybackInfo struct {
	MediaSources []MediaSource `json:"MediaSources"`
	PlaySessionId string       `json:"PlaySessionId"`
}

// GetPlaybackInfo 获取播放信息
func (c *EmbyClient) GetPlaybackInfo(userID, itemID string) (*PlaybackInfo, error) {
	path := fmt.Sprintf("/Items/%s/PlaybackInfo?UserId=%s", itemID, userID)
	data, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var info PlaybackInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetStreamURL 获取流媒体URL
func (c *EmbyClient) GetStreamURL(itemID, mediaSourceID string, audioStreamIndex, subtitleStreamIndex int) string {
	url := fmt.Sprintf("%s/Videos/%s/stream?static=true&mediaSourceId=%s&api_key=%s",
		c.baseURL, itemID, mediaSourceID, c.apiKey)
	
	if audioStreamIndex >= 0 {
		url += fmt.Sprintf("&AudioStreamIndex=%d", audioStreamIndex)
	}
	if subtitleStreamIndex >= 0 {
		url += fmt.Sprintf("&SubtitleStreamIndex=%d", subtitleStreamIndex)
	}
	
	return url
}

// ========== 收藏/播放状态 API ==========

// MarkFavorite 标记为收藏
func (c *EmbyClient) MarkFavorite(userID, itemID string) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/Users/%s/FavoriteItems/%s", userID, itemID), nil)
	return err
}

// UnmarkFavorite 取消收藏
func (c *EmbyClient) UnmarkFavorite(userID, itemID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/Users/%s/FavoriteItems/%s", userID, itemID), nil)
	return err
}

// MarkPlayed 标记为已播放
func (c *EmbyClient) MarkPlayed(userID, itemID string) error {
	_, err := c.doRequest("POST", fmt.Sprintf("/Users/%s/PlayedItems/%s", userID, itemID), nil)
	return err
}

// UnmarkPlayed 取消已播放标记
func (c *EmbyClient) UnmarkPlayed(userID, itemID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/Users/%s/PlayedItems/%s", userID, itemID), nil)
	return err
}

// ReportPlaybackProgress 报告播放进度
func (c *EmbyClient) ReportPlaybackProgress(userID, itemID string, positionTicks int64, isPaused bool) error {
	req := map[string]interface{}{
		"ItemId":        itemID,
		"PositionTicks": positionTicks,
		"IsPaused":      isPaused,
	}
	_, err := c.doRequest("POST", "/Sessions/Playing/Progress", req)
	return err
}

// ReportPlaybackStopped 报告播放停止
func (c *EmbyClient) ReportPlaybackStopped(userID, itemID string, positionTicks int64) error {
	req := map[string]interface{}{
		"ItemId":        itemID,
		"PositionTicks": positionTicks,
	}
	_, err := c.doRequest("POST", "/Sessions/Playing/Stopped", req)
	return err
}

// ========== 设备和会话管理 API ==========

// EmbySession Emby会话信息
type EmbySession struct {
	ID                       string           `json:"Id"`
	UserID                   string           `json:"UserId,omitempty"`
	UserName                 string           `json:"UserName,omitempty"`
	Client                   string           `json:"Client"`
	DeviceID                 string           `json:"DeviceId"`
	DeviceName               string           `json:"DeviceName"`
	DeviceType               string           `json:"DeviceType,omitempty"`
	ApplicationVersion       string           `json:"ApplicationVersion,omitempty"`
	LastActivityDate         string           `json:"LastActivityDate"`
	LastPlaybackCheckIn      string           `json:"LastPlaybackCheckIn,omitempty"`
	NowPlayingItem           *NowPlayingItem  `json:"NowPlayingItem,omitempty"`
	PlayState                *PlayState       `json:"PlayState,omitempty"`
	RemoteEndPoint           string           `json:"RemoteEndPoint,omitempty"`
	SupportsRemoteControl    bool             `json:"SupportsRemoteControl"`
	SupportsMediaControl     bool             `json:"SupportsMediaControl"`
	PlayableMediaTypes       []string         `json:"PlayableMediaTypes,omitempty"`
	SupportedCommands        []string         `json:"SupportedCommands,omitempty"`
	TranscodingInfo          *TranscodingInfo `json:"TranscodingInfo,omitempty"`
}

// NowPlayingItem 正在播放的媒体项
type NowPlayingItem struct {
	ID               string `json:"Id"`
	Name             string `json:"Name"`
	Type             string `json:"Type"`
	SeriesName       string `json:"SeriesName,omitempty"`
	SeasonName       string `json:"SeasonName,omitempty"`
	IndexNumber      int    `json:"IndexNumber,omitempty"`
	ParentIndexNumber int   `json:"ParentIndexNumber,omitempty"`
	RunTimeTicks     int64  `json:"RunTimeTicks,omitempty"`
	ProductionYear   int    `json:"ProductionYear,omitempty"`
}

// PlayState 播放状态
type PlayState struct {
	PositionTicks     int64  `json:"PositionTicks"`
	CanSeek           bool   `json:"CanSeek"`
	IsPaused          bool   `json:"IsPaused"`
	IsMuted           bool   `json:"IsMuted"`
	VolumeLevel       int    `json:"VolumeLevel,omitempty"`
	AudioStreamIndex  int    `json:"AudioStreamIndex,omitempty"`
	SubtitleStreamIndex int  `json:"SubtitleStreamIndex,omitempty"`
	MediaSourceID     string `json:"MediaSourceId,omitempty"`
	PlayMethod        string `json:"PlayMethod,omitempty"`
	RepeatMode        string `json:"RepeatMode,omitempty"`
}

// TranscodingInfo 转码信息
type TranscodingInfo struct {
	AudioCodec             string  `json:"AudioCodec,omitempty"`
	VideoCodec             string  `json:"VideoCodec,omitempty"`
	Container              string  `json:"Container,omitempty"`
	IsVideoDirect          bool    `json:"IsVideoDirect"`
	IsAudioDirect          bool    `json:"IsAudioDirect"`
	Bitrate                int     `json:"Bitrate,omitempty"`
	Width                  int     `json:"Width,omitempty"`
	Height                 int     `json:"Height,omitempty"`
	AudioChannels          int     `json:"AudioChannels,omitempty"`
	TranscodeReasons       []string `json:"TranscodeReasons,omitempty"`
	CompletionPercentage   float64 `json:"CompletionPercentage,omitempty"`
}

// EmbyDevice Emby设备信息
type EmbyDevice struct {
	ID               string `json:"Id"`                         // Emby 内部设备记录 ID
	ReportedDeviceID string `json:"ReportedDeviceId,omitempty"` // 客户端报告的设备 ID（用于 EnabledDevices 策略）
	Name             string `json:"Name"`                       // 设备名称
	AppName          string `json:"AppName"`                    // 应用名称
	AppVersion       string `json:"AppVersion"`                 // 应用版本
	LastUserID       string `json:"LastUserId,omitempty"`       // 最后使用用户 ID
	LastUserName     string `json:"LastUserName,omitempty"`     // 最后使用用户名
	DateLastActivity string `json:"DateLastActivity"`           // 最后活动时间
	IconURL          string `json:"IconUrl,omitempty"`          // 图标 URL
	IPAddress        string `json:"IpAddress,omitempty"`        // IP 地址
}

// DevicesResponse 设备列表响应
type DevicesResponse struct {
	Items            []EmbyDevice `json:"Items"`
	TotalRecordCount int          `json:"TotalRecordCount"`
}

// GetSessions 获取所有活动会话
func (c *EmbyClient) GetSessions() ([]EmbySession, error) {
	data, err := c.doRequest("GET", "/Sessions", nil)
	if err != nil {
		return nil, err
	}

	var sessions []EmbySession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return nil, fmt.Errorf("解析会话列表失败: %w", err)
	}

	return sessions, nil
}

// GetSessionsByUserID 获取指定用户的活动会话
func (c *EmbyClient) GetSessionsByUserID(userID string) ([]EmbySession, error) {
	sessions, err := c.GetSessions()
	if err != nil {
		return nil, err
	}

	var userSessions []EmbySession
	// 忽略大小写比较 UserID
	userIDLower := strings.ToLower(userID)
	for _, s := range sessions {
		if strings.ToLower(s.UserID) == userIDLower {
			userSessions = append(userSessions, s)
		}
	}

	return userSessions, nil
}

// GetDevices 获取所有注册设备
func (c *EmbyClient) GetDevices() ([]EmbyDevice, error) {
	data, err := c.doRequest("GET", "/Devices", nil)
	if err != nil {
		return nil, err
	}

	var resp DevicesResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("解析设备列表失败: %w", err)
	}

	return resp.Items, nil
}

// GetDevicesByUserID 获取指定用户的设备
func (c *EmbyClient) GetDevicesByUserID(userID string) ([]EmbyDevice, error) {
	devices, err := c.GetDevices()
	if err != nil {
		return nil, err
	}

	var userDevices []EmbyDevice
	for _, d := range devices {
		if d.LastUserID == userID {
			userDevices = append(userDevices, d)
		}
	}

	return userDevices, nil
}

// DeleteDevice 删除设备
func (c *EmbyClient) DeleteDevice(deviceID string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/Devices?Id=%s", url.QueryEscape(deviceID)), nil)
	return err
}

// KillSession 终止会话（强制下线）
func (c *EmbyClient) KillSession(sessionID string) error {
	// 尝试发送停止播放命令
	// Emby API: POST /Sessions/{sessionId}/Playing/Stop
	_, err := c.doRequest("POST", fmt.Sprintf("/Sessions/%s/Playing/Stop", sessionID), nil)
	if err != nil {
		// 如果停止播放失败，尝试发送消息通知用户
		c.SendMessageToSession(sessionID, "系统通知", "您的播放已被管理员终止", 5000)
	}
	return err
}

// StopPlayback 停止指定会话的播放
func (c *EmbyClient) StopPlayback(sessionID string) error {
	// Emby API: POST /Sessions/{sessionId}/Playing/Stop
	_, err := c.doRequest("POST", fmt.Sprintf("/Sessions/%s/Playing/Stop", sessionID), nil)
	return err
}

// SendMessageToSession 向会话发送消息
func (c *EmbyClient) SendMessageToSession(sessionID, header, text string, timeoutMs int) error {
	req := map[string]interface{}{
		"Header":    header,
		"Text":      text,
		"TimeoutMs": timeoutMs,
	}
	_, err := c.doRequest("POST", fmt.Sprintf("/Sessions/%s/Message", sessionID), req)
	return err
}

// SetUserSimultaneousStreamLimit 设置用户同时在线流数限制
func (c *EmbyClient) SetUserSimultaneousStreamLimit(userID string, limit int) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.SimultaneousStreamLimit = limit

	return c.UpdateUserPolicy(userID, user.Policy)
}

// GetUserSimultaneousStreamLimit 获取用户同时在线流数限制
func (c *EmbyClient) GetUserSimultaneousStreamLimit(userID string) (int, error) {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return 0, err
	}

	if user.Policy == nil {
		return 0, nil
	}

	return user.Policy.SimultaneousStreamLimit, nil
}


// ========== 用户设备策略管理 API ==========

// GetUserEnabledDevices 获取用户允许的设备列表
func (c *EmbyClient) GetUserEnabledDevices(userID string) ([]string, bool, error) {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return nil, false, err
	}

	if user.Policy == nil {
		return []string{}, true, nil
	}

	return user.Policy.EnabledDevices, user.Policy.EnableAllDevices, nil
}

// SetUserEnableAllDevices 设置用户是否允许所有设备
func (c *EmbyClient) SetUserEnableAllDevices(userID string, enableAll bool) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.EnableAllDevices = enableAll

	return c.UpdateUserPolicy(userID, user.Policy)
}

// SetUserEnabledDevices 设置用户允许的设备列表
func (c *EmbyClient) SetUserEnabledDevices(userID string, deviceIDs []string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.EnabledDevices = deviceIDs

	return c.UpdateUserPolicy(userID, user.Policy)
}

// AddDeviceToUserWhitelist 添加设备到用户白名单
func (c *EmbyClient) AddDeviceToUserWhitelist(userID, deviceID string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}

	// 检查设备是否已在白名单中
	for _, d := range user.Policy.EnabledDevices {
		if d == deviceID {
			return nil // 已存在，无需添加
		}
	}

	user.Policy.EnabledDevices = append(user.Policy.EnabledDevices, deviceID)
	return c.UpdateUserPolicy(userID, user.Policy)
}

// RemoveDeviceFromUserWhitelist 从用户白名单移除设备
func (c *EmbyClient) RemoveDeviceFromUserWhitelist(userID, deviceID string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		return nil
	}

	// 过滤掉要移除的设备
	newDevices := make([]string, 0, len(user.Policy.EnabledDevices))
	for _, d := range user.Policy.EnabledDevices {
		if d != deviceID {
			newDevices = append(newDevices, d)
		}
	}

	user.Policy.EnabledDevices = newDevices
	return c.UpdateUserPolicy(userID, user.Policy)
}

// SetUserDevicePolicy 设置用户设备策略（同时设置EnableAllDevices和EnabledDevices）
func (c *EmbyClient) SetUserDevicePolicy(userID string, enableAll bool, deviceIDs []string) error {
	user, err := c.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.Policy == nil {
		user.Policy = &UserPolicy{}
	}
	user.Policy.EnableAllDevices = enableAll
	user.Policy.EnabledDevices = deviceIDs

	return c.UpdateUserPolicy(userID, user.Policy)
}
