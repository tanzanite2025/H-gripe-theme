package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"
)

const (
	defaultYanwenTrackingPollingBatchLimit = 20
	maximumYanwenTrackingPollingBatchLimit = 30
	defaultYanwenTrackingPollingLease      = 15 * time.Minute
	maximumYanwenTrackingRetryDelay        = 6 * time.Hour
)

// YanwenTrackingQueryService is the narrow tracking dependency required by the
// production polling worker. It deliberately excludes catalog, waybill, and
// configuration mutation methods.
type YanwenTrackingQueryService interface {
	QueryYanwenTracking(context.Context, []string) ([]YanwenTrackingResult, error)
}

type yanwenTrackingSnapshotBoundService interface {
	YanwenTrackingQueryService
	yanwenTrackingSnapshotRepository() *repository.YanwenTrackingSnapshotRepository
}

// YanwenTrackingPollingService owns the production-only background polling
// workflow. It depends on the narrow tracking service and the Yanwen-owned
// snapshot repository, never on generic tracking repositories.
type YanwenTrackingPollingService struct {
	trackingService YanwenTrackingQueryService
	snapshots       *repository.YanwenTrackingSnapshotRepository
	interval        time.Duration
	batchLimit      int
	leaseDuration   time.Duration
}

type YanwenTrackingPollingFailure struct {
	TrackingNumber string `json:"tracking_number"`
	Error          string `json:"error"`
}

type YanwenTrackingPollingBatchResult struct {
	Claimed  int                            `json:"claimed"`
	Synced   int                            `json:"synced"`
	Terminal int                            `json:"terminal"`
	Failed   int                            `json:"failed"`
	Errors   []YanwenTrackingPollingFailure `json:"errors,omitempty"`
}

func NewYanwenTrackingPollingService(
	trackingService YanwenTrackingQueryService,
	snapshots *repository.YanwenTrackingSnapshotRepository,
	interval time.Duration,
	batchLimit int,
) *YanwenTrackingPollingService {
	if interval <= 0 {
		interval = defaultYanwenTrackingPollingInterval
	}
	if batchLimit <= 0 {
		batchLimit = defaultYanwenTrackingPollingBatchLimit
	}
	if batchLimit > maximumYanwenTrackingPollingBatchLimit {
		batchLimit = maximumYanwenTrackingPollingBatchLimit
	}
	leaseDuration := 2 * interval
	if leaseDuration < defaultYanwenTrackingPollingLease {
		leaseDuration = defaultYanwenTrackingPollingLease
	}
	return &YanwenTrackingPollingService{
		trackingService: trackingService,
		snapshots:       snapshots,
		interval:        interval,
		batchLimit:      batchLimit,
		leaseDuration:   leaseDuration,
	}
}

