package service

import (
	"errors"
	"net"
	"strings"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/pkg/secretbox"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsConnectorNormalizeInputEncryptsCredentials(t *testing.T) {
	t.Setenv(OpsConnectorMasterKeyEnv, "test-ops-connector-master-key")

	service := &OpsConnectorService{}
	record, err := service.normalizeInput(OpsConnectorInput{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthAPIToken,
		Credentials: map[string]string{"token": "secret-token-value"},
	}, nil)
	if err != nil {
		t.Fatalf("normalizeInput() error = %v", err)
	}
	if record.CredentialsEncrypted == "" {
		t.Fatal("normalizeInput() did not encrypt credentials")
	}
	if strings.Contains(record.CredentialsEncrypted, "secret-token-value") {
		t.Fatal("encrypted credentials expose plaintext")
	}
	if got := decodeCredentialFields(record.CredentialFields); len(got) != 1 || got[0] != "token" {
		t.Fatalf("credential fields = %#v, want [token]", got)
	}

	plaintext, err := secretbox.DecryptString(record.CredentialsEncrypted, "test-ops-connector-master-key")
	if err != nil {
		t.Fatalf("DecryptString() error = %v", err)
	}
	if !strings.Contains(plaintext, "secret-token-value") {
		t.Fatalf("decrypted credentials = %q, want token value", plaintext)
	}
}

func TestOpsConnectorViewDoesNotExposeCredentials(t *testing.T) {
	view := connectorView(ops.Connector{
		ID:                   7,
		Name:                 "Cloudflare Production",
		Provider:             ops.ConnectorProviderCloudflare,
		AuthType:             ops.ConnectorAuthAPIToken,
		CredentialsEncrypted: "v1:encrypted-value",
		CredentialFields:     `["token"]`,
	})

	if !view.CredentialConfigured {
		t.Fatal("CredentialConfigured = false, want true")
	}
	if strings.Contains(view.Endpoint, "encrypted-value") {
		t.Fatal("connector view contains encrypted credential payload")
	}
	if len(view.CredentialFields) != 1 || view.CredentialFields[0] != "token" {
		t.Fatalf("CredentialFields = %#v, want [token]", view.CredentialFields)
	}
}

func TestValidateConnectorEndpointRejectsUnsafeTargets(t *testing.T) {
	tests := []string{
		"http://127.0.0.1:8080/health",
		"https://10.0.0.5/health",
		"https://localhost/health",
		"http://example.com/health",
	}
	for _, endpoint := range tests {
		t.Run(endpoint, func(t *testing.T) {
			if err := validateConnectorEndpoint(endpoint, ops.ConnectorEnvironmentProduction); err == nil {
				t.Fatalf("validateConnectorEndpoint(%q) expected error", endpoint)
			}
		})
	}

	if err := validateConnectorEndpoint("http://localhost:8080/health", ops.ConnectorEnvironmentLocal); err == nil {
		t.Fatal("local endpoint validation should still reject localhost")
	}
	if err := validateConnectorEndpoint("https://api.example.com/health", ops.ConnectorEnvironmentProduction); err != nil {
		t.Fatalf("validateConnectorEndpoint() error = %v", err)
	}
}

func TestIsUnsafeConnectorIP(t *testing.T) {
	unsafe := []string{"127.0.0.1", "10.0.0.1", "172.16.0.1", "192.168.1.1", "::1", "fe80::1"}
	for _, raw := range unsafe {
		if !isUnsafeConnectorIP(net.ParseIP(raw)) {
			t.Fatalf("isUnsafeConnectorIP(%q) = false, want true", raw)
		}
	}
	if isUnsafeConnectorIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("isUnsafeConnectorIP(public IP) = true, want false")
	}
}

func TestConnectorCredentialConfiguredChecksEnvironmentVariable(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "")
	record := ops.Connector{
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
	}
	if connectorCredentialConfigured(record) {
		t.Fatal("empty environment variable should not be reported as configured")
	}

	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	if !connectorCredentialConfigured(record) {
		t.Fatal("non-empty environment variable should be reported as configured")
	}
}

