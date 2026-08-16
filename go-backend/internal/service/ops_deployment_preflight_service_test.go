package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPreflightBoundaryChecksRequireProductionServicesAndNetworks(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			Services: "db, redis, migrate, edge-config, api, storefront, admin, web",
			Networks: "db, cache, app, shared-edge",
		},
	}

	if check := checkServices(project); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkServices status = %q, want pass: %#v", check.Status, check)
	}
	if check := checkNetworks(project); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkNetworks status = %q, want pass: %#v", check.Status, check)
	}

	project.Services = "db, redis, migrate, api, storefront, admin, web"
	if check := checkServices(project); check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("checkServices missing admin status = %q, want block", check.Status)
	}
}

func TestPreflightRemoteHealthRequiresObservedHealthyContainers(t *testing.T) {
	now := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			HealthStatus:           ops.ProjectHealthHealthy,
			LastCheckedAt:          &now,
			ObservedContainerCount: 6,
			ObservedRunningCount:   6,
			ObservedHealthyCount:   6,
		},
	}

	if check := checkRemoteHealth(project); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkRemoteHealth status = %q, want pass: %#v", check.Status, check)
	}

	project.ObservedHealthyCount = 5
	if check := checkRemoteHealth(project); check.Status != ops.DeploymentCheckWarning {
		t.Fatalf("checkRemoteHealth partial evidence status = %q, want warning", check.Status)
	}

	project.HealthStatus = ops.ProjectHealthUnknown
	if check := checkRemoteHealth(project); check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("checkRemoteHealth unknown status = %q, want block", check.Status)
	}
}

func TestPreflightBackupsWarnWhenRestoreExerciseIsPending(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			BackupPolicy: "Daily PostgreSQL and uploads backups with off-VPS copy.",
			RestoreNotes: "恢复演练结果待补录。",
		},
	}

	if check := checkBackups(project); check.Status != ops.DeploymentCheckWarning {
		t.Fatalf("checkBackups pending restore status = %q, want warning", check.Status)
	}

	project.RestoreNotes = "2026-08-13 已完成恢复演练，PostgreSQL 与 uploads 校验通过。"
	if check := checkBackups(project); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkBackups complete restore status = %q, want pass", check.Status)
	}
}

func TestPreflightHostingerConnectorAcceptsEnvironmentCredentials(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")

	connectorID := uint(7)
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ConnectorID: &connectorID,
		},
	}
	connector := &ops.Connector{
		ID:            7,
		Name:          "Hostinger Production",
		Provider:      ops.ConnectorProviderHostinger,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Scopes:        "vps:read,project:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}

	if check := checkHostingerConnector(project, connector, nil); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkHostingerConnector status = %q, want pass: %#v", check.Status, check)
	}
}

func TestPreflightUsesVPSConnectorWhenProjectConnectorIsOmitted(t *testing.T) {
	vpsConnectorID := uint(8)
	vps := &ops.VPSBinding{
		Environment: ops.VPSEnvironmentProduction,
		ConnectorID: &vpsConnectorID,
	}
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			Environment: ops.ProjectEnvironmentProduction,
		},
	}

	if got := effectiveProjectConnectorID(project, vps); got == nil || *got != vpsConnectorID {
		t.Fatalf("effectiveProjectConnectorID = %#v, want VPS connector %d", got, vpsConnectorID)
	}
	if got := projectConnectorBindingSource(project, vps); got != "vps" {
		t.Fatalf("projectConnectorBindingSource = %q, want vps", got)
	}
}

func TestPreflightReportsExplicitProjectConnectorSource(t *testing.T) {
	projectConnectorID := uint(9)
	vpsConnectorID := uint(8)
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			Environment: ops.ProjectEnvironmentProduction,
			ConnectorID: &projectConnectorID,
		},
	}
	vps := &ops.VPSBinding{ConnectorID: &vpsConnectorID}
	connector := &ops.Connector{
		ID:            projectConnectorID,
		Name:          "Hostinger Production",
		Provider:      ops.ConnectorProviderHostinger,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Scopes:        "vps:read,project:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")

	check := checkHostingerConnectorWithSource(project, connector, nil, projectConnectorBindingSource(project, vps))
	if check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkHostingerConnectorWithSource status = %q, want pass: %#v", check.Status, check)
	}
	if !strings.Contains(check.Detail, "项目显式绑定") {
		t.Fatalf("check.Detail = %q, want explicit project connector evidence", check.Detail)
	}
}

