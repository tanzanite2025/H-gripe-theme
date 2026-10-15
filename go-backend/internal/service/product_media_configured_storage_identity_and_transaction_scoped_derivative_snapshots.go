package service

import (
	"gorm.io/gorm"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
)

func (s *MediaService) ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference string) (string, error) {
	if s == nil {
		return "", ErrMediaStorageUnavailable
	}
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, s.storage, s.siteURL)
}

// The product repository has already locked all participating asset rows. Use
// its transaction for both preset and derivative reads, including fallback saves.
func (s *MediaService) ProductMediaImageVariantsWithinProductTransaction(tx *gorm.DB, assetID uint) (map[string]product.ProductMediaImageVariant, string, error) {
	if s == nil || s.repo == nil {
		return nil, "", ErrMediaStorageUnavailable
	}
	transactionService := *s
	transactionService.repo = s.repo.WithTx(tx)
	if s.derivativePresets != nil {
		transactionService.derivativePresets = s.derivativePresets.WithTx(tx)
	}
	return transactionService.ProductMediaImageVariants(assetID)
}
