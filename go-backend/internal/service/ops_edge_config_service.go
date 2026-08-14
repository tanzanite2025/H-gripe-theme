package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var ErrInvalidOpsEdgeConfig = errors.New("invalid operations edge configuration")

type OpsEdgeConfigService struct {
	domainRepo  *repository.OpsDomainBindingRepository
	projectRepo *repository.OpsProjectBindingRepository
}

func NewOpsEdgeConfigService(
	domainRepo *repository.OpsDomainBindingRepository,
	projectRepo *repository.OpsProjectBindingRepository,
) *OpsEdgeConfigService {
	return &OpsEdgeConfigService{
		domainRepo:  domainRepo,
		projectRepo: projectRepo,
	}
}

func (s *OpsEdgeConfigService) Render(environment string) (*ops.EdgeConfig, error) {
	if s == nil || s.domainRepo == nil || s.projectRepo == nil {
		return nil, errors.New("operations edge config service is not configured")
	}

	environment = normalizeEdgeEnvironment(environment)
	if environment == "" {
		return nil, fmt.Errorf("%w: environment is required", ErrInvalidOpsEdgeConfig)
	}

	domains, err := s.domainRepo.List()
	if err != nil {
		return nil, err
	}
	projects, err := s.projectRepo.List()
	if err != nil {
		return nil, err
	}

	projectByID := make(map[uint]ops.ProjectBindingView, len(projects))
	for _, project := range projects {
		projectByID[project.ID] = project
	}

	routes := make([]ops.EdgeDomainRoute, 0)
	cloudflareRules := make([]ops.EdgeCloudflareRule, 0)
	storefrontNames := make([]string, 0)
	adminNames := make([]string, 0)
	canonical := ""

	for _, domain := range domains {
		if domain.Environment != environment {
			continue
		}
		if !isEdgeDomainRole(domain.Role) {
			continue
		}
		if !domain.Enabled {
			continue
		}
		if domain.Status != ops.DomainStatusActive {
			status := strings.TrimSpace(domain.Status)
			if status == "" {
				status = "unknown"
			}
			return nil, fmt.Errorf(
				"%w: enabled domain %s has non-active desired status %s",
				ErrInvalidOpsEdgeConfig,
				domain.Domain,
				status,
			)
		}

		project, ok := projectByID[nonNilUint(domain.ProjectBindingID)]
		if !ok || project.ID == 0 {
			return nil, fmt.Errorf("%w: domain %s is not bound to a project", ErrInvalidOpsEdgeConfig, domain.Domain)
		}
		if !project.Enabled || project.Status == ops.ProjectStatusDisabled {
			return nil, fmt.Errorf("%w: project %s for domain %s is disabled", ErrInvalidOpsEdgeConfig, project.Name, domain.Domain)
		}
		if project.Environment != environment {
			return nil, fmt.Errorf(
				"%w: project %s environment %s does not match domain environment %s",
				ErrInvalidOpsEdgeConfig,
				project.Name,
				project.Environment,
				environment,
			)
		}

		route := ops.EdgeDomainRoute{
			Domain:           domain.Domain,
			Role:             domain.Role,
			ProjectBindingID: project.ID,
			Project:          project.Name,
			GatewayAlias:     strings.TrimSpace(project.GatewayAlias),
		}
		if domain.Role == ops.DomainRoleRedirect {
			route.RedirectTarget = strings.TrimSpace(domain.RedirectTarget)
			if route.RedirectTarget == "" {
				return nil, fmt.Errorf("%w: redirect domain %s has no redirect target", ErrInvalidOpsEdgeConfig, domain.Domain)
			}
			parsedTarget, err := url.ParseRequestURI(route.RedirectTarget)
			if err != nil ||
				parsedTarget == nil ||
				parsedTarget.Host == "" ||
				(parsedTarget.Scheme != "http" && parsedTarget.Scheme != "https") {
				return nil, fmt.Errorf("%w: redirect domain %s has an invalid redirect target", ErrInvalidOpsEdgeConfig, domain.Domain)
			}
		} else {
			route.Upstream, err = edgeUpstream(route.GatewayAlias)
			if err != nil {
				return nil, fmt.Errorf("%w: domain %s: %v", ErrInvalidOpsEdgeConfig, domain.Domain, err)
			}
			policy, err := ops.ParseQuickBuyRateLimitPolicy(project.QuickBuyRateLimitPolicy)
			if err != nil {
				return nil, fmt.Errorf("%w: project %s quick buy rate limit policy is invalid: %v", ErrInvalidOpsEdgeConfig, project.Name, err)
			}
			route.QuickBuyRateLimit = &policy
		}
		routes = append(routes, route)

		if domain.Role == ops.DomainRoleCanonical {
			if canonical != "" {
				return nil, fmt.Errorf("%w: multiple canonical domains are enabled", ErrInvalidOpsEdgeConfig)
			}
			canonical = domain.Domain
		}
		switch domain.Role {
		case ops.DomainRoleCanonical, ops.DomainRoleAlias:
			storefrontNames = append(storefrontNames, domain.Domain)
		case ops.DomainRoleAdmin:
			adminNames = append(adminNames, domain.Domain)
		}

		if domain.Provider == ops.DomainProviderCloudflare {
			cloudflareRules = append(cloudflareRules, ops.EdgeCloudflareRule{
				Domain:        domain.Domain,
				Role:          domain.Role,
				Zone:          strings.TrimSpace(domain.Zone),
				Target:        strings.TrimSpace(domain.Target),
				ProxyMode:     domain.ProxyMode,
				TLSMode:       domain.TLSMode,
				Status:        domain.Status,
				ObservedState: domain.ObservedStatus,
				Enabled:       domain.Enabled,
			})
		}
	}

	if canonical == "" {
		return nil, fmt.Errorf("%w: no active canonical domain is configured", ErrInvalidOpsEdgeConfig)
	}
	if len(storefrontNames) == 0 {
		return nil, fmt.Errorf("%w: no active storefront domain is configured", ErrInvalidOpsEdgeConfig)
	}
	if len(adminNames) == 0 {
		return nil, fmt.Errorf("%w: no active admin domain is configured", ErrInvalidOpsEdgeConfig)
	}

	sort.Slice(routes, func(i, j int) bool { return routes[i].Domain < routes[j].Domain })
	sort.Strings(storefrontNames)
	sort.Strings(adminNames)
	sort.Slice(cloudflareRules, func(i, j int) bool { return cloudflareRules[i].Domain < cloudflareRules[j].Domain })

	config := &ops.EdgeConfig{
		SchemaVersion: ops.EdgeConfigSchemaVersion,
		Environment:   environment,
		GeneratedAt:   time.Now().UTC(),
		Canonical:     canonical,
		Domains:       routes,
		Cloudflare:    cloudflareRules,
		Caddy:         renderCaddy(routes),
		Nginx: ops.EdgeNginxConfig{
			StorefrontServerNames: storefrontNames,
			AdminServerNames:      adminNames,
		},
	}
	return config, nil
}

