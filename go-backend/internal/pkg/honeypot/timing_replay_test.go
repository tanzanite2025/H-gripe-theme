package honeypot

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestTimingReplayObserverRecordsFirstAndRepeatedNonce(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	observer := NewTimingReplayObserver(client)
	now := time.Unix(1_700_000_000, 0).UTC()
	claims := TimingClaims{Form: "newsletter", IssuedAt: now, Nonce: "nonce-1"}

	if got := observer.Observe(context.Background(), "newsletter", claims, now.Add(time.Second), time.Minute); got != TimingReplayFirstSeen {
		t.Fatalf("first replay observation = %q, want %q", got, TimingReplayFirstSeen)
	}
	if got := observer.Observe(context.Background(), "newsletter", claims, now.Add(2*time.Second), time.Minute); got != TimingReplayReused {
		t.Fatalf("second replay observation = %q, want %q", got, TimingReplayReused)
	}
	for _, key := range mini.Keys() {
		if strings.Contains(key, claims.Nonce) {
			t.Fatalf("Redis replay key leaked raw nonce: %q", key)
		}
	}
}

func TestTimingReplayObserverFailsOpenWhenRedisIsUnavailable(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	observer := NewTimingReplayObserver(client)
	_ = client.Close()

	result := observer.Observe(context.Background(), "newsletter", TimingClaims{
		Form:     "newsletter",
		IssuedAt: time.Now().Add(-time.Second),
		Nonce:    "nonce-2",
	}, time.Now(), time.Minute)
	if result != TimingReplayError {
		t.Fatalf("unavailable Redis result = %q, want %q", result, TimingReplayError)
	}
}
