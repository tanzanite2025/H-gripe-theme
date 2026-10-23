package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/product"
)

var ErrProductCategoryImageReferenceInvalid = errors.New("product category image reference invalid")

func (r *ProductCategoryRepository) lockValidateAndResolveCategoryImageFromRegisteredAsset(tx *gorm.DB, category *product.ProductCategory, previous ...product.ProductCategory) error {
	if category.ImageMediaAssetID == nil {
		category.ImageURL = ""
		return nil
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, nil, []uint{*category.ImageMediaAssetID})
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrProductCategoryImageReferenceInvalid, err)
	}
	if err != nil {
		return err
	}
	return validateAndResolveCategoryImageFromLockedRegisteredAsset(tx, category, resolved, previous...)
}

func validateAndResolveCategoryImageFromLockedRegisteredAsset(tx *gorm.DB, category *product.ProductCategory, resolved *lockedPublicBusinessMediaReferenceIdentities, previous ...product.ProductCategory) error {
	if category.ImageMediaAssetID == nil {
		category.ImageURL = ""
		return nil
	}
	asset := resolved.assets[*category.ImageMediaAssetID]
	if asset == nil || asset.MediaType != "image" || strings.TrimSpace(asset.URL) == "" {
		return ErrProductCategoryImageReferenceInvalid
	}
	if resolved.dedicatedObjectNamespaceForResolvedPublicMediaReference("", asset.ID) != "" {
		if len(previous) == 0 || previous[0].ImageMediaAssetID == nil || *previous[0].ImageMediaAssetID != asset.ID {
			return fmt.Errorf("%w: dedicated objects cannot be newly selected as category images", ErrProductCategoryImageReferenceInvalid)
		}
		category.ImageURL = previous[0].ImageURL
		return nil
	}
	category.ImageURL = asset.URL
	return NewMediaRepository(tx).BindBusinessDraftMediaUploadReceiptWithinTransaction(asset)
}
