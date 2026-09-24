package honeypot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"commerce-platform/internal/pkg/metrics"

	"github.com/redis/go-redis/v9"
)

const timingReplayTimeout = 25 * time.Millisecond

// TimingReplayObserver records whether a valid timing nonce has already been
// seen. It is intentionally observation-only: reused tokens are not rejected
// until a separately reviewed enforce rollout enables that policy.
type TimingReplayObserver struct {
	redisClient redis.UniversalClient
	timeout     time.Duration
}

func NewTimingReplayObserver(redisClient redis.UniversalClient) *TimingReplayObserver {
	if redisClient == nil {
		return nil
	}
	return &TimingReplayObserver{
		redisClient: redisClient,
		timeout:     timingReplayTimeout,
	}
}

// Observe stores only a hash of the nonce and expires it with the token's
// remaining lifetime. Redis failures are fail-open for the form submission;
// they are surfaced as a bounded metric result instead.
func (o *TimingReplayObserver) Observe(ctx context.Context, form string, claims TimingClaims, now time.Time, ttl time.Duration) string {
	if o == nil || o.redisClient == nil {
		return TimingReplayError
	}
	if ttl <= 0 {
		ttl = DefaultTimingTokenTTL()
	}
	remaining := claims.IssuedAt.Add(ttl).Sub(now)
	if remaining <= 0 || claims.Nonce == "" {
		metrics.HoneypotTimingReplays.WithLabelValues(normalizeTimingForm(form), TimingReplayError).Inc()
		return TimingReplayError
	}

	keyDigest := sha256.Sum256([]byte(normalizeTimingForm(form) + "\x00" + claims.Nonce))
	key := "commerce_platform:honeypot-timing:nonce:" + hex.EncodeToString(keyDigest[:])
	if ctx == nil {
		ctx = context.Background()
	}
	checkContext, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()
	accepted, err := o.redisClient.SetNX(checkContext, key, "1", remaining).Result()
	if err != nil {
		metrics.HoneypotTimingReplays.WithLabelValues(normalizeTimingForm(form), TimingReplayError).Inc()
		return TimingReplayError
	}
	if !accepted {
		metrics.HoneypotTimingReplays.WithLabelValues(normalizeTimingForm(form), TimingReplayReused).Inc()
		return TimingReplayReused
	}
	metrics.HoneypotTimingReplays.WithLabelValues(normalizeTimingForm(form), TimingReplayFirstSeen).Inc()
	return TimingReplayFirstSeen
}
