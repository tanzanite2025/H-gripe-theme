package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// The hash tag keeps the Stream and its event-id dedupe keys in one Redis
	// Cluster slot, so the atomic publish Lua script works in clustered Redis.
	defaultCustomerServiceRealtimeStream          = "customer_service:{realtime}:v1"
	defaultCustomerServiceRealtimeStreamMaxLen    = 10000
	defaultCustomerServiceRealtimeReplayLimit     = 200
	defaultCustomerServiceRealtimeConsumerBlock   = 5 * time.Second
	defaultCustomerServiceRealtimeDedupRetention  = 24 * time.Hour
	customerServiceRealtimeRelayReadBatchLimit    = 100
	customerServiceRealtimeRelayRetryDelay        = time.Second
	customerServiceRealtimeRelayDedupKeyNamespace = ":dedupe:"
)

var errInvalidCustomerServiceRealtimeCursor = errors.New("invalid customer-service realtime cursor")

var publishCustomerServiceRealtimeEventScript = redis.NewScript(`
local existing = redis.call('GET', KEYS[2])
if existing then
  return existing
end

local streamID = redis.call('XADD', KEYS[1], 'MAXLEN', '~', ARGV[1], '*', 'event', ARGV[2])
redis.call('SET', KEYS[2], streamID, 'EX', ARGV[3])
return streamID
`)

// CustomerServiceRealtimeRelayConfig controls the bounded Redis Stream used
// to distribute already-persisted message events to every API instance.
type CustomerServiceRealtimeRelayConfig struct {
	Stream         string
	StreamMaxLen   int64
	ReplayLimit    int
	ConsumerBlock  time.Duration
	DedupRetention time.Duration
}

func DefaultCustomerServiceRealtimeRelayConfig() CustomerServiceRealtimeRelayConfig {
	return CustomerServiceRealtimeRelayConfig{
		Stream:         defaultCustomerServiceRealtimeStream,
		StreamMaxLen:   defaultCustomerServiceRealtimeStreamMaxLen,
		ReplayLimit:    defaultCustomerServiceRealtimeReplayLimit,
		ConsumerBlock:  defaultCustomerServiceRealtimeConsumerBlock,
		DedupRetention: defaultCustomerServiceRealtimeDedupRetention,
	}
}

type CustomerServiceRealtimeRelay struct {
	client *redis.Client
	hub    *CustomerServiceEventHub
	config CustomerServiceRealtimeRelayConfig

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewCustomerServiceRealtimeRelay(
	client *redis.Client,
	hub *CustomerServiceEventHub,
	config CustomerServiceRealtimeRelayConfig,
) (*CustomerServiceRealtimeRelay, error) {
	if client == nil {
		return nil, errors.New("customer-service realtime relay requires redis")
	}
	if hub == nil {
		return nil, errors.New("customer-service realtime relay requires an event hub")
	}
	if strings.TrimSpace(config.Stream) == "" {
		config.Stream = defaultCustomerServiceRealtimeStream
	}
	if !hasCustomerServiceRealtimeRedisHashTag(config.Stream) {
		return nil, errors.New("customer-service realtime stream must include a Redis hash tag, for example customer_service:{realtime}:v1")
	}
	if config.StreamMaxLen <= 0 {
		config.StreamMaxLen = defaultCustomerServiceRealtimeStreamMaxLen
	}
	if config.ReplayLimit <= 0 {
		config.ReplayLimit = defaultCustomerServiceRealtimeReplayLimit
	}
	if config.ConsumerBlock <= 0 {
		config.ConsumerBlock = defaultCustomerServiceRealtimeConsumerBlock
	}
	if config.DedupRetention <= 0 {
		config.DedupRetention = defaultCustomerServiceRealtimeDedupRetention
	}

	return &CustomerServiceRealtimeRelay{
		client: client,
		hub:    hub,
		config: config,
	}, nil
}

// Start positions this process at the current tail before it begins XREAD.
// Using the concrete tail rather than '$' closes the small race between
// establishing the cursor and beginning the blocking read.
func (r *CustomerServiceRealtimeRelay) Start(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return nil
	}

	cursor, err := r.currentTail(ctx)
	if err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.done = make(chan struct{})
	go r.consume(runCtx, cursor, r.done)
	return nil
}

