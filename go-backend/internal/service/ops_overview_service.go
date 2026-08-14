package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

type OpsOverviewService struct {
	domainRepo    *repository.OpsDomainBindingRepository
	connectorRepo *repository.OpsConnectorRepository
	vpsRepo       *repository.OpsVPSBindingRepository
	projectRepo   *repository.OpsProjectBindingRepository
	auditService  *AuditService
}

func NewOpsOverviewService(
	domainRepo *repository.OpsDomainBindingRepository,
	connectorRepo *repository.OpsConnectorRepository,
	vpsRepo *repository.OpsVPSBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	auditService *AuditService,
) *OpsOverviewService {
	return &OpsOverviewService{
		domainRepo:    domainRepo,
		connectorRepo: connectorRepo,
		vpsRepo:       vpsRepo,
		projectRepo:   projectRepo,
		auditService:  auditService,
	}
}

func (s *OpsOverviewService) Get() (*ops.Overview, error) {
	if s == nil || s.domainRepo == nil || s.connectorRepo == nil || s.vpsRepo == nil || s.projectRepo == nil {
		return nil, errors.New("operations overview service is not configured")
	}

	domains, err := s.domainRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations domains: %w", err)
	}
	connectors, err := s.connectorRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations connectors: %w", err)
	}
	vps, err := s.vpsRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations VPS bindings: %w", err)
	}
	projects, err := s.projectRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load operations project bindings: %w", err)
	}

	overview := &ops.Overview{
		Environment: "production",
		GeneratedAt: time.Now().UTC(),
		Summary: map[string]ops.OverviewSummary{
			"domains":    summarizeDomains(domains),
			"connectors": summarizeConnectors(connectors),
			"vps":        summarizeVPS(vps),
			"projects":   summarizeProjects(projects),
		},
		Topology: ops.OverviewTopology{
			VPS:      vps,
			Projects: projects,
			Domains:  domains,
		},
		Attention:   buildOverviewAttention(domains, connectors, vps, projects),
		RecentAudit: []audit.AuditLog{},
	}

	if s.auditService != nil {
		recent, err := s.auditService.GetRecentActivities(50)
		if err == nil {
			overview.RecentAudit = filterOpsAudit(recent, 12)
		}
	}

	return overview, nil
}

func summarizeDomains(records []ops.DomainBinding) ops.OverviewSummary {
	summary := ops.OverviewSummary{Total: len(records)}
	for _, record := range records {
		if record.Enabled {
			summary.Enabled++
		}
		if record.ObservedStatus == ops.DomainObservedUnknown {
			summary.Unknown++
		}
		if record.ObservedStatus == ops.DomainObservedMatched {
			summary.Healthy++
		}
		if isDomainAttention(record) {
			summary.Attention++
		}
	}
	return summary
}

func summarizeConnectors(records []ops.Connector) ops.OverviewSummary {
	summary := ops.OverviewSummary{Total: len(records)}
	for _, record := range records {
		if record.Enabled {
			summary.Enabled++
		}
		if strings.TrimSpace(record.CredentialsEncrypted) != "" || record.AuthType == ops.ConnectorAuthNone || record.AuthType == ops.ConnectorAuthManual {
			summary.Configured++
		}
		if record.Status == ops.ConnectorStatusPending || record.Status == ops.ConnectorStatusError {
			summary.Attention++
		}
		if record.LastTestStatus == "" {
			summary.Unknown++
		}
	}
	return summary
}

func summarizeVPS(records []ops.VPSBinding) ops.OverviewSummary {
	summary := ops.OverviewSummary{Total: len(records)}
	for _, record := range records {
		if record.Enabled {
			summary.Enabled++
		}
		if record.ObservedStatus == ops.VPSObservedUnknown {
			summary.Unknown++
		}
		if record.ObservedStatus == ops.VPSObservedHealthy {
			summary.Healthy++
		}
		if isOpsAttentionStatus(record.Status) || record.ObservedStatus == ops.VPSObservedDegraded || record.ObservedStatus == ops.VPSObservedOffline {
			summary.Attention++
		}
	}
	return summary
}

func summarizeProjects(records []ops.ProjectBindingView) ops.OverviewSummary {
	summary := ops.OverviewSummary{Total: len(records)}
	for _, record := range records {
		if record.Enabled {
			summary.Enabled++
		}
		if record.HealthStatus == ops.ProjectHealthUnknown {
			summary.Unknown++
		}
		if record.HealthStatus == ops.ProjectHealthHealthy {
			summary.Healthy++
		}
		if isOpsAttentionStatus(record.Status) || record.HealthStatus == ops.ProjectHealthDegraded || record.HealthStatus == ops.ProjectHealthOffline {
			summary.Attention++
		}
	}
	return summary
}

