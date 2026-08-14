package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadAdminPasswordRejectsBothSources(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "N3w-Admin-Secret!")
	t.Setenv("ADMIN_PASSWORD_FILE", "secret.txt")

	_, _, err := readAdminPassword()

	require.EqualError(t, err, "set only one of ADMIN_PASSWORD or ADMIN_PASSWORD_FILE")
}

func TestReadAdminPasswordReadsFileWithoutTrailingNewline(t *testing.T) {
	secretFile := writeAdminctlTestSecret(t, "N3w-Admin-Secret!\n")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("ADMIN_PASSWORD_FILE", secretFile)

	password, source, err := readAdminPassword()

	require.NoError(t, err)
	require.Equal(t, "N3w-Admin-Secret!", password)
	require.Equal(t, "ADMIN_PASSWORD_FILE", source)
}

func TestRequireProductionConfirmation(t *testing.T) {
	t.Setenv("ADMINCTL_CONFIRM", "")

	err := requireProductionConfirmation("production")
	require.Error(t, err)

	t.Setenv("ADMINCTL_CONFIRM", productionConfirmationValue)
	require.NoError(t, requireProductionConfirmation("production"))
	require.NoError(t, requireProductionConfirmation("debug"))
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	err := run([]string{"missing"}, os.Stdout, os.Stderr)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown command")
}

func writeAdminctlTestSecret(t *testing.T, value string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "admin-password-*")
	require.NoError(t, err)
	_, err = file.WriteString(value)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return file.Name()
}
