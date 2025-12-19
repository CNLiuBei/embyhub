// Package middleware 限流中间件
package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"feiniu-user-system/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RateLimiter 限流器
type RateLimiter struct {
	requests map[string]*userLimit
	mu       sync.RWMutex
	limit    int           // 允许的请求次数
	window   time.Duration // 时间窗口
}

// userLimit 用户限制
type userLimit struct {
	count      int
	firstTime  time.Time
	lastTime   time.Time
	violations int // 违规次数
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userLimit),
		limit:    limit,
		window:   window,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// cleanup 定期清理过期数据
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, limit := range rl.requests {
			if now.Sub(limit.lastTime) > rl.window*2 {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) (bool, *userLimit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.requests[key]

	if !exists {
		// 首次请求
		rl.requests[key] = &userLimit{
			count:     1,
			firstTime: now,
			lastTime:  now,
		}
		return true, rl.requests[key]
	}

	// 检查时间窗口
	if now.Sub(limit.firstTime) > rl.window {
		// 重置窗口
		limit.count = 1
		limit.firstTime = now
		limit.lastTime = now
		return true, limit
	}

	// 在窗口内
	limit.count++
	limit.lastTime = now

	if limit.count > rl.limit {
		limit.violations++
		return false, limit
	}

	return true, limit
}

// RedeemRateLimit 卡密兑换限流中间件
func RedeemRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// 获取用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			response.BadRequest(c, "用户ID格式错误")
			c.Abort()
			return
		}

		// 检查用户限流
		userKey := fmt.Sprintf("user:%s", uid.String())
		allowed, userLimit := limiter.Allow(userKey)

		if !allowed {
			remaining := int(window.Seconds()) - int(time.Since(userLimit.firstTime).Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("兑换过于频繁，请在%d秒后再试", remaining),
			})
			c.Abort()
			return
		}

		// 同时检查IP限流（防止换账号刷）
		ip := c.ClientIP()
		ipKey := fmt.Sprintf("ip:%s", ip)
		allowed, ipLimit := limiter.Allow(ipKey)

		if !allowed {
			remaining := int(window.Seconds()) - int(time.Since(ipLimit.firstTime).Seconds())

			// 记录可疑行为
			if ipLimit.violations > 5 {
				// TODO: 可以在这里添加到黑名单或发送警告
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("操作过于频繁，请在%d秒后再试", remaining),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GlobalRateLimit 全局限流中间件
func GlobalRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := fmt.Sprintf("global:%s", ip)

		allowed, _ := limiter.Allow(key)
		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPBlacklist IP黑名单中间件
type IPBlacklist struct {
	blacklist map[string]time.Time
	mu        sync.RWMutex
}

// NewIPBlacklist 创建IP黑名单
func NewIPBlacklist() *IPBlacklist {
	bl := &IPBlacklist{
		blacklist: make(map[string]time.Time),
	}

	// 启动清理协程
	go bl.cleanup()

	return bl
}

// cleanup 定期清理过期黑名单
func (bl *IPBlacklist) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.mu.Lock()
		now := time.Now()
		for ip, expireTime := range bl.blacklist {
			if now.After(expireTime) {
				delete(bl.blacklist, ip)
			}
		}
		bl.mu.Unlock()
	}
}

// Add 添加到黑名单
func (bl *IPBlacklist) Add(ip string, duration time.Duration) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.blacklist[ip] = time.Now().Add(duration)
}

// IsBlocked 检查是否在黑名单中
func (bl *IPBlacklist) IsBlocked(ip string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expireTime, exists := bl.blacklist[ip]
	if !exists {
		return false
	}

	if time.Now().After(expireTime) {
		return false
	}

	return true
}

// Middleware 黑名单中间件
func (bl *IPBlacklist) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if bl.IsBlocked(ip) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "您的IP已被暂时封禁",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 全局黑名单实例
var GlobalBlacklist = NewIPBlacklist()


// 全局限流器实例
var (
	// 发帖限流：每用户每小时最多10个帖子
	topicLimiter = NewRateLimiter(10, time.Hour)
	// 评论限流：每用户每分钟最多20条评论
	commentLimiter = NewRateLimiter(20, time.Minute)
	// 私信限流：每用户每分钟最多30条消息
	messageLimiter = NewRateLimiter(30, time.Minute)
)

// TopicRateLimitMiddleware 发帖限流中间件
func TopicRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("topic:%s", uid.String())
		allowed, limit := topicLimiter.Allow(key)
		if !allowed {
			remaining := int(time.Hour.Seconds()) - int(time.Since(limit.firstTime).Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("发帖过于频繁，每小时最多发布10个帖子，请在%d分钟后再试", remaining/60),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CommentRateLimitMiddleware 评论限流中间件
func CommentRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("comment:%s", uid.String())
		allowed, limit := commentLimiter.Allow(key)
		if !allowed {
			remaining := int(time.Minute.Seconds()) - int(time.Since(limit.firstTime).Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("评论过于频繁，每分钟最多发布20条评论，请在%d秒后再试", remaining),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// MessageRateLimitMiddleware 私信限流中间件
func MessageRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}

		key := fmt.Sprintf("message:%s", uid.String())
		allowed, limit := messageLimiter.Allow(key)
		if !allowed {
			remaining := int(time.Minute.Seconds()) - int(time.Since(limit.firstTime).Seconds())
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("发送消息过于频繁，请在%d秒后再试", remaining),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
