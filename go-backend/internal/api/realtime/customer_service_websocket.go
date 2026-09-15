package realtime

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/service"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	customerServiceWebSocketReadBufferSize      = 1024
	customerServiceWebSocketWriteBufferSize     = 1024
	customerServiceWebSocketMaxConnections      = 500
	customerServiceWebSocketMaxMessageSize      = 4 << 10
	customerServiceWebSocketWriteWait           = 10 * time.Second
	customerServiceWebSocketPongWait            = 60 * time.Second
	customerServiceWebSocketPingPeriod          = 50 * time.Second
	customerServiceWebSocketOutboundBuffer      = 256
	customerServiceWebSocketMaxConnectionsPerIP = 5
	customerServiceWebSocketLeaseTTL            = 2 * time.Minute
)

var customerServiceWebSocketConnections atomic.Int64

// CustomerServiceWebSocketControl is deliberately limited to transient
// controls. Durable conversation mutations continue to use the HTTP commands.
type CustomerServiceWebSocketControl struct {
	Type     string `json:"type"`
	IsTyping *bool  `json:"is_typing,omitempty"`
}

type customerServiceWebSocketFrame struct {
	Type   string                                `json:"type"`
	Cursor string                                `json:"cursor,omitempty"`
	Event  *service.CustomerServiceRealtimeEvent `json:"event,omitempty"`
	Code   string                                `json:"code,omitempty"`
}

// CustomerServiceWebSocketOptions scopes one already-authorized connection.
// Subscription must be created before Serve so replay and live delivery cannot
// leave an intentional subscription gap.
type CustomerServiceWebSocketOptions struct {
	CheckOrigin       func(*http.Request) bool
	Subscription      *service.CustomerServiceEventSubscription
	Replay            []service.CustomerServiceRealtimeEvent
	AllowEvent        func(service.CustomerServiceRealtimeEvent) bool
	HandleControl     func(CustomerServiceWebSocketControl)
	ConnectionLimiter *CustomerServiceWebSocketLimiter
	ClientIP          string
}

// CustomerServiceWebSocketLimiter stores expiring connection leases in Redis.
// A sorted set gives every socket its own lease, so a dead process cannot leave
// an IP permanently blocked and concurrent replicas share one atomic limit.
type CustomerServiceWebSocketLimiter struct {
	client   redis.UniversalClient
	maxPerIP int
	ttl      time.Duration
}

type CustomerServiceWebSocketLease struct {
	limiter *CustomerServiceWebSocketLimiter
	key     string
	token   string
	once    sync.Once
}

var (
	acquireCustomerServiceWebSocketLeaseScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
local max_connections = tonumber(ARGV[3])
redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now)
local count = redis.call("ZCARD", KEYS[1])
if count >= max_connections then
  if count > 0 then redis.call("EXPIRE", KEYS[1], ttl) end
  return 0
end
redis.call("ZADD", KEYS[1], now + (ttl * 1000), ARGV[4])
redis.call("EXPIRE", KEYS[1], ttl)
return 1
`)
	renewCustomerServiceWebSocketLeaseScript = redis.NewScript(`
local now = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])
if redis.call("ZSCORE", KEYS[1], ARGV[3]) == false then return 0 end
redis.call("ZADD", KEYS[1], now + (ttl * 1000), ARGV[3])
redis.call("EXPIRE", KEYS[1], ttl)
return 1
`)
	releaseCustomerServiceWebSocketLeaseScript = redis.NewScript(`
