package resilience

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrCircuitStoreUnavailable = errors.New("outbound circuit breaker store is unavailable")

const circuitStoreOperationTimeout = 2 * time.Second

// DistributedCircuitBreakerConfig controls the Redis-backed state shared by
// all application replicas. ProbeTimeout must be at least as long as the
// largest outbound HTTP timeout protected by this breaker.
type DistributedCircuitBreakerConfig struct {
	Enabled          bool
	FailureThreshold int
	FailureWindow    time.Duration
	OpenDuration     time.Duration
	ProbeTimeout     time.Duration
	KeyPrefix        string
}

func DefaultDistributedCircuitBreakerConfig() DistributedCircuitBreakerConfig {
	return DistributedCircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 4,
		FailureWindow:    60 * time.Second,
		OpenDuration:     30 * time.Second,
		ProbeTimeout:     30 * time.Second,
		KeyPrefix:        "commerce-platform:outbound-circuit",
	}
}

type DistributedCircuitBreaker struct {
	client redis.UniversalClient
	config DistributedCircuitBreakerConfig
}

// ProbeTimeout exposes the maximum lifetime of a half-open probe. The HTTP
// client uses it to prevent a slow probe from outliving its Redis lease.
func (b *DistributedCircuitBreaker) ProbeTimeout() time.Duration {
	if b == nil {
		return 0
	}
	return b.config.ProbeTimeout
}

func NewDistributedCircuitBreaker(
	client redis.UniversalClient,
	config DistributedCircuitBreakerConfig,
) *DistributedCircuitBreaker {
	return &DistributedCircuitBreaker{
		client: client,
		config: normalizeDistributedCircuitBreakerConfig(config),
	}
}

func (b *DistributedCircuitBreaker) Acquire(ctx context.Context, key string) (CircuitPermit, error) {
	if b == nil {
		return noopCircuitPermit{}, nil
	}
	if !b.config.Enabled {
		return noopCircuitPermit{}, nil
	}
	if b.client == nil {
		return nil, ErrCircuitStoreUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	key = normalizeKey(key)
	if key == "" {
		return noopCircuitPermit{}, nil
	}

	keys := b.keys(key)
	token := uuid.NewString()
	values, err := distributedCircuitAcquireScript.Run(
		ctx,
		b.client,
		[]string{keys.state, keys.probe},
		token,
		b.config.ProbeTimeout.Milliseconds(),
	).Int64Slice()
	if err != nil {
		return nil, fmt.Errorf("%w: acquire %s: %v", ErrCircuitStoreUnavailable, key, err)
	}
	if len(values) != 3 {
		return nil, fmt.Errorf("%w: acquire %s returned invalid state", ErrCircuitStoreUnavailable, key)
	}
	if values[0] != 1 {
		retryAfter := time.Duration(values[1]) * time.Millisecond
		if retryAfter <= 0 {
			retryAfter = time.Millisecond
		}
		return nil, circuitOpenError{key: key, retryAfter: retryAfter}
	}

	return distributedCircuitPermit{
		breaker: b,
		key:     key,
		token:   token,
		probe:   values[2] == 1,
	}, nil
}

type circuitOpenError struct {
	key        string
	retryAfter time.Duration
}

func (e circuitOpenError) Error() string {
	if e.retryAfter <= 0 {
		return fmt.Sprintf("%s: %s", ErrCircuitOpen, e.key)
	}
	return fmt.Sprintf("%s: %s (retry after %s)", ErrCircuitOpen, e.key, e.retryAfter.Round(time.Millisecond))
}

func (e circuitOpenError) Is(target error) bool {
	return target == ErrCircuitOpen
}

func RetryAfter(err error) time.Duration {
	var openErr circuitOpenError
	if errors.As(err, &openErr) {
		return openErr.retryAfter
	}
	return 0
}

type distributedCircuitPermit struct {
	breaker *DistributedCircuitBreaker
	key     string
	token   string
	probe   bool
}

func (p distributedCircuitPermit) IsProbe() bool {
	return p.probe
}

func (p distributedCircuitPermit) RecordSuccess(ctx context.Context) {
	if p.breaker == nil {
		return
	}
	ctx, cancel := circuitStoreOperationContext()
	defer cancel()
	keys := p.breaker.keys(p.key)
	_ = distributedCircuitSuccessScript.Run(
		ctx,
		p.breaker.client,
		[]string{keys.state, keys.probe, keys.failures},
		p.token,
		boolToRedisArgument(p.probe),
	).Err()
}

func (p distributedCircuitPermit) RecordFailure(ctx context.Context) {
	if p.breaker == nil {
		return
	}
	ctx, cancel := circuitStoreOperationContext()
	defer cancel()
	keys := p.breaker.keys(p.key)
	_ = distributedCircuitFailureScript.Run(
		ctx,
		p.breaker.client,
		[]string{keys.state, keys.probe, keys.failures},
		p.token,
		boolToRedisArgument(p.probe),
		p.breaker.config.FailureThreshold,
		p.breaker.config.FailureWindow.Milliseconds(),
		p.breaker.config.OpenDuration.Milliseconds(),
		p.breaker.config.ProbeTimeout.Milliseconds(),
	).Err()
}

func (p distributedCircuitPermit) Release(ctx context.Context) {
	if p.breaker == nil || !p.probe {
		return
	}
	ctx, cancel := circuitStoreOperationContext()
	defer cancel()
	keys := p.breaker.keys(p.key)
	_ = distributedCircuitReleaseProbeScript.Run(ctx, p.breaker.client, []string{keys.probe}, p.token).Err()
}

func circuitStoreOperationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), circuitStoreOperationTimeout)
}

