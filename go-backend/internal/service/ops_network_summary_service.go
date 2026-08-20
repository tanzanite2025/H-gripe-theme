package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrInvalidOpsNetworkEnvironment = errors.New("invalid operations network environment")

type OpsNetworkSummaryService struct {
	networkRuleRepo *repository.OpsNetworkRuleRepository
	vpsRepo         *repository.OpsVPSBindingRepository
	projectRepo     *repository.OpsProjectBindingRepository
	domainRepo      *repository.OpsDomainBindingRepository
	connectorRepo   *repository.OpsConnectorRepository
}

func NewOpsNetworkSummaryService(
	networkRuleRepo *repository.OpsNetworkRuleRepository,
	vpsRepo *repository.OpsVPSBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	domainRepo *repository.OpsDomainBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
) *OpsNetworkSummaryService {
	return &OpsNetworkSummaryService{
		networkRuleRepo: networkRuleRepo,
		vpsRepo:         vpsRepo,
		projectRepo:     projectRepo,
		domainRepo:      domainRepo,
		connectorRepo:   connectorRepo,
	}
}

func (s *OpsNetworkSummaryService) Get(environment string) (*ops.NetworkSummary, error) {
	environment, err := normalizeOpsNetworkEnvironment(environment)
	if err != nil {
		return nil, err
	}
	if s == nil || s.networkRuleRepo == nil || s.vpsRepo == nil || s.projectRepo == nil || s.domainRepo == nil || s.connectorRepo == nil {
		return nil, errors.New("operations network summary service is not configured")
	}

	rules, err := s.networkRuleRepo.ListByEnvironment(environment)
	if err != nil {
		return nil, fmt.Errorf("load network rules: %w", err)
	}
	vps, err := s.vpsRepo.ListByEnvironment(environment)
	if err != nil {
		return nil, fmt.Errorf("load network VPS bindings: %w", err)
	}
	projects, err := s.projectRepo.ListByEnvironment(environment)
	if err != nil {
		return nil, fmt.Errorf("load network project bindings: %w", err)
	}
	domains, err := s.domainRepo.ListByEnvironment(environment)
	if err != nil {
		return nil, fmt.Errorf("load network domain bindings: %w", err)
	}
	connectors, err := s.connectorRepo.ListByEnvironment(environment)
	if err != nil {
		return nil, fmt.Errorf("load network connectors: %w", err)
	}

	vpsByID := make(map[uint]ops.VPSBinding, len(vps))
	for _, record := range vps {
		vpsByID[record.ID] = record
	}
	projectsByID := make(map[uint]ops.ProjectBindingView, len(projects))
	for _, record := range projects {
		projectsByID[record.ID] = record
	}
	domainsByID := make(map[uint]ops.DomainBinding, len(domains))
	for _, record := range domains {
		domainsByID[record.ID] = record
	}
	connectorsByID := make(map[uint]ops.ConnectorView, len(connectors))
	for _, record := range connectors {
		connectorsByID[record.ID] = ops.ConnectorView{
			ID:                   record.ID,
			Name:                 record.Name,
			Provider:             record.Provider,
			Environment:          record.Environment,
			Endpoint:             record.Endpoint,
			AuthType:             record.AuthType,
			CredentialRef:        record.CredentialRef,
			CredentialConfigured: strings.TrimSpace(record.CredentialsEncrypted) != "",
			Scopes:               record.Scopes,
			Status:               record.Status,
			Enabled:              record.Enabled,
			LastTestStatus:       record.LastTestStatus,
			LastTestedAt:         record.LastTestedAt,
			LastError:            record.LastError,
			Notes:                record.Notes,
			CreatedAt:            record.CreatedAt,
			UpdatedAt:            record.UpdatedAt,
		}
	}

	items := make([]ops.NetworkSummaryItem, 0, len(rules)+len(domains))
	for _, rule := range rules {
		items = append(items, buildNetworkRuleSummaryItem(rule, vpsByID, projectsByID, domainsByID, connectorsByID))
	}
	for _, domain := range domains {
		if !isNetworkDomainBinding(domain) {
			continue
		}
		items = append(items, buildDomainNetworkSummaryItem(domain, vpsByID, projectsByID, domainsByID, connectorsByID))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind == items[j].Kind {
			return items[i].Name < items[j].Name
		}
		return items[i].Kind < items[j].Kind
	})

	return &ops.NetworkSummary{
		Environment: environment,
		GeneratedAt: time.Now().UTC(),
		Summary:     summarizeNetworkItems(items),
		Items:       items,
	}, nil
}

