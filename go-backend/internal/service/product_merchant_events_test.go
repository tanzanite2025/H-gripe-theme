package service

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/assert"
)

func TestMerchantProductUpdateAffectsChannelWhenBrandChanges(t *testing.T) {
	assert.True(t, merchantProductUpdateAffectsChannel(ProductUpdateInput{
		UpdateBrandID: true,
	}))
	assert.Equal(t, "product_brand_changed", merchantProductSourceChangeReason(ProductUpdateInput{
		UpdateBrandID: true,
	}))
}

func TestMerchantProductCoreChangedWhenBrandChanges(t *testing.T) {
	previousBrandID := uint(1)
	nextBrandID := uint(2)
	previous := &product.Product{BrandID: &previousBrandID}
	next := &product.Product{BrandID: &nextBrandID}

	assert.True(t, merchantProductCoreChanged(previous, next))
}
