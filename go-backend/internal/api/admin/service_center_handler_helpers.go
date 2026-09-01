package admin

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func serviceCenterCacheRuleTarget(c *gin.Context) (uint, string, bool) {
	rawConnectorID := strings.TrimSpace(c.Query("connector_id"))
	connectorID64, err := strconv.ParseUint(rawConnectorID, 10, 32)
	if err != nil || connectorID64 == 0 {
		apierror.RespondBadRequest(c, "connector_id is required")
		return 0, "", false
	}
	zone := strings.TrimSpace(c.Query("zone"))
	if zone == "" {
		apierror.RespondBadRequest(c, "zone is required")
		return 0, "", false
	}
	return uint(connectorID64), zone, true
}

func newServiceCenterProviderSummary(
	id string,
	label string,
	route string,
	connectors []ops.ConnectorView,
	resourceCount int,
) serviceCenterProviderSummary {
	return serviceCenterProviderSummary{
		ID:                    id,
		Label:                 label,
		Route:                 route,
		ConnectionCount:       len(connectors),
		ActiveConnectionCount: countActiveServiceConnections(connectors),
		ResourceCount:         resourceCount,
		Status:                serviceCenterProviderStatus(connectors),
	}
}

func serviceCenterProviderStatus(connectors []ops.ConnectorView) string {
	if len(connectors) == 0 {
		return "not_connected"
	}
	for _, connector := range connectors {
		if connector.Enabled && connector.Status == ops.ConnectorStatusActive {
			return "active"
		}
	}
	for _, connector := range connectors {
		if connector.Status == ops.ConnectorStatusError {
			return "attention"
		}
	}
	return "pending"
}

func filterServiceConnectors(connectors []ops.ConnectorView, provider string) []ops.ConnectorView {
	filtered := make([]ops.ConnectorView, 0)
	for _, connector := range connectors {
		if connector.Provider == provider {
			filtered = append(filtered, connector)
		}
	}
	return filtered
}

func filterServiceDomains(domains []ops.DomainBinding, provider string) []ops.DomainBinding {
	filtered := make([]ops.DomainBinding, 0)
	for _, domain := range domains {
		if domain.Provider == provider {
			filtered = append(filtered, domain)
		}
	}
	return filtered
}

func countActiveServiceConnections(connectors []ops.ConnectorView) int {
	count := 0
	for _, connector := range connectors {
		if connector.Enabled && connector.Status == ops.ConnectorStatusActive {
			count++
		}
	}
	return count
}

func countCredentialConfiguredServiceConnections(connectors []ops.ConnectorView) int {
	count := 0
	for _, connector := range connectors {
		if connector.CredentialConfigured {
			count++
		}
	}
	return count
}

func countServiceConnectionAttention(connectors []ops.ConnectorView) int {
	count := 0
	for _, connector := range connectors {
		if connector.Status == ops.ConnectorStatusPending || connector.Status == ops.ConnectorStatusError ||
			connector.LastTestStatus == ops.ConnectorTestFailed {
			count++
		}
	}
	return count
}

func countServiceVPS(records []ops.VPSBinding, provider string) int {
	count := 0
	for _, record := range records {
		if record.Provider == provider {
			count++
		}
	}
	return count
}

func filterServiceProjectsForConnectors(
	projects []ops.ProjectBindingView,
	connectors []ops.ConnectorView,
) []ops.ProjectBindingView {
	connectorIDs := make(map[uint]struct{}, len(connectors))
	for _, connector := range connectors {
		connectorIDs[connector.ID] = struct{}{}
	}

	filtered := make([]ops.ProjectBindingView, 0)
	for _, project := range projects {
		if project.ConnectorID == nil {
			continue
		}
		if _, ok := connectorIDs[*project.ConnectorID]; ok {
			filtered = append(filtered, project)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Environment == filtered[j].Environment {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].Environment < filtered[j].Environment
	})
	return filtered
}

func countServiceProjectAttention(projects []ops.ProjectBindingView) int {
	count := 0
	for _, project := range projects {
		if project.Status == ops.ProjectStatusPending || project.Status == ops.ProjectStatusDrifted ||
			project.Status == ops.ProjectStatusError || project.HealthStatus == ops.ProjectHealthUnknown ||
			project.HealthStatus == ops.ProjectHealthDegraded || project.HealthStatus == ops.ProjectHealthOffline {
			count++
		}
	}
	return count
}

func countServiceProjectsForConnector(
	projects []ops.ProjectBindingView,
	connectors []ops.ConnectorView,
) int {
	connectorIDs := make(map[uint]struct{}, len(connectors))
	for _, connector := range connectors {
		connectorIDs[connector.ID] = struct{}{}
	}

	count := 0
	for _, project := range projects {
		if project.ConnectorID == nil {
			continue
		}
		if _, ok := connectorIDs[*project.ConnectorID]; ok {
			count++
		}
	}
	return count
}

