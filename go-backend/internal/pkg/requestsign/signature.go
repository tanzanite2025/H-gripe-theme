package requestsign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Sign(method, requestTarget string, timestamp time.Time, nonce string, body []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(canonical(method, requestTarget, timestamp.Unix(), nonce, body)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func Verify(method, requestTarget, timestampValue, nonce, signature string, body []byte, key string, now time.Time, maxSkew time.Duration) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("request signing key is not configured")
	}
	timestamp, err := strconv.ParseInt(strings.TrimSpace(timestampValue), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid request timestamp")
	}
	if abs(now.Unix()-timestamp) > int64(maxSkew/time.Second) {
		return fmt.Errorf("request timestamp is outside the allowed window")
	}
	if strings.TrimSpace(nonce) == "" || len(nonce) > 128 {
		return fmt.Errorf("invalid request nonce")
	}

	expected := Sign(method, requestTarget, time.Unix(timestamp, 0), nonce, body, key)
	actual, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return fmt.Errorf("invalid request signature encoding")
	}
	expectedBytes, err := base64.RawURLEncoding.DecodeString(expected)
	if err != nil || !hmac.Equal(expectedBytes, actual) {
		return fmt.Errorf("invalid request signature")
	}
	return nil
}

func BodyDigest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func canonical(method, requestTarget string, timestamp int64, nonce string, body []byte) string {
	return strings.Join([]string{
		strings.ToUpper(strings.TrimSpace(method)),
		normalizeTarget(requestTarget),
		strconv.FormatInt(timestamp, 10),
		strings.TrimSpace(nonce),
		BodyDigest(body),
	}, "\n")
}

func normalizeTarget(requestTarget string) string {
	parsed, err := url.Parse(requestTarget)
	if err != nil || parsed.Path == "" {
		return requestTarget
	}
	target := parsed.EscapedPath()
	if parsed.RawQuery != "" {
		target += "?" + parsed.RawQuery
	}
	return target
}

func abs(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
