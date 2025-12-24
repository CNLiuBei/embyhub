// Package middleware 中间件
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"feiniu-user-system/internal/database"
	"feiniu-user-system/internal/service"
	"feiniu-user-system/pkg/auth"
	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ContextUserID   = "user_id"
	ContextUserRole = "user_role"
)

// JWTAuth JWT认证中间件
func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Token格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 检查Token是否在黑名单
		ctx := context.Background()
		blacklistKey := database.KeyBlacklist + tokenString
		if exists, _ := database.RDB.Exists(ctx, blacklistKey).Result(); exists > 0 {
			response.Unauthorized(c, "Token已失效")
			c.Abort()
			return
		}

		// 解析Token
		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			if err == auth.ErrTokenExpired {
				response.Unauthorized(c, "Token已过期")
			} else {
				response.Unauthorized(c, "Token无效")
			}
			c.Abort()
			return
		}

		// 设置用户信息到上下文
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUserRole, claims.Role)
		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...int8) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextUserRole)
		if !exists {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		userRole := role.(int8)
		for _, r := range roles {
			if userRole >= r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "没有操作权限")
		c.Abort()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	id, exists := c.Get(ContextUserID)
	if !exists {
		return uuid.Nil, false
	}
	return id.(uuid.UUID), true
}

// GetUserRole 从上下文获取用户角色
func GetUserRole(c *gin.Context) int8 {
	role, exists := c.Get(ContextUserRole)
	if !exists {
		return 0
	}
	return role.(int8)
}

