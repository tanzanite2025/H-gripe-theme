package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type recordingMerchantPublisher struct {
	upserts           int
	withdraws         int
	revalidate        int
	revalidateOfferID uint
	revalidateReason  string
}

func (p *recordingMerchantPublisher) EnqueueProductUpsert(uint, string) error {
	p.upserts++
	return nil
}

func (p *recordingMerchantPublisher) EnqueueProductWithdraw(uint, string) error {
	p.withdraws++
	return nil
}

func (p *recordingMerchantPublisher) EnqueueOfferRevalidate(offerID uint, reason string) error {
	p.revalidate++
	p.revalidateOfferID = offerID
	p.revalidateReason = reason
	return nil
}

func TestProductSEOUpdateUsesDedicatedBoundaryWithoutMerchantEvent(t *testing.T) {
	_, productService := newTestProductService(t)
	created, err := productService.CreateAdminProduct(ProductCreateInput{
		Name:   "SEO Boundary Product",
		Slug:   "seo-boundary-product",
		Status: "active",
		Locale: "en",
		Variants: []ProductVariantInput{{
			SKU:       "SEO-BOUNDARY-001",
			Price:     99,
			Stock:     1,
			IsDefault: true,
			IsActive:  boolPtr(true),
		}},
	})
	require.NoError(t, err)

	publisher := &recordingMerchantPublisher{}
	productService.ConfigureMerchantEventPublisher(publisher)
	title := "SEO Boundary Product | Tanzanite"
	description := "A description maintained by the SEO control plane."

	updated, err := productService.UpdateProductSEO(created.ID, ProductSEOUpdateInput{
		MetaTitle:       &title,
		MetaDescription: &description,
	})

	require.NoError(t, err)
	require.Equal(t, title, updated.MetaTitle)
	require.Equal(t, description, updated.MetaDesc)
	require.Zero(t, publisher.upserts)
	require.Zero(t, publisher.withdraws)
	require.Zero(t, publisher.revalidate)
}
