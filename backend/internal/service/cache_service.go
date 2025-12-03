package service

import (
	"encoding/json"
	"fmt"
	"time"

	"embyhub/internal/model"
	"embyhub/pkg/redis"
)

// CacheService 缓存服务
type CacheService struct{}

// NewCacheService 创建缓存服务
func NewCacheService() *CacheService {
	return &CacheService{}
}

// 缓存键前缀
const (
	CacheKeyUserInfo   = "emby_ums:cache:user:%d"    // 用户信息缓存
	CacheKeyUserPerms  = "emby_ums:cache:perms:%d"   // 用户权限缓存
	CacheKeyStatistics = "emby_ums:cache:statistics" // 统计数据缓存
	CacheKeyCardStats  = "emby_ums:cache:card_stats" // 卡密统计缓存
	CacheKeyVipStats   = "emby_ums:cache:vip_stats"  // VIP统计缓存
)

// 缓存过期时间
const (
	UserCacheTTL       = 30 * time.Minute // 用户信息30分钟
	PermsCacheTTL      = 1 * time.Hour    // 权限1小时
	StatisticsCacheTTL = 1 * time.Minute  // 统计1分钟
	CardStatsCacheTTL  = 1 * time.Minute  // 卡密统计1分钟
)

// GetUserInfo 获取用户信息（带缓存）
func (s *CacheService) GetUserInfo(userID int) (*model.User, error) {
	key := fmt.Sprintf(CacheKeyUserInfo, userID)

	// 尝试从缓存获取
	data, err := redis.Get(key)
	if err == nil && data != "" {
		var user model.User
		if err := json.Unmarshal([]byte(data), &user); err == nil {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("缓存未命中")
}

// SetUserInfo 设置用户信息缓存
func (s *CacheService) SetUserInfo(user *model.User) error {
	key := fmt.Sprintf(CacheKeyUserInfo, user.UserID)
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return redis.Set(key, string(data), UserCacheTTL)
}

// InvalidateUserInfo 使用户信息缓存失效
func (s *CacheService) InvalidateUserInfo(userID int) {
	key := fmt.Sprintf(CacheKeyUserInfo, userID)
	redis.Del(key)
}

// GetUserPermissions 获取用户权限（带缓存）
func (s *CacheService) GetUserPermissions(userID int) ([]string, error) {
	key := fmt.Sprintf(CacheKeyUserPerms, userID)

	data, err := redis.Get(key)
	if err == nil && data != "" {
		var perms []string
		if err := json.Unmarshal([]byte(data), &perms); err == nil {
			return perms, nil
		}
	}

	return nil, fmt.Errorf("缓存未命中")
}

// SetUserPermissions 设置用户权限缓存
func (s *CacheService) SetUserPermissions(userID int, permissions []string) error {
	key := fmt.Sprintf(CacheKeyUserPerms, userID)
	data, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return redis.Set(key, string(data), PermsCacheTTL)
}

// InvalidateUserPermissions 使用户权限缓存失效
func (s *CacheService) InvalidateUserPermissions(userID int) {
	key := fmt.Sprintf(CacheKeyUserPerms, userID)
	redis.Del(key)
}

// GetStatistics 获取统计数据（带缓存）
func (s *CacheService) GetStatistics() (*model.StatisticsResponse, error) {
	data, err := redis.Get(CacheKeyStatistics)
	if err == nil && data != "" {
		var stats model.StatisticsResponse
		if err := json.Unmarshal([]byte(data), &stats); err == nil {
			return &stats, nil
		}
	}
	return nil, fmt.Errorf("缓存未命中")
}

// SetStatistics 设置统计数据缓存
func (s *CacheService) SetStatistics(stats *model.StatisticsResponse) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	return redis.Set(CacheKeyStatistics, string(data), StatisticsCacheTTL)
}

// InvalidateStatistics 使统计数据缓存失效
func (s *CacheService) InvalidateStatistics() {
	redis.Del(CacheKeyStatistics)
	redis.Del(CacheKeyCardStats)
	redis.Del(CacheKeyVipStats)
}

// 全局缓存服务实例
var cacheService = NewCacheService()

// Cache 获取全局缓存服务
func Cache() *CacheService {
	return cacheService
}