func (s *OpsEdgeConfigService) RenderToDirectory(environment, outputDir string) (*ops.EdgeConfig, error) {
	config, err := s.Render(environment)
	if err != nil {
		return nil, err
	}
	if err := writeEdgeConfigFiles(config, outputDir); err != nil {
		return nil, err
	}
	return config, nil
}

func writeEdgeConfigFiles(config *ops.EdgeConfig, outputDir string) error {
	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return fmt.Errorf("%w: output directory is required", ErrInvalidOpsEdgeConfig)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create edge config directory: %w", err)
	}

	files := map[string][]byte{}
	manifest, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("encode edge config manifest: %w", err)
	}
	files["manifest.json"] = append(manifest, '\n')
	files["caddy.caddy"] = []byte(config.Caddy)
	files["cloudflare.json"], err = json.MarshalIndent(config.Cloudflare, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Cloudflare edge rules: %w", err)
	}
	files["cloudflare.json"] = append(files["cloudflare.json"], '\n')
	files["storefront-server-names.conf"] = []byte(renderNginxServerNames(config.Nginx.StorefrontServerNames))
	files["admin-server-names.conf"] = []byte(renderNginxServerNames(config.Nginx.AdminServerNames))

	stagingDir, err := os.MkdirTemp(outputDir, ".edge-config-staging-*")
	if err != nil {
		return fmt.Errorf("create edge config staging directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(stagingDir)
	}()

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if err := writeAtomic(filepath.Join(stagingDir, name), files[name]); err != nil {
			return err
		}
	}

	for _, name := range names {
		contents, err := os.ReadFile(filepath.Join(stagingDir, name))
		if err != nil {
			return fmt.Errorf("read staged edge config %s: %w", name, err)
		}
		if err := writeAtomic(filepath.Join(outputDir, name), contents); err != nil {
			return err
		}
	}
	return nil
}

