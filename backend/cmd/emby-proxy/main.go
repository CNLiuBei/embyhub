// Emby代理服务 - 独立端口运行
// 用于客户端无法添加路径的情况
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"feiniu-user-system/internal/config"
	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/service"

	"gorm.io/gorm"
)

// Token 到 Emby 用户 ID 的缓存
var (
	tokenUserCache = make(map[string]string) // token -> emby_user_id
	tokenCacheMu   sync.RWMutex
)

// 全局速率限制器（按用户ID分组，每个用户独立限速）
var (
	userRateLimiters = make(map[string]*sharedRateLimiter) // emby_user_id -> limiter
	rateLimiterMu    sync.RWMutex
)

// sharedRateLimiter 共享速率限制器（令牌桶算法）
type sharedRateLimiter struct {
	mu           sync.Mutex
	tokens       float64   // 当前可用令牌数
	maxTokens    float64   // 最大令牌数（桶容量）
	refillRate   float64   // 每秒补充令牌数（字节/秒）
	lastRefill   time.Time // 上次补充时间
}

// newSharedRateLimiter 创建共享速率限制器
func newSharedRateLimiter(bytesPerSecond int64) *sharedRateLimiter {
	// 桶容量设为1秒的流量，允许短暂突发
	maxTokens := float64(bytesPerSecond)
	return &sharedRateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: float64(bytesPerSecond),
		lastRefill: time.Now(),
	}
}

// updateRate 更新速率限制
func (l *sharedRateLimiter) updateRate(bytesPerSecond int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refillRate = float64(bytesPerSecond)
	l.maxTokens = float64(bytesPerSecond)
}

// consume 消费令牌，返回实际可消费的字节数和需要等待的时间
func (l *sharedRateLimiter) consume(requested int64) (int64, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(l.lastRefill).Seconds()
	l.tokens += elapsed * l.refillRate
	if l.tokens > l.maxTokens {
		l.tokens = l.maxTokens
	}
	l.lastRefill = now
	
	// 如果没有足够的令牌，计算需要等待的时间
	if l.tokens < 1 {
		waitTime := time.Duration((1 - l.tokens) / l.refillRate * float64(time.Second))
		return 0, waitTime
	}
	
	// 消费令牌
	canConsume := int64(l.tokens)
	if canConsume > requested {
		canConsume = requested
	}
	l.tokens -= float64(canConsume)
	
	return canConsume, 0
}

// getOrCreateUserRateLimiter 获取或创建用户对应的速率限制器
func getOrCreateUserRateLimiter(userID string, bytesPerSecond int64) *sharedRateLimiter {
	if userID == "" {
		return nil
	}
	
	rateLimiterMu.Lock()
	defer rateLimiterMu.Unlock()
	
	limiter, exists := userRateLimiters[userID]
	if !exists {
		limiter = newSharedRateLimiter(bytesPerSecond)
		userRateLimiters[userID] = limiter
		log.Printf("创建用户速率限制器: userID=%s, limit=%d B/s", userID, bytesPerSecond)
	} else {
		// 更新速率（配置可能已更改）
		limiter.updateRate(bytesPerSecond)
	}
	return limiter
}

