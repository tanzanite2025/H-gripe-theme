package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

type OpsDeploymentHealthCheckService struct {
	domainRepo *repository.OpsDomainBindingRepository
	httpClient *http.Client
	lookupHost func(context.Context, string) ([]string, error)
}

func NewOpsDeploymentHealthCheckService(domainRepo *repository.OpsDomainBindingRepository) *OpsDeploymentHealthCheckService {
	return &OpsDeploymentHealthCheckService{
		domainRepo: domainRepo,
		httpClient: &http.Client{Timeout: 8 * time.Second},
		lookupHost: func(ctx context.Context, host string) ([]string, error) {
			return net.DefaultResolver.LookupHost(ctx, host)
		},
	}
}

func (s *OpsDeploymentHealthCheckService) CheckProject(
	ctx context.Context,
	project *ops.ProjectBindingView,
) (*ops.DeploymentHealthCheckReport, error) {
	if s == nil || s.domainRepo == nil {
		return nil, errors.New("operations deployment health check service is not configured")
	}
	if project == nil || project.ID == 0 {
		return nil, errors.New("project binding is required for deployment health checks")
	}
	domains, err := s.domainRepo.List()
	if err != nil {
		return nil, fmt.Errorf("load domains for deployment health checks: %w", err)
	}
	report := &ops.DeploymentHealthCheckReport{
		ProjectID:   project.ID,
		Project:     project.Name,
		Environment: project.Environment,
		Status:      ops.DeploymentHealthFailed,
		CheckedAt:   time.Now().UTC(),
		Checks:      make([]ops.DeploymentHealthItem, 0),
	}
	for _, domain := range domains {
		if !isHealthCheckDomain(domain, project) {
			continue
		}
		report.Checks = append(report.Checks, s.checkDomain(ctx, domain.Domain)...)
	}
	if len(report.Checks) == 0 {
		report.Summary = "没有可用于发布后检查的已启用项目域名。"
		return report, nil
	}

	failed := 0
	degraded := 0
	for _, check := range report.Checks {
		switch check.Status {
		case "failed":
			failed++
		case "degraded":
			degraded++
		}
	}
	switch {
	case failed > 0:
		report.Status = ops.DeploymentHealthFailed
		report.Summary = fmt.Sprintf("发布后健康检查失败：%d 项失败，%d 项降级。", failed, degraded)
	case degraded > 0:
		report.Status = ops.DeploymentHealthDegraded
		report.Summary = fmt.Sprintf("发布后健康检查部分通过：%d 项降级。", degraded)
	default:
		report.Status = ops.DeploymentHealthHealthy
		report.Summary = fmt.Sprintf("发布后健康检查通过：%d 项检查。", len(report.Checks))
	}
	return report, nil
}

func (s *OpsDeploymentHealthCheckService) checkDomain(ctx context.Context, domain string) []ops.DeploymentHealthItem {
	domain = strings.TrimSpace(strings.ToLower(domain))
	if domain == "" {
		return nil
	}
	checks := make([]ops.DeploymentHealthItem, 0, 3)
	startedAt := time.Now()
	_, err := s.lookupHost(ctx, domain)
	dnsItem := ops.DeploymentHealthItem{
		Domain:     domain,
		Check:      "dns",
		Target:     domain,
		DurationMS: time.Since(startedAt).Milliseconds(),
	}
	if err != nil {
		dnsItem.Status = "failed"
		dnsItem.Message = "DNS 解析失败"
	} else {
		dnsItem.Status = "pass"
		dnsItem.Message = "DNS 已解析"
	}
	checks = append(checks, dnsItem)

	checks = append(checks, s.checkHTTP(ctx, domain, "http", false))
	checks = append(checks, s.checkHTTP(ctx, domain, "https", true))
	return checks
}

func (s *OpsDeploymentHealthCheckService) checkHTTP(ctx context.Context, domain, scheme string, requireSuccess bool) ops.DeploymentHealthItem {
	target := fmt.Sprintf("%s://%s/healthz", scheme, domain)
	item := ops.DeploymentHealthItem{
		Domain: domain,
		Check:  scheme + "_health",
		Target: target,
	}
	parsed, err := url.Parse(target)
	if err != nil {
		item.Status = "failed"
		item.Message = "健康检查地址无效"
		return item
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		item.Status = "failed"
		item.Message = "创建健康检查请求失败"
		return item
	}
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	startedAt := time.Now()
	response, err := clientCopy.Do(req)
	item.DurationMS = time.Since(startedAt).Milliseconds()
	if err != nil {
		item.Status = "failed"
		item.Message = "HTTP 健康检查请求失败"
		return item
	}
	defer response.Body.Close()
	item.StatusCode = response.StatusCode
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		item.Status = "pass"
		item.Message = "HTTP 健康检查通过"
		return item
	}
	if !requireSuccess && response.StatusCode >= 300 && response.StatusCode < 400 {
		item.Status = "degraded"
		item.Message = "HTTP 已返回跳转"
		return item
	}
	item.Status = "failed"
	item.Message = "HTTP 健康检查返回非成功状态"
	return item
}

func isHealthCheckDomain(domain ops.DomainBinding, project *ops.ProjectBindingView) bool {
	if project == nil || !domain.Enabled || domain.Environment != project.Environment {
		return false
	}
	if domain.ProjectBindingID == nil || *domain.ProjectBindingID != project.ID {
		return false
	}
	switch domain.Role {
	case ops.DomainRoleCanonical, ops.DomainRoleAlias, ops.DomainRoleAdmin:
		return strings.TrimSpace(domain.Domain) != ""
	default:
		return false
	}
}
