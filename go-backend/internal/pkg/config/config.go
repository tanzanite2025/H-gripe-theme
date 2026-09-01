package config

import (
	"commerce-platform/internal/pkg/locales"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server                       ServerConfig                       `mapstructure:"server"`
	Database                     DatabaseConfig                     `mapstructure:"database"`
	Redis                        RedisConfig                        `mapstructure:"redis"`
	JWT                          JWTConfig                          `mapstructure:"jwt"`
	OAuth                        OAuthConfig                        `mapstructure:"oauth"`
	GoogleMerchant               GoogleMerchantConfig               `mapstructure:"google_merchant"`
	GoogleIndexing               GoogleIndexingConfig               `mapstructure:"google_indexing"`
	SocialOAuth                  SocialOAuthConfig                  `mapstructure:"social_oauth"`
	I18n                         I18nConfig                         `mapstructure:"i18n"`
	CORS                         CORSConfig                         `mapstructure:"cors"`
	Cookie                       CookieConfig                       `mapstructure:"cookie"`
	Cache                        CacheConfig                        `mapstructure:"cache"`
	Log                          LogConfig                          `mapstructure:"log"`
	Worker                       WorkerConfig                       `mapstructure:"worker"`
	SiteQuality                  SiteQualityConfig                  `mapstructure:"site_quality"`
	CustomerServiceRealtime      CustomerServiceRealtimeConfig      `mapstructure:"customer_service_realtime"`
	BehaviorEvents               BehaviorEventsConfig               `mapstructure:"behavior_events"`
	AntiAbuse                    AntiAbuseConfig                    `mapstructure:"anti_abuse"`
	OrderAbuse                   OrderAbuseConfig                   `mapstructure:"order_abuse"`
	OrderNumber                  OrderNumberConfig                  `mapstructure:"order_number"`
	PaymentRisk                  PaymentRiskConfig                  `mapstructure:"payment_risk"`
	PaymentBINRateLimit          PaymentBINRateLimitConfig          `mapstructure:"payment_bin_rate_limit"`
	PaymentGatewayCircuitBreaker PaymentGatewayCircuitBreakerConfig `mapstructure:"payment_gateway_circuit_breaker"`
	OutboundHTTPResilience       OutboundHTTPResilienceConfig       `mapstructure:"outbound_http_resilience"`
	PaymentRiskMonitoring        PaymentRiskMonitoringConfig        `mapstructure:"payment_risk_monitoring"`
	PaymentProtection            PaymentProtectionConfig            `mapstructure:"payment_protection"`
	PaymentThreeDS               PaymentThreeDSConfig               `mapstructure:"payment_3ds"`
	VisitorRisk                  VisitorRiskConfig                  `mapstructure:"visitor_risk"`
	RequestSigning               RequestSigningConfig               `mapstructure:"request_signing"`
	QuickBuyRateLimit            QuickBuyRateLimitConfig            `mapstructure:"quick_buy_rate_limit"`
	FeedbackRateLimit            FeedbackRateLimitConfig            `mapstructure:"feedback_rate_limit"`
	MediaUpload                  MediaUploadConfig                  `mapstructure:"media_upload"`
	ShowcaseUploadProtection     ShowcaseUploadProtectionConfig     `mapstructure:"showcase_upload_protection"`
}

// MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds is intentionally longer
// than the fixed provider request timeout because PayPal can perform token
// acquisition during gateway initialization before it creates the payment.
const MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds = 60

// MinimumOutboundHTTPHalfOpenProbeTimeoutSeconds leaves enough time for the
// longest configured non-payment HTTP client plus a small Redis/network margin.
const MinimumOutboundHTTPHalfOpenProbeTimeoutSeconds = 30

const DefaultVisitorProfileIPAddressRetentionDays = 30

const (
	developmentSiteQualityRunnerToken = "dev-site-quality-runner-token-0123456789"
	placeholderSiteQualityRunnerToken = "change_me_site_quality_runner_token"
)

type ServerConfig struct {
	Port              string   `mapstructure:"port"`
	Mode              string   `mapstructure:"mode"`
	BaseURL           string   `mapstructure:"base_url"`
	ReadTimeout       int      `mapstructure:"read_timeout"`
	ReadHeaderTimeout int      `mapstructure:"read_header_timeout"`
	WriteTimeout      int      `mapstructure:"write_timeout"`
	IdleTimeout       int      `mapstructure:"idle_timeout"`
	MaxHeaderBytes    int      `mapstructure:"max_header_bytes"`
	TrustedProxies    []string `mapstructure:"trusted_proxies"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	AutoMigrate     bool   `mapstructure:"auto_migrate"`
	LogLevel        string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Mode             string   `mapstructure:"mode"`
	Host             string   `mapstructure:"host"`
	Port             int      `mapstructure:"port"`
	Addrs            []string `mapstructure:"addrs"`
	Username         string   `mapstructure:"username"`
	Password         string   `mapstructure:"password"`
	DB               int      `mapstructure:"db"`
	PoolSize         int      `mapstructure:"pool_size"`
	MasterName       string   `mapstructure:"master_name"`
	SentinelUsername string   `mapstructure:"sentinel_username"`
	SentinelPassword string   `mapstructure:"sentinel_password"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

type OAuthConfig struct {
	GoogleClientID string `mapstructure:"google_client_id"`
}

type GoogleMerchantConfig struct {
	ClientID           string `mapstructure:"client_id"`
	ClientSecret       string `mapstructure:"client_secret"`
	RedirectURL        string `mapstructure:"redirect_url"`
	PostConnectURL     string `mapstructure:"post_connect_url"`
	TokenEncryptionKey string `mapstructure:"token_encryption_key"`
	StateTTLSeconds    int    `mapstructure:"state_ttl_seconds"`
}

type GoogleIndexingConfig struct {
	Enabled               bool   `mapstructure:"enabled"`
	ServiceAccountJSON    string `mapstructure:"service_account_json"`
	ServiceAccountFile    string `mapstructure:"service_account_file"`
	RequestTimeoutSeconds int    `mapstructure:"request_timeout_seconds"`
}

type SocialOAuthProviderConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
	Scopes       string `mapstructure:"scopes"`
}

type SocialOAuthConfig struct {
	Meta                SocialOAuthProviderConfig `mapstructure:"meta"`
	X                   SocialOAuthProviderConfig `mapstructure:"x"`
	YouTube             SocialOAuthProviderConfig `mapstructure:"youtube"`
	Reddit              SocialOAuthProviderConfig `mapstructure:"reddit"`
	PostConnectURL      string                    `mapstructure:"post_connect_url"`
	TokenEncryptionKey  string                    `mapstructure:"token_encryption_key"`
	StateTTLSeconds     int                       `mapstructure:"state_ttl_seconds"`
	MetaGraphAPIVersion string                    `mapstructure:"meta_graph_api_version"`
}

type I18nConfig struct {
	DefaultLocale    string   `mapstructure:"default_locale"`
	SupportedLocales []string `mapstructure:"supported_locales"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type CookieConfig struct {
	Secure   string `mapstructure:"secure"`
	SameSite string `mapstructure:"same_site"`
	Domain   string `mapstructure:"domain"`
}

type CacheConfig struct {
	DefaultTTL     int `mapstructure:"default_ttl"`
	PostTTL        int `mapstructure:"post_ttl"`
	ProductTTL     int `mapstructure:"product_ttl"`
	ProductLockTTL int `mapstructure:"product_lock_ttl"`
	SettingsTTL    int `mapstructure:"settings_ttl"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type WorkerConfig struct {
	Enabled                               bool `mapstructure:"enabled"`
	TrackingPollingEnabled                bool `mapstructure:"tracking_polling_enabled"`
	TrackingPollingIntervalSeconds        int  `mapstructure:"tracking_polling_interval_seconds"`
	TrackingPollingBatchLimit             int  `mapstructure:"tracking_polling_batch_limit"`
	VisitorProfileCleanupEnabled          bool `mapstructure:"visitor_profile_cleanup_enabled"`
	VisitorProfileCleanupIntervalSeconds  int  `mapstructure:"visitor_profile_cleanup_interval_seconds"`
	VisitorProfileIPAddressRetentionDays  int  `mapstructure:"visitor_profile_ip_address_retention_days"`
	BehaviorEventCleanupEnabled           bool `mapstructure:"behavior_event_cleanup_enabled"`
	BehaviorEventCleanupIntervalSeconds   int  `mapstructure:"behavior_event_cleanup_interval_seconds"`
	OutboxDispatchEnabled                 bool `mapstructure:"outbox_dispatch_enabled"`
	OutboxDispatchIntervalSeconds         int  `mapstructure:"outbox_dispatch_interval_seconds"`
	OutboxDispatchBatchLimit              int  `mapstructure:"outbox_dispatch_batch_limit"`
	OutboxDispatchLockTimeoutSeconds      int  `mapstructure:"outbox_dispatch_lock_timeout_seconds"`
	PaymentExpirationEnabled              bool `mapstructure:"payment_expiration_enabled"`
	PaymentExpirationIntervalSeconds      int  `mapstructure:"payment_expiration_interval_seconds"`
	PaymentPendingTTLSeconds              int  `mapstructure:"payment_pending_ttl_seconds"`
	PaymentExpirationBatchLimit           int  `mapstructure:"payment_expiration_batch_limit"`
	PaymentRiskMonitoringEnabled          bool `mapstructure:"payment_risk_monitoring_enabled"`
	PaymentRiskMonitoringIntervalSeconds  int  `mapstructure:"payment_risk_monitoring_interval_seconds"`
	ExchangeRateSyncEnabled               bool `mapstructure:"exchange_rate_sync_enabled"`
	ExchangeRateSyncIntervalSeconds       int  `mapstructure:"exchange_rate_sync_interval_seconds"`
	SiteQualityEnabled                    bool `mapstructure:"site_quality_enabled"`
	SiteQualityAutoScanEnabled            bool `mapstructure:"site_quality_auto_scan_enabled"`
	SiteQualityDispatchIntervalSeconds    int  `mapstructure:"site_quality_dispatch_interval_seconds"`
	SiteQualityBatchLimit                 int  `mapstructure:"site_quality_batch_limit"`
	SiteQualityLeaseTimeoutSeconds        int  `mapstructure:"site_quality_lease_timeout_seconds"`
	SiteQualitySampleCount                int  `mapstructure:"site_quality_sample_count"`
	SiteQualityConfirmations              int  `mapstructure:"site_quality_confirmations"`
	SiteQualityCleanEvaluations           int  `mapstructure:"site_quality_clean_evaluations"`
	SiteQualityProviderConcurrency        int  `mapstructure:"site_quality_provider_concurrency"`
	SiteQualityProviderSpacingSeconds     int  `mapstructure:"site_quality_provider_spacing_seconds"`
	MediaDerivativeRebuildEnabled         bool `mapstructure:"media_derivative_rebuild_enabled"`
	MediaDerivativeRebuildIntervalSeconds int  `mapstructure:"media_derivative_rebuild_interval_seconds"`
	MediaDerivativeRebuildBatchLimit      int  `mapstructure:"media_derivative_rebuild_batch_limit"`
	ShowcaseCleanupEnabled                bool `mapstructure:"showcase_cleanup_enabled"`
	ShowcaseCleanupIntervalSeconds        int  `mapstructure:"showcase_cleanup_interval_seconds"`
	ShowcasePendingTTLSeconds             int  `mapstructure:"showcase_pending_ttl_seconds"`
	ShowcaseCleanupBatchLimit             int  `mapstructure:"showcase_cleanup_batch_limit"`
	HotDataArchiveEnabled                 bool `mapstructure:"hot_data_archive_enabled"`
	HotDataArchiveIntervalSeconds         int  `mapstructure:"hot_data_archive_interval_seconds"`
	HotDataArchiveBatchLimit              int  `mapstructure:"hot_data_archive_batch_limit"`
}

// SiteQualityConfig describes the internal-only Lighthouse runner endpoint.
// It is deployment wiring, never an operator-managed integration credential.
type SiteQualityConfig struct {
	RunnerURL   string `mapstructure:"runner_url"`
	RunnerToken string `mapstructure:"runner_token"`
}

// CustomerServiceRealtimeConfig controls the optional Redis Stream relay for
// durable customer-service message events. HTTP remains authoritative when it
// is disabled or the bounded replay window has expired.
type CustomerServiceRealtimeConfig struct {
	Enabled               bool   `mapstructure:"enabled"`
	Stream                string `mapstructure:"stream"`
	StreamMaxLen          int    `mapstructure:"stream_max_len"`
	ReplayLimit           int    `mapstructure:"replay_limit"`
	ConsumerBlockSeconds  int    `mapstructure:"consumer_block_seconds"`
	DedupRetentionSeconds int    `mapstructure:"dedup_retention_seconds"`
}

type BehaviorEventsConfig struct {
	LowIntentRetentionDays      int `mapstructure:"low_intent_retention_days"`
	StandardIntentRetentionDays int `mapstructure:"standard_intent_retention_days"`
	HighIntentRetentionDays     int `mapstructure:"high_intent_retention_days"`
	CleanupBatchLimit           int `mapstructure:"cleanup_batch_limit"`
}

type AntiAbuseConfig struct {
	TurnstileRequired                    bool   `mapstructure:"turnstile_required"`
	TurnstileSecretKey                   string `mapstructure:"turnstile_secret_key"`
	VerificationIPWindowSeconds          int    `mapstructure:"verification_ip_window_seconds"`
	VerificationDestinationWindowSeconds int    `mapstructure:"verification_destination_window_seconds"`
	VerificationDailyLimit               int    `mapstructure:"verification_daily_limit"`
	VerificationGlobalWindowSeconds      int    `mapstructure:"verification_global_window_seconds"`
	VerificationGlobalLimit              int    `mapstructure:"verification_global_limit"`
	VerificationCircuitSeconds           int    `mapstructure:"verification_circuit_seconds"`
}

type OrderAbuseConfig struct {
	Enabled                     bool `mapstructure:"enabled"`
	OrderCreateWindowSeconds    int  `mapstructure:"order_create_window_seconds"`
	MaxOrderCreationsPerUser    int  `mapstructure:"max_order_creations_per_user"`
	MaxOrderCreationsPerSession int  `mapstructure:"max_order_creations_per_session"`
	MaxOrderCreationsPerIP      int  `mapstructure:"max_order_creations_per_ip"`
}

type OrderNumberConfig struct {
	Secret         string `mapstructure:"secret"`
	PreviousSecret string `mapstructure:"previous_secret"`
	NodeID         uint16 `mapstructure:"node_id"`
}

func (c OrderNumberConfig) EffectiveSecret(jwtSecret string) string {
	if secret := strings.TrimSpace(c.Secret); secret != "" {
		return secret
	}
	return strings.TrimSpace(jwtSecret)
}

func (c OrderNumberConfig) EffectivePreviousSecret() string {
	return strings.TrimSpace(c.PreviousSecret)
}

type PaymentRiskConfig struct {
	FailureWindowSeconds int `mapstructure:"failure_window_seconds"`
	FailureThreshold     int `mapstructure:"failure_threshold"`
	DelaySeconds         int `mapstructure:"delay_seconds"`
	HighRiskScore        int `mapstructure:"high_risk_score"`
}

type PaymentBINRateLimitConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	WindowSeconds        int  `mapstructure:"window_seconds"`
	FailureThreshold     int  `mapstructure:"failure_threshold"`
	BlockDurationSeconds int  `mapstructure:"block_duration_seconds"`
}

