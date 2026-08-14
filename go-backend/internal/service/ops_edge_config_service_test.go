package service

import (
	"errors"
	"strings"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsEdgeConfigRendersProductionRoutesFromBindings(t *testing.T) {
	repos := newOpsEdgeConfigTestRepositories(t)
	project := createOpsEdgeConfigProject(t, repos, "commerce-platform", true, ops.ProjectStatusActive)

	createOpsEdgeConfigDomain(t, repos, "learn.example.com", ops.DomainRoleCanonical, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "www.example.com", ops.DomainRoleAlias, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "admin.example.com", ops.DomainRoleAdmin, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "old.example.com", ops.DomainRoleRedirect, &project.ID, "https://learn.example.com")

	config, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if config.Canonical != "learn.example.com" {
		t.Fatalf("canonical = %q, want learn.example.com", config.Canonical)
	}
	if len(config.Domains) != 4 {
		t.Fatalf("domain route count = %d, want 4", len(config.Domains))
	}
	if strings.Join(config.Nginx.StorefrontServerNames, ",") != "learn.example.com,www.example.com" {
		t.Fatalf("storefront server names = %#v", config.Nginx.StorefrontServerNames)
	}
	if strings.Join(config.Nginx.AdminServerNames, ",") != "admin.example.com" {
		t.Fatalf("admin server names = %#v", config.Nginx.AdminServerNames)
	}
	if !strings.Contains(config.Caddy, "learn.example.com {") ||
		!strings.Contains(config.Caddy, "reverse_proxy theme-web:8080") ||
		!strings.Contains(config.Caddy, "redir https://learn.example.com 308") {
		t.Fatalf("generated Caddy config does not contain expected routes:\n%s", config.Caddy)
	}
	if config.Domains[0].QuickBuyRateLimit == nil {
		t.Fatalf("generated route is missing Quick Buy rate limit policy: %#v", config.Domains[0])
	}
	if !strings.Contains(config.Caddy, "Quick Buy API rate limit: backend token buckets") ||
		strings.Contains(config.Caddy, "\trate_limit ") {
		t.Fatalf("generated Caddy config should annotate policy without unsupported directives:\n%s", config.Caddy)
	}
	if len(config.Cloudflare) != 4 {
		t.Fatalf("Cloudflare rule count = %d, want 4", len(config.Cloudflare))
	}
}

func TestOpsEdgeConfigRejectsMissingCanonical(t *testing.T) {
	repos := newOpsEdgeConfigTestRepositories(t)
	project := createOpsEdgeConfigProject(t, repos, "commerce-platform", true, ops.ProjectStatusActive)

	createOpsEdgeConfigDomain(t, repos, "www.example.com", ops.DomainRoleAlias, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "admin.example.com", ops.DomainRoleAdmin, &project.ID, "")

	_, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
	if !errors.Is(err, ErrInvalidOpsEdgeConfig) {
		t.Fatalf("Render error = %v, want ErrInvalidOpsEdgeConfig", err)
	}
	if !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("Render error = %v, want canonical diagnostic", err)
	}
}

func TestOpsEdgeConfigRejectsUnboundEnabledDomain(t *testing.T) {
	repos := newOpsEdgeConfigTestRepositories(t)
	createOpsEdgeConfigDomain(t, repos, "learn.example.com", ops.DomainRoleCanonical, nil, "")

	_, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
	if !errors.Is(err, ErrInvalidOpsEdgeConfig) {
		t.Fatalf("Render error = %v, want ErrInvalidOpsEdgeConfig", err)
	}
	if !strings.Contains(err.Error(), "not bound to a project") {
		t.Fatalf("Render error = %v, want project binding diagnostic", err)
	}
}

func TestOpsEdgeConfigRejectsDisabledProject(t *testing.T) {
	repos := newOpsEdgeConfigTestRepositories(t)
	project := createOpsEdgeConfigProject(t, repos, "commerce-platform", false, ops.ProjectStatusDisabled)
	createOpsEdgeConfigDomain(t, repos, "learn.example.com", ops.DomainRoleCanonical, &project.ID, "")

	_, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
	if !errors.Is(err, ErrInvalidOpsEdgeConfig) {
		t.Fatalf("Render error = %v, want ErrInvalidOpsEdgeConfig", err)
	}
	if !strings.Contains(err.Error(), "is disabled") {
		t.Fatalf("Render error = %v, want disabled project diagnostic", err)
	}
}