func writeAtomic(path string, contents []byte) error {
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, contents, 0o644); err != nil {
		return fmt.Errorf("write generated edge config %s: %w", filepath.Base(path), err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("publish generated edge config %s: %w", filepath.Base(path), err)
	}
	return nil
}

func renderCaddy(routes []ops.EdgeDomainRoute) string {
	var builder strings.Builder
	builder.WriteString("# Generated from ops_domain_bindings. Do not edit manually.\n\n")
	for _, route := range routes {
		builder.WriteString(route.Domain)
		builder.WriteString(" {\n")
		if route.Role == ops.DomainRoleRedirect {
			fmt.Fprintf(&builder, "\tredir %s 308\n", route.RedirectTarget)
		} else {
			builder.WriteString("\tencode zstd gzip\n")
			writeQuickBuyRateLimitCaddyComment(&builder, route)
			fmt.Fprintf(&builder, "\treverse_proxy %s\n", route.Upstream)
		}
		builder.WriteString("}\n\n")
	}
	return builder.String()
}

func writeQuickBuyRateLimitCaddyComment(builder *strings.Builder, route ops.EdgeDomainRoute) {
	if builder == nil || route.QuickBuyRateLimit == nil {
		return
	}
	policy := *route.QuickBuyRateLimit
	if !policy.Enabled {
		builder.WriteString("\t# Quick Buy API rate limit is disabled in Ops project policy.\n")
		return
	}
	fmt.Fprintf(
		builder,
		"\t# Quick Buy API rate limit: backend token buckets ip=%d/min burst=%d, session=%d/min burst=%d.\n",
		policy.IPRequestsPerMinute,
		policy.IPBurst,
		policy.SessionRequestsPerMinute,
		policy.SessionBurst,
	)
	if policy.CaddyRateLimitEnabled {
		fmt.Fprintf(
			builder,
			"\t# Caddy edge rate limit requested by Ops policy: ip=%d/min burst=%d. Standard gateway images must include a compatible rate-limit module before rendering executable directives.\n",
			policy.EdgeIPRequestsPerMinute,
			policy.EdgeIPBurst,
		)
	} else {
		builder.WriteString("\t# Caddy edge rate limit directives are not rendered because the shared gateway has no required plugin contract.\n")
	}
}

func renderNginxServerNames(domains []string) string {
	return fmt.Sprintf("# Generated from ops_domain_bindings. Do not edit manually.\nserver_name %s;\n", strings.Join(domains, " "))
}

func edgeUpstream(alias string) (string, error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", errors.New("project gateway alias is required")
	}
	if strings.ContainsAny(alias, " \t\r\n{}();") {
		return "", errors.New("project gateway alias contains invalid characters")
	}
	if strings.Contains(alias, "://") || strings.Contains(alias, "/") {
		return "", errors.New("project gateway alias must be a host or host:port")
	}
	if strings.Contains(alias, ":") {
		return alias, nil
	}
	return alias + ":8080", nil
}

func normalizeEdgeEnvironment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case ops.DomainEnvironmentProduction, ops.DomainEnvironmentStaging, ops.DomainEnvironmentTest, ops.DomainEnvironmentLocal:
		return value
	default:
		return ""
	}
}

func isEdgeDomainRole(role string) bool {
	switch role {
	case ops.DomainRoleCanonical, ops.DomainRoleAlias, ops.DomainRoleAdmin, ops.DomainRoleRedirect:
		return true
	default:
		return false
	}
}

func nonNilUint(value *uint) uint {
	if value == nil {
		return 0
	}
	return *value
}
