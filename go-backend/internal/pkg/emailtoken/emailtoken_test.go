package emailtoken

import (
	"testing"
	"time"
)

func TestSignAndVerify(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	token, err := Sign("test-email-token-secret", Claims{
		Purpose:   "subscription:confirm",
		Email:     "rider@example.test",
		Subject:   "rider@example.test",
		ExpiresAt: now.Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	claims, err := Verify("test-email-token-secret", token, "subscription:confirm", now)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if claims.Email != "rider@example.test" || claims.Subject != "rider@example.test" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestVerifyRejectsTamperingAndExpiry(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	token, err := Sign("test-email-token-secret", Claims{
		Purpose:   "subscription:confirm",
		Subject:   "rider@example.test",
		ExpiresAt: now.Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := Verify("test-email-token-secret", token+"x", "subscription:confirm", now); err == nil {
		t.Fatal("tampered token should be rejected")
	}

	expired, err := Sign("test-email-token-secret", Claims{
		Purpose:   "subscription:confirm",
		Subject:   "rider@example.test",
		ExpiresAt: now.Add(-time.Second).Unix(),
	})
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	if _, err := Verify("test-email-token-secret", expired, "subscription:confirm", now); err != ErrExpiredToken {
		t.Fatalf("verify expired token error = %v, want %v", err, ErrExpiredToken)
	}
}
