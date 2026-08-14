package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrOpsDomainSync = errors.New("operations domain sync failed")

type OpsDomainSyncService struct {
	domainRepo       *repository.OpsDomainBindingRepository
	connectorService *OpsConnectorService
}

type cloudflareZone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type cloudflareDNSRecord struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

type cloudflareSSLSetting struct {
	Value string `json:"value"`
}

func NewOpsDomainSyncService(
	domainRepo *repository.OpsDomainBindingRepository,
	connectorService *OpsConnectorService,
) *OpsDomainSyncService {
	return &OpsDomainSyncService{
		domainRepo:       domainRepo,
		connectorService: connectorService,
	}
}

func (s *OpsDomainSyncService) Sync(ctx context.Context, id uint) (*ops.DomainSyncResult, error) {
	if s == nil || s.domainRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations domain sync service is not configured")
	}

	domain, err := s.domainRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	checkedAt := time.Now().UTC()
	result := &ops.DomainSyncResult{
		DomainID:          domain.ID,
		Domain:            domain.Domain,
		ObservedStatus:    ops.DomainObservedError,
		ObservedProxyMode: ops.DomainProxyUnknown,
		ObservedTLSMode:   ops.DomainTLSUnknown,
		LastObservedAt:    checkedAt,
	}

	if domain.ConnectorID == nil {
		return s.persistFailure(result, "未绑定 Cloudflare 只读连接器")
	}
	result.ConnectorID = *domain.ConnectorID

	connector, err := s.connectorService.Get(*domain.ConnectorID)
	if err != nil {
		return s.persistFailure(result, "读取 Cloudflare 连接器失败")
	}
	result.ConnectorName = connector.Name
	result.ObservedSource = "cloudflare:" + connector.Name
	if connector.Provider != ops.ConnectorProviderCloudflare {
		return s.persistFailure(result, "域名绑定的连接器不是 Cloudflare")
	}

	zoneName := strings.TrimSpace(domain.Zone)
	if zoneName == "" {
		return s.persistFailure(result, "域名未登记 Cloudflare Zone")
	}

	var zones struct {
		Result []cloudflareZone `json:"result"`
	}
	_, err = s.connectorService.CloudflareRead(ctx, *domain.ConnectorID, "/zones", url.Values{
		"name":     []string{zoneName},
		"status":   []string{"active"},
		"per_page": []string{"50"},
	}, &zones)
	if err != nil {
		return s.persistFailure(result, "读取 Cloudflare Zone 失败")
	}
	zoneID := ""
	for _, zone := range zones.Result {
		if strings.EqualFold(strings.TrimSpace(zone.Name), zoneName) {
			zoneID = strings.TrimSpace(zone.ID)
			break
		}
	}
	if zoneID == "" {
		result.ObservedStatus = ops.DomainObservedDrifted
		result.ObservedError = "Cloudflare Zone 未找到"
		result.Message = result.ObservedError
		result.ObservedSource = "cloudflare:" + connector.Name
		return result, s.persistResult(result)
	}
	result.ZoneID = zoneID

	var records struct {
		Result []cloudflareDNSRecord `json:"result"`
	}
	_, err = s.connectorService.CloudflareRead(ctx, *domain.ConnectorID, "/zones/"+url.PathEscape(zoneID)+"/dns_records", url.Values{
		"name":     []string{domain.Domain},
		"per_page": []string{"100"},
	}, &records)
	if err != nil {
		return s.persistFailure(result, "读取 Cloudflare DNS 记录失败")
	}

	relevantRecords := relevantDNSRecords(records.Result)
	result.DNSRecordCount = len(relevantRecords)
	result.ObservedTarget = observedTarget(relevantRecords)
	result.ObservedProxyMode = observedProxyMode(relevantRecords)

	var ssl cloudflareSSLSetting
	_, err = s.connectorService.CloudflareRead(ctx, *domain.ConnectorID, "/zones/"+url.PathEscape(zoneID)+"/settings/ssl", nil, &ssl)
	if err != nil {
		return s.persistFailure(result, "读取 Cloudflare SSL/TLS 模式失败")
	}
	result.ObservedTLSMode = normalizeObservedTLSMode(ssl.Value)
	result.ObservedStatus, result.ObservedError = compareDomainObservedState(*domain, *result)
	if result.ObservedStatus == ops.DomainObservedMatched {
		result.Message = "Cloudflare 域名状态已匹配"
	} else {
		result.Message = "Cloudflare 域名状态存在差异"
	}
	return result, s.persistResult(result)
}

