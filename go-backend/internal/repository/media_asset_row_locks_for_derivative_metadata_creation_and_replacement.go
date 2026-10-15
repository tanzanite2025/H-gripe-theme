package repository

import (
	"sort"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
)

func lockActiveMediaAssetsBeforePersistingDerivativeMetadata(tx *gorm.DB, derivatives []media.MediaAssetDerivative) error {
	unique := make(map[uint]struct{})
	for _, derivative := range derivatives {
		unique[derivative.MediaAssetID] = struct{}{}
	}
	ids := make([]uint, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		asset, err := NewMediaRepository(tx).FindAssetByIDForUpdate(id)
		if err != nil {
			return err
		}
		if asset.Status != "active" {
			return gorm.ErrRecordNotFound
		}
	}
	return nil
}
