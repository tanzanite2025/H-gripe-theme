package admin

import (
	"errors"
	"net/http"
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
	serviceProviderGitHub     = "github"
	serviceProviderGHCR       = "ghcr"
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

type serviceCenterGitHubRepository struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	FullName      string     `json:"full_name"`
	Private       bool       `json:"private"`
	Visibility    string     `json:"visibility"`
	HTMLURL       string     `json:"html_url"`
	DefaultBranch string     `json:"default_branch"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	PushedAt      *time.Time `json:"pushed_at,omitempty"`
	ConnectorID   uint       `json:"connector_id"`
	ConnectorName string     `json:"connector_name"`
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
				serviceProviderGitHub,
				"GitHub",
				"/services/github",
				githubConnectors,
				countServiceProjectsForConnector(assets.Projects, githubConnectors),
			),
			newServiceCenterProviderSummary(
				serviceProviderGHCR,
				"GHCR",
				"/services/github",
				ghcrConnectors,
				countServiceProjectsForConnector(assets.Projects, ghcrConnectors),
			),
		},
		"environment": environment,
		"assets":      assets,
		"network":     network,
	})
}

func (h *ServiceCenterHandler) GitHub(c *gin.Context) {
	environment := strings.TrimSpace(c.Query("environment"))
	connectors, _, err := h.loadResources(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}
	assets, err := h.loadAssets(environment)
	if err != nil {
		h.respondResourceError(c, err)
		return
	}

	githubConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderGitHub)
	ghcrConnectors := filterServiceConnectors(connectors, ops.ConnectorProviderGHCR)
	serviceConnectors := append([]ops.ConnectorView{}, githubConnectors...)
	serviceConnectors = append(serviceConnectors, ghcrConnectors...)
	linkedProjects := filterServiceProjectsForConnectors(assets.Projects, serviceConnectors)
	repositories, repositoryReadErrors := h.readGitHubRepositories(c.Request.Context(), githubConnectors)

	response.Success(c, gin.H{
		"environment":                 environment,
		"generated_at":                time.Now().UTC(),
		"connections":                 serviceConnectors,
		"projects":                    linkedProjects,
		"repositories":                repositories,
		"repository_read_errors":      repositoryReadErrors,
		"connection_count":            len(serviceConnectors),
		"active_connection_count":     countActiveServiceConnections(serviceConnectors),
		"credential_configured_count": countCredentialConfiguredServiceConnections(serviceConnectors),
		"project_count":               len(linkedProjects),
		"repository_count":            len(repositories),
		"attention_count": countServiceConnectionAttention(serviceConnectors) +
			countServiceProjectAttention(linkedProjects) +
			len(repositoryReadErrors),
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

func (h *ServiceCenterHandler) StartGitHubOAuth(c *gin.Context) {
	if h == nil || h.connectorHandler == nil || h.connectorHandler.oauthService == nil {
		apierror.RespondInternalError(c, errors.New("GitHub OAuth service is not configured"))
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
		Provider:    ops.ConnectorProviderGitHub,
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

func (h *ServiceCenterHandler) TestGitHubConnection(c *gin.Context) {
	if h == nil || h.connectorService == nil || h.connectorHandler == nil {
		apierror.RespondInternalError(c, errors.New("GitHub connector service is not configured"))
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
	if connector.Provider != ops.ConnectorProviderGitHub && connector.Provider != ops.ConnectorProviderGHCR {
		apierror.RespondError(c, http.StatusNotFound, "service_not_found", "GitHub connection was not found")
		return
	}
	h.connectorHandler.Test(c)
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