// 错误页面模板
const errorPageTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            color: #fff;
            padding: 20px;
        }
        .container {
            text-align: center;
            max-width: 500px;
            animation: fadeIn 0.5s ease-out;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .icon {
            width: 120px;
            height: 120px;
            margin: 0 auto 30px;
            background: linear-gradient(135deg, #e94560 0%, #ff6b6b 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 10px 40px rgba(233, 69, 96, 0.3);
        }
        .icon svg {
            width: 60px;
            height: 60px;
            fill: #fff;
        }
        h1 {
            font-size: 28px;
            font-weight: 600;
            margin-bottom: 15px;
            color: #fff;
        }
        .message {
            font-size: 16px;
            color: rgba(255, 255, 255, 0.7);
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .error-code {
            display: inline-block;
            background: rgba(255, 255, 255, 0.1);
            padding: 8px 20px;
            border-radius: 20px;
            font-size: 14px;
            color: rgba(255, 255, 255, 0.5);
            margin-bottom: 30px;
        }
        .clients {
            background: rgba(255, 255, 255, 0.05);
            border-radius: 16px;
            padding: 25px;
            margin-top: 20px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .clients h3 {
            font-size: 14px;
            color: rgba(255, 255, 255, 0.5);
            margin-bottom: 15px;
            text-transform: uppercase;
            letter-spacing: 1px;
        }
        .client-list {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            justify-content: center;
        }
        .client-tag {
            background: linear-gradient(135deg, #00d9ff 0%, #00b4d8 100%);
            color: #fff;
            padding: 8px 16px;
            border-radius: 20px;
            font-size: 14px;
            font-weight: 500;
        }
        .footer {
            margin-top: 40px;
            font-size: 12px;
            color: rgba(255, 255, 255, 0.3);
        }
        .footer a {
            color: #00d9ff;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                {{if eq .Type "unknown"}}
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
                {{else if eq .Type "forbidden"}}
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 15h2v2h-2v-2zm0-12h2v8h-2V5z"/>
                {{else}}
                <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                {{end}}
            </svg>
        </div>
        <h1>{{.Title}}</h1>
        <p class="message">{{.Message}}</p>
        <div class="error-code">错误代码: {{.Code}}</div>
        {{if .Clients}}
        <div class="clients">
            <h3>支持的客户端</h3>
            <div class="client-list">
                {{range .Clients}}
                <span class="client-tag">{{.}}</span>
                {{end}}
            </div>
        </div>
        {{end}}
        <div class="footer">
            <p>如需帮助，请联系管理员</p>
        </div>
    </div>
</body>
</html>`

// ErrorPageData 错误页面数据
type ErrorPageData struct {
	Title   string
	Message string
	Code    int
	Type    string   // unknown, forbidden, error
	Clients []string // 支持的客户端列表
}

var (
	port       = flag.Int("port", 54682, "代理服务端口")
	configPath = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	db := database.GetDB()
	settingService := service.NewSettingService(db)

	// 创建代理处理器
	handler := &EmbyProxyHandler{
		db:             db,
		settingService: settingService,
	}

	// 启动服务
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Emby代理服务启动在 %s", addr)
	log.Printf("客户端可以使用 http://你的IP%s 连接", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// EmbyProxyHandler Emby代理处理器
type EmbyProxyHandler struct {
	db             *gorm.DB
	settingService *service.SettingService
}

// renderErrorPage 渲染错误页面
func renderErrorPage(w http.ResponseWriter, data ErrorPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(data.Code)
	
	tmpl, err := template.New("error").Parse(errorPageTemplate)
	if err != nil {
		http.Error(w, data.Message, data.Code)
		return
	}
	tmpl.Execute(w, data)
}

// getEnabledClientNames 获取启用的客户端名称列表
func getEnabledClientNames(settings *service.ClientWhitelistSettings) []string {
	var names []string
	for _, client := range settings.Clients {
		if client.Enabled {
			names = append(names, client.DisplayName)
		}
	}
	return names
}

// ServeHTTP 处理所有请求
func (h *EmbyProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 调试日志：记录所有请求路径
	log.Printf("代理请求: method=%s, path=%s", r.Method, r.URL.Path)

	// 获取Emby设置
	embySettings, err := h.settingService.GetEmbySettings()
	if err != nil || !embySettings.Enabled {
		renderErrorPage(w, ErrorPageData{
			Title:   "服务不可用",
			Message: "Emby 媒体服务当前未启用，请联系管理员。",
			Code:    http.StatusServiceUnavailable,
			Type:    "error",
		})
		return
	}

	// 获取客户端白名单设置
	clientSettings, err := h.settingService.GetClientWhitelistSettings()
	if err != nil {
		renderErrorPage(w, ErrorPageData{
			Title:   "服务器错误",
			Message: "获取客户端配置失败，请稍后重试。",
			Code:    http.StatusInternalServerError,
			Type:    "error",
		})
		return
	}

	// 检查客户端白名单
	if clientSettings.Enabled {
		clientName := r.Header.Get("X-Emby-Client")
		if clientName == "" {
			// 尝试从 Authorization 头解析
			authHeader := r.Header.Get("X-Emby-Authorization")
			if authHeader != "" {
				clientName = parseClientFromAuth(authHeader)
			}
		}

		enabledClients := getEnabledClientNames(clientSettings)

		// 如果白名单为空或没有启用的客户端，拒绝所有请求
		if len(clientSettings.Clients) == 0 || !hasEnabledClients(clientSettings) {
			log.Printf("客户端白名单为空，拒绝访问")
			renderErrorPage(w, ErrorPageData{
				Title:   "暂无可用客户端",
				Message: "管理员尚未配置允许的客户端，请联系管理员开通访问权限。",
				Code:    http.StatusForbidden,
				Type:    "forbidden",
			})
			return
		}

		// 如果无法识别客户端，拒绝访问
		if clientName == "" {
			log.Printf("无法识别客户端，拒绝访问")
			renderErrorPage(w, ErrorPageData{
				Title:   "无法识别客户端",
				Message: "请使用支持的播放器客户端访问媒体库。浏览器直接访问不被支持。",
				Code:    http.StatusForbidden,
				Type:    "unknown",
				Clients: enabledClients,
			})
			return
		}

		// 检查客户端是否在白名单中
		if !h.isClientAllowed(clientSettings, clientName) {
			log.Printf("客户端 %s 未授权访问", clientName)
			renderErrorPage(w, ErrorPageData{
				Title:   "客户端未授权",
				Message: fmt.Sprintf("您使用的客户端「%s」未获得访问授权。请使用以下支持的客户端。", clientName),
				Code:    http.StatusForbidden,
				Type:    "forbidden",
				Clients: enabledClients,
			})
			return
		}
	}

	// 检查播放限制
	if h.isPlaybackRequest(r.URL.Path) {
		authHeader := r.Header.Get("X-Emby-Authorization")
		playToken := parseTokenFromAuth(authHeader)
		playPathUserID := parseUserIDFromPath(r.URL.Path)
		
		// 优先使用路径中的用户ID，其次使用Token缓存
		userID := playPathUserID
		if userID == "" && playToken != "" {
			tokenCacheMu.RLock()
			userID = tokenUserCache[playToken]
			tokenCacheMu.RUnlock()
		}
		
		log.Printf("检测到播放请求: path=%s, pathUserID=%s, cachedUserID=%s, token=%s...", r.URL.Path, playPathUserID, userID, playToken[:min(8, len(playToken))])
		if userID != "" {
			blocked, msg := h.checkPlaybackLimit(userID)
			if blocked {
				log.Printf("播放请求被阻止: 用户ID=%s, 路径=%s, 原因=%s", userID, r.URL.Path, msg)
				renderErrorPage(w, ErrorPageData{
					Title:   "播放限制",
					Message: msg,
					Code:    http.StatusForbidden,
					Type:    "forbidden",
				})
				return
			}
		}
	}

	// 解析目标URL
	targetURL, err := url.Parse(embySettings.BaseURL)
	if err != nil {
		http.Error(w, "Emby服务器地址配置错误", http.StatusInternalServerError)
		return
	}

	// 缓存 Token 对应的用户 ID（从 /Users/{userId}/ 路径中提取）
	authHeader := r.Header.Get("X-Emby-Authorization")
	token := parseTokenFromAuth(authHeader)
	pathUserID := parseUserIDFromPath(r.URL.Path)
	if token != "" && pathUserID != "" {
		tokenCacheMu.Lock()
		tokenUserCache[token] = pathUserID
		tokenCacheMu.Unlock()
	}

	// 获取播放限制设置（用于速率限制）
	playLimitSettings, _ := h.settingService.GetPlayLimitSettings()

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// 自定义Director
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = targetURL.Host

		// 如果请求没有带API Key，添加服务器的API Key
		if req.URL.Query().Get("api_key") == "" && req.Header.Get("X-Emby-Token") == "" {
			q := req.URL.Query()
			q.Set("api_key", embySettings.APIKey)
			req.URL.RawQuery = q.Encode()
		}
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("代理错误: %v", err)
		http.Error(w, "无法连接到Emby服务器", http.StatusBadGateway)
	}

	// 检查是否需要速率限制
	isStreamRequest := h.isStreamRequest(r.URL.Path)
	var speedLimitBps int64 = 0
	var streamUserID string = "" // 用于闭包中获取用户限制器
	if isStreamRequest && playLimitSettings != nil && playLimitSettings.SpeedEnabled {
		// 获取用户ID - 优先从URL路径，其次从Token缓存，最后从Emby API
		// 注意：不使用 authHeader 中的 UserId，因为某些客户端（如SenPlayer）发送的是错误的值
		userID := pathUserID
		
		// 尝试获取 Token（从多个来源）
		streamToken := token
		if streamToken == "" {
			// 尝试从 X-Emby-Token 头获取
			streamToken = r.Header.Get("X-Emby-Token")
		}
		if streamToken == "" {
			// 尝试从 URL 参数获取
			streamToken = r.URL.Query().Get("api_key")
		}
		
		if userID == "" && streamToken != "" {
			tokenCacheMu.RLock()
			userID = tokenUserCache[streamToken]
			tokenCacheMu.RUnlock()
			
			// 如果缓存中没有，通过 Emby API 获取
			if userID == "" {
				userID = h.getUserIDByToken(streamToken, embySettings.BaseURL)
				if userID != "" {
					tokenCacheMu.Lock()
					tokenUserCache[streamToken] = userID
					tokenCacheMu.Unlock()
					log.Printf("通过Emby API获取用户ID并缓存: token=%s..., userId=%s", streamToken[:min(8, len(streamToken))], userID)
				}
			}
		}
		userRole := h.getUserRoleByEmbyID(userID)
		streamUserID = userID // 保存到外部变量供闭包使用
		tokenPrefix := streamToken
		if len(tokenPrefix) > 8 {
			tokenPrefix = tokenPrefix[:8]
		}
		cachedID := ""
		if streamToken != "" {
			tokenCacheMu.RLock()
			cachedID = tokenUserCache[streamToken]
			tokenCacheMu.RUnlock()
		}
		log.Printf("速率限制用户查询: pathUserID=%s, cachedUserID=%s, token=%s..., finalUserID=%s, role=%d", pathUserID, cachedID, tokenPrefix, userID, userRole)
		
		// 根据角色获取速率限制（单位：MB/s）
		var speedLimitMBps int
		switch {
		case userRole >= 2: // 管理员
			speedLimitMBps = playLimitSettings.SpeedAdmin
		case userRole == 1: // 会员
			speedLimitMBps = playLimitSettings.SpeedMember
		default: // 普通用户
			speedLimitMBps = playLimitSettings.SpeedUser
		}
		
		if speedLimitMBps > 0 {
			speedLimitBps = int64(speedLimitMBps) * 1024 * 1024 // MB/s 转 Bytes/s
			log.Printf("应用速率限制: 用户角色=%d, 限制=%d MB/s (%d B/s)", userRole, speedLimitMBps, speedLimitBps)
		}
	}

	// 检查是否是登录请求，需要拦截响应获取用户信息
	isLoginRequest := strings.Contains(strings.ToLower(r.URL.Path), "/users/authenticatebyname")
	
	// 设置响应修改器
	proxy.ModifyResponse = func(resp *http.Response) error {
		// 拦截登录响应，缓存 Token 和 UserId
		if isLoginRequest && resp.StatusCode == http.StatusOK {
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil {
				// 解析登录响应
				var loginResp struct {
					AccessToken string `json:"AccessToken"`
					User        struct {
						Id   string `json:"Id"`
						Name string `json:"Name"`
					} `json:"User"`
				}
				if json.Unmarshal(bodyBytes, &loginResp) == nil && loginResp.AccessToken != "" && loginResp.User.Id != "" {
					tokenCacheMu.Lock()
					tokenUserCache[loginResp.AccessToken] = loginResp.User.Id
					tokenCacheMu.Unlock()
					log.Printf("登录成功，缓存用户信息: username=%s, userId=%s, token=%s...", 
						loginResp.User.Name, loginResp.User.Id, loginResp.AccessToken[:8])
				}
				// 重新设置响应体
				resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}
		
		// 对视频流响应应用速率限制（按用户独立限速）
		if speedLimitBps > 0 && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusPartialContent) {
			limiter := getOrCreateUserRateLimiter(streamUserID, speedLimitBps)
			if limiter != nil {
				resp.Body = &rateLimitedReader{
					reader:  resp.Body,
					limiter: limiter,
				}
			}
		}
		return nil
	}

	// 执行代理
	proxy.ServeHTTP(w, r)
}

// hasEnabledClients 检查是否有启用的客户端
func hasEnabledClients(settings *service.ClientWhitelistSettings) bool {
	for _, client := range settings.Clients {
		if client.Enabled {
			return true
		}
	}
	return false
}

// isClientAllowed 检查客户端是否允许
func (h *EmbyProxyHandler) isClientAllowed(settings *service.ClientWhitelistSettings, clientName string) bool {
	for _, client := range settings.Clients {
		if !client.Enabled {
			continue
		}
		// 精确匹配
		if client.Name == clientName {
			return true
		}
		// 前缀匹配（处理版本号等情况）
		if strings.HasPrefix(clientName, client.Name) {
			return true
		}
	}
	return false
}

// parseClientFromAuth 从Authorization头解析客户端名称
func parseClientFromAuth(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "Client=") || strings.HasPrefix(part, "MediaBrowser Client=") {
			idx := strings.Index(part, "Client=")
			if idx >= 0 {
				value := part[idx+7:]
				value = strings.Trim(value, "\"")
				return value
			}
		}
	}
	return ""
}


// getUserIDByToken 通过 Token 从 Emby API 获取用户 ID
func (h *EmbyProxyHandler) getUserIDByToken(token, baseURL string) string {
	if token == "" || baseURL == "" {
		return ""
	}
	
	// 调用 Emby /Users/Me 接口获取当前用户信息
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/Users/Me", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("X-Emby-Token", token)
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("通过Token获取用户失败: %v", err)
		return ""
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	
	// 简单解析 JSON 获取 Id 字段
	var result struct {
		Id string `json:"Id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ""
	}
	
	log.Printf("通过Token获取用户ID成功: %s", result.Id)
	return result.Id
}

// parseTokenFromAuth 从Authorization头解析Token
func parseTokenFromAuth(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "Token=") {
			idx := strings.Index(part, "Token=")
			if idx >= 0 {
				value := part[idx+6:]
				return strings.Trim(value, "\"")
			}
		}
	}
	return ""
}

// parseUserIDFromPath 从URL路径解析用户ID
// 例如: /emby/Users/2064bb799a1047e191472eabbbae316b/Items/6
func parseUserIDFromPath(path string) string {
	// 匹配 /Users/{userId}/ 或 /emby/Users/{userId}/
	lowerPath := strings.ToLower(path)
	patterns := []string{"/users/", "/emby/users/"}
	for _, pattern := range patterns {
		idx := strings.Index(lowerPath, pattern)
		if idx >= 0 {
			start := idx + len(pattern)
			rest := path[start:]
			// 找到下一个 / 或字符串结束
			endIdx := strings.Index(rest, "/")
			if endIdx > 0 {
				return rest[:endIdx]
			} else if len(rest) > 0 {
				return rest
			}
		}
	}
	return ""
}

// parseUserIDFromAuth 从Authorization头解析用户ID
func parseUserIDFromAuth(authHeader string) string {
	parts := strings.Split(authHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "UserId=") {
			idx := strings.Index(part, "UserId=")
			if idx >= 0 {
				value := part[idx+7:]
				return strings.Trim(value, "\"")
			}
		}
	}
	return ""
}

// isPlaybackRequest 检查是否是播放开始请求（用于播放数量限制）
// 只检查真正的播放开始请求，不检查视频流请求
func (h *EmbyProxyHandler) isPlaybackRequest(path string) bool {
	lowerPath := strings.ToLower(path)
	// 只匹配播放开始请求和播放信息请求
	// 不匹配视频流请求和进度更新请求
	playbackPatterns := []string{
		`^/items/[^/]+/playbackinfo$`,
		`^/sessions/playing$`,
		`^/emby/items/[^/]+/playbackinfo$`,
		`^/emby/sessions/playing$`,
	}
	for _, pattern := range playbackPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerPath); matched {
			return true
		}
	}
	return false
}


