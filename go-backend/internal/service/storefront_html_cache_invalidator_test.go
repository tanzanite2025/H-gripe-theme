package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStorefrontHTMLCacheInvalidatorFromEnvRequiresURLAndToken(t *testing.T) {
	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_URL", "")
	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_TOKEN", "token")
	require.Nil(t, NewStorefrontHTMLCacheInvalidatorFromEnv())

	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_URL", "http://storefront:3000/_internal/html-cache/purge")
	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_TOKEN", "")
	require.Nil(t, NewStorefrontHTMLCacheInvalidatorFromEnv())

	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_URL", " http://storefront:3000/_internal/html-cache/purge ")
	t.Setenv("STOREFRONT_HTML_CACHE_PURGE_TOKEN", " secret-token ")

	invalidator := NewStorefrontHTMLCacheInvalidatorFromEnv()
	require.NotNil(t, invalidator)
	assert.True(t, invalidator.Enabled())
	assert.Equal(t, "http://storefront:3000/_internal/html-cache/purge", invalidator.purgeURL)
	assert.Equal(t, "secret-token", invalidator.token)
}

func TestStorefrontHTMLCacheInvalidatorPurgeAllSendsTokenAndReason(t *testing.T) {
	var receivedReason string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "secret-token", r.Header.Get(storefrontHTMLCachePurgeHeader))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		receivedReason = payload["reason"]

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	invalidator := &StorefrontHTMLCacheInvalidator{
		purgeURL: server.URL,
		token:    "secret-token",
		client:   server.Client(),
		timeout:  time.Second,
	}

	require.NoError(t, invalidator.PurgeAll(context.Background(), " admin product update "))
	assert.Equal(t, "admin product update", receivedReason)
}

func TestStorefrontHTMLCacheInvalidatorPurgeAllReturnsErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "purge denied", http.StatusUnauthorized)
	}))
	defer server.Close()

	invalidator := &StorefrontHTMLCacheInvalidator{
		purgeURL: server.URL,
		token:    "secret-token",
		client:   server.Client(),
		timeout:  time.Second,
	}

	err := invalidator.PurgeAll(context.Background(), "admin product update")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status=401")
	assert.Contains(t, err.Error(), "purge denied")
}

func TestStorefrontHTMLCacheInvalidatorPurgeAllNoopsWhenDisabled(t *testing.T) {
	require.NoError(t, (*StorefrontHTMLCacheInvalidator)(nil).PurgeAll(context.Background(), "noop"))

	invalidator := &StorefrontHTMLCacheInvalidator{
		purgeURL: strings.TrimSpace(""),
		token:    "secret-token",
		client:   http.DefaultClient,
		timeout:  time.Second,
	}
	require.NoError(t, invalidator.PurgeAll(context.Background(), "noop"))
}

func TestStorefrontHTMLCacheInvalidatorPurgeAllAsyncCoalescesWithinDebounceWindow(t *testing.T) {
	var requestCount atomic.Int32
	reasons := make(chan string, 3)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		var payload map[string]string
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		reasons <- payload["reason"]

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	invalidator := &StorefrontHTMLCacheInvalidator{
		purgeURL: server.URL,
		token:    "secret-token",
		client:   server.Client(),
		timeout:  time.Second,
		debounce: 20 * time.Millisecond,
	}

	invalidator.PurgeAllAsync("first")
	invalidator.PurgeAllAsync("second")
	invalidator.PurgeAllAsync("third")

	select {
	case reason := <-reasons:
		assert.Equal(t, "third", reason)
	case <-time.After(time.Second):
		t.Fatal("expected debounced purge request")
	}

	time.Sleep(80 * time.Millisecond)
	assert.Equal(t, int32(1), requestCount.Load())
}
