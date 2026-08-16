package config

import (
	"net/http"
	"testing"
)

func TestCookieSecureAutoDisablesForLocalHTTP(t *testing.T) {
	cookie := CookieConfig{Secure: "auto"}
	server := ServerConfig{Mode: "debug", BaseURL: "http://localhost:9200"}

	if cookie.SecureEnabled(server) {
		t.Fatal("local HTTP debug server should not force Secure cookies")
	}
}

func TestCookieSecureAutoEnablesForRelease(t *testing.T) {
	cookie := CookieConfig{Secure: "auto"}
	server := ServerConfig{Mode: "release", BaseURL: "http://127.0.0.1:9200"}

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

func TestValidateConfigRejectsShortJWTSecretInRelease(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.Mode = "release"
	cfg.JWT.Secret = "short-secret"

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject short JWT secrets in release mode")
	}
}

func TestValidateConfigAllowsShortJWTSecretInDebug(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.Mode = "debug"
	cfg.JWT.Secret = "short-secret"

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig should allow short JWT secrets outside release mode: %v", err)
	}
}

func TestValidateConfigRejectsInvalidPaymentExpirationConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker.PaymentExpirationEnabled = true
	cfg.Worker.PaymentExpirationIntervalSeconds = 900
	cfg.Worker.PaymentPendingTTLSeconds = 0
	cfg.Worker.PaymentExpirationBatchLimit = 100

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid payment expiration config")
	}
}

func TestValidateConfigRejectsInvalidPaymentRiskMonitoringSchedule(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker.PaymentRiskMonitoringEnabled = true
	cfg.Worker.PaymentRiskMonitoringIntervalSeconds = 0

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject an invalid payment risk monitoring schedule")
	}
}

func TestValidateConfigRejectsOrderAbuseWithoutIdentityLimit(t *testing.T) {
	cfg := validTestConfig()
	cfg.OrderAbuse = OrderAbuseConfig{
		Enabled:                  true,
		OrderCreateWindowSeconds: 600,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject enabled order abuse protection without an identity limit")
	}
}

func TestValidateConfigRejectsInvalidQuickBuyRateLimitConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.QuickBuyRateLimit = QuickBuyRateLimitConfig{
		Enabled:                  true,
		IPRequestsPerMinute:      120,
		IPBurst:                  0,
		SessionRequestsPerMinute: 60,
		SessionBurst:             20,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid Quick Buy rate limit config")
	}
}

func TestValidateConfigRejectsShortPreviousOrderNumberSecretInRelease(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.Mode = "release"
	cfg.JWT.Secret = "test-production-secret-at-least-32-chars"
	cfg.OrderNumber.PreviousSecret = "short-secret"

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject a short previous order number secret in release mode")
	}
}

func TestValidateConfigRejectsInvalidPaymentThreeDSConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.PaymentThreeDS = PaymentThreeDSConfig{
		AdaptiveEnabled:     true,
		LowRiskMaxAmount:    100,
		TrustedPaidOrders:   0,
		VisitorRiskLookback: 30,
		StepUpRiskScore:     80,
		ChallengeRiskScore:  60,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid payment 3DS config")
	}
}

func TestValidateConfigRejectsInvalidPaymentProtectionDurationHierarchy(t *testing.T) {
	cfg := validTestConfig()
	cfg.PaymentProtection = PaymentProtectionConfig{
		Enabled:                            true,
		MaxControlDurationHours:            24,
		MaxPausePaymentDurationHours:       48,
		MaxGlobalPausePaymentDurationHours: 2,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject pause duration above the broader control duration")
	}

	cfg.PaymentProtection.MaxPausePaymentDurationHours = 24
	cfg.PaymentProtection.MaxGlobalPausePaymentDurationHours = 25
	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject global pause duration above the broader pause duration")
	}
}

func TestValidateConfigRejectsInvalidOutboxDispatchConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker.OutboxDispatchEnabled = true
	cfg.Worker.OutboxDispatchIntervalSeconds = 10
	cfg.Worker.OutboxDispatchBatchLimit = 0
	cfg.Worker.OutboxDispatchLockTimeoutSeconds = 300

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid outbox dispatch config")
	}
}

func TestValidateConfigRejectsInvalidCustomerServiceRealtimeConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker = WorkerConfig{
		OutboxDispatchEnabled:            true,
		OutboxDispatchIntervalSeconds:    2,
		OutboxDispatchBatchLimit:         100,
		OutboxDispatchLockTimeoutSeconds: 300,
	}
	cfg.CustomerServiceRealtime = CustomerServiceRealtimeConfig{
		Enabled:               true,
		Stream:                "customer_service:{realtime}:v1",
		StreamMaxLen:          10000,
		ReplayLimit:           0,
		ConsumerBlockSeconds:  5,
		DedupRetentionSeconds: 86400,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject invalid customer-service realtime configuration")
	}
}

func TestValidateConfigRejectsCustomerServiceRealtimeStreamWithoutRedisHashTag(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker = WorkerConfig{
		OutboxDispatchEnabled:            true,
		OutboxDispatchIntervalSeconds:    2,
		OutboxDispatchBatchLimit:         100,
		OutboxDispatchLockTimeoutSeconds: 300,
	}
	cfg.CustomerServiceRealtime = CustomerServiceRealtimeConfig{
		Enabled:               true,
		Stream:                "customer_service:realtime:v1",
		StreamMaxLen:          10000,
		ReplayLimit:           200,
		ConsumerBlockSeconds:  5,
		DedupRetentionSeconds: 86400,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should require a Redis Cluster hash tag for customer-service realtime")
	}
}

