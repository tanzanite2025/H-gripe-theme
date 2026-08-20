package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"commerce-platform/internal/domain/outbox"
	appLogger "commerce-platform/internal/pkg/logger"
	"commerce-platform/internal/pkg/metrics"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultOutboxBatchLimit            = 100
	defaultOutboxLockTimeout           = 5 * time.Minute
	defaultOutboxRetryBaseDelay        = 30 * time.Second
	defaultOutboxRetryMaxDelay         = 15 * time.Minute
	defaultOutboxRetryJitter           = 15 * time.Second
	defaultOutboxUnknownReconcileDelay = 15 * time.Minute
	maxOutboxDispatchErrorMessage      = 2000
)

var outboxRetryJitterInt63n = rand.Int63n

type OutboxEventHandler func(ctx context.Context, event outbox.Event) error

type OutboxDispatchResult struct {
	Claimed    int    `json:"claimed"`
	Processed  int    `json:"processed"`
	Failed     int    `json:"failed"`
	Unknown    int    `json:"unknown"`
	DeadLetter int    `json:"dead_letter"`
	Unhandled  int    `json:"unhandled"`
	BatchLimit int    `json:"batch_limit"`
	WorkerID   string `json:"worker_id"`
}

type OutboxService struct {
	repo        *repository.OutboxRepository
	workerID    string
	handlers    map[string]OutboxEventHandler
	lockTimeout time.Duration
}

func NewOutboxService(repo *repository.OutboxRepository) *OutboxService {
	return &OutboxService{
		repo:        repo,
		workerID:    fmt.Sprintf("outbox-%s", uuid.NewString()),
		handlers:    map[string]OutboxEventHandler{},
		lockTimeout: defaultOutboxLockTimeout,
	}
}

func (s *OutboxService) RegisterHandler(eventType string, handler OutboxEventHandler) {
	if s == nil || eventType == "" || handler == nil {
		return
	}
	s.handlers[eventType] = handler
}

func (s *OutboxService) HandlerCount() int {
	if s == nil {
		return 0
	}
	return len(s.handlers)
}

func (s *OutboxService) ConfigureLockTimeout(timeout time.Duration) {
	if s == nil || timeout <= 0 {
		return
	}
	s.lockTimeout = timeout
}

func (s *OutboxService) ProcessPending(ctx context.Context, now time.Time, limit int) (OutboxDispatchResult, error) {
	result := OutboxDispatchResult{}
	if s == nil || s.repo == nil {
		return result, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 {
		limit = defaultOutboxBatchLimit
	}
	result.BatchLimit = limit
	result.WorkerID = s.workerID

	events, err := s.repo.ClaimReadyEvents(now, s.workerID, limit, s.lockTimeout)
	if err != nil {
		return result, err
	}
	result.Claimed = len(events)

	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		handler := s.handlers[event.EventType]
		if handler == nil {
			result.Unhandled++
			if markErr := s.markDispatchFailure(event, fmt.Errorf("no handler registered for outbox event type %s", event.EventType), now); markErr != nil {
				return result, markErr
			}
			if event.Attempts >= event.MaxAttempts {
				result.DeadLetter++
			} else {
				result.Failed++
			}
			s.recordCustomerServiceRealtimeOutboxOutcome(event, "unhandled")
			continue
		}

		if err := s.handleEventWithLeaseHeartbeat(ctx, event, handler); err != nil {
			if errors.Is(err, repository.ErrOutboxOwnershipLost) {
				return result, err
			}
			if errors.Is(err, resilience.ErrExternalOutcomeUnknown) {
				if markErr := s.markDispatchUnknown(event, err, now); markErr != nil {
					return result, markErr
				}
				result.Unknown++
				s.recordCustomerServiceRealtimeOutboxOutcome(event, "unknown")
				continue
			}
			if ctxErr := ctx.Err(); ctxErr != nil {
				return result, ctxErr
			}
			if markErr := s.markDispatchFailure(event, err, now); markErr != nil {
				return result, markErr
			}
			if event.Attempts >= event.MaxAttempts {
				result.DeadLetter++
			} else {
				result.Failed++
			}
			s.recordCustomerServiceRealtimeOutboxOutcome(event, "failed")
			continue
		}

		if err := s.repo.MarkProcessedByWorker(event.ID, s.workerID, now); err != nil {
			return result, err
		}
		result.Processed++
		s.recordCustomerServiceRealtimeOutboxOutcome(event, "processed")
	}

	s.refreshCustomerServiceRealtimeOutboxMetrics()
	return result, nil
}

func (s *OutboxService) recordCustomerServiceRealtimeOutboxOutcome(event outbox.Event, result string) {
	if event.EventType != outbox.EventTypeCustomerServiceRealtime {
		return
	}
	metrics.CustomerServiceRealtimeOutboxDeliveries.WithLabelValues(result).Inc()
}

// RefreshCustomerServiceRealtimeMetrics exposes only aggregate durable state
// for the customer-service event type. The scheduler calls this after every
// dispatch pass; it is also useful at startup before the first event arrives.
func (s *OutboxService) RefreshCustomerServiceRealtimeMetrics() error {
	if s == nil || s.repo == nil {
		return nil
	}
	counts, err := s.repo.CountEventsByStatus(outbox.EventTypeCustomerServiceRealtime)
	if err != nil {
		return err
	}
	for _, status := range []string{
		outbox.EventStatusPending,
		outbox.EventStatusProcessing,
		outbox.EventStatusFailed,
		outbox.EventStatusUnknown,
		outbox.EventStatusDeadLetter,
	} {
		metrics.CustomerServiceRealtimeOutboxEvents.WithLabelValues(status).Set(float64(counts[status]))
	}
	return nil
}

