package orderabuse

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"tanzanite/internal/pkg/config"
)

func TestEvaluateBlocksIdentityAfterConfiguredLimit(t *testing.T) {
	store := &fakeCounterStore{}
	service := &Service{
		store: store,
		cfg: config.OrderAbuseConfig{
			Enabled:                     true,
			OrderCreateWindowSeconds:    600,
			MaxOrderCreationsPerUser:    2,
			MaxOrderCreationsPerSession: 0,
			MaxOrderCreationsPerIP:      0,
		},
	}
	identity := Identity{UserID: 42}

	for attempt := 0; attempt < 2; attempt++ {
		decision, err := service.Evaluate(context.Background(), identity)
		if err != nil {
			t.Fatalf("Evaluate() error = %v", err)
		}
		if !decision.Allowed {
			t.Fatalf("attempt %d unexpectedly blocked: %#v", attempt+1, decision)
		}
	}

	decision, err := service.Evaluate(context.Background(), identity)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Allowed {
		t.Fatal("third attempt should be blocked")
	}
	if decision.Action != ActionBlock || decision.Dimension != "user" {
		t.Fatalf("decision = %#v, want user block", decision)
	}
	if decision.Count != 3 || decision.Limit != 2 || decision.RetryAfter != 10*time.Minute {
		t.Fatalf("decision = %#v, want count=3 limit=2 retry_after=10m", decision)
	}
}

func TestEvaluateHashesSensitiveIdentityValues(t *testing.T) {
	store := &fakeCounterStore{}
	service := &Service{
		store: store,
		cfg: config.OrderAbuseConfig{
			Enabled:                     true,
			OrderCreateWindowSeconds:    60,
			MaxOrderCreationsPerUser:    1,
			MaxOrderCreationsPerSession: 1,
			MaxOrderCreationsPerIP:      1,
		},
	}
	identity := Identity{
		UserID:    42,
		SessionID: "session-secret",
		IPAddress: "203.0.113.10",
	}

	_, err := service.Evaluate(context.Background(), identity)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	joined := strings.Join(store.keys, "|")
	for _, raw := range []string{
		"tanzanite:order-abuse:create:user:42",
		"session-secret",
		"203.0.113.10",
	} {
		if strings.Contains(joined, raw) {
			t.Fatalf("counter key leaked raw value %q in %q", raw, joined)
		}
	}
	if len(store.keys) != 3 {
		t.Fatalf("counter keys = %#v, want three dimensions", store.keys)
	}
}

func TestEvaluateDoesNotUseStoreWhenDisabled(t *testing.T) {
	store := &fakeCounterStore{}
	service := &Service{
		store: store,
		cfg: config.OrderAbuseConfig{
			Enabled: false,
		},
	}

	decision, err := service.Evaluate(context.Background(), Identity{UserID: 42})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !decision.Allowed || decision.Action != ActionAllow {
		t.Fatalf("decision = %#v, want allow", decision)
	}
	if len(store.keys) != 0 {
		t.Fatalf("disabled service should not write counters: %#v", store.keys)
	}
}

func TestEvaluateReturnsUnavailableWhenCounterFails(t *testing.T) {
	service := &Service{
		store: &fakeCounterStore{err: errors.New("redis unavailable")},
		cfg: config.OrderAbuseConfig{
			Enabled:                  true,
			OrderCreateWindowSeconds: 60,
			MaxOrderCreationsPerUser: 1,
		},
	}

	_, err := service.Evaluate(context.Background(), Identity{UserID: 42})
	if !errors.Is(err, ErrServiceUnavailable) {
		t.Fatalf("Evaluate() error = %v, want ErrServiceUnavailable", err)
	}
}

type fakeCounterStore struct {
	counts map[string]int64
	keys   []string
	err    error
}

func (s *fakeCounterStore) Increment(_ context.Context, key string, _ time.Duration) (int64, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.counts == nil {
		s.counts = map[string]int64{}
	}
	s.keys = append(s.keys, key)
	s.counts[key]++
	return s.counts[key], nil
}
