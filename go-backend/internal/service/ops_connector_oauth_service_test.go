package service

import (
	"testing"

	"commerce-platform/internal/domain/ops"
)

func TestFinalizeOAuthBindingResultMarksPartialSuccess(t *testing.T) {
	result := &OpsConnectorOAuthCallbackResult{
		BoundVPSCount:     1,
		BoundProjectCount: 2,
		BoundDomainCount:  3,
		BindingWarnings:   []string{"project sync failed"},
	}

	finalizeOAuthBindingResult("Hostinger Production", result)

	if result.Status != "connected_with_warnings" {
		t.Fatalf("status = %q, want connected_with_warnings", result.Status)
	}
	if result.Message != "Hostinger Production connected; bound 1 VPS, 2 projects, and 3 domains; 1 binding or sync warnings require review: project sync failed" {
		t.Fatalf("message = %q, want partial-success summary", result.Message)
	}
}

func TestFinalizeOAuthBindingResultMarksCompleteSuccess(t *testing.T) {
	result := &OpsConnectorOAuthCallbackResult{}

	finalizeOAuthBindingResult("Cloudflare Production", result)

	if result.Status != "connected" {
		t.Fatalf("status = %q, want connected", result.Status)
	}
	if result.Message != "Cloudflare Production connected; bound 0 VPS, 0 projects, and 0 domains" {
		t.Fatalf("message = %q, want complete-success summary", result.Message)
	}
}

func TestNormalizeOAuthConnectorEnvironment(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "", want: ops.ConnectorEnvironmentProduction},
		{input: " staging ", want: ops.ConnectorEnvironmentStaging},
		{input: "TEST", want: ops.ConnectorEnvironmentTest},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got, err := normalizeOAuthConnectorEnvironment(test.input)
			if err != nil {
				t.Fatalf("normalizeOAuthConnectorEnvironment(%q) error = %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("normalizeOAuthConnectorEnvironment(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
	if _, err := normalizeOAuthConnectorEnvironment("qa"); err == nil {
		t.Fatal("normalizeOAuthConnectorEnvironment(qa) error = nil, want validation error")
	}
}

func TestOAuthResolveConnectorUsesRequestedEnvironment(t *testing.T) {
	connectorService, connectorRepo := newOpsConnectorTestService(t)
	oauthService := &OpsConnectorOAuthService{
		connectorRepo:    connectorRepo,
		connectorService: connectorService,
	}

	connector, err := oauthService.resolveConnector(
		ops.ConnectorProviderCloudflare,
		nil,
		ops.ConnectorEnvironmentStaging,
	)
	if err != nil {
		t.Fatalf("resolveConnector returned error: %v", err)
	}
	if connector.Environment != ops.ConnectorEnvironmentStaging {
		t.Fatalf("connector environment = %q, want staging", connector.Environment)
	}
	if connector.Name != "Cloudflare Staging" {
		t.Fatalf("connector name = %q, want Cloudflare Staging", connector.Name)
	}

	production := &ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthBearer,
		Status:      ops.ConnectorStatusPending,
		Enabled:     true,
	}
	if err := connectorRepo.Create(production); err != nil {
		t.Fatalf("create production connector: %v", err)
	}
	reused, err := oauthService.resolveConnector(
		ops.ConnectorProviderCloudflare,
		nil,
		ops.ConnectorEnvironmentStaging,
	)
	if err != nil {
		t.Fatalf("resolveConnector reuse returned error: %v", err)
	}
	if reused.ID != connector.ID {
		t.Fatalf("resolveConnector reused connector id = %d, want staging connector %d", reused.ID, connector.ID)
	}

	if _, err := oauthService.resolveConnector(
		ops.ConnectorProviderCloudflare,
		&production.ID,
		ops.ConnectorEnvironmentStaging,
	); err == nil {
		t.Fatal("resolveConnector accepted a connector from another environment")
	}
}

func TestFirstEnvironmentVPS(t *testing.T) {
	records := []ops.VPSBinding{
		{ID: 1, Environment: ops.VPSEnvironmentProduction, Enabled: true},
		{ID: 2, Environment: ops.VPSEnvironmentStaging, Enabled: false},
		{ID: 3, Environment: ops.VPSEnvironmentStaging, Enabled: true},
	}
	selected := firstEnvironmentVPS(records, ops.VPSEnvironmentStaging)
	if selected == nil || selected.ID != 3 {
		t.Fatalf("firstEnvironmentVPS(staging) = %#v, want VPS 3", selected)
	}
	if selected := firstEnvironmentVPS(records, ops.VPSEnvironmentTest); selected != nil {
		t.Fatalf("firstEnvironmentVPS(test) = %#v, want nil", selected)
	}
}
