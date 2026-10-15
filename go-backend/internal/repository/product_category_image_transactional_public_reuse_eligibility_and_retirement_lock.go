package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/product"
)

var ErrProductCategoryImageReferenceInvalid = errors.New("product category image reference invalid")

func lockValidateAndResolveCategoryImageFromRegisteredAsset(tx *gorm.DB, category *product.ProductCategory) error {
	if category.ImageMediaAssetID == nil {
		category.ImageURL = ""
		return nil
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, nil, nil, []uint{*category.ImageMediaAssetID})
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrProductCategoryImageReferenceInvalid, err)
	}
	if err != nil {
		return err
	}
	asset := resolved.assets[*category.ImageMediaAssetID]
	if asset.MediaType != "image" || strings.TrimSpace(asset.URL) == "" {
		return ErrProductCategoryImageReferenceInvalid
	}
	category.ImageURL = asset.URL
	return nil
}