func buildOverviewAttention(
	domains []ops.DomainBinding,
	connectors []ops.Connector,
	vps []ops.VPSBinding,
	projects []ops.ProjectBindingView,
) []ops.OverviewResource {
	items := make([]ops.OverviewResource, 0)
	for _, record := range domains {
		if !isDomainAttention(record) {
			continue
		}
		message := "域名尚未完成实际状态同步"
		switch record.ObservedStatus {
		case ops.DomainObservedDrifted:
			message = "域名期望状态与观察状态不一致"
		case ops.DomainObservedError:
			message = record.ObservedError
			if strings.TrimSpace(message) == "" {
				message = "域名状态检查失败"
			}
		default:
			if isOpsAttentionStatus(record.Status) {
				message = "域名期望状态需要确认"
			}
		}
		items = append(items, ops.OverviewResource{
			Kind:           "domain",
			ID:             record.ID,
			Name:           record.Domain,
			Environment:    record.Environment,
			Status:         record.Status,
			ObservedStatus: record.ObservedStatus,
			Message:        message,
			Target:         firstNonEmpty(record.Target, record.ObservedTarget),
			UpdatedAt:      record.UpdatedAt,
		})
	}
	for _, record := range connectors {
		if record.Status != ops.ConnectorStatusPending && record.Status != ops.ConnectorStatusError {
			continue
		}
		message := "连接器尚未完成测试"
		if record.Status == ops.ConnectorStatusError {
			message = record.LastError
			if strings.TrimSpace(message) == "" {
				message = "连接器测试失败"
			}
		}
		items = append(items, ops.OverviewResource{
			Kind:        "connector",
			ID:          record.ID,
			Name:        record.Name,
			Environment: record.Environment,
			Status:      record.Status,
			Message:     message,
			Target:      record.Provider,
			UpdatedAt:   record.UpdatedAt,
		})
	}
	for _, record := range vps {
		if !isOpsAttentionStatus(record.Status) && record.ObservedStatus != ops.VPSObservedUnknown && record.ObservedStatus != ops.VPSObservedDegraded && record.ObservedStatus != ops.VPSObservedOffline {
			continue
		}
		message := "VPS 尚未完成实际状态同步"
		if record.ObservedStatus == ops.VPSObservedDegraded || record.ObservedStatus == ops.VPSObservedOffline {
			message = "VPS 观察状态需要处理"
		}
		if strings.TrimSpace(record.LastError) != "" {
			message = record.LastError
		}
		items = append(items, ops.OverviewResource{
			Kind:           "vps",
			ID:             record.ID,
			Name:           record.Name,
			Environment:    record.Environment,
			Status:         record.Status,
			ObservedStatus: record.ObservedStatus,
			Message:        message,
			Target:         firstNonEmpty(record.Hostname, record.IPv4),
			UpdatedAt:      record.UpdatedAt,
		})
	}
	for _, record := range projects {
		if !isOpsAttentionStatus(record.Status) && record.HealthStatus != ops.ProjectHealthUnknown && record.HealthStatus != ops.ProjectHealthDegraded && record.HealthStatus != ops.ProjectHealthOffline {
			continue
		}
		message := "项目尚未完成实际健康检查"
		if record.HealthStatus == ops.ProjectHealthDegraded || record.HealthStatus == ops.ProjectHealthOffline {
			message = "项目健康状态需要处理"
		}
		if strings.TrimSpace(record.LastError) != "" {
			message = record.LastError
		}
		items = append(items, ops.OverviewResource{
			Kind:         "project",
			ID:           record.ID,
			Name:         record.Name,
			Environment:  record.Environment,
			Status:       record.Status,
			HealthStatus: record.HealthStatus,
			Message:      message,
			Target:       firstNonEmpty(record.ComposeProjectName, record.VPSName),
			UpdatedAt:    record.UpdatedAt,
		})
	}
	return items
}

func filterOpsAudit(records []audit.AuditLog, limit int) []audit.AuditLog {
	filtered := make([]audit.AuditLog, 0, limit)
	for _, record := range records {
		if !strings.HasPrefix(record.Resource, "ops_") {
			continue
		}
		filtered = append(filtered, record)
		if len(filtered) >= limit {
			break
		}
	}
	return filtered
}

func isOpsAttentionStatus(status string) bool {
	return status == "pending" || status == "drifted" || status == "error"
}

func isDomainAttention(record ops.DomainBinding) bool {
	return isOpsAttentionStatus(record.Status) ||
		record.ObservedStatus == ops.DomainObservedUnknown ||
		record.ObservedStatus == ops.DomainObservedDrifted ||
		record.ObservedStatus == ops.DomainObservedError
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