func normalizeOpsNetworkEnvironment(environment string) (string, error) {
	environment = strings.ToLower(strings.TrimSpace(environment))
	if environment == "" {
		return "", nil
	}
	switch environment {
	case ops.DomainEnvironmentProduction,
		ops.DomainEnvironmentStaging,
		ops.DomainEnvironmentTest,
		ops.DomainEnvironmentLocal:
		return environment, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidOpsNetworkEnvironment, environment)
	}
}

func buildNetworkRuleSummaryItem(
	rule ops.NetworkRule,
	vpsByID map[uint]ops.VPSBinding,
	projectsByID map[uint]ops.ProjectBindingView,
	domainsByID map[uint]ops.DomainBinding,
	connectorsByID map[uint]ops.ConnectorView,
) ops.NetworkSummaryItem {
	item := ops.NetworkSummaryItem{
		Key:              fmt.Sprintf("rule-%d", rule.ID),
		Kind:             ops.NetworkItemKindRule,
		ID:               rule.ID,
		Name:             rule.Name,
		Environment:      rule.Environment,
		OwnerKind:        rule.OwnerKind,
		OwnerID:          rule.OwnerID,
		VPSBindingID:     rule.VPSBindingID,
		ProjectBindingID: rule.ProjectBindingID,
		DomainBindingID:  rule.DomainBindingID,
		ConnectorID:      rule.ConnectorID,
		ManagedBy:        rule.ManagedBy,
		SourceKind:       firstNetworkValue(rule.SourceKind, ops.NetworkSourceFirewallRule),
		Scope:            rule.Scope,
		Direction:        rule.Direction,
		Protocol:         rule.Protocol,
		Ports:            rule.Ports,
		SourceCIDR:       rule.SourceCIDR,
		Target:           rule.Target,
		DesiredState:     rule.DesiredState,
		ObservedState:    rule.ObservedState,
		EffectiveState:   rule.EffectiveState,
		Status:           rule.Status,
		Enabled:          rule.Enabled,
		LastObservedAt:   rule.LastObservedAt,
		LastError:        rule.LastError,
		Notes:            rule.Notes,
	}
	attachNetworkReferences(&item, vpsByID, projectsByID, domainsByID, connectorsByID)
	if item.OwnerName == "" {
		item.OwnerName = networkOwnerName(item)
	}
	return item
}

func buildDomainNetworkSummaryItem(
	domain ops.DomainBinding,
	vpsByID map[uint]ops.VPSBinding,
	projectsByID map[uint]ops.ProjectBindingView,
	domainsByID map[uint]ops.DomainBinding,
	connectorsByID map[uint]ops.ConnectorView,
) ops.NetworkSummaryItem {
	item := ops.NetworkSummaryItem{
		Key:              fmt.Sprintf("domain-dns-%d", domain.ID),
		Kind:             ops.NetworkItemKindDomainDNS,
		ID:               domain.ID,
		Name:             domain.Domain,
		Environment:      domain.Environment,
		OwnerKind:        ops.NetworkOwnerDomain,
		OwnerID:          domain.ID,
		DomainBindingID:  networkUintPointer(domain.ID),
		ProjectBindingID: domain.ProjectBindingID,
		ConnectorID:      domain.ConnectorID,
		ManagedBy:        firstNetworkValue(domain.Provider, ops.NetworkManagedByOther),
		SourceKind:       ops.NetworkSourceDomainBinding,
		Scope:            networkDomainScope(domain),
		Direction:        ops.NetworkDirectionIngress,
		Protocol:         ops.NetworkProtocolTCP,
		Target:           firstNetworkValue(domain.Target, domain.ObservedTarget),
		DesiredState:     networkDomainDesiredState(domain),
		ObservedState:    networkDomainObservedState(domain),
		EffectiveState:   networkDomainEffectiveState(domain),
		Status:           domain.Status,
		Enabled:          domain.Enabled,
		LastObservedAt:   domain.LastObservedAt,
		LastError:        domain.ObservedError,
		Notes:            domain.Notes,
		DomainName:       domain.Domain,
		OwnerName:        domain.Domain,
	}
	attachNetworkReferences(&item, vpsByID, projectsByID, domainsByID, connectorsByID)
	return item
}

