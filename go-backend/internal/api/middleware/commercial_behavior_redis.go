package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const commercialBehaviorRedisPrefix = "commerce:commercial-behavior:v1"

var commercialBehaviorObserveScript = redis.NewScript(`
local request_count = redis.call("INCR", KEYS[1])
if request_count == 1 then
  redis.call("EXPIRE", KEYS[1], ARGV[1])
end

local target_added = redis.call("SADD", KEYS[2], ARGV[2])
if target_added == 1 then
  redis.call("EXPIRE", KEYS[2], ARGV[1])
end

local unique_target_count = redis.call("SCARD", KEYS[2])
return {request_count, unique_target_count}
`)

func optionalCommercialBehaviorRedisClient(clients []*redis.Client) *redis.Client {
	if len(clients) == 0 {
		return nil
	}
	return clients[0]
}

func newCommercialBehaviorTrackerWithRedis(redisClient *redis.Client) *commercialBehaviorTracker {
	if redisClient == nil {
		return newCommercialBehaviorTracker()
	}
	return &commercialBehaviorTracker{
		redisClient: redisClient,
		fallback:    newCommercialBehaviorTracker(),
	}
}

func (t *commercialBehaviorTracker) observeRedis(
	key string,
	target string,
	numericTarget *uint64,
	now time.Time,
) (requestCount, uniqueTargetCount, sequenceStreak int, exceeded bool) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	_ = numericTarget

	target = targetOrPlaceholder(target)
	digest := commercialBehaviorDigest(key)
	countKey := commercialBehaviorRedisPrefix + ":count:" + digest
	targetsKey := commercialBehaviorRedisPrefix + ":targets:" + digest
	ttlSeconds := int(commercialBehaviorWindow / time.Second)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}

	result, err := commercialBehaviorObserveScript.Run(
		context.Background(),
		t.redisClient,
		[]string{countKey, targetsKey},
		ttlSeconds,
		target,
	).Result()
	if err != nil {
		if t.fallback != nil {
			return t.fallback.observe(key, target, numericTarget, now)
		}
		return 0, 0, 0, false
	}

	values, ok := result.([]interface{})
	if !ok || len(values) != 2 {
		if t.fallback != nil {
			return t.fallback.observe(key, target, numericTarget, now)
		}
		return 0, 0, 0, false
	}

	requestCount, ok = redisResultInt(values[0])
	if !ok {
		return 0, 0, 0, false
	}
	uniqueTargetCount, ok = redisResultInt(values[1])
	if !ok {
		return 0, 0, 0, false
	}
	return requestCount, uniqueTargetCount, 0, false
}

func commercialBehaviorDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func redisResultInt(value interface{}) (int, bool) {
	switch typed := value.(type) {
	case int64:
		return int(typed), true
	case int:
		return typed, true
	case string:
		parsed, err := strconv.Atoi(typed)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func targetOrPlaceholder(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
