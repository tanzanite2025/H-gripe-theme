package honeypot

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/pkg/metrics"
)

const (
	timingTokenVersion       = "v1"
	timingTokenNonceBytes    = 16
	timingKeyContext         = "commerce-platform/honeypot-timing/v1"
	defaultTimingTokenTTL    = 15 * time.Minute
	maxTimingTokenFutureSkew = 5 * time.Second
)

var (
	ErrTimingTokenMalformed = errors.New("malformed honeypot timing token")
	ErrTimingTokenInvalid   = errors.New("invalid honeypot timing token")
	ErrTimingTokenExpired   = errors.New("expired honeypot timing token")
)

const (
	TimingResultValid     = "valid"
	TimingResultMissing   = "missing"
	TimingResultMalformed = "malformed"
	TimingResultInvalid   = "invalid"
	TimingResultExpired   = "expired"
	TimingReplayFirstSeen = "first_seen"
	TimingReplayReused    = "reused"
	TimingReplayError     = "error"
)

// TimingClaims are the server-issued values used to measure the time between
// rendering a form and submitting it. IssuedAt is never trusted unless the
// token's HMAC verifies successfully.
type TimingClaims struct {
	Form     string
	IssuedAt time.Time
	Nonce    string
}

// TimingObservation is the non-blocking result recorded for one submission.
// Claims are returned only after a valid HMAC and time-window check.
type TimingObservation struct {
	Result  string
	Claims  TimingClaims
	Elapsed time.Duration
	Valid   bool
}

type timingPayload struct {
	Version  string `json:"v"`
	Form     string `json:"form"`
	IssuedAt int64  `json:"iat"` // Unix milliseconds
	Nonce    string `json:"nonce"`
}

// DefaultTimingTokenTTL is used when a deployment leaves the optional TTL
// setting unset.
func DefaultTimingTokenTTL() time.Duration {
	return defaultTimingTokenTTL
}

// DeriveTimingKey provides key separation when a deployment chooses to use
// an existing server-side master secret instead of a dedicated timing key.
func DeriveTimingKey(masterSecret string) string {
	masterSecret = strings.TrimSpace(masterSecret)
	if masterSecret == "" {
		return ""
	}
	derived := timingSignature(timingKeyContext, masterSecret)
	return base64.RawURLEncoding.EncodeToString(derived)
}

// IssueTimingToken creates a short-lived, server-signed token. The key must
// stay server-side; the token itself is safe to send to a browser.
func IssueTimingToken(form, key string, now time.Time) (string, error) {
	form = normalizeTimingForm(form)
	if form == "" || strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("form and timing token key are required")
	}

	nonceBytes := make([]byte, timingTokenNonceBytes)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("generate timing token nonce: %w", err)
	}
	payload := timingPayload{
		Version:  timingTokenVersion,
		Form:     form,
		IssuedAt: now.UTC().UnixMilli(),
		Nonce:    base64.RawURLEncoding.EncodeToString(nonceBytes),
	}
	return encodeTimingToken(payload, key)
}

// VerifyTimingToken validates the signature, form binding, clock bounds, and
// expiry. A caller that needs replay prevention can persist claims.Nonce in a
// short-lived store after this function returns successfully.
func VerifyTimingToken(token, expectedForm, key string, now time.Time, ttl time.Duration) (TimingClaims, error) {
	var zero TimingClaims
	if strings.TrimSpace(key) == "" || normalizeTimingForm(expectedForm) == "" {
		return zero, ErrTimingTokenInvalid
	}
	if ttl <= 0 {
		ttl = defaultTimingTokenTTL
	}

	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 || parts[0] != timingTokenVersion {
		return zero, ErrTimingTokenMalformed
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || len(payloadBytes) == 0 {
		return zero, ErrTimingTokenMalformed
	}
	actualSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return zero, ErrTimingTokenMalformed
	}
	expectedSignature := timingSignature(parts[0]+"."+parts[1], key)
	if !hmac.Equal(expectedSignature, actualSignature) {
		return zero, ErrTimingTokenInvalid
	}

	var payload timingPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return zero, ErrTimingTokenMalformed
	}
	if payload.Version != timingTokenVersion || payload.Form != normalizeTimingForm(expectedForm) ||
		payload.IssuedAt <= 0 || strings.TrimSpace(payload.Nonce) == "" || len(payload.Nonce) > 128 {
		return zero, ErrTimingTokenInvalid
	}
	issuedAt := time.UnixMilli(payload.IssuedAt)
	age := now.Sub(issuedAt)
	if age < -maxTimingTokenFutureSkew {
		return zero, ErrTimingTokenInvalid
	}
	if age > ttl {
		return zero, ErrTimingTokenExpired
	}

	return TimingClaims{Form: payload.Form, IssuedAt: issuedAt, Nonce: payload.Nonce}, nil
}

// ObserveTimingToken records a bounded validation result for every submission
// and a duration histogram only for authentic tokens. It deliberately returns
// no blocking decision: timing is a shadow-mode signal until real traffic has
// established a safe threshold and false-positive rate.
func ObserveTimingToken(token, expectedForm, key string, now time.Time, ttl time.Duration) (time.Duration, bool) {
	observation := ObserveTimingTokenDetailed(token, expectedForm, key, now, ttl)
	return observation.Elapsed, observation.Valid
}

// ObserveTimingTokenDetailed records a bounded validation result for every
// submission and a duration histogram only for authentic tokens. It
// deliberately returns no blocking decision: timing is a shadow-mode signal
// until real traffic has established a safe threshold and false-positive rate.
func ObserveTimingTokenDetailed(token, expectedForm, key string, now time.Time, ttl time.Duration) TimingObservation {
	form := normalizeTimingForm(expectedForm)
	if strings.TrimSpace(token) == "" {
		metrics.HoneypotTimingEvaluations.WithLabelValues(form, TimingResultMissing).Inc()
		return TimingObservation{Result: TimingResultMissing}
	}

	claims, err := VerifyTimingToken(token, form, key, now, ttl)
	if err != nil {
		result := TimingResultInvalid
		switch {
		case errors.Is(err, ErrTimingTokenMalformed):
			result = TimingResultMalformed
		case errors.Is(err, ErrTimingTokenExpired):
			result = TimingResultExpired
		}
		metrics.HoneypotTimingEvaluations.WithLabelValues(form, result).Inc()
		return TimingObservation{Result: result}
	}

	elapsed := now.Sub(claims.IssuedAt)
	if elapsed < 0 {
		elapsed = 0
	}
	metrics.HoneypotTimingEvaluations.WithLabelValues(form, TimingResultValid).Inc()
	metrics.HoneypotTimingSeconds.WithLabelValues(form).Observe(elapsed.Seconds())
	return TimingObservation{
		Result:  TimingResultValid,
		Claims:  claims,
		Elapsed: elapsed,
		Valid:   true,
	}
}

func encodeTimingToken(payload timingPayload, key string) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode timing token: %w", err)
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := timingSignature(timingTokenVersion+"."+payloadPart, key)
	return timingTokenVersion + "." + payloadPart + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func timingSignature(message, key string) []byte {
	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(key)))
	_, _ = mac.Write([]byte(message))
	return mac.Sum(nil)
}

func normalizeTimingForm(form string) string {
	return strings.ToLower(strings.TrimSpace(form))
}
