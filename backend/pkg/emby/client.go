// Package emby Emby API客户端
package emby

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	secretKey = "NDzZTVxnRKP8Z0jXg1VAMonaG8akvh"
	apiKey    = "16CCEB3D-AB42-077D-36A1-F355324E4237"
)

// Client Emby API客户端
type Client struct {
	baseURL   string
	adminUser string
	adminPass string
	token     string
	client    *http.Client
}

// Config 客户端配置
type Config struct {
	BaseURL   string // API基础地址
	AdminUser string // 管理员账号
	AdminPass string // 管理员密码
}

// Response 统一响应格式
type Response struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username        string   `json:"username"`
	Password        string   `json:"password"`
	IsAdmin         int      `json:"is_admin"`
	MediaPermission int      `json:"media_permission"`
	MediaDBList     []string `json:"mediadb_list"`
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	GUID string `json:"guid"`
}

// NewClient 创建客户端
func NewClient(cfg *Config) *Client {
	return &Client{
		baseURL:   cfg.BaseURL,
		adminUser: cfg.AdminUser,
		adminPass: cfg.AdminPass,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// generateAuthx 生成authx签名
func generateAuthx(path, method string, params map[string]string, data interface{}) string {
	// 确保路径包含 /v/api/v1
	if !strings.HasPrefix(path, "/v/api/v1") {
		path = "/v/api/v1" + path
	}

	// 计算数据哈希
	var content string
	if method == "GET" && len(params) > 0 {
		// GET请求：参数排序后拼接
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(params[k])))
		}
		content = strings.Join(parts, "&")
	} else if (method == "POST" || method == "PUT") && data != nil {
		// POST/PUT请求：JSON序列化
		jsonData, _ := json.Marshal(data)
		content = string(jsonData)
	} else {
		content = ""
	}

	// 计算数据MD5
	dataHash := md5Hash(content)

	// 生成随机nonce和时间戳
	nonce := fmt.Sprintf("%06d", rand.Intn(900000)+100000)
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)

	// 拼接签名字符串
	signString := fmt.Sprintf("%s_%s_%s_%s_%s_%s",
		secretKey, path, nonce, timestamp, dataHash, apiKey)

	// 计算签名
	sign := md5Hash(signString)

	return fmt.Sprintf("nonce=%s&timestamp=%s&sign=%s", nonce, timestamp, sign)
}

// md5Hash 计算MD5哈希
func md5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GetToken 获取当前token
func (c *Client) GetToken() string {
	return c.token
}

