package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/pkg/config"
	appLogger "commerce-platform/internal/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client redis.UniversalClient
	ctx    context.Context
}

var ErrCacheMiss = errors.New("cache miss")

const releaseLockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`

const refreshLockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("pexpire", KEYS[1], ARGV[2])
end
return 0
`

type RedisLock struct {
	cache *RedisCache
	key   string
	token string
}

// Client returns the underlying redis client
func (r *RedisCache) Client() redis.UniversalClient {
	return r.client
}

// Init 初始化Redis连接
func Init(cfg config.RedisConfig) (*RedisCache, error) {
	options := &redis.UniversalOptions{
		Addrs:                 cfg.GetRedisAddrs(),
		Username:              cfg.Username,
		Password:              cfg.Password,
		DB:                    cfg.DB,
		PoolSize:              cfg.PoolSize,
		MasterName:            cfg.MasterName,
		SentinelUsername:      cfg.SentinelUsername,
		SentinelPassword:      cfg.SentinelPassword,
		ContextTimeoutEnabled: true,
	}
	if cfg.NormalizedMode() != "sentinel" {
		options.MasterName = ""
		options.SentinelUsername = ""
		options.SentinelPassword = ""
	}
	client := redis.NewUniversalClient(options)

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	appLogger.Info("redis connected successfully")

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisCache) context(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	if r != nil && r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// SetContext 设置缓存
func (r *RedisCache) SetContext(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(r.context(ctx), key, data, ttl).Err()
}

// Set 设置缓存
func (r *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	return r.SetContext(r.ctx, key, value, ttl)
}

// GetContext 获取缓存
func (r *RedisCache) GetContext(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(r.context(ctx), key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	return nil
}

// Get 获取缓存
func (r *RedisCache) Get(key string, dest interface{}) error {
	return r.GetContext(r.ctx, key, dest)
}

// DeleteContext 删除缓存
func (r *RedisCache) DeleteContext(ctx context.Context, key string) error {
	return r.client.Del(r.context(ctx), key).Err()
}

// Delete 删除缓存
func (r *RedisCache) Delete(key string) error {
	return r.DeleteContext(r.ctx, key)
}

// DeletePatternContext 删除匹配模式的所有键
func (r *RedisCache) DeletePatternContext(ctx context.Context, pattern string) error {
	ctx = r.context(ctx)
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// DeletePattern 删除匹配模式的所有键
func (r *RedisCache) DeletePattern(pattern string) error {
	return r.DeletePatternContext(r.ctx, pattern)
}

// ExistsContext 检查键是否存在
func (r *RedisCache) ExistsContext(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(r.context(ctx), key).Result()
	return n > 0, err
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(key string) (bool, error) {
	return r.ExistsContext(r.ctx, key)
}

func (r *RedisCache) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*RedisLock, bool, error) {
	if ttl <= 0 {
		return nil, false, fmt.Errorf("lock ttl must be positive")
	}
	token, err := randomLockToken()
	if err != nil {
		return nil, false, err
	}
	ok, err := r.client.SetNX(r.context(ctx), key, token, ttl).Result()
	if err != nil || !ok {
		return nil, ok, err
	}
	return &RedisLock{cache: r, key: key, token: token}, true, nil
}

func (l *RedisLock) Release(ctx context.Context) error {
	if l == nil || l.cache == nil {
		return nil
	}
	return l.cache.client.Eval(l.cache.context(ctx), releaseLockScript, []string{l.key}, l.token).Err()
}

func (l *RedisLock) Refresh(ctx context.Context, ttl time.Duration) error {
	if l == nil || l.cache == nil {
		return nil
	}
	if ttl <= 0 {
		return fmt.Errorf("lock ttl must be positive")
	}

	result, err := l.cache.client.Eval(
		l.cache.context(ctx),
		refreshLockScript,
		[]string{l.key},
		l.token,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return fmt.Errorf("redis lock is no longer owned")
	}
	return nil
}

func randomLockToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate redis lock token: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// Close 关闭连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// PublishEvent 发布事件到 Redis Stream
func (r *RedisCache) PublishEvent(ctx context.Context, stream string, values map[string]interface{}) error {
	return r.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}

// ConsumeEventGroup 从 Redis Stream 的消费者组读取事件
func (r *RedisCache) ConsumeEventGroup(ctx context.Context, stream, group, consumer string) ([]redis.XMessage, error) {
	streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(streams) > 0 {
		return streams[0].Messages, nil
	}
	return []redis.XMessage{}, nil
}
