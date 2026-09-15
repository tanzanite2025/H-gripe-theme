package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const releaseSchedulerLockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`

type schedulerLock struct {
	client redis.UniversalClient
	key    string
	token  string
}

func acquireSchedulerLock(ctx context.Context, client redis.UniversalClient, key string, ttl time.Duration) (*schedulerLock, bool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		return nil, false, fmt.Errorf("redis client is required for scheduler lease")
	}
	if ttl <= 0 {
		return nil, false, fmt.Errorf("scheduler lease ttl must be positive")
	}
	token, err := schedulerLockToken()
	if err != nil {
		return nil, false, err
	}
	ok, err := client.SetNX(ctx, key, token, ttl).Result()
	if err != nil || !ok {
		return nil, ok, err
	}
	return &schedulerLock{client: client, key: key, token: token}, true, nil
}

func (l *schedulerLock) release(ctx context.Context) error {
	if l == nil || l.client == nil {
		return nil
	}
	return l.client.Eval(ctx, releaseSchedulerLockScript, []string{l.key}, l.token).Err()
}

func (l *schedulerLock) refresh(ctx context.Context, ttl time.Duration) error {
	if l == nil || l.client == nil {
		return nil
	}
	if ttl <= 0 {
		return fmt.Errorf("scheduler lease ttl must be positive")
	}
	result, err := l.client.Eval(ctx, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("pexpire", KEYS[1], ARGV[2])
end
return 0
`, []string{l.key}, l.token, ttl.Milliseconds()).Int64()
	if err != nil {
		return err
	}
	if result != 1 {
		return fmt.Errorf("scheduler lease is no longer owned")
	}
	return nil
}

func maintainSchedulerLock(lock *schedulerLock, ttl time.Duration) func() {
	if lock == nil || ttl <= 0 {
		return func() {}
	}
	refreshCtx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	interval := ttl / 3
	if interval < time.Second {
		interval = time.Second
	}
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				if err := lock.refresh(refreshCtx, ttl); err != nil {
					return
				}
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

func schedulerLockToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate scheduler lease token: %w", err)
	}
	return hex.EncodeToString(token[:]), nil
}