func TestValidateConfigRejectsCustomerServiceRealtimeWithoutOutboxDispatcher(t *testing.T) {
	cfg := validTestConfig()
	cfg.CustomerServiceRealtime = CustomerServiceRealtimeConfig{
		Enabled:               true,
		Stream:                "customer_service:{realtime}:v1",
		StreamMaxLen:          10000,
		ReplayLimit:           200,
		ConsumerBlockSeconds:  5,
		DedupRetentionSeconds: 86400,
	}

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should require the outbox dispatcher when customer-service realtime is enabled")
	}
}

func TestValidateConfigAllowsCustomerServiceRealtimeWithOutboxDispatcher(t *testing.T) {
	cfg := validTestConfig()
	cfg.Worker = WorkerConfig{
		OutboxDispatchEnabled:            true,
		OutboxDispatchIntervalSeconds:    2,
		OutboxDispatchBatchLimit:         100,
		OutboxDispatchLockTimeoutSeconds: 300,
	}
	cfg.CustomerServiceRealtime = CustomerServiceRealtimeConfig{
		Enabled:               true,
		Stream:                "customer_service:{realtime}:v1",
		StreamMaxLen:          10000,
		ReplayLimit:           200,
		ConsumerBlockSeconds:  5,
		DedupRetentionSeconds: 86400,
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig should allow a complete customer-service realtime configuration: %v", err)
	}
}

func TestValidateConfigRejectsMissingMediaUploadQuota(t *testing.T) {
	cfg := validTestConfig()
	cfg.MediaUpload.AccountStorageQuotaBytes = 0

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject a missing media upload storage quota")
	}
}

func TestValidateConfigRejectsInvalidProductLockTTL(t *testing.T) {
	cfg := validTestConfig()
	cfg.Cache.ProductLockTTL = 0

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject a non-positive product cache lock TTL")
	}
}

func TestValidateConfigRejectsReleaseWithoutTrustedProxies(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.Mode = "release"
	cfg.JWT.Secret = "test-production-secret-at-least-32-chars"

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig should reject release mode without trusted proxies")
	}
}

func TestSplitEnvListTrimsAndDropsEmptyValues(t *testing.T) {
	got := splitEnvList("https://example.com, https://admin.example.com, ,")
	want := []string{"https://example.com", "https://admin.example.com"}

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
	t.Setenv("SERVER_BASE_URL", "https://example.com")
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USERNAME", "commerce_platform_prod")
	t.Setenv("DB_PASSWORD", "test-database-password")
	t.Setenv("DB_NAME", "commerce_platform_prod")
	t.Setenv("DB_AUTO_MIGRATE", "false")
	t.Setenv("REDIS_HOST", "redis")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_PASSWORD", "test-redis-password")
	t.Setenv("JWT_SECRET", "test-production-secret-at-least-32-chars")
	t.Setenv("GOOGLE_CLIENT_ID", "test-google-client")
	t.Setenv("CORS_ORIGINS", "https://example.com,https://admin.example.com")
	t.Setenv("TRUSTED_PROXIES", "10.0.0.0/8, 172.16.0.0/12")
	t.Setenv("QUICK_BUY_RATE_LIMIT_IP_REQUESTS_PER_MINUTE", "88")
	t.Setenv("QUICK_BUY_RATE_LIMIT_SESSION_BURST", "9")

	cfg, err := Load("../../../config/config.production.yaml")
	if err != nil {
		t.Fatalf("Load production config: %v", err)
	}

	if cfg.Database.Host != "db" || cfg.Database.Username != "commerce_platform_prod" {
		t.Fatalf("database environment overrides not applied: %+v", cfg.Database)
	}
	if cfg.Database.AutoMigrate {
		t.Fatal("production config must keep GORM AutoMigrate disabled")
	}
	if len(cfg.CORS.AllowedOrigins) != 2 || cfg.CORS.AllowedOrigins[1] != "https://admin.example.com" {
		t.Fatalf("CORS_ORIGINS override not applied: %v", cfg.CORS.AllowedOrigins)
	}
	if len(cfg.Server.TrustedProxies) != 2 || cfg.Server.TrustedProxies[1] != "172.16.0.0/12" {
		t.Fatalf("TRUSTED_PROXIES override not applied: %v", cfg.Server.TrustedProxies)
	}
	if cfg.QuickBuyRateLimit.IPRequestsPerMinute != 88 || cfg.QuickBuyRateLimit.SessionBurst != 9 {
		t.Fatalf("Quick Buy rate limit overrides not applied: %+v", cfg.QuickBuyRateLimit)
	}
	if !cfg.CustomerServiceRealtime.Enabled || !cfg.Worker.OutboxDispatchEnabled {
		t.Fatalf("production config must enable the durable customer-service relay and dispatcher: realtime=%+v worker=%+v", cfg.CustomerServiceRealtime, cfg.Worker)
	}
	if cfg.Worker.OutboxDispatchIntervalSeconds != 2 {
		t.Fatalf("production outbox dispatcher interval = %d, want 2", cfg.Worker.OutboxDispatchIntervalSeconds)
	}
}

func validTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Mode:              "debug",
			BaseURL:           "http://localhost:9200",
			ReadTimeout:       60,
			ReadHeaderTimeout: 10,
			WriteTimeout:      60,
			IdleTimeout:       120,
			MaxHeaderBytes:    1 << 20,
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Database: "commerce_platform",
		},
		JWT: JWTConfig{Secret: "test-secret"},
		Cache: CacheConfig{
			ProductLockTTL: 5,
		},
		MediaUpload: MediaUploadConfig{
			AccountStorageQuotaBytes: 20 << 30,
		},
	}
}
