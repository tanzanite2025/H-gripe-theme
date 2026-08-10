package service

import "tanzanite/internal/domain/product"

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
	if input.MetaTitle != nil {
		existingProduct.MetaTitle = *input.MetaTitle
	}
	if input.MetaDescription != nil {
		existingProduct.MetaDesc = *input.MetaDescription
	}
	if err := s.productRepo.Update(existingProduct); err != nil {
		return nil, err
	}

	s.clearProductCache(existingProduct)
	s.invalidateStorefrontHTMLCache("product SEO update")
	return s.findProduct(id)
}
