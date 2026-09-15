package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestAcquireSchedulerLockAllowsSingleOwnerAndOwnerOnlyRelease(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	first, acquired, err := acquireSchedulerLock(context.Background(), client, "scheduler:test", time.Minute)
	require.NoError(t, err)
	require.True(t, acquired)

	second, acquired, err := acquireSchedulerLock(context.Background(), client, "scheduler:test", time.Minute)
	require.NoError(t, err)
	require.False(t, acquired)
	require.Nil(t, second)

	third := &schedulerLock{client: client, key: "scheduler:test", token: "wrong-owner"}
	require.NoError(t, third.release(context.Background()))
	value, err := server.Get("scheduler:test")
	require.NoError(t, err)
	require.Equal(t, first.token, value)

	require.NoError(t, first.release(context.Background()))
	require.False(t, server.Exists("scheduler:test"))
}

func TestSchedulerLockRefreshExtendsOnlyOwnedLease(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	lock, acquired, err := acquireSchedulerLock(context.Background(), client, "scheduler:refresh", time.Second)
	require.NoError(t, err)
	require.True(t, acquired)
	require.NoError(t, lock.refresh(context.Background(), time.Minute))
	require.Greater(t, server.TTL("scheduler:refresh"), time.Second)

	wrong := &schedulerLock{client: client, key: "scheduler:refresh", token: "wrong-owner"}
	require.Error(t, wrong.refresh(context.Background(), time.Minute))
}
