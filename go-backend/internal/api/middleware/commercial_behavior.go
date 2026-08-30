package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"commerce-platform/internal/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	commercialBehaviorWindow               = time.Minute
	commercialInventoryRequestThreshold    = 120
	commercialInventoryTargetThreshold     = 40
	commercialOrderRequestThreshold        = 20
	commercialOrderSequenceThreshold       = 4
	commercialBehaviorMaxTrackedIdentities = 10000
)

type commercialBehaviorBucket struct {
	expiresAt         time.Time
	requestCount      int
	targets           map[string]struct{}
	lastNumericTarget uint64
	sequenceStreak    int
}

type commercialBehaviorTracker struct {
	mu          sync.Mutex
	buckets     map[string]*commercialBehaviorBucket
	redisClient redis.UniversalClient
	fallback    *commercialBehaviorTracker
}

func newCommercialBehaviorTracker() *commercialBehaviorTracker {
	return &commercialBehaviorTracker{
		buckets: make(map[string]*commercialBehaviorBucket),
	}
}

func (t *commercialBehaviorTracker) observe(
	ctx context.Context,
	key string,
	target string,
	numericTarget *uint64,
	now time.Time,
) (requestCount, uniqueTargetCount, sequenceStreak int, exceeded bool) {
	if t == nil || strings.TrimSpace(key) == "" {
		return 0, 0, 0, false
	}
	if t.redisClient != nil {
		return t.observeRedis(ctx, key, target, numericTarget, now)
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	bucket := t.buckets[key]
	if bucket == nil || !now.Before(bucket.expiresAt) {
		t.cleanupExpired(now)
		if bucket == nil && len(t.buckets) >= commercialBehaviorMaxTrackedIdentities {
			return 0, 0, 0, false
		}
		bucket = &commercialBehaviorBucket{
			expiresAt: now.Add(commercialBehaviorWindow),
			targets:   make(map[string]struct{}),
		}
		t.buckets[key] = bucket
	}

	bucket.requestCount++
	target = strings.TrimSpace(target)
	if target != "" {
		bucket.targets[target] = struct{}{}
	}
	if numericTarget != nil {
		if bucket.lastNumericTarget != 0 &&
			commercialNumericTargetsAreAdjacent(*numericTarget, bucket.lastNumericTarget) {
			bucket.sequenceStreak++
		} else {
			bucket.sequenceStreak = 0
		}
		bucket.lastNumericTarget = *numericTarget
	}

	requestCount = bucket.requestCount
	uniqueTargetCount = len(bucket.targets)
	sequenceStreak = bucket.sequenceStreak
	return requestCount, uniqueTargetCount, sequenceStreak, false
}

func commercialRequestContext(c *gin.Context) context.Context {
	if c == nil || c.Request == nil {
		return context.Background()
	}
	if ctx := c.Request.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func commercialNumericTargetsAreAdjacent(first, second uint64) bool {
	return (first > second && first-second == 1) ||
		(second > first && second-first == 1)
}

func (t *commercialBehaviorTracker) cleanupExpired(now time.Time) {
	for key, bucket := range t.buckets {
		if bucket == nil || !now.Before(bucket.expiresAt) {
			delete(t.buckets, key)
		}
	}
}

// CommercialInventoryProbeGuard observes public product reads and rate-limits
// high-volume catalog enumeration. Internal SSR calls without a forwarded
// client identity are deliberately ignored.
func CommercialInventoryProbeGuard(redisClients ...redis.UniversalClient) gin.HandlerFunc {
	tracker := newCommercialBehaviorTrackerWithRedis(optionalCommercialBehaviorRedisClient(redisClients))

	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet || !isCommercialProductReadPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		identity := commercialRequestIdentity(c)
		if identity == "" {
			c.Next()
			return
		}

		target := commercialProductTarget(c.Request.URL.Path)
		requestCount, uniqueTargetCount, _, _ := tracker.observe(
			commercialRequestContext(c),
			identity,
			target,
			nil,
			time.Now().UTC(),
		)
		if requestCount > commercialInventoryRequestThreshold ||
			uniqueTargetCount > commercialInventoryTargetThreshold {
			abortCommercialIntelligenceRateLimit(c, "inventory-crawlers")
			return
		}

		c.Next()
	}
}

// CommercialOrderEnumerationGuard protects authenticated order detail reads
// from high-volume guessing. Public order numbers are opaque, so a numeric
// adjacency signal is no longer available or needed.
func CommercialOrderEnumerationGuard(redisClients ...redis.UniversalClient) gin.HandlerFunc {
	tracker := newCommercialBehaviorTrackerWithRedis(optionalCommercialBehaviorRedisClient(redisClients))

	return func(c *gin.Context) {
		orderNumber := strings.TrimSpace(c.Param("order_number"))
		if c.Request.Method != http.MethodGet || orderNumber == "" {
			c.Next()
			return
		}

		identity := commercialRequestIdentity(c)
		if userID, exists := c.Get("user_id"); exists {
			identity = "user:" + strconv.FormatUint(uint64(userID.(uint)), 10) + "|" + identity
		}
		if identity == "" {
			c.Next()
			return
		}

		requestCount, _, _, _ := tracker.observe(
			commercialRequestContext(c),
			identity,
			orderNumber,
			nil,
			time.Now().UTC(),
		)
		if requestCount > commercialOrderRequestThreshold {
			abortCommercialIntelligenceRateLimit(c, "order-scrapers")
			return
		}

		c.Next()
	}
}

// CommercialCartProbeGuard limits the high-volume cart mutations commonly used
// to discover inventory thresholds. It never fabricates inventory responses:
// authoritative availability remains validated by the cart service.
func CommercialCartProbeGuard(redisClients ...redis.UniversalClient) gin.HandlerFunc {
	tracker := newCommercialBehaviorTrackerWithRedis(optionalCommercialBehaviorRedisClient(redisClients))

	return func(c *gin.Context) {
		if !isCommercialCartWrite(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		identity := commercialRequestIdentity(c)
		if identity == "" {
			c.Next()
			return
		}

		inspection := inspectCommercialCartRequest(c)
		if inspection.quantityExceeded {
			abortCommercialCartQuantityLimit(c)
			return
		}

		identities := commercialCartIdentities(c)
		if len(identities) == 0 {
			c.Next()
			return
		}

		threshold := commercialCartWriteRequestThreshold
		if !commercialRequestHasCartSession(c) {
			threshold = commercialAnonymousCartWriteThreshold
		}

		now := time.Now().UTC()
		pathTarget := "cart:" + strings.Trim(strings.TrimSpace(c.Request.URL.Path), "/")
		for _, identityKey := range identities {
			requestCount, _, _, _ := tracker.observe(commercialRequestContext(c), identityKey, pathTarget, nil, now)
			if requestCount > threshold {
				abortCommercialIntelligenceRateLimit(c, "inventory-crawlers")
				return
			}
			for _, itemTarget := range inspection.targets {
				targetCount, _, _, _ := tracker.observe(
					commercialRequestContext(c),
					identityKey+"|target:"+itemTarget,
					"",
					nil,
					now,
				)
				if targetCount > commercialCartTargetRequestThreshold {
					abortCommercialIntelligenceRateLimit(c, "inventory-crawlers")
					return
				}
			}
		}

		c.Next()
	}
}

func isCommercialProductReadPath(path string) bool {
	path = strings.Trim(strings.TrimSpace(path), "/")
	if path == "api/v1/products" || path == "api/v1/customer-service/products" {
		return true
	}
	if !strings.HasPrefix(path, "api/v1/products/") {
		return false
	}

	suffix := strings.TrimPrefix(path, "api/v1/products/")
	return suffix != "types" && suffix != "attributes/filterable"
}

func isCommercialCartWrite(method, path string) bool {
	if method != http.MethodPost && method != http.MethodPut {
		return false
	}

	path = strings.Trim(strings.TrimSpace(path), "/")
	return path == "api/v1/cart/add" ||
		path == "api/v1/cart/sync" ||
		strings.HasPrefix(path, "api/v1/cart/items/")
}

func commercialRequestHasCartSession(c *gin.Context) bool {
	if c == nil {
		return false
	}
	sessionID, err := c.Cookie("session_id")
	return err == nil && strings.TrimSpace(sessionID) != ""
}

func commercialProductTarget(path string) string {
	path = strings.Trim(strings.TrimSpace(path), "/")
	if path == "api/v1/products" {
		return "catalog:list"
	}
	if path == "api/v1/customer-service/products" {
		return "catalog:customer-service-search"
	}
	return strings.TrimPrefix(path, "api/v1/products/")
}

func commercialRequestIdentity(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}

	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return ""
	}

	// Nuxt SSR calls the API from inside the Docker network. Without a
	// forwarded client address, do not aggregate those calls as one crawler.
	if strings.TrimSpace(c.GetHeader("X-Forwarded-For")) == "" && isPrivateNetworkIP(ip) {
		return ""
	}
	return ip
}

