package cart

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/service"
)

func TestPublicCartSummaryUsesMinimalMediaContract(t *testing.T) {
	variantID := uint(9)
	shippingTemplateID := uint(12)
	productSpecificationTemplateID := uint(13)
	mediaAssetID := uint(14)

	summary := &productdomain.CartSummary{
		ItemCount: 1,
		Total:     79.5,
		Items: []productdomain.CartItem{
			{
				ID:        21,
				CartID:    22,
				ProductID: 7,
				VariantID: &variantID,
				Quantity:  1,
				Price:     79.5,
				Currency:  "USD",
				Product: &productdomain.Product{
					ID:                             7,
					ProductSpecificationTemplateID: &productSpecificationTemplateID,
					ShippingTemplateID:             &shippingTemplateID,
					SKU:                            "INTERNAL-PRODUCT-SKU",
					Name:                           "Public Cart Product",
					Slug:                           "public-cart-product",
					Media: []productdomain.ProductMedia{
						{
							ID:           31,
							ProductID:    7,
							VariantID:    &variantID,
							MediaAssetID: &mediaAssetID,
							MediaType:    "image",
							Role:         "primary",
							URL:          "https://example.test/public.jpg",
							Alt:          "Public product",
							Locale:       "en",
							IsPrimary:    true,
							IsVisible:    true,
						},
						{
							ID:        32,
							ProductID: 7,
							MediaType: "image",
							URL:       "https://example.test/hidden.jpg",
							IsVisible: false,
						},
					},
				},
				Variant: &productdomain.ProductVariant{
					ID:                 variantID,
					ProductID:          7,
					ShippingTemplateID: &shippingTemplateID,
					SKU:                "INTERNAL-VARIANT-SKU",
					Title:              "700c / Black",
					Weight:             1450,
					IsDefault:          true,
					IsActive:           true,
					Stock:              4,
				},
			},
		},
	}

	payload, err := json.Marshal(PublicCartSummaryFromDomain(summary))
	if err != nil {
		t.Fatalf("marshal public cart summary: %v", err)
	}
	body := string(payload)

	for _, field := range []string{
		`"product_specification_template_id"`,
		`"shipping_template_id"`,
		`"sku"`,
		`"weight_grams"`,
		`"media_asset_id"`,
		`"is_visible"`,
	} {
		if strings.Contains(body, field) {
			t.Fatalf("public cart response exposes internal field %s: %s", field, body)
		}
	}
	if strings.Contains(body, "hidden.jpg") {
		t.Fatalf("public cart response exposes hidden media: %s", body)
	}

	var response PublicCartSummary
	if err := json.Unmarshal(payload, &response); err != nil {
		t.Fatalf("unmarshal public cart summary: %v", err)
	}
	if len(response.Items) != 1 || response.Items[0].Product == nil {
		t.Fatalf("public cart response omitted product: %s", body)
	}
	if len(response.Items[0].Product.Media) != 1 {
		t.Fatalf("expected one visible media item, got %d: %s", len(response.Items[0].Product.Media), body)
	}
	if got := response.Items[0].Currency; got != "USD" {
		t.Fatalf("unexpected public cart item currency: %s", got)
	}
	if got := response.Items[0].Product.Media[0].URL; got != "https://example.test/public.jpg" {
		t.Fatalf("unexpected public media URL: %s", got)
	}
}

func TestPublicCartSummaryCanonicalizesFirstPartyMediaURLs(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	summary := &productdomain.CartSummary{
		ItemCount: 1,
		Items: []productdomain.CartItem{
			{
				ID:       1,
				Quantity: 1,
				Product: &productdomain.Product{
					ID:   2,
					Name: "Media Cart Product",
					Slug: "media-cart-product",
					Media: []productdomain.ProductMedia{
						{
							MediaType:    "image",
							URL:          "http://media.internal:8080/uploads/cart/full.webp",
							ThumbnailURL: "http://media.internal:8080/uploads/cart/thumb.webp",
							IsVisible:    true,
						},
					},
				},
			},
		},
	}

	payload, err := json.Marshal(PublicCartSummaryFromDomain(summary, resolver))
	if err != nil {
		t.Fatalf("marshal public cart summary: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "media.internal") {
		t.Fatalf("public cart response exposes internal media origin: %s", body)
	}
	if !strings.Contains(body, `"url":"https://shop.example.test/uploads/cart/full.webp"`) ||
		!strings.Contains(body, `"thumbnail_url":"https://shop.example.test/uploads/cart/thumb.webp"`) {
		t.Fatalf("public cart response missing canonical media URLs: %s", body)
	}
}
