package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/config"

	"github.com/redis/go-redis/v9"
)

var ErrShowcaseUploadProtectionUnavailable = errors.New("showcase upload protection is unavailable")

const (
	UGCShowcaseUploadProtectionDecisionAllow = "allow"
	UGCShowcaseUploadProtectionDecisionBlock = "block"
)

type UGCShowcaseUploadProtectionIdentity struct {
	UserID    uint
	IPAddress string
}

type UGCShowcaseUploadProtectionInput struct {
	Identity           UGCShowcaseUploadProtectionIdentity
	UploadBytes        int64
	PendingSubmissions int64
}

type UGCShowcaseUploadProtectionDecision struct {
	Allowed    bool
	Action     string
	Dimension  string
	Count      int64
	Limit      int64
	RetryAfter time.Duration
}

type UGCShowcaseUploadProtectionService struct {
	store       showcaseUploadProtectionStore
	cfg         config.ShowcaseUploadProtectionConfig
	now         func() time.Time
	unavailable bool
}

type showcaseUploadProtectionStore interface {
	Evaluate(context.Context, showcaseUploadProtectionEvaluation) (showcaseUploadProtectionStoreResult, error)
	RecordFailure(context.Context, showcaseUploadProtectionFailure) error
}

type showcaseUploadProtectionEvaluation struct {
	Keys []string
	Args []interface{}
}

type showcaseUploadProtectionStoreResult struct {
	Blocked      bool
	Reason       string
	Count        int64
	Limit        int64
	RetrySeconds int64
}

type showcaseUploadProtectionFailure struct {
	Keys []string
	Args []interface{}
}

type redisShowcaseUploadProtectionStore struct {
	client redis.UniversalClient
}

func NewUGCShowcaseUploadProtectionService(redisClient redis.UniversalClient, cfg config.ShowcaseUploadProtectionConfig) *UGCShowcaseUploadProtectionService {
	if redisClient == nil {
		return &UGCShowcaseUploadProtectionService{
			cfg:         cfg,
			now:         time.Now,
			unavailable: cfg.Enabled,
		}
	}
	return &UGCShowcaseUploadProtectionService{
		store: redisShowcaseUploadProtectionStore{client: redisClient},
		cfg:   cfg,
		now:   time.Now,
	}
}

func (s *UGCShowcaseUploadProtectionService) Evaluate(ctx context.Context, input UGCShowcaseUploadProtectionInput) (UGCShowcaseUploadProtectionDecision, error) {
	decision := UGCShowcaseUploadProtectionDecision{
		Allowed: true,
		Action:  UGCShowcaseUploadProtectionDecisionAllow,
	}
	if s == nil || !s.cfg.Enabled {
		return decision, nil
	}
	if input.PendingSubmissions >= int64(s.cfg.MaxPendingSubmissionsPerUser) {
		return UGCShowcaseUploadProtectionDecision{
			Allowed:   false,
			Action:    UGCShowcaseUploadProtectionDecisionBlock,
			Dimension: "pending_submissions",
			Count:     input.PendingSubmissions,
			Limit:     int64(s.cfg.MaxPendingSubmissionsPerUser),
		}, nil
	}
	if s.unavailable || s.store == nil {
		return decision, ErrShowcaseUploadProtectionUnavailable
	}

	evaluation := s.evaluation(input)
	result, err := s.store.Evaluate(ctx, evaluation)
	if err != nil {
		return decision, ErrShowcaseUploadProtectionUnavailable
	}
	if !result.Blocked {
		return decision, nil
	}
	return UGCShowcaseUploadProtectionDecision{
		Allowed:    false,
		Action:     UGCShowcaseUploadProtectionDecisionBlock,
		Dimension:  result.Reason,
		Count:      result.Count,
		Limit:      result.Limit,
		RetryAfter: time.Duration(result.RetrySeconds) * time.Second,
	}, nil
}

func (s *UGCShowcaseUploadProtectionService) RecordFailure(ctx context.Context, identity UGCShowcaseUploadProtectionIdentity) error {
	if s == nil || !s.cfg.Enabled {
		return nil
	}
	if s.unavailable || s.store == nil {
		return ErrShowcaseUploadProtectionUnavailable
	}
	if err := s.store.RecordFailure(ctx, s.failure(identity)); err != nil {
		return ErrShowcaseUploadProtectionUnavailable
	}
	return nil
}

