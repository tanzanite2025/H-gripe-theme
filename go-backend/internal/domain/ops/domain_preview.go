package ops

import "time"

type DomainPreview struct {
	DomainID    uint              `json:"domain_id"`
	Domain      string            `json:"domain"`
	Environment string            `json:"environment"`
	GeneratedAt time.Time         `json:"generated_at"`
	Warnings    []string          `json:"warnings"`
	DNS         DomainDNSPreview  `json:"dns"`
	Caddy       DomainTextPreview `json:"caddy"`
	Nginx       DomainTextPreview `json:"nginx"`
}

type DomainDNSPreview struct {
	Provider       string `json:"provider"`
	Zone           string `json:"zone"`
	RecordType     string `json:"record_type"`
	Name           string `json:"name"`
	Content        string `json:"content"`
	ProxyMode      string `json:"proxy_mode"`
	TLSMode        string `json:"tls_mode"`
	Redirect       bool   `json:"redirect"`
	RedirectTarget string `json:"redirect_target,omitempty"`
}

type DomainTextPreview struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}