func TestPreflightImageCommitConsistencyBlocksMismatchedSHATag(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			CurrentImageTag:  "sha-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			CurrentCommitSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
	}

	check := checkImageCommitConsistency(project)
	if check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("checkImageCommitConsistency status = %q, want block: %#v", check.Status, check)
	}
}

func TestPreflightSummaryKeepsProjectCounts(t *testing.T) {
	generatedAt := time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)
	report := &ops.DeploymentPreflight{
		ProjectID:     42,
		Project:       "theme",
		Environment:   ops.ProjectEnvironmentProduction,
		GeneratedAt:   generatedAt,
		Ready:         false,
		StatusLevel:   "blocked",
		BlockingCount: 2,
		WarningCount:  1,
		PassCount:     12,
		InfoCount:     1,
		Summary:       "存在 2 项阻断问题。",
		NextActions: []string{
			"Compose 来源：补齐或确认部署使用的 Compose 来源。",
			"镜像标签：登记不可变镜像标签或 digest。",
		},
		Categories: []ops.DeploymentPreflightGroup{
			{
				Category:      "compose",
				Label:         "Compose",
				TotalCount:    1,
				BlockingCount: 1,
			},
		},
		Checks: []ops.DeploymentPreflightCheck{
			{
				Label:   "Compose 来源",
				Status:  ops.DeploymentCheckBlock,
				Message: "未登记 Compose 来源",
			},
			{
				Label:   "镜像标签",
				Status:  ops.DeploymentCheckWarning,
				Message: "当前镜像标签不是不可变 SHA 标签或 digest。",
			},
		},
	}

	summary := summarizeDeploymentPreflight(report)
	if summary.ProjectID != report.ProjectID || summary.Project != report.Project || summary.Environment != report.Environment {
		t.Fatalf("summary identity mismatch: %#v", summary)
	}
	if summary.Ready || summary.StatusLevel != "blocked" || summary.BlockingCount != 2 || summary.WarningCount != 1 || summary.PassCount != 12 || summary.InfoCount != 1 {
		t.Fatalf("summary counts mismatch: %#v", summary)
	}
	if !summary.GeneratedAt.Equal(generatedAt) {
		t.Fatalf("summary generated_at = %s, want %s", summary.GeneratedAt, generatedAt)
	}
	if len(summary.BlockReasons) != 1 || summary.BlockReasons[0] != "Compose 来源：未登记 Compose 来源" {
		t.Fatalf("summary block reasons = %#v", summary.BlockReasons)
	}
	if len(summary.WarnReasons) != 1 || summary.WarnReasons[0] != "镜像标签：当前镜像标签不是不可变 SHA 标签或 digest。" {
		t.Fatalf("summary warn reasons = %#v", summary.WarnReasons)
	}
	if len(summary.NextActions) != 2 || summary.NextActions[0] != "Compose 来源：补齐或确认部署使用的 Compose 来源。" {
		t.Fatalf("summary next actions = %#v", summary.NextActions)
	}
	if len(summary.Categories) != 1 || summary.Categories[0].Category != "compose" {
		t.Fatalf("summary categories = %#v", summary.Categories)
	}
}

func TestPreflightStatusLevelAndNextActions(t *testing.T) {
	checks := []ops.DeploymentPreflightCheck{
		blockedCheck("domains", "domain", "域名边界", "域名状态存在阻断问题。", "learn.gripe drifted"),
		warningCheck("deployment_record", "evidence", "部署记录", "未登记最近一次部署时间。", ""),
	}

	if level := preflightStatusLevel(1, 1); level != ops.DeploymentStatusBlocked {
		t.Fatalf("preflightStatusLevel blocked = %q, want blocked", level)
	}
	if level := preflightStatusLevel(0, 1); level != ops.DeploymentStatusReview {
		t.Fatalf("preflightStatusLevel warning = %q, want review", level)
	}
	if level := preflightStatusLevel(0, 0); level != ops.DeploymentStatusReady {
		t.Fatalf("preflightStatusLevel clean = %q, want ready", level)
	}

	actions := preflightNextActions(checks)
	if len(actions) != 2 {
		t.Fatalf("preflightNextActions len = %d, want 2: %#v", len(actions), actions)
	}
	if actions[0] != "域名边界：修正域名目标、启用状态或刷新 Cloudflare 观察结果。" {
		t.Fatalf("first action = %q", actions[0])
	}
	if actions[1] != "部署记录：补录最近一次部署时间。" {
		t.Fatalf("second action = %q", actions[1])
	}
}

