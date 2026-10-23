package maintenancegate

import "errors"

const ProductViewCountImmutableBatchSQLReceiptProtocol = "immutable_batch_sql_receipt_compare_ack_v1"

// Optional for existing whole-storefront boundaries. Product handoff inspection
// requires this explicit registration; absence never means Redis has no deltas.
type ProductCatalogExternalProjectionBoundary struct {
	Version                      int    `json:"version"`
	SearchMode                   string `json:"search_mode"`
	DetailCacheMode              string `json:"detail_cache_mode"`
	DetailCacheContractSHA256    string `json:"detail_cache_contract_sha256"`
	ViewCountMode                string `json:"view_count_mode"`
	ViewCountKeysSHA256          string `json:"view_count_keys_sha256,omitempty"`
	ViewCountFlushProtocol       string `json:"view_count_flush_protocol,omitempty"`
	ViewCountFlushContractSHA256 string `json:"view_count_flush_contract_sha256,omitempty"`
}

func (boundary ProductCatalogExternalProjectionBoundary) Validate() error {
	if (boundary.Version != 1 && boundary.Version != 2) || boundary.SearchMode != "sql_only" || boundary.DetailCacheMode != "redis" || !externalProjectionSHA256.MatchString(boundary.DetailCacheContractSHA256) {
		return errors.New("product projection registration requires the known SQL search and Redis detail-cache contract")
	}
	switch boundary.ViewCountMode {
	case "sql_direct":
		if boundary.ViewCountKeysSHA256 != "" {
			return errors.New("direct SQL view counting cannot register a Redis delta buffer")
		}
	case "redis_delta_buffer":
		if !externalProjectionSHA256.MatchString(boundary.ViewCountKeysSHA256) {
			return errors.New("buffered product views require both active and processing key identities")
		}
	default:
		return errors.New("product view count persistence configuration is unknown")
	}
	if boundary.Version == 1 && (boundary.ViewCountFlushProtocol != "" || boundary.ViewCountFlushContractSHA256 != "") {
		return errors.New("legacy product projection cannot claim an idempotent flush protocol")
	}
	if boundary.Version == 2 && (boundary.ViewCountMode != "redis_delta_buffer" || boundary.ViewCountFlushProtocol != ProductViewCountImmutableBatchSQLReceiptProtocol || !externalProjectionSHA256.MatchString(boundary.ViewCountFlushContractSHA256)) {
		return errors.New("product projection version 2 requires the registered immutable batch SQL receipt protocol")
	}
	return nil
}