func (s *YanwenTrackingPollingService) PollDueYanwenTrackingTargets(ctx context.Context) (YanwenTrackingPollingBatchResult, error) {
	if s == nil || s.trackingService == nil {
		return YanwenTrackingPollingBatchResult{}, errors.New("Yanwen tracking service is not configured")
	}
	if s.snapshots == nil {
		return YanwenTrackingPollingBatchResult{}, errors.New("Yanwen tracking snapshot repository is not configured")
	}
	if boundService, ok := s.trackingService.(yanwenTrackingSnapshotBoundService); ok && boundService.yanwenTrackingSnapshotRepository() != s.snapshots {
		return YanwenTrackingPollingBatchResult{}, errors.New("Yanwen tracking service is not connected to the polling snapshot repository")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	now := time.Now().UTC()
	leaseToken, err := newYanwenTrackingPollingLeaseToken()
	if err != nil {
		return YanwenTrackingPollingBatchResult{}, err
	}
	claimed, err := s.snapshots.ClaimDueProductionTrackingSnapshots(
		ctx,
		now,
		now.Add(s.leaseDuration),
		s.batchLimit,
		leaseToken,
	)
	if err != nil {
		return YanwenTrackingPollingBatchResult{}, fmt.Errorf("claim Yanwen tracking polling targets: %w", err)
	}
	result := YanwenTrackingPollingBatchResult{Claimed: len(claimed), Errors: make([]YanwenTrackingPollingFailure, 0)}
	if len(claimed) == 0 {
		return result, nil
	}

	trackingNumbers := make([]string, 0, len(claimed))
	for _, target := range claimed {
		trackingNumbers = append(trackingNumbers, target.TrackingNumber)
	}
	officialResults, queryErr := s.trackingService.QueryYanwenTracking(ctx, trackingNumbers)
	if queryErr != nil {
		for _, target := range claimed {
			s.recordYanwenTrackingPollingFailure(ctx, &result, target, leaseToken, now, queryErr.Error())
		}
		return result, fmt.Errorf("query Yanwen tracking batch: %w", queryErr)
	}

	byTrackingNumber := make(map[string]YanwenTrackingResult, len(officialResults))
	for _, officialResult := range officialResults {
		byTrackingNumber[strings.TrimSpace(officialResult.TrackingNumber)] = officialResult
	}
	for _, target := range claimed {
		officialResult, found := byTrackingNumber[target.TrackingNumber]
		if !found {
			s.recordYanwenTrackingPollingFailure(
				ctx,
				&result,
				target,
				leaseToken,
				now,
				"official tracking response did not include this tracking number",
			)
			continue
		}
		pollingState := shipping.YanwenTrackingPollingStateActive
		var nextSyncAt *time.Time
		if shipping.IsYanwenTrackingTerminalStatus(officialResult.TrackingStatus) {
			pollingState = shipping.YanwenTrackingPollingStateTerminal
			result.Terminal++
		} else {
			next := now.Add(s.interval)
			nextSyncAt = &next
		}
		if err := s.snapshots.MarkProductionTrackingPollingSucceeded(
			ctx,
			target.TrackingNumber,
			leaseToken,
			now,
			nextSyncAt,
			pollingState,
		); err != nil {
			result.Errors = append(result.Errors, YanwenTrackingPollingFailure{
				TrackingNumber: target.TrackingNumber,
				Error:          fmt.Sprintf("save polling success state: %v", err),
			})
			result.Failed++
			continue
		}
		result.Synced++
	}
	if len(result.Errors) > 0 {
		return result, errors.New("one or more Yanwen tracking polling targets could not be finalized")
	}
	return result, nil
}

func (s *YanwenTrackingPollingService) recordYanwenTrackingPollingFailure(
	ctx context.Context,
	result *YanwenTrackingPollingBatchResult,
	target shipping.YanwenTrackingSnapshot,
	leaseToken string,
	now time.Time,
	failureMessage string,
) {
	if result == nil {
		return
	}
	retryAt := now.Add(calculateYanwenTrackingRetryDelay(s.interval, target.PollingFailureCount+1))
	if err := s.snapshots.MarkProductionTrackingPollingFailed(
		ctx,
		target.TrackingNumber,
		leaseToken,
		now,
		retryAt,
		failureMessage,
	); err != nil {
		failureMessage = fmt.Sprintf("%s; save failure state: %v", failureMessage, err)
	}
	result.Failed++
	result.Errors = append(result.Errors, YanwenTrackingPollingFailure{
		TrackingNumber: target.TrackingNumber,
		Error:          failureMessage,
	})
}

func calculateYanwenTrackingRetryDelay(interval time.Duration, failureCount int) time.Duration {
	if interval <= 0 {
		interval = defaultYanwenTrackingPollingInterval
	}
	if failureCount < 1 {
		failureCount = 1
	}
	delay := interval
	for attempt := 1; attempt < failureCount && delay < maximumYanwenTrackingRetryDelay; attempt++ {
		delay *= 2
	}
	if delay > maximumYanwenTrackingRetryDelay {
		return maximumYanwenTrackingRetryDelay
	}
	return delay
}

func newYanwenTrackingPollingLeaseToken() (string, error) {
	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", fmt.Errorf("generate Yanwen tracking polling lease token: %w", err)
	}
	return hex.EncodeToString(token[:]), nil
}
