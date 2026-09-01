package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"
)

func TestShowcaseUploadProtectionPendingLimitBlocksBeforeStore(t *testing.T) {
	store := newMemoryShowcaseUploadProtectionStore()
	protection := &UGCShowcaseUploadProtectionService{
		store: store,
		cfg: config.ShowcaseUploadProtectionConfig{
			Enabled:                      true,
			MaxPendingSubmissionsPerUser: 2,
		},
		now: time.Now,
	}

	decision, err := protection.Evaluate(context.Background(), UGCShowcaseUploadProtectionInput{
		Identity:           UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"},
		PendingSubmissions: 2,
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Allowed || decision.Dimension != "pending_submissions" {
		t.Fatalf("decision = %#v, want pending_submissions block", decision)
	}
	if store.evaluateCalls != 0 {
		t.Fatalf("pending limit should block before Redis evaluation, got %d calls", store.evaluateCalls)
	}
}

func TestShowcaseUploadProtectionBlocksAfterUserWindowLimit(t *testing.T) {
	protection := newTestShowcaseUploadProtectionService(newMemoryShowcaseUploadProtectionStore(), config.ShowcaseUploadProtectionConfig{
		WindowSeconds:          60,
		MaxUploadsPerUser:      2,
		MaxUploadsPerIP:        100,
		MaxUploadsPerIPPrefix:  100,
		DailyMaxUploadsPerUser: 100,
		DailyMaxUploadsPerIP:   100,
		DailyMaxBytesPerUser:   100 << 20,
		DailyMaxBytesPerIP:     100 << 20,
	})

	input := UGCShowcaseUploadProtectionInput{
		Identity:    UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"},
		UploadBytes: 1024,
	}
	for attempt := 1; attempt <= 2; attempt++ {
		decision, err := protection.Evaluate(context.Background(), input)
		if err != nil {
			t.Fatalf("Evaluate() attempt %d error = %v", attempt, err)
		}
		if !decision.Allowed {
			t.Fatalf("attempt %d unexpectedly blocked: %#v", attempt, decision)
		}
	}

	decision, err := protection.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Allowed || decision.Dimension != "user_window" {
		t.Fatalf("decision = %#v, want user_window block", decision)
	}
	if decision.Count != 3 || decision.Limit != 2 || decision.RetryAfter != time.Minute {
		t.Fatalf("decision = %#v, want count=3 limit=2 retry_after=1m", decision)
	}
}

func TestShowcaseUploadProtectionEvaluationChargesNormalCounters(t *testing.T) {
	store := newMemoryShowcaseUploadProtectionStore()
	protection := newTestShowcaseUploadProtectionService(store, config.ShowcaseUploadProtectionConfig{})
	input := UGCShowcaseUploadProtectionInput{
		Identity:    UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"},
		UploadBytes: 1024,
	}

	decision, err := protection.Evaluate(context.Background(), input)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !decision.Allowed {
		t.Fatalf("decision = %#v, want allow", decision)
	}
	if len(store.evaluateKeys) != 9 {
		t.Fatalf("Evaluate() keys = %d, want 9", len(store.evaluateKeys))
	}
	if len(store.failureKeys) != 0 {
		t.Fatalf("Evaluate() unexpectedly touched failure keys: %v", store.failureKeys)
	}
	for index, key := range store.evaluateKeys[:7] {
		want := int64(1)
		if index >= 5 {
			want = input.UploadBytes
		}
		if got := store.counts[key]; got != want {
			t.Fatalf("normal counter %q = %d, want %d", key, got, want)
		}
	}
	if got := store.counts[store.evaluateKeys[7]]; got != 0 {
		t.Fatalf("block counter %q = %d, want 0", store.evaluateKeys[7], got)
	}
	if got := store.counts[store.evaluateKeys[8]]; got != 0 {
		t.Fatalf("block counter %q = %d, want 0", store.evaluateKeys[8], got)
	}
}

func TestShowcaseUploadProtectionFailureOnlyChargesFailureCounters(t *testing.T) {
	store := newMemoryShowcaseUploadProtectionStore()
	protection := newTestShowcaseUploadProtectionService(store, config.ShowcaseUploadProtectionConfig{})
	identity := UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"}

	if err := protection.RecordFailure(context.Background(), identity); err != nil {
		t.Fatalf("RecordFailure() error = %v", err)
	}
	if len(store.evaluateKeys) != 0 {
		t.Fatalf("RecordFailure() unexpectedly evaluated normal keys: %v", store.evaluateKeys)
	}
	if len(store.failureKeys) != 4 {
		t.Fatalf("RecordFailure() keys = %d, want 4", len(store.failureKeys))
	}
	if got := store.counts[store.failureKeys[0]]; got != 1 {
		t.Fatalf("user failure counter = %d, want 1", got)
	}
	if got := store.counts[store.failureKeys[1]]; got != 1 {
		t.Fatalf("IP failure counter = %d, want 1", got)
	}
}

func TestShowcaseUploadProtectionFailureThresholdBlocksFutureEvaluation(t *testing.T) {
	store := newMemoryShowcaseUploadProtectionStore()
	protection := newTestShowcaseUploadProtectionService(store, config.ShowcaseUploadProtectionConfig{
		MaxFailuresPerUser: 2,
		MaxFailuresPerIP:   100,
	})
	identity := UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"}

	for attempt := 1; attempt <= 2; attempt++ {
		if err := protection.RecordFailure(context.Background(), identity); err != nil {
			t.Fatalf("RecordFailure() attempt %d error = %v", attempt, err)
		}
	}

	decision, err := protection.Evaluate(context.Background(), UGCShowcaseUploadProtectionInput{
		Identity:    identity,
		UploadBytes: 1024,
	})
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if decision.Allowed || decision.Dimension != "user_block" {
		t.Fatalf("decision = %#v, want user_block", decision)
	}
}

func TestShowcaseUploadProtectionKeysDoNotLeakRawIP(t *testing.T) {
	store := newMemoryShowcaseUploadProtectionStore()
	protection := newTestShowcaseUploadProtectionService(store, config.ShowcaseUploadProtectionConfig{})
	identity := UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"}

	if _, err := protection.Evaluate(context.Background(), UGCShowcaseUploadProtectionInput{
		Identity:    identity,
		UploadBytes: 1024,
	}); err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if err := protection.RecordFailure(context.Background(), identity); err != nil {
		t.Fatalf("RecordFailure() error = %v", err)
	}

	joined := strings.Join(append(store.evaluateKeys, store.failureKeys...), "|")
	for _, raw := range []string{"203.0.113.10", "203.0.113.0/24"} {
		if strings.Contains(joined, raw) {
			t.Fatalf("showcase upload protection key leaked raw value %q in %q", raw, joined)
		}
	}
}

func TestShowcaseUploadProtectionIPPrefixNormalization(t *testing.T) {
	cases := map[string]string{
		"203.0.113.10":           "203.0.113.0/24",
		" 2001:db8:abcd:12::99 ": "2001:db8:abcd:12::/64",
		"not-an-ip-address":      "not-an-ip-address",
	}

	for input, expected := range cases {
		if actual := ipPrefix(input); actual != expected {
			t.Fatalf("ipPrefix(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestShowcaseUploadProtectionReturnsUnavailableWhenStoreFails(t *testing.T) {
	expectedErr := errors.New("redis unavailable")
	protection := newTestShowcaseUploadProtectionService(&memoryShowcaseUploadProtectionStore{err: expectedErr}, config.ShowcaseUploadProtectionConfig{})

	_, err := protection.Evaluate(context.Background(), UGCShowcaseUploadProtectionInput{
		Identity:    UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"},
		UploadBytes: 1024,
	})
	if !errors.Is(err, ErrShowcaseUploadProtectionUnavailable) {
		t.Fatalf("Evaluate() error = %v, want ErrShowcaseUploadProtectionUnavailable", err)
	}

	err = protection.RecordFailure(context.Background(), UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"})
	if !errors.Is(err, ErrShowcaseUploadProtectionUnavailable) {
		t.Fatalf("RecordFailure() error = %v, want ErrShowcaseUploadProtectionUnavailable", err)
	}
}

func TestShowcaseUploadProtectionFailsClosedWhenRedisIsMissing(t *testing.T) {
	cfg := config.ShowcaseUploadProtectionConfig{
		Enabled:                      true,
		MaxPendingSubmissionsPerUser: 10,
	}
	protection := NewUGCShowcaseUploadProtectionService(nil, cfg)

	_, err := protection.Evaluate(context.Background(), UGCShowcaseUploadProtectionInput{
		Identity:    UGCShowcaseUploadProtectionIdentity{UserID: 42, IPAddress: "203.0.113.10"},
		UploadBytes: 1024,
	})
	if !errors.Is(err, ErrShowcaseUploadProtectionUnavailable) {
		t.Fatalf("Evaluate() error = %v, want ErrShowcaseUploadProtectionUnavailable", err)
	}
}

func newTestShowcaseUploadProtectionService(store showcaseUploadProtectionStore, overrides config.ShowcaseUploadProtectionConfig) *UGCShowcaseUploadProtectionService {
	cfg := config.ShowcaseUploadProtectionConfig{
		Enabled:                      true,
		WindowSeconds:                60,
		MaxUploadsPerUser:            100,
		MaxUploadsPerIP:              100,
		MaxUploadsPerIPPrefix:        100,
		DailyMaxUploadsPerUser:       100,
		DailyMaxUploadsPerIP:         100,
		DailyMaxBytesPerUser:         100 << 20,
		DailyMaxBytesPerIP:           100 << 20,
		MaxPendingSubmissionsPerUser: 10,
		FailureWindowSeconds:         900,
		MaxFailuresPerUser:           100,
		MaxFailuresPerIP:             100,
		BlockDurationSeconds:         1800,
	}
	applyShowcaseUploadProtectionConfigOverrides(&cfg, overrides)
	return &UGCShowcaseUploadProtectionService{
		store: store,
		cfg:   cfg,
		now: func() time.Time {
			return time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
		},
	}
}

func applyShowcaseUploadProtectionConfigOverrides(cfg *config.ShowcaseUploadProtectionConfig, overrides config.ShowcaseUploadProtectionConfig) {
	if overrides.WindowSeconds != 0 {
		cfg.WindowSeconds = overrides.WindowSeconds
	}
	if overrides.MaxUploadsPerUser != 0 {
		cfg.MaxUploadsPerUser = overrides.MaxUploadsPerUser
	}
	if overrides.MaxUploadsPerIP != 0 {
		cfg.MaxUploadsPerIP = overrides.MaxUploadsPerIP
	}
	if overrides.MaxUploadsPerIPPrefix != 0 {
		cfg.MaxUploadsPerIPPrefix = overrides.MaxUploadsPerIPPrefix
	}
	if overrides.DailyMaxUploadsPerUser != 0 {
		cfg.DailyMaxUploadsPerUser = overrides.DailyMaxUploadsPerUser
	}
	if overrides.DailyMaxUploadsPerIP != 0 {
		cfg.DailyMaxUploadsPerIP = overrides.DailyMaxUploadsPerIP
	}
	if overrides.DailyMaxBytesPerUser != 0 {
		cfg.DailyMaxBytesPerUser = overrides.DailyMaxBytesPerUser
	}
	if overrides.DailyMaxBytesPerIP != 0 {
		cfg.DailyMaxBytesPerIP = overrides.DailyMaxBytesPerIP
	}
	if overrides.MaxPendingSubmissionsPerUser != 0 {
		cfg.MaxPendingSubmissionsPerUser = overrides.MaxPendingSubmissionsPerUser
	}
	if overrides.FailureWindowSeconds != 0 {
		cfg.FailureWindowSeconds = overrides.FailureWindowSeconds
	}
	if overrides.MaxFailuresPerUser != 0 {
		cfg.MaxFailuresPerUser = overrides.MaxFailuresPerUser
	}
	if overrides.MaxFailuresPerIP != 0 {
		cfg.MaxFailuresPerIP = overrides.MaxFailuresPerIP
	}
	if overrides.BlockDurationSeconds != 0 {
		cfg.BlockDurationSeconds = overrides.BlockDurationSeconds
	}
}

type memoryShowcaseUploadProtectionStore struct {
	counts        map[string]int64
	blocked       map[string]bool
	evaluateCalls int
	evaluateKeys  []string
	failureKeys   []string
	err           error
}

func newMemoryShowcaseUploadProtectionStore() *memoryShowcaseUploadProtectionStore {
	return &memoryShowcaseUploadProtectionStore{
		counts:  map[string]int64{},
		blocked: map[string]bool{},
	}
}

func (s *memoryShowcaseUploadProtectionStore) Evaluate(_ context.Context, evaluation showcaseUploadProtectionEvaluation) (showcaseUploadProtectionStoreResult, error) {
	if s.err != nil {
		return showcaseUploadProtectionStoreResult{}, s.err
	}
	s.evaluateCalls++
	s.evaluateKeys = append(s.evaluateKeys, evaluation.Keys...)

	if s.blocked[evaluation.Keys[7]] {
		return showcaseUploadProtectionStoreResult{Blocked: true, Reason: "user_block", RetrySeconds: toShowcaseUploadProtectionTestInt64(evaluation.Args[10])}, nil
	}
	if s.blocked[evaluation.Keys[8]] {
		return showcaseUploadProtectionStoreResult{Blocked: true, Reason: "ip_block", RetrySeconds: toShowcaseUploadProtectionTestInt64(evaluation.Args[10])}, nil
	}

	windowSeconds := toShowcaseUploadProtectionTestInt64(evaluation.Args[0])
	dailySeconds := toShowcaseUploadProtectionTestInt64(evaluation.Args[9])
	checks := []struct {
		key          string
		increment    int64
		limit        int64
		reason       string
		retrySeconds int64
	}{
		{key: evaluation.Keys[0], increment: 1, limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[1]), reason: "user_window", retrySeconds: windowSeconds},
		{key: evaluation.Keys[1], increment: 1, limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[2]), reason: "ip_window", retrySeconds: windowSeconds},
		{key: evaluation.Keys[2], increment: 1, limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[3]), reason: "ip_prefix_window", retrySeconds: windowSeconds},
		{key: evaluation.Keys[3], increment: 1, limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[4]), reason: "user_daily", retrySeconds: dailySeconds},
		{key: evaluation.Keys[4], increment: 1, limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[5]), reason: "ip_daily", retrySeconds: dailySeconds},
		{key: evaluation.Keys[5], increment: toShowcaseUploadProtectionTestInt64(evaluation.Args[8]), limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[6]), reason: "user_daily_bytes", retrySeconds: dailySeconds},
		{key: evaluation.Keys[6], increment: toShowcaseUploadProtectionTestInt64(evaluation.Args[8]), limit: toShowcaseUploadProtectionTestInt64(evaluation.Args[7]), reason: "ip_daily_bytes", retrySeconds: dailySeconds},
	}
	for _, check := range checks {
		s.counts[check.key] += check.increment
		if s.counts[check.key] > check.limit {
			return showcaseUploadProtectionStoreResult{
				Blocked:      true,
				Reason:       check.reason,
				Count:        s.counts[check.key],
				Limit:        check.limit,
				RetrySeconds: check.retrySeconds,
			}, nil
		}
	}
	return showcaseUploadProtectionStoreResult{}, nil
}

func (s *memoryShowcaseUploadProtectionStore) RecordFailure(_ context.Context, failure showcaseUploadProtectionFailure) error {
	if s.err != nil {
		return s.err
	}
	s.failureKeys = append(s.failureKeys, failure.Keys...)

	userFailures := s.incrementFailure(failure.Keys[0])
	if userFailures >= toShowcaseUploadProtectionTestInt64(failure.Args[1]) {
		s.blocked[failure.Keys[2]] = true
	}
	ipFailures := s.incrementFailure(failure.Keys[1])
	if ipFailures >= toShowcaseUploadProtectionTestInt64(failure.Args[2]) {
		s.blocked[failure.Keys[3]] = true
	}
	return nil
}

func (s *memoryShowcaseUploadProtectionStore) incrementFailure(key string) int64 {
	if s.counts == nil {
		s.counts = map[string]int64{}
	}
	s.counts[key]++
	return s.counts[key]
}

func toShowcaseUploadProtectionTestInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}