func (h *ServiceCenterHandler) readGitHubRepositories(
	ctx context.Context,
	connectors []ops.ConnectorView,
) ([]serviceCenterGitHubRepository, []string) {
	repositories := make([]serviceCenterGitHubRepository, 0)
	readErrors := make([]string, 0)
	if h == nil || h.connectorService == nil {
		return repositories, readErrors
	}
	query := url.Values{
		"per_page":  []string{"20"},
		"sort":      []string{"updated"},
		"direction": []string{"desc"},
	}
	for _, connector := range connectors {
		if !connector.Enabled || !connector.CredentialConfigured {
			continue
		}
		var current []serviceCenterGitHubRepository
		if _, err := h.connectorService.GitHubRead(ctx, connector.ID, "/user/repos", query, &current); err != nil {
			readErrors = append(readErrors, connector.Name+": "+err.Error())
			continue
		}
		for i := range current {
			current[i].ConnectorID = connector.ID
			current[i].ConnectorName = connector.Name
			if strings.TrimSpace(current[i].Visibility) == "" {
				if current[i].Private {
					current[i].Visibility = "private"
				} else {
					current[i].Visibility = "public"
				}
			}
			repositories = append(repositories, current[i])
		}
	}
	sort.Slice(repositories, func(i, j int) bool {
		left := repositories[i].UpdatedAt
		right := repositories[j].UpdatedAt
		if left != nil && right != nil && !left.Equal(*right) {
			return left.After(*right)
		}
		if repositories[i].ConnectorID == repositories[j].ConnectorID {
			return repositories[i].FullName < repositories[j].FullName
		}
		return repositories[i].ConnectorID < repositories[j].ConnectorID
	})
	return repositories, readErrors
}

func buildServiceCenterCloudflareZones(
	domains []ops.DomainBinding,
	connectors []ops.ConnectorView,
) []serviceCenterCloudflareZone {
	connectorNames := make(map[uint]string, len(connectors))
	for _, connector := range connectors {
		connectorNames[connector.ID] = connector.Name
	}

	zones := make(map[string]*serviceCenterCloudflareZone)
	for _, domain := range domains {
		name := strings.TrimSpace(domain.Zone)
		if name == "" {
			name = domain.Domain
		}
		key := domain.Environment + ":" + name
		if domain.ConnectorID != nil {
			key += ":" + strconv.FormatUint(uint64(*domain.ConnectorID), 10)
		}
		zone := zones[key]
		if zone == nil {
			zone = &serviceCenterCloudflareZone{
				Name:        name,
				Environment: domain.Environment,
				ConnectorID: domain.ConnectorID,
				Domains:     []ops.DomainBinding{},
			}
			if domain.ConnectorID != nil {
				zone.ConnectorName = connectorNames[*domain.ConnectorID]
			}
			zones[key] = zone
		}
		zone.Domains = append(zone.Domains, domain)
		zone.DomainCount++
	}

	result := make([]serviceCenterCloudflareZone, 0, len(zones))
	for _, zone := range zones {
		sort.Slice(zone.Domains, func(i, j int) bool {
			return zone.Domains[i].Domain < zone.Domains[j].Domain
		})
		result = append(result, *zone)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Environment == result[j].Environment {
			return result[i].Name < result[j].Name
		}
		return result[i].Environment < result[j].Environment
	})
	return result
}

func (h *ServiceCenterHandler) loadResources(environment string) ([]ops.ConnectorView, []ops.DomainBinding, error) {
	if h == nil || h.connectorService == nil || h.domainService == nil {
		return nil, nil, errors.New("service center is not configured")
	}
	connectors, err := h.connectorService.ListForEnvironment(environment)
	if err != nil {
		return nil, nil, err
	}
	domains, err := h.domainService.ListForEnvironment(environment)
	if err != nil {
		return nil, nil, err
	}
	return connectors, domains, nil
}

func (h *ServiceCenterHandler) loadAssets(environment string) (*serviceCenterAssetOverview, error) {
	result := &serviceCenterAssetOverview{
		VPS:      []ops.VPSBinding{},
		Projects: []ops.ProjectBindingView{},
		Domains:  []ops.DomainBinding{},
	}
	if h == nil || h.opsOverview == nil {
		return result, nil
	}

	environments := []string{environment}
	if strings.TrimSpace(environment) == "" {
		environments = []string{
			ops.DomainEnvironmentProduction,
			ops.DomainEnvironmentStaging,
			ops.DomainEnvironmentTest,
			ops.DomainEnvironmentLocal,
		}
	}

	for _, currentEnvironment := range environments {
		overview, err := h.opsOverview.GetForEnvironment(currentEnvironment)
		if err != nil {
			return nil, err
		}
		result.VPS = append(result.VPS, overview.Topology.VPS...)
		result.Projects = append(result.Projects, overview.Topology.Projects...)
		result.Domains = append(result.Domains, overview.Topology.Domains...)
	}
	return result, nil
}

func (h *ServiceCenterHandler) loadNetworkSummary(environment string) (*ops.NetworkSummary, error) {
	if h == nil || h.opsNetworkSummary == nil {
		return &ops.NetworkSummary{
			Environment: strings.TrimSpace(environment),
			GeneratedAt: time.Now().UTC(),
			Summary: ops.NetworkSummaryCounts{
				ManagedBy: map[string]int{},
				Scopes:    map[string]int{},
			},
			Items: []ops.NetworkSummaryItem{},
		}, nil
	}
	return h.opsNetworkSummary.Get(environment)
}

func (h *ServiceCenterHandler) respondResourceError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrInvalidOpsConnectorEnvironment) ||
		errors.Is(err, service.ErrInvalidOpsDomainEnvironment) ||
		errors.Is(err, service.ErrInvalidOpsOverviewEnvironment) ||
		errors.Is(err, service.ErrInvalidOpsNetworkEnvironment) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	apierror.RespondInternalError(c, err)
}

func (h *ServiceCenterHandler) respondCacheRulesError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCloudflareCacheRule):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrCloudflareCacheRuleNotFound):
		apierror.RespondNotFound(c, "Cloudflare cache rule")
	default:
		apierror.RespondError(c, http.StatusBadGateway, "cloudflare_cache_rules_failed", err.Error())
	}
}

func (h *ServiceCenterHandler) recordCloudflareCacheRuleAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	_ = recordAdminAudit(h.auditService, c, event)
}