func (s *OpsDomainSyncService) persistFailure(result *ops.DomainSyncResult, message string) (*ops.DomainSyncResult, error) {
	result.ObservedStatus = ops.DomainObservedError
	result.ObservedError = message
	result.Message = message
	if err := s.persistResult(result); err != nil {
		return result, err
	}
	return result, fmt.Errorf("%w: %s", ErrOpsDomainSync, message)
}

func (s *OpsDomainSyncService) persistResult(result *ops.DomainSyncResult) error {
	return s.domainRepo.UpdateObservedState(
		result.DomainID,
		result.ObservedStatus,
		result.ObservedTarget,
		result.ObservedProxyMode,
		result.ObservedTLSMode,
		result.ObservedSource,
		result.LastObservedAt,
		result.ObservedError,
	)
}

func observedTarget(records []cloudflareDNSRecord) string {
	records = relevantDNSRecords(records)
	values := make([]string, 0, len(records))
	for _, record := range records {
		content := strings.TrimSpace(record.Content)
		if content != "" {
			values = append(values, content)
		}
	}
	sort.Strings(values)
	return strings.Join(uniqueStrings(values), ", ")
}

func observedProxyMode(records []cloudflareDNSRecord) string {
	records = relevantDNSRecords(records)
	if len(records) == 0 {
		return ops.DomainProxyUnknown
	}
	hasProxied := false
	hasDNSOnly := false
	for _, record := range records {
		if record.Proxied {
			hasProxied = true
		} else {
			hasDNSOnly = true
		}
	}
	if hasProxied && hasDNSOnly {
		return ops.DomainProxyUnknown
	}
	if hasProxied {
		return ops.DomainProxyProxied
	}
	return ops.DomainProxyDNSOnly
}

func relevantDNSRecords(records []cloudflareDNSRecord) []cloudflareDNSRecord {
	relevant := make([]cloudflareDNSRecord, 0, len(records))
	for _, record := range records {
		switch strings.ToUpper(strings.TrimSpace(record.Type)) {
		case "A", "AAAA", "CNAME":
			relevant = append(relevant, record)
		}
	}
	return relevant
}

func normalizeObservedTLSMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "strict":
		return ops.DomainTLSFullStrict
	case "full":
		return ops.DomainTLSFull
	case "flexible":
		return ops.DomainTLSFlexible
	case "off":
		return ops.DomainTLSOff
	default:
		return ops.DomainTLSUnknown
	}
}

func compareDomainObservedState(domain ops.DomainBinding, result ops.DomainSyncResult) (string, string) {
	if result.DNSRecordCount == 0 {
		return ops.DomainObservedDrifted, "Cloudflare 未找到该域名的 A、AAAA 或 CNAME 记录"
	}
	if domain.ProxyMode != ops.DomainProxyUnknown && domain.ProxyMode != result.ObservedProxyMode {
		return ops.DomainObservedDrifted, "代理模式与期望状态不一致"
	}
	if domain.TLSMode != ops.DomainTLSUnknown && domain.TLSMode != result.ObservedTLSMode {
		return ops.DomainObservedDrifted, "SSL/TLS 模式与期望状态不一致"
	}
	if comparableDomainTarget(domain.Target) && !targetContains(result.ObservedTarget, domain.Target) {
		return ops.DomainObservedDrifted, "DNS 目标与期望状态不一致"
	}
	return ops.DomainObservedMatched, ""
}

func comparableDomainTarget(target string) bool {
	target = strings.TrimSpace(target)
	return target != "" &&
		!strings.Contains(target, "://") &&
		!strings.ContainsAny(target, ":/ ")
}

func targetContains(observed, expected string) bool {
	for _, value := range strings.Split(observed, ",") {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(expected)) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if len(unique) == 0 || unique[len(unique)-1] != value {
			unique = append(unique, value)
		}
	}
	return unique
}
