package config

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"tanzanite/internal/pkg/locales"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig         `mapstructure:"server"`
	Database       DatabaseConfig       `mapstructure:"database"`
	Redis          RedisConfig          `mapstructure:"redis"`
	JWT            JWTConfig            `mapstructure:"jwt"`
	OAuth          OAuthConfig          `mapstructure:"oauth"`
	I18n           I18nConfig           `mapstructure:"i18n"`
	CORS           CORSConfig           `mapstructure:"cors"`
	Cookie         CookieConfig         `mapstructure:"cookie"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Log            LogConfig            `mapstructure:"log"`
	Worker         WorkerConfig         `mapstructure:"worker"`
	AntiAbuse      AntiAbuseConfig      `mapstructure:"anti_abuse"`
	PaymentRisk    PaymentRiskConfig    `mapstructure:"payment_risk"`
	RequestSigning RequestSigningConfig `mapstructure:"request_signing"`
}

type ServerConfig struct {
	Port           string   `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"`
	BaseURL        string   `mapstructure:"base_url"`
	ReadTimeout    int      `mapstructure:"read_timeout"`
	WriteTimeout   int      `mapstructure:"write_timeout"`
	TrustedProxies []string `mapstructure:"trusted_proxies"`
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
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	ExpireHours        int    `mapstructure:"expire_hours"`
	RefreshExpireHours int    `mapstructure:"refresh_expire_hours"`
}

