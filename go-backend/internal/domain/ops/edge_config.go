package ops

import "time"

const EdgeConfigSchemaVersion = 1

type EdgeConfig struct {
	SchemaVersion int                  `json:"schema_version"`
	Environment   string               `json:"environment"`
	GeneratedAt   time.Time            `json:"generated_at"`
	Canonical     string               `json:"canonical"`
	Domains       []EdgeDomainRoute    `json:"domains"`
	Cloudflare    []EdgeCloudflareRule `json:"cloudflare"`
	Caddy         string               `json:"caddy"`
	Nginx         EdgeNginxConfig      `json:"nginx"`
}

type EdgeDomainRoute struct {
	Domain            string                   `json:"domain"`
	Role              string                   `json:"role"`
	ProjectBindingID  uint                     `json:"project_binding_id"`
	Project           string                   `json:"project"`
	GatewayAlias      string                   `json:"gateway_alias"`
	Upstream          string                   `json:"upstream,omitempty"`
	RedirectTarget    string                   `json:"redirect_target,omitempty"`
	QuickBuyRateLimit *QuickBuyRateLimitPolicy `json:"quick_buy_rate_limit,omitempty"`
}

type EdgeCloudflareRule struct {
	Domain        string `json:"domain"`
	Role          string `json:"role"`
	Zone          string `json:"zone"`
	Target        string `json:"target"`
	ProxyMode     string `json:"proxy_mode"`
	TLSMode       string `json:"tls_mode"`
	Status        string `json:"status"`
	ObservedState string `json:"observed_status"`
	Enabled       bool   `json:"enabled"`
}

type EdgeNginxConfig struct {
	StorefrontServerNames []string `json:"storefront_server_names"`
	AdminServerNames      []string `json:"admin_server_names"`
}
