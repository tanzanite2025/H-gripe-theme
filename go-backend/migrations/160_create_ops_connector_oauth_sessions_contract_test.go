package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func readOpsOAuthMigrationFile(t *testing.T, name string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), name))
	if err != nil {
		t.Fatalf("read migration %s: %v", name, err)
	}
	return string(data)
}

func TestOpsConnectorOAuthSessionMigrationContract(t *testing.T) {
	sql := readOpsOAuthMigrationFile(t, "160_create_ops_connector_oauth_sessions.up.sql")
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS ops_connector_oauth_sessions",
		"state_hash VARCHAR(128) NOT NULL",
		"code_verifier_encrypted TEXT NOT NULL DEFAULT ''",
		"expires_at TIMESTAMPTZ NOT NULL",
		"consumed_at TIMESTAMPTZ NULL",
		"FOREIGN KEY (connector_id) REFERENCES ops_connectors (id)",
		"idx_ops_connector_oauth_session_state_hash",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("OAuth session migration is missing contract fragment %q", fragment)
		}
	}
}
