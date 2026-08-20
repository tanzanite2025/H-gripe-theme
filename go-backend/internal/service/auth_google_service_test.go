package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/resilience"
)

func TestVerifyGoogleIDTokenRetriesTransientFailures(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("id_token") != "test-token" {
			t.Fatalf("id_token = %q, want test-token", request.URL.Query().Get("id_token"))
		}
		if atomic.AddInt32(&requests, 1) < 3 {
			http.Error(writer, "temporary failure", http.StatusServiceUnavailable)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(googleTokenInfo{
			Issuer:        "https://accounts.google.com",
			Subject:       "google-subject",
			Audience:      "google-client-id",
			Email:         "buyer@example.com",
			EmailVerified: "true",
		})
	}))
	defer server.Close()

	authService := NewAuthService(nil, config.JWTConfig{}, config.OAuthConfig{
		GoogleClientID: "google-client-id",
	})
	authService.ConfigureGoogleTokenInfoEndpoint(server.URL)
	authService.ConfigureGoogleOAuthResilience(
		resilience.HTTPRetryPolicy{
			MaxAttempts: 3,
			Backoff: resilience.BackoffPolicy{
				BaseDelay: time.Millisecond,
				MaxDelay:  time.Millisecond,
			},
		},
		resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
			FailureThreshold: 5,
			FailureWindow:    time.Minute,
			OpenDuration:     time.Minute,
		}),
	)

	tokenInfo, err := authService.verifyGoogleIDToken(nil, "test-token")
	if err != nil {
		t.Fatalf("verifyGoogleIDToken returned error: %v", err)
	}
	if tokenInfo.Email != "buyer@example.com" {
		t.Fatalf("email = %q, want buyer@example.com", tokenInfo.Email)
	}
	if got := atomic.LoadInt32(&requests); got != 3 {
		t.Fatalf("request count = %d, want 3", got)
	}
}

func TestVerifyGoogleIDTokenCircuitBreakerStopsRepeatedFailures(t *testing.T) {
	var requests int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		atomic.AddInt32(&requests, 1)
		http.Error(writer, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	authService := NewAuthService(nil, config.JWTConfig{}, config.OAuthConfig{
		GoogleClientID: "google-client-id",
	})
	authService.ConfigureGoogleTokenInfoEndpoint(server.URL)
	authService.ConfigureGoogleOAuthResilience(
		resilience.HTTPRetryPolicy{
			MaxAttempts: 1,
			Backoff: resilience.BackoffPolicy{
				BaseDelay: time.Millisecond,
				MaxDelay:  time.Millisecond,
			},
		},
		resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
			FailureThreshold: 1,
			FailureWindow:    time.Minute,
			OpenDuration:     time.Minute,
		}),
	)

	_, firstErr := authService.verifyGoogleIDToken(context.Background(), "test-token")
	if firstErr == nil {
		t.Fatal("first verification error = nil, want upstream failure")
	}

	_, secondErr := authService.verifyGoogleIDToken(context.Background(), "test-token")
	if !errors.Is(secondErr, resilience.ErrCircuitOpen) {
		t.Fatalf("second verification error = %v, want ErrCircuitOpen", secondErr)
	}
	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Fatalf("request count = %d, want 1 after circuit opens", got)
	}
}