func TestPreflightReviewIsNotReady(t *testing.T) {
	report := &ops.DeploymentPreflight{
		BlockingCount: 0,
		WarningCount:  1,
	}
	report.StatusLevel = preflightStatusLevel(report.BlockingCount, report.WarningCount)
	report.Ready = report.StatusLevel == ops.DeploymentStatusReady

	if report.StatusLevel != ops.DeploymentStatusReview {
		t.Fatalf("status level = %q, want review", report.StatusLevel)
	}
	if report.Ready {
		t.Fatal("review report ready = true, want false")
	}
}

func TestPreflightVPSConnectorRequiresTheVPSBindingConnector(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	vpsConnectorID := uint(8)
	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Environment: ops.VPSEnvironmentProduction,
		ConnectorID: &vpsConnectorID,
	}
	projectConnector := &ops.Connector{
		ID:            vpsConnectorID,
		Name:          "Project Hostinger",
		Provider:      ops.ConnectorProviderHostinger,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
		Scopes:        "vps:read,project:read",
	}
	vpsConnector := *projectConnector
	vpsConnector.ID = vpsConnectorID
	vpsConnector.Name = "VPS Hostinger"

	if check := checkVPSConnector(vps, &vpsConnector, nil); check.Status != ops.DeploymentCheckWarning && check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkVPSConnector status = %q, want warning or pass: %#v", check.Status, check)
	}
	if check := checkVPSConnector(vps, nil, repository.ErrRecordNotFound); check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("checkVPSConnector missing status = %q, want block: %#v", check.Status, check)
	}
}

func TestPreflightProductionVolumeBaselineAcceptsResolvedAndLogicalNames(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			Environment:        ops.ProjectEnvironmentProduction,
			ComposeProjectName: "commerce-platform",
			Volumes:            "commerce-platform-postgres-data, commerce-platform-redis-data, uploads",
		},
	}

	check := checkVolumes(project)
	if check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkVolumes status = %q, want pass: %#v", check.Status, check)
	}
}

func TestPreflightNonProductionBoundaryBaselineNeedsReview(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			Environment: ops.ProjectEnvironmentStaging,
			Services:    "api",
		},
	}

	check := checkServices(project)
	if check.Status != ops.DeploymentCheckWarning {
		t.Fatalf("checkServices staging status = %q, want warning: %#v", check.Status, check)
	}
}

func TestPreflightGroupSummarySortsByRisk(t *testing.T) {
	groups := summarizePreflightGroups([]ops.DeploymentPreflightCheck{
		blockedCheck("domains", "domain", "域名边界", "域名状态存在阻断问题。", ""),
		blockedCheck("remote_health", "runtime", "远端观察健康", "尚未获得远端项目健康观察结果。", ""),
		warningCheck("ghcr_connector", "connector", "镜像连接器", "未登记 GHCR/GitHub 只读连接器。", ""),
		passCheck("compose_source", "compose", "Compose 来源", "compose.prod.yml", ""),
		infoCheck("image_commit", "version", "镜像 / Commit 一致性", "镜像标签与 Commit 暂无可校验关系。", ""),
	})

	if len(groups) != 5 {
		t.Fatalf("summarizePreflightGroups len = %d, want 5: %#v", len(groups), groups)
	}
	if groups[0].Category != "domain" || groups[0].BlockingCount != 1 {
		t.Fatalf("first group = %#v, want domain block", groups[0])
	}
	if groups[1].Category != "runtime" || groups[1].BlockingCount != 1 {
		t.Fatalf("second group = %#v, want runtime block", groups[1])
	}
	if groups[2].Category != "connector" || groups[2].WarningCount != 1 {
		t.Fatalf("third group = %#v, want connector warning", groups[2])
	}
}

