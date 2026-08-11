package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorefrontContentReleaseNotifierFromEnvRequiresURLAndToken(t *testing.T) {
	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_URL", "")
	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_TOKEN", "token")
	require.Nil(t, NewStorefrontContentReleaseNotifierFromEnv())

	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_URL", "http://storefront-rebuild")
	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_TOKEN", "")
	require.Nil(t, NewStorefrontContentReleaseNotifierFromEnv())

	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_URL", " http://storefront-rebuild ")
	t.Setenv("STOREFRONT_CONTENT_RELEASE_WEBHOOK_TOKEN", " secret-token ")

	notifier := NewStorefrontContentReleaseNotifierFromEnv()
	require.NotNil(t, notifier)
	assert.True(t, notifier.Enabled())
	assert.Equal(t, "http://storefront-rebuild", notifier.webhookURL)
	assert.Equal(t, "secret-token", notifier.token)
}

func TestStorefrontContentReleaseNotifierTriggerSendsTokenAndPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "secret-token", r.Header.Get(storefrontContentReleaseTokenHeader))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "storefront_content_published", payload["event"])
		assert.Equal(t, "admin faq update", payload["reason"])
		assert.Equal(t, "commerce_platform-go-backend", payload["source"])

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	notifier := &StorefrontContentReleaseNotifier{
		webhookURL: server.URL,
		token:      "secret-token",
		client:     server.Client(),
		timeout:    time.Second,
	}

	require.NoError(t, notifier.Trigger(context.Background(), " admin faq update "))
}

func TestStorefrontContentReleaseNotifierTriggerReturnsErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "release denied", http.StatusUnauthorized)
	}))
	defer server.Close()

	notifier := &StorefrontContentReleaseNotifier{
		webhookURL: server.URL,
		token:      "secret-token",
		client:     server.Client(),
		timeout:    time.Second,
	}

	err := notifier.Trigger(context.Background(), "admin faq update")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status=401")
	assert.Contains(t, err.Error(), "release denied")
}
