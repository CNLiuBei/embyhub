package emby

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"embyhub/config"
	"embyhub/internal/model"
)

type Client struct {
	ServerURL string
	APIKey    string
	Timeout   time.Duration
	client    *http.Client
}

// NewClient 创建Emby客户端
func NewClient(cfg *config.EmbyConfig) *Client {
	return &Client{
		ServerURL: cfg.ServerURL,
		APIKey:    cfg.APIKey,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// GetUsers 获取Emby用户列表
func (c *Client) GetUsers() ([]*model.EmbyUser, error) {
	url := fmt.Sprintf("%s/Users", c.ServerURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var users []*model.EmbyUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return users, nil
}

// GetUserByName 根据用户名获取Emby用户
func (c *Client) GetUserByName(username string) (*model.EmbyUser, error) {
	users, err := c.GetUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Name == username {
			return user, nil
		}
	}
	return nil, nil // 未找到返回nil
}

// GetUser 获取单个Emby用户信息
func (c *Client) GetUser(userID string) (*model.EmbyUser, error) {
	url := fmt.Sprintf("%s/Users/%s", c.ServerURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var user model.EmbyUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &user, nil
}

// TestConnection 测试Emby连接
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/System/Info/Public", c.ServerURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("连接Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Emby服务器返回错误: %d", resp.StatusCode)
	}

	return nil
}

// CreateUser 在Emby服务器创建用户
func (c *Client) CreateUser(username, password string) (*model.EmbyUser, error) {
	url := fmt.Sprintf("%s/Users/New", c.ServerURL)

	// 构建请求体
	requestBody := map[string]interface{}{
		"Name":     username,
		"Password": password,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var user model.EmbyUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &user, nil
}

// SetUserPassword 设置Emby用户密码
func (c *Client) SetUserPassword(userID, password string) error {
	url := fmt.Sprintf("%s/Users/%s/Password", c.ServerURL, userID)

	// 构建请求体
	requestBody := map[string]interface{}{
		"CurrentPw": "",
		"NewPw":     password,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("设置密码失败: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteUser 删除Emby用户
func (c *Client) DeleteUser(userID string) error {
	url := fmt.Sprintf("%s/Users/%s", c.ServerURL, userID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除用户失败: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// ========== 媒体库相关API ==========

// MediaLibrary 媒体库信息
type MediaLibrary struct {
	Id              string   `json:"Id,omitempty"` // Views API 返回 Id
	Name            string   `json:"Name"`
	CollectionType  string   `json:"CollectionType"`
	ItemId          string   `json:"ItemId,omitempty"` // VirtualFolders API 返回 ItemId
	Locations       []string `json:"Locations,omitempty"`
	RefreshStatus   string   `json:"RefreshStatus,omitempty"`
	PrimaryImageTag string   `json:"PrimaryImageTag,omitempty"`
}

// MediaItem 媒体项目
type MediaItem struct {
	Id                 string      `json:"Id"`
	Name               string      `json:"Name"`
	Type               string      `json:"Type"`
	Overview           string      `json:"Overview,omitempty"`
	ProductionYear     int         `json:"ProductionYear,omitempty"`
	CommunityRating    float64     `json:"CommunityRating,omitempty"`
	OfficialRating     string      `json:"OfficialRating,omitempty"`
	RunTimeTicks       int64       `json:"RunTimeTicks,omitempty"`
	PremiereDate       string      `json:"PremiereDate,omitempty"`
	DateCreated        string      `json:"DateCreated,omitempty"`
	Genres             []string    `json:"Genres,omitempty"`
	Studios            []NamedItem `json:"Studios,omitempty"`
	People             []Person    `json:"People,omitempty"`
	ImageTags          ImageTags   `json:"ImageTags,omitempty"`
	BackdropImageTags  []string    `json:"BackdropImageTags,omitempty"`
	ParentId           string      `json:"ParentId,omitempty"`
	SeriesId           string      `json:"SeriesId,omitempty"`
	SeriesName         string      `json:"SeriesName,omitempty"`
	SeasonId           string      `json:"SeasonId,omitempty"`
	SeasonName         string      `json:"SeasonName,omitempty"`
	IndexNumber        int         `json:"IndexNumber,omitempty"`
	ParentIndexNumber  int         `json:"ParentIndexNumber,omitempty"`
	ChildCount         int         `json:"ChildCount,omitempty"`
	RecursiveItemCount int         `json:"RecursiveItemCount,omitempty"`
	MediaSources       []any       `json:"MediaSources,omitempty"`
	UserData           *UserData   `json:"UserData,omitempty"`
}

type NamedItem struct {
	Name string      `json:"Name"`
	Id   interface{} `json:"Id,omitempty"` // Emby 返回数字或字符串
}

type Person struct {
	Name            string      `json:"Name"`
	Id              interface{} `json:"Id,omitempty"` // Emby 返回数字或字符串
	Role            string      `json:"Role,omitempty"`
	Type            string      `json:"Type,omitempty"`
	PrimaryImageTag string      `json:"PrimaryImageTag,omitempty"`
}

type ImageTags struct {
	Primary string `json:"Primary,omitempty"`
	Thumb   string `json:"Thumb,omitempty"`
	Banner  string `json:"Banner,omitempty"`
	Logo    string `json:"Logo,omitempty"`
}

type UserData struct {
	PlaybackPositionTicks int64  `json:"PlaybackPositionTicks"`
	PlayCount             int    `json:"PlayCount"`
	IsFavorite            bool   `json:"IsFavorite"`
	Played                bool   `json:"Played"`
	LastPlayedDate        string `json:"LastPlayedDate,omitempty"`
}

// MediaItemsResponse 媒体项目列表响应
type MediaItemsResponse struct {
	Items            []MediaItem `json:"Items"`
	TotalRecordCount int         `json:"TotalRecordCount"`
}

// UserViewsResponse 用户视图响应
type UserViewsResponse struct {
	Items []MediaLibrary `json:"Items"`
}

// GetLibraries 获取媒体库列表（使用用户视图API，保持与Emby一致的顺序）
func (c *Client) GetLibraries(userId string) ([]MediaLibrary, error) {
	var apiUrl string
	if userId != "" {
		// 使用用户视图API，获取按用户设置排序的媒体库
		apiUrl = fmt.Sprintf("%s/Users/%s/Views", c.ServerURL, userId)
	} else {
		// 回退到旧API
		apiUrl = fmt.Sprintf("%s/Library/VirtualFolders", c.ServerURL)
	}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	if userId != "" {
		// 用户视图API返回格式不同
		var viewsResp UserViewsResponse
		if err := json.NewDecoder(resp.Body).Decode(&viewsResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		// 填充 ItemId 字段（Views API 返回的是 Id）
		for i := range viewsResp.Items {
			if viewsResp.Items[i].ItemId == "" {
				viewsResp.Items[i].ItemId = viewsResp.Items[i].Id
			}
		}
		return viewsResp.Items, nil
	}

	var libraries []MediaLibrary
	if err := json.NewDecoder(resp.Body).Decode(&libraries); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return libraries, nil
}

// GetItems 获取媒体项目列表
func (c *Client) GetItems(parentId string, itemType string, startIndex, limit int, sortBy, sortOrder, searchTerm string) (*MediaItemsResponse, error) {
	// 不使用 Recursive=true，只获取直接子项（避免显示到电视剧的每一集）
	apiUrl := fmt.Sprintf("%s/Items?Fields=Overview,Genres,Studios,People,DateCreated,PremiereDate,CommunityRating,OfficialRating,ChildCount,RecursiveItemCount&StartIndex=%d&Limit=%d",
		c.ServerURL, startIndex, limit)

	if parentId != "" {
		apiUrl += "&ParentId=" + parentId
	}
	if itemType != "" {
		apiUrl += "&IncludeItemTypes=" + itemType
	}
	if sortBy != "" {
		apiUrl += "&SortBy=" + sortBy
	}
	if sortOrder != "" {
		apiUrl += "&SortOrder=" + sortOrder
	}
	// 搜索关键词（需要URL编码）
	if searchTerm != "" {
		apiUrl += "&SearchTerm=" + url.QueryEscape(searchTerm) + "&Recursive=true"
	}

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var result MediaItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// GetItem 获取单个媒体项目详情
func (c *Client) GetItem(itemId string) (*MediaItem, error) {
	url := fmt.Sprintf("%s/Items/%s?Fields=Overview,Genres,Studios,People,DateCreated,PremiereDate,CommunityRating,OfficialRating,MediaSources", c.ServerURL, itemId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var item MediaItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &item, nil
}

// GetLatestItems 获取最新添加的媒体
func (c *Client) GetLatestItems(userId string, parentId string, limit int) ([]MediaItem, error) {
	if userId == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 只获取电影和电视剧，不显示具体集数；GroupItems=true 合并同一电视剧
	url := fmt.Sprintf("%s/Users/%s/Items/Latest?Limit=%d&Fields=Overview,Genres,DateCreated,PremiereDate,CommunityRating,ImageTags&IncludeItemTypes=Movie,Series&GroupItems=true", c.ServerURL, userId, limit)
	if parentId != "" {
		url += "&ParentId=" + parentId
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Emby服务器返回错误: %d - %s", resp.StatusCode, string(body))
	}

	var items []MediaItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return items, nil
}

// GetImageURL 获取媒体图片URL
func (c *Client) GetImageURL(itemId string, imageType string, tag string) string {
	if tag == "" {
		return fmt.Sprintf("%s/Items/%s/Images/%s", c.ServerURL, itemId, imageType)
	}
	return fmt.Sprintf("%s/Items/%s/Images/%s?tag=%s", c.ServerURL, itemId, imageType, tag)
}

// SetUserPolicy 设置Emby用户权限策略（受限普通用户）
func (c *Client) SetUserPolicy(userID string) error {
	url := fmt.Sprintf("%s/Users/%s/Policy", c.ServerURL, userID)

	// 受限普通用户权限配置（与test用户相同）
	policy := map[string]interface{}{
		"IsAdministrator":                 false,
		"IsHidden":                        true, // 隐藏用户
		"IsHiddenRemotely":                true,
		"IsHiddenFromUnusedDevices":       true,
		"IsDisabled":                      false,
		"EnableUserPreferenceAccess":      true,
		"EnableRemoteControlOfOtherUsers": false,
		"EnableSharedDeviceControl":       true,
		"EnableRemoteAccess":              true, // 允许远程访问
		"EnableLiveTvManagement":          false,
		"EnableLiveTvAccess":              true, // 允许看直播
		"EnableMediaPlayback":             true, // 允许播放
		"EnableAudioPlaybackTranscoding":  true, // 允许音频转码
		"EnableVideoPlaybackTranscoding":  true, // 允许视频转码
		"EnablePlaybackRemuxing":          true,
		"EnableContentDeletion":           false, // 禁止删除
		"EnableContentDownloading":        false, // 禁止下载
		"EnableSubtitleDownloading":       false, // 禁止下载字幕
		"EnableSubtitleManagement":        false,
		"EnableSyncTranscoding":           false,
		"EnableMediaConversion":           false,
		"EnableAllChannels":               true,
		"EnableAllFolders":                true, // 访问所有媒体库
		"EnableAllDevices":                true,
		"EnablePublicSharing":             false, // 禁止公开分享
		"AllowCameraUpload":               false,
		"AllowSharingPersonalItems":       false,
		"SimultaneousStreamLimit":         0, // 无并发限制
		"RemoteClientBitrateLimit":        0, // 无码率限制
	}

	bodyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Token", c.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求Emby服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("设置用户权限失败: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
