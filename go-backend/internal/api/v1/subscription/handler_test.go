package subscription

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	domainsubscription "tanzanite/internal/domain/subscription"
	"testing"

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

func TestUnsubscribeByEmailDoesNotMutateDirectly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil)
	router := gin.New()
	router.POST("/subscriptions/unsubscribe", handler.UnsubscribeByEmail)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/subscriptions/unsubscribe", strings.NewReader(`{"email":"victim@example.test"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("unsubscribe by email status = %d, want %d", recorder.Code, http.StatusAccepted)
	}
}
