package service

import (
	"net/url"
	"os"
	"strings"
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

func TestNormalizeOAuthProviderSupportsGitHub(t *testing.T) {
	if got := normalizeOAuthProvider(" GitHub "); got != ops.ConnectorProviderGitHub {
		t.Fatalf("normalizeOAuthProvider(GitHub) = %q, want %q", got, ops.ConnectorProviderGitHub)
	}
	if got := normalizeOAuthProvider("ghcr"); got != "" {
		t.Fatalf("normalizeOAuthProvider(ghcr) = %q, want empty provider", got)
	}
}

func TestGitHubOAuthScopesUseConfiguredValue(t *testing.T) {
	previous := os.Getenv("OPS_GITHUB_OAUTH_SCOPES")
	t.Cleanup(func() {
		_ = os.Setenv("OPS_GITHUB_OAUTH_SCOPES", previous)
	})
	_ = os.Setenv("OPS_GITHUB_OAUTH_SCOPES", "read:user, user:email\nrepo")

	if got := githubOAuthScopes(); got != "read:user user:email repo" {
		t.Fatalf("githubOAuthScopes() = %q, want normalized configured scopes", got)
	}
	if got := githubOAuthScopesFromConnector(""); got != "read:user user:email repo" {
		t.Fatalf("githubOAuthScopesFromConnector(empty) = %q, want configured scopes", got)
	}
}

func TestBuildGitHubAuthorizationURL(t *testing.T) {
	oauthService := &OpsConnectorOAuthService{}
	authorizationURL, err := oauthService.buildAuthorizationURL(
		ops.ConnectorProviderGitHub,
		"client-id",
		"https://ops.example.com/api/admin/ops/connectors/oauth/callback",
		"oauth-state",
		"pkce-challenge",
		"read:user,repo",
	)
	if err != nil {
		t.Fatalf("buildGitHubAuthorizationURL returned error: %v", err)
	}
	parsed, err := url.Parse(authorizationURL)
	if err != nil {
		t.Fatalf("parse GitHub authorization URL: %v", err)
	}
	if parsed.Scheme != "https" || parsed.Host != "github.com" || parsed.Path != "/login/oauth/authorize" {
		t.Fatalf("authorization URL = %q, want GitHub authorize endpoint", authorizationURL)
	}
	query := parsed.Query()
	if query.Get("client_id") != "client-id" ||
		query.Get("redirect_uri") != "https://ops.example.com/api/admin/ops/connectors/oauth/callback" ||
		query.Get("state") != "oauth-state" ||
		query.Get("code_challenge") != "pkce-challenge" ||
		query.Get("code_challenge_method") != "S256" ||
		query.Get("scope") != "read:user repo" {
		t.Fatalf("authorization query = %v, want GitHub OAuth parameters", query)
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

	github, err := oauthService.resolveConnector(
		ops.ConnectorProviderGitHub,
		nil,
		ops.ConnectorEnvironmentProduction,
	)
	if err != nil {
		t.Fatalf("resolveConnector GitHub returned error: %v", err)
	}
	if github.Name != "GitHub Production" || github.Endpoint != "https://api.github.com/user" {
		t.Fatalf("GitHub connector = %#v, want default GitHub connector", github)
	}
	if !strings.Contains(github.Scopes, "read:packages") {
		t.Fatalf("GitHub connector scopes = %q, want read:packages", github.Scopes)
	}

	ghcr, err := oauthService.resolveConnector(
		ops.ConnectorProviderGHCR,
		nil,
		ops.ConnectorEnvironmentProduction,
	)
	if err != nil {
		t.Fatalf("resolveConnector GHCR returned error: %v", err)
	}
	if ghcr.Name != "GHCR Production" || ghcr.Endpoint != "https://api.github.com/user" {
		t.Fatalf("GHCR connector = %#v, want default GHCR connector", ghcr)
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
