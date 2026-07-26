package requestsign

import (
	"strconv"
	"testing"
	"time"
)

func TestVerifyAcceptsCanonicalRequest(t *testing.T) {
	now := time.Unix(1_735_689_600, 0)
	key := "test-request-signing-key-with-at-least-32-bytes"
	body := []byte(`{"amount":12.50}`)
	nonce := "nonce-1"
	signature := Sign("post", "/api/v1/orders?preview=true", now, nonce, body, key)

	if err := Verify(
		"POST",
		"/api/v1/orders?preview=true",
		strconv.FormatInt(now.Unix(), 10),
		nonce,
		signature,
		body,
		key,
		now,
		30*time.Second,
	); err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
}

func TestVerifyRejectsStaleTimestamp(t *testing.T) {
	now := time.Unix(1_735_689_600, 0)
	key := "test-request-signing-key-with-at-least-32-bytes"
	signature := Sign("POST", "/api/v1/orders", now.Add(-31*time.Second), "nonce-1", nil, key)

	if err := Verify(
		"POST",
		"/api/v1/orders",
		strconv.FormatInt(now.Add(-31*time.Second).Unix(), 10),
		"nonce-1",
		signature,
		nil,
		key,
		now,
		30*time.Second,
	); err == nil {
		t.Fatal("Verify should reject a stale timestamp")
	}
}

func TestVerifyRejectsModifiedBodyAndInvalidNonce(t *testing.T) {
	now := time.Unix(1_735_689_600, 0)
	key := "test-request-signing-key-with-at-least-32-bytes"
	signature := Sign("POST", "/api/v1/orders", now, "nonce-1", []byte("original"), key)

	if err := Verify(
		"POST",
		"/api/v1/orders",
		strconv.FormatInt(now.Unix(), 10),
		"nonce-1",
		signature,
		[]byte("modified"),
		key,
		now,
		30*time.Second,
	); err == nil {
		t.Fatal("Verify should reject a modified body")
	}

	if err := Verify(
		"POST",
		"/api/v1/orders",
		strconv.FormatInt(now.Unix(), 10),
		"",
		signature,
		[]byte("original"),
		key,
		now,
		30*time.Second,
	); err == nil {
		t.Fatal("Verify should reject an empty nonce")
	}
}

func TestNormalizeTargetKeepsPathAndQueryOnly(t *testing.T) {
	got := normalizeTarget("https://example.test/api/v1/orders?preview=true")
	if got != "/api/v1/orders?preview=true" {
		t.Fatalf("normalizeTarget = %q", got)
	}
}
