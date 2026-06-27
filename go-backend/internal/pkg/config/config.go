package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	OAuth    OAuthConfig    `mapstructure:"oauth"`
	I18n     I18nConfig     `mapstructure:"i18n"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Log      LogConfig      `mapstructure:"log"`
	Worker   WorkerConfig   `mapstructure:"worker"`
}

type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	BaseURL      string `mapstructure:"base_url"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
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
	Enabled bool `mapstructure:"enabled"`
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
	viper.SetDefault("server.port", ":9000")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.base_url", "http://localhost:9000")
	viper.SetDefault("server.read_timeout", 60)
	viper.SetDefault("server.write_timeout", 60)

	viper.SetDefault("database.driver", "postgres")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "tanzanite")
	viper.SetDefault("database.password", "tanzanite_password")
	viper.SetDefault("database.database", "tanzanite")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)
	viper.SetDefault("database.auto_migrate", true)
	viper.SetDefault("database.log_level", "silent")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("jwt.expire_hours", 24)
	viper.SetDefault("jwt.refresh_expire_hours", 168)

	viper.SetDefault("oauth.google_client_id", "")

	viper.SetDefault("i18n.default_locale", "en")
	viper.SetDefault("i18n.supported_locales", []string{"en", "zh", "fr", "de", "es", "ja", "ko", "it", "pt", "ru", "ar", "fi", "da", "th"})

	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:5173"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Authorization", "Accept-Language"})
	viper.SetDefault("cors.expose_headers", []string{"Content-Length"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 43200)

	viper.SetDefault("cache.default_ttl", 3600)
	viper.SetDefault("cache.post_ttl", 3600)
	viper.SetDefault("cache.product_ttl", 1800)
	viper.SetDefault("cache.settings_ttl", 7200)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	viper.SetDefault("worker.enabled", false)
}

func bindEnvironment() {
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.mode", "SERVER_MODE")
	_ = viper.BindEnv("server.base_url", "SERVER_BASE_URL")

	_ = viper.BindEnv("database.driver", "DB_DRIVER", "DATABASE_DRIVER")
	_ = viper.BindEnv("database.host", "DB_HOST", "DATABASE_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT", "DATABASE_PORT")
	_ = viper.BindEnv("database.username", "DB_USER", "DB_USERNAME", "DATABASE_USERNAME")
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

	_ = viper.BindEnv("log.level", "LOG_LEVEL")
	_ = viper.BindEnv("log.format", "LOG_FORMAT")
	_ = viper.BindEnv("log.output", "LOG_OUTPUT")

	_ = viper.BindEnv("worker.enabled", "WORKER_ENABLED", "ASYNQ_WORKER_ENABLED")
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

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}