func isPrivateNetworkIP(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	privateBlocks := []*net.IPNet{
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
		{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
	}
	for _, block := range privateBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func abortCommercialIntelligenceRateLimit(c *gin.Context, seedID string) {
	metrics.CommercialIntelligenceActions.WithLabelValues(seedID, "rate_limited").Inc()
	logCommercialProtectionAction(c, commercialProtectionAuditInput{
		SeedID:      seedID,
		Outcome:     "rate_limited",
		Action:      "rate_limit_429",
		Reason:      "behavior_threshold_exceeded",
		Window:      commercialBehaviorWindow,
		ReleaseMode: "automatic_window_expiry",
	})
	c.Header("Retry-After", strconv.Itoa(int(commercialBehaviorWindow.Seconds())))
	c.Header("Cache-Control", "no-store, max-age=0")
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error":   "commercial_intelligence_rate_limited",
		"message": "Too many requests. Please try again later.",
	})
}

func abortCommercialCartQuantityLimit(c *gin.Context) {
	metrics.CommercialIntelligenceActions.WithLabelValues("inventory-crawlers", "quantity_rejected").Inc()
	logCommercialProtectionAction(c, commercialProtectionAuditInput{
		SeedID:      "inventory-crawlers",
		Outcome:     "quantity_rejected",
		Action:      "reject_quantity_400",
		Reason:      "cart_quantity_exceeds_public_limit",
		ReleaseMode: "request_scope",
	})
	c.Header("Cache-Control", "no-store, max-age=0")
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error":   "cart_quantity_limit",
		"message": "The requested quantity cannot be added to the cart.",
	})
}
