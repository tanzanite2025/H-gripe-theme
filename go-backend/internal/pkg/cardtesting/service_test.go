package cardtesting

import (
	"context"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"
)

type fakeStore struct {
	failures  map[string]int64
	blocks    map[string]time.Duration
	bindings  map[string]string
	deletions []string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		failures: make(map[string]int64),
		blocks:   make(map[string]time.Duration),
		bindings: make(map[string]string),
	}
}

func (s *fakeStore) CheckBlocked(_ context.Context, key string) (bool, time.Duration, error) {
	ttl, exists := s.blocks[key]
	return exists, ttl, nil
}

func (s *fakeStore) RecordFailure(
	_ context.Context,
	failureKeyValue string,
	blockKeyValue string,
	_ time.Duration,
	blockDuration time.Duration,
	threshold int64,
) (int64, bool, time.Duration, error) {
	if ttl, blocked := s.blocks[blockKeyValue]; blocked {
		return 0, true, ttl, nil
	}
	s.failures[failureKeyValue]++
	count := s.failures[failureKeyValue]
	if count >= threshold {
		s.blocks[blockKeyValue] = blockDuration
		return count, true, blockDuration, nil
	}
	return count, false, 0, nil
}

func (s *fakeStore) Delete(_ context.Context, key string) error {
	delete(s.failures, key)
	s.deletions = append(s.deletions, key)
	return nil
}

func (s *fakeStore) BindPaymentIntent(_ context.Context, paymentIntentID, failureKeyValue string, _ time.Duration) error {
	s.bindings[paymentIntentID] = failureKeyValue
	return nil
}

func (s *fakeStore) PaymentIntentBinding(_ context.Context, paymentIntentID string) (string, error) {
	return s.bindings[paymentIntentID], nil
}

func (s *fakeStore) DeletePaymentIntentBinding(_ context.Context, paymentIntentID string) error {
	delete(s.bindings, paymentIntentID)
	return nil
}

func testConfig() config.PaymentBINRateLimitConfig {
	return config.PaymentBINRateLimitConfig{
		Enabled:              true,
		WindowSeconds:        60,
		FailureThreshold:     5,
		BlockDurationSeconds: 1800,
	}
}

func TestBINLimiterBlocksAfterFiveFailuresAcrossPaymentIntents(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	service := newWithStore(store, testConfig())
	bin := "41111111"

	for index := 1; index <= 4; index++ {
		paymentIntentID := "pi_failure_" + string(rune('0'+index))
		if err := service.BindPaymentIntent(ctx, paymentIntentID, bin); err != nil {
			t.Fatalf("BindPaymentIntent() error = %v", err)
		}
		decision, err := service.RecordPaymentIntentFailure(ctx, paymentIntentID)
		if err != nil {
			t.Fatalf("RecordPaymentIntentFailure() error = %v", err)
		}
		if decision.Blocked {
			t.Fatalf("failure %d unexpectedly blocked BIN", index)
		}
	}

	if decision, err := service.Check(ctx, bin); err != nil {
		t.Fatalf("Check() error = %v", err)
	} else if decision.Blocked {
		t.Fatal("BIN should not be blocked before the fifth failure")
	}

	if err := service.BindPaymentIntent(ctx, "pi_failure_5", bin); err != nil {
		t.Fatalf("BindPaymentIntent() error = %v", err)
	}
	decision, err := service.RecordPaymentIntentFailure(ctx, "pi_failure_5")
	if err != nil {
		t.Fatalf("RecordPaymentIntentFailure() error = %v", err)
	}
	if !decision.Blocked || decision.Count != 5 {
		t.Fatalf("decision = %#v, want blocked at count 5", decision)
	}
	if decision.RetryAfter != 1800*time.Second {
		t.Fatalf("RetryAfter = %s, want 30m", decision.RetryAfter)
	}

	if decision, err := service.Check(ctx, bin); err != nil {
		t.Fatalf("Check() error = %v", err)
	} else if !decision.Blocked {
		t.Fatal("BIN should remain blocked after the fifth failure")
	}
	if decision, err := service.Check(ctx, "555555"); err != nil {
		t.Fatalf("Check() for another BIN error = %v", err)
	} else if decision.Blocked {
		t.Fatal("a different BIN must not be blocked")
	}
}

func TestBINLimiterSuccessClearsFailureWindow(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()
	service := newWithStore(store, testConfig())
	bin := "411111"

	for index := 1; index <= 4; index++ {
		if _, err := service.RecordFailure(ctx, bin); err != nil {
			t.Fatalf("RecordFailure() error = %v", err)
		}
	}
	if err := service.RecordSuccess(ctx, bin); err != nil {
		t.Fatalf("RecordSuccess() error = %v", err)
	}
	if len(store.failures) != 0 {
		t.Fatalf("failures = %#v, want empty after success", store.failures)
	}
	decision, err := service.RecordFailure(ctx, bin)
	if err != nil {
		t.Fatalf("RecordFailure() after success error = %v", err)
	}
	if decision.Count != 1 || decision.Blocked {
		t.Fatalf("decision = %#v, want a fresh first failure", decision)
	}
}

func TestBINLimiterDoesNotPersistRawBIN(t *testing.T) {
	store := newFakeStore()
	service := newWithStore(store, testConfig())

	if err := service.BindPaymentIntent(context.Background(), "pi_secret", "41111111"); err != nil {
		t.Fatalf("BindPaymentIntent() error = %v", err)
	}
	for paymentIntentID, value := range store.bindings {
		if strings.Contains(value, "41111111") || strings.Contains(paymentIntentID, "41111111") {
			t.Fatalf("binding leaked raw BIN: %q -> %q", paymentIntentID, value)
		}
	}
	for key := range store.failures {
		if strings.Contains(key, "41111111") {
			t.Fatalf("failure key leaked raw BIN: %q", key)
		}
	}
}
