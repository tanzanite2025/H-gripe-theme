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
