package orderabuse

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/redis/go-redis/v9"
)

var ErrServiceUnavailable = errors.New("order abuse service is unavailable")

const (
	ActionAllow = "allow"
	ActionBlock = "block"
)

type Identity struct {
	UserID    uint
	SessionID string
	IPAddress string
}

type Decision struct {
	Allowed    bool
	Action     string
	Dimension  string
	Count      int64
	Limit      int
	RetryAfter time.Duration
}

type counterStore interface {
	Increment(context.Context, string, time.Duration) (int64, error)
}

type redisCounterStore struct {
	client *redis.Client
}

const incrementWithTTLScript = `
local count = redis.call("INCR", KEYS[1])
if count == 1 then
	local ttl = tonumber(ARGV[1])
	if ttl > 0 then
		redis.call("EXPIRE", KEYS[1], ttl)
	end
end
return count
`

func (s redisCounterStore) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	return s.client.Eval(ctx, incrementWithTTLScript, []string{key}, int64(ttl/time.Second)).Int64()
}

type Service struct {
	store counterStore
	cfg   config.OrderAbuseConfig
}

func New(redisClient *redis.Client, cfg config.OrderAbuseConfig) *Service {
	if redisClient == nil {
		return nil
	}
	return &Service{
		store: redisCounterStore{client: redisClient},
		cfg:   cfg,
	}
}

func (s *Service) Evaluate(ctx context.Context, identity Identity) (Decision, error) {
	decision := Decision{Allowed: true, Action: ActionAllow}
	if s == nil || !s.cfg.Enabled || s.store == nil {
		return decision, nil
	}

	ttl := time.Duration(s.cfg.OrderCreateWindowSeconds) * time.Second
	for _, subject := range s.subjects(identity) {
		count, err := s.store.Increment(ctx, orderCreateKey(subject.dimension, subject.value), ttl)
		if err != nil {
			return decision, ErrServiceUnavailable
		}
		if count > int64(subject.limit) {
			return Decision{
				Allowed:    false,
				Action:     ActionBlock,
				Dimension:  subject.dimension,
				Count:      count,
				Limit:      subject.limit,
				RetryAfter: ttl,
			}, nil
		}
	}

	return decision, nil
}

type subject struct {
	dimension string
	value     string
	limit     int
}

func (s *Service) subjects(identity Identity) []subject {
	items := make([]subject, 0, 3)
	if identity.UserID > 0 && s.cfg.MaxOrderCreationsPerUser > 0 {
		items = append(items, subject{
			dimension: "user",
			value:     stringUint(identity.UserID),
			limit:     s.cfg.MaxOrderCreationsPerUser,
		})
	}
	if value := strings.TrimSpace(identity.SessionID); value != "" && s.cfg.MaxOrderCreationsPerSession > 0 {
		items = append(items, subject{
			dimension: "session",
			value:     value,
			limit:     s.cfg.MaxOrderCreationsPerSession,
		})
	}
	if value := strings.TrimSpace(identity.IPAddress); value != "" && s.cfg.MaxOrderCreationsPerIP > 0 {
		items = append(items, subject{
			dimension: "ip",
			value:     value,
			limit:     s.cfg.MaxOrderCreationsPerIP,
		})
	}
	return items
}

func orderCreateKey(dimension, value string) string {
	return "commerce_platform:order-abuse:create:" + strings.TrimSpace(dimension) + ":" + digest(value)
}

func digest(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func stringUint(value uint) string {
	const digits = "0123456789"
	if value == 0 {
		return "0"
	}

	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = digits[value%10]
		value /= 10
	}
	return string(buffer[index:])
}
