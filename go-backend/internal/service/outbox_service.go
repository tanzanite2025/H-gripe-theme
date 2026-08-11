package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

const (
	defaultOutboxBatchLimit       = 100
	defaultOutboxLockTimeout      = 5 * time.Minute
	defaultOutboxRetryBaseDelay   = 30 * time.Second
	defaultOutboxRetryMaxDelay    = 15 * time.Minute
	maxOutboxDispatchErrorMessage = 2000
)

type OutboxEventHandler func(ctx context.Context, event outbox.Event) error

type OutboxDispatchResult struct {
	Claimed    int    `json:"claimed"`
	Processed  int    `json:"processed"`
	Failed     int    `json:"failed"`
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
		workerID:    fmt.Sprintf("outbox-%d-%d", os.Getpid(), time.Now().UnixNano()),
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
			continue
		}

		if err := handler(ctx, event); err != nil {
			if markErr := s.markDispatchFailure(event, err, now); markErr != nil {
				return result, markErr
			}
			if event.Attempts >= event.MaxAttempts {
				result.DeadLetter++
			} else {
				result.Failed++
			}
			continue
		}

		if err := s.repo.MarkProcessed(event.ID, now); err != nil {
			return result, err
		}
		result.Processed++
	}

	return result, nil
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
	return s.repo.MarkFailed(event.ID, status, truncateOutboxError(dispatchErr.Error()), nextRetryAt, now)
}

func outboxRetryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return defaultOutboxRetryBaseDelay
	}
	multiplier := math.Pow(2, float64(attempt-1))
	delay := time.Duration(multiplier) * defaultOutboxRetryBaseDelay
	if delay > defaultOutboxRetryMaxDelay {
		return defaultOutboxRetryMaxDelay
	}
	return delay
}

func truncateOutboxError(value string) string {
	if len(value) <= maxOutboxDispatchErrorMessage {
		return value
	}
	return value[:maxOutboxDispatchErrorMessage]
}
