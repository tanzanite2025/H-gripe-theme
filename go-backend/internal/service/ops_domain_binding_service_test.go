package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsDomainBindingRequiresProjectForDeploymentRoles(t *testing.T) {
	repos := newOpsDomainBindingTestRepositories(t)
	service := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)

	enabled := true
	_, err := service.Create(OpsDomainBindingInput{
		Domain:      "learn.example.com",
		Role:        ops.DomainRoleCanonical,
		Environment: ops.DomainEnvironmentProduction,
		Provider:    ops.DomainProviderCloudflare,
		Status:      ops.DomainStatusActive,
		Enabled:     &enabled,
	})
	if err == nil {
		t.Fatal("Create without project binding error = nil, want validation error")
	}
}

func TestOpsDomainBindingValidatesProjectEnvironment(t *testing.T) {
	repos := newOpsDomainBindingTestRepositories(t)
	service := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)

	vps := &ops.VPSBinding{
		Name:        "Staging VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	if err := repos.vps.Create(vps); err != nil {
		t.Fatalf("create VPS binding: %v", err)
	}
	project := &ops.ProjectBinding{
		Name:         "staging-project",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentStaging,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	if err := repos.projects.Create(project); err != nil {
		t.Fatalf("create project binding: %v", err)
	}

	enabled := true
	_, err := service.Create(OpsDomainBindingInput{
		Domain:           "learn.example.com",
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Status:           ops.DomainStatusActive,
		Enabled:          &enabled,
	})
	if err == nil {
		t.Fatal("Create with mismatched environment error = nil, want validation error")
	}
}

func TestOpsDomainBindingUpdateCanClearConnectorAndKeepsProjectBinding(t *testing.T) {
	repos := newOpsDomainBindingTestRepositories(t)
	service := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)

	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	if err := repos.vps.Create(vps); err != nil {
		t.Fatalf("create VPS binding: %v", err)
	}
	project := &ops.ProjectBinding{
		Name:         "production-project",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	if err := repos.projects.Create(project); err != nil {
		t.Fatalf("create project binding: %v", err)
	}
	connector := &ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	if err := repos.connectors.Create(connector); err != nil {
		t.Fatalf("create connector: %v", err)
	}
	domain := &ops.DomainBinding{
		Domain:           "learn.example.com",
		ConnectorID:      &connector.ID,
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Status:           ops.DomainStatusActive,
		Enabled:          true,
	}
	if err := repos.domains.Create(domain); err != nil {
		t.Fatalf("create domain binding: %v", err)
	}

	updated, err := service.Update(domain.ID, OpsDomainBindingInput{
		Domain:              domain.Domain,
		ConnectorIDSet:      true,
		ProjectBindingIDSet: true,
		ProjectBindingID:    &project.ID,
		Role:                domain.Role,
		Environment:         domain.Environment,
		Provider:            domain.Provider,
		Status:              domain.Status,
		Enabled:             &domain.Enabled,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.ConnectorID != nil {
		t.Fatalf("connector_id = %v, want nil after explicit clear", *updated.ConnectorID)
	}
	if updated.ProjectBindingID == nil || *updated.ProjectBindingID != project.ID {
		t.Fatalf("project_binding_id = %#v, want %d", updated.ProjectBindingID, project.ID)
	}
}

func TestOpsDomainBindingUpdatePreservesOmittedReferences(t *testing.T) {
	repos := newOpsDomainBindingTestRepositories(t)
	service := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)

	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	if err := repos.vps.Create(vps); err != nil {
		t.Fatalf("create VPS binding: %v", err)
	}
	project := &ops.ProjectBinding{
		Name:         "production-project",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	if err := repos.projects.Create(project); err != nil {
		t.Fatalf("create project binding: %v", err)
	}
	connector := &ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	if err := repos.connectors.Create(connector); err != nil {
		t.Fatalf("create connector: %v", err)
	}
	domain := &ops.DomainBinding{
		Domain:           "preserve.example.com",
		ConnectorID:      &connector.ID,
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Status:           ops.DomainStatusActive,
		Enabled:          true,
	}
	if err := repos.domains.Create(domain); err != nil {
		t.Fatalf("create domain binding: %v", err)
	}

	updated, err := service.Update(domain.ID, OpsDomainBindingInput{
		Domain:      domain.Domain,
		Role:        domain.Role,
		Environment: domain.Environment,
		Provider:    domain.Provider,
		Status:      domain.Status,
		Enabled:     &domain.Enabled,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.ConnectorID == nil || *updated.ConnectorID != connector.ID {
		t.Fatalf("connector_id = %#v, want %d", updated.ConnectorID, connector.ID)
	}
	if updated.ProjectBindingID == nil || *updated.ProjectBindingID != project.ID {
		t.Fatalf("project_binding_id = %#v, want %d", updated.ProjectBindingID, project.ID)
	}
}

func TestOpsDomainBindingListForEnvironment(t *testing.T) {
	repos := newOpsDomainBindingTestRepositories(t)
	service := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)

	for _, domain := range []*ops.DomainBinding{
		{
			Domain:      "learn.example.com",
			Role:        ops.DomainRoleInternal,
			Environment: ops.DomainEnvironmentProduction,
			Provider:    ops.DomainProviderCloudflare,
			Status:      ops.DomainStatusActive,
			Enabled:     true,
		},
		{
			Domain:      "staging.learn.example.com",
			Role:        ops.DomainRoleInternal,
			Environment: ops.DomainEnvironmentStaging,
			Provider:    ops.DomainProviderCloudflare,
			Status:      ops.DomainStatusActive,
			Enabled:     true,
		},
	} {
		if err := repos.domains.Create(domain); err != nil {
			t.Fatalf("create domain binding %s: %v", domain.Domain, err)
		}
	}

	records, err := service.ListForEnvironment(ops.DomainEnvironmentStaging)
	if err != nil {
		t.Fatalf("ListForEnvironment returned error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("ListForEnvironment record count = %d, want 1", len(records))
	}
	if records[0].Domain != "staging.learn.example.com" {
		t.Fatalf("ListForEnvironment domain = %q, want staging.learn.example.com", records[0].Domain)
	}

	if _, err := service.ListForEnvironment("qa"); !errors.Is(err, ErrInvalidOpsDomainEnvironment) {
		t.Fatalf("ListForEnvironment invalid environment error = %v, want ErrInvalidOpsDomainEnvironment", err)
	}
}

type opsDomainBindingTestRepositories struct {
	domains    *repository.OpsDomainBindingRepository
	projects   *repository.OpsProjectBindingRepository
	vps        *repository.OpsVPSBindingRepository
	connectors *repository.OpsConnectorRepository
}

func newOpsDomainBindingTestRepositories(t *testing.T) opsDomainBindingTestRepositories {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&ops.Connector{},
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.DomainBinding{},
	); err != nil {
		t.Fatalf("migrate ops binding tables: %v", err)
	}
	return opsDomainBindingTestRepositories{
		domains:    repository.NewOpsDomainBindingRepository(db),
		projects:   repository.NewOpsProjectBindingRepository(db),
		vps:        repository.NewOpsVPSBindingRepository(db),
		connectors: repository.NewOpsConnectorRepository(db),
	}
}