// checkPlaybackLimit 检查用户播放限制
func (h *EmbyProxyHandler) checkPlaybackLimit(userID string) (bool, string) {
	limitSettings, err := h.settingService.GetPlayLimitSettings()
	if err != nil {
		log.Printf("获取播放限制设置失败: %v", err)
		return false, ""
	}
	if !limitSettings.Enabled || limitSettings.MaxPlaying <= 0 {
		return false, ""
	}

	adapter := service.GetMediaAdapterFromDB(h.db)
	if adapter == nil {
		log.Printf("无法获取媒体适配器")
		return false, ""
	}

	sessions, err := adapter.GetSessionsByUserID(userID)
	if err != nil {
		log.Printf("获取用户会话失败: %v", err)
		return false, ""
	}

	playingCount := 0
	for _, session := range sessions {
		if session.IsPlaying {
			playingCount++
		}
	}
	log.Printf("用户 %s 当前播放数: %d, 限制: %d", userID, playingCount, limitSettings.MaxPlaying)

	if playingCount >= limitSettings.MaxPlaying {
		return true, fmt.Sprintf("已达到最大同时播放数量限制（%d），请先停止其他设备的播放", limitSettings.MaxPlaying)
	}
	return false, ""
}


// rateLimitedReader 速率限制的读取器（使用共享限制器）
type rateLimitedReader struct {
	reader  io.ReadCloser
	limiter *sharedRateLimiter
}