type PaymentGatewayCircuitBreakerConfig struct {
	Enabled                 bool    `mapstructure:"enabled"`
	WindowSeconds           int     `mapstructure:"window_seconds"`
	FailureRateThreshold    float64 `mapstructure:"failure_rate_threshold"`
	MinimumSampleCount      int     `mapstructure:"minimum_sample_count"`
	OpenDurationSeconds     int     `mapstructure:"open_duration_seconds"`
	HalfOpenProbeTimeoutSec int     `mapstructure:"half_open_probe_timeout_seconds"`
}

// OutboundHTTPResilienceConfig protects non-payment third-party HTTP calls.
// State is stored in Redis so every API and worker replica observes the same
// circuit state and shares a single half-open probe.
type OutboundHTTPResilienceConfig struct {
	Enabled                 bool `mapstructure:"enabled"`
	RetryMaxAttempts        int  `mapstructure:"retry_max_attempts"`
	RetryBaseDelayMillis    int  `mapstructure:"retry_base_delay_millis"`
	RetryMaxDelayMillis     int  `mapstructure:"retry_max_delay_millis"`
	RetryJitterMillis       int  `mapstructure:"retry_jitter_millis"`
	FailureThreshold        int  `mapstructure:"failure_threshold"`
	FailureWindowSeconds    int  `mapstructure:"failure_window_seconds"`
	OpenDurationSeconds     int  `mapstructure:"open_duration_seconds"`
	HalfOpenProbeTimeoutSec int  `mapstructure:"half_open_probe_timeout_seconds"`
}

type PaymentRiskMonitoringConfig struct {
	Enabled                     bool    `mapstructure:"enabled"`
	AlertEnabled                bool    `mapstructure:"alert_enabled"`
	WindowDays                  int     `mapstructure:"window_days"`
	MinimumSuccessfulPayments   int     `mapstructure:"minimum_successful_payments"`
	WarningDisputeActivityRate  float64 `mapstructure:"warning_dispute_activity_rate"`
	CriticalDisputeActivityRate float64 `mapstructure:"critical_dispute_activity_rate"`
	WarningEarlyFraudRate       float64 `mapstructure:"warning_early_fraud_rate"`
	CriticalEarlyFraudRate      float64 `mapstructure:"critical_early_fraud_rate"`
	WarningRefundRate           float64 `mapstructure:"warning_refund_rate"`
	CriticalRefundRate          float64 `mapstructure:"critical_refund_rate"`
	AutoStepUpEnabled           bool    `mapstructure:"auto_step_up_enabled"`
}

type PaymentProtectionConfig struct {
	Enabled                            bool `mapstructure:"enabled"`
	MaxControlDurationHours            int  `mapstructure:"max_control_duration_hours"`
	MaxPausePaymentDurationHours       int  `mapstructure:"max_pause_payment_duration_hours"`
	MaxGlobalPausePaymentDurationHours int  `mapstructure:"max_global_pause_payment_duration_hours"`
}

type PaymentThreeDSConfig struct {
	AdaptiveEnabled                                 bool    `mapstructure:"adaptive_enabled"`
	LowRiskMaxAmount                                float64 `mapstructure:"low_risk_max_amount"`
	AVSBillingShippingMismatchHighValueThresholdUSD float64 `mapstructure:"avs_billing_shipping_mismatch_high_value_threshold_usd"`
	TrustedPaidOrders                               int     `mapstructure:"trusted_paid_orders"`
	VisitorRiskLookback                             int     `mapstructure:"visitor_risk_lookback_days"`
	StepUpRiskScore                                 int     `mapstructure:"step_up_risk_score"`
	ChallengeRiskScore                              int     `mapstructure:"challenge_risk_score"`
}

type VisitorRiskConfig struct {
	Enabled              bool   `mapstructure:"enabled"`
	HashSalt             string `mapstructure:"hash_salt"`
	FlushIntervalSeconds int    `mapstructure:"flush_interval_seconds"`
	MaxPendingFacts      int    `mapstructure:"max_pending_facts"`
	SamplePathLimit      int    `mapstructure:"sample_path_limit"`
	RetentionDays        int    `mapstructure:"retention_days"`
}

type RequestSigningConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	Key            string   `mapstructure:"key"`
	MaxSkewSeconds int      `mapstructure:"max_skew_seconds"`
	RequiredPaths  []string `mapstructure:"required_paths"`
}

type QuickBuyRateLimitConfig struct {
	Enabled                  bool `mapstructure:"enabled"`
	IPRequestsPerMinute      int  `mapstructure:"ip_requests_per_minute"`
	IPBurst                  int  `mapstructure:"ip_burst"`
	SessionRequestsPerMinute int  `mapstructure:"session_requests_per_minute"`
	SessionBurst             int  `mapstructure:"session_burst"`
	FailOpen                 bool `mapstructure:"fail_open"`
}

type FeedbackRateLimitConfig struct {
	Enabled                    bool `mapstructure:"enabled"`
	ReadIPRequestsPerMinute    int  `mapstructure:"read_ip_requests_per_minute"`
	ReadIPBurst                int  `mapstructure:"read_ip_burst"`
	WriteIPRequestsPerMinute   int  `mapstructure:"write_ip_requests_per_minute"`
	WriteIPBurst               int  `mapstructure:"write_ip_burst"`
	WriteUserRequestsPerMinute int  `mapstructure:"write_user_requests_per_minute"`
	WriteUserBurst             int  `mapstructure:"write_user_burst"`
	FailOpen                   bool `mapstructure:"fail_open"`
}

type MediaUploadConfig struct {
	AccountStorageQuotaBytes int64 `mapstructure:"account_storage_quota_bytes"`
}

type ShowcaseUploadProtectionConfig struct {
	Enabled                      bool  `mapstructure:"enabled"`
	WindowSeconds                int   `mapstructure:"window_seconds"`
	MaxUploadsPerUser            int   `mapstructure:"max_uploads_per_user"`
	MaxUploadsPerIP              int   `mapstructure:"max_uploads_per_ip"`
	MaxUploadsPerIPPrefix        int   `mapstructure:"max_uploads_per_ip_prefix"`
	DailyMaxUploadsPerUser       int   `mapstructure:"daily_max_uploads_per_user"`
	DailyMaxUploadsPerIP         int   `mapstructure:"daily_max_uploads_per_ip"`
	DailyMaxBytesPerUser         int64 `mapstructure:"daily_max_bytes_per_user"`
	DailyMaxBytesPerIP           int64 `mapstructure:"daily_max_bytes_per_ip"`
	MaxPendingSubmissionsPerUser int   `mapstructure:"max_pending_submissions_per_user"`
	FailureWindowSeconds         int   `mapstructure:"failure_window_seconds"`
	MaxFailuresPerUser           int   `mapstructure:"max_failures_per_user"`
	MaxFailuresPerIP             int   `mapstructure:"max_failures_per_ip"`
	BlockDurationSeconds         int   `mapstructure:"block_duration_seconds"`
}

