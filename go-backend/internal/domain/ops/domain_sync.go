package ops

import "time"

type DomainSyncResult struct {
	DomainID          uint      `json:"domain_id"`
	Domain            string    `json:"domain"`
	ConnectorID       uint      `json:"connector_id"`
	ConnectorName     string    `json:"connector_name"`
	ZoneID            string    `json:"zone_id,omitempty"`
	ObservedStatus    string    `json:"observed_status"`
	ObservedTarget    string    `json:"observed_target,omitempty"`
	ObservedProxyMode string    `json:"observed_proxy_mode"`
	ObservedTLSMode   string    `json:"observed_tls_mode"`
	ObservedSource    string    `json:"observed_source"`
	LastObservedAt    time.Time `json:"last_observed_at"`
	ObservedError     string    `json:"observed_error,omitempty"`
	DNSRecordCount    int       `json:"dns_record_count"`
	Message           string    `json:"message"`
}
