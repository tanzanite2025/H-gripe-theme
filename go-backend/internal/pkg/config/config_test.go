package config

import (
	"net/http"
	"testing"
)

func TestCookieSecureAutoDisablesForLocalHTTP(t *testing.T) {
	cookie := CookieConfig{Secure: "auto"}
	server := ServerConfig{Mode: "debug", BaseURL: "http://localhost:9000"}

	if cookie.SecureEnabled(server) {
		t.Fatal("local HTTP debug server should not force Secure cookies")
	}
}

func TestCookieSecureAutoEnablesForRelease(t *testing.T) {
	cookie := CookieConfig{Secure: "auto"}
	server := ServerConfig{Mode: "release", BaseURL: "http://127.0.0.1:9000"}

	if !cookie.SecureEnabled(server) {
		t.Fatal("release server must force Secure cookies")
	}
}

func TestCookieSecureAutoEnablesForHTTPSBaseURL(t *testing.T) {
	cookie := CookieConfig{Secure: "auto"}
	server := ServerConfig{Mode: "debug", BaseURL: "https://api.example.com"}

	if !cookie.SecureEnabled(server) {
		t.Fatal("HTTPS base URL must force Secure cookies")
	}
}

func TestCookieSameSiteDefaultsToLax(t *testing.T) {
	cookie := CookieConfig{}

	if got := cookie.SameSiteMode(); got != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want %v", got, http.SameSiteLaxMode)
	}
}

func TestValidateConfigRejectsSameSiteNoneWithoutSecure(t *testing.T) {
	cfg := validTestConfig()
	cfg.Cookie = CookieConfig{Secure: "never", SameSite: "none"}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject SameSite=None without Secure cookies")
	}
}

func TestValidateConfigRejectsInvalidCookieSecureMode(t *testing.T) {
	cfg := validTestConfig()
	cfg.Cookie = CookieConfig{Secure: "maybe", SameSite: "lax"}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid cookie.secure values")
	}
}

func TestSplitEnvListTrimsAndDropsEmptyValues(t *testing.T) {
	got := splitEnvList("https://tanzanite.site, https://admin.tanzanite.site, ,")
	want := []string{"https://tanzanite.site", "https://admin.tanzanite.site"}

	if len(got) != len(want) {
		t.Fatalf("splitEnvList length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("splitEnvList[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func TestLoadProductionConfigUsesEnvironmentOverrides(t *testing.T) {
	t.Setenv("SERVER_BASE_URL", "https://tanzanite.site")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USERNAME", "tanzanite_prod")
	t.Setenv("DB_PASSWORD", "test-database-password")
	t.Setenv("DB_NAME", "tanzanite_prod")
	t.Setenv("DB_AUTO_MIGRATE", "false")
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_PASSWORD", "test-redis-password")
	t.Setenv("JWT_SECRET", "test-production-secret")
	t.Setenv("GOOGLE_CLIENT_ID", "test-google-client")
	t.Setenv("CORS_ORIGINS", "https://tanzanite.site,https://admin.tanzanite.site")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.0/12")

	cfg, err := Load("../../../config/config.production.yaml")
	if err != nil {
		t.Fatalf("Load production config: %v", err)
	}

	if cfg.Database.Host != "db" || cfg.Database.Username != "tanzanite_prod" {
		t.Fatalf("database environment overrides not applied: %+v", cfg.Database)
	}
	if cfg.Database.AutoMigrate {
		t.Fatal("production config must keep GORM AutoMigrate disabled")
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[1] != "https://admin.tanzanite.site" {
		t.Fatalf("CORS_ORIGINS override not applied: %v", cfg.CORS.AllowedOrigins)
	}
	if len(cfg.Server.TrustedProxies) != 2 || cfg.Server.TrustedProxies[1] != "172.16.0.0/12" {
		t.Fatalf("TRUSTED_PROXIES override not applied: %v", cfg.Server.TrustedProxies)
	}
}

func validTestConfig() *Config {
	return &Config{
		Server: ServerConfig{Mode: "debug", BaseURL: "http://localhost:9000"},
		Database: DatabaseConfig{
			Host:     "localhost",
			Database: "tanzanite",
		},
		JWT: JWTConfig{Secret: "test-secret"},
	}
}
