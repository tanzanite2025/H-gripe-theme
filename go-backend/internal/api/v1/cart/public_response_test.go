package cart

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "tanzanite/internal/domain/product"
)

func TestPublicCartSummaryUsesMinimalMediaContract(t *testing.T) {
	variantID := uint(9)
	shippingTemplateID := uint(12)
	productTypeID := uint(13)
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
					ID:                 7,
					ProductTypeID:      &productTypeID,
					ShippingTemplateID: &shippingTemplateID,
					SKU:                "INTERNAL-PRODUCT-SKU",
					Name:               "Public Cart Product",
					Slug:               "public-cart-product",
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
		`"product_type_id"`,
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