func attachNetworkReferences(
	item *ops.NetworkSummaryItem,
	vpsByID map[uint]ops.VPSBinding,
	projectsByID map[uint]ops.ProjectBindingView,
	domainsByID map[uint]ops.DomainBinding,
	connectorsByID map[uint]ops.ConnectorView,
) {
	if item == nil {
		return
	}
	if item.VPSBindingID != nil {
		if record, ok := vpsByID[*item.VPSBindingID]; ok {
			item.VPSName = record.Name
		}
	}
	if item.ProjectBindingID != nil {
		if record, ok := projectsByID[*item.ProjectBindingID]; ok {
			item.ProjectName = record.Name
			if item.VPSBindingID == nil && record.VPSBindingID > 0 {
				item.VPSBindingID = networkUintPointer(record.VPSBindingID)
				item.VPSName = record.VPSName
			}
		}
	}
	if item.DomainBindingID != nil {
		if record, ok := domainsByID[*item.DomainBindingID]; ok {
			item.DomainName = record.Domain
			if item.ProjectBindingID == nil && record.ProjectBindingID != nil {
				item.ProjectBindingID = networkUintPointer(*record.ProjectBindingID)
				if project, ok := projectsByID[*record.ProjectBindingID]; ok {
					item.ProjectName = project.Name
					if item.VPSBindingID == nil && project.VPSBindingID > 0 {
						item.VPSBindingID = networkUintPointer(project.VPSBindingID)
						item.VPSName = project.VPSName
					}
				}
			}
			if item.ConnectorID == nil && record.ConnectorID != nil {
				item.ConnectorID = networkUintPointer(*record.ConnectorID)
			}
		}
	}
	if item.ConnectorID != nil {
		if record, ok := connectorsByID[*item.ConnectorID]; ok {
			item.ConnectorName = record.Name
		}
	}
}

func isNetworkDomainBinding(domain ops.DomainBinding) bool {
	return domain.Enabled && domain.Role != ops.DomainRoleInternal
}

func networkDomainDesiredState(domain ops.DomainBinding) string {
	if domain.Status == ops.DomainStatusDisabled {
		return ops.NetworkStateClosed
	}
	if domain.Status == ops.DomainStatusActive {
		return ops.NetworkStateOpen
	}
	return ops.NetworkStateUnknown
}

func networkDomainObservedState(domain ops.DomainBinding) string {
	switch domain.ObservedStatus {
	case ops.DomainObservedMatched:
		return ops.NetworkStateOpen
	case ops.DomainObservedDrifted:
		return ops.NetworkStateDrifted
	case ops.DomainObservedError:
		return ops.NetworkStateError
	default:
		return ops.NetworkStateUnknown
	}
}

func networkDomainScope(domain ops.DomainBinding) string {
	if domain.Provider == ops.DomainProviderCloudflare && domain.ProxyMode == ops.DomainProxyProxied {
		return ops.NetworkScopeEdge
	}
	return ops.NetworkScopeDNS
}

func networkDomainEffectiveState(domain ops.DomainBinding) string {
	observed := networkDomainObservedState(domain)
	if observed != ops.NetworkStateUnknown {
		return observed
	}
	return networkDomainDesiredState(domain)
}

func networkOwnerName(item ops.NetworkSummaryItem) string {
	switch item.OwnerKind {
	case ops.NetworkOwnerVPS:
		return item.VPSName
	case ops.NetworkOwnerProject:
		return item.ProjectName
	case ops.NetworkOwnerDomain:
		return item.DomainName
	case ops.NetworkOwnerConnector:
		return item.ConnectorName
	default:
		return ""
	}
}

func summarizeNetworkItems(items []ops.NetworkSummaryItem) ops.NetworkSummaryCounts {
	summary := ops.NetworkSummaryCounts{
		ManagedBy: map[string]int{},
		Scopes:    map[string]int{},
	}
	vpsIDs := map[uint]struct{}{}
	for _, item := range items {
		summary.Total++
		if item.Enabled {
			summary.Enabled++
		}
		if item.Kind == ops.NetworkItemKindRule {
			summary.ExplicitRuleCount++
		} else {
			summary.InferredItemCount++
		}
		if item.VPSBindingID != nil {
			vpsIDs[*item.VPSBindingID] = struct{}{}
		}
		if item.ManagedBy != "" {
			summary.ManagedBy[item.ManagedBy]++
		}
		if item.Scope != "" {
			summary.Scopes[item.Scope]++
		}
		if isNetworkItemAttention(item) {
			summary.Attention++
		}
		if item.ObservedState == ops.NetworkStateUnknown || item.EffectiveState == ops.NetworkStateUnknown {
			summary.Unknown++
		}
	}
	summary.VPSCount = len(vpsIDs)
	return summary
}

func isNetworkItemAttention(item ops.NetworkSummaryItem) bool {
	return item.Status == ops.NetworkStatusPending ||
		item.Status == ops.NetworkStatusDrifted ||
		item.Status == ops.NetworkStatusError ||
		item.ObservedState == ops.NetworkStateUnknown ||
		item.ObservedState == ops.NetworkStateDrifted ||
		item.ObservedState == ops.NetworkStateError ||
		item.EffectiveState == ops.NetworkStateUnknown
}

func firstNetworkValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func networkUintPointer(value uint) *uint {
	if value == 0 {
		return nil
	}
	result := value
	return &result
}
