package cache

import (
	"context"
	"time"

	"golang-rest-api-template/pkg/config"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Keys(context.Context, string) *redis.StringSliceCmd
	Del(context.Context, ...string) *redis.IntCmd
}

// NewRedisClient 根据配置创建 Redis 客户端
func NewRedisClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.RedisHost + ":" + cfg.RedisPort,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		IdleTimeout:  time.Duration(cfg.RedisIdleTimeoutSec) * time.Second,
		DialTimeout:  time.Duration(cfg.RedisDialTimeoutMs) * time.Millisecond,
		ReadTimeout:  time.Duration(cfg.RedisReadTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.RedisWriteTimeoutMs) * time.Millisecond,
	})
}
