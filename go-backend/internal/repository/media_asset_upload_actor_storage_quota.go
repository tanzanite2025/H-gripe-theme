package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/media"

	"gorm.io/gorm"
)

var ErrMediaUploadActorStorageQuotaExceeded = errors.New("media upload actor storage quota exceeded")

// AssetStorageUsageByUploadActor includes soft-deleted originals until their
// physical deletion is confirmed. Derivative capacity accounting is separate.
// Legacy rows cannot identify an account namespace, so typed account uploads
// conservatively include legacy rows with the same numeric ID.
func (r *MediaRepository) AssetStorageUsageByUploadActor(identity media.MediaUploadActorIdentity) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}
	if !identity.IsValidMediaUploadActorIdentity() {
		return 0, fmt.Errorf("invalid media upload actor identity")
	}
	if identity.ActorType == media.MediaUploadActorLegacyAccount {
		// Preserve the aggregate quota semantics for callers using only UploaderID.
		return r.AssetStorageUsageByUploaderID(identity.AccountID)
	}
	query := r.db.Unscoped().Model(&media.MediaAsset{})
	switch identity.ActorType {
	case media.MediaUploadActorVisitor:
		query = query.Where("uploader_actor_type = ? AND uploader_visitor_session_hash = ?", identity.ActorType, identity.VisitorSessionHash)
	default:
		query = query.Where("uploader_id = ? AND uploader_actor_type IN ?", identity.AccountID,
			[]media.MediaUploadActorType{identity.ActorType, media.MediaUploadActorLegacyAccount})
	}
	var total int64
	err := query.Select("COALESCE(SUM(CASE WHEN size > 0 THEN size ELSE 0 END), 0)").Scan(&total).Error
	return total, err
}

// CreateAssetWithinUploadActorStorageQuota rechecks usage while serializing
// PostgreSQL uploads for the quota owner. Object upload stays outside the short
// transaction; the caller compensates by deleting the object on rejection.
func (r *MediaRepository) CreateAssetWithinUploadActorStorageQuota(ctx context.Context, asset *media.MediaAsset, quotaBytes int64, customerServiceUploadSessionIDs ...string) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if asset == nil {
		return fmt.Errorf("media asset is required for upload quota registration")
	}
	identity := media.MediaUploadActorIdentity{
		ActorType: asset.UploaderActorType, AccountID: asset.UploaderID,
		VisitorSessionHash: asset.UploaderVisitorSessionHash,
	}
	if !identity.IsValidMediaUploadActorIdentity() || asset.Size <= 0 {
		return fmt.Errorf("invalid media upload actor or file size")
	}
	var transactionOptions []*sql.TxOptions
	if r.db.Name() == "postgres" {
		transactionOptions = append(transactionOptions, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if tx.Name() == "postgres" {
			// Account namespaces share a lock because their legacy usage overlaps.
			lockOwner := fmt.Sprintf("media-upload-quota/account/%d", identity.AccountID)
			if identity.ActorType == media.MediaUploadActorVisitor {
				lockOwner = "media-upload-quota/visitor/" + identity.VisitorSessionHash
			}
			lockDigest := sha256.Sum256([]byte(lockOwner))
			var lockID int64
			if _, err := binary.Decode(lockDigest[:8], binary.BigEndian, &lockID); err != nil {
				return fmt.Errorf("decode media upload quota lock identity: %w", err)
			}
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockID).Error; err != nil {
				return err
			}
		}
		transactionRepository := r.WithTx(tx)
		used, err := transactionRepository.AssetStorageUsageByUploadActor(identity)
		if err != nil {
			return err
		}
		if quotaBytes <= 0 || used >= quotaBytes || asset.Size > quotaBytes-used {
			return ErrMediaUploadActorStorageQuotaExceeded
		}
		if err := transactionRepository.CreateAsset(asset); err != nil {
			return err
		}
		if len(customerServiceUploadSessionIDs) > 0 && customerServiceUploadSessionIDs[0] != "" {
			return transactionRepository.RegisterCustomerServiceUploadSessionAssetWithinTransaction(customerServiceUploadSessionIDs[0], asset)
		}
		return nil
	}, transactionOptions...)
}