func TestPreflightLookupConnectorSnapshotUsesLoadedConnectors(t *testing.T) {
	connectorID := uint(3)
	connectors := []ops.Connector{
		{ID: 2, Name: "Cloudflare", Provider: ops.ConnectorProviderCloudflare},
		{ID: 3, Name: "Hostinger", Provider: ops.ConnectorProviderHostinger},
	}

	connector, err := lookupConnectorSnapshot(connectors, &connectorID)
	if err != nil {
		t.Fatalf("lookupConnectorSnapshot returned error: %v", err)
	}
	if connector == nil || connector.Name != "Hostinger" {
		t.Fatalf("connector = %#v, want Hostinger", connector)
	}

	missingID := uint(99)
	if _, err := lookupConnectorSnapshot(connectors, &missingID); err == nil {
		t.Fatal("lookupConnectorSnapshot missing connector error = nil, want error")
	}
}

func TestPreflightDomainsBlockObservedDrift(t *testing.T) {
	projectID := uint(42)
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:             projectID,
			Environment:    ops.ProjectEnvironmentProduction,
			GatewayAlias:   "theme-web",
			GatewayNetwork: "shared-edge",
		},
	}

	check := checkDomains(project, []ops.DomainBinding{
		{
			Domain:           "learn.gripe",
			ProjectBindingID: &projectID,
			Role:             ops.DomainRoleCanonical,
			Environment:      ops.DomainEnvironmentProduction,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedDrifted,
			Enabled:          true,
		},
	})
	if check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("checkDomains drift status = %q, want block: %#v", check.Status, check)
	}
}

func TestPreflightDomainsPassMatchedGatewayDomains(t *testing.T) {
	projectID := uint(42)
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:           projectID,
			Environment:  ops.ProjectEnvironmentProduction,
			GatewayAlias: "theme-web",
		},
	}

	check := checkDomains(project, []ops.DomainBinding{
		{
			Domain:           "learn.gripe",
			ProjectBindingID: &projectID,
			Role:             ops.DomainRoleCanonical,
			Environment:      ops.DomainEnvironmentProduction,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedMatched,
			Enabled:          true,
		},
		{
			Domain:           "admin.learn.gripe",
			ProjectBindingID: &projectID,
			Role:             ops.DomainRoleAdmin,
			Environment:      ops.DomainEnvironmentProduction,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedMatched,
			Enabled:          true,
		},
	})
	if check.Status != ops.DeploymentCheckPass {
		t.Fatalf("checkDomains matched status = %q, want pass: %#v", check.Status, check)
	}
}

func TestPreflightDomainsDoNotLeakDriftAcrossProjectsInTheSameEnvironment(t *testing.T) {
	projectAID := uint(41)
	projectBID := uint(42)
	projectA := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:                 projectAID,
			Environment:        ops.ProjectEnvironmentProduction,
			ComposeProjectName: "project-a",
			GatewayAlias:       "theme-web",
		},
	}
	projectB := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:                 projectBID,
			Environment:        ops.ProjectEnvironmentProduction,
			ComposeProjectName: "project-b",
			GatewayAlias:       "theme-web",
		},
	}
	domains := []ops.DomainBinding{
		{
			Domain:           "a.example.com",
			ProjectBindingID: &projectAID,
			Role:             ops.DomainRoleCanonical,
			Environment:      ops.DomainEnvironmentProduction,
			Provider:         ops.DomainProviderCloudflare,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedMatched,
			Enabled:          true,
		},
		{
			Domain:           "b.example.com",
			ProjectBindingID: &projectBID,
			Role:             ops.DomainRoleCanonical,
			Environment:      ops.DomainEnvironmentProduction,
			Provider:         ops.DomainProviderCloudflare,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedDrifted,
			Enabled:          true,
		},
	}
	connectors := []ops.Connector{
		{
			ID:            1,
			Name:          "Cloudflare A",
			Provider:      ops.ConnectorProviderCloudflare,
			Environment:   ops.ConnectorEnvironmentProduction,
			AuthType:      ops.ConnectorAuthEnvironment,
			CredentialRef: "CLOUDFLARE_A_TOKEN",
			Scopes:        "zones:read,dns:read",
			Status:        ops.ConnectorStatusActive,
			Enabled:       true,
		},
		{
			ID:          2,
			Name:        "Cloudflare B",
			Provider:    ops.ConnectorProviderCloudflare,
			Environment: ops.ConnectorEnvironmentProduction,
			Scopes:      "zones:read",
			Status:      ops.ConnectorStatusError,
			Enabled:     false,
		},
	}
	domains[0].ConnectorID = deploymentTestUintPointer(1)
	domains[1].ConnectorID = deploymentTestUintPointer(2)

	if check := checkDomains(projectA, domains); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("project A domain status = %q, want pass: %#v", check.Status, check)
	}
	if check := checkDomains(projectB, domains); check.Status != ops.DeploymentCheckBlock {
		t.Fatalf("project B domain status = %q, want block: %#v", check.Status, check)
	}
	t.Setenv("CLOUDFLARE_A_TOKEN", "configured-token")
	if check := checkCloudflareConnectors(connectors, projectA, domains); check.Status != ops.DeploymentCheckPass {
		t.Fatalf("project A Cloudflare connector status = %q, want pass: %#v", check.Status, check)
	}
	if check := checkCloudflareConnectors(connectors, projectB, domains); check.Status != ops.DeploymentCheckWarning {
		t.Fatalf("project B Cloudflare connector status = %q, want warning: %#v", check.Status, check)
	}
}

