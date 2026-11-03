package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/config"
)

func TestLocalDevelopmentBackupCredentialCreationPreservesExistingDecryptionMaterial(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "backup-decryption-credential")
	first, err := createPrivateLocalDevelopmentBackupCredential(filename)
	require.NoError(t, err)
	require.Len(t, first, 64)
	second, err := createPrivateLocalDevelopmentBackupCredential(filename)
	require.NoError(t, err)
	require.Equal(t, first, second, "restart must never rotate the key needed by downloaded packages")
	require.NoError(t, os.WriteFile(filename, []byte("broken"), 0600))
	_, err = createPrivateLocalDevelopmentBackupCredential(filename)
	require.Error(t, err)
	retained, err := os.ReadFile(filename)
	require.NoError(t, err)
	require.Equal(t, "broken", string(retained), "invalid existing credentials must be retained for diagnosis")
}

func TestLocalDevelopmentBackupPreparationRejectsReleaseOrRemoteDatabaseBeforeAccess(t *testing.T) {
	originalVersion := Version
	Version = "synthetic-local-backup-build"
	defer func() { Version = originalVersion }()
	for _, scenario := range []struct{ mode, host string }{{"release", "127.0.0.1"}, {"debug", "production.invalid"}} {
		cfg := &config.Config{}
		cfg.Server.Mode = scenario.mode
		cfg.Database.Host = scenario.host
		require.ErrorContains(t, prepareLocalDevelopmentBackupFromEffectiveApplicationConfiguration(cfg), "loopback debug database")
	}
}

func TestLocalDevelopmentBackupHistoricalOriginsPreserveExplicitPortsAndRejectUnreviewedHosts(t *testing.T) {
	origins, err := collectLocalDevelopmentBackupExplicitHistoricalMediaOrigins("http://localhost:10241", []string{"http://localhost:9199", "http://localhost:10241"}, "http://LOCALHOST:9200/,http://127.0.0.1:10241")
	require.NoError(t, err)
	require.Equal(t, []string{"http://localhost:10241", "http://localhost:9199", "http://localhost:9200", "http://127.0.0.1:10241"}, origins)
	for _, value := range []string{"https://store.example.invalid", "http://localhost.evil.invalid", "http://192.168.1.1", "http://user:password@localhost:9199", "http://localhost:9199/uploads", "http://localhost:9199?token=secret", "http://localhost:9199#fragment", "//localhost:9199", "file:///uploads", "http://*", ","} {
		_, err := collectLocalDevelopmentBackupExplicitHistoricalMediaOrigins("http://localhost:10241", nil, value)
		require.Error(t, err, value)
		require.NotContains(t, err.Error(), "password")
		require.NotContains(t, err.Error(), "token=")
	}
}
