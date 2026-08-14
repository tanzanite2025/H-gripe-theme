package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newTestRedisCache(t *testing.T) (*miniredis.Miniredis, *RedisCache) {
	t.Helper()

	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() {
		_ = client.Close()
	})

	return server, &RedisCache{
		client: client,
		ctx:    context.Background(),
	}
}

func TestRedisCacheGetContextReturnsStableCacheMissError(t *testing.T) {
	_, redisCache := newTestRedisCache(t)

	var destination map[string]string
	err := redisCache.GetContext(context.Background(), "missing", &destination)

	require.ErrorIs(t, err, ErrCacheMiss)
}

func TestRedisCacheDistributedLockUsesCompareAndDeleteRelease(t *testing.T) {
	server, redisCache := newTestRedisCache(t)
	ctx := context.Background()

	firstLock, acquired, err := redisCache.AcquireLock(ctx, "lock:product:1", time.Second)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotNil(t, firstLock)

	_, acquired, err = redisCache.AcquireLock(ctx, "lock:product:1", time.Second)
	require.NoError(t, err)
	require.False(t, acquired)

	server.FastForward(2 * time.Second)
	secondLock, acquired, err := redisCache.AcquireLock(ctx, "lock:product:1", time.Second)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotNil(t, secondLock)

	require.NoError(t, firstLock.Release(ctx))
	_, acquired, err = redisCache.AcquireLock(ctx, "lock:product:1", time.Second)
	require.NoError(t, err)
	require.False(t, acquired)

	require.NoError(t, secondLock.Release(ctx))
	thirdLock, acquired, err := redisCache.AcquireLock(ctx, "lock:product:1", time.Second)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NotNil(t, thirdLock)
	require.NoError(t, thirdLock.Release(ctx))
}

func TestRedisCacheDistributedLockCanRefreshLease(t *testing.T) {
	server, redisCache := newTestRedisCache(t)
	ctx := context.Background()

	lock, acquired, err := redisCache.AcquireLock(ctx, "lock:product:refresh", time.Second)
	require.NoError(t, err)
	require.True(t, acquired)

	server.FastForward(750 * time.Millisecond)
	require.NoError(t, lock.Refresh(ctx, time.Second))

	server.FastForward(500 * time.Millisecond)
	_, acquired, err = redisCache.AcquireLock(ctx, "lock:product:refresh", time.Second)
	require.NoError(t, err)
	require.False(t, acquired)

	require.NoError(t, lock.Release(ctx))
}

func TestRedisCacheSetAndGetContextUsesRequestContext(t *testing.T) {
	_, redisCache := newTestRedisCache(t)
	ctx := context.Background()

	require.NoError(t, redisCache.SetContext(ctx, "product:1", map[string]string{"status": "active"}, time.Minute))

	var destination map[string]string
	require.NoError(t, redisCache.GetContext(ctx, "product:1", &destination))
	require.Equal(t, map[string]string{"status": "active"}, destination)

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	err := redisCache.GetContext(cancelled, "product:1", &destination)
	require.Error(t, err)
	require.False(t, errors.Is(err, ErrCacheMiss))
}