// Load 加载配置文件
func Load(configFiles ...string) (*Config, error) {
	viper.Reset()

	if len(configFiles) > 0 && configFiles[0] != "" {
		viper.SetConfigFile(configFiles[0])
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	setDefaults()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvironment()

	// 允许环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	if origins := strings.TrimSpace(os.Getenv("CORS_ORIGINS")); origins != "" {
		viper.Set("cors.allowed_origins", splitEnvList(origins))
	}
	if proxies, configured := os.LookupEnv("TRUSTED_PROXIES"); configured {
		viper.Set("server.trusted_proxies", splitEnvList(proxies))
	}
	if paths, configured := os.LookupEnv("REQUEST_SIGNING_REQUIRED_PATHS"); configured {
		viper.Set("request_signing.required_paths", splitEnvList(paths))
	}
	if addrs, configured := os.LookupEnv("REDIS_ADDRS"); configured {
		viper.Set("redis.addrs", splitEnvList(addrs))
	}
	if addrs, configured := os.LookupEnv("REDIS_SENTINEL_ADDRS"); configured {
		viper.Set("redis.addrs", splitEnvList(addrs))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证关键配置
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.port", ":9200")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.base_url", "http://localhost:9200")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.read_header_timeout", 10)
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.idle_timeout", 120)
	viper.SetDefault("server.max_header_bytes", 1<<20)
	viper.SetDefault("server.trusted_proxies", []string{})

	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 9400)
	viper.SetDefault("database.username", "commerce_platform")
	viper.SetDefault("database.password", "commerce_platform_password")
	viper.SetDefault("database.database", "commerce_platform")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.log_level", "silent")

	viper.SetDefault("redis.mode", "standalone")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 9510)
	viper.SetDefault("redis.addrs", []string{})
	viper.SetDefault("redis.username", "")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.master_name", "")
	viper.SetDefault("redis.sentinel_username", "")
	viper.SetDefault("redis.sentinel_password", "")

	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.refresh_expire_hours", 168)

	viper.SetDefault("oauth.google_client_id", "")

	viper.SetDefault("google_merchant.client_id", "")
	viper.SetDefault("google_merchant.client_secret", "")
	viper.SetDefault("google_merchant.redirect_url", "")
	viper.SetDefault("google_merchant.post_connect_url", "http://localhost:9300/google-merchant")
	viper.SetDefault("google_merchant.token_encryption_key", "")
	viper.SetDefault("google_merchant.state_ttl_seconds", 600)

	viper.SetDefault("google_indexing.enabled", false)
	viper.SetDefault("google_indexing.service_account_json", "")
	viper.SetDefault("google_indexing.service_account_file", "")
	viper.SetDefault("google_indexing.request_timeout_seconds", 15)

	viper.SetDefault("social_oauth.meta.client_id", "")
	viper.SetDefault("social_oauth.meta.client_secret", "")
	viper.SetDefault("social_oauth.meta.redirect_url", "")
	viper.SetDefault("social_oauth.meta.scopes", "")
	viper.SetDefault("social_oauth.x.client_id", "")
	viper.SetDefault("social_oauth.x.client_secret", "")
	viper.SetDefault("social_oauth.x.redirect_url", "")
	viper.SetDefault("social_oauth.x.scopes", "")
	viper.SetDefault("social_oauth.youtube.client_id", "")
	viper.SetDefault("social_oauth.youtube.client_secret", "")
	viper.SetDefault("social_oauth.youtube.redirect_url", "")
	viper.SetDefault("social_oauth.youtube.scopes", "")
	viper.SetDefault("social_oauth.reddit.client_id", "")
	viper.SetDefault("social_oauth.reddit.client_secret", "")
	viper.SetDefault("social_oauth.reddit.redirect_url", "")
	viper.SetDefault("social_oauth.reddit.scopes", "")
	viper.SetDefault("social_oauth.post_connect_url", "http://localhost:9300/social")
	viper.SetDefault("social_oauth.token_encryption_key", "")
	viper.SetDefault("social_oauth.state_ttl_seconds", 600)
	viper.SetDefault("social_oauth.meta_graph_api_version", "v23.0")

	viper.SetDefault("i18n.default_locale", "en")
	viper.SetDefault("i18n.supported_locales", locales.SupportedLocaleCodes())

	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:9100", "http://localhost:9300"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{
		"Origin",
		"Content-Type",
		"Accept-Language",
		"Idempotency-Key",
		"X-CSRF-Token",
		"X-Locale",
		"X-Timezone",
		"X-Display-Currency",
		"X-Market-Country",
		"X-Country-Code",
		"X-Device-Fingerprint",
		"X-Request-Timestamp",
		"X-Request-Nonce",
		"X-Request-Signature",
		"X-Quick-Buy-Session",
		"X-Anonymous-ID",
	})
	viper.SetDefault("cors.expose_headers", []string{
		"Content-Length",
		"Idempotency-Replayed",
		"Retry-After",
	})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 43200)

	viper.SetDefault("cookie.secure", "auto")
	viper.SetDefault("cookie.same_site", "lax")
	viper.SetDefault("cookie.domain", "")

	viper.SetDefault("cache.default_ttl", 3600)
	viper.SetDefault("cache.post_ttl", 3600)
	viper.SetDefault("cache.product_ttl", 1800)
	viper.SetDefault("cache.product_lock_ttl", 5)
	viper.SetDefault("cache.settings_ttl", 7200)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	viper.SetDefault("worker.enabled", false)
	viper.SetDefault("worker.tracking_polling_enabled", false)
	viper.SetDefault("worker.tracking_polling_interval_seconds", 300)
	viper.SetDefault("worker.tracking_polling_batch_limit", 20)
	viper.SetDefault("worker.visitor_profile_cleanup_enabled", false)
	viper.SetDefault("worker.visitor_profile_cleanup_interval_seconds", 86400)
	viper.SetDefault("worker.visitor_profile_ip_address_retention_days", DefaultVisitorProfileIPAddressRetentionDays)
	viper.SetDefault("worker.behavior_event_cleanup_enabled", false)
	viper.SetDefault("worker.behavior_event_cleanup_interval_seconds", 86400)
	viper.SetDefault("worker.outbox_dispatch_enabled", false)
	viper.SetDefault("worker.outbox_dispatch_interval_seconds", 10)
	viper.SetDefault("worker.outbox_dispatch_batch_limit", 100)
	viper.SetDefault("worker.outbox_dispatch_lock_timeout_seconds", 300)
	viper.SetDefault("worker.payment_expiration_enabled", false)
	viper.SetDefault("worker.payment_expiration_interval_seconds", 900)
	viper.SetDefault("worker.payment_pending_ttl_seconds", 1800)
	viper.SetDefault("worker.payment_expiration_batch_limit", 100)
	viper.SetDefault("worker.payment_risk_monitoring_enabled", false)
	viper.SetDefault("worker.payment_risk_monitoring_interval_seconds", 3600)
	viper.SetDefault("worker.exchange_rate_sync_enabled", true)
	viper.SetDefault("worker.exchange_rate_sync_interval_seconds", 86400)
	viper.SetDefault("worker.site_quality_enabled", false)
	viper.SetDefault("worker.site_quality_auto_scan_enabled", false)
	viper.SetDefault("worker.site_quality_dispatch_interval_seconds", 30)
	viper.SetDefault("worker.site_quality_batch_limit", 1)
	viper.SetDefault("worker.site_quality_lease_timeout_seconds", 900)
	viper.SetDefault("worker.site_quality_sample_count", 3)
	viper.SetDefault("worker.site_quality_confirmations", 2)
	viper.SetDefault("worker.site_quality_clean_evaluations", 2)
	viper.SetDefault("worker.site_quality_provider_concurrency", 1)
	viper.SetDefault("worker.site_quality_provider_spacing_seconds", 5)
	viper.SetDefault("worker.media_derivative_rebuild_enabled", true)
	viper.SetDefault("worker.media_derivative_rebuild_interval_seconds", 5)
	viper.SetDefault("worker.media_derivative_rebuild_batch_limit", 10)
	viper.SetDefault("site_quality.runner_url", "")
	viper.SetDefault("site_quality.runner_token", "")
	viper.SetDefault("worker.showcase_cleanup_enabled", true)
	viper.SetDefault("worker.showcase_cleanup_interval_seconds", 86400)
	viper.SetDefault("worker.showcase_pending_ttl_seconds", 2592000)
	viper.SetDefault("worker.showcase_cleanup_batch_limit", 100)
	viper.SetDefault("worker.hot_data_archive_enabled", false)
	viper.SetDefault("worker.hot_data_archive_interval_seconds", 86400)
	viper.SetDefault("worker.hot_data_archive_batch_limit", 500)

	viper.SetDefault("customer_service_realtime.enabled", false)
	viper.SetDefault("customer_service_realtime.stream", "customer_service:{realtime}:v1")
	viper.SetDefault("customer_service_realtime.stream_max_len", 10000)
	viper.SetDefault("customer_service_realtime.replay_limit", 200)
	viper.SetDefault("customer_service_realtime.consumer_block_seconds", 5)
	viper.SetDefault("customer_service_realtime.dedup_retention_seconds", 86400)

	viper.SetDefault("behavior_events.low_intent_retention_days", 30)
	viper.SetDefault("behavior_events.standard_intent_retention_days", 60)
	viper.SetDefault("behavior_events.high_intent_retention_days", 180)
	viper.SetDefault("behavior_events.cleanup_batch_limit", 5000)

	viper.SetDefault("anti_abuse.turnstile_required", false)
	viper.SetDefault("anti_abuse.turnstile_secret_key", "")
	viper.SetDefault("anti_abuse.verification_ip_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_destination_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_daily_limit", 8)
	viper.SetDefault("anti_abuse.verification_global_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_global_limit", 100)
	viper.SetDefault("anti_abuse.verification_circuit_seconds", 300)

	viper.SetDefault("order_abuse.enabled", false)
	viper.SetDefault("order_abuse.order_create_window_seconds", 600)
	viper.SetDefault("order_abuse.max_order_creations_per_user", 3)
	viper.SetDefault("order_abuse.max_order_creations_per_session", 3)
	viper.SetDefault("order_abuse.max_order_creations_per_ip", 12)

	viper.SetDefault("order_number.secret", "")
	viper.SetDefault("order_number.previous_secret", "")
	viper.SetDefault("order_number.node_id", 0)

	viper.SetDefault("payment_risk.failure_window_seconds", 600)
	viper.SetDefault("payment_risk.failure_threshold", 3)
	viper.SetDefault("payment_risk.delay_seconds", 2)
	viper.SetDefault("payment_risk.high_risk_score", 60)

	viper.SetDefault("payment_bin_rate_limit.enabled", true)
	viper.SetDefault("payment_bin_rate_limit.window_seconds", 60)
	viper.SetDefault("payment_bin_rate_limit.failure_threshold", 5)
	viper.SetDefault("payment_bin_rate_limit.block_duration_seconds", 1800)

	viper.SetDefault("payment_gateway_circuit_breaker.enabled", true)
	viper.SetDefault("payment_gateway_circuit_breaker.window_seconds", 60)
	viper.SetDefault("payment_gateway_circuit_breaker.failure_rate_threshold", 0.15)
	viper.SetDefault("payment_gateway_circuit_breaker.minimum_sample_count", 20)
	viper.SetDefault("payment_gateway_circuit_breaker.open_duration_seconds", 30)
	viper.SetDefault(
		"payment_gateway_circuit_breaker.half_open_probe_timeout_seconds",
		MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds,
	)

	viper.SetDefault("outbound_http_resilience.enabled", true)
	viper.SetDefault("outbound_http_resilience.retry_max_attempts", 3)
	viper.SetDefault("outbound_http_resilience.retry_base_delay_millis", 250)
	viper.SetDefault("outbound_http_resilience.retry_max_delay_millis", 3000)
	viper.SetDefault("outbound_http_resilience.retry_jitter_millis", 250)
	viper.SetDefault("outbound_http_resilience.failure_threshold", 4)
	viper.SetDefault("outbound_http_resilience.failure_window_seconds", 60)
	viper.SetDefault("outbound_http_resilience.open_duration_seconds", 30)
	viper.SetDefault("outbound_http_resilience.half_open_probe_timeout_seconds", 30)

	viper.SetDefault("payment_risk_monitoring.enabled", true)
	viper.SetDefault("payment_risk_monitoring.alert_enabled", false)
	viper.SetDefault("payment_risk_monitoring.window_days", 30)
	viper.SetDefault("payment_risk_monitoring.minimum_successful_payments", 20)
	viper.SetDefault("payment_risk_monitoring.warning_dispute_activity_rate", 0.005)
	viper.SetDefault("payment_risk_monitoring.critical_dispute_activity_rate", 0.008)
	viper.SetDefault("payment_risk_monitoring.warning_early_fraud_rate", 0.005)
	viper.SetDefault("payment_risk_monitoring.critical_early_fraud_rate", 0.009)
	viper.SetDefault("payment_risk_monitoring.warning_refund_rate", 0.08)
	viper.SetDefault("payment_risk_monitoring.critical_refund_rate", 0.15)
	viper.SetDefault("payment_risk_monitoring.auto_step_up_enabled", true)

	viper.SetDefault("payment_protection.enabled", true)
	viper.SetDefault("payment_protection.max_control_duration_hours", 168)
	viper.SetDefault("payment_protection.max_pause_payment_duration_hours", 24)
	viper.SetDefault("payment_protection.max_global_pause_payment_duration_hours", 2)

	viper.SetDefault("payment_3ds.adaptive_enabled", true)
	viper.SetDefault("payment_3ds.low_risk_max_amount", 100.0)
	viper.SetDefault("payment_3ds.avs_billing_shipping_mismatch_high_value_threshold_usd", 800.0)
	viper.SetDefault("payment_3ds.trusted_paid_orders", 1)
	viper.SetDefault("payment_3ds.visitor_risk_lookback_days", 30)
	viper.SetDefault("payment_3ds.step_up_risk_score", 20)
	viper.SetDefault("payment_3ds.challenge_risk_score", 60)

	viper.SetDefault("visitor_risk.enabled", false)
	viper.SetDefault("visitor_risk.hash_salt", "")
	viper.SetDefault("visitor_risk.flush_interval_seconds", 60)
	viper.SetDefault("visitor_risk.max_pending_facts", 5000)
	viper.SetDefault("visitor_risk.sample_path_limit", 8)
	viper.SetDefault("visitor_risk.retention_days", 365)

	viper.SetDefault("request_signing.enabled", false)
	viper.SetDefault("request_signing.key", "")
	viper.SetDefault("request_signing.max_skew_seconds", 30)
	viper.SetDefault("request_signing.required_paths", []string{})

	viper.SetDefault("quick_buy_rate_limit.enabled", true)
	viper.SetDefault("quick_buy_rate_limit.ip_requests_per_minute", 120)
	viper.SetDefault("quick_buy_rate_limit.ip_burst", 40)
	viper.SetDefault("quick_buy_rate_limit.session_requests_per_minute", 60)
	viper.SetDefault("quick_buy_rate_limit.session_burst", 20)
	viper.SetDefault("quick_buy_rate_limit.fail_open", true)

	viper.SetDefault("feedback_rate_limit.enabled", true)
	viper.SetDefault("feedback_rate_limit.read_ip_requests_per_minute", 120)
	viper.SetDefault("feedback_rate_limit.read_ip_burst", 40)
	viper.SetDefault("feedback_rate_limit.write_ip_requests_per_minute", 12)
	viper.SetDefault("feedback_rate_limit.write_ip_burst", 4)
	viper.SetDefault("feedback_rate_limit.write_user_requests_per_minute", 6)
	viper.SetDefault("feedback_rate_limit.write_user_burst", 2)
	viper.SetDefault("feedback_rate_limit.fail_open", false)

	viper.SetDefault("media_upload.account_storage_quota_bytes", 20<<30)

	viper.SetDefault("showcase_upload_protection.enabled", true)
	viper.SetDefault("showcase_upload_protection.window_seconds", 60)
	viper.SetDefault("showcase_upload_protection.max_uploads_per_user", 3)
	viper.SetDefault("showcase_upload_protection.max_uploads_per_ip", 6)
	viper.SetDefault("showcase_upload_protection.max_uploads_per_ip_prefix", 18)
	viper.SetDefault("showcase_upload_protection.daily_max_uploads_per_user", 12)
	viper.SetDefault("showcase_upload_protection.daily_max_uploads_per_ip", 24)
	viper.SetDefault("showcase_upload_protection.daily_max_bytes_per_user", 100<<20)
	viper.SetDefault("showcase_upload_protection.daily_max_bytes_per_ip", 200<<20)
	viper.SetDefault("showcase_upload_protection.max_pending_submissions_per_user", 10)
	viper.SetDefault("showcase_upload_protection.failure_window_seconds", 900)
	viper.SetDefault("showcase_upload_protection.max_failures_per_user", 5)
	viper.SetDefault("showcase_upload_protection.max_failures_per_ip", 10)
	viper.SetDefault("showcase_upload_protection.block_duration_seconds", 1800)
}