func TestOpsEdgeConfigRejectsMultipleCanonicalDomains(t *testing.T) {
	repos := newOpsEdgeConfigTestRepositories(t)
	project := createOpsEdgeConfigProject(t, repos, "commerce-platform", true, ops.ProjectStatusActive)
	createOpsEdgeConfigDomain(t, repos, "learn.example.com", ops.DomainRoleCanonical, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "shop.example.com", ops.DomainRoleCanonical, &project.ID, "")
	createOpsEdgeConfigDomain(t, repos, "admin.example.com", ops.DomainRoleAdmin, &project.ID, "")

	_, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
	if !errors.Is(err, ErrInvalidOpsEdgeConfig) {
		t.Fatalf("Render error = %v, want ErrInvalidOpsEdgeConfig", err)
	}
	if !strings.Contains(err.Error(), "multiple canonical") {
		t.Fatalf("Render error = %v, want multiple canonical diagnostic", err)
	}
}

func TestOpsEdgeConfigRejectsInvalidRedirectTarget(t *testing.T) {
	tests := []struct {
		name   string
		target string
	}{
		{name: "missing target", target: ""},
		{name: "relative target", target: "/new-location"},
		{name: "malformed target", target: "https://"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repos := newOpsEdgeConfigTestRepositories(t)
			project := createOpsEdgeConfigProject(t, repos, "commerce-platform", true, ops.ProjectStatusActive)
			createOpsEdgeConfigDomain(t, repos, "learn.example.com", ops.DomainRoleCanonical, &project.ID, "")
			createOpsEdgeConfigDomain(t, repos, "admin.example.com", ops.DomainRoleAdmin, &project.ID, "")
			createOpsEdgeConfigDomain(t, repos, "old.example.com", ops.DomainRoleRedirect, &project.ID, test.target)

			_, err := NewOpsEdgeConfigService(repos.domains, repos.projects).Render(ops.DomainEnvironmentProduction)
			if !errors.Is(err, ErrInvalidOpsEdgeConfig) {
				t.Fatalf("Render error = %v, want ErrInvalidOpsEdgeConfig", err)
			}
			if !strings.Contains(err.Error(), "redirect domain") {
				t.Fatalf("Render error = %v, want redirect diagnostic", err)
			}
		})
	}
}

type opsEdgeConfigTestRepositories struct {
	domains  *repository.OpsDomainBindingRepository
	projects *repository.OpsProjectBindingRepository
	vps      *repository.OpsVPSBindingRepository
}

func newOpsEdgeConfigTestRepositories(t *testing.T) opsEdgeConfigTestRepositories {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.DomainBinding{},
	); err != nil {
		t.Fatalf("migrate ops binding tables: %v", err)
	}

	return opsEdgeConfigTestRepositories{
		domains:  repository.NewOpsDomainBindingRepository(db),
		projects: repository.NewOpsProjectBindingRepository(db),
		vps:      repository.NewOpsVPSBindingRepository(db),
	}
}

func createOpsEdgeConfigProject(
	t *testing.T,
	repos opsEdgeConfigTestRepositories,
	name string,
	enabled bool,
	status string,
) ops.ProjectBinding {
	t.Helper()

	vps := &ops.VPSBinding{
		Name:        name + "-vps",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	if err := repos.vps.Create(vps); err != nil {
		t.Fatalf("create VPS binding: %v", err)
	}

	project := ops.ProjectBinding{
		Name:                    name,
		VPSBindingID:            vps.ID,
		Environment:             ops.ProjectEnvironmentProduction,
		GatewayNetwork:          "shared-edge",
		GatewayAlias:            "theme-web",
		Status:                  status,
		HealthStatus:            ops.ProjectHealthUnknown,
		Enabled:                 enabled,
		QuickBuyRateLimitPolicy: `{"enabled":true,"ip_requests_per_minute":90,"ip_burst":30,"session_requests_per_minute":45,"session_burst":15,"edge_ip_requests_per_minute":180,"edge_ip_burst":60,"caddy_rate_limit_enabled":false}`,
	}
	if err := repos.projects.Create(&project); err != nil {
		t.Fatalf("create project binding: %v", err)
	}
	return project
}

func createOpsEdgeConfigDomain(
	t *testing.T,
	repos opsEdgeConfigTestRepositories,
	domain string,
	role string,
	projectID *uint,
	redirectTarget string,
) {
	t.Helper()

	record := &ops.DomainBinding{
		Domain:           domain,
		ProjectBindingID: projectID,
		Role:             role,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Zone:             "example.com",
		Target:           "theme-web",
		ProxyMode:        ops.DomainProxyProxied,
		TLSMode:          ops.DomainTLSFullStrict,
		RedirectTarget:   redirectTarget,
		Status:           ops.DomainStatusActive,
		ObservedStatus:   ops.DomainObservedUnknown,
		Enabled:          true,
	}
	if err := repos.domains.Create(record); err != nil {
		t.Fatalf("create domain binding %s: %v", domain, err)
	}
}
