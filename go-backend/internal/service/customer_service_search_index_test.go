package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"commerce-platform/internal/pkg/resilience"
	"github.com/stretchr/testify/require"
)

func TestCustomerServiceHTTPSearchIndexDeletesConversationWithIdempotencyHeaders(t *testing.T) {
	var method, path, authorization, idempotencyKey string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		method = request.Method
		path = request.URL.Path
		authorization = request.Header.Get("Authorization")
		idempotencyKey = request.Header.Get("Idempotency-Key")
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	index, err := NewCustomerServiceHTTPSearchIndex(
		server.URL+"/conversations/{ticket_id}",
		"search-token",
		server.Client(),
		resilience.HTTPRetryPolicy{MaxAttempts: 1},
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, index.DeleteConversation(context.Background(), 42))
	require.Equal(t, http.MethodDelete, method)
	require.Equal(t, "/conversations/42", path)
	require.Equal(t, "Bearer search-token", authorization)
	require.Equal(t, "customer-service-conversation:42", idempotencyKey)
}

func TestCustomerServiceHTTPSearchIndexTreatsMissingProjectionAsSuccess(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	index, err := NewCustomerServiceHTTPSearchIndex(
		server.URL+"/conversations",
		"",
		server.Client(),
		resilience.HTTPRetryPolicy{MaxAttempts: 1},
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, index.DeleteConversation(nil, 9))
}

func TestCustomerServiceHTTPSearchIndexRetriesTransientDeleteFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	index, err := NewCustomerServiceHTTPSearchIndex(
		server.URL+"/conversations/{ticket_id}",
		"",
		server.Client(),
		resilience.HTTPRetryPolicy{
			MaxAttempts: 2,
			Backoff:     resilience.BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
		},
		nil,
	)
	require.NoError(t, err)
	require.NoError(t, index.DeleteConversation(context.Background(), 11))
	require.Equal(t, int32(2), requests.Load())
}

func TestNewCustomerServiceHTTPSearchIndexRejectsUnsafeEndpoint(t *testing.T) {
	for _, endpoint := range []string{
		"file:///tmp/search-index",
		"https://user:pass@example.test/conversations",
		"https://example.test/conversations#fragment",
	} {
		index, err := NewCustomerServiceHTTPSearchIndex(endpoint, "", nil, resilience.HTTPRetryPolicy{MaxAttempts: 1}, nil)
		require.Nil(t, index)
		require.Error(t, err)
		require.NotContains(t, strings.ToLower(err.Error()), "panic")
	}
}
