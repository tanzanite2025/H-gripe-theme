package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

const opsCloudflareCachePurgeBatchSize = 30

var ErrOpsCloudflareCachePurge = errors.New("operations Cloudflare cache purge failed")

type opsCloudflareCachePurgeClient interface {
	CloudflareRead(context.Context, uint, string, url.Values, interface{}) (int, error)
	CloudflareWrite(context.Context, uint, string, string, []byte, interface{}) (int, error)
}

type OpsCloudflareCachePurgeService struct {
	domainRepo       *repository.OpsDomainBindingRepository
	connectorService opsCloudflareCachePurgeClient
}

type OpsCloudflareCachePurgeResult struct {
	ProjectID    uint                                 `json:"project_id"`
	DomainCount  int                                  `json:"domain_count"`
	HostCount    int                                  `json:"host_count"`
	ZoneCount    int                                  `json:"zone_count"`
	RequestCount int                                  `json:"request_count"`
	Skipped      bool                                 `json:"skipped"`
	Groups       []OpsCloudflareCachePurgeGroupResult `json:"groups,omitempty"`
	Summary      string                               `json:"summary"`
}

type OpsCloudflareCachePurgeGroupResult struct {
	ConnectorID  uint     `json:"connector_id"`
	Zone         string   `json:"zone"`
	ZoneID       string   `json:"zone_id"`
	Hosts        []string `json:"hosts"`
	RequestCount int      `json:"request_count"`
	OperationIDs []string `json:"operation_ids,omitempty"`
}

type cloudflareCachePurgeZoneLookup struct {
	Result []cloudflareZone `json:"result"`
}

type cloudflareCachePurgeResponse struct {
	ID string `json:"id"`
}

func NewOpsCloudflareCachePurgeService(
	domainRepo *repository.OpsDomainBindingRepository,
	connectorService *OpsConnectorService,
) *OpsCloudflareCachePurgeService {
	return &OpsCloudflareCachePurgeService{
		domainRepo:       domainRepo,
		connectorService: connectorService,
	}
}

func (s *OpsCloudflareCachePurgeService) PurgeProject(ctx context.Context, projectID uint) (*OpsCloudflareCachePurgeResult, error) {
	if s == nil || s.domainRepo == nil || s.connectorService == nil {
		return nil, errors.New("operations Cloudflare cache purge service is not configured")
	}
	if projectID == 0 {
		return nil, fmt.Errorf("%w: project id is required", ErrOpsCloudflareCachePurge)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	domains, err := s.domainRepo.ListByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("%w: load project domains: %v", ErrOpsCloudflareCachePurge, err)
	}

	result := &OpsCloudflareCachePurgeResult{
		ProjectID: projectID,
		Groups:    []OpsCloudflareCachePurgeGroupResult{},
	}
	groups, err := buildCloudflareCachePurgeGroups(domains)
	if err != nil {
		return result, err
	}
	result.DomainCount = countCloudflareCachePurgeDomains(domains)
	if len(groups) == 0 {
		result.Skipped = true
		result.Summary = "没有可清理的 Cloudflare 项目域名，已跳过缓存清理。"
		return result, nil
	}

	groupKeys := make([]string, 0, len(groups))
	for key := range groups {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)
	result.ZoneCount = len(groupKeys)

	for _, groupKey := range groupKeys {
		if err := contextError(ctx); err != nil {
			return result, fmt.Errorf("%w: %v", ErrOpsCloudflareCachePurge, err)
		}
		group := groups[groupKey]
		zoneID, err := s.findZoneID(ctx, group.connectorID, group.zone)
		if err != nil {
			return result, err
		}
		hosts := sortedCachePurgeHosts(group.hosts)
		groupResult := OpsCloudflareCachePurgeGroupResult{
			ConnectorID: group.connectorID,
			Zone:        group.zone,
			ZoneID:      zoneID,
			Hosts:       hosts,
		}
		for start := 0; start < len(hosts); start += opsCloudflareCachePurgeBatchSize {
			if err := contextError(ctx); err != nil {
				return result, fmt.Errorf("%w: %v", ErrOpsCloudflareCachePurge, err)
			}
			end := start + opsCloudflareCachePurgeBatchSize
			if end > len(hosts) {
				end = len(hosts)
			}
			batch := hosts[start:end]
			payload, err := json.Marshal(map[string][]string{"hosts": batch})
			if err != nil {
				return result, fmt.Errorf("%w: encode purge request: %v", ErrOpsCloudflareCachePurge, err)
			}
			var response cloudflareCachePurgeResponse
			_, err = s.connectorService.CloudflareWrite(
				ctx,
				group.connectorID,
				http.MethodPost,
				"/zones/"+url.PathEscape(zoneID)+"/purge_cache",
				payload,
				&response,
			)
			if err != nil {
				return result, fmt.Errorf("%w: zone %s hosts %d-%d: %v", ErrOpsCloudflareCachePurge, group.zone, start+1, end, err)
			}
			groupResult.RequestCount++
			if operationID := strings.TrimSpace(response.ID); operationID != "" {
				groupResult.OperationIDs = append(groupResult.OperationIDs, operationID)
			}
			result.RequestCount++
			result.HostCount += len(batch)
		}
		result.Groups = append(result.Groups, groupResult)
	}

	result.Summary = fmt.Sprintf(
		"Cloudflare 缓存清理完成：%d 个域名、%d 个 Zone、%d 个请求。",
		result.HostCount,
		result.ZoneCount,
		result.RequestCount,
	)
	return result, nil
}

