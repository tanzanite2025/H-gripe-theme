package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteQualityTargetRepository struct {
	db *gorm.DB
}

var ErrSiteQualityTargetCanonicalConflict = errors.New("SiteQuality target canonical URL belongs to another route entry")

func NewSiteQualityTargetRepository(db *gorm.DB) *SiteQualityTargetRepository {
	return &SiteQualityTargetRepository{db: db}
}

func (r *SiteQualityTargetRepository) WithTx(tx *gorm.DB) *SiteQualityTargetRepository {
	return &SiteQualityTargetRepository{db: tx}
}

func (r *SiteQualityTargetRepository) Upsert(
	input sitequalitydomain.SiteQualityTargetInput,
	now time.Time,
) (*sitequalitydomain.SiteQualityTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality target repository is unavailable")
	}
	input.CanonicalURL = strings.TrimSpace(input.CanonicalURL)
	if input.CanonicalURL == "" {
		return nil, errors.New("SiteQuality quality target canonical URL is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if input.SamplingTier == "" {
		input.SamplingTier = sitequalitydomain.SiteQualityTargetTierStandard
	}
	if input.SamplingIntervalSeconds <= 0 {
		return nil, errors.New("SiteQuality quality target sampling interval must be positive")
	}
	input.Source = strings.TrimSpace(input.Source)
	if input.Source == "" {
		input.Source = sitequalitydomain.SiteQualityTargetSourceOperator
	}
	if input.Source != sitequalitydomain.SiteQualityTargetSourceRouteCatalog &&
		input.Source != sitequalitydomain.SiteQualityTargetSourceOperator {
		return nil, errors.New("SiteQuality quality target source is invalid")
	}
	if input.Source == sitequalitydomain.SiteQualityTargetSourceRouteCatalog {
		input.LedgerSynced = true
		if input.LedgerSyncedAt == nil {
			input.LedgerSyncedAt = &now
		}
	}

	var result sitequalitydomain.SiteQualityTarget
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var target sitequalitydomain.SiteQualityTarget
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"})
		var err error
		if input.RouteEntryID != nil && *input.RouteEntryID != 0 {
			err = query.Where("route_entry_id = ?", *input.RouteEntryID).First(&target).Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || input.RouteEntryID == nil || *input.RouteEntryID == 0 {
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("canonical_url = ?", input.CanonicalURL).
				First(&target).Error
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			target = sitequalitydomain.SiteQualityTarget{
				RouteEntryID:            input.RouteEntryID,
				CanonicalURL:            input.CanonicalURL,
				Locale:                  strings.TrimSpace(input.Locale),
				Source:                  input.Source,
				SourceType:              strings.TrimSpace(input.SourceType),
				Title:                   strings.TrimSpace(input.Title),
				SamplingTier:            input.SamplingTier,
				SamplingIntervalSeconds: input.SamplingIntervalSeconds,
				Enabled:                 input.Enabled,
				LedgerSynced:            input.LedgerSynced,
				LedgerSyncMarker:        strings.TrimSpace(input.LedgerSyncMarker),
				LedgerSyncedAt:          input.LedgerSyncedAt,
				DisableReason:           strings.TrimSpace(input.DisableReason),
				NextScheduledAt:         &now,
			}
			if err := tx.Select("*").Create(&target).Error; err != nil {
				return err
			}
			if !input.Enabled {
				if err := tx.Model(&target).Update("enabled", false).Error; err != nil {
					return err
				}
				target.Enabled = false
			}
			result = target
			return nil
		}
		if err != nil {
			return err
		}
		if input.LedgerSyncedAt != nil &&
			target.LedgerSyncedAt != nil &&
			input.LedgerSyncedAt.Before(*target.LedgerSyncedAt) {
			result = target
			return nil
		}
		var canonicalTarget sitequalitydomain.SiteQualityTarget
		canonicalErr := tx.
			Where("canonical_url = ?", input.CanonicalURL).
			First(&canonicalTarget).Error
		if canonicalErr == nil && canonicalTarget.ID != target.ID {
			return fmt.Errorf(
				"%w: %s",
				ErrSiteQualityTargetCanonicalConflict,
				input.CanonicalURL,
			)
		}
		if canonicalErr != nil && !errors.Is(canonicalErr, gorm.ErrRecordNotFound) {
			return canonicalErr
		}
		if target.RouteEntryID != nil &&
			input.RouteEntryID != nil &&
			*target.RouteEntryID != *input.RouteEntryID {
			return fmt.Errorf(
				"%w: %s",
				ErrSiteQualityTargetCanonicalConflict,
				input.CanonicalURL,
			)
		}

		updates := map[string]interface{}{
			"route_entry_id":            input.RouteEntryID,
			"canonical_url":             input.CanonicalURL,
			"locale":                    strings.TrimSpace(input.Locale),
			"source":                    input.Source,
			"source_type":               strings.TrimSpace(input.SourceType),
			"title":                     strings.TrimSpace(input.Title),
			"sampling_tier":             input.SamplingTier,
			"sampling_interval_seconds": input.SamplingIntervalSeconds,
			"enabled":                   input.Enabled,
			"ledger_synced":             input.LedgerSynced,
			"ledger_sync_marker":        strings.TrimSpace(input.LedgerSyncMarker),
			"ledger_synced_at":          input.LedgerSyncedAt,
			"disabled_at":               nil,
			"disable_reason":            strings.TrimSpace(input.DisableReason),
			"updated_at":                now,
		}
		if err := tx.Model(&target).Updates(updates).Error; err != nil {
			return err
		}
		return tx.First(&result, target.ID).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *SiteQualityTargetRepository) FindByID(id uint) (*sitequalitydomain.SiteQualityTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality target repository is unavailable")
	}
	if id == 0 {
		return nil, errors.New("SiteQuality quality target ID is required")
	}
	var target sitequalitydomain.SiteQualityTarget
	if err := r.db.First(&target, id).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *SiteQualityTargetRepository) FindByRouteEntryID(routeEntryID uint) (*sitequalitydomain.SiteQualityTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality target repository is unavailable")
	}
	if routeEntryID == 0 {
		return nil, errors.New("SiteQuality route entry ID is required")
	}
	var target sitequalitydomain.SiteQualityTarget
	if err := r.db.Where("route_entry_id = ?", routeEntryID).First(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *SiteQualityTargetRepository) FindByCanonicalURL(canonicalURL string) (*sitequalitydomain.SiteQualityTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality target repository is unavailable")
	}
	var target sitequalitydomain.SiteQualityTarget
	if err := r.db.Where("canonical_url = ?", strings.TrimSpace(canonicalURL)).First(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *SiteQualityTargetRepository) DisableByRouteEntryID(
	routeEntryID uint,
	reason string,
	marker string,
	now time.Time,
	ledgerSyncedAt ...time.Time,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality target repository is unavailable")
	}
	if routeEntryID == 0 {
		return errors.New("SiteQuality route entry ID is required")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	disabledReason := strings.TrimSpace(reason)
	if disabledReason == "" {
		disabledReason = "route catalog target is no longer active or checkable"
	}
	syncedAt := now
	if len(ledgerSyncedAt) > 0 && !ledgerSyncedAt[0].IsZero() {
		syncedAt = ledgerSyncedAt[0].UTC()
	}
	return r.db.Model(&sitequalitydomain.SiteQualityTarget{}).
		Where("route_entry_id = ? AND (ledger_synced_at IS NULL OR ledger_synced_at <= ?)", routeEntryID, syncedAt).
		Updates(map[string]interface{}{
			"enabled":            false,
			"ledger_synced":      true,
			"ledger_sync_marker": strings.TrimSpace(marker),
			"ledger_synced_at":   syncedAt,
			"disabled_at":        now,
			"disable_reason":     disabledReason,
			"updated_at":         now,
		}).Error
}