func (r *CustomerServiceRealtimeRelay) Stop() {
	if r == nil {
		return
	}

	r.mu.Lock()
	cancel := r.cancel
	done := r.done
	r.cancel = nil
	r.done = nil
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// Publish atomically appends one durable event and remembers its Stream ID.
// The stable event ID protects the Outbox retry window after XADD succeeds but
// before the database worker marks the Outbox row processed.
func (r *CustomerServiceRealtimeRelay) Publish(ctx context.Context, event CustomerServiceRealtimeEvent) (string, error) {
	if r == nil || r.client == nil {
		metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("not_configured").Inc()
		return "", errors.New("customer-service realtime relay is not configured")
	}
	if err := validateCustomerServiceRealtimeEvent(event); err != nil {
		metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("invalid_event").Inc()
		return "", err
	}
	if ctx == nil {
		ctx = context.Background()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("encode_failed").Inc()
		return "", fmt.Errorf("encode customer-service realtime event: %w", err)
	}

	seconds := int64(r.config.DedupRetention / time.Second)
	if r.config.DedupRetention%time.Second != 0 {
		seconds++
	}
	if seconds <= 0 {
		seconds = 1
	}

	streamID, err := publishCustomerServiceRealtimeEventScript.Run(
		ctx,
		r.client,
		[]string{r.config.Stream, r.dedupKey(event.EventID)},
		strconv.FormatInt(r.config.StreamMaxLen, 10),
		string(payload),
		strconv.FormatInt(seconds, 10),
	).Text()
	if err != nil {
		metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("redis_failed").Inc()
		return "", fmt.Errorf("publish customer-service realtime event to redis stream: %w", err)
	}
	if !isCustomerServiceRealtimeStreamID(streamID) {
		metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("invalid_stream_id").Inc()
		return "", fmt.Errorf("redis returned invalid customer-service stream id %q", streamID)
	}
	metrics.CustomerServiceRealtimeRelayPublishes.WithLabelValues("published").Inc()
	return streamID, nil
}

func (r *CustomerServiceRealtimeRelay) ReplayAfter(ctx context.Context, afterID string, limit int) ([]CustomerServiceRealtimeEvent, error) {
	if r == nil || r.client == nil || strings.TrimSpace(afterID) == "" {
		return nil, nil
	}
	if !isCustomerServiceRealtimeStreamID(afterID) {
		metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("invalid_cursor").Inc()
		return nil, errInvalidCustomerServiceRealtimeCursor
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if limit <= 0 || limit > r.config.ReplayLimit {
		limit = r.config.ReplayLimit
	}
	firstID, err := r.firstID(ctx)
	if err != nil {
		metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("redis_failed").Inc()
		return nil, err
	}
	if firstID == "" || compareCustomerServiceRealtimeStreamIDs(afterID, firstID) < 0 {
		// The client cursor predates the retained window. A partial replay could
		// not prove what was missed, so leave recovery to scoped HTTP reads.
		metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("truncated").Inc()
		return nil, nil
	}

	// Replay is a finite snapshot for an already-connected WebSocket.
	// XREAD without BLOCK can still wait at a Stream tail in Redis-compatible
	// implementations, so use XRANGE's exclusive lower bound instead.
	messages, err := r.client.XRangeN(ctx, r.config.Stream, "("+afterID, "+", int64(limit)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("empty").Inc()
			return nil, nil
		}
		metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("redis_failed").Inc()
		return nil, fmt.Errorf("replay customer-service realtime events: %w", err)
	}

	events := r.eventsFromMessages(messages, "replay")
	metrics.CustomerServiceRealtimeReplayRequests.WithLabelValues("served").Inc()
	return events, nil
}

func (r *CustomerServiceRealtimeRelay) firstID(ctx context.Context) (string, error) {
	messages, err := r.client.XRangeN(ctx, r.config.Stream, "-", "+", 1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("read customer-service realtime stream head: %w", err)
	}
	if len(messages) == 0 {
		return "", nil
	}
	if !isCustomerServiceRealtimeStreamID(messages[0].ID) {
		return "", fmt.Errorf("redis returned invalid customer-service stream head %q", messages[0].ID)
	}
	return messages[0].ID, nil
}

func (r *CustomerServiceRealtimeRelay) currentTail(ctx context.Context) (string, error) {
	messages, err := r.client.XRevRangeN(ctx, r.config.Stream, "+", "-", 1).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("read customer-service realtime stream tail: %w", err)
	}
	if len(messages) == 0 {
		return "0-0", nil
	}
	if !isCustomerServiceRealtimeStreamID(messages[0].ID) {
		return "", fmt.Errorf("redis returned invalid customer-service stream tail %q", messages[0].ID)
	}
	return messages[0].ID, nil
}

