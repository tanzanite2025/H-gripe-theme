package honeypot

import (
	"strings"
	"testing"
	"time"
)

func TestTimingTokenRoundTrip(t *testing.T) {
	now := time.Date(2026, 9, 16, 12, 0, 0, 123000000, time.UTC)
	token, err := IssueTimingToken("newsletter", "timing-secret", now)
	if err != nil {
		t.Fatalf("IssueTimingToken() error = %v", err)
	}
	claims, err := VerifyTimingToken(token, "NEWSLETTER", "timing-secret", now.Add(2*time.Second), time.Minute)
	if err != nil {
		t.Fatalf("VerifyTimingToken() error = %v", err)
	}
	if claims.Form != "newsletter" || claims.Nonce == "" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
	if !claims.IssuedAt.Equal(now) {
		t.Fatalf("issued time = %v, want %v", claims.IssuedAt, now)
	}
}

func TestDeriveTimingKeyIsDeterministicAndDomainSeparated(t *testing.T) {
	first := DeriveTimingKey("master-secret")
	second := DeriveTimingKey("master-secret")
	if first == "" || first != second || first == "master-secret" {
		t.Fatalf("unexpected derived timing key %q", first)
	}
	if got := DeriveTimingKey(" "); got != "" {
		t.Fatalf("empty master secret derived key = %q", got)
	}
}

func TestTimingTokenRejectsTamperingAndWrongForm(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	token, err := IssueTimingToken("newsletter", "timing-secret", now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	parts[1] = parts[1] + "a"
	if _, err := VerifyTimingToken(strings.Join(parts, "."), "newsletter", "timing-secret", now, time.Minute); err == nil {
		t.Fatal("tampered token should be rejected")
	}
	if _, err := VerifyTimingToken(token, "feedback", "timing-secret", now, time.Minute); err != ErrTimingTokenInvalid {
		t.Fatalf("wrong form error = %v, want ErrTimingTokenInvalid", err)
	}
}

func TestTimingTokenRejectsExpiredAndFarFuture(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	token, err := IssueTimingToken("newsletter", "timing-secret", now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyTimingToken(token, "newsletter", "timing-secret", now.Add(time.Minute+time.Millisecond), time.Minute); err != ErrTimingTokenExpired {
		t.Fatalf("expired token error = %v, want ErrTimingTokenExpired", err)
	}
	futureToken, err := IssueTimingToken("newsletter", "timing-secret", now.Add(6*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := VerifyTimingToken(futureToken, "newsletter", "timing-secret", now, time.Minute); err != ErrTimingTokenInvalid {
		t.Fatalf("future token error = %v, want ErrTimingTokenInvalid", err)
	}
}

func TestObserveTimingTokenIsShadowOnly(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	token, err := IssueTimingToken("newsletter", "timing-secret", now)
	if err != nil {
		t.Fatal(err)
	}
	elapsed, ok := ObserveTimingToken(token, "newsletter", "timing-secret", now.Add(2*time.Second), time.Minute)
	if !ok || elapsed != 2*time.Second {
		t.Fatalf("ObserveTimingToken() = (%v, %v), want (2s, true)", elapsed, ok)
	}
	if _, ok := ObserveTimingToken("not-a-token", "newsletter", "timing-secret", now, time.Minute); ok {
		t.Fatal("malformed token should be observed as invalid, not accepted")
	}
}
