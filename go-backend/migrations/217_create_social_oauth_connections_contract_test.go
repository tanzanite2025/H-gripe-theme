package migrations_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSocialOAuthMigrationStoresOnlyEncryptedTokens(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "217_create_social_oauth_connections.up.sql"))
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	sql := string(data)
	for _, fragment := range []string{
		"CREATE TABLE IF NOT EXISTS social_oauth_connections",
		"access_token_encrypted TEXT NOT NULL",
		"refresh_token_encrypted TEXT NOT NULL",
		"provider_resources JSONB NOT NULL",
		"CREATE TABLE IF NOT EXISTS social_oauth_sessions",
		"state_hash VARCHAR(128) NOT NULL",
		"code_verifier_encrypted TEXT NOT NULL",
		"UNIQUE (provider)",
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("social OAuth migration is missing contract fragment %q", fragment)
		}
	}
}
