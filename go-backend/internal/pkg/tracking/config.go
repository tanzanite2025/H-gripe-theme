package tracking

import (
	"time"

	"commerce-platform/internal/pkg/resilience"
)

// Config 追踪服务配置
type Config struct {
	Provider string // 17track, aftership, etc.
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
	Retry    resilience.HTTPRetryPolicy
	Breaker  resilience.CircuitController
}