func (s *OpsCloudflareCachePurgeService) findZoneID(ctx context.Context, connectorID uint, zone string) (string, error) {
	var zones cloudflareCachePurgeZoneLookup
	_, err := s.connectorService.CloudflareRead(ctx, connectorID, "/zones", url.Values{
		"name":     []string{zone},
		"status":   []string{"active"},
		"per_page": []string{"50"},
	}, &zones)
	if err != nil {
		return "", fmt.Errorf("%w: read Cloudflare Zone %s: %v", ErrOpsCloudflareCachePurge, zone, err)
	}
	for _, candidate := range zones.Result {
		if strings.EqualFold(strings.TrimSuffix(strings.TrimSpace(candidate.Name), "."), zone) {
			zoneID := strings.TrimSpace(candidate.ID)
			if zoneID != "" {
				return zoneID, nil
			}
		}
	}
	return "", fmt.Errorf("%w: Cloudflare Zone %s was not found", ErrOpsCloudflareCachePurge, zone)
}

type cloudflareCachePurgeGroup struct {
	connectorID uint
	zone        string
	hosts       map[string]struct{}
}

func buildCloudflareCachePurgeGroups(domains []ops.DomainBinding) (map[string]*cloudflareCachePurgeGroup, error) {
	groups := make(map[string]*cloudflareCachePurgeGroup)
	for _, domain := range domains {
		if !domain.Enabled || domain.Provider != ops.DomainProviderCloudflare || !isCachePurgeDomainRole(domain.Role) {
			continue
		}
		if domain.ConnectorID == nil || *domain.ConnectorID == 0 {
			return nil, fmt.Errorf("%w: domain %s has no Cloudflare connector", ErrOpsCloudflareCachePurge, domain.Domain)
		}
		zone := normalizeCachePurgeZone(domain.Zone)
		if zone == "" {
			return nil, fmt.Errorf("%w: domain %s has no Cloudflare Zone", ErrOpsCloudflareCachePurge, domain.Domain)
		}
		host := normalizeCachePurgeHost(domain.Domain)
		if host == "" {
			return nil, fmt.Errorf("%w: domain %s is not a valid cache purge host", ErrOpsCloudflareCachePurge, domain.Domain)
		}
		key := fmt.Sprintf("%d:%s", *domain.ConnectorID, zone)
		group := groups[key]
		if group == nil {
			group = &cloudflareCachePurgeGroup{
				connectorID: *domain.ConnectorID,
				zone:        zone,
				hosts:       make(map[string]struct{}),
			}
			groups[key] = group
		}
		group.hosts[host] = struct{}{}
	}
	return groups, nil
}

func countCloudflareCachePurgeDomains(domains []ops.DomainBinding) int {
	count := 0
	for _, domain := range domains {
		if domain.Enabled && domain.Provider == ops.DomainProviderCloudflare && isCachePurgeDomainRole(domain.Role) {
			count++
		}
	}
	return count
}

func sortedCachePurgeHosts(hosts map[string]struct{}) []string {
	values := make([]string, 0, len(hosts))
	for host := range hosts {
		values = append(values, host)
	}
	sort.Strings(values)
	return values
}

func normalizeCachePurgeZone(value string) string {
	value = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if value == "" || strings.ContainsAny(value, " \t\r\n/") || strings.Contains(value, "://") {
		return ""
	}
	return value
}

func normalizeCachePurgeHost(value string) string {
	value = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	if value == "" || len(value) > 253 || strings.ContainsAny(value, " \t\r\n/") || strings.Contains(value, "://") {
		return ""
	}
	return value
}

func isCachePurgeDomainRole(role string) bool {
	switch role {
	case ops.DomainRoleCanonical, ops.DomainRoleAlias, ops.DomainRoleAdmin, ops.DomainRoleRedirect:
		return true
	default:
		return false
	}
}
