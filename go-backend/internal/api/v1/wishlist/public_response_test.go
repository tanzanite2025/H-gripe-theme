package wishlist

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	wishlistdomain "commerce-platform/internal/domain/wishlist"
)

func TestPublicWishlistResponseOmitsCommercialFields(t *testing.T) {
	productTypeID := uint(2)
	shippingTemplateID := uint(3)
	mediaAssetID := uint(4)

	item := wishlistdomain.Item{
		ID:        11,
		UserID:    12,
		ProductID: 13,
		CreatedAt: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
		Product: &productdomain.Product{
			ID:                 13,
			ProductTypeID:      &productTypeID,
			ShippingTemplateID: &shippingTemplateID,
			SKU:                "INTERNAL-WISHLIST-SKU",
			Name:               "Saved Product",
			Slug:               "saved-product",
			Stock:              48,
			Status:             "active",
			Locale:             "en",
			ViewCount:          302,
			Media: []productdomain.ProductMedia{
				{
					ID:           15,
					ProductID:    13,
					MediaAssetID: &mediaAssetID,
					MediaType:    "image",
					URL:          "https://example.test/full.jpg",
					ThumbnailURL: "https://example.test/thumb.jpg",
					IsPrimary:    true,
					IsVisible:    true,
				},
			},
			Variants: []productdomain.ProductVariant{
				{
					ID:        16,
					ProductID: 13,
					SKU:       "INTERNAL-WISHLIST-VARIANT-SKU",
					Stock:     6,
					Weight:    1300,
					IsActive:  true,
					IsDefault: true,
				},
			},
		},
	}

	payload, err := json.Marshal(publicWishlistResponse(item))
	if err != nil {
		t.Fatalf("marshal public wishlist response: %v", err)
	}
	body := string(payload)
	for _, field := range []string{
		`"user_id"`,
		`"created_at"`,
		`"sku"`,
		`"stock"`,
		`"status"`,
		`"locale"`,
		`"view_count"`,
		`"product_type_id"`,
		`"shipping_template_id"`,
		`"media_asset_id"`,
		`"weight_grams"`,
	} {
		if strings.Contains(body, field) {
			t.Fatalf("public wishlist response exposes internal field %s: %s", field, body)
		}
	}
	if !strings.Contains(body, `"thumbnail":"https://example.test/thumb.jpg"`) {
		t.Fatalf("public wishlist response omits selected thumbnail: %s", body)
	}
	if !strings.Contains(body, `"availability":"in_stock"`) {
		t.Fatalf("public wishlist response omits coarse availability: %s", body)
	}
}