func TestPreflightDomainsTreatMatchingLegacyUnboundDomainsAsReview(t *testing.T) {
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:                 42,
			Environment:        ops.ProjectEnvironmentProduction,
			ComposeProjectName: "commerce-platform",
			GatewayAlias:       "theme-web",
		},
	}
	check := checkDomains(project, []ops.DomainBinding{
		{
			Domain:         "legacy.example.com",
			Role:           ops.DomainRoleCanonical,
			Environment:    ops.DomainEnvironmentProduction,
			Target:         "theme-web:8080",
			Status:         ops.DomainStatusActive,
			ObservedStatus: ops.DomainObservedMatched,
			Enabled:        true,
		},
	})
	if check.Status != ops.DeploymentCheckWarning {
		t.Fatalf("legacy unbound domain status = %q, want warning: %#v", check.Status, check)
	}
}

func TestDeploymentPreflightServiceUsesEffectiveVPSConnector(t *testing.T) {
	repos := newDeploymentPreflightTestRepositories(t)
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	t.Setenv("CLOUDFLARE_API_TOKEN", "configured-token")

	hostinger := &ops.Connector{
		Name:          "Hostinger Production",
		Provider:      ops.ConnectorProviderHostinger,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Scopes:        "vps:read,project:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	cloudflare := &ops.Connector{
		Name:          "Cloudflare Production",
		Provider:      ops.ConnectorProviderCloudflare,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "CLOUDFLARE_API_TOKEN",
		Scopes:        "zones:read,dns:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	if err := repos.connectors.Create(hostinger); err != nil {
		t.Fatalf("create Hostinger connector: %v", err)
	}
	if err := repos.connectors.Create(cloudflare); err != nil {
		t.Fatalf("create Cloudflare connector: %v", err)
	}

	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	vps := &ops.VPSBinding{
		Name:               "Production VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ConnectorID:        &hostinger.ID,
		ProviderResourceID: "vm-1",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		ObservedState:      "running",
		ObservedSource:     "hostinger:Hostinger Production",
		Hostname:           "vps.example.com",
		ObservedHostname:   "vps.example.com",
		IPv4:               "203.0.113.10",
		ObservedIPv4:       "203.0.113.10",
		LastObservedAt:     &now,
		Enabled:            true,
	}
	if err := repos.vps.Create(vps); err != nil {
		t.Fatalf("create VPS binding: %v", err)
	}

	project := &ops.ProjectBinding{
		Name:                   "commerce-platform",
		VPSBindingID:           vps.ID,
		Environment:            ops.ProjectEnvironmentProduction,
		ComposeSource:          "compose.prod.yml",
		ComposeProjectName:     "commerce-platform",
		GatewayNetwork:         "shared-edge",
		GatewayAlias:           "theme-web",
		Services:               "db, redis, migrate, edge-config, api, storefront, admin, web",
		Networks:               "db, cache, app, shared-edge",
		Volumes:                "commerce-platform-postgres-data, commerce-platform-redis-data, uploads",
		CurrentImageTag:        "sha-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		CurrentCommitSHA:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Status:                 ops.ProjectStatusActive,
		HealthStatus:           ops.ProjectHealthHealthy,
		ObservedContainerCount: 7,
		ObservedRunningCount:   7,
		ObservedHealthyCount:   7,
		LastCheckedAt:          &now,
		LastDeploymentAt:       &now,
		BackupPolicy:           "Daily PostgreSQL and uploads backups with off-VPS copy.",
		RestoreNotes:           "2026-08-13 已完成恢复演练，PostgreSQL 与 uploads 校验通过。",
		Enabled:                true,
	}
	if err := repos.projects.Create(project); err != nil {
		t.Fatalf("create project binding: %v", err)
	}
	domain := &ops.DomainBinding{
		Domain:           "learn.gripe",
		ConnectorID:      &cloudflare.ID,
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Target:           "theme-web:8080",
		Status:           ops.DomainStatusActive,
		ObservedStatus:   ops.DomainObservedMatched,
		Enabled:          true,
	}
	if err := repos.domains.Create(domain); err != nil {
		t.Fatalf("create domain binding: %v", err)
	}

	report, err := NewOpsDeploymentPreflightService(
		repos.projects,
		repos.vps,
		repos.connectors,
		repos.domains,
	).EvaluateProject(project.ID)
	if err != nil {
		t.Fatalf("EvaluateProject returned error: %v", err)
	}
	if report.StatusLevel != ops.DeploymentStatusReview {
		t.Fatalf("status level = %q, want review: %#v", report.StatusLevel, report.Checks)
	}
	if report.Ready {
		t.Fatal("review report ready = true, want false")
	}
	if report.BlockingCount != 0 {
		t.Fatalf("blocking count = %d, want 0: %#v", report.BlockingCount, report.Checks)
	}
	if findPreflightCheck(report.Checks, "hostinger_connector").Status != ops.DeploymentCheckPass {
		t.Fatalf("effective VPS connector was not accepted: %#v", findPreflightCheck(report.Checks, "hostinger_connector"))
	}
	if findPreflightCheck(report.Checks, "vps_connector").Status != ops.DeploymentCheckPass {
		t.Fatalf("VPS connector check did not pass: %#v", findPreflightCheck(report.Checks, "vps_connector"))
	}
}