func TestOpsConnectorCreateNormalizesAndPersistsConnector(t *testing.T) {
	t.Setenv(OpsConnectorMasterKeyEnv, "test-ops-connector-master-key")
	service, repo := newOpsConnectorTestService(t)

	view, err := service.Create(OpsConnectorInput{
		Name:        "  Cloudflare Production  ",
		Provider:    " CLOUDFLARE ",
		Environment: "",
		AuthType:    "",
		Endpoint:    "https://api.example.com/health",
		Credentials: map[string]string{" token ": "secret-token"},
		Scopes:      "zones:read",
		Enabled:     boolPointer(true),
		Notes:       "  read only  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if view.ID == 0 {
		t.Fatal("Create() returned zero connector ID")
	}
	if view.Name != "Cloudflare Production" ||
		view.Provider != ops.ConnectorProviderCloudflare ||
		view.Environment != ops.ConnectorEnvironmentProduction ||
		view.AuthType != ops.ConnectorAuthAPIToken ||
		view.Status != ops.ConnectorStatusPending ||
		!view.Enabled {
		t.Fatalf("Create() view = %#v, want normalized active defaults", view)
	}
	if !view.CredentialConfigured || len(view.CredentialFields) != 1 || view.CredentialFields[0] != "token" {
		t.Fatalf("Create() credential projection = %#v, want configured token", view)
	}

	record, err := repo.FindByID(view.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if record.CredentialsEncrypted == "" || strings.Contains(record.CredentialsEncrypted, "secret-token") {
		t.Fatalf("stored credentials = %q, want encrypted non-plaintext value", record.CredentialsEncrypted)
	}
}

func TestOpsConnectorCreateRejectsDuplicateName(t *testing.T) {
	service, _ := newOpsConnectorTestService(t)
	input := OpsConnectorInput{
		Name:     "Duplicate Connector",
		Provider: ops.ConnectorProviderOther,
		Endpoint: "https://api.example.com/health",
	}
	if _, err := service.Create(input); err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	_, err := service.Create(input)
	if err == nil {
		t.Fatal("second Create() error = nil, want duplicate-name error")
	}
	if !errors.Is(err, ErrInvalidOpsConnector) {
		t.Fatalf("second Create() error = %v, want ErrInvalidOpsConnector", err)
	}
}

func TestOpsConnectorUpdatePreservesCredentialsAndSetEnabled(t *testing.T) {
	t.Setenv(OpsConnectorMasterKeyEnv, "test-ops-connector-master-key")
	service, repo := newOpsConnectorTestService(t)

	created, err := service.Create(OpsConnectorInput{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		AuthType:    ops.ConnectorAuthAPIToken,
		Credentials: map[string]string{"token": "original-token"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	before, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID() before update error = %v", err)
	}

	updated, err := service.Update(created.ID, OpsConnectorInput{
		Name:        "Hostinger Staging",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentStaging,
		AuthType:    ops.ConnectorAuthAPIToken,
		Endpoint:    "https://api.example.com/staging",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	after, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}
	if after.CredentialsEncrypted != before.CredentialsEncrypted {
		t.Fatal("Update() replaced encrypted credentials when no new credentials were supplied")
	}
	if updated.Name != "Hostinger Staging" ||
		updated.Environment != ops.ConnectorEnvironmentStaging ||
		!updated.CredentialConfigured {
		t.Fatalf("Update() view = %#v, want changed metadata with preserved credentials", updated)
	}

	disabled, err := service.SetEnabled(created.ID, false)
	if err != nil {
		t.Fatalf("SetEnabled(false) error = %v", err)
	}
	if disabled.Enabled || disabled.Status != ops.ConnectorStatusDisabled {
		t.Fatalf("SetEnabled(false) view = %#v, want disabled status", disabled)
	}

	enabled, err := service.SetEnabled(created.ID, true)
	if err != nil {
		t.Fatalf("SetEnabled(true) error = %v", err)
	}
	if !enabled.Enabled || enabled.Status != ops.ConnectorStatusPending {
		t.Fatalf("SetEnabled(true) view = %#v, want pending status", enabled)
	}
}

func TestOpsConnectorTestPersistsDisabledManualAndMissingCredentialResults(t *testing.T) {
	t.Setenv(OpsConnectorMasterKeyEnv, "test-ops-connector-master-key")
	service, repo := newOpsConnectorTestService(t)

	disabled, err := service.Create(OpsConnectorInput{
		Name:     "Disabled Connector",
		Provider: ops.ConnectorProviderOther,
	})
	if err != nil {
		t.Fatalf("Create(disabled) error = %v", err)
	}
	if _, err := service.SetEnabled(disabled.ID, false); err != nil {
		t.Fatalf("SetEnabled(disabled) error = %v", err)
	}
	disabledResult, err := service.Test(t.Context(), disabled.ID)
	if err != nil {
		t.Fatalf("Test(disabled) error = %v", err)
	}
	if disabledResult.Success || disabledResult.Message != "connector is disabled" {
		t.Fatalf("Test(disabled) result = %#v, want disabled failure", disabledResult)
	}
	disabledRecord, err := repo.FindByID(disabled.ID)
	if err != nil {
		t.Fatalf("FindByID(disabled) error = %v", err)
	}
	if disabledRecord.LastTestStatus != ops.ConnectorTestFailed ||
		disabledRecord.Status != ops.ConnectorStatusDisabled ||
		disabledRecord.LastError != "connector is disabled" {
		t.Fatalf("disabled persisted state = %#v, want failed/disabled state", disabledRecord)
	}

	manual, err := service.Create(OpsConnectorInput{
		Name:     "Manual Connector",
		Provider: ops.ConnectorProviderOther,
		AuthType: ops.ConnectorAuthManual,
	})
	if err != nil {
		t.Fatalf("Create(manual) error = %v", err)
	}
	manualResult, err := service.Test(t.Context(), manual.ID)
	if err != nil {
		t.Fatalf("Test(manual) error = %v", err)
	}
	if manualResult.Success || manualResult.Message != "manual connector does not support an automated test" {
		t.Fatalf("Test(manual) result = %#v, want manual failure", manualResult)
	}

	missingCredentials, err := service.Create(OpsConnectorInput{
		Name:     "Missing Credentials Connector",
		Provider: ops.ConnectorProviderCloudflare,
		AuthType: ops.ConnectorAuthAPIToken,
		Endpoint: "https://api.example.com/health",
	})
	if err != nil {
		t.Fatalf("Create(missing credentials) error = %v", err)
	}
	missingResult, err := service.Test(t.Context(), missingCredentials.ID)
	if err != nil {
		t.Fatalf("Test(missing credentials) error = %v", err)
	}
	if missingResult.Success || missingResult.Message != "connector credentials are not configured" {
		t.Fatalf("Test(missing credentials) result = %#v, want credential failure", missingResult)
	}
}

func newOpsConnectorTestService(t *testing.T) (*OpsConnectorService, *repository.OpsConnectorRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ops.Connector{}); err != nil {
		t.Fatalf("migrate connector table: %v", err)
	}
	repo := repository.NewOpsConnectorRepository(db)
	return NewOpsConnectorService(repo), repo
}

func boolPointer(value bool) *bool {
	return &value
}
