package service

import "commerce-platform/internal/domain/product"

// ProductSEOUpdateInput is the only product write contract exposed to the SEO
// control plane. Catalog create/update inputs deliberately do not contain SEO
// fields, so product editing cannot become a second SEO entry point.
type ProductSEOUpdateInput struct {
	MetaTitle       *string
	MetaDescription *string
}

func (s *ProductService) UpdateProductSEO(id uint, input ProductSEOUpdateInput) (*product.Product, error) {
	if s == nil {
		return nil, ErrProductNotFound
	}

	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}
	metaTitle := existingProduct.MetaTitle
	metaDescription := existingProduct.MetaDesc
	if input.MetaTitle != nil {
		metaTitle = *input.MetaTitle
	}
	if input.MetaDescription != nil {
		metaDescription = *input.MetaDescription
	}
	if err := s.productRepo.UpdateSEO(id, metaTitle, metaDescription); err != nil {
		return nil, err
	}
	existingProduct.MetaTitle = metaTitle
	existingProduct.MetaDesc = metaDescription

	s.clearProductCache(existingProduct)
	if err := s.enqueueProductCacheInvalidationByIDs([]uint{existingProduct.ID}, "product SEO update"); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("product SEO update")
	return s.findProduct(id)
}