func TestDeploymentPreflightOverviewPreservesStatusesAndMissingVPSEvidence(t *testing.T) {
	repos := newDeploymentPreflightTestRepositories(t)
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	t.Setenv("CLOUDFLARE_API_TOKEN", "configured-token")

	hostinger := &ops.Connector{
		Name:          "Hostinger Production",
		Provider:      ops.ConnectorProviderHostinger,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Scopes:        "vps:read,project:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	cloudflare := &ops.Connector{
		Name:          "Cloudflare Production",
		Provider:      ops.ConnectorProviderCloudflare,
		Environment:   ops.ConnectorEnvironmentProduction,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "CLOUDFLARE_API_TOKEN",
		Scopes:        "zones:read,dns:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	if err := repos.connectors.Create(hostinger); err != nil {
		t.Fatalf("create Hostinger connector: %v", err)
	}
	if err := repos.connectors.Create(cloudflare); err != nil {
		t.Fatalf("create Cloudflare connector: %v", err)
	}

	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)
	healthyVPS := &ops.VPSBinding{
		Name:               "Healthy VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ConnectorID:        &hostinger.ID,
		ProviderResourceID: "vm-healthy",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		ObservedState:      "running",
		ObservedSource:     "hostinger",
		LastObservedAt:     &now,
		Enabled:            true,
	}
	missingConnectorVPS := &ops.VPSBinding{
		Name:               "Unconnected VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ProviderResourceID: "vm-unconnected",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		ObservedState:      "running",
		ObservedSource:     "hostinger",
		LastObservedAt:     &now,
		Enabled:            true,
	}
	if err := repos.vps.Create(healthyVPS); err != nil {
		t.Fatalf("create healthy VPS: %v", err)
	}
	if err := repos.vps.Create(missingConnectorVPS); err != nil {
		t.Fatalf("create unconnected VPS: %v", err)
	}

	reviewProject := newPreflightFixtureProject("review-project", healthyVPS.ID, now)
	reviewProject.ProviderResourceID = ""
	blockedProject := newPreflightFixtureProject("blocked-project", missingConnectorVPS.ID, now)
	missingVPSProject := newPreflightFixtureProject("missing-vps-project", 9999, now)
	if err := repos.projects.Create(reviewProject); err != nil {
		t.Fatalf("create review project: %v", err)
	}
	if err := repos.projects.Create(blockedProject); err != nil {
		t.Fatalf("create blocked project: %v", err)
	}
	if err := repos.projects.Create(missingVPSProject); err != nil {
		t.Fatalf("create missing VPS project: %v", err)
	}

	for _, domain := range []*ops.DomainBinding{
		{
			Domain:           "review.example.com",
			ConnectorID:      &cloudflare.ID,
			ProjectBindingID: &reviewProject.ID,
			Role:             ops.DomainRoleCanonical,
			Environment:      ops.DomainEnvironmentProduction,
			Provider:         ops.DomainProviderCloudflare,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedMatched,
			Enabled:          true,
		},
		{
			Domain:           "blocked.example.com",
			ConnectorID:      &cloudflare.ID,
			ProjectBindingID: &blockedProject.ID,
			Role:             ops.DomainRoleAlias,
			Environment:      ops.DomainEnvironmentProduction,
			Provider:         ops.DomainProviderCloudflare,
			Target:           "theme-web:8080",
			Status:           ops.DomainStatusActive,
			ObservedStatus:   ops.DomainObservedMatched,
			Enabled:          true,
		},
	} {
		if err := repos.domains.Create(domain); err != nil {
			t.Fatalf("create domain binding: %v", err)
		}
	}

	overview, err := NewOpsDeploymentPreflightService(
		repos.projects,
		repos.vps,
		repos.connectors,
		repos.domains,
	).EvaluateOverview()
	if err != nil {
		t.Fatalf("EvaluateOverview returned error: %v", err)
	}
	if overview.ProjectCount != 3 {
		t.Fatalf("project count = %d, want 3", overview.ProjectCount)
	}
	if overview.ReadyCount != 0 || overview.ReviewCount != 1 || overview.BlockedCount != 2 {
		t.Fatalf("overview status counts = ready:%d review:%d blocked:%d, want 0:1:2", overview.ReadyCount, overview.ReviewCount, overview.BlockedCount)
	}
	if len(overview.Projects) != 3 {
		t.Fatalf("overview projects len = %d, want 3", len(overview.Projects))
	}
	if overview.GeneratedAt.IsZero() {
		t.Fatal("overview generated_at is zero")
	}

	summaries := make(map[string]ops.DeploymentPreflightSummary, len(overview.Projects))
	for _, summary := range overview.Projects {
		if !summary.GeneratedAt.Equal(overview.GeneratedAt) {
			t.Fatalf("summary %q generated_at = %s, want overview batch time %s", summary.Project, summary.GeneratedAt, overview.GeneratedAt)
		}
		summaries[summary.Project] = summary
	}
	if summaries["review-project"].StatusLevel != ops.DeploymentStatusReview || summaries["review-project"].Ready {
		t.Fatalf("review project summary = %#v", summaries["review-project"])
	}
	if summaries["blocked-project"].StatusLevel != ops.DeploymentStatusBlocked {
		t.Fatalf("blocked project summary = %#v", summaries["blocked-project"])
	}
	if summaries["missing-vps-project"].StatusLevel != ops.DeploymentStatusBlocked {
		t.Fatalf("missing VPS project summary = %#v", summaries["missing-vps-project"])
	}
	if len(summaries["missing-vps-project"].BlockReasons) == 0 {
		t.Fatalf("missing VPS project has no block reasons: %#v", summaries["missing-vps-project"])
	}
}

