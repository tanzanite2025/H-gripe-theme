package repository

import (
	"context"

	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) WithContextForMediaLibraryClassification(ctx context.Context) *MediaRepository {
	return NewMediaRepository(r.db.WithContext(ctx))
}

// Caller owns the transaction and the asset lock. Update only management
// classification, so upload origins, URLs, files and business references persist.
func (r *MediaRepository) UpdateMediaAssetLibraryClassificationAtExpectedRevision(asset *media.MediaAsset, expectedRevision int64) (bool, error) {
	updatedAt := r.db.NowFunc()
	result := r.db.Model(&media.MediaAsset{}).Where("id = ? AND classification_revision = ?", asset.ID, expectedRevision).
		Updates(map[string]interface{}{"library_partition": asset.LibraryPartition, "classification_method": asset.ClassificationMethod,
			"customer_service_media_purpose": asset.CustomerServiceMediaPurpose, "classification_revision": asset.ClassificationRevision, "updated_at": updatedAt})
	if result.Error == nil && result.RowsAffected == 1 {
		asset.UpdatedAt = updatedAt
	}
	return result.RowsAffected == 1, result.Error
}

func (r *MediaRepository) UpdateMediaAssetEditableMetadataWithoutReplacingLibraryClassification(asset *media.MediaAsset) error {
	updatedAt := r.db.NowFunc()
	err := r.db.Model(&media.MediaAsset{}).Where("id = ?", asset.ID).
		Updates(map[string]interface{}{"alt": asset.Alt, "caption": asset.Caption, "status": asset.Status, "visibility": asset.Visibility, "updated_at": updatedAt}).Error
	if err == nil {
		asset.UpdatedAt = updatedAt
	}
	return err
}
