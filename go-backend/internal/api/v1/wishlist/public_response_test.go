package wishlist

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	productdomain "commerce-platform/internal/domain/product"
	wishlistdomain "commerce-platform/internal/domain/wishlist"
	"commerce-platform/internal/service"
)

func TestPublicWishlistResponseOmitsCommercialFields(t *testing.T) {
	productSpecificationTemplateID := uint(2)
	shippingTemplateID := uint(3)
	mediaAssetID := uint(4)

	item := wishlistdomain.Item{
		ID:        11,
		UserID:    12,
		ProductID: 13,
		CreatedAt: time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC),
		Product: &productdomain.Product{
			ID:                             13,
			ProductSpecificationTemplateID: &productSpecificationTemplateID,
			ShippingTemplateID:             &shippingTemplateID,
			SKU:                            "INTERNAL-WISHLIST-SKU",
			Name:                           "Saved Product",
			Slug:                           "saved-product",
			Stock:                          48,
			Status:                         "active",
			Locale:                         "en",
			ViewCount:                      302,
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
		`"product_specification_template_id"`,
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

func TestPublicWishlistResponseCanonicalizesFirstPartyThumbnailURL(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := wishlistdomain.Item{
		ID:        1,
		ProductID: 2,
		Product: &productdomain.Product{
			ID:     2,
			Name:   "Saved Media Product",
			Slug:   "saved-media-product",
			Status: "active",
			Media: []productdomain.ProductMedia{
				{
					MediaType:    "image",
					ThumbnailURL: "http://media.internal:8080/uploads/wishlist/thumb.webp",
					URL:          "http://media.internal:8080/uploads/wishlist/full.webp",
					IsPrimary:    true,
					IsVisible:    true,
				},
			},
		},
	}

	payload, err := json.Marshal(publicWishlistResponse(item, resolver))
	if err != nil {
		t.Fatalf("marshal public wishlist response: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "media.internal") {
		t.Fatalf("public wishlist response exposes internal media origin: %s", body)
	}
	if !strings.Contains(body, `"thumbnail":"https://shop.example.test/uploads/wishlist/thumb.webp"`) {
		t.Fatalf("public wishlist response missing canonical thumbnail: %s", body)
	}
}
