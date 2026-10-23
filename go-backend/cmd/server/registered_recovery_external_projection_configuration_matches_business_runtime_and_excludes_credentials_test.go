package main

import (
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestRegisteredRecoveryExternalProjectionConfigurationMatchesRuntimeAndNeverSerializesCredentials(t *testing.T) {
	t.Setenv("CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL", "")
	t.Setenv("CUSTOMER_SERVICE_SEARCH_INDEX_TOKEN", "synthetic-search-token")
	cfg := &config.Config{Redis: config.RedisConfig{Host: "synthetic-redis", Port: 6379, DB: 3, Password: "synthetic-redis-password"}}
	shared := &storage.Config{BaseURL: "https://media.example.test", PrivateBaseURL: "https://private.example.test"}
	logo := &storage.Config{BaseURL: "https://logo.example.test"}
	boundary, err := collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
	require.NoError(t, err)
	require.Equal(t, "sql_only", boundary.CustomerServiceSearchMode)
	require.Equal(t, "process_local", boundary.CustomerServiceRealtimeMode)
	t.Setenv("CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL", "https://search.example.test/delete/{ticket_id}?key=synthetic-query-secret")
	cfg.CustomerServiceRealtime.Enabled = true
	cfg.CustomerServiceRealtime.Stream = "customer_service:{synthetic-realtime}:v1"
	boundary, err = collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
	require.NoError(t, err)
	require.Equal(t, "external_delete_only", boundary.CustomerServiceSearchMode)
	require.Equal(t, maintenancegate.FingerprintExternalProjectionConfiguration("customer_service:{synthetic-realtime}:dedupe:"), boundary.RealtimeDedupeNamespaceSHA256)
	encoded, err := json.Marshal(boundary)
	require.NoError(t, err)
	for _, secret := range []string{"synthetic-redis-password", "synthetic-search-token", "synthetic-query-secret", "search.example.test", "synthetic-redis", "synthetic-realtime", "media.example.test"} {
		require.NotContains(t, string(encoded), secret)
	}
	cfg.CustomerServiceRealtime.Stream = "customer_service:{synthetic-realtime}:v2"
	changed, err := collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
	require.NoError(t, err)
	require.NotEqual(t, boundary.RealtimeStreamSHA256, changed.RealtimeStreamSHA256)
	require.Equal(t, boundary.RealtimeDedupeNamespaceSHA256, changed.RealtimeDedupeNamespaceSHA256, "suffix-only rotation must not claim isolated dedupe")
	cfg.CustomerServiceRealtime.Stream = "customer_service:{other-realtime}:v2"
	changed, err = collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
	require.NoError(t, err)
	require.NotEqual(t, boundary.RealtimeDedupeNamespaceSHA256, changed.RealtimeDedupeNamespaceSHA256)
	cfg.Redis.DB++
	changed, err = collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
	require.NoError(t, err)
	require.NotEqual(t, boundary.RedisDestinationSHA256, changed.RedisDestinationSHA256)
	for _, endpoint := range []string{"ftp://search.example.test", "https://user:secret@search.example.test", "https://search.example.test/#fragment"} {
		t.Setenv("CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL", endpoint)
		_, err = collectRegisteredRecoveryExternalProjectionConfiguration(cfg, shared, logo)
		require.Error(t, err)
		require.False(t, strings.Contains(err.Error(), endpoint))
	}
}
