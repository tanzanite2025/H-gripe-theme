package admin

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	serviceProviderCloudflare = "cloudflare"
	serviceProviderHostinger  = "hostinger"

	adminAuditResourceServiceCloudflareCacheRule = "service_cloudflare_cache_rule"
)

type ServiceCenterHandler struct {
	connectorService  *service.OpsConnectorService
	domainService     *service.OpsDomainBindingService
	connectorHandler  *OpsConnectorHandler
	cacheRulesService *service.CloudflareCacheRulesService
	opsOverview       *service.OpsOverviewService
	opsNetworkSummary *service.OpsNetworkSummaryService
	auditService      adminAuditRecorder
}

func NewServiceCenterHandler(
	connectorService *service.OpsConnectorService,
	domainService *service.OpsDomainBindingService,
	connectorHandler *OpsConnectorHandler,
) *ServiceCenterHandler {
	return &ServiceCenterHandler{
		connectorService: connectorService,
		domainService:    domainService,
		connectorHandler: connectorHandler,
	}
}

func (h *ServiceCenterHandler) ConfigureCloudflareCacheRulesService(cacheRulesService *service.CloudflareCacheRulesService) {
	if h == nil {
		return
	}
	h.cacheRulesService = cacheRulesService
}

func (h *ServiceCenterHandler) ConfigureOpsOverviewService(opsOverview *service.OpsOverviewService) {
	if h == nil {
		return
	}
	h.opsOverview = opsOverview
}

func (h *ServiceCenterHandler) ConfigureOpsNetworkSummaryService(opsNetworkSummary *service.OpsNetworkSummaryService) {
	if h == nil {
		return
	}
	h.opsNetworkSummary = opsNetworkSummary
}

func (h *ServiceCenterHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

type serviceCenterProviderSummary struct {
	ID                    string `json:"id"`
	Label                 string `json:"label"`
	Route                 string `json:"route,omitempty"`
	ConnectionCount       int    `json:"connection_count"`
	ActiveConnectionCount int    `json:"active_connection_count"`
	ResourceCount         int    `json:"resource_count"`
	Status                string `json:"status"`
}

type serviceCenterCloudflareZone struct {
	Name          string              `json:"name"`
	Environment   string              `json:"environment"`
	ConnectorID   *uint               `json:"connector_id,omitempty"`
	ConnectorName string              `json:"connector_name,omitempty"`
	DomainCount   int                 `json:"domain_count"`
	Domains       []ops.DomainBinding `json:"domains"`
}

type serviceCenterAssetOverview struct {
	VPS      []ops.VPSBinding         `json:"vps"`
	Projects []ops.ProjectBindingView `json:"projects"`
	Domains  []ops.DomainBinding      `json:"domains"`
}

func (h *ServiceCenterHandler) Overview(c *gin.Context) {
	environment := strings.TrimSpace(c.Query("environment"))
	connectors, domains, err := h.loadResources(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}
	assets, err := h.loadAssets(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}
	network, err := h.loadNetworkSummary(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}

	cloudflareConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderCloudflare)
	hostingerConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderHostinger)
	githubConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderGitHub)
	ghcrConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderGHCR)
	cloudflareDomains := filterServiceDomains(domains, ops.DomainProviderCloudflare)
	hostingerVPSCount := countServiceVPS(assets.VPS, ops.VPSProviderHostinger)

	response.Success(c, gin.H{
		"generated_at": time.Now().UTC(),
		"providers": []serviceCenterProviderSummary{
			newServiceCenterProviderSummary(
				serviceProviderCloudflare,
				"Cloudflare",
				"/services/cloudflare",
				cloudflareConnectors,
				len(cloudflareDomains),
			),
			newServiceCenterProviderSummary(
				serviceProviderHostinger,
				"Hostinger",
				"/ops/vps",
				hostingerConnectors,
				hostingerVPSCount,
			),
			newServiceCenterProviderSummary(
				"github",
				"GitHub",
				"/ops/deployments",
				githubConnectors,
				countServiceProjectsForConnector(assets.Projects, githubConnectors),
			),
			newServiceCenterProviderSummary(
				"ghcr",
				"GHCR",
				"/ops/deployments",
				ghcrConnectors,
				countServiceProjectsForConnector(assets.Projects, ghcrConnectors),
			),
		},
		"environment": environment,
		"assets":      assets,
		"network":     network,
	})
}

