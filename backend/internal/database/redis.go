// Package database Redis连接管理
package database

import (
	"context"
	"fmt"
	"time"

	"feiniu-user-system/internal/config"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

// Redis键前缀常量
const (
	KeyUserToken     = "user:token:"      // 用户Token
	KeyUserInfo      = "user:info:"       // 用户信息缓存
	KeyMemberInfo    = "member:info:"     // 会员信息缓存
	KeyLoginLimit    = "limit:login:"     // 登录限流
	KeyAPILimit      = "limit:api:"       // API限流
	KeyVerifyCode    = "verify:code:"     // 验证码
	KeyBlacklist     = "blacklist:token:" // Token黑名单
	KeyResetPwdCode  = "reset:pwd:code:"  // 密码重置验证码
	KeyResetPwdLimit = "reset:pwd:limit:" // 密码重置限流
	KeyRegisterCode  = "register:code:"   // 注册验证码
	KeyRegisterLimit = "register:limit:"  // 注册验证码限流
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	return nil
}

// GetRedis 获取Redis实例
func GetRedis() *redis.Client {
	return RDB
}

// SetCache 设置缓存
func SetCache(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return RDB.Set(ctx, key, value, expiration).Err()
}

// GetCache 获取缓存
func GetCache(ctx context.Context, key string) (string, error) {
	return RDB.Get(ctx, key).Result()
}

// DeleteCache 删除缓存
func DeleteCache(ctx context.Context, keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}

// IncrLimit 限流计数器
func IncrLimit(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	pipe := RDB.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}