func TestDeploymentPreflightOverviewFiltersProjectsByEnvironment(t *testing.T) {
	repos := newDeploymentPreflightTestRepositories(t)
	now := time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)

	productionProject := newPreflightFixtureProject("production-project", 1001, now)
	stagingProject := newPreflightFixtureProject("staging-project", 1002, now)
	stagingProject.Environment = ops.ProjectEnvironmentStaging
	stagingProject.ComposeSource = "compose.staging.yml"
	stagingProject.GatewayNetwork = "staging-edge"
	stagingProject.GatewayAlias = "staging-web"
	if err := repos.projects.Create(productionProject); err != nil {
		t.Fatalf("create production project: %v", err)
	}
	if err := repos.projects.Create(stagingProject); err != nil {
		t.Fatalf("create staging project: %v", err)
	}

	service := NewOpsDeploymentPreflightService(repos.projects, repos.vps, repos.connectors, repos.domains)
	overview, err := service.EvaluateOverviewForEnvironment(ops.ProjectEnvironmentStaging)
	if err != nil {
		t.Fatalf("EvaluateOverviewForEnvironment returned error: %v", err)
	}
	if overview.Environment != ops.ProjectEnvironmentStaging {
		t.Fatalf("overview environment = %q, want staging", overview.Environment)
	}
	if overview.ProjectCount != 1 || len(overview.Projects) != 1 {
		t.Fatalf("overview project count = %d/%d, want 1", overview.ProjectCount, len(overview.Projects))
	}
	if overview.Projects[0].Project != stagingProject.Name || overview.Projects[0].Environment != ops.ProjectEnvironmentStaging {
		t.Fatalf("overview project = %#v, want staging project", overview.Projects[0])
	}

	_, err = service.EvaluateOverviewForEnvironment("qa")
	if !errors.Is(err, ErrInvalidOpsProjectEnvironment) {
		t.Fatalf("invalid environment error = %v, want ErrInvalidOpsProjectEnvironment", err)
	}
}

