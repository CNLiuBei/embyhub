package redis

import (
	"context"
	"fmt"
	"time"

	"embyhub/config"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	Client = redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// 测试连接
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("连接Redis失败: %w", err)
	}

	return nil
}

// Close 关闭Redis连接
func Close() error {
	if Client != nil {
		return Client.Close()
	}
	return nil
}

// Set 设置键值
func Set(key string, value interface{}, expiration time.Duration) error {
	return Client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

// Del 删除键
func Del(keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func Exists(keys ...string) (int64, error) {
	return Client.Exists(ctx, keys...).Result()
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return Client.Expire(ctx, key, expiration).Err()
}

// HSet 设置哈希字段
func HSet(key string, values ...interface{}) error {
	return Client.HSet(ctx, key, values...).Err()
}

// HGet 获取哈希字段
func HGet(key, field string) (string, error) {
	return Client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有哈希字段
func HGetAll(key string) (map[string]string, error) {
	return Client.HGetAll(ctx, key).Result()
}

// Incr 自增
func Incr(key string) (int64, error) {
	return Client.Incr(ctx, key).Result()
}

// TTL 获取剩余过期时间
func TTL(key string) (time.Duration, error) {
	return Client.TTL(ctx, key).Result()
}

// ExistsKey 检查单个键是否存在
func ExistsKey(key string) (bool, error) {
	result, err := Client.Exists(ctx, key).Result()
	return result > 0, err
}
