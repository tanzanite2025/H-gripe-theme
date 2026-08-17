package service

import productdomain "commerce-platform/internal/domain/product"

type productMediaImageVariantResolver interface {
	ProductMediaImageVariants(assetID uint) (map[string]productdomain.ProductMediaImageVariant, string, error)
}
