package maintenancegate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
)

// ExternalProjectionBoundary records configuration facts supplied by the business
// process, never inferred from the maintenance process's environment. Digests
// identify destinations without publishing URLs, credentials or Redis keys.
type ExternalProjectionBoundary struct {
	Version                          int                                       `json:"version"`
	CustomerServiceSearchMode        string                                    `json:"customer_service_search_mode"`
	SearchDestinationSHA256          string                                    `json:"search_destination_sha256,omitempty"`
	RedisDestinationSHA256           string                                    `json:"redis_destination_sha256"`
	CustomerServiceRealtimeMode      string                                    `json:"customer_service_realtime_mode"`
	RealtimeStreamSHA256             string                                    `json:"realtime_stream_sha256,omitempty"`
	RealtimeDedupeNamespaceSHA256    string                                    `json:"realtime_dedupe_namespace_sha256,omitempty"`
	MediaDeliveryConfigurationSHA256 string                                    `json:"media_delivery_configuration_sha256"`
	ProductCatalog                   *ProductCatalogExternalProjectionBoundary `json:"product_catalog,omitempty"`
}

var externalProjectionSHA256 = regexp.MustCompile(`^[a-f0-9]{64}$`)

// FingerprintExternalProjectionConfiguration accepts only the caller's explicit
// non-secret configuration facts. Passwords and API tokens must not be supplied.
func FingerprintExternalProjectionConfiguration(configuration any) string {
	encoded, err := json.Marshal(configuration)
	if err != nil {
		panic("external projection configuration must be JSON serializable")
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}

func (boundary ExternalProjectionBoundary) Validate() error {
	if boundary.Version != 1 || !externalProjectionSHA256.MatchString(boundary.RedisDestinationSHA256) || !externalProjectionSHA256.MatchString(boundary.MediaDeliveryConfigurationSHA256) {
		return errors.New("registered external projections require versioned Redis and media delivery identities")
	}
	switch boundary.CustomerServiceSearchMode {
	case "sql_only":
		if boundary.SearchDestinationSHA256 != "" {
			return errors.New("SQL-only search cannot register an external search destination")
		}
	case "external_delete_only":
		if !externalProjectionSHA256.MatchString(boundary.SearchDestinationSHA256) {
			return errors.New("external search requires its configured destination identity")
		}
	default:
		return errors.New("customer-service search configuration is unknown")
	}
	switch boundary.CustomerServiceRealtimeMode {
	case "process_local":
		if boundary.RealtimeStreamSHA256 != "" || boundary.RealtimeDedupeNamespaceSHA256 != "" {
			return errors.New("process-local realtime cannot register Redis stream identities")
		}
	case "redis_stream":
		if !externalProjectionSHA256.MatchString(boundary.RealtimeStreamSHA256) || !externalProjectionSHA256.MatchString(boundary.RealtimeDedupeNamespaceSHA256) {
			return errors.New("Redis realtime requires both stream and dedupe namespace identities")
		}
	default:
		return errors.New("customer-service realtime configuration is unknown")
	}
	if boundary.ProductCatalog != nil {
		return boundary.ProductCatalog.Validate()
	}
	return nil
}
