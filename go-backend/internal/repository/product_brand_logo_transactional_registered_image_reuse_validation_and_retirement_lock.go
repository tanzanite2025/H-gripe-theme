package repository

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/product"
	"gorm.io/gorm"
)

var ErrProductBrandLogoReferenceInvalid = errors.New("product brand logo reference is invalid")

type ProductBrandLogoFieldMutation struct {
	UpdateLogo bool
}

func (r *ProductBrandRepository) validateAndCanonicalizeProductBrandLogoUnderMediaAssetRetirementLock(tx *gorm.DB, brand *product.ProductBrand, previousURL string) error {
	// Preserve an unchanged historical placement. An explicit restore after
	// detachment is a new binding and must validate against the locked row.
	if brand.LogoURL == previousURL {
		return nil
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, []string{brand.LogoURL}, nil)
	if err == nil {
		err = resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(publicBusinessMediaReferenceBinding{URL: &brand.LogoURL, ExpectedMediaType: "image"})
	}
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrProductBrandLogoReferenceInvalid, err)
	}
	return err
}
