package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrBusinessDraftMediaUploadReceiptUnavailable = errors.New("business draft media upload receipt is unavailable or expired")

func (r *MediaRepository) FindUnboundBusinessDraftMediaUploadExpiryForAsset(asset *media.MediaAsset) *time.Time {
	id, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(asset.StorageKey)
	if !draft || r == nil || r.db == nil {
		return nil
	}
	var receipt media.BusinessDraftMediaUploadReceipt
	if err := r.db.Select("expires_at").Where("id = ? AND asset_id = ? AND status = ?", id, asset.ID, "ready").Take(&receipt).Error; err != nil {
		return nil
	}
	return &receipt.ExpiresAt
}

func (r *MediaRepository) CreateBusinessDraftMediaUploadReceipt(ctx context.Context, receipt *media.BusinessDraftMediaUploadReceipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

func (r *MediaRepository) FindBusinessDraftMediaUploadReceiptForUpdate(id string) (*media.BusinessDraftMediaUploadReceipt, error) {
	if err := r.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction(nil, nil); err != nil {
		return nil, err
	}
	var receipt media.BusinessDraftMediaUploadReceipt
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&receipt, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *MediaRepository) RequireIndependentMediaProviderOperationTransaction() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("media provider operations require their own transaction lifetime")
	}
	return nil
}

func (r *MediaRepository) LockBusinessDraftMediaUploadObjectsWithinTransaction(receipt *media.BusinessDraftMediaUploadReceipt) error {
	keys, err := BusinessDraftMediaUploadReceiptObjectKeys(receipt)
	if err != nil {
		return err
	}
	return r.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction(keys, nil)
}

func BusinessDraftMediaUploadReceiptObjectKeys(receipt *media.BusinessDraftMediaUploadReceipt) ([]string, error) {
	var keys []string
	if receipt == nil || json.Unmarshal(receipt.ObjectKeysJSON, &keys) != nil || len(keys) == 0 || len(keys) > 13 || keys[0] != receipt.OriginalObjectKey {
		return nil, ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	seen := make(map[string]bool)
	for _, key := range keys {
		id, valid := storage.BusinessDraftMediaReceiptIDFromObjectKey(key)
		if !valid || id != receipt.ID || seen[key] {
			return nil, ErrBusinessDraftMediaUploadReceiptUnavailable
		}
		seen[key] = true
	}
	return keys, nil
}

func (r *MediaRepository) validateBusinessDraftMediaRegistrationReceiptWithinTransaction(receiptID string, asset *media.MediaAsset) error {
	receipt, err := r.FindBusinessDraftMediaUploadReceiptForUpdate(receiptID)
	if err != nil {
		return err
	}
	if receipt.Status != "uploading" || receipt.AssetID != nil || !receipt.ExpiresAt.After(time.Now().UTC()) ||
		asset.UploaderActorType != media.MediaUploadActorAdministrator || asset.UploaderID != receipt.UploaderID || asset.UploadOrigin != receipt.UploadOrigin ||
		asset.StorageKey != receipt.OriginalObjectKey || asset.ContentSHA256 != receipt.ContentSHA256 || asset.Size != receipt.Size || asset.MimeType != receipt.MimeType || asset.Visibility != "public" {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	keys, err := BusinessDraftMediaUploadReceiptObjectKeys(receipt)
	if err != nil {
		return err
	}
	if len(keys) != 1+len(asset.Derivatives) {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	if r.resolveRegisteredMediaObjectKeyForCleanupOwnership(asset.URL) != asset.StorageKey {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	for i, derivative := range asset.Derivatives {
		if derivative.StorageKey != keys[i+1] || r.resolveRegisteredMediaObjectKeyForCleanupOwnership(derivative.URL) != derivative.StorageKey {
			return ErrBusinessDraftMediaUploadReceiptUnavailable
		}
	}
	return nil
}

// The asset row is already locked by the common business binding plan. Its
// receipt update rolls back with the owning business write.
func (r *MediaRepository) BindBusinessDraftMediaUploadReceiptWithinTransaction(asset *media.MediaAsset) error {
	id, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(asset.StorageKey)
	if !draft {
		return nil
	}
	receipt, err := r.FindBusinessDraftMediaUploadReceiptForUpdate(id)
	if err != nil {
		return err
	}
	if receipt.AssetID == nil || *receipt.AssetID != asset.ID || (receipt.Status != "ready" && receipt.Status != "bound") {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	return r.db.Model(receipt).Update("status", "bound").Error
}

func (r *MediaRepository) RemoveBusinessDraftMediaUploadReceiptWithinRetirementTransaction(asset *media.MediaAsset) error {
	if _, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(asset.StorageKey); !draft {
		return nil
	}
	return r.db.Where("asset_id = ?", asset.ID).Delete(&media.BusinessDraftMediaUploadReceipt{}).Error
}

func (r *MediaRepository) FindExpiredBusinessDraftMediaUploadReceipts(ctx context.Context, now time.Time, limit int) ([]media.BusinessDraftMediaUploadReceipt, error) {
	var receipts []media.BusinessDraftMediaUploadReceipt
	if !r.hasTable("business_draft_media_upload_receipts") {
		return receipts, nil
	}
	if limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	err := r.db.WithContext(ctx).Where("status IN ? AND expires_at <= ? AND next_expiry_attempt_at <= ?", []string{"uploading", "ready"}, now, now).Order("expires_at ASC, id ASC").Limit(limit).Find(&receipts).Error
	return receipts, err
}

func (r *MediaRepository) DelayRetryOfBlockedBusinessDraftMediaUploadExpiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&media.BusinessDraftMediaUploadReceipt{}).Where("id = ? AND status IN ?", id, []string{"uploading", "ready"}).Update("next_expiry_attempt_at", time.Now().UTC().Add(5*time.Minute)).Error
}

func (r *MediaRepository) DeleteBusinessDraftMediaUploadReceiptWithinTransaction(id string) error {
	return r.db.Delete(&media.BusinessDraftMediaUploadReceipt{}, "id = ?", id).Error
}

func (r *MediaRepository) ScheduleFailedBusinessDraftMediaUploadReceiptForExpiry(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&media.BusinessDraftMediaUploadReceipt{}).Where("id = ? AND status = ?", id, "uploading").Update("expires_at", time.Now().UTC()).Error
}

func (r *MediaRepository) ValidateBusinessDraftMediaDerivativeGenerationIsPermanentlyBound(asset *media.MediaAsset) error {
	id, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(asset.StorageKey)
	if !draft {
		return nil
	}
	var receipt media.BusinessDraftMediaUploadReceipt
	if err := r.db.First(&receipt, "id = ? AND asset_id = ? AND status = ?", id, asset.ID, "bound").Error; err != nil {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	return nil
}

func (r *MediaRepository) RejectUnpreparedBusinessDraftMediaObjectRegistration(key string) error {
	if !strings.HasPrefix(key, storage.BusinessDraftMediaObjectPrefix+"/") {
		return nil
	}
	id, valid := storage.BusinessDraftMediaReceiptIDFromObjectKey(key)
	if !valid || r.preparedBusinessDraftUploadReceiptID != id {
		return fmt.Errorf("%w: draft objects require their prepared upload registration", ErrBusinessDraftMediaUploadReceiptUnavailable)
	}
	return nil
}

func (r *MediaRepository) EnsureBusinessDraftMediaReceiptIsAbsentBeforeExpiredObjectCleanup(id string) error {
	var count int64
	if err := r.db.Model(&media.BusinessDraftMediaUploadReceipt{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count != 0 {
		return ErrBusinessDraftMediaUploadReceiptUnavailable
	}
	return nil
}