type OAuthConfig struct {
	GoogleClientID string `mapstructure:"google_client_id"`
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
	DefaultTTL  int `mapstructure:"default_ttl"`
	PostTTL     int `mapstructure:"post_ttl"`
	ProductTTL  int `mapstructure:"product_ttl"`
	SettingsTTL int `mapstructure:"settings_ttl"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type WorkerConfig struct {
	Enabled                        bool `mapstructure:"enabled"`
	TrackingPollingEnabled         bool `mapstructure:"tracking_polling_enabled"`
	TrackingPollingIntervalSeconds int  `mapstructure:"tracking_polling_interval_seconds"`
	TrackingPollingBatchLimit      int  `mapstructure:"tracking_polling_batch_limit"`
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

type PaymentRiskConfig struct {
	FailureWindowSeconds int `mapstructure:"failure_window_seconds"`
	FailureThreshold     int `mapstructure:"failure_threshold"`
	DelaySeconds         int `mapstructure:"delay_seconds"`
	HighRiskScore        int `mapstructure:"high_risk_score"`
}

type RequestSigningConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	Key            string   `mapstructure:"key"`
	MaxSkewSeconds int      `mapstructure:"max_skew_seconds"`
	RequiredPaths  []string `mapstructure:"required_paths"`
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
	viper.SetDefault("server.write_timeout", 60)
	viper.SetDefault("server.trusted_proxies", []string{})

	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 9400)
	viper.SetDefault("database.username", "tanzanite")
	viper.SetDefault("database.password", "tanzanite_password")
	viper.SetDefault("database.database", "tanzanite")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.log_level", "silent")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 9500)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.refresh_expire_hours", 168)

	viper.SetDefault("oauth.google_client_id", "")

	viper.SetDefault("i18n.default_locale", "en")
	viper.SetDefault("i18n.supported_locales", locales.SupportedLocaleCodes())

	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:9100", "http://localhost:9300"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{
		"Origin",
		"Content-Type",
		"Accept-Language",
		"X-CSRF-Token",
		"X-Request-Timestamp",
		"X-Request-Nonce",
		"X-Request-Signature",
	})
	viper.SetDefault("cors.expose_headers", []string{"Content-Length"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 43200)

	viper.SetDefault("cookie.secure", "auto")
	viper.SetDefault("cookie.same_site", "lax")
	viper.SetDefault("cookie.domain", "")

	viper.SetDefault("cache.default_ttl", 3600)
	viper.SetDefault("cache.post_ttl", 3600)
	viper.SetDefault("cache.product_ttl", 1800)
	viper.SetDefault("cache.settings_ttl", 7200)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	viper.SetDefault("worker.enabled", false)
	viper.SetDefault("worker.tracking_polling_enabled", false)
	viper.SetDefault("worker.tracking_polling_interval_seconds", 300)
	viper.SetDefault("worker.tracking_polling_batch_limit", 20)

	viper.SetDefault("anti_abuse.turnstile_required", false)
	viper.SetDefault("anti_abuse.turnstile_secret_key", "")
	viper.SetDefault("anti_abuse.verification_ip_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_destination_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_daily_limit", 8)
	viper.SetDefault("anti_abuse.verification_global_window_seconds", 60)
	viper.SetDefault("anti_abuse.verification_global_limit", 100)
	viper.SetDefault("anti_abuse.verification_circuit_seconds", 300)

	viper.SetDefault("payment_risk.failure_window_seconds", 600)
	viper.SetDefault("payment_risk.failure_threshold", 2)
	viper.SetDefault("payment_risk.delay_seconds", 2)
	viper.SetDefault("payment_risk.high_risk_score", 60)

	viper.SetDefault("request_signing.enabled", false)
	viper.SetDefault("request_signing.key", "")
	viper.SetDefault("request_signing.max_skew_seconds", 30)
	viper.SetDefault("request_signing.required_paths", []string{})
}

func bindEnvironment() {
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.mode", "SERVER_MODE")
	_ = viper.BindEnv("server.base_url", "SERVER_BASE_URL")
	_ = viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	_ = viper.BindEnv("database.driver", "DB_DRIVER", "DATABASE_DRIVER")
	_ = viper.BindEnv("database.host", "DB_HOST", "DATABASE_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT", "DATABASE_PORT")
	_ = viper.BindEnv("database.username", "DB_USERNAME", "DATABASE_USERNAME")
	_ = viper.BindEnv("database.password", "DB_PASSWORD", "DATABASE_PASSWORD")
	_ = viper.BindEnv("database.database", "DB_NAME", "DATABASE_NAME")
	_ = viper.BindEnv("database.auto_migrate", "DB_AUTO_MIGRATE", "DATABASE_AUTO_MIGRATE")
	_ = viper.BindEnv("database.log_level", "DB_LOG_LEVEL", "DATABASE_LOG_LEVEL")

	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = viper.BindEnv("redis.db", "REDIS_DB")

	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("jwt.expire_hours", "JWT_EXPIRE_HOURS")
	_ = viper.BindEnv("jwt.refresh_expire_hours", "JWT_REFRESH_EXPIRE_HOURS")

	_ = viper.BindEnv("oauth.google_client_id", "GOOGLE_CLIENT_ID", "GOOGLE_OAUTH_CLIENT_ID", "NUXT_PUBLIC_GOOGLE_CLIENT_ID")

	_ = viper.BindEnv("cookie.secure", "COOKIE_SECURE")
	_ = viper.BindEnv("cookie.same_site", "COOKIE_SAME_SITE")
	_ = viper.BindEnv("cookie.domain", "COOKIE_DOMAIN")

	_ = viper.BindEnv("log.level", "LOG_LEVEL")
	_ = viper.BindEnv("log.format", "LOG_FORMAT")
	_ = viper.BindEnv("log.output", "LOG_OUTPUT")

	_ = viper.BindEnv("worker.enabled", "WORKER_ENABLED", "ASYNQ_WORKER_ENABLED")
	_ = viper.BindEnv("worker.tracking_polling_enabled", "WORKER_TRACKING_POLLING_ENABLED", "TRACKING_POLLING_ENABLED")
	_ = viper.BindEnv("worker.tracking_polling_interval_seconds", "WORKER_TRACKING_POLLING_INTERVAL_SECONDS", "TRACKING_POLLING_INTERVAL_SECONDS")
	_ = viper.BindEnv("worker.tracking_polling_batch_limit", "WORKER_TRACKING_POLLING_BATCH_LIMIT", "TRACKING_POLLING_BATCH_LIMIT")

	_ = viper.BindEnv("anti_abuse.turnstile_required", "TURNSTILE_REQUIRED")
	_ = viper.BindEnv("anti_abuse.turnstile_secret_key", "TURNSTILE_SECRET_KEY")
	_ = viper.BindEnv("anti_abuse.verification_ip_window_seconds", "VERIFICATION_IP_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_destination_window_seconds", "VERIFICATION_DESTINATION_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_daily_limit", "VERIFICATION_DAILY_LIMIT")
	_ = viper.BindEnv("anti_abuse.verification_global_window_seconds", "VERIFICATION_GLOBAL_WINDOW_SECONDS")
	_ = viper.BindEnv("anti_abuse.verification_global_limit", "VERIFICATION_GLOBAL_LIMIT")
	_ = viper.BindEnv("anti_abuse.verification_circuit_seconds", "VERIFICATION_CIRCUIT_SECONDS")

	_ = viper.BindEnv("payment_risk.failure_window_seconds", "PAYMENT_RISK_FAILURE_WINDOW_SECONDS")
	_ = viper.BindEnv("payment_risk.failure_threshold", "PAYMENT_RISK_FAILURE_THRESHOLD")
	_ = viper.BindEnv("payment_risk.delay_seconds", "PAYMENT_RISK_DELAY_SECONDS")
	_ = viper.BindEnv("payment_risk.high_risk_score", "PAYMENT_RISK_HIGH_RISK_SCORE")

	_ = viper.BindEnv("request_signing.enabled", "REQUEST_SIGNING_ENABLED")
	_ = viper.BindEnv("request_signing.key", "REQUEST_SIGNING_KEY")
	_ = viper.BindEnv("request_signing.max_skew_seconds", "REQUEST_SIGNING_MAX_SKEW_SECONDS")
	_ = viper.BindEnv("request_signing.required_paths", "REQUEST_SIGNING_REQUIRED_PATHS")
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

// validateConfig 验证配置是否完整
func validateConfig(cfg *Config) error {
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required. Please set JWT_SECRET environment variable or jwt.secret in config file")
	}

	if strings.EqualFold(cfg.Server.Mode, "release") && len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters in release mode")
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
	if cfg.RequestSigning.Enabled && len(strings.TrimSpace(cfg.RequestSigning.Key)) < 32 {
		return fmt.Errorf("REQUEST_SIGNING_KEY must be at least 32 characters when request signing is enabled")
	}
	if cfg.RequestSigning.Enabled && cfg.RequestSigning.MaxSkewSeconds <= 0 {
		return fmt.Errorf("request signing max skew must be positive")
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
