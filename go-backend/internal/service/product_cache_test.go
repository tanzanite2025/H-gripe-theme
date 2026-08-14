package service

import (
	"testing"

	"commerce-platform/internal/domain/product"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductDetailCacheInvalidatorDeletesIDAndSlugKeys(t *testing.T) {
	repo := &fakeProductCacheIdentityRepository{
		productsByID: map[uint]product.Product{
			7: {ID: 7, Slug: "carbon-rim", Locale: "zh-CN"},
		},
	}
	cache := &recordingProductDetailCache{}
	invalidator := NewProductDetailCacheInvalidator(repo, cache)

	invalidator.InvalidateProductCacheByIDs([]uint{7, 7, 0})

	require.ElementsMatch(t, []uint{7}, repo.requestedIDs)
	assert.ElementsMatch(t, []string{
		"product:7",
		"product:slug:carbon-rim:",
		"product:slug:carbon-rim:zh-CN",
		"product:slug:carbon-rim:zh_cn",
	}, cache.deletedKeys)
}

func TestProductDetailCacheInvalidatorDeletesDependentProductKeys(t *testing.T) {
	repo := &fakeProductCacheIdentityRepository{
		productsByProductTypeID: map[uint][]product.Product{
			3: {
				{ID: 11, Slug: "wheelset", Locale: "en"},
			},
		},
		productsByTemplateID: map[uint][]product.Product{
			5: {
				{ID: 12, Slug: "packing-rim", Locale: "en"},
			},
		},
	}
	cache := &recordingProductDetailCache{}
	invalidator := NewProductDetailCacheInvalidator(repo, cache)

	invalidator.InvalidateProductCacheByProductTypeID(3)
	invalidator.InvalidateProductCacheByInformationTemplateID(5)

	assert.Equal(t, []uint{3}, repo.requestedProductTypeIDs)
	assert.Equal(t, []uint{5}, repo.requestedTemplateIDs)
	assert.Contains(t, cache.deletedKeys, "product:11")
	assert.Contains(t, cache.deletedKeys, "product:slug:wheelset:en")
	assert.Contains(t, cache.deletedKeys, "product:12")
	assert.Contains(t, cache.deletedKeys, "product:slug:packing-rim:en")
}

func TestProductDetailCacheInvalidatorDeletesBrandProductKeys(t *testing.T) {
	repo := &fakeProductCacheIdentityRepository{
		productsByBrandID: map[uint][]product.Product{
			4: {
				{ID: 21, Slug: "dt-swiss-rim", Locale: "en"},
				{ID: 22, Slug: "dt-swiss-wheelset", Locale: "zh-CN"},
			},
		},
	}
	cache := &recordingProductDetailCache{}
	invalidator := NewProductDetailCacheInvalidator(repo, cache)

	result, err := invalidator.InvalidateProductCacheByBrandIDWithSource(4, productCacheInvalidationSourceDirect)

	require.NoError(t, err)
	assert.Equal(t, ProductCacheInvalidationResult{Products: 2, Keys: 7}, result)
	assert.Equal(t, []uint{4}, repo.requestedBrandIDs)
	assert.Contains(t, cache.deletedKeys, "product:21")
	assert.Contains(t, cache.deletedKeys, "product:slug:dt-swiss-rim:en")
	assert.Contains(t, cache.deletedKeys, "product:22")
	assert.Contains(t, cache.deletedKeys, "product:slug:dt-swiss-wheelset:zh-CN")
}

type recordingProductDetailCache struct {
	deletedKeys []string
}

func (c *recordingProductDetailCache) Delete(key string) error {
	c.deletedKeys = append(c.deletedKeys, key)
	return nil
}

type recordingProductCacheInvalidator struct {
	productIDs     []uint
	productTypeIDs []uint
	brandIDs       []uint
	templateIDs    []uint
}

func (i *recordingProductCacheInvalidator) InvalidateProductCacheByIDs(ids []uint) {
	i.productIDs = append(i.productIDs, ids...)
}

func (i *recordingProductCacheInvalidator) InvalidateProductCacheByProductTypeID(productTypeID uint) {
	i.productTypeIDs = append(i.productTypeIDs, productTypeID)
}

func (i *recordingProductCacheInvalidator) InvalidateProductCacheByBrandID(brandID uint) {
	i.brandIDs = append(i.brandIDs, brandID)
}

func (i *recordingProductCacheInvalidator) InvalidateProductCacheByInformationTemplateID(templateID uint) {
	i.templateIDs = append(i.templateIDs, templateID)
}

type fakeProductCacheIdentityRepository struct {
	productsByID            map[uint]product.Product
	productsByProductTypeID map[uint][]product.Product
	productsByBrandID       map[uint][]product.Product
	productsByTemplateID    map[uint][]product.Product
	requestedIDs            []uint
	requestedProductTypeIDs []uint
	requestedBrandIDs       []uint
	requestedTemplateIDs    []uint
}

func (r *fakeProductCacheIdentityRepository) FindProductCacheIdentitiesByIDs(ids []uint) ([]product.Product, error) {
	r.requestedIDs = append(r.requestedIDs, ids...)
	products := make([]product.Product, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.productsByID[id]; ok {
			products = append(products, item)
		}
	}
	return products, nil
}

func (r *fakeProductCacheIdentityRepository) FindProductCacheIdentitiesByProductTypeID(productTypeID uint) ([]product.Product, error) {
	r.requestedProductTypeIDs = append(r.requestedProductTypeIDs, productTypeID)
	return append([]product.Product(nil), r.productsByProductTypeID[productTypeID]...), nil
}

func (r *fakeProductCacheIdentityRepository) FindProductCacheIdentitiesByBrandID(brandID uint) ([]product.Product, error) {
	r.requestedBrandIDs = append(r.requestedBrandIDs, brandID)
	return append([]product.Product(nil), r.productsByBrandID[brandID]...), nil
}

func (r *fakeProductCacheIdentityRepository) FindProductCacheIdentitiesByInformationTemplateID(templateID uint) ([]product.Product, error) {
	r.requestedTemplateIDs = append(r.requestedTemplateIDs, templateID)
	return append([]product.Product(nil), r.productsByTemplateID[templateID]...), nil
}
