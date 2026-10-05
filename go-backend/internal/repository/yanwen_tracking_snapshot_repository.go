package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/shipping"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// YanwenTrackingSnapshotRepository owns the latest official tracking facts
// inside the Yanwen domain. It never reads or writes generic tracking_events.
type YanwenTrackingSnapshotRepository struct{ db *gorm.DB }

func NewYanwenTrackingSnapshotRepository(db *gorm.DB) *YanwenTrackingSnapshotRepository {
	return &YanwenTrackingSnapshotRepository{db: db}
}

func (r *YanwenTrackingSnapshotRepository) UpsertYanwenTrackingSnapshots(snapshots []*shipping.YanwenTrackingSnapshot) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if len(snapshots) == 0 {
		return errors.New("at least one Yanwen tracking snapshot is required")
	}
	for _, snapshot := range snapshots {
		if err := snapshot.Validate(); err != nil {
			return err
		}
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, snapshot := range snapshots {
			updatedAt := snapshot.UpdatedAt
			if updatedAt.IsZero() {
				updatedAt = snapshot.LastSyncedAt
			}
			var existing shipping.YanwenTrackingSnapshot
			err := tx.Where(
				"environment = ? AND tracking_number = ?",
				snapshot.Environment,
				snapshot.TrackingNumber,
			).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(snapshot).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}
			updates := map[string]interface{}{
				"waybill_number":               snapshot.WaybillNumber,
				"exchange_number":              snapshot.ExchangeNumber,
				"last_mile_carrier":            snapshot.LastMileCarrier,
				"last_mile_carrier_website":    snapshot.LastMileCarrierWebsite,
				"last_mile_carrier_contact":    snapshot.LastMileCarrierContact,
				"tracking_status":              snapshot.TrackingStatus,
				"tracking_status_level1":       snapshot.TrackingStatusLevel1,
				"tracking_status_level2":       snapshot.TrackingStatusLevel2,
				"tracking_status_level3":       snapshot.TrackingStatusLevel3,
				"last_mile_tracking_expected":  snapshot.LastMileTrackingExpected,
				"origin_country":               snapshot.OriginCountry,
				"destination_country":          snapshot.DestinationCountry,
				"latest_checkpoint_status":     snapshot.LatestCheckpointStatus,
				"latest_checkpoint_time_stamp": snapshot.LatestCheckpointTimeStamp,
				"checkpoints_data":             snapshot.CheckpointsData,
				"response_data":                snapshot.ResponseData,
				"has_official_result":          true,
				"polling_state":                snapshot.PollingState,
				"next_sync_at":                 snapshot.NextSyncAt,
				"last_synced_at":               snapshot.LastSyncedAt.UTC(),
				"updated_at":                   updatedAt.UTC(),
			}
			if err := tx.Model(&shipping.YanwenTrackingSnapshot{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RegisterProductionTrackingNumbers creates internal polling targets without
// inventing an official status or checkpoint. A target becomes an official
// snapshot only after the tracking endpoint returns a validated result.
func (r *YanwenTrackingSnapshotRepository) RegisterProductionTrackingNumbers(ctx context.Context, trackingNumbers []string, nextSyncAt time.Time) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	normalized := make([]string, 0, len(trackingNumbers))
	seen := make(map[string]struct{}, len(trackingNumbers))
	for _, value := range trackingNumbers {
		number := strings.TrimSpace(value)
		if number == "" {
			continue
		}
		if _, exists := seen[number]; exists {
			continue
		}
		seen[number] = struct{}{}
		normalized = append(normalized, number)
	}
	if len(normalized) == 0 {
		return nil
	}
	if nextSyncAt.IsZero() {
		nextSyncAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, number := range normalized {
			target := &shipping.YanwenTrackingSnapshot{
				Environment:       "production",
				TrackingNumber:    number,
				WaybillNumber:     number,
				PollingState:      shipping.YanwenTrackingPollingStatePending,
				NextSyncAt:        &nextSyncAt,
				CheckpointsData:   []byte("[]"),
				ResponseData:      []byte("{}"),
				HasOfficialResult: false,
			}
			if err := target.Validate(); err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(target).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ClaimDueProductionTrackingSnapshots leases due production targets so a
// crashed worker can be recovered after the lease expires and two workers do
// not query the same target concurrently.
func (r *YanwenTrackingSnapshotRepository) ClaimDueProductionTrackingSnapshots(
	ctx context.Context,
	now time.Time,
	leaseUntil time.Time,
	limit int,
	leaseToken string,
) ([]shipping.YanwenTrackingSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if now.IsZero() || leaseUntil.IsZero() || leaseUntil.Before(now) {
		return nil, errors.New("valid Yanwen tracking polling lease times are required")
	}
	if limit <= 0 {
		return nil, errors.New("Yanwen tracking polling claim limit must be positive")
	}
	if strings.TrimSpace(leaseToken) == "" {
		return nil, errors.New("Yanwen tracking polling lease token is required")
	}
	claimed := make([]shipping.YanwenTrackingSnapshot, 0, limit)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := tx.Where(
			"environment = ? AND polling_state <> ? AND (next_sync_at IS NULL OR next_sync_at <= ?) AND (polling_lease_until IS NULL OR polling_lease_until <= ?)",
			"production",
			shipping.YanwenTrackingPollingStateTerminal,
			now.UTC(),
			now.UTC(),
		).Order("COALESCE(next_sync_at, created_at) ASC").Order("id ASC").Limit(limit)
		if tx.Dialector.Name() == "postgres" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		var candidates []shipping.YanwenTrackingSnapshot
		if err := query.Find(&candidates).Error; err != nil {
			return err
		}
		for index := range candidates {
			candidate := &candidates[index]
			result := tx.Model(&shipping.YanwenTrackingSnapshot{}).
				Where("id = ?", candidate.ID).
				Updates(map[string]interface{}{
					"polling_lease_token":     leaseToken,
					"polling_lease_until":     leaseUntil.UTC(),
					"last_polling_attempt_at": now.UTC(),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				continue
			}
			candidate.PollingLeaseToken = leaseToken
			candidate.PollingLeaseUntil = &leaseUntil
			candidate.LastPollingAttemptAt = &now
			claimed = append(claimed, *candidate)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (r *YanwenTrackingSnapshotRepository) MarkProductionTrackingPollingSucceeded(
	ctx context.Context,
	trackingNumber string,
	leaseToken string,
	now time.Time,
	nextSyncAt *time.Time,
	pollingState string,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	leaseToken = strings.TrimSpace(leaseToken)
	if trackingNumber == "" || leaseToken == "" || now.IsZero() {
		return errors.New("Yanwen tracking polling success identifiers are required")
	}
	if pollingState != shipping.YanwenTrackingPollingStateActive && pollingState != shipping.YanwenTrackingPollingStateTerminal {
		return errors.New("Yanwen tracking polling success state is invalid")
	}
	updates := map[string]interface{}{
		"polling_state":         pollingState,
		"next_sync_at":          nextSyncAt,
		"polling_lease_token":   "",
		"polling_lease_until":   nil,
		"polling_failure_count": 0,
		"last_polling_error":    "",
		"updated_at":            now.UTC(),
	}
	result := r.db.WithContext(ctx).Model(&shipping.YanwenTrackingSnapshot{}).
		Where("environment = ? AND tracking_number = ? AND polling_lease_token = ?", "production", trackingNumber, leaseToken).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *YanwenTrackingSnapshotRepository) MarkProductionTrackingPollingFailed(
	ctx context.Context,
	trackingNumber string,
	leaseToken string,
	now time.Time,
	nextSyncAt time.Time,
	errMessage string,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	trackingNumber = strings.TrimSpace(trackingNumber)
	leaseToken = strings.TrimSpace(leaseToken)
	errMessage = strings.TrimSpace(errMessage)
	if trackingNumber == "" || leaseToken == "" || now.IsZero() || nextSyncAt.IsZero() || errMessage == "" {
		return errors.New("Yanwen tracking polling failure details are required")
	}
	if len(errMessage) > 1000 {
		errMessage = errMessage[:1000]
	}
	result := r.db.WithContext(ctx).Model(&shipping.YanwenTrackingSnapshot{}).
		Where("environment = ? AND tracking_number = ? AND polling_lease_token = ?", "production", trackingNumber, leaseToken).
		Updates(map[string]interface{}{
			"polling_state": gorm.Expr(
				"CASE WHEN has_official_result = ? THEN ? ELSE ? END",
				true,
				shipping.YanwenTrackingPollingStateActive,
				shipping.YanwenTrackingPollingStatePending,
			),
			"next_sync_at":          nextSyncAt.UTC(),
			"polling_lease_token":   "",
			"polling_lease_until":   nil,
			"polling_failure_count": gorm.Expr("polling_failure_count + 1"),
			"last_polling_error":    errMessage,
			"updated_at":            now.UTC(),
		}).Error
	if result != nil {
		return result
	}
	return nil
}

func (r *YanwenTrackingSnapshotRepository) FindYanwenTrackingSnapshotByEnvironmentAndTrackingNumber(environment, trackingNumber string) (*shipping.YanwenTrackingSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	var snapshot shipping.YanwenTrackingSnapshot
	err := r.db.Where(
		"environment = ? AND tracking_number = ?",
		strings.ToLower(strings.TrimSpace(environment)),
		strings.TrimSpace(trackingNumber),
	).First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *YanwenTrackingSnapshotRepository) FindYanwenTrackingSnapshotsByEnvironment(environment string) ([]shipping.YanwenTrackingSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	snapshots := make([]shipping.YanwenTrackingSnapshot, 0)
	// Polling targets without an official response are internal worker state;
	// callers listing snapshots must only receive real Yanwen facts.
	query := r.db.Where("has_official_result = ?", true).Order("last_synced_at DESC").Order("id DESC")
	if environment = strings.ToLower(strings.TrimSpace(environment)); environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if err := query.Find(&snapshots).Error; err != nil {
		return nil, err
	}
	return snapshots, nil
}