func (h *ServiceCenterHandler) Cloudflare(c *gin.Context) {
	environment := strings.TrimSpace(c.Query("environment"))
	connectors, domains, err := h.loadResources(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}

	cloudflareConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderCloudflare)
	cloudflareDomains := filterServiceDomains(domains, ops.DomainProviderCloudflare)
	zones := buildServiceCenterCloudflareZones(cloudflareDomains, cloudflareConnectors)

	credentialConfiguredCount := 0
	attentionCount := 0
	for _, connector := range cloudflareConnectors {
		if connector.CredentialConfigured {
			credentialConfiguredCount++
		}
		if connector.Status == ops.ConnectorStatusPending || connector.Status == ops.ConnectorStatusError {
			attentionCount++
		}
	}
	for _, domain := range cloudflareDomains {
		if domain.Status == ops.DomainStatusDrifted || domain.Status == ops.DomainStatusError ||
			domain.ObservedStatus == ops.DomainObservedDrifted || domain.ObservedStatus == ops.DomainObservedError {
			attentionCount++
		}
	}

	response.Success(c, gin.H{
		"environment":                 environment,
		"generated_at":                time.Now().UTC(),
		"connections":                 cloudflareConnectors,
		"domains":                     cloudflareDomains,
		"zones":                       zones,
		"connection_count":            len(cloudflareConnectors),
		"active_connection_count":     countActiveServiceConnections(cloudflareConnectors),
		"credential_configured_count": credentialConfiguredCount,
		"domain_count":                len(cloudflareDomains),
		"zone_count":                  len(zones),
		"attention_count":             attentionCount,
	})
}

func (h *ServiceCenterHandler) StartCloudflareOAuth(c *gin.Context) {
	if h == nil || h.connectorHandler == nil || h.connectorHandler.oauthService == nil {
		apierror.RespondInternalError(c, errors.New("Cloudflare OAuth service is not configured"))
		return
	}
	userID, ok := currentOpsConnectorUserID(c)
	if !ok {
		return
	}

	var req struct {
		ConnectorID *uint  `json:"connector_id"`
		Environment string `json:"environment"`
		ReturnPath  string `json:"return_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.connectorHandler.oauthService.Start(c.Request.Context(), userID, service.OpsConnectorOAuthStartInput{
		Provider:    ops.ConnectorProviderCloudflare,
		ConnectorID: req.ConnectorID,
		Environment: req.Environment,
		ReturnPath:  req.ReturnPath,
	})
	if err != nil {
		respondOpsConnectorOAuthError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ServiceCenterHandler) TestCloudflareConnection(c *gin.Context) {
	if h == nil || h.connectorService == nil || h.connectorHandler == nil {
		apierror.RespondInternalError(c, errors.New("Cloudflare connector service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid connector id")
	if err != nil {
		return
	}
	connector, err := h.connectorService.Get(id)
	if err != nil {
		respondOpsConnectorError(c, err)
		return
	}
	if connector.Provider != ops.ConnectorProviderCloudflare {
		apierror.RespondError(c, http.StatusNotFound, "service_not_found", "Cloudflare connection was not found")
		return
	}
	h.connectorHandler.Test(c)
}

func (h *ServiceCenterHandler) GetCloudflareCacheRules(c *gin.Context) {
	connectorID, zone, ok := serviceCenterCacheRuleTarget(c)
	if !ok {
		return
	}
	if h == nil || h.cacheRulesService == nil {
		apierror.RespondInternalError(c, errors.New("Cloudflare cache rules service is not configured"))
		return
	}
	result, err := h.cacheRulesService.Get(c.Request.Context(), connectorID, zone)
	if err != nil {
		h.respondCacheRulesError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ServiceCenterHandler) UpdateCloudflareCacheRuleEnabled(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	connectorID, zone, ok := serviceCenterCacheRuleTarget(c)
	if !ok {
		return
	}
	ruleID := strings.TrimSpace(c.Param("rule_id"))
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordCloudflareCacheRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceServiceCloudflareCacheRule,
			ResourceID:   connectorID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"zone": zone, "rule_id": ruleID},
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if h == nil || h.cacheRulesService == nil {
		err := errors.New("Cloudflare cache rules service is not configured")
		h.recordCloudflareCacheRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceServiceCloudflareCacheRule,
			ResourceID:   connectorID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"zone": zone, "rule_id": ruleID, "enabled": req.Enabled},
		})
		apierror.RespondInternalError(c, err)
		return
	}
	result, err := h.cacheRulesService.SetRuleEnabled(c.Request.Context(), connectorID, zone, ruleID, req.Enabled)
	if err != nil {
		h.recordCloudflareCacheRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceServiceCloudflareCacheRule,
			ResourceID:   connectorID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"zone": zone, "rule_id": ruleID, "enabled": req.Enabled},
		})
		h.respondCacheRulesError(c, err)
		return
	}
	h.recordCloudflareCacheRuleAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceServiceCloudflareCacheRule,
		ResourceID: connectorID,
		Status:     adminAuditStatusSuccess,
		Changes:    gin.H{"zone": zone, "rule_id": ruleID, "enabled": req.Enabled},
		NewValue:   result,
	})
	response.Success(c, result)
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
	recordAdminAudit(h.auditService, c, event)
}

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

func countServiceVPS(records []ops.VPSBinding, provider string) int {
	count := 0
	for _, record := range records {
		if record.Provider == provider {
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
