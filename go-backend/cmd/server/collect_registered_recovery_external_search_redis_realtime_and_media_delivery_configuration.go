package main

import (
	"errors"
	"net/url"
	"os"
	"strings"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/service"
)

func collectRegisteredRecoveryExternalProjectionConfiguration(cfg *config.Config, sharedStorage, logoStorage *storage.Config) (maintenancegate.ExternalProjectionBoundary, error) {
	fingerprint := maintenancegate.FingerprintExternalProjectionConfiguration
	boundary := maintenancegate.ExternalProjectionBoundary{
		Version:                          1,
		CustomerServiceSearchMode:        "sql_only",
		CustomerServiceRealtimeMode:      "process_local",
		RedisDestinationSHA256:           maintenancegate.FingerprintRedisRecoveryDestination(cfg.Redis.NormalizedMode(), cfg.Redis.GetRedisAddrs(), cfg.Redis.DB, cfg.Redis.MasterName),
		MediaDeliveryConfigurationSHA256: fingerprint([]string{sharedStorage.BaseURL, sharedStorage.PrivateBaseURL, logoStorage.BaseURL, logoStorage.PrivateBaseURL}),
	}
	productProjections := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(cfg.Worker.ProductViewCountFlushEnabled)
	boundary.ProductCatalog = &productProjections
	if endpoint := strings.TrimSpace(os.Getenv("CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL")); endpoint != "" {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
			return boundary, errors.New("cannot register invalid customer-service external search configuration")
		}
		boundary.CustomerServiceSearchMode = "external_delete_only"
		boundary.SearchDestinationSHA256 = fingerprint(endpoint)
	}
	if cfg.CustomerServiceRealtime.Enabled {
		stream := cfg.CustomerServiceRealtime.Stream
		if strings.TrimSpace(stream) == "" {
			stream = "customer_service:{realtime}:v1"
		}
		start := strings.IndexByte(stream, '{')
		if start < 0 {
			return boundary, errors.New("cannot register realtime stream without a Redis hash tag")
		}
		end := strings.IndexByte(stream[start+1:], '}')
		if end <= 0 {
			return boundary, errors.New("cannot register realtime stream without a nonempty Redis hash tag")
		}
		boundary.CustomerServiceRealtimeMode = "redis_stream"
		boundary.RealtimeStreamSHA256 = fingerprint(stream)
		// The service derives dedupe keys from the prefix THROUGH the first hash
		// tag, not the full stream name. Changing only a suffix reuses dedupe keys.
		boundary.RealtimeDedupeNamespaceSHA256 = fingerprint(stream[:start+end+2] + ":dedupe:")
	}
	return boundary, boundary.Validate()
}