type distributedCircuitKeys struct {
	state    string
	probe    string
	failures string
}

func (b *DistributedCircuitBreaker) keys(key string) distributedCircuitKeys {
	prefix := strings.TrimSpace(b.config.KeyPrefix)
	if prefix == "" {
		prefix = DefaultDistributedCircuitBreakerConfig().KeyPrefix
	}
	sum := sha256.Sum256([]byte(normalizeKey(key)))
	tag := hex.EncodeToString(sum[:16])
	base := prefix + ":{" + tag + "}"
	return distributedCircuitKeys{
		state:    base + ":state",
		probe:    base + ":probe",
		failures: base + ":failures",
	}
}

func normalizeDistributedCircuitBreakerConfig(
	config DistributedCircuitBreakerConfig,
) DistributedCircuitBreakerConfig {
	defaults := DefaultDistributedCircuitBreakerConfig()
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = defaults.FailureThreshold
	}
	if config.FailureWindow <= 0 {
		config.FailureWindow = defaults.FailureWindow
	}
	if config.OpenDuration <= 0 {
		config.OpenDuration = defaults.OpenDuration
	}
	if config.ProbeTimeout <= 0 {
		config.ProbeTimeout = defaults.ProbeTimeout
	}
	if strings.TrimSpace(config.KeyPrefix) == "" {
		config.KeyPrefix = defaults.KeyPrefix
	}
	return config
}

func boolToRedisArgument(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

var distributedCircuitAcquireScript = redis.NewScript(`
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)

local state = redis.call("GET", KEYS[1])
if not state then
	return {1, 0, 0}
end

local openUntil = tonumber(state)
if not openUntil then
	redis.call("DEL", KEYS[1])
	return {1, 0, 0}
end
if openUntil > now then
	return {0, openUntil - now, 0}
end

local probeTTL = tonumber(ARGV[2])
if redis.call("SET", KEYS[2], ARGV[1], "PX", probeTTL, "NX") then
	return {1, 0, 1}
end

local retryAfter = redis.call("PTTL", KEYS[2])
if retryAfter <= 0 then
	retryAfter = probeTTL
end
return {0, retryAfter, 1}
`)

var distributedCircuitSuccessScript = redis.NewScript(`
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)

local token = ARGV[1]
local probe = ARGV[2]

if probe == "1" then
	if redis.call("GET", KEYS[2]) == token then
		redis.call("DEL", KEYS[1])
		redis.call("DEL", KEYS[2])
		redis.call("DEL", KEYS[3])
		return 1
	end
	return 0
end

local state = redis.call("GET", KEYS[1])
if not state or tonumber(state) <= now then
	redis.call("DEL", KEYS[3])
end
return 1
`)

var distributedCircuitFailureScript = redis.NewScript(`
local redisTime = redis.call("TIME")
local now = (tonumber(redisTime[1]) * 1000) + math.floor(tonumber(redisTime[2]) / 1000)

local token = ARGV[1]
local probe = ARGV[2]
local threshold = tonumber(ARGV[3])
local window = tonumber(ARGV[4])
local openDuration = tonumber(ARGV[5])
local probeTimeout = tonumber(ARGV[6])
local stateTTL = openDuration + probeTimeout

if probe == "1" then
	if redis.call("GET", KEYS[2]) == token then
		redis.call("SET", KEYS[1], now + openDuration, "PX", stateTTL)
		redis.call("DEL", KEYS[2])
		redis.call("DEL", KEYS[3])
		return 1
	end
	return 0
end

local state = redis.call("GET", KEYS[1])
if state and tonumber(state) > now then
	return 0
end

redis.call("ZREMRANGEBYSCORE", KEYS[3], 0, now - window)
redis.call("ZADD", KEYS[3], now, token)
redis.call("PEXPIRE", KEYS[3], window + 1000)
local failures = redis.call("ZCARD", KEYS[3])
if failures >= threshold then
	redis.call("SET", KEYS[1], now + openDuration, "PX", stateTTL)
	redis.call("DEL", KEYS[3])
	return 1
end
return 0
`)

var distributedCircuitReleaseProbeScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)
