package service

import "commerce-platform/internal/domain/product"

// Category translations are excluded from the domain model's API JSON. Store
// them explicitly in the internal cache document so cache hits retain the same
// localized category names as database reads.
type productDetailCacheDocumentWithCategoryTranslations struct {
	Product              product.Product                      `json:"product"`
	CategoryTranslations []product.ProductCategoryTranslation `json:"category_translations,omitempty"`
}

func buildProductDetailCacheDocumentWithCategoryTranslations(item *product.Product) productDetailCacheDocumentWithCategoryTranslations {
	document := productDetailCacheDocumentWithCategoryTranslations{Product: *item}
	if item.ProductCategory != nil {
		document.CategoryTranslations = item.ProductCategory.Translations
	}
	return document
}
