package subscription

import (
	domainsubscription "commerce-platform/internal/domain/subscription"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/honeypot"

	"github.com/gin-gonic/gin"
)

func TestPublicSubscriptionResponseOmitsUnsubscribeToken(t *testing.T) {
	body, err := json.Marshal(publicSubscriptionResponse(domainsubscription.Subscription{
		ID:         1,
		Email:      "customer@example.test",
		Status:     "active",
		Locale:     "en",
		UnsubToken: "unsubscribe-secret-token",
	}))
	if err != nil {
		t.Fatalf("marshal subscription response: %v", err)
	}

	payload := string(body)
	if strings.Contains(payload, "unsub") || strings.Contains(payload, "unsubscribe-secret-token") {
		t.Fatalf("subscription response leaked unsubscribe token: %s", payload)
	}
}

func TestIssueTimingTokenReturnsServerSignedNewsletterToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fixedNow := time.Date(2026, 9, 16, 14, 30, 0, 250000000, time.UTC)
	handler := NewHandler(nil)
	handler.ConfigureTimingToken("server-only-secret", 10*time.Minute)
	handler.now = func() time.Time { return fixedNow }

	router := gin.New()
	router.GET("/subscriptions/timing-token", handler.IssueTimingToken)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/subscriptions/timing-token", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("timing token status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
	var payload struct {
		Token     string `json:"token"`
		ExpiresIn int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode timing token response: %v", err)
	}
	if payload.ExpiresIn != 600 {
		t.Fatalf("expires_in = %d, want 600", payload.ExpiresIn)
	}
	claims, err := honeypot.VerifyTimingToken(payload.Token, "newsletter", "server-only-secret", fixedNow, 10*time.Minute)
	if err != nil {
		t.Fatalf("issued token did not verify: %v", err)
	}
	if !claims.IssuedAt.Equal(fixedNow) {
		t.Fatalf("issued_at = %v, want %v", claims.IssuedAt, fixedNow)
	}
}

func TestUnsubscribeByEmailDoesNotMutateDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil)
	router := gin.New()
	router.POST("/subscriptions/unsubscribe", handler.UnsubscribeByEmail)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/subscriptions/unsubscribe", strings.NewReader(`{"email":"victim@example.test"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unsubscribe by email status = %d, want %d", recorder.Code, http.StatusAccepted)
	}
}