type deploymentPreflightTestRepositories struct {
	projects   *repository.OpsProjectBindingRepository
	vps        *repository.OpsVPSBindingRepository
	connectors *repository.OpsConnectorRepository
	domains    *repository.OpsDomainBindingRepository
}

func newDeploymentPreflightTestRepositories(t *testing.T) deploymentPreflightTestRepositories {
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
		t.Fatalf("migrate deployment preflight tables: %v", err)
	}
	return deploymentPreflightTestRepositories{
		projects:   repository.NewOpsProjectBindingRepository(db),
		vps:        repository.NewOpsVPSBindingRepository(db),
		connectors: repository.NewOpsConnectorRepository(db),
		domains:    repository.NewOpsDomainBindingRepository(db),
	}
}

func newPreflightFixtureProject(name string, vpsID uint, now time.Time) *ops.ProjectBinding {
	return &ops.ProjectBinding{
		Name:                   name,
		VPSBindingID:           vpsID,
		Environment:            ops.ProjectEnvironmentProduction,
		ComposeSource:          "compose.prod.yml",
		ComposeProjectName:     name,
		GatewayNetwork:         "shared-edge",
		GatewayAlias:           "theme-web",
		Services:               "db, redis, migrate, edge-config, api, storefront, admin, web",
		Networks:               "db, cache, app, shared-edge",
		Volumes:                name + "-postgres-data, " + name + "-redis-data, uploads",
		CurrentImageTag:        "sha-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		CurrentCommitSHA:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Status:                 ops.ProjectStatusActive,
		HealthStatus:           ops.ProjectHealthHealthy,
		ObservedContainerCount: 7,
		ObservedRunningCount:   7,
		ObservedHealthyCount:   7,
		LastCheckedAt:          &now,
		LastDeploymentAt:       &now,
		BackupPolicy:           "Daily PostgreSQL and uploads backups with off-VPS copy.",
		RestoreNotes:           "2026-08-13 restore exercise completed.",
		Enabled:                true,
	}
}

func findPreflightCheck(checks []ops.DeploymentPreflightCheck, key string) ops.DeploymentPreflightCheck {
	for _, check := range checks {
		if check.Key == key {
			return check
		}
	}
	return ops.DeploymentPreflightCheck{Key: key}
}

func deploymentTestUintPointer(value uint) *uint {
	return &value
}