func (r *CustomerServiceRealtimeRelay) consume(ctx context.Context, cursor string, done chan<- struct{}) {
	defer close(done)

	for {
		streams, err := r.client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{r.config.Stream, cursor},
			Count:   customerServiceRealtimeRelayReadBatchLimit,
			Block:   r.config.ConsumerBlock,
		}).Result()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, redis.Nil) {
				metrics.CustomerServiceRealtimeRelayReads.WithLabelValues("empty").Inc()
				continue
			}
			metrics.CustomerServiceRealtimeRelayReads.WithLabelValues("redis_failed").Inc()
			appLogger.Warn("customer-service realtime relay read failed", zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(customerServiceRealtimeRelayRetryDelay):
			}
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				cursor = message.ID
				event, err := customerServiceRealtimeEventFromStreamMessage(message)
				if err != nil {
					metrics.CustomerServiceRealtimeRelayEvents.WithLabelValues("live", "invalid").Inc()
					appLogger.Warn("customer-service realtime relay skipped invalid stream event", zap.Error(err), zap.String("stream_id", message.ID))
					continue
				}
				metrics.CustomerServiceRealtimeRelayEvents.WithLabelValues("live", "delivered").Inc()
				r.hub.Publish(event)
			}
		}
	}
}

func (r *CustomerServiceRealtimeRelay) eventsFromMessages(messages []redis.XMessage, path string) []CustomerServiceRealtimeEvent {
	events := make([]CustomerServiceRealtimeEvent, 0, len(messages))
	for _, message := range messages {
		event, err := customerServiceRealtimeEventFromStreamMessage(message)
		if err != nil {
			metrics.CustomerServiceRealtimeRelayEvents.WithLabelValues(path, "invalid").Inc()
			appLogger.Warn("customer-service realtime replay skipped invalid stream event", zap.Error(err), zap.String("stream_id", message.ID))
			continue
		}
		metrics.CustomerServiceRealtimeRelayEvents.WithLabelValues(path, "delivered").Inc()
		events = append(events, event)
	}
	return events
}

func (r *CustomerServiceRealtimeRelay) dedupKey(eventID string) string {
	sum := sha256.Sum256([]byte(eventID))
	return customerServiceRealtimeRedisHashTagPrefix(r.config.Stream) + customerServiceRealtimeRelayDedupKeyNamespace + hex.EncodeToString(sum[:])
}

func customerServiceRealtimeEventFromStreamMessage(message redis.XMessage) (CustomerServiceRealtimeEvent, error) {
	if !isCustomerServiceRealtimeStreamID(message.ID) {
		return CustomerServiceRealtimeEvent{}, fmt.Errorf("invalid stream id %q", message.ID)
	}
	raw, ok := message.Values["event"]
	if !ok {
		return CustomerServiceRealtimeEvent{}, errors.New("stream event payload is missing")
	}
	value, ok := raw.(string)
	if !ok {
		return CustomerServiceRealtimeEvent{}, errors.New("stream event payload has an unexpected type")
	}

	var event CustomerServiceRealtimeEvent
	if err := json.Unmarshal([]byte(value), &event); err != nil {
		return CustomerServiceRealtimeEvent{}, fmt.Errorf("decode stream event payload: %w", err)
	}
	if err := validateCustomerServiceRealtimeEvent(event); err != nil {
		return CustomerServiceRealtimeEvent{}, err
	}
	event.StreamID = message.ID
	return event, nil
}

func isCustomerServiceRealtimeStreamID(value string) bool {
	_, _, ok := parseCustomerServiceRealtimeStreamID(value)
	return ok
}

func compareCustomerServiceRealtimeStreamIDs(left, right string) int {
	leftMilliseconds, leftSequence, leftOK := parseCustomerServiceRealtimeStreamID(left)
	rightMilliseconds, rightSequence, rightOK := parseCustomerServiceRealtimeStreamID(right)
	if !leftOK || !rightOK {
		return 0
	}
	if leftMilliseconds < rightMilliseconds {
		return -1
	}
	if leftMilliseconds > rightMilliseconds {
		return 1
	}
	if leftSequence < rightSequence {
		return -1
	}
	if leftSequence > rightSequence {
		return 1
	}
	return 0
}

func parseCustomerServiceRealtimeStreamID(value string) (uint64, uint64, bool) {
	parts := strings.Split(value, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, false
	}
	milliseconds, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	sequence, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return milliseconds, sequence, true
}

func hasCustomerServiceRealtimeRedisHashTag(stream string) bool {
	start := strings.IndexByte(stream, '{')
	if start < 0 {
		return false
	}
	endOffset := strings.IndexByte(stream[start+1:], '}')
	return endOffset > 0
}

func customerServiceRealtimeRedisHashTagPrefix(stream string) string {
	start := strings.IndexByte(stream, '{')
	end := start + 1 + strings.IndexByte(stream[start+1:], '}')
	return stream[:end+1]
}