func (s *OutboxService) ListUnknownEvents(limit int) ([]outbox.Event, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	return s.repo.FindUnknownEvents(limit)
}

// ResumeUnknownEvent is an explicit reconciliation decision. It is not called
// by the normal worker because an unknown external result must be verified
// before the side effect is attempted again.
func (s *OutboxService) ResumeUnknownEvent(id uint, nextAvailableAt time.Time, note string, resumedAt time.Time) error {
	if s == nil || s.repo == nil {
		return repository.ErrOutboxUnknownNotFound
	}
	return s.repo.ResumeUnknownEvent(id, nextAvailableAt, note, resumedAt)
}

// MarkUnknownEventProcessed records that reconciliation confirmed the external
// side effect already happened, so the local event must not be sent again.
func (s *OutboxService) MarkUnknownEventProcessed(id uint, note string, processedAt time.Time) error {
	if s == nil || s.repo == nil {
		return repository.ErrOutboxUnknownNotFound
	}
	return s.repo.MarkUnknownProcessed(id, note, processedAt)
}

func (s *OutboxService) refreshCustomerServiceRealtimeOutboxMetrics() {
	if err := s.RefreshCustomerServiceRealtimeMetrics(); err != nil {
		// Processing must not become unavailable merely because a monitoring
		// aggregate query failed. The next scheduler pass retries the aggregate.
		appLogger.Warn("customer-service realtime outbox metric refresh failed", zap.Error(err))
	}
}

func (s *OutboxService) markDispatchFailure(event outbox.Event, dispatchErr error, now time.Time) error {
	if dispatchErr == nil {
		dispatchErr = errors.New("outbox dispatch failed")
	}
	status := outbox.EventStatusFailed
	if event.Attempts >= event.MaxAttempts {
		status = outbox.EventStatusDeadLetter
	}
	nextRetryAt := now
	if status != outbox.EventStatusDeadLetter {
		nextRetryAt = now.Add(outboxRetryDelay(event.Attempts))
	}
	return s.repo.MarkFailedByWorker(
		event.ID,
		s.workerID,
		status,
		truncateOutboxError(dispatchErr.Error()),
		nextRetryAt,
		now,
	)
}

func (s *OutboxService) markDispatchUnknown(event outbox.Event, dispatchErr error, now time.Time) error {
	if dispatchErr == nil {
		dispatchErr = resilience.ErrExternalOutcomeUnknown
	}
	return s.repo.MarkUnknownByWorker(
		event.ID,
		s.workerID,
		truncateOutboxError(dispatchErr.Error()),
		now.Add(defaultOutboxUnknownReconcileDelay),
		now,
	)
}

func (s *OutboxService) handleEventWithLeaseHeartbeat(
	ctx context.Context,
	event outbox.Event,
	handler OutboxEventHandler,
) error {
	if s == nil || s.repo == nil || handler == nil {
		return nil
	}
	interval := outboxLeaseHeartbeatInterval(s.lockTimeout)
	if interval <= 0 {
		return handler(ctx, event)
	}

	handlerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stop := make(chan struct{})
	done := make(chan struct{})
	heartbeatErr := make(chan error, 1)
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-handlerCtx.Done():
				return
			case <-stop:
				return
			case tick := <-ticker.C:
				if err := s.repo.RefreshProcessingLockByWorker(event.ID, s.workerID, tick.UTC()); err != nil {
					// Once the lease heartbeat is no longer reliable, the
					// external outcome cannot be proven safe to replay.
					heartbeatErr <- errors.Join(err, resilience.ErrExternalOutcomeUnknown)
					cancel()
					return
				}
			}
		}
	}()

	err := handler(handlerCtx, event)
	close(stop)
	<-done

	select {
	case leaseErr := <-heartbeatErr:
		if leaseErr != nil {
			return leaseErr
		}
	default:
	}
	return err
}

func outboxLeaseHeartbeatInterval(lockTimeout time.Duration) time.Duration {
	if lockTimeout <= 0 {
		return 0
	}
	interval := lockTimeout / 3
	if interval > 30*time.Second {
		interval = 30 * time.Second
	}
	if interval <= 0 {
		interval = lockTimeout / 2
	}
	if interval <= 0 {
		return 0
	}
	return interval
}

func outboxRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(multiplier) * defaultOutboxRetryBaseDelay
	if delay > defaultOutboxRetryMaxDelay {
		delay = defaultOutboxRetryMaxDelay
	}
	return delay + outboxRetryJitter(delay)
}

func outboxRetryJitter(delay time.Duration) time.Duration {
	if defaultOutboxRetryJitter <= 0 || delay >= defaultOutboxRetryMaxDelay {
		return 0
	}
	jitterLimit := defaultOutboxRetryJitter
	if remaining := defaultOutboxRetryMaxDelay - delay; remaining < jitterLimit {
		jitterLimit = remaining
	}
	if jitterLimit <= 0 {
		return 0
	}
	return time.Duration(outboxRetryJitterInt63n(int64(jitterLimit)))
}

func truncateOutboxError(value string) string {
	if len(value) <= maxOutboxDispatchErrorMessage {
		return value
	}
	return value[:maxOutboxDispatchErrorMessage]
}