func (r *SiteQualityTargetRepository) ListDue(now time.Time, limit int) ([]sitequalitydomain.SiteQualityTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality quality target repository is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var targets []sitequalitydomain.SiteQualityTarget
	err := r.db.
		Where("enabled = ? AND (next_scheduled_at IS NULL OR next_scheduled_at <= ?)", true, now).
		Order("CASE sampling_tier WHEN 'critical' THEN 1 WHEN 'standard' THEN 2 ELSE 3 END").
		Order("next_scheduled_at ASC NULLS FIRST").
		Order("id ASC").
		Limit(limit).
		Find(&targets).Error
	return targets, err
}

func (r *SiteQualityTargetRepository) MarkScheduled(
	id uint,
	scheduledAt time.Time,
	nextScheduledAt time.Time,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality target repository is unavailable")
	}
	if id == 0 {
		return errors.New("SiteQuality quality target ID is required")
	}
	if scheduledAt.IsZero() {
		scheduledAt = time.Now().UTC()
	} else {
		scheduledAt = scheduledAt.UTC()
	}
	if nextScheduledAt.IsZero() {
		return errors.New("SiteQuality quality next scheduled time is required")
	}
	nextScheduledAt = nextScheduledAt.UTC()
	return r.db.Model(&sitequalitydomain.SiteQualityTarget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_scheduled_at": scheduledAt,
			"next_scheduled_at": nextScheduledAt,
			"updated_at":        scheduledAt,
		}).Error
}

func (r *SiteQualityTargetRepository) MarkCompleted(
	id uint,
	completedAt time.Time,
) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality quality target repository is unavailable")
	}
	if id == 0 {
		return errors.New("SiteQuality quality target ID is required")
	}
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	} else {
		completedAt = completedAt.UTC()
	}
	return r.db.Model(&sitequalitydomain.SiteQualityTarget{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_completed_at": completedAt,
			"updated_at":        completedAt,
		}).Error
}

func (r *SiteQualityTargetRepository) Stats(now time.Time) (SiteQualityTargetStats, error) {
	if r == nil || r.db == nil {
		return SiteQualityTargetStats{}, errors.New("SiteQuality quality target repository is unavailable")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}

	var stats SiteQualityTargetStats
	err := r.db.Model(&sitequalitydomain.SiteQualityTarget{}).
		Select(`
			COALESCE(COUNT(*), 0) AS total,
			COALESCE(SUM(CASE WHEN enabled THEN 1 ELSE 0 END), 0) AS enabled,
			COALESCE(SUM(CASE WHEN enabled AND (next_scheduled_at IS NULL OR next_scheduled_at <= ?) THEN 1 ELSE 0 END), 0) AS due,
			COALESCE(SUM(CASE WHEN sampling_tier = 'critical' THEN 1 ELSE 0 END), 0) AS critical,
			COALESCE(SUM(CASE WHEN sampling_tier = 'standard' THEN 1 ELSE 0 END), 0) AS standard,
			COALESCE(SUM(CASE WHEN sampling_tier = 'background' THEN 1 ELSE 0 END), 0) AS background
		`, now).
		Scan(&stats).Error
	return stats, err
}