func bindEnvironment() {
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.mode", "SERVER_MODE", "GIN_MODE")
	_ = viper.BindEnv("server.base_url", "SERVER_BASE_URL")
	_ = viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = viper.BindEnv("server.read_header_timeout", "SERVER_READ_HEADER_TIMEOUT")
	_ = viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = viper.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	_ = viper.BindEnv("server.max_header_bytes", "SERVER_MAX_HEADER_BYTES")

	_ = viper.BindEnv("database.driver", "DB_DRIVER", "DATABASE_DRIVER")
	_ = viper.BindEnv("database.host", "DB_HOST", "DATABASE_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT", "DATABASE_PORT")
	_ = viper.BindEnv("database.username", "DB_USERNAME", "DATABASE_USERNAME")
	_ = viper.BindEnv("database.password", "DB_PASSWORD", "DATABASE_PASSWORD")
	_ = viper.BindEnv("database.database", "DB_NAME", "DATABASE_NAME")
	_ = viper.BindEnv("database.auto_migrate", "DB_AUTO_MIGRATE", "DATABASE_AUTO_MIGRATE")
	_ = viper.BindEnv("database.log_level", "DB_LOG_LEVEL", "DATABASE_LOG_LEVEL")

	_ = viper.BindEnv("redis.mode", "REDIS_MODE")
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.username", "REDIS_USERNAME")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")
	_ = viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	_ = viper.BindEnv("redis.master_name", "REDIS_MASTER_NAME", "REDIS_SENTINEL_MASTER_NAME")
	_ = viper.BindEnv("redis.sentinel_username", "REDIS_SENTINEL_USERNAME")
	_ = viper.BindEnv("redis.sentinel_password", "REDIS_SENTINEL_PASSWORD")

	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")
	_ = viper.BindEnv("jwt.refresh_expire_hours", "JWT_REFRESH_EXPIRE_HOURS")

	_ = viper.BindEnv("oauth.google_client_id", "GOOGLE_CLIENT_ID", "GOOGLE_OAUTH_CLIENT_ID", "NUXT_PUBLIC_GOOGLE_CLIENT_ID")

	_ = viper.BindEnv("google_merchant.client_id", "GOOGLE_MERCHANT_CLIENT_ID")
	_ = viper.BindEnv("google_merchant.client_secret", "GOOGLE_MERCHANT_CLIENT_SECRET")
	_ = viper.BindEnv("google_merchant.redirect_url", "GOOGLE_MERCHANT_REDIRECT_URL")
	_ = viper.BindEnv("google_merchant.post_connect_url", "GOOGLE_MERCHANT_POST_CONNECT_URL")
	_ = viper.BindEnv("google_merchant.token_encryption_key", "GOOGLE_MERCHANT_TOKEN_ENCRYPTION_KEY")
	_ = viper.BindEnv("google_merchant.state_ttl_seconds", "GOOGLE_MERCHANT_STATE_TTL_SECONDS")

	_ = viper.BindEnv("google_indexing.enabled", "GOOGLE_INDEXING_ENABLED")
	_ = viper.BindEnv("google_indexing.service_account_json", "GOOGLE_INDEXING_SERVICE_ACCOUNT_JSON")
	_ = viper.BindEnv("google_indexing.service_account_file", "GOOGLE_INDEXING_SERVICE_ACCOUNT_FILE")
	_ = viper.BindEnv("google_indexing.request_timeout_seconds", "GOOGLE_INDEXING_REQUEST_TIMEOUT_SECONDS")

	_ = viper.BindEnv("social_oauth.meta.client_id", "SOCIAL_OAUTH_META_CLIENT_ID")
	_ = viper.BindEnv("social_oauth.meta.client_secret", "SOCIAL_OAUTH_META_CLIENT_SECRET")
	_ = viper.BindEnv("social_oauth.meta.redirect_url", "SOCIAL_OAUTH_META_REDIRECT_URL")
	_ = viper.BindEnv("social_oauth.meta.scopes", "SOCIAL_OAUTH_META_SCOPES")
	_ = viper.BindEnv("social_oauth.x.client_id", "SOCIAL_OAUTH_X_CLIENT_ID")
	_ = viper.BindEnv("social_oauth.x.client_secret", "SOCIAL_OAUTH_X_CLIENT_SECRET")
	_ = viper.BindEnv("social_oauth.x.redirect_url", "SOCIAL_OAUTH_X_REDIRECT_URL")
	_ = viper.BindEnv("social_oauth.x.scopes", "SOCIAL_OAUTH_X_SCOPES")
	_ = viper.BindEnv("social_oauth.youtube.client_id", "SOCIAL_OAUTH_YOUTUBE_CLIENT_ID")
	_ = viper.BindEnv("social_oauth.youtube.client_secret", "SOCIAL_OAUTH_YOUTUBE_CLIENT_SECRET")
	_ = viper.BindEnv("social_oauth.youtube.redirect_url", "SOCIAL_OAUTH_YOUTUBE_REDIRECT_URL")
	_ = viper.BindEnv("social_oauth.youtube.scopes", "SOCIAL_OAUTH_YOUTUBE_SCOPES")
	_ = viper.BindEnv("social_oauth.reddit.client_id", "SOCIAL_OAUTH_REDDIT_CLIENT_ID")
	_ = viper.BindEnv("social_oauth.reddit.client_secret", "SOCIAL_OAUTH_REDDIT_CLIENT_SECRET")
	_ = viper.BindEnv("social_oauth.reddit.redirect_url", "SOCIAL_OAUTH_REDDIT_REDIRECT_URL")
	_ = viper.BindEnv("social_oauth.reddit.scopes", "SOCIAL_OAUTH_REDDIT_SCOPES")
	_ = viper.BindEnv("social_oauth.post_connect_url", "SOCIAL_OAUTH_POST_CONNECT_URL")
	_ = viper.BindEnv("social_oauth.token_encryption_key", "SOCIAL_OAUTH_TOKEN_ENCRYPTION_KEY")
	_ = viper.BindEnv("social_oauth.state_ttl_seconds", "SOCIAL_OAUTH_STATE_TTL_SECONDS")
	_ = viper.BindEnv("social_oauth.meta_graph_api_version", "SOCIAL_OAUTH_META_GRAPH_API_VERSION")

	_ = viper.BindEnv("cookie.secure", "COOKIE_SECURE")
	_ = viper.BindEnv("cookie.same_site", "COOKIE_SAME_SITE")
	_ = viper.BindEnv("cookie.domain", "COOKIE_DOMAIN")

	_ = viper.BindEnv("cache.default_ttl", "CACHE_DEFAULT_TTL")
	_ = viper.BindEnv("cache.post_ttl", "CACHE_POST_TTL")
	_ = viper.BindEnv("cache.product_ttl", "CACHE_PRODUCT_TTL")
	_ = viper.BindEnv("cache.product_lock_ttl", "CACHE_PRODUCT_LOCK_TTL")
	_ = viper.BindEnv("cache.settings_ttl", "CACHE_SETTINGS_TTL")

	_ = viper.BindEnv("log.level", "LOG_LEVEL")
	_ = viper.BindEnv("log.format", "LOG_FORMAT")
	_ = viper.BindEnv("log.output", "LOG_OUTPUT")

	_ = viper.BindEnv("worker.enabled", "WORKER_ENABLED", "ASYNQ_WORKER_ENABLED")
	_ = viper.BindEnv("worker.tracking_polling_enabled", "WORKER_TRACKING_POLLING_ENABLED", "TRACKING_POLLING_ENABLED")
	_ = viper.BindEnv("worker.tracking_polling_interval_seconds", "WORKER_TRACKING_POLLING_INTERVAL_SECONDS", "TRACKING_POLLING_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.tracking_polling_batch_limit", "WORKER_TRACKING_POLLING_BATCH_LIMIT", "TRACKING_POLLING_BATCH_LIMIT")
	_ = viper.BindEnv("worker.visitor_profile_cleanup_enabled", "WORKER_VISITOR_PROFILE_CLEANUP_ENABLED", "VISITOR_PROFILE_CLEANUP_ENABLED")
	_ = viper.BindEnv("worker.visitor_profile_cleanup_interval_seconds", "WORKER_VISITOR_PROFILE_CLEANUP_INTERVAL_SECONDS", "VISITOR_PROFILE_CLEANUP_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.visitor_profile_ip_address_retention_days", "WORKER_VISITOR_PROFILE_IP_ADDRESS_RETENTION_DAYS", "VISITOR_PROFILE_IP_ADDRESS_RETENTION_DAYS")
	_ = viper.BindEnv("worker.behavior_event_cleanup_enabled", "WORKER_BEHAVIOR_EVENT_CLEANUP_ENABLED", "BEHAVIOR_EVENT_CLEANUP_ENABLED")
	_ = viper.BindEnv("worker.behavior_event_cleanup_interval_seconds", "WORKER_BEHAVIOR_EVENT_CLEANUP_INTERVAL_SECONDS", "BEHAVIOR_EVENT_CLEANUP_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.outbox_dispatch_enabled", "WORKER_OUTBOX_DISPATCH_ENABLED", "OUTBOX_DISPATCH_ENABLED")
	_ = viper.BindEnv("worker.outbox_dispatch_interval_seconds", "WORKER_OUTBOX_DISPATCH_INTERVAL_SECONDS", "OUTBOX_DISPATCH_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.outbox_dispatch_batch_limit", "WORKER_OUTBOX_DISPATCH_BATCH_LIMIT", "OUTBOX_DISPATCH_BATCH_LIMIT")
	_ = viper.BindEnv("worker.outbox_dispatch_lock_timeout_seconds", "WORKER_OUTBOX_DISPATCH_LOCK_TIMEOUT_SECONDS", "OUTBOX_DISPATCH_LOCK_TIMEOUT_SECONDS")
	_ = viper.BindEnv("worker.payment_expiration_enabled", "WORKER_PAYMENT_EXPIRATION_ENABLED", "PAYMENT_EXPIRATION_ENABLED")
	_ = viper.BindEnv("worker.payment_expiration_interval_seconds", "WORKER_PAYMENT_EXPIRATION_INTERVAL_SECONDS", "PAYMENT_EXPIRATION_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.payment_pending_ttl_seconds", "WORKER_PAYMENT_PENDING_TTL_SECONDS", "PAYMENT_PENDING_TTL_SECONDS")
	_ = viper.BindEnv("worker.payment_expiration_batch_limit", "WORKER_PAYMENT_EXPIRATION_BATCH_LIMIT", "PAYMENT_EXPIRATION_BATCH_LIMIT")
	_ = viper.BindEnv("worker.payment_risk_monitoring_enabled", "WORKER_PAYMENT_RISK_MONITORING_ENABLED", "PAYMENT_RISK_MONITORING_WORKER_ENABLED")
	_ = viper.BindEnv("worker.payment_risk_monitoring_interval_seconds", "WORKER_PAYMENT_RISK_MONITORING_INTERVAL_SECONDS", "PAYMENT_RISK_MONITORING_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.exchange_rate_sync_enabled", "WORKER_EXCHANGE_RATE_SYNC_ENABLED", "EXCHANGE_RATE_SYNC_ENABLED")
	_ = viper.BindEnv("worker.exchange_rate_sync_interval_seconds", "WORKER_EXCHANGE_RATE_SYNC_INTERVAL_SECONDS", "EXCHANGE_RATE_SYNC_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.site_quality_enabled", "WORKER_SITE_QUALITY_ENABLED")
	_ = viper.BindEnv("worker.site_quality_auto_scan_enabled", "WORKER_SITE_QUALITY_AUTO_SCAN_ENABLED")
	_ = viper.BindEnv("worker.site_quality_dispatch_interval_seconds", "WORKER_SITE_QUALITY_DISPATCH_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.site_quality_batch_limit", "WORKER_SITE_QUALITY_BATCH_LIMIT")
	_ = viper.BindEnv("worker.site_quality_lease_timeout_seconds", "WORKER_SITE_QUALITY_LEASE_TIMEOUT_SECONDS")
	_ = viper.BindEnv("worker.site_quality_sample_count", "WORKER_SITE_QUALITY_SAMPLE_COUNT")
	_ = viper.BindEnv("worker.site_quality_confirmations", "WORKER_SITE_QUALITY_CONFIRMATIONS")
	_ = viper.BindEnv("worker.site_quality_clean_evaluations", "WORKER_SITE_QUALITY_CLEAN_EVALUATIONS")
	_ = viper.BindEnv("worker.site_quality_provider_concurrency", "WORKER_SITE_QUALITY_PROVIDER_CONCURRENCY")
	_ = viper.BindEnv("worker.site_quality_provider_spacing_seconds", "WORKER_SITE_QUALITY_PROVIDER_SPACING_SECONDS")
	_ = viper.BindEnv("worker.media_derivative_rebuild_enabled", "WORKER_MEDIA_DERIVATIVE_REBUILD_ENABLED")
	_ = viper.BindEnv("worker.media_derivative_rebuild_interval_seconds", "WORKER_MEDIA_DERIVATIVE_REBUILD_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.media_derivative_rebuild_batch_limit", "WORKER_MEDIA_DERIVATIVE_REBUILD_BATCH_LIMIT")
	_ = viper.BindEnv("site_quality.runner_url", "SITE_QUALITY_RUNNER_URL")
	_ = viper.BindEnv("site_quality.runner_token", "SITE_QUALITY_RUNNER_TOKEN")
	_ = viper.BindEnv("worker.showcase_cleanup_enabled", "WORKER_SHOWCASE_CLEANUP_ENABLED", "SHOWCASE_CLEANUP_ENABLED")
	_ = viper.BindEnv("worker.showcase_cleanup_interval_seconds", "WORKER_SHOWCASE_CLEANUP_INTERVAL_SECONDS", "SHOWCASE_CLEANUP_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.showcase_pending_ttl_seconds", "WORKER_SHOWCASE_PENDING_TTL_SECONDS", "SHOWCASE_PENDING_TTL_SECONDS")
	_ = viper.BindEnv("worker.showcase_cleanup_batch_limit", "WORKER_SHOWCASE_CLEANUP_BATCH_LIMIT", "SHOWCASE_CLEANUP_BATCH_LIMIT")
	_ = viper.BindEnv("worker.hot_data_archive_enabled", "WORKER_HOT_DATA_ARCHIVE_ENABLED")
	_ = viper.BindEnv("worker.hot_data_archive_interval_seconds", "WORKER_HOT_DATA_ARCHIVE_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.hot_data_archive_batch_limit", "WORKER_HOT_DATA_ARCHIVE_BATCH_LIMIT")

	_ = viper.BindEnv("customer_service_realtime.enabled", "CUSTOMER_SERVICE_REALTIME_ENABLED")
	_ = viper.BindEnv("customer_service_realtime.stream", "CUSTOMER_SERVICE_REALTIME_STREAM")
	_ = viper.BindEnv("customer_service_realtime.stream_max_len", "CUSTOMER_SERVICE_REALTIME_STREAM_MAX_LEN")
	_ = viper.BindEnv("customer_service_realtime.replay_limit", "CUSTOMER_SERVICE_REALTIME_REPLAY_LIMIT")
	_ = viper.BindEnv("customer_service_realtime.consumer_block_seconds", "CUSTOMER_SERVICE_REALTIME_CONSUMER_BLOCK_SECONDS")
	_ = viper.BindEnv("customer_service_realtime.dedup_retention_seconds", "CUSTOMER_SERVICE_REALTIME_DEDUP_RETENTION_SECONDS")

	_ = viper.BindEnv("behavior_events.low_intent_retention_days", "BEHAVIOR_EVENTS_LOW_INTENT_RETENTION_DAYS")
	_ = viper.BindEnv("behavior_events.standard_intent_retention_days", "BEHAVIOR_EVENTS_STANDARD_INTENT_RETENTION_DAYS")
	_ = viper.BindEnv("behavior_events.high_intent_retention_days", "BEHAVIOR_EVENTS_HIGH_INTENT_RETENTION_DAYS")
	_ = viper.BindEnv("behavior_events.cleanup_batch_limit", "BEHAVIOR_EVENTS_CLEANUP_BATCH_LIMIT")

	_ = viper.BindEnv("anti_abuse.turnstile_required", "TURNSTILE_REQUIRED")
	_ = viper.BindEnv("anti_abuse.turnstile_secret_key", "TURNSTILE_SECRET_KEY")
	_ = viper.BindEnv("anti_abuse.verification_ip_window_seconds", "VERIFICATION_IP_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_destination_window_seconds", "VERIFICATION_DESTINATION_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_daily_limit", "VERIFICATION_DAILY_LIMIT")
	_ = viper.BindEnv("anti_abuse.verification_global_window_seconds", "VERIFICATION_GLOBAL_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_global_limit", "VERIFICATION_GLOBAL_LIMIT")
	_ = viper.BindEnv("anti_abuse.verification_circuit_seconds", "VERIFICATION_CIRCUIT_SECONDS")

	_ = viper.BindEnv("order_abuse.enabled", "ORDER_ABUSE_ENABLED")
	_ = viper.BindEnv("order_abuse.order_create_window_seconds", "ORDER_ABUSE_ORDER_CREATE_WINDOW_SECONDS")
	_ = viper.BindEnv("order_abuse.max_order_creations_per_user", "ORDER_ABUSE_MAX_ORDER_CREATIONS_PER_USER")
	_ = viper.BindEnv("order_abuse.max_order_creations_per_session", "ORDER_ABUSE_MAX_ORDER_CREATIONS_PER_SESSION")
	_ = viper.BindEnv("order_abuse.max_order_creations_per_ip", "ORDER_ABUSE_MAX_ORDER_CREATIONS_PER_IP")

	_ = viper.BindEnv("order_number.secret", "ORDER_NUMBER_SECRET")
	_ = viper.BindEnv("order_number.previous_secret", "ORDER_NUMBER_PREVIOUS_SECRET")
	_ = viper.BindEnv("order_number.node_id", "ORDER_NUMBER_NODE_ID")

	_ = viper.BindEnv("payment_risk.failure_window_seconds", "PAYMENT_RISK_FAILURE_WINDOW_SECONDS")
	_ = viper.BindEnv("payment_risk.failure_threshold", "PAYMENT_RISK_FAILURE_THRESHOLD")
	_ = viper.BindEnv("payment_risk.delay_seconds", "PAYMENT_RISK_DELAY_SECONDS")
	_ = viper.BindEnv("payment_risk.high_risk_score", "PAYMENT_RISK_HIGH_RISK_SCORE")

	_ = viper.BindEnv("payment_bin_rate_limit.enabled", "PAYMENT_BIN_RATE_LIMIT_ENABLED")
	_ = viper.BindEnv("payment_bin_rate_limit.window_seconds", "PAYMENT_BIN_RATE_LIMIT_WINDOW_SECONDS")
	_ = viper.BindEnv("payment_bin_rate_limit.failure_threshold", "PAYMENT_BIN_RATE_LIMIT_FAILURE_THRESHOLD")
	_ = viper.BindEnv("payment_bin_rate_limit.block_duration_seconds", "PAYMENT_BIN_RATE_LIMIT_BLOCK_DURATION_SECONDS")

	_ = viper.BindEnv("payment_gateway_circuit_breaker.enabled", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_ENABLED")
	_ = viper.BindEnv("payment_gateway_circuit_breaker.window_seconds", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_WINDOW_SECONDS")
	_ = viper.BindEnv("payment_gateway_circuit_breaker.failure_rate_threshold", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_FAILURE_RATE_THRESHOLD")
	_ = viper.BindEnv("payment_gateway_circuit_breaker.minimum_sample_count", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_MINIMUM_SAMPLE_COUNT")
	_ = viper.BindEnv("payment_gateway_circuit_breaker.open_duration_seconds", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_OPEN_DURATION_SECONDS")
	_ = viper.BindEnv("payment_gateway_circuit_breaker.half_open_probe_timeout_seconds", "PAYMENT_GATEWAY_CIRCUIT_BREAKER_HALF_OPEN_PROBE_TIMEOUT_SECONDS")

	_ = viper.BindEnv("outbound_http_resilience.enabled", "OUTBOUND_HTTP_RESILIENCE_ENABLED")
	_ = viper.BindEnv("outbound_http_resilience.retry_max_attempts", "OUTBOUND_HTTP_RESILIENCE_RETRY_MAX_ATTEMPTS")
	_ = viper.BindEnv("outbound_http_resilience.retry_base_delay_millis", "OUTBOUND_HTTP_RESILIENCE_RETRY_BASE_DELAY_MILLIS")
	_ = viper.BindEnv("outbound_http_resilience.retry_max_delay_millis", "OUTBOUND_HTTP_RESILIENCE_RETRY_MAX_DELAY_MILLIS")
	_ = viper.BindEnv("outbound_http_resilience.retry_jitter_millis", "OUTBOUND_HTTP_RESILIENCE_RETRY_JITTER_MILLIS")
	_ = viper.BindEnv("outbound_http_resilience.failure_threshold", "OUTBOUND_HTTP_RESILIENCE_FAILURE_THRESHOLD")
	_ = viper.BindEnv("outbound_http_resilience.failure_window_seconds", "OUTBOUND_HTTP_RESILIENCE_FAILURE_WINDOW_SECONDS")
	_ = viper.BindEnv("outbound_http_resilience.open_duration_seconds", "OUTBOUND_HTTP_RESILIENCE_OPEN_DURATION_SECONDS")
	_ = viper.BindEnv("outbound_http_resilience.half_open_probe_timeout_seconds", "OUTBOUND_HTTP_RESILIENCE_HALF_OPEN_PROBE_TIMEOUT_SECONDS")

	_ = viper.BindEnv("payment_risk_monitoring.enabled", "PAYMENT_RISK_MONITORING_ENABLED")
	_ = viper.BindEnv("payment_risk_monitoring.alert_enabled", "PAYMENT_RISK_MONITORING_ALERT_ENABLED")
	_ = viper.BindEnv("payment_risk_monitoring.window_days", "PAYMENT_RISK_MONITORING_WINDOW_DAYS")
	_ = viper.BindEnv("payment_risk_monitoring.minimum_successful_payments", "PAYMENT_RISK_MONITORING_MINIMUM_SUCCESSFUL_PAYMENTS")
	_ = viper.BindEnv("payment_risk_monitoring.warning_dispute_activity_rate", "PAYMENT_RISK_MONITORING_WARNING_DISPUTE_ACTIVITY_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.critical_dispute_activity_rate", "PAYMENT_RISK_MONITORING_CRITICAL_DISPUTE_ACTIVITY_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.warning_early_fraud_rate", "PAYMENT_RISK_MONITORING_WARNING_EARLY_FRAUD_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.critical_early_fraud_rate", "PAYMENT_RISK_MONITORING_CRITICAL_EARLY_FRAUD_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.warning_refund_rate", "PAYMENT_RISK_MONITORING_WARNING_REFUND_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.critical_refund_rate", "PAYMENT_RISK_MONITORING_CRITICAL_REFUND_RATE")
	_ = viper.BindEnv("payment_risk_monitoring.auto_step_up_enabled", "PAYMENT_RISK_MONITORING_AUTO_STEP_UP_ENABLED")

	_ = viper.BindEnv("payment_protection.enabled", "PAYMENT_PROTECTION_ENABLED")
	_ = viper.BindEnv("payment_protection.max_control_duration_hours", "PAYMENT_PROTECTION_MAX_CONTROL_DURATION_HOURS")
	_ = viper.BindEnv("payment_protection.max_pause_payment_duration_hours", "PAYMENT_PROTECTION_MAX_PAUSE_PAYMENT_DURATION_HOURS")
	_ = viper.BindEnv("payment_protection.max_global_pause_payment_duration_hours", "PAYMENT_PROTECTION_MAX_GLOBAL_PAUSE_PAYMENT_DURATION_HOURS")

	_ = viper.BindEnv("payment_3ds.adaptive_enabled", "PAYMENT_3DS_ADAPTIVE_ENABLED")
	_ = viper.BindEnv("payment_3ds.low_risk_max_amount", "PAYMENT_3DS_LOW_RISK_MAX_AMOUNT")
	_ = viper.BindEnv("payment_3ds.avs_billing_shipping_mismatch_high_value_threshold_usd", "PAYMENT_3DS_AVS_BILLING_SHIPPING_MISMATCH_HIGH_VALUE_THRESHOLD_USD")
	_ = viper.BindEnv("payment_3ds.trusted_paid_orders", "PAYMENT_3DS_TRUSTED_PAID_ORDERS")
	_ = viper.BindEnv("payment_3ds.visitor_risk_lookback_days", "PAYMENT_3DS_VISITOR_RISK_LOOKBACK_DAYS")
	_ = viper.BindEnv("payment_3ds.step_up_risk_score", "PAYMENT_3DS_STEP_UP_RISK_SCORE")
	_ = viper.BindEnv("payment_3ds.challenge_risk_score", "PAYMENT_3DS_CHALLENGE_RISK_SCORE")

	_ = viper.BindEnv("visitor_risk.enabled", "VISITOR_RISK_ENABLED")
	_ = viper.BindEnv("visitor_risk.hash_salt", "VISITOR_RISK_HASH_SALT")
	_ = viper.BindEnv("visitor_risk.flush_interval_seconds", "VISITOR_RISK_FLUSH_INTERVAL_SECONDS")
	_ = viper.BindEnv("visitor_risk.max_pending_facts", "VISITOR_RISK_MAX_PENDING_FACTS")
	_ = viper.BindEnv("visitor_risk.sample_path_limit", "VISITOR_RISK_SAMPLE_PATH_LIMIT")
	_ = viper.BindEnv("visitor_risk.retention_days", "VISITOR_RISK_RETENTION_DAYS")

	_ = viper.BindEnv("request_signing.enabled", "REQUEST_SIGNING_ENABLED")
	_ = viper.BindEnv("request_signing.key", "REQUEST_SIGNING_KEY")
	_ = viper.BindEnv("request_signing.max_skew_seconds", "REQUEST_SIGNING_MAX_SKEW_SECONDS")
	_ = viper.BindEnv("request_signing.required_paths", "REQUEST_SIGNING_REQUIRED_PATHS")

	_ = viper.BindEnv("quick_buy_rate_limit.enabled", "QUICK_BUY_RATE_LIMIT_ENABLED")
	_ = viper.BindEnv("quick_buy_rate_limit.ip_requests_per_minute", "QUICK_BUY_RATE_LIMIT_IP_REQUESTS_PER_MINUTE")
	_ = viper.BindEnv("quick_buy_rate_limit.ip_burst", "QUICK_BUY_RATE_LIMIT_IP_BURST")
	_ = viper.BindEnv("quick_buy_rate_limit.session_requests_per_minute", "QUICK_BUY_RATE_LIMIT_SESSION_REQUESTS_PER_MINUTE")
	_ = viper.BindEnv("quick_buy_rate_limit.session_burst", "QUICK_BUY_RATE_LIMIT_SESSION_BURST")
	_ = viper.BindEnv("quick_buy_rate_limit.fail_open", "QUICK_BUY_RATE_LIMIT_FAIL_OPEN")

	_ = viper.BindEnv("feedback_rate_limit.enabled", "FEEDBACK_RATE_LIMIT_ENABLED")
	_ = viper.BindEnv("feedback_rate_limit.read_ip_requests_per_minute", "FEEDBACK_RATE_LIMIT_READ_IP_REQUESTS_PER_MINUTE")
	_ = viper.BindEnv("feedback_rate_limit.read_ip_burst", "FEEDBACK_RATE_LIMIT_READ_IP_BURST")
	_ = viper.BindEnv("feedback_rate_limit.write_ip_requests_per_minute", "FEEDBACK_RATE_LIMIT_WRITE_IP_REQUESTS_PER_MINUTE")
	_ = viper.BindEnv("feedback_rate_limit.write_ip_burst", "FEEDBACK_RATE_LIMIT_WRITE_IP_BURST")
	_ = viper.BindEnv("feedback_rate_limit.write_user_requests_per_minute", "FEEDBACK_RATE_LIMIT_WRITE_USER_REQUESTS_PER_MINUTE")
	_ = viper.BindEnv("feedback_rate_limit.write_user_burst", "FEEDBACK_RATE_LIMIT_WRITE_USER_BURST")
	_ = viper.BindEnv("feedback_rate_limit.fail_open", "FEEDBACK_RATE_LIMIT_FAIL_OPEN")

	_ = viper.BindEnv("media_upload.account_storage_quota_bytes", "MEDIA_UPLOAD_ACCOUNT_STORAGE_QUOTA_BYTES")

	_ = viper.BindEnv("showcase_upload_protection.enabled", "SHOWCASE_UPLOAD_PROTECTION_ENABLED")
	_ = viper.BindEnv("showcase_upload_protection.window_seconds", "SHOWCASE_UPLOAD_WINDOW_SECONDS")
	_ = viper.BindEnv("showcase_upload_protection.max_uploads_per_user", "SHOWCASE_UPLOAD_MAX_UPLOADS_PER_USER")
	_ = viper.BindEnv("showcase_upload_protection.max_uploads_per_ip", "SHOWCASE_UPLOAD_MAX_UPLOADS_PER_IP")
	_ = viper.BindEnv("showcase_upload_protection.max_uploads_per_ip_prefix", "SHOWCASE_UPLOAD_MAX_UPLOADS_PER_IP_PREFIX")
	_ = viper.BindEnv("showcase_upload_protection.daily_max_uploads_per_user", "SHOWCASE_UPLOAD_DAILY_MAX_UPLOADS_PER_USER")
	_ = viper.BindEnv("showcase_upload_protection.daily_max_uploads_per_ip", "SHOWCASE_UPLOAD_DAILY_MAX_UPLOADS_PER_IP")
	_ = viper.BindEnv("showcase_upload_protection.daily_max_bytes_per_user", "SHOWCASE_UPLOAD_DAILY_MAX_BYTES_PER_USER")
	_ = viper.BindEnv("showcase_upload_protection.daily_max_bytes_per_ip", "SHOWCASE_UPLOAD_DAILY_MAX_BYTES_PER_IP")
	_ = viper.BindEnv("showcase_upload_protection.max_pending_submissions_per_user", "SHOWCASE_UPLOAD_MAX_PENDING_SUBMISSIONS_PER_USER")
	_ = viper.BindEnv("showcase_upload_protection.failure_window_seconds", "SHOWCASE_UPLOAD_FAILURE_WINDOW_SECONDS")
	_ = viper.BindEnv("showcase_upload_protection.max_failures_per_user", "SHOWCASE_UPLOAD_MAX_FAILURES_PER_USER")
	_ = viper.BindEnv("showcase_upload_protection.max_failures_per_ip", "SHOWCASE_UPLOAD_MAX_FAILURES_PER_IP")
	_ = viper.BindEnv("showcase_upload_protection.block_duration_seconds", "SHOWCASE_UPLOAD_BLOCK_DURATION_SECONDS")
}

func splitEnvList(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
			c.Host, c.Port, c.Username, c.Password, c.Database,
		)
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database,
		)
	default:
		return ""
	}
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c RedisConfig) NormalizedMode() string {
	switch strings.ToLower(strings.TrimSpace(c.Mode)) {
	case "", "single", "standalone":
		return "standalone"
	case "sentinel", "failover":
		return "sentinel"
	case "cluster":
		return "cluster"
	default:
		return strings.ToLower(strings.TrimSpace(c.Mode))
	}
}

func (c RedisConfig) GetRedisAddrs() []string {
	addrs := make([]string, 0, len(c.Addrs)+1)
	for _, addr := range c.Addrs {
		if trimmed := strings.TrimSpace(addr); trimmed != "" {
			addrs = append(addrs, trimmed)
		}
	}
	if len(addrs) == 0 && strings.TrimSpace(c.Host) != "" {
		addrs = append(addrs, c.GetRedisAddr())
	}
	return addrs
}

func (c *CookieConfig) SecureEnabled(server ServerConfig) bool {
	switch strings.ToLower(strings.TrimSpace(c.Secure)) {
	case "false", "no", "never", "off", "0":
		return false
	case "true", "yes", "always", "on", "1":
		return true
	default:
		baseURL := strings.ToLower(strings.TrimSpace(server.BaseURL))
		return strings.EqualFold(server.Mode, "release") || strings.HasPrefix(baseURL, "https://")
	}
}

func (c *CookieConfig) SameSiteMode() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(c.SameSite)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// GetJWTExpireDuration 获取JWT过期时间
func (c *JWTConfig) GetJWTExpireDuration() time.Duration {
	return time.Duration(c.ExpireHours) * time.Hour
}

// GetRefreshExpireDuration 获取刷新令牌过期时间
func (c *JWTConfig) GetRefreshExpireDuration() time.Duration {
	return time.Duration(c.RefreshExpireHours) * time.Hour
}

func validateRedisConfig(cfg RedisConfig) error {
	switch cfg.NormalizedMode() {
	case "standalone":
		if len(cfg.GetRedisAddrs()) != 1 {
			return fmt.Errorf("standalone Redis requires exactly one address")
		}
	case "sentinel":
		if len(cfg.Addrs) == 0 {
			return fmt.Errorf("Redis Sentinel requires at least one sentinel address")
		}
		if strings.TrimSpace(cfg.MasterName) == "" {
			return fmt.Errorf("Redis Sentinel requires redis.master_name or REDIS_MASTER_NAME")
		}
	case "cluster":
		return fmt.Errorf("Redis Cluster mode is not enabled yet: multi-key rate limit and anti-abuse scripts still need cluster-wide hash-tag migration")
	default:
		return fmt.Errorf("unsupported Redis mode %q", cfg.Mode)
	}
	if cfg.Port <= 0 && len(cfg.Addrs) == 0 {
		return fmt.Errorf("Redis port must be positive")
	}
	if cfg.DB < 0 {
		return fmt.Errorf("Redis DB must not be negative")
	}
	if cfg.PoolSize < 0 {
		return fmt.Errorf("Redis pool size must not be negative")
	}
	return nil
}

const minimumTrustedProxyIPv4PrefixRelease = 24

func validateTrustedProxies(values []string, releaseMode bool) error {
	hasProxy := false
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		hasProxy = true

		var (
			ip           net.IP
			network      *net.IPNet
			prefixLength int
			addressBits  int
		)
		if strings.Contains(value, "/") {
			parsedIP, parsedNetwork, err := net.ParseCIDR(value)
			if err != nil {
				return fmt.Errorf("invalid trusted proxy network %q: %w", value, err)
			}
			ip = parsedIP
			network = parsedNetwork
			prefixLength, addressBits = parsedNetwork.Mask.Size()
			if prefixLength < 0 {
				return fmt.Errorf("invalid trusted proxy network %q", value)
			}
		} else {
			ip = net.ParseIP(value)
			if ip == nil {
				return fmt.Errorf("invalid trusted proxy address %q", value)
			}
			if ip4 := ip.To4(); ip4 != nil {
				ip = ip4
				addressBits = 32
				prefixLength = 32
			} else {
				addressBits = 128
				prefixLength = 128
			}
			network = &net.IPNet{
				IP:   ip,
				Mask: net.CIDRMask(addressBits, addressBits),
			}
		}

		if ip.IsUnspecified() || prefixLength == 0 {
			return fmt.Errorf("trusted proxy %q must not be an unspecified or default route", value)
		}
		if releaseMode &&
			addressBits == 32 &&
			prefixLength < minimumTrustedProxyIPv4PrefixRelease &&
			trustedProxyNetworkOverlapsRFC1918(network) {
			return fmt.Errorf(
				"trusted proxy network %q is too broad in release mode; use a specific proxy address or a subnet no broader than /%d",
				value,
				minimumTrustedProxyIPv4PrefixRelease,
			)
		}
	}

	if releaseMode && !hasProxy {
		return fmt.Errorf("trusted proxies are required in release mode")
	}
	return nil
}

