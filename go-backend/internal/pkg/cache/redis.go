package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"tanzanite/internal/pkg/config"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// Init 初始化Redis连接
func Init(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Redis connected successfully")

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set 设置缓存
func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(r.ctx, key, data, ttl).Err()
}

// Get 获取缓存
func (r *RedisCache) Get(key string, dest interface{}) error {
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("cache miss")
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// DeletePattern 删除匹配模式的所有键
func (r *RedisCache) DeletePattern(pattern string) error {
	iter := r.client.Scan(r.ctx, 0, pattern, 0).Iterator()
	for iter.Next(r.ctx) {
		if err := r.client.Del(r.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	n, err := r.client.Exists(r.ctx, key).Result()
	return n > 0, err
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}
