package tracking

import (
	"os"
	"time"
)

// Config 追踪服务配置
type Config struct {
	Provider string // 17track, aftership, etc.
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	return &Config{
		Provider: getEnv("TRACKING_PROVIDER", "17track"),
		APIKey:   os.Getenv("TRACKING_API_KEY"),
		BaseURL:  getEnv("TRACKING_BASE_URL", "https://api.17track.net/track/v2"),
		Timeout:  30 * time.Second,
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
