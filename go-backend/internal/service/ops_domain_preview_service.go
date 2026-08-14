package service

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

type OpsDomainPreviewService struct {
	domainRepo *repository.OpsDomainBindingRepository
}

func NewOpsDomainPreviewService(domainRepo *repository.OpsDomainBindingRepository) *OpsDomainPreviewService {
	return &OpsDomainPreviewService{domainRepo: domainRepo}
}

func (s *OpsDomainPreviewService) Preview(id uint) (*ops.DomainPreview, error) {
	if s == nil || s.domainRepo == nil {
		return nil, errors.New("operations domain preview service is not configured")
	}
	domain, err := s.domainRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return BuildOpsDomainPreview(*domain), nil
}

func BuildOpsDomainPreview(domain ops.DomainBinding) *ops.DomainPreview {
	generatedAt := time.Now().UTC()
	target := strings.TrimSpace(domain.Target)
	content, recordType, targetWarning := previewDNSContent(target)
	warnings := make([]string, 0, 4)
	if targetWarning != "" {
		warnings = append(warnings, targetWarning)
	}
	if domain.Provider != ops.DomainProviderCloudflare {
		warnings = append(warnings, "当前域名提供商不是 Cloudflare，DNS 草稿仅作为记录参考，不会自动下发。")
	}
	if strings.TrimSpace(domain.Zone) == "" {
		warnings = append(warnings, "未登记 DNS Zone，无法确认最终记录所属区域。")
	}
	if domain.Role == ops.DomainRoleRedirect {
		if strings.TrimSpace(domain.RedirectTarget) == "" {
			warnings = append(warnings, "跳转域未登记跳转目标，Caddy/Nginx 只生成占位说明。")
		}
	} else if target == "" {
		warnings = append(warnings, "非跳转域未登记目标，网关配置不会自动指向具体上游。")
	}

	return &ops.DomainPreview{
		DomainID:    domain.ID,
		Domain:      domain.Domain,
		Environment: domain.Environment,
		GeneratedAt: generatedAt,
		Warnings:    warnings,
		DNS: ops.DomainDNSPreview{
			Provider:       domain.Provider,
			Zone:           strings.TrimSpace(domain.Zone),
			RecordType:     recordType,
			Name:           domain.Domain,
			Content:        content,
			ProxyMode:      domain.ProxyMode,
			TLSMode:        domain.TLSMode,
			Redirect:       domain.Role == ops.DomainRoleRedirect,
			RedirectTarget: strings.TrimSpace(domain.RedirectTarget),
		},
		Caddy: ops.DomainTextPreview{
			Filename: "Caddyfile.preview",
			Content:  buildCaddyPreview(domain, target),
		},
		Nginx: ops.DomainTextPreview{
			Filename: "nginx-domain.preview.conf",
			Content:  buildNginxPreview(domain, target),
		},
	}
}

func previewDNSContent(target string) (string, string, string) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", "A", "DNS 目标为空，请先补录 VPS IP、CNAME 或其他解析目标。"
	}
	if parsed := net.ParseIP(target); parsed != nil {
		if parsed.To4() != nil {
			return parsed.To4().String(), "A", ""
		}
		return parsed.String(), "AAAA", ""
	}
	if strings.Contains(target, "://") {
		return target, "CNAME", "目标包含 URL scheme，DNS 记录不能直接使用完整 URL。"
	}
	if parsed, err := url.Parse("https://" + target); err == nil && parsed.Host == target && strings.Contains(target, ".") {
		return strings.TrimSuffix(strings.ToLower(target), "."), "CNAME", ""
	}
	return target, "CNAME", "目标无法明确识别为 IPv4、IPv6 或标准 hostname，请人工确认。"
}

func buildCaddyPreview(domain ops.DomainBinding, target string) string {
	host := domain.Domain
	if domain.Role == ops.DomainRoleRedirect {
		return fmt.Sprintf("%s {\n\tredir %s 308\n}", host, strings.TrimSpace(domain.RedirectTarget))
	}
	upstream := previewUpstream(target)
	return fmt.Sprintf("%s {\n\tencode zstd gzip\n\treverse_proxy %s\n}", host, upstream)
}

func buildNginxPreview(domain ops.DomainBinding, target string) string {
	host := domain.Domain
	if domain.Role == ops.DomainRoleRedirect {
		return fmt.Sprintf("server {\n\tserver_name %s;\n\treturn 308 %s$request_uri;\n}", host, strings.TrimSpace(domain.RedirectTarget))
	}
	upstream := previewUpstream(target)
	return fmt.Sprintf("server {\n\tserver_name %s;\n\tlocation / {\n\t\tproxy_pass http://%s;\n\t\tproxy_set_header Host $host;\n\t\tproxy_set_header X-Forwarded-Proto $scheme;\n\t}\n}", host, upstream)
}

func previewUpstream(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return "REPLACE_WITH_UPSTREAM"
	}
	if strings.Contains(target, "://") {
		return strings.TrimPrefix(strings.TrimPrefix(target, "https://"), "http://")
	}
	return target
}