func trustedProxyNetworkOverlapsRFC1918(network *net.IPNet) bool {
	if network == nil || network.IP.To4() == nil {
		return false
	}
	for _, privateCIDR := range []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	} {
		_, privateNetwork, err := net.ParseCIDR(privateCIDR)
		if err != nil {
			continue
		}
		if network.Contains(privateNetwork.IP) || privateNetwork.Contains(network.IP) {
			return true
		}
	}
	return false
}

func isDefaultSiteQualityRunnerToken(token string) bool {
	switch strings.ToLower(strings.TrimSpace(token)) {
	case developmentSiteQualityRunnerToken,
		placeholderSiteQualityRunnerToken:
		return true
	default:
		return false
	}
}

// validateConfig 验证配置是否完整
func validateConfig(cfg *Config) error {
	releaseMode := strings.EqualFold(cfg.Server.Mode, "release")
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required. Please set JWT_SECRET environment variable or jwt.secret in config file")
	}

	if releaseMode && len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters in release mode")
	}
	if cfg.Server.ReadTimeout <= 0 ||
		cfg.Server.ReadHeaderTimeout <= 0 ||
		cfg.Server.WriteTimeout <= 0 ||
		cfg.Server.IdleTimeout <= 0 ||
		cfg.Server.MaxHeaderBytes <= 0 {
		return fmt.Errorf("server timeout and header size limits must be positive")
	}
	if err := validateTrustedProxies(cfg.Server.TrustedProxies, releaseMode); err != nil {
		return err
	}
	if err := validateRedisConfig(cfg.Redis); err != nil {
		return err
	}

	if cfg.AntiAbuse.TurnstileRequired {
		if strings.TrimSpace(cfg.AntiAbuse.TurnstileSecretKey) == "" {
			return fmt.Errorf("TURNSTILE_SECRET_KEY is required when Turnstile protection is enabled")
		}
		if cfg.AntiAbuse.VerificationDailyLimit <= 0 ||
			cfg.AntiAbuse.VerificationGlobalLimit <= 0 ||
			cfg.AntiAbuse.VerificationIPWindowSeconds <= 0 ||
			cfg.AntiAbuse.VerificationDestinationWindowSeconds <= 0 ||
			cfg.AntiAbuse.VerificationGlobalWindowSeconds <= 0 ||
			cfg.AntiAbuse.VerificationCircuitSeconds <= 0 {
			return fmt.Errorf("anti-abuse verification limits must be positive")
		}
	}
	if cfg.OrderAbuse.Enabled {
		if cfg.OrderAbuse.OrderCreateWindowSeconds <= 0 {
			return fmt.Errorf("order abuse order create window must be positive")
		}
		if cfg.OrderAbuse.MaxOrderCreationsPerUser <= 0 &&
			cfg.OrderAbuse.MaxOrderCreationsPerSession <= 0 &&
			cfg.OrderAbuse.MaxOrderCreationsPerIP <= 0 {
			return fmt.Errorf("order abuse requires at least one positive identity limit")
		}
	}
	if cfg.OrderNumber.NodeID > 1023 {
		return fmt.Errorf("ORDER_NUMBER_NODE_ID must be between 0 and 1023")
	}
	if releaseMode && len(cfg.OrderNumber.EffectiveSecret(cfg.JWT.Secret)) < 32 {
		return fmt.Errorf("ORDER_NUMBER_SECRET or JWT_SECRET must be at least 32 characters in release mode")
	}
	if releaseMode &&
		cfg.OrderNumber.EffectivePreviousSecret() != "" &&
		len(cfg.OrderNumber.EffectivePreviousSecret()) < 32 {
		return fmt.Errorf("ORDER_NUMBER_PREVIOUS_SECRET must be at least 32 characters in release mode when configured")
	}
	if cfg.PaymentRisk.FailureThreshold != 0 ||
		cfg.PaymentRisk.FailureWindowSeconds != 0 ||
		cfg.PaymentRisk.DelaySeconds != 0 ||
		cfg.PaymentRisk.HighRiskScore != 0 {
		if cfg.PaymentRisk.FailureThreshold <= 0 ||
			cfg.PaymentRisk.FailureWindowSeconds <= 0 ||
			cfg.PaymentRisk.DelaySeconds < 0 ||
			cfg.PaymentRisk.HighRiskScore <= 0 {
			return fmt.Errorf("payment risk configuration is invalid")
		}
	}
	if cfg.PaymentBINRateLimit.Enabled {
		if cfg.PaymentBINRateLimit.WindowSeconds <= 0 ||
			cfg.PaymentBINRateLimit.FailureThreshold <= 0 ||
			cfg.PaymentBINRateLimit.BlockDurationSeconds <= 0 {
			return fmt.Errorf("payment BIN rate limit configuration is invalid")
		}
	}
	if cfg.PaymentGatewayCircuitBreaker.Enabled {
		if cfg.PaymentGatewayCircuitBreaker.WindowSeconds <= 0 ||
			cfg.PaymentGatewayCircuitBreaker.FailureRateThreshold <= 0 ||
			cfg.PaymentGatewayCircuitBreaker.FailureRateThreshold > 1 ||
			cfg.PaymentGatewayCircuitBreaker.MinimumSampleCount <= 0 ||
			cfg.PaymentGatewayCircuitBreaker.OpenDurationSeconds <= 0 ||
			cfg.PaymentGatewayCircuitBreaker.HalfOpenProbeTimeoutSec <
				MinimumPaymentGatewayHalfOpenProbeTimeoutSeconds {
			return fmt.Errorf("payment gateway circuit breaker configuration is invalid")
		}
	}
	if cfg.OutboundHTTPResilience.Enabled {
		if cfg.OutboundHTTPResilience.RetryMaxAttempts <= 0 ||
			cfg.OutboundHTTPResilience.RetryBaseDelayMillis <= 0 ||
			cfg.OutboundHTTPResilience.RetryMaxDelayMillis <= 0 ||
			cfg.OutboundHTTPResilience.RetryBaseDelayMillis > cfg.OutboundHTTPResilience.RetryMaxDelayMillis ||
			cfg.OutboundHTTPResilience.RetryJitterMillis < 0 ||
			cfg.OutboundHTTPResilience.FailureThreshold <= 0 ||
			cfg.OutboundHTTPResilience.FailureWindowSeconds <= 0 ||
			cfg.OutboundHTTPResilience.OpenDurationSeconds <= 0 ||
			cfg.OutboundHTTPResilience.HalfOpenProbeTimeoutSec <
				MinimumOutboundHTTPHalfOpenProbeTimeoutSeconds {
			return fmt.Errorf("outbound HTTP resilience configuration is invalid")
		}
	}
	if cfg.PaymentRiskMonitoring.Enabled {
		if cfg.PaymentRiskMonitoring.WindowDays <= 0 ||
			cfg.PaymentRiskMonitoring.MinimumSuccessfulPayments <= 0 ||
			cfg.PaymentRiskMonitoring.WarningDisputeActivityRate < 0 ||
			cfg.PaymentRiskMonitoring.CriticalDisputeActivityRate < cfg.PaymentRiskMonitoring.WarningDisputeActivityRate ||
			cfg.PaymentRiskMonitoring.WarningEarlyFraudRate < 0 ||
			cfg.PaymentRiskMonitoring.CriticalEarlyFraudRate < cfg.PaymentRiskMonitoring.WarningEarlyFraudRate ||
			cfg.PaymentRiskMonitoring.WarningRefundRate < 0 ||
			cfg.PaymentRiskMonitoring.CriticalRefundRate < cfg.PaymentRiskMonitoring.WarningRefundRate {
			return fmt.Errorf("payment risk monitoring configuration is invalid")
		}
	}
	if cfg.PaymentThreeDS.AdaptiveEnabled {
		if cfg.PaymentThreeDS.LowRiskMaxAmount < 0 ||
			cfg.PaymentThreeDS.TrustedPaidOrders <= 0 ||
			cfg.PaymentThreeDS.VisitorRiskLookback <= 0 ||
			cfg.PaymentThreeDS.StepUpRiskScore <= 0 ||
			cfg.PaymentThreeDS.ChallengeRiskScore <= 0 ||
			cfg.PaymentThreeDS.StepUpRiskScore > cfg.PaymentThreeDS.ChallengeRiskScore {
			return fmt.Errorf("payment 3DS configuration is invalid")
		}
	}
	if cfg.PaymentThreeDS.AVSBillingShippingMismatchHighValueThresholdUSD <= 0 {
		return fmt.Errorf("payment 3DS configuration is invalid")
	}
	if cfg.PaymentProtection.Enabled {
		if cfg.PaymentProtection.MaxControlDurationHours <= 0 ||
			cfg.PaymentProtection.MaxPausePaymentDurationHours <= 0 ||
			cfg.PaymentProtection.MaxGlobalPausePaymentDurationHours <= 0 {
			return fmt.Errorf("payment protection max control durations must be positive when protection is enabled")
		}
		if cfg.PaymentProtection.MaxPausePaymentDurationHours > cfg.PaymentProtection.MaxControlDurationHours ||
			cfg.PaymentProtection.MaxGlobalPausePaymentDurationHours > cfg.PaymentProtection.MaxPausePaymentDurationHours {
			return fmt.Errorf("payment protection pause duration limits must not exceed broader control duration limits")
		}
	}
	if cfg.Worker.VisitorProfileCleanupEnabled && cfg.Worker.VisitorProfileCleanupIntervalSeconds <= 0 {
		return fmt.Errorf("visitor profile cleanup interval must be positive when cleanup is enabled")
	}
	if cfg.Worker.VisitorProfileIPAddressRetentionDays < 0 ||
		(cfg.Worker.VisitorProfileCleanupEnabled && cfg.Worker.VisitorProfileIPAddressRetentionDays <= 0) {
		return fmt.Errorf("visitor profile IP address retention days must be positive when cleanup is enabled")
	}
	if cfg.Worker.BehaviorEventCleanupEnabled && cfg.Worker.BehaviorEventCleanupIntervalSeconds <= 0 {
		return fmt.Errorf("behavior event cleanup interval must be positive when cleanup is enabled")
	}
	if cfg.Worker.OutboxDispatchEnabled {
		if cfg.Worker.OutboxDispatchIntervalSeconds <= 0 ||
			cfg.Worker.OutboxDispatchBatchLimit <= 0 ||
			cfg.Worker.OutboxDispatchLockTimeoutSeconds <= 0 {
			return fmt.Errorf("outbox dispatch configuration is invalid")
		}
	}
	if cfg.CustomerServiceRealtime.Enabled {
		if !cfg.Worker.OutboxDispatchEnabled {
			return fmt.Errorf("customer-service realtime requires outbox dispatch to be enabled")
		}
		if strings.TrimSpace(cfg.CustomerServiceRealtime.Stream) == "" ||
			!hasRedisHashTag(cfg.CustomerServiceRealtime.Stream) ||
			cfg.CustomerServiceRealtime.StreamMaxLen <= 0 ||
			cfg.CustomerServiceRealtime.ReplayLimit <= 0 ||
			cfg.CustomerServiceRealtime.ConsumerBlockSeconds <= 0 ||
			cfg.CustomerServiceRealtime.DedupRetentionSeconds <= 0 {
			return fmt.Errorf("customer-service realtime configuration is invalid")
		}
	}
	if cfg.Worker.PaymentExpirationEnabled {
		if cfg.Worker.PaymentExpirationIntervalSeconds <= 0 ||
			cfg.Worker.PaymentPendingTTLSeconds <= 0 ||
			cfg.Worker.PaymentExpirationBatchLimit <= 0 {
			return fmt.Errorf("payment expiration configuration is invalid")
		}
	}
	if cfg.Worker.PaymentRiskMonitoringEnabled && cfg.Worker.PaymentRiskMonitoringIntervalSeconds <= 0 {
		return fmt.Errorf("payment risk monitoring interval must be positive when monitoring is enabled")
	}
	if cfg.Worker.ExchangeRateSyncEnabled && cfg.Worker.ExchangeRateSyncIntervalSeconds <= 0 {
		return fmt.Errorf("exchange rate sync interval must be positive when sync is enabled")
	}
	if cfg.Worker.SiteQualityEnabled {
		if cfg.Worker.SiteQualityDispatchIntervalSeconds <= 0 ||
			cfg.Worker.SiteQualityBatchLimit <= 0 ||
			cfg.Worker.SiteQualityLeaseTimeoutSeconds <= 0 ||
			cfg.Worker.SiteQualitySampleCount <= 0 ||
			cfg.Worker.SiteQualityConfirmations <= 0 ||
			cfg.Worker.SiteQualityConfirmations > cfg.Worker.SiteQualitySampleCount ||
			cfg.Worker.SiteQualityCleanEvaluations <= 0 ||
			cfg.Worker.SiteQualityProviderConcurrency <= 0 ||
			cfg.Worker.SiteQualityProviderSpacingSeconds < 0 {
			return fmt.Errorf("site quality job worker configuration is invalid")
		}
		if strings.TrimSpace(cfg.SiteQuality.RunnerURL) == "" {
			return fmt.Errorf("SITE_QUALITY_RUNNER_URL is required when the Site Quality job worker is enabled")
		}
		runnerURL, err := url.ParseRequestURI(strings.TrimSpace(cfg.SiteQuality.RunnerURL))
		if err != nil || runnerURL.Scheme == "" || runnerURL.Host == "" ||
			(runnerURL.Scheme != "http" && runnerURL.Scheme != "https") {
			return fmt.Errorf("SITE_QUALITY_RUNNER_URL must be an absolute HTTP(S) URL")
		}
		if strings.TrimSpace(cfg.SiteQuality.RunnerToken) == "" {
			return fmt.Errorf("SITE_QUALITY_RUNNER_TOKEN is required when the Site Quality job worker is enabled")
		}
		if len(strings.TrimSpace(cfg.SiteQuality.RunnerToken)) < 32 {
			return fmt.Errorf("SITE_QUALITY_RUNNER_TOKEN must be at least 32 characters when the Site Quality job worker is enabled")
		}
		if releaseMode && isDefaultSiteQualityRunnerToken(cfg.SiteQuality.RunnerToken) {
			return fmt.Errorf("SITE_QUALITY_RUNNER_TOKEN must not use a development/default value in release mode")
		}
	} else if releaseMode {
		if strings.TrimSpace(cfg.SiteQuality.RunnerURL) != "" && strings.TrimSpace(cfg.SiteQuality.RunnerToken) == "" {
			return fmt.Errorf("SITE_QUALITY_RUNNER_TOKEN is required in release mode when the Site Quality runner is configured")
		}
		if isDefaultSiteQualityRunnerToken(cfg.SiteQuality.RunnerToken) {
			return fmt.Errorf("SITE_QUALITY_RUNNER_TOKEN must not use a development/default value in release mode")
		}
	}
	if cfg.Worker.MediaDerivativeRebuildEnabled {
		if cfg.Worker.MediaDerivativeRebuildIntervalSeconds <= 0 ||
			cfg.Worker.MediaDerivativeRebuildBatchLimit <= 0 ||
			cfg.Worker.MediaDerivativeRebuildBatchLimit > 100 {
			return fmt.Errorf("media derivative rebuild worker configuration is invalid")
		}
	}
	if cfg.Worker.ShowcaseCleanupEnabled {
		if cfg.Worker.ShowcaseCleanupIntervalSeconds <= 0 ||
			cfg.Worker.ShowcasePendingTTLSeconds <= 0 ||
			cfg.Worker.ShowcaseCleanupBatchLimit <= 0 {
			return fmt.Errorf("showcase cleanup configuration is invalid")
		}
	}
	if cfg.Worker.HotDataArchiveEnabled {
		if cfg.Worker.HotDataArchiveIntervalSeconds <= 0 ||
			cfg.Worker.HotDataArchiveBatchLimit <= 0 {
			return fmt.Errorf("hot data archive configuration is invalid")
		}
	}
	if cfg.BehaviorEvents.LowIntentRetentionDays != 0 ||
		cfg.BehaviorEvents.StandardIntentRetentionDays != 0 ||
		cfg.BehaviorEvents.HighIntentRetentionDays != 0 ||
		cfg.BehaviorEvents.CleanupBatchLimit != 0 {
		if cfg.BehaviorEvents.LowIntentRetentionDays <= 0 ||
			cfg.BehaviorEvents.StandardIntentRetentionDays <= 0 ||
			cfg.BehaviorEvents.HighIntentRetentionDays <= 0 ||
			cfg.BehaviorEvents.CleanupBatchLimit <= 0 {
			return fmt.Errorf("behavior events retention configuration is invalid")
		}
		if cfg.BehaviorEvents.LowIntentRetentionDays > cfg.BehaviorEvents.StandardIntentRetentionDays ||
			cfg.BehaviorEvents.StandardIntentRetentionDays > cfg.BehaviorEvents.HighIntentRetentionDays {
			return fmt.Errorf("behavior events retention days must be ordered low <= standard <= high")
		}
	}
	if cfg.VisitorRisk.Enabled {
		if cfg.VisitorRisk.FlushIntervalSeconds <= 0 ||
			cfg.VisitorRisk.MaxPendingFacts <= 0 ||
			cfg.VisitorRisk.SamplePathLimit <= 0 ||
			cfg.VisitorRisk.RetentionDays <= 0 {
			return fmt.Errorf("visitor risk configuration is invalid")
		}
	}
	if cfg.RequestSigning.Enabled && len(strings.TrimSpace(cfg.RequestSigning.Key)) < 32 {
		return fmt.Errorf("REQUEST_SIGNING_KEY must be at least 32 characters when request signing is enabled")
	}
	if cfg.RequestSigning.Enabled && cfg.RequestSigning.MaxSkewSeconds <= 0 {
		return fmt.Errorf("request signing max skew must be positive")
	}
	if cfg.QuickBuyRateLimit.Enabled {
		if cfg.QuickBuyRateLimit.IPRequestsPerMinute <= 0 ||
			cfg.QuickBuyRateLimit.IPBurst <= 0 ||
			cfg.QuickBuyRateLimit.SessionRequestsPerMinute <= 0 ||
			cfg.QuickBuyRateLimit.SessionBurst <= 0 {
			return fmt.Errorf("quick buy rate limit configuration is invalid")
		}
	}
	if cfg.FeedbackRateLimit.Enabled {
		if cfg.FeedbackRateLimit.ReadIPRequestsPerMinute <= 0 ||
			cfg.FeedbackRateLimit.ReadIPBurst <= 0 ||
			cfg.FeedbackRateLimit.WriteIPRequestsPerMinute <= 0 ||
			cfg.FeedbackRateLimit.WriteIPBurst <= 0 ||
			cfg.FeedbackRateLimit.WriteUserRequestsPerMinute <= 0 ||
			cfg.FeedbackRateLimit.WriteUserBurst <= 0 {
			return fmt.Errorf("feedback rate limit configuration is invalid")
		}
	}
	if cfg.Cache.ProductLockTTL <= 0 {
		return fmt.Errorf("CACHE_PRODUCT_LOCK_TTL must be positive")
	}
	if cfg.MediaUpload.AccountStorageQuotaBytes <= 0 {
		return fmt.Errorf("media upload account storage quota must be positive")
	}
	if cfg.ShowcaseUploadProtection.Enabled {
		if cfg.ShowcaseUploadProtection.WindowSeconds <= 0 ||
			cfg.ShowcaseUploadProtection.MaxUploadsPerUser <= 0 ||
			cfg.ShowcaseUploadProtection.MaxUploadsPerIP <= 0 ||
			cfg.ShowcaseUploadProtection.MaxUploadsPerIPPrefix <= 0 ||
			cfg.ShowcaseUploadProtection.DailyMaxUploadsPerUser <= 0 ||
			cfg.ShowcaseUploadProtection.DailyMaxUploadsPerIP <= 0 ||
			cfg.ShowcaseUploadProtection.DailyMaxBytesPerUser <= 0 ||
			cfg.ShowcaseUploadProtection.DailyMaxBytesPerIP <= 0 ||
			cfg.ShowcaseUploadProtection.MaxPendingSubmissionsPerUser <= 0 ||
			cfg.ShowcaseUploadProtection.FailureWindowSeconds <= 0 ||
			cfg.ShowcaseUploadProtection.MaxFailuresPerUser <= 0 ||
			cfg.ShowcaseUploadProtection.MaxFailuresPerIP <= 0 ||
			cfg.ShowcaseUploadProtection.BlockDurationSeconds <= 0 {
			return fmt.Errorf("showcase upload protection configuration is invalid")
		}
	}
	googleMerchantOAuthFields := []string{
		strings.TrimSpace(cfg.GoogleMerchant.ClientID),
		strings.TrimSpace(cfg.GoogleMerchant.ClientSecret),
		strings.TrimSpace(cfg.GoogleMerchant.RedirectURL),
	}
	googleMerchantConfiguredFields := 0
	for _, field := range googleMerchantOAuthFields {
		if field != "" {
			googleMerchantConfiguredFields++
		}
	}
	if googleMerchantConfiguredFields != 0 && googleMerchantConfiguredFields != len(googleMerchantOAuthFields) {
		return fmt.Errorf("Google Merchant OAuth requires client id, client secret, and redirect URL together")
	}
	if googleMerchantConfiguredFields == len(googleMerchantOAuthFields) &&
		len(strings.TrimSpace(cfg.GoogleMerchant.TokenEncryptionKey)) < 32 {
		return fmt.Errorf("GOOGLE_MERCHANT_TOKEN_ENCRYPTION_KEY must be at least 32 characters when Google Merchant OAuth is configured")
	}
	if googleMerchantConfiguredFields != 0 && cfg.GoogleMerchant.StateTTLSeconds <= 0 {
		return fmt.Errorf("Google Merchant OAuth state TTL must be positive")
	}
	if err := validateGoogleIndexingConfig(cfg.GoogleIndexing); err != nil {
		return err
	}
	if err := validateSocialOAuthConfig(cfg.SocialOAuth); err != nil {
		return err
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if !isValidCookieSecureMode(cfg.Cookie.Secure) {
		return fmt.Errorf("cookie.secure must be auto, always/true, or never/false")
	}

	if !isValidCookieSameSite(cfg.Cookie.SameSite) {
		return fmt.Errorf("cookie.same_site must be one of lax, strict, or none")
	}

	if cfg.Cookie.SameSiteMode() == http.SameSiteNoneMode && !cfg.Cookie.SecureEnabled(cfg.Server) {
		return fmt.Errorf("cookie.same_site=none requires secure cookies")
	}

	return nil
}