func (s *UGCShowcaseUploadProtectionService) evaluation(input UGCShowcaseUploadProtectionInput) showcaseUploadProtectionEvaluation {
	day := s.now().UTC().Format("20060102")
	userKey := showcaseUploadProtectionUserKey(input.Identity.UserID)
	ipKey := digestShowcaseUploadProtectionValue(input.Identity.IPAddress)
	ipPrefixKey := digestShowcaseUploadProtectionValue(ipPrefix(input.Identity.IPAddress))
	return showcaseUploadProtectionEvaluation{
		Keys: []string{
			showcaseUploadProtectionCounterKey("window:user", userKey),
			showcaseUploadProtectionCounterKey("window:ip", ipKey),
			showcaseUploadProtectionCounterKey("window:ip-prefix", ipPrefixKey),
			showcaseUploadProtectionCounterKey("daily:user", day+":"+userKey),
			showcaseUploadProtectionCounterKey("daily:ip", day+":"+ipKey),
			showcaseUploadProtectionCounterKey("daily-bytes:user", day+":"+userKey),
			showcaseUploadProtectionCounterKey("daily-bytes:ip", day+":"+ipKey),
			showcaseUploadProtectionCounterKey("block:user", userKey),
			showcaseUploadProtectionCounterKey("block:ip", ipKey),
		},
		Args: []interface{}{
			s.cfg.WindowSeconds,
			s.cfg.MaxUploadsPerUser,
			s.cfg.MaxUploadsPerIP,
			s.cfg.MaxUploadsPerIPPrefix,
			s.cfg.DailyMaxUploadsPerUser,
			s.cfg.DailyMaxUploadsPerIP,
			s.cfg.DailyMaxBytesPerUser,
			s.cfg.DailyMaxBytesPerIP,
			input.UploadBytes,
			secondsUntilNextUTCDay(s.now()),
			s.cfg.BlockDurationSeconds,
		},
	}
}

func (s *UGCShowcaseUploadProtectionService) failure(identity UGCShowcaseUploadProtectionIdentity) showcaseUploadProtectionFailure {
	userKey := showcaseUploadProtectionUserKey(identity.UserID)
	ipKey := digestShowcaseUploadProtectionValue(identity.IPAddress)
	return showcaseUploadProtectionFailure{
		Keys: []string{
			showcaseUploadProtectionCounterKey("failure:user", userKey),
			showcaseUploadProtectionCounterKey("failure:ip", ipKey),
			showcaseUploadProtectionCounterKey("block:user", userKey),
			showcaseUploadProtectionCounterKey("block:ip", ipKey),
		},
		Args: []interface{}{
			s.cfg.FailureWindowSeconds,
			s.cfg.MaxFailuresPerUser,
			s.cfg.MaxFailuresPerIP,
			s.cfg.BlockDurationSeconds,
		},
	}
}

func (s redisShowcaseUploadProtectionStore) Evaluate(ctx context.Context, evaluation showcaseUploadProtectionEvaluation) (showcaseUploadProtectionStoreResult, error) {
	values, err := s.client.Eval(ctx, showcaseUploadProtectionEvaluateScript, evaluation.Keys, evaluation.Args...).Slice()
	if err != nil {
		return showcaseUploadProtectionStoreResult{}, err
	}
	return decodeShowcaseUploadProtectionStoreResult(values), nil
}

func (s redisShowcaseUploadProtectionStore) RecordFailure(ctx context.Context, failure showcaseUploadProtectionFailure) error {
	return s.client.Eval(ctx, showcaseUploadProtectionFailureScript, failure.Keys, failure.Args...).Err()
}

func decodeShowcaseUploadProtectionStoreResult(values []interface{}) showcaseUploadProtectionStoreResult {
	result := showcaseUploadProtectionStoreResult{}
	if len(values) < 5 {
		return result
	}
	result.Blocked = redisScriptInt64(values[0]) == 1
	result.Reason = redisScriptString(values[1])
	result.Count = redisScriptInt64(values[2])
	result.Limit = redisScriptInt64(values[3])
	result.RetrySeconds = redisScriptInt64(values[4])
	return result
}

func redisScriptInt64(value interface{}) int64 {
	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	case string:
		var result int64
		for _, r := range typed {
			if r < '0' || r > '9' {
				return 0
			}
			result = result*10 + int64(r-'0')
		}
		return result
	default:
		return 0
	}
}

func redisScriptString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func showcaseUploadProtectionCounterKey(scope, value string) string {
	return "commerce_platform:showcase-upload:" + strings.TrimSpace(scope) + ":" + digestShowcaseUploadProtectionValue(value)
}

func showcaseUploadProtectionUserKey(userID uint) string {
	return strconv.FormatUint(uint64(userID), 10)
}

func digestShowcaseUploadProtectionValue(value string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:])
}