// Login 管理员登录获取token
func (c *Client) Login() error {
	payload := map[string]string{
		"username": c.adminUser,
		"password": c.adminPass,
		"app_name": "trimemedia-web",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(c.baseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("登录Emby失败: %v", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("登录失败: %s", result.Msg)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(result.Data, &loginResp); err != nil {
		return err
	}

	c.token = loginResp.Token
	return nil
}

// CreateUser 创建用户
func (c *Client) CreateUser(username, password string) (*CreateUserResponse, error) {
	// 确保已登录
	if c.token == "" {
		if err := c.Login(); err != nil {
			return nil, fmt.Errorf("飞牛影视登录失败: %w", err)
		}
	}

	req := &CreateUserRequest{
		Username:        username,
		Password:        password,
		IsAdmin:         0,
		MediaPermission: 2, // 全部权限
		MediaDBList:     []string{},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("PUT", c.baseURL+"/manager/user/create", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求飞牛影视API失败: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w (响应: %s)", err, string(data))
	}

	// token过期重试
	if result.Code != 0 && result.Msg == "token invalid" {
		c.token = ""
		return c.CreateUser(username, password)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("飞牛影视API错误 (code=%d): %s", result.Code, result.Msg)
	}

	var createResp CreateUserResponse
	if err := json.Unmarshal(result.Data, &createResp); err != nil {
		return nil, fmt.Errorf("解析用户数据失败: %w", err)
	}

	return &createResp, nil
}

// UserInfo 用户信息
type UserInfo struct {
	GUID     string `json:"guid"`
	Username string `json:"username"`
	IsAdmin  int    `json:"is_admin"`
}

// MediaDB 媒体库信息
type MediaDB struct {
	GUID     string   `json:"guid"`
	Title    string   `json:"title"`
	Posters  []string `json:"posters"`
	Category string   `json:"category"`
	ViewType int      `json:"view_type"`
}

// MediaDBSum 媒体库统计
type MediaDBSum struct {
	Total    int `json:"total"`
	Movie    int `json:"movie"`
	TV       int `json:"tv"`
	Video    int `json:"video"`
	Favorite int `json:"favorite"`
}

// GetUserList 获取用户列表
func (c *Client) GetUserList() ([]UserInfo, error) {
	if c.token == "" {
		if err := c.Login(); err != nil {
			return nil, err
		}
	}

	httpReq, err := http.NewRequest("GET", c.baseURL+"/manager/user/list", nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("authorization", c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var users []UserInfo
	if err := json.Unmarshal(result.Data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteUser 删除用户
func (c *Client) DeleteUser(guid string) error {
	if c.token == "" {
		if err := c.Login(); err != nil {
			return err
		}
	}

	httpReq, err := http.NewRequest("DELETE", c.baseURL+"/manager/user/"+guid, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("authorization", c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteUserByUsername 根据用户名删除用户
func (c *Client) DeleteUserByUsername(username string) error {
	users, err := c.GetUserList()
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.Username == username {
			return c.DeleteUser(u.GUID)
		}
	}

	return nil // 用户不存在，无需删除
}

// UpdateUserStatus 更新用户状态 (通过media_permission控制)
func (c *Client) UpdateUserStatus(username string, enabled bool) error {
	users, err := c.GetUserList()
	if err != nil {
		return err
	}

	var guid string
	for _, u := range users {
		if u.Username == username {
			guid = u.GUID
			break
		}
	}

	if guid == "" {
		return nil // 用户不存在
	}

	if c.token == "" {
		if err := c.Login(); err != nil {
			return err
		}
	}

	// enabled=true: media_permission=2 (全部权限)
	// enabled=false: media_permission=0 (无权限，相当于禁用)
	permission := 0
	if enabled {
		permission = 2
	}

	req := map[string]interface{}{
		"username":         username,
		"is_admin":         0,
		"media_permission": permission,
		"mediadb_list":     []string{},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/manager/user/"+guid, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// UpdateUserPassword 更新用户密码
func (c *Client) UpdateUserPassword(username, password string) error {
	users, err := c.GetUserList()
	if err != nil {
		return err
	}

	var guid string
	for _, u := range users {
		if u.Username == username {
			guid = u.GUID
			break
		}
	}

	if guid == "" {
		return nil // 用户不存在
	}

	if c.token == "" {
		if err := c.Login(); err != nil {
			return err
		}
	}

	req := &CreateUserRequest{
		Username:        username,
		Password:        password,
		IsAdmin:         0,
		MediaPermission: 2,
		MediaDBList:     []string{},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/manager/user/"+guid, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", c.token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetMediaDBList 获取媒体库列表 (需要用户token)
func (c *Client) GetMediaDBList(userToken string) ([]MediaDB, error) {
	path := "/v/api/v1/mediadb/list"
	requestURL := c.baseURL + "/mediadb/list" // baseURL已包含/v/api/v1
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", nil, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var mediaDBs []MediaDB
	if err := json.Unmarshal(result.Data, &mediaDBs); err != nil {
		return nil, err
	}

	return mediaDBs, nil
}

// LoginUser 用户登录飞牛影视获取token
func (c *Client) LoginUser(username, password string) (string, error) {
	payload := map[string]string{
		"username": username,
		"password": password,
		"app_name": "trimemedia-web",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Post(c.baseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", fmt.Errorf("登录失败: %s", result.Msg)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(result.Data, &loginResp); err != nil {
		return "", err
	}

	return loginResp.Token, nil
}

// MediaItem 媒体项
type MediaItem struct {
	GUID                  string `json:"guid"`
	Title                 string `json:"title"`
	OriginalTitle         string `json:"original_title"`
	Type                  string `json:"type"` // Movie/TV
	Poster                string `json:"poster"`
	VoteAverage           string `json:"vote_average"`
	ReleaseDate           string `json:"release_date"`
	Year                  int    `json:"year"`
	Overview              string `json:"overview"`
	NumberOfEpisodes      int    `json:"number_of_episodes"`
	LocalNumberOfEpisodes int    `json:"local_number_of_episodes"`
	IsFavorite            int    `json:"is_favorite"`
	Watched               int    `json:"watched"`
	Duration              int    `json:"duration"`
}

// MediaDetail 媒体详情
type MediaDetail struct {
	GUID          string      `json:"guid"`
	Title         string      `json:"title"`
	OriginalTitle string      `json:"original_title"`
	Year          int         `json:"year"`
	Rating        float64     `json:"rating"`
	Posters       []string    `json:"posters"`
	Backdrops     []string    `json:"backdrops"`
	Genres        []string    `json:"genres"`
	Overview      string      `json:"overview"`
	Duration      int         `json:"duration"`
	Category      string      `json:"category"`
	Seasons       []Season    `json:"seasons,omitempty"`
	Files         []MediaFile `json:"files,omitempty"`
}

// Season 季信息（从飞牛API返回）
type Season struct {
	GUID              string `json:"guid"`
	Title             string `json:"title"`
	SeasonNumber      int    `json:"season_number"`
	EpisodeCount      int    `json:"episode_count"`
	LocalEpisodeCount int    `json:"local_episode_count"`
	Poster            string `json:"poster"`
}

// Episode 集信息（从飞牛API返回）
type Episode struct {
	GUID          string `json:"guid"`
	Title         string `json:"title"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonNumber  int    `json:"season_number"`
	Overview      string `json:"overview"`
	StillPath     string `json:"still_path"`
	Runtime       int    `json:"runtime"`
	IsWatched     int    `json:"is_watched"`
	PlayPosition  int    `json:"play_position"`
}

// MediaFile 媒体文件
type MediaFile struct {
	GUID     string `json:"guid"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Duration int    `json:"duration"`
	Codec    string `json:"codec"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// GetMediaDBItems 获取媒体库中的媒体列表
func (c *Client) GetMediaDBItems(userToken, mediaDBGUID string, page, pageSize int) ([]MediaItem, int, error) {
	path := "/v/api/v1/item/list"
	requestURL := c.baseURL + "/item/list"

	// 构造POST请求体
	requestBody := map[string]interface{}{
		"ancestor_guid":         mediaDBGUID,
		"page":                  page,
		"page_size":             pageSize,
		"sort_column":           "create_time",
		"sort_type":             "DESC",
		"exclude_grouped_video": 1, // 排除分组视频（剧集）
		"tags": map[string]interface{}{
			"type": []string{"Movie", "TV"}, // 只显示电影和电视剧
		},
	}

	bodyData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, 0, err
	}

	httpReq, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(bodyData))
	if err != nil {
		return nil, 0, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "POST", nil, requestBody))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code != 0 {
		return nil, 0, fmt.Errorf("飞牛API错误: %s", result.Msg)
	}

	var itemsResp struct {
		List  []MediaItem `json:"list"` // 注意：飞牛API用的是list而不是items
		Total int         `json:"total"`
	}
	if err := json.Unmarshal(result.Data, &itemsResp); err != nil {
		return nil, 0, fmt.Errorf("解析items失败: %v", err)
	}

	return itemsResp.List, itemsResp.Total, nil
}

// GetMediaDetail 获取媒体详情
func (c *Client) GetMediaDetail(userToken, mediaGUID string) (*MediaDetail, error) {
	path := fmt.Sprintf("/v/api/v1/media/%s", mediaGUID)
	requestURL := fmt.Sprintf("%s/media/%s", c.baseURL, mediaGUID)
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", nil, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var detail MediaDetail
	if err := json.Unmarshal(result.Data, &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// SearchMedia 搜索媒体
func (c *Client) SearchMedia(userToken, keyword string, page, pageSize int) ([]MediaItem, int, error) {
	path := "/v/api/v1/media/search"
	requestURL := fmt.Sprintf("%s/media/search?keyword=%s&page=%d&page_size=%d", c.baseURL, url.QueryEscape(keyword), page, pageSize)
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, 0, err
	}

	params := map[string]string{
		"keyword":   keyword,
		"page":      fmt.Sprintf("%d", page),
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", params, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, 0, err
	}

	if result.Code != 0 {
		return nil, 0, errors.New(result.Msg)
	}

	var searchResp struct {
		Items []MediaItem `json:"items"`
		Total int         `json:"total"`
	}
	if err := json.Unmarshal(result.Data, &searchResp); err != nil {
		return nil, 0, err
	}

	return searchResp.Items, searchResp.Total, nil
}

// GetMediaDBSum 获取媒体库统计
func (c *Client) GetMediaDBSum(userToken, mediaDBGUID string) (*MediaDBSum, error) {
	path := fmt.Sprintf("/v/api/v1/mediadb/%s/sum", mediaDBGUID)
	requestURL := fmt.Sprintf("%s/mediadb/%s/sum", c.baseURL, mediaDBGUID)
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", nil, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var sum MediaDBSum
	if err := json.Unmarshal(result.Data, &sum); err != nil {
		return nil, err
	}

	return &sum, nil
}

// GetMediaSeasons 获取媒体的季列表
func (c *Client) GetMediaSeasons(userToken, mediaGUID string) ([]Season, error) {
	path := fmt.Sprintf("/v/api/v1/season/list/%s", mediaGUID)
	requestURL := fmt.Sprintf("%s/season/list/%s", c.baseURL, mediaGUID)
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", nil, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var seasons []Season
	if err := json.Unmarshal(result.Data, &seasons); err != nil {
		return nil, err
	}

	return seasons, nil
}

// GetSeasonEpisodes 获取季的剧集列表
func (c *Client) GetSeasonEpisodes(userToken, seasonGUID string) ([]Episode, error) {
	path := fmt.Sprintf("/v/api/v1/episode/list/%s", seasonGUID)
	requestURL := fmt.Sprintf("%s/episode/list/%s", c.baseURL, seasonGUID)
	httpReq, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("authorization", userToken)
	httpReq.Header.Set("authx", generateAuthx(path, "GET", nil, nil))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, errors.New(result.Msg)
	}

	var episodes []Episode
	if err := json.Unmarshal(result.Data, &episodes); err != nil {
		return nil, err
	}

	return episodes, nil
}
