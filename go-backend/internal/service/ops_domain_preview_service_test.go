package service

import (
	"strings"
	"testing"

	"commerce-platform/internal/domain/ops"
)

func TestBuildOpsDomainPreviewGeneratesIPv4DNSAndGatewayDrafts(t *testing.T) {
	preview := BuildOpsDomainPreview(ops.DomainBinding{
		ID:          7,
		Domain:      "learn.gripe",
		Role:        ops.DomainRoleCanonical,
		Environment: ops.DomainEnvironmentProduction,
		Provider:    ops.DomainProviderCloudflare,
		Zone:        "learn.gripe",
		Target:      "2.25.85.201",
		ProxyMode:   ops.DomainProxyProxied,
		TLSMode:     ops.DomainTLSFullStrict,
	})

	if preview.DomainID != 7 || preview.Domain != "learn.gripe" {
		t.Fatalf("preview domain identity = (%d, %q), want (7, learn.gripe)", preview.DomainID, preview.Domain)
	}
	if preview.GeneratedAt.IsZero() {
		t.Fatal("preview generated_at is zero")
	}
	if len(preview.Warnings) != 0 {
		t.Fatalf("preview warnings = %#v, want none", preview.Warnings)
	}
	if preview.DNS.RecordType != "A" || preview.DNS.Content != "2.25.85.201" {
		t.Fatalf("preview DNS = (%q, %q), want A 2.25.85.201", preview.DNS.RecordType, preview.DNS.Content)
	}
	if !strings.Contains(preview.Caddy.Content, "reverse_proxy 2.25.85.201") {
		t.Fatalf("Caddy preview missing reverse proxy target: %s", preview.Caddy.Content)
	}
	if !strings.Contains(preview.Nginx.Content, "proxy_pass http://2.25.85.201;") {
		t.Fatalf("Nginx preview missing proxy target: %s", preview.Nginx.Content)
	}
}

func TestBuildOpsDomainPreviewGeneratesRedirectDrafts(t *testing.T) {
	preview := BuildOpsDomainPreview(ops.DomainBinding{
		Domain:         "www.learn.gripe",
		Role:           ops.DomainRoleRedirect,
		Environment:    ops.DomainEnvironmentProduction,
		Provider:       ops.DomainProviderCloudflare,
		Zone:           "learn.gripe",
		Target:         "2.25.85.201",
		ProxyMode:      ops.DomainProxyProxied,
		TLSMode:        ops.DomainTLSFullStrict,
		RedirectTarget: "https://learn.gripe",
	})

	if !preview.DNS.Redirect || preview.DNS.RedirectTarget != "https://learn.gripe" {
		t.Fatalf("preview redirect DNS metadata = (%v, %q), want true target", preview.DNS.Redirect, preview.DNS.RedirectTarget)
	}
	if !strings.Contains(preview.Caddy.Content, "redir https://learn.gripe 308") {
		t.Fatalf("Caddy preview missing redirect: %s", preview.Caddy.Content)
	}
	if !strings.Contains(preview.Nginx.Content, "return 308 https://learn.gripe$request_uri;") {
		t.Fatalf("Nginx preview missing redirect: %s", preview.Nginx.Content)
	}
}

func TestBuildOpsDomainPreviewWarnsWhenDNSTargetIsFullURL(t *testing.T) {
	preview := BuildOpsDomainPreview(ops.DomainBinding{
		Domain:      "admin.learn.gripe",
		Role:        ops.DomainRoleAdmin,
		Environment: ops.DomainEnvironmentProduction,
		Provider:    ops.DomainProviderCloudflare,
		Zone:        "learn.gripe",
		Target:      "https://theme-admin.example.com",
		ProxyMode:   ops.DomainProxyDNSOnly,
		TLSMode:     ops.DomainTLSFullStrict,
	})

	if preview.DNS.RecordType != "CNAME" {
		t.Fatalf("preview DNS record_type = %q, want CNAME", preview.DNS.RecordType)
	}
	if !hasWarning(preview.Warnings, "DNS 记录不能直接使用完整 URL") {
		t.Fatalf("preview warnings = %#v, want URL target warning", preview.Warnings)
	}
}

func TestBuildOpsDomainPreviewUsesPlaceholderForMissingUpstream(t *testing.T) {
	preview := BuildOpsDomainPreview(ops.DomainBinding{
		Domain:      "api.learn.gripe",
		Role:        ops.DomainRoleAlias,
		Environment: ops.DomainEnvironmentProduction,
		Provider:    ops.DomainProviderCloudflare,
		Zone:        "learn.gripe",
		ProxyMode:   ops.DomainProxyProxied,
		TLSMode:     ops.DomainTLSFullStrict,
	})

	if !hasWarning(preview.Warnings, "DNS 目标为空") {
		t.Fatalf("preview warnings = %#v, want empty DNS target warning", preview.Warnings)
	}
	if !hasWarning(preview.Warnings, "非跳转域未登记目标") {
		t.Fatalf("preview warnings = %#v, want empty upstream warning", preview.Warnings)
	}
	if !strings.Contains(preview.Caddy.Content, "reverse_proxy REPLACE_WITH_UPSTREAM") {
		t.Fatalf("Caddy preview missing placeholder upstream: %s", preview.Caddy.Content)
	}
}

func hasWarning(warnings []string, fragment string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, fragment) {
			return true
		}
	}
	return false
}
