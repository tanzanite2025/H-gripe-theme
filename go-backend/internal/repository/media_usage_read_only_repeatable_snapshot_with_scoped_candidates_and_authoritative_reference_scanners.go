package repository

import (
	"context"
	"database/sql"
	"errors"

	"commerce-platform/internal/domain/media"
	"gorm.io/gorm"
)

var ErrMediaUsageSnapshotLimitExceeded = errors.New("media usage snapshot exceeded its candidate or reference limit")

// WalkReadOnlyMediaUsageSnapshot uses one database snapshot for candidates and
// every business scanner. Its callback must only build a report; it must never
// perform provider IO, authorize deletion, or mutate business data.
func (r *MediaRepository) WalkReadOnlyMediaUsageSnapshot(ctx context.Context, filter MediaAssetFilter, candidateLimit, referenceLimit int, consume func(media.MediaAsset, []media.AssetReference)) error {
	if r == nil || r.db == nil || consume == nil {
		return gorm.ErrInvalidDB
	}
	if candidateLimit < 1 || referenceLimit < 1 {
		return ErrMediaUsageSnapshotLimitExceeded
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		snapshotRepository := r.WithTx(tx)
		var assets []media.MediaAsset
		if err := snapshotRepository.mediaAssetLibraryListQueryWithBusinessPartitionAndAccessScope(filter).
			Select("id, filename, original_filename, library_partition, visibility, url, storage_key").Order("id ASC").Limit(candidateLimit + 1).Find(&assets).Error; err != nil {
			return err
		}
		if len(assets) > candidateLimit {
			return ErrMediaUsageSnapshotLimitExceeded
		}
		// Draft workflow receipts are not business-reference/deletion authority,
		// but an unfinished upload must not appear in the unused review filter.
		if len(assets) > 0 && snapshotRepository.hasTable("business_draft_media_upload_receipts") {
			ids := make([]uint, 0, len(assets))
			for _, asset := range assets {
				ids = append(ids, asset.ID)
			}
			var receipts []media.BusinessDraftMediaUploadReceipt
			if err := tx.Select("asset_id, expires_at").Where("asset_id IN ? AND status <> ?", ids, "bound").Find(&receipts).Error; err != nil {
				return err
			}
			for _, receipt := range receipts {
				if receipt.AssetID == nil {
					continue
				}
				for index := range assets {
					if assets[index].ID == *receipt.AssetID {
						expiry := receipt.ExpiresAt
						assets[index].BusinessDraftUploadExpiresAt = &expiry
						break
					}
				}
			}
		}
		remainingReferences := referenceLimit
		for _, asset := range assets {
			if err := ctx.Err(); err != nil {
				return err
			}
			references, err := snapshotRepository.FindAssetReferences(&asset)
			if err != nil {
				return err
			}
			remainingReferences -= len(references)
			if remainingReferences < 0 {
				return ErrMediaUsageSnapshotLimitExceeded
			}
			consume(asset, references)
		}
		return ctx.Err()
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}