// RequireMember 会员权限检查中间件 - 检查用户是否为有效会员
// 只有有效会员（未过期）才能访问受保护的资源
func RequireMember(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 从数据库获取用户会员信息
		var user struct {
			MemberLevel  int8
			MemberExpire *time.Time
			Role         int8
		}
		if err := db.Table("users").Select("member_level, member_expire, role").Where("id = ?", userID).First(&user).Error; err != nil {
			response.Unauthorized(c, "用户不存在")
			c.Abort()
			return
		}

		// 管理员和超级管理员直接放行
		if user.Role >= 2 {
			c.Next()
			return
		}

		// 检查会员是否有效
		if user.MemberLevel == 0 {
			response.Forbidden(c, "您不是会员，请先开通会员后访问")
			c.Abort()
			return
		}

		// 检查会员是否过期
		if user.MemberExpire == nil || user.MemberExpire.Before(time.Now()) {
			response.Forbidden(c, "您的会员已过期，请续费后访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimit 限流中间件（基于用户ID+路径，未登录则基于IP+路径）
func RateLimit(limit int, window time.Duration, keyPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath() // 使用路由路径，如 /api/v1/payment/order/:order_no
		var key string
		
		// 优先使用用户ID作为限流key
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uuid.UUID); ok {
				key = keyPrefix + "user:" + uid.String() + ":" + path
			}
		}
		
		// 如果没有用户ID，则使用IP
		if key == "" {
			key = keyPrefix + "ip:" + c.ClientIP() + ":" + path
		}

		ctx := context.Background()
		count, err := database.IncrLimit(ctx, key, window)
		if err != nil {
			c.Next()
			return
		}

		if count > int64(limit) {
			response.Error(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit 登录限流中间件
// 1. 检查IP是否被拉黑
// 2. 检查账号失败次数（只有失败才计数，成功会清除）
// 3. 检测同一IP尝试多个不同账号失败的攻击行为
func LoginRateLimit(limit int, window time.Duration, ipBlacklistService *service.IPBlacklistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := context.Background()

		// 1. 先检查IP是否已被拉黑
		if ipBlacklistService != nil && ipBlacklistService.IsBlocked(ip) {
			response.Error(c, 403, "您的IP已被封禁，请联系管理员")
			c.Abort()
			return
		}

		// 先读取请求体获取账号
		var loginReq struct {
			Account string `json:"account"`
		}

		// 复制请求体以便后续handler可以再次读取
		bodyBytes, err := c.GetRawData()
		if err != nil {
			c.Next()
			return
		}

		// 解析账号
		if err := json.Unmarshal(bodyBytes, &loginReq); err != nil || loginReq.Account == "" {
			// 无法解析账号，直接放行
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			c.Next()
			return
		}

		// 恢复请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 2. 检查该账号的失败次数（只读取，不增加）
		key := database.KeyLoginLimit + "account:" + loginReq.Account
		countStr, _ := database.GetRedis().Get(ctx, key).Result()
		var count int64
		if countStr != "" {
			count, _ = database.GetRedis().Get(ctx, key).Int64()
		}

		if count >= int64(limit) {
			response.Error(c, 429, "登录失败次数过多，请15分钟后再试")
			c.Abort()
			return
		}

		// 3. 检查该IP的失败账号数量
		ipFailedAccountsKey := database.KeyLoginLimit + "ip_failed_accounts:" + ip
		failedAccountCount, _ := database.GetRedis().SCard(ctx, ipFailedAccountsKey).Result()

		if failedAccountCount >= 5 {
			// 自动拉黑该IP（首次24小时，再犯永久）
			var isPermanent bool
			if ipBlacklistService != nil {
				isPermanent, _ = ipBlacklistService.AutoBlock(ip, "系统检测：短时间内多个账号登录失败，疑似暴力破解攻击", 24*60)
			}
			// 清除记录
			database.GetRedis().Del(ctx, ipFailedAccountsKey)

			if isPermanent {
				response.Error(c, 403, "检测到重复攻击行为，您的IP已被永久封禁")
			} else {
				response.Error(c, 403, "检测到异常登录行为，您的IP已被临时封禁24小时")
			}
			c.Abort()
			return
		}

		// 将账号和IP信息存入context，供handler使用
		c.Set("login_account", loginReq.Account)
		c.Set("login_ip", ip)
		c.Set("ip_blacklist_service", ipBlacklistService)

		c.Next()
	}
}

// RecordLoginFailure 记录登录失败（在handler中调用）
func RecordLoginFailure(ctx context.Context, account, ip string, window time.Duration) {
	// 增加账号失败计数
	key := database.KeyLoginLimit + "account:" + account
	database.GetRedis().Incr(ctx, key)
	database.GetRedis().Expire(ctx, key, window)

	// 将该账号添加到IP的失败账号集合
	ipFailedAccountsKey := database.KeyLoginLimit + "ip_failed_accounts:" + ip
	database.GetRedis().SAdd(ctx, ipFailedAccountsKey, account)
	database.GetRedis().Expire(ctx, ipFailedAccountsKey, 10*time.Minute)
}

// ClearLoginFailure 清除登录失败记录（登录成功时调用）
func ClearLoginFailure(ctx context.Context, account, ip string) {
	// 清除账号失败计数
	key := database.KeyLoginLimit + "account:" + account
	database.GetRedis().Del(ctx, key)

	// 从IP的失败账号集合中移除该账号
	ipFailedAccountsKey := database.KeyLoginLimit + "ip_failed_accounts:" + ip
	database.GetRedis().SRem(ctx, ipFailedAccountsKey, account)
}

// CheckIPAttack 检查IP是否触发攻击检测（在登录失败后调用）
func CheckIPAttack(ctx context.Context, ip string, ipBlacklistService *service.IPBlacklistService) (blocked bool, permanent bool) {
	ipFailedAccountsKey := database.KeyLoginLimit + "ip_failed_accounts:" + ip
	failedAccountCount, _ := database.GetRedis().SCard(ctx, ipFailedAccountsKey).Result()

	if failedAccountCount >= 5 && ipBlacklistService != nil {
		permanent, _ = ipBlacklistService.AutoBlock(ip, "系统检测：短时间内多个账号登录失败，疑似暴力破解攻击", 24*60)
		database.GetRedis().Del(ctx, ipFailedAccountsKey)
		return true, permanent
	}
	return false, false
}

// DomainWhitelist 域名白名单中间件 - 完全由后端控制，无法绕过
func DomainWhitelist(db *gorm.DB) gin.HandlerFunc {
	settingService := service.NewSettingService(db)

	// 错误页面HTML模板
	const errorPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>域名未授权 - Domain Not Authorized</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 60px 40px;
            max-width: 600px;
            width: 100%;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            text-align: center;
            animation: slideIn 0.5s ease-out;
        }
        @keyframes slideIn {
            from { opacity: 0; transform: translateY(-30px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .icon {
            font-size: 80px;
            margin-bottom: 20px;
            animation: bounce 2s infinite;
        }
        @keyframes bounce {
            0%, 100% { transform: translateY(0); }
            50% { transform: translateY(-10px); }
        }
        h1 {
            color: #333;
            font-size: 32px;
            margin-bottom: 10px;
            font-weight: 700;
        }
        .subtitle {
            color: #666;
            font-size: 18px;
            margin-bottom: 30px;
            font-weight: 400;
        }
        .domain-box {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 12px;
            margin: 30px 0;
            font-family: 'Courier New', monospace;
            font-size: 18px;
            font-weight: bold;
            word-break: break-all;
            box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
        }
        .info {
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            text-align: left;
        }
        .info-title {
            color: #667eea;
            font-weight: 600;
            margin-bottom: 10px;
            font-size: 16px;
        }
        .info-text {
            color: #555;
            line-height: 1.6;
            font-size: 14px;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 2px solid #eee;
            color: #999;
            font-size: 12px;
        }
        .tech-info {
            display: inline-block;
            background: #f0f0f0;
            padding: 5px 10px;
            border-radius: 5px;
            margin: 5px;
            font-size: 11px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🚫</div>
        <h1>域名未授权</h1>
        <div class="subtitle">Domain Not Authorized</div>
        
        <div class="domain-box">
            {{.Domain}}
        </div>
        
        <div class="info">
            <div class="info-title">访问受限原因</div>
            <div class="info-text">
                此域名未在系统白名单中，无法访问本系统。<br>
                如需访问，请联系系统管理员将域名添加到授权列表。
            </div>
        </div>
        
        <div class="info">
            <div class="info-title">安全提示</div>
            <div class="info-text">
                本系统采用域名白名单机制，只允许授权域名访问。<br>
                这有助于防止未授权访问、域名劫持和钓鱼攻击。
            </div>
        </div>
        
        <div class="footer">
            <div class="tech-info">Error Code: 403</div>
            <div class="tech-info">Client IP: {{.ClientIP}}</div>
            <div class="tech-info">Request ID: {{.RequestID}}</div>
            <div style="margin-top: 10px; color: #bbb;">
                Protected by Go Backend Domain Whitelist Middleware
            </div>
        </div>
    </div>
</body>
</html>`

	return func(c *gin.Context) {
		// 获取请求的Host（支持反向代理）
		host := c.Request.Host

		// 尝试从X-Forwarded-Host获取原始域名
		if forwardedHost := c.GetHeader("X-Forwarded-Host"); forwardedHost != "" {
			host = forwardedHost
		} else if originalHostHeader := c.GetHeader("X-Original-Host"); originalHostHeader != "" {
			host = originalHostHeader
		}

		// 移除端口号
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// 检查域名是否在白名单中
		allowed := settingService.IsDomainAllowed(host)

		if !allowed {
			// 域名被拒绝
			// 判断请求类型，返回不同格式
			acceptHeader := c.GetHeader("Accept")

			// 如果是API请求（JSON格式）
			if strings.Contains(acceptHeader, "application/json") || strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(403, gin.H{
					"code":    403,
					"message": "域名未授权访问",
					"data": gin.H{
						"domain":    host,
						"clientIP":  c.ClientIP(),
						"requestID": c.GetString("request_id"),
					},
				})
			} else {
				// 浏览器请求，返回HTML错误页面
				htmlContent := strings.ReplaceAll(errorPageHTML, "{{.Domain}}", host)
				htmlContent = strings.ReplaceAll(htmlContent, "{{.ClientIP}}", c.ClientIP())
				htmlContent = strings.ReplaceAll(htmlContent, "{{.RequestID}}", c.GetString("request_id"))

				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(403, htmlContent)
			}

			c.Abort()
			return
		}

		c.Next()
	}
}
