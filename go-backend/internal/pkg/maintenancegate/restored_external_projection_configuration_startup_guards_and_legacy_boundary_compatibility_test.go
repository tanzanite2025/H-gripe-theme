package maintenancegate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestoredExternalProjectionBoundaryRejectsConfigurationDriftAndOmissionButAllowsBuildUpgrade(t *testing.T) {
	directory := t.TempDir()
	roots := map[string]string{"uploads": t.TempDir()}
	keys := map[string]string{"JWT_SECRET": "synthetic-key"}
	boundary := ExternalProjectionBoundary{Version: 1, CustomerServiceSearchMode: "sql_only", RedisDestinationSHA256: strings.Repeat("a", 64), CustomerServiceRealtimeMode: "process_local", MediaDeliveryConfigurationSHA256: strings.Repeat("b", 64)}
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-before", roots, keys, boundary))
	data, err := os.ReadFile(filepath.Join(directory, "application-boundary.json"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "restored-data-identity-boundary.json"), data, 0600))
	var registeredBoundary ApplicationBoundary
	require.NoError(t, json.Unmarshal(data, &registeredBoundary))
	require.Equal(t, 2, registeredBoundary.Protocol, "older binaries require protocol 1 and must reject this guard")
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, boundary))
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys))
	changed := boundary
	changed.RedisDestinationSHA256 = strings.Repeat("c", 64)
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, changed))
	changed = boundary
	changed.MediaDeliveryConfigurationSHA256 = strings.Repeat("c", 64)
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, changed))
	changed = boundary
	changed.CustomerServiceSearchMode = "external_delete_only"
	changed.SearchDestinationSHA256 = strings.Repeat("c", 64)
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, changed))
	changed = boundary
	changed.CustomerServiceRealtimeMode = "redis_stream"
	changed.RealtimeStreamSHA256 = strings.Repeat("c", 64)
	changed.RealtimeDedupeNamespaceSHA256 = strings.Repeat("d", 64)
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, changed))
	// Legacy guards can be upgraded; legacy reports still cannot release a new
	// recovery until explicit external configuration and evidence are supplied.
	var legacy ApplicationBoundary
	require.NoError(t, json.Unmarshal(data, &legacy))
	legacy.ExternalProjections = nil
	legacy.Protocol = 1
	legacyData, err := json.Marshal(legacy)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(directory, "restored-data-identity-boundary.json"), legacyData, 0600))
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-after", roots, keys, boundary))
}

func TestExternalProjectionBoundaryRejectsUnknownOrContradictoryConfiguration(t *testing.T) {
	valid := ExternalProjectionBoundary{Version: 1, CustomerServiceSearchMode: "sql_only", RedisDestinationSHA256: strings.Repeat("a", 64), CustomerServiceRealtimeMode: "process_local", MediaDeliveryConfigurationSHA256: strings.Repeat("b", 64)}
	for _, scenario := range []struct {
		name   string
		change func(*ExternalProjectionBoundary)
	}{
		{"unknown_version", func(b *ExternalProjectionBoundary) { b.Version = 2 }},
		{"missing_redis", func(b *ExternalProjectionBoundary) { b.RedisDestinationSHA256 = "" }},
		{"missing_media", func(b *ExternalProjectionBoundary) { b.MediaDeliveryConfigurationSHA256 = "" }},
		{"unknown_search", func(b *ExternalProjectionBoundary) { b.CustomerServiceSearchMode = "" }},
		{"sql_with_external_destination", func(b *ExternalProjectionBoundary) { b.SearchDestinationSHA256 = strings.Repeat("c", 64) }},
		{"external_without_destination", func(b *ExternalProjectionBoundary) { b.CustomerServiceSearchMode = "external_delete_only" }},
		{"unknown_realtime", func(b *ExternalProjectionBoundary) { b.CustomerServiceRealtimeMode = "" }},
		{"local_with_stream", func(b *ExternalProjectionBoundary) { b.RealtimeStreamSHA256 = strings.Repeat("c", 64) }},
		{"stream_without_dedupe", func(b *ExternalProjectionBoundary) {
			b.CustomerServiceRealtimeMode = "redis_stream"
			b.RealtimeStreamSHA256 = strings.Repeat("c", 64)
		}},
	} {
		t.Run(scenario.name, func(t *testing.T) { changed := valid; scenario.change(&changed); require.Error(t, changed.Validate()) })
	}
}