func secondsUntilNextUTCDay(now time.Time) int64 {
	now = now.UTC()
	next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	seconds := int64(next.Sub(now).Seconds())
	if seconds <= 0 {
		return int64((24 * time.Hour) / time.Second)
	}
	return seconds
}

func ipPrefix(ip string) string {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return strings.TrimSpace(ip)
	}
	if ipv4 := parsed.To4(); ipv4 != nil {
		return net.IPv4(ipv4[0], ipv4[1], ipv4[2], 0).String() + "/24"
	}
	ipv6 := parsed.To16()
	if ipv6 == nil {
		return strings.TrimSpace(ip)
	}
	prefix := make(net.IP, len(ipv6))
	copy(prefix, ipv6)
	for i := 8; i < len(prefix); i++ {
		prefix[i] = 0
	}
	return prefix.String() + "/64"
}

const showcaseUploadProtectionEvaluateScript = `
local window_ttl = tonumber(ARGV[1])
local max_user_window = tonumber(ARGV[2])
local max_ip_window = tonumber(ARGV[3])
local max_prefix_window = tonumber(ARGV[4])
local max_user_daily = tonumber(ARGV[5])
local max_ip_daily = tonumber(ARGV[6])
local max_user_daily_bytes = tonumber(ARGV[7])
local max_ip_daily_bytes = tonumber(ARGV[8])
local upload_bytes = tonumber(ARGV[9])
local daily_ttl = tonumber(ARGV[10])
local block_ttl = tonumber(ARGV[11])

if redis.call("EXISTS", KEYS[8]) == 1 then
	return {1, "user_block", 1, 0, block_ttl}
end
if redis.call("EXISTS", KEYS[9]) == 1 then
	return {1, "ip_block", 1, 0, block_ttl}
end

local function increment_with_ttl(key, ttl)
	local count = redis.call("INCR", key)
	if count == 1 and ttl > 0 then
		redis.call("EXPIRE", key, ttl)
	end
	return count
end

local function increment_by_with_ttl(key, amount, ttl)
	local count = redis.call("INCRBY", key, amount)
	if count == amount and ttl > 0 then
		redis.call("EXPIRE", key, ttl)
	end
	return count
end

local user_window = increment_with_ttl(KEYS[1], window_ttl)
if user_window > max_user_window then
	return {1, "user_window", user_window, max_user_window, window_ttl}
end
local ip_window = increment_with_ttl(KEYS[2], window_ttl)
if ip_window > max_ip_window then
	return {1, "ip_window", ip_window, max_ip_window, window_ttl}
end
local prefix_window = increment_with_ttl(KEYS[3], window_ttl)
if prefix_window > max_prefix_window then
	return {1, "ip_prefix_window", prefix_window, max_prefix_window, window_ttl}
end
local user_daily = increment_with_ttl(KEYS[4], daily_ttl)
if user_daily > max_user_daily then
	return {1, "user_daily", user_daily, max_user_daily, daily_ttl}
end
local ip_daily = increment_with_ttl(KEYS[5], daily_ttl)
if ip_daily > max_ip_daily then
	return {1, "ip_daily", ip_daily, max_ip_daily, daily_ttl}
end
local user_daily_bytes = increment_by_with_ttl(KEYS[6], upload_bytes, daily_ttl)
if user_daily_bytes > max_user_daily_bytes then
	return {1, "user_daily_bytes", user_daily_bytes, max_user_daily_bytes, daily_ttl}
end
local ip_daily_bytes = increment_by_with_ttl(KEYS[7], upload_bytes, daily_ttl)
if ip_daily_bytes > max_ip_daily_bytes then
	return {1, "ip_daily_bytes", ip_daily_bytes, max_ip_daily_bytes, daily_ttl}
end

return {0, "", 0, 0, 0}
`

const showcaseUploadProtectionFailureScript = `
local failure_ttl = tonumber(ARGV[1])
local max_user_failures = tonumber(ARGV[2])
local max_ip_failures = tonumber(ARGV[3])
local block_ttl = tonumber(ARGV[4])

local function increment_with_ttl(key, ttl)
	local count = redis.call("INCR", key)
	if count == 1 and ttl > 0 then
		redis.call("EXPIRE", key, ttl)
	end
	return count
end

local user_failures = increment_with_ttl(KEYS[1], failure_ttl)
if user_failures >= max_user_failures then
	redis.call("SET", KEYS[3], "1", "EX", block_ttl)
end
local ip_failures = increment_with_ttl(KEYS[2], failure_ttl)
if ip_failures >= max_ip_failures then
	redis.call("SET", KEYS[4], "1", "EX", block_ttl)
end
return 0
`
