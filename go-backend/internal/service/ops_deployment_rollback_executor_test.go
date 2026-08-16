package service

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"

	"github.com/stretchr/testify/require"
)

func TestOpsDeploymentRollbackCommandIsFixedAndRefBounded(t *testing.T) {
	const ref = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	command := opsDeploymentRollbackCommand(ref, "/srv/commerce-platform")

	require.Equal(t, "cd -- /srv/commerce-platform && DEPLOY_REF="+ref+" ./deploy.sh", command)
	require.NotContains(t, command, ";")
	require.NotContains(t, command, "&& DEPLOY_REF="+ref+" ./deploy.sh &&")
}

func TestValidateOpsDeploymentRollbackRejectsInjectionAndWrongTarget(t *testing.T) {
	config := OpsDeploymentRollbackSSHConfig{
		Enabled:        true,
		Host:           "prod.example.com",
		Port:           22,
		User:           "deploy",
		PrivateKeyPath: "/run/secrets/deploy.key",
		KnownHostsPath: "/run/secrets/known_hosts",
		Workdir:        "/srv/commerce-platform",
		Timeout:        15 * time.Minute,
		MaxOutputBytes: 1024,
	}
	vps := &ops.VPSBinding{
		ID:               3,
		Hostname:         "prod.example.com",
		ObservedHostname: "prod.example.com",
		IPv4:             "203.0.113.10",
	}

	err := validateOpsDeploymentRollbackInput(config, OpsDeploymentRollbackExecutionInput{
		Environment: "production",
		RollbackRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa;id",
		VPS:         vps,
	})
	require.ErrorIs(t, err, ErrOpsDeploymentRollbackInvalidRef)

	config.Host = "other.example.com"
	err = validateOpsDeploymentRollbackInput(config, OpsDeploymentRollbackExecutionInput{
		Environment: "production",
		RollbackRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		VPS:         vps,
	})
	require.ErrorIs(t, err, ErrOpsDeploymentRollbackInvalidTarget)

	config.Host = "prod.example.com"
	config.Workdir = "/srv/commerce-platform;touch"
	err = validateOpsDeploymentRollbackInput(config, OpsDeploymentRollbackExecutionInput{
		Environment: "production",
		RollbackRef: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		VPS:         vps,
	})
	require.ErrorIs(t, err, ErrOpsDeploymentRollbackInvalidConfig)
}

func TestOpsDeploymentRollbackOutputIsBounded(t *testing.T) {
	output := newOpsDeploymentBoundedOutput(8)
	_, err := output.Write([]byte("1234567890"))
	require.NoError(t, err)
	require.Equal(t, "12345678\n[stream output truncated]", strings.TrimSpace(output.String()))
}

func TestOpsDeploymentRollbackExecutorDisabledByDefault(t *testing.T) {
	t.Setenv(OpsDeploymentRollbackEnabledEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHHostEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHPortEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHUserEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHPrivateKeyEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHKnownHostsEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHWorkdirEnv, "")
	t.Setenv(OpsDeploymentRollbackSSHTimeoutEnv, "")

	executor := NewOpsDeploymentSSHRollbackExecutorFromEnv()
	_, err := executor.ExecuteRollback(nil, OpsDeploymentRollbackExecutionInput{})

	require.ErrorIs(t, err, ErrOpsDeploymentRollbackDisabled)
}
