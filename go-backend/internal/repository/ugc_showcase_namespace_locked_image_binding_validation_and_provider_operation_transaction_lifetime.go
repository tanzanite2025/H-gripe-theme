package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUGCShowcaseImageOwnershipValidationUnavailable = errors.New("UGC showcase image ownership validation is unavailable")

func (r *UGCShowcaseRepository) RequireIndependentUGCShowcaseProviderOperationTransaction() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, active := r.db.Statement.ConnPool.(gorm.TxCommitter); active {
		return fmt.Errorf("UGC provider operation requires its own transaction lifetime")
	}
	return nil
}

func (r *UGCShowcaseRepository) WithContext(ctx context.Context) *UGCShowcaseRepository {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.WithTx(r.db.WithContext(ctx))
}

func (r *UGCShowcaseRepository) ConfigureTransactionalUGCShowcaseImageOwnershipValidation(validate func(context.Context, *ugcshowcase.UGCShowcase, *ugcshowcase.UGCShowcase) error, resolve func(string) (string, error)) {
	r.validateImageBindings, r.resolveImageObjectKey = validate, resolve
}

// Publication copies and physical cleanup retain this lock until synchronous
// provider IO returns, including cancellation. SQL and IO keep the request ctx.
func (r *UGCShowcaseRepository) WithLockedUGCShowcaseLifecycleUntilProviderOperationReturns(ctx context.Context, operation func(*UGCShowcaseRepository, *gorm.DB) error) error {
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, dedicatedUGCShowcaseLifecycleNamespace, func(tx *gorm.DB) error {
		return operation(r.WithTx(tx), tx)
	})
}

func (r *UGCShowcaseRepository) FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(id uint) (*ugcshowcase.UGCShowcase, error) {
	var item ugcshowcase.UGCShowcase
	query := r.db
	if query.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *UGCShowcaseRepository) ListUGCShowcaseImageReferenceScanPageAfterID(afterID uint) ([]ugcshowcase.UGCShowcase, error) {
	var items []ugcshowcase.UGCShowcase
	err := r.db.Select("id", "images", "status").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&items).Error
	return items, err
}

func (r *UGCShowcaseRepository) mutateUGCShowcaseImagesAndStatusWithNamespaceAndCurrentRowLocks(id uint, expectedStatus string, images *datatypes.JSON, status *string, reason string) (bool, error) {
	updated := false
	err := r.WithTransaction(func(repo *UGCShowcaseRepository, tx *gorm.DB) error {
		current, err := repo.FindUGCShowcaseByIDForUpdateWithinLifecycleTransaction(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if expectedStatus != "" && current.Status != expectedStatus {
			return nil
		}
		next := *current
		updates := map[string]any{}
		if images != nil {
			next.Images = *images
			updates["images"] = *images
		}
		if status != nil {
			next.Status, next.RejectedReason = *status, reason
			updates["status"], updates["rejected_reason"] = *status, reason
			if *status == ugcshowcase.StatusApproved {
				updates["approved_at"] = time.Now().UTC()
			}
		}
		if repo.validateImageBindings == nil {
			return ErrUGCShowcaseImageOwnershipValidationUnavailable
		}
		if err := repo.validateImageBindings(tx.Statement.Context, current, &next); err != nil {
			return err
		}
		result := tx.Model(&ugcshowcase.UGCShowcase{}).Where("id = ?", id).Updates(updates)
		updated = result.RowsAffected == 1
		return result.Error
	})
	return updated, err
}

func (r *UGCShowcaseRepository) listExactPendingUGCShowcaseCleanupCandidates(cutoff time.Time, limit int) ([]ugcshowcase.UGCShowcase, error) {
	var candidates []ugcshowcase.UGCShowcase
	if limit <= 0 {
		return candidates, nil
	}
	if r.resolveImageObjectKey == nil {
		return nil, ErrUGCShowcaseImageOwnershipValidationUnavailable
	}
	var afterCreatedAt *time.Time
	afterID := uint(0)
	for {
		var page []ugcshowcase.UGCShowcase
		query := r.db.Where("((status = ? AND created_at < ?) OR status = ?)", ugcshowcase.StatusPending, cutoff, ugcshowcase.StatusRejected)
		if afterCreatedAt != nil {
			query = query.Where("created_at > ? OR (created_at = ? AND id > ?)", *afterCreatedAt, *afterCreatedAt, afterID)
		}
		if err := query.Order("created_at ASC, id ASC").Limit(256).Find(&page).Error; err != nil {
			return nil, err
		}
		for _, item := range page {
			pending := item.Status == ugcshowcase.StatusPending
			if !pending {
				var references []string
				if len(strings.TrimSpace(string(item.Images))) > 0 {
					if err := json.Unmarshal(item.Images, &references); err != nil {
						return nil, err
					}
				}
				for _, reference := range references {
					key, err := r.resolveImageObjectKey(reference)
					if err == nil && storage.DedicatedPublicObjectNamespaceForStorageKey(key) == "showcase/pending" && strings.HasPrefix(key, "showcase/pending/") {
						pending = true
						break
					}
				}
			}
			if pending {
				candidates = append(candidates, item)
				if len(candidates) == limit {
					return candidates, nil
				}
			}
		}
		if len(page) < 256 {
			return candidates, nil
		}
		afterID = page[len(page)-1].ID
		afterCreatedAt = &page[len(page)-1].CreatedAt
	}
}