// Read 实现 io.Reader 接口，带速率限制
func (r *rateLimitedReader) Read(p []byte) (n int, err error) {
	if r.limiter == nil {
		return r.reader.Read(p)
	}
	
	// 请求消费令牌
	requested := int64(len(p))
	for {
		canConsume, waitTime := r.limiter.consume(requested)
		if canConsume > 0 {
			// 限制读取大小
			if canConsume < int64(len(p)) {
				p = p[:canConsume]
			}
			return r.reader.Read(p)
		}
		// 等待令牌补充
		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}
}

// Close 实现 io.Closer 接口
func (r *rateLimitedReader) Close() error {
	return r.reader.Close()
}

// isStreamRequest 检查是否是视频/音频流请求
func (h *EmbyProxyHandler) isStreamRequest(path string) bool {
	lowerPath := strings.ToLower(path)
	streamPatterns := []string{
		`^/videos/[^/]+/stream`,
		`^/audio/[^/]+/stream`,
		`^/emby/videos/[^/]+/stream`,
		`^/emby/audio/[^/]+/stream`,
		`^/videos/[^/]+/original`,
		`^/emby/videos/[^/]+/original`,
		`\.mkv$`,
		`\.mp4$`,
		`\.avi$`,
		`\.ts$`,
		`\.m3u8$`,
	}
	for _, pattern := range streamPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerPath); matched {
			return true
		}
	}
	return false
}


// getUserRoleByEmbyID 根据 Emby 用户 ID 获取本地用户角色
func (h *EmbyProxyHandler) getUserRoleByEmbyID(embyUserID string) int8 {
	if embyUserID == "" {
		return 0
	}
	
	// 直接从数据库查询（emby_user_id 已保存在 users 表中）
	// 支持带连字符和不带连字符的 ID 格式
	normalizedID := strings.ReplaceAll(strings.ToLower(embyUserID), "-", "")
	
	var user struct {
		Username string
		Role     int8
	}
	err := h.db.Table("users").
		Select("username, role").
		Where("LOWER(REPLACE(emby_user_id, '-', '')) = ?", normalizedID).
		First(&user).Error
	
	if err != nil {
		log.Printf("根据Emby用户ID查询角色失败: embyUserID=%s, normalizedID=%s, err=%v", embyUserID, normalizedID, err)
		return 0
	}
	
	log.Printf("用户角色查询成功: embyUserID=%s, username=%s, role=%d", embyUserID, user.Username, user.Role)
	return user.Role
}