func validateGoogleIndexingConfig(cfg GoogleIndexingConfig) error {
	jsonConfigured := strings.TrimSpace(cfg.ServiceAccountJSON) != ""
	fileConfigured := strings.TrimSpace(cfg.ServiceAccountFile) != ""
	if jsonConfigured && fileConfigured {
		return fmt.Errorf("Google Indexing requires either service account JSON or service account file, not both")
	}
	if cfg.RequestTimeoutSeconds < 0 {
		return fmt.Errorf("Google Indexing request timeout must not be negative")
	}
	if cfg.Enabled && !jsonConfigured && !fileConfigured {
		return fmt.Errorf("Google Indexing requires GOOGLE_INDEXING_SERVICE_ACCOUNT_JSON or GOOGLE_INDEXING_SERVICE_ACCOUNT_FILE when enabled")
	}
	return nil
}

func validateSocialOAuthConfig(cfg SocialOAuthConfig) error {
	providers := map[string]SocialOAuthProviderConfig{
		"Meta":    cfg.Meta,
		"X":       cfg.X,
		"YouTube": cfg.YouTube,
		"Reddit":  cfg.Reddit,
	}
	configured := false
	for name, provider := range providers {
		fields := []string{
			strings.TrimSpace(provider.ClientID),
			strings.TrimSpace(provider.ClientSecret),
			strings.TrimSpace(provider.RedirectURL),
		}
		configuredFields := 0
		for _, field := range fields {
			if field != "" {
				configuredFields++
			}
		}
		if configuredFields != 0 && configuredFields != len(fields) {
			return fmt.Errorf("Social OAuth %s requires client id, client secret, and redirect URL together", name)
		}
		configured = configured || configuredFields == len(fields)
	}
	if !configured {
		return nil
	}
	if len(strings.TrimSpace(cfg.TokenEncryptionKey)) < 32 {
		return fmt.Errorf("SOCIAL_OAUTH_TOKEN_ENCRYPTION_KEY must be at least 32 characters when Social OAuth is configured")
	}
	if cfg.StateTTLSeconds <= 0 {
		return fmt.Errorf("Social OAuth state TTL must be positive")
	}
	if strings.TrimSpace(cfg.PostConnectURL) == "" {
		return fmt.Errorf("SOCIAL_OAUTH_POST_CONNECT_URL is required when Social OAuth is configured")
	}
	return nil
}

func hasRedisHashTag(value string) bool {
	start := strings.IndexByte(value, '{')
	if start < 0 {
		return false
	}
	endOffset := strings.IndexByte(value[start+1:], '}')
	return endOffset > 0
}

func isValidCookieSecureMode(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto", "false", "no", "never", "off", "0", "true", "yes", "always", "on", "1":
		return true
	default:
		return false
	}
}

func isValidCookieSameSite(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "lax", "strict", "none":
		return true
	default:
		return false
	}
}
