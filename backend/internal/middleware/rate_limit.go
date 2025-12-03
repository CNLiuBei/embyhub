package middleware

import (
	"fmt"
	"net/http"
	"time"

	"embyhub/pkg/redis"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	MaxRequests int           // 最大请求数
	Window      time.Duration // 时间窗口
	KeyPrefix   string        // Redis键前缀
}

// 默认配置：每分钟60次请求
var defaultConfig = RateLimitConfig{
	MaxRequests: 60,
	Window:      time.Minute,
	KeyPrefix:   "emby_ums:rate_limit:",
}

// RateLimitMiddleware 请求频率限制中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(defaultConfig)
}

// RateLimitMiddlewareWithConfig 带配置的限流中间件
func RateLimitMiddlewareWithConfig(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端IP
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s%s", config.KeyPrefix, clientIP)

		// 增加计数
		count, err := redis.Incr(key)
		if err != nil {
			// Redis错误时放行，避免影响正常请求
			c.Next()
			return
		}

		// 首次请求，设置过期时间
		if count == 1 {
			redis.Expire(key, config.Window)
		}

		// 获取剩余时间
		ttl, _ := redis.TTL(key)

		// 设置响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.MaxRequests)-count)))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

		// 超过限制
		if count > int64(config.MaxRequests) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": fmt.Sprintf("请求过于频繁，请%d秒后重试", int(ttl.Seconds())),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimitMiddleware 登录接口限流（更严格：每分钟10次）
func LoginRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(RateLimitConfig{
		MaxRequests: 10,
		Window:      time.Minute,
		KeyPrefix:   "emby_ums:rate_limit:login:",
	})
}

// APIRateLimitMiddleware API接口限流（每分钟100次）
func APIRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddlewareWithConfig(RateLimitConfig{
		MaxRequests: 100,
		Window:      time.Minute,
		KeyPrefix:   "emby_ums:rate_limit:api:",
	})
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