redis.call("ZREM", KEYS[1], ARGV[1])
if redis.call("ZCARD", KEYS[1]) == 0 then redis.call("DEL", KEYS[1]) end
return 1
`)
)

func NewCustomerServiceWebSocketLimiter(client redis.UniversalClient, maxPerIP int, leaseTTL time.Duration) *CustomerServiceWebSocketLimiter {
	if maxPerIP <= 0 {
		maxPerIP = customerServiceWebSocketMaxConnectionsPerIP
	}
	if leaseTTL <= 0 {
		leaseTTL = customerServiceWebSocketLeaseTTL
	}
	return &CustomerServiceWebSocketLimiter{client: client, maxPerIP: maxPerIP, ttl: leaseTTL}
}

func (l *CustomerServiceWebSocketLimiter) Acquire(ctx context.Context, clientIP string) (*CustomerServiceWebSocketLease, error) {
	if l == nil || l.client == nil {
		return nil, errors.New("customer-service websocket connection limiter unavailable")
	}
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" {
		return nil, errors.New("customer-service websocket client IP unavailable")
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)
	key := customerServiceWebSocketIPKey(clientIP)
	ttlSeconds := int64(l.ttl / time.Second)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	result, err := acquireCustomerServiceWebSocketLeaseScript.Run(ctx, l.client, []string{key},
		time.Now().UnixMilli(), ttlSeconds, l.maxPerIP, token).Int()
	if err != nil {
		return nil, err
	}
	if result != 1 {
		return nil, errCustomerServiceWebSocketIPLimit
	}
	return &CustomerServiceWebSocketLease{limiter: l, key: key, token: token}, nil
}

func (l *CustomerServiceWebSocketLease) Renew(ctx context.Context) error {
	if l == nil || l.limiter == nil || l.limiter.client == nil {
		return errors.New("customer-service websocket lease unavailable")
	}
	ttlSeconds := int64(l.limiter.ttl / time.Second)
	if ttlSeconds < 1 {
		ttlSeconds = 1
	}
	result, err := renewCustomerServiceWebSocketLeaseScript.Run(ctx, l.limiter.client, []string{l.key},
		time.Now().UnixMilli(), ttlSeconds, l.token).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("customer-service websocket lease expired")
	}
	return nil
}

func (l *CustomerServiceWebSocketLease) Release(ctx context.Context) error {
	if l == nil || l.limiter == nil || l.limiter.client == nil {
		return nil
	}
	var err error
	l.once.Do(func() {
		err = releaseCustomerServiceWebSocketLeaseScript.Run(ctx, l.limiter.client, []string{l.key}, l.token).Err()
	})
	return err
}

func customerServiceWebSocketIPKey(clientIP string) string {
	// The IP is already resolved by Gin using its trusted-proxy configuration.
	// Hashing keeps the Redis key bounded and avoids exposing it in key dumps.
	sum := sha256.Sum256([]byte(clientIP))
	return "commerce_platform:customer_service:websocket:ip:" + hex.EncodeToString(sum[:])
}

var errCustomerServiceWebSocketIPLimit = errors.New("customer-service websocket IP connection limit reached")

// ServeCustomerServiceWebSocket owns all writes for one socket. A full outbound
// queue closes the connection so the client reconnects and reconciles through
// the authoritative HTTP APIs instead of ever blocking an event publisher.
func ServeCustomerServiceWebSocket(w http.ResponseWriter, r *http.Request, options CustomerServiceWebSocketOptions) {
	if options.Subscription == nil || options.Subscription.Events() == nil {
		http.Error(w, "customer service realtime unavailable", http.StatusServiceUnavailable)
		return
	}
	// The handler creates the subscription before entering Serve to avoid an
	// event gap. Consequently Serve must own cleanup even when capacity checks
	// reject the handshake before a WebSocket is upgraded.
	defer options.Subscription.Cancel()
	if !acquireCustomerServiceWebSocketConnection() {
		metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("capacity_rejected").Inc()
		http.Error(w, "too many websocket connections", http.StatusServiceUnavailable)
		return
	}
	defer releaseCustomerServiceWebSocketConnection()

	var lease *CustomerServiceWebSocketLease
	if options.ConnectionLimiter != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		lease, err := options.ConnectionLimiter.Acquire(ctx, options.ClientIP)
		cancel()
		if err != nil {
			if errors.Is(err, errCustomerServiceWebSocketIPLimit) {
				metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("ip_capacity_rejected").Inc()
				http.Error(w, "too many websocket connections from this IP", http.StatusTooManyRequests)
			} else {
				metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("limiter_unavailable").Inc()
				appLogger.Warn("customer-service websocket distributed limiter unavailable", zap.Error(err))
				http.Error(w, "customer service realtime unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := lease.Release(ctx); err != nil {
				appLogger.Warn("customer-service websocket lease release failed", zap.Error(err))
			}
		}()
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  customerServiceWebSocketReadBufferSize,
		WriteBufferSize: customerServiceWebSocketWriteBufferSize,
		CheckOrigin:     options.CheckOrigin,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("upgrade_failed").Inc()
		appLogger.Warn("customer-service websocket upgrade failed", zap.Error(err))
		return
	}
	metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("accepted").Inc()

	done := make(chan struct{})
	var stopOnce sync.Once
	stop := func() {
		stopOnce.Do(func() {
			close(done)
		})
	}
	if lease != nil {
		go func() {
			renewEvery := lease.limiter.ttl / 2
			if renewEvery < time.Second {
				renewEvery = time.Second
			}
			ticker := time.NewTicker(renewEvery)
			defer ticker.Stop()
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
					err := lease.Renew(ctx)
					cancel()
					if err != nil {
						metrics.CustomerServiceRealtimeWebSocketConnectionAttempts.WithLabelValues("lease_renewal_failed").Inc()
						appLogger.Warn("customer-service websocket lease renewal failed", zap.Error(err))
						stop()
						return
					}
				}
			}
		}()
	}

	outbound := make(chan customerServiceWebSocketFrame, customerServiceWebSocketOutboundBuffer)
	enqueue := func(frame customerServiceWebSocketFrame) bool {
		select {
		case <-done:
			return false
		case outbound <- frame:
			return true
		default:
			metrics.CustomerServiceRealtimeWebSocketOutboundOverflows.Inc()
			appLogger.Warn("customer-service websocket outbound queue overflow")
			stop()
			return false
		}
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		defer func() { _ = conn.Close() }()

		pingTicker := time.NewTicker(customerServiceWebSocketPingPeriod)
		defer pingTicker.Stop()

		for {
			select {
			case <-done:
				_ = conn.WriteControl(
					websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing"),
					time.Now().Add(customerServiceWebSocketWriteWait),
				)
				return
			case frame := <-outbound:
				if err := conn.SetWriteDeadline(time.Now().Add(customerServiceWebSocketWriteWait)); err != nil {
					stop()
					return
				}
				if err := conn.WriteJSON(frame); err != nil {
					stop()
					return
				}
			case <-pingTicker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(customerServiceWebSocketWriteWait)); err != nil {
					stop()
					return
				}
			}
		}
	}()

	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		defer stop()

		conn.SetReadLimit(customerServiceWebSocketMaxMessageSize)
		_ = conn.SetReadDeadline(time.Now().Add(customerServiceWebSocketPongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(customerServiceWebSocketPongWait))
		})

		for {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					appLogger.Warn("customer-service websocket unexpected close", zap.Error(err))
				}
				return
			}
			if messageType != websocket.TextMessage {
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "text_frame_required"})
				continue
			}

			control, ok := decodeCustomerServiceWebSocketControl(payload)
			if !ok {
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "invalid_control"})
				continue
			}
			switch control.Type {
			case "ping":
				enqueue(customerServiceWebSocketFrame{Type: "pong"})
			case "typing":
				if options.HandleControl != nil {
					options.HandleControl(control)
				}
			default:
				enqueue(customerServiceWebSocketFrame{Type: "error", Code: "unsupported_control"})
			}
		}
	}()

	defer func() {
		stop()
		<-writerDone
		<-readerDone
	}()

	deduper := service.NewCustomerServiceRealtimeEventDeduper(2048)
	deliver := func(event service.CustomerServiceRealtimeEvent) {
		if options.AllowEvent != nil && !options.AllowEvent(event) {
			if event.StreamID != "" {
				enqueue(customerServiceWebSocketFrame{Type: "cursor", Cursor: event.StreamID})
			}
			return
		}
		if !deduper.First(event) {
			return
		}
		eventCopy := event
		enqueue(customerServiceWebSocketFrame{Type: "event", Cursor: event.StreamID, Event: &eventCopy})
	}

	if !enqueue(customerServiceWebSocketFrame{Type: "ready"}) {
		return
	}
	for _, event := range options.Replay {
		deliver(event)
	}

	for {
		select {
		case <-done:
			return
		case <-readerDone:
			return
		case event, ok := <-options.Subscription.Events():
			if !ok {
				return
			}
			deliver(event)
		}
	}
}

func decodeCustomerServiceWebSocketControl(payload []byte) (CustomerServiceWebSocketControl, bool) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()

	var control CustomerServiceWebSocketControl
	if err := decoder.Decode(&control); err != nil {
		return CustomerServiceWebSocketControl{}, false
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return CustomerServiceWebSocketControl{}, false
	}

	control.Type = strings.ToLower(strings.TrimSpace(control.Type))
	return control, control.Type != ""
}

// CustomerServiceWebSocketOriginAllowed requires an explicit same-origin or
// configured allowed-origin browser upgrade. Native browser sockets always send
// Origin, so accepting a missing value would weaken the cookie-auth boundary.
func CustomerServiceWebSocketOriginAllowed(r *http.Request, allowedOrigins []string) bool {
	if r == nil {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Scheme == "" || originURL.Host == "" {
		return false
	}
	if sameCustomerServiceWebSocketOrigin(originURL, requestCustomerServiceWebSocketOrigin(r)) {
		return true
	}

	for _, allowedOrigin := range allowedOrigins {
		allowedURL, err := url.Parse(strings.TrimSpace(allowedOrigin))
		if err != nil || allowedURL.Scheme == "" || allowedURL.Host == "" {
			continue
		}
		if sameCustomerServiceWebSocketOrigin(originURL, allowedURL) {
			return true
		}
	}
	return false
}

func requestCustomerServiceWebSocketOrigin(r *http.Request) *url.URL {
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if comma := strings.IndexByte(scheme, ','); comma >= 0 {
		scheme = strings.TrimSpace(scheme[:comma])
	}
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return &url.URL{Scheme: scheme, Host: r.Host}
}

func sameCustomerServiceWebSocketOrigin(left, right *url.URL) bool {
	if left == nil || right == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(left.Scheme), strings.TrimSpace(right.Scheme)) &&
		strings.EqualFold(strings.TrimSpace(left.Host), strings.TrimSpace(right.Host))
}

func acquireCustomerServiceWebSocketConnection() bool {
	if customerServiceWebSocketConnections.Add(1) <= customerServiceWebSocketMaxConnections {
		metrics.CustomerServiceRealtimeWebSocketConnections.Inc()
		return true
	}
	customerServiceWebSocketConnections.Add(-1)
	return false
}

func releaseCustomerServiceWebSocketConnection() {
	customerServiceWebSocketConnections.Add(-1)
	metrics.CustomerServiceRealtimeWebSocketConnections.Dec()
}
