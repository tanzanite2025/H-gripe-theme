package product

import (
	"encoding/json"
	"strings"
	"testing"

	productdomain "tanzanite/internal/domain/product"
)

func TestPublicProductFromDomainExposesPurchasableSkuFactsWithoutExactInventory(t *testing.T) {
	item := productdomain.Product{
		ID:                 11,
		ProductTypeID:      pointerTo(uint(3)),
		ShippingTemplateID: pointerTo(uint(7)),
		SKU:                "PUBLIC-PRODUCT",
		Name:               "Public Product",
		Slug:               "public-product",
		Stock:              47,
		Status:             "active",
		ViewCount:          146,
		Variants: []productdomain.ProductVariant{
			{
				ID:                 12,
				ProductID:          11,
				ShippingTemplateID: pointerTo(uint(7)),
				SKU:                "PUBLIC-PRODUCT-DEFAULT",
				Stock:              3,
				Weight:             1400,
				IsActive:           true,
			},
		},
	}

	payload, err := json.Marshal(PublicProductFromDomain(item))
	if err != nil {
		t.Fatalf("marshal public product: %v", err)
	}
	if strings.Contains(string(payload), `"stock"`) {
		t.Fatalf("public product response exposes exact inventory: %s", payload)
	}
	if strings.Contains(string(payload), `"view_count"`) {
		t.Fatalf("public product response exposes exact view count: %s", payload)
	}
	for _, field := range []string{
		`"product_type_id"`,
		`"shipping_template_id"`,
		`"product_id"`,
		`"status"`,
		`"locale"`,
		`"created_at"`,
		`"updated_at"`,
	} {
		if strings.Contains(string(payload), field) {
			t.Fatalf("public product response exposes internal field %s: %s", field, payload)
		}
	}
	if !strings.Contains(string(payload), `"availability":"in_stock"`) {
		t.Fatalf("public product response omits availability: %s", payload)
	}
	if !strings.Contains(string(payload), `"sku":"PUBLIC-PRODUCT-DEFAULT"`) {
		t.Fatalf("public product response omits selected SKU: %s", payload)
	}
	if !strings.Contains(string(payload), `"weight_grams":1400`) {
		t.Fatalf("public product response omits SKU weight: %s", payload)
	}
}

func TestPublicProductFromDomainStatusOverridesSkuAvailability(t *testing.T) {
	for _, status := range []string{"inactive", "out_of_stock"} {
		t.Run(status, func(t *testing.T) {
			item := productdomain.Product{
				ID:     31,
				SKU:    "STATUS-PRODUCT",
				Name:   "Status Product",
				Slug:   "status-product",
				Status: status,
				Variants: []productdomain.ProductVariant{
					{
						ID:       32,
						SKU:      "STATUS-PRODUCT-DEFAULT",
						Stock:    9,
						IsActive: true,
					},
				},
			}

			publicProduct := PublicProductFromDomain(item)
			if publicProduct.Availability != AvailabilityOutOfStock {
				t.Fatalf("expected product availability to follow status %q, got %q", status, publicProduct.Availability)
			}
			if len(publicProduct.Variants) != 1 {
				t.Fatalf("expected one public variant, got %d", len(publicProduct.Variants))
			}
			if publicProduct.Variants[0].Availability != AvailabilityOutOfStock {
				t.Fatalf("expected variant availability to follow product status %q, got %q", status, publicProduct.Variants[0].Availability)
			}

			payload, err := json.Marshal(publicProduct)
			if err != nil {
				t.Fatalf("marshal public product: %v", err)
			}
			if strings.Contains(string(payload), `"stock"`) {
				t.Fatalf("public product response exposes exact inventory: %s", payload)
			}
		})
	}
}

func TestPublicChatProductOmitsExactInventory(t *testing.T) {
	item := productdomain.Product{
		ID:    21,
		Name:  "Chat Product",
		Slug:  "chat-product",
		Stock: 28,
		Variants: []productdomain.ProductVariant{
			{
				ID:        22,
				SKU:       "CHAT-PRODUCT-DEFAULT",
				Stock:     2,
				IsDefault: true,
				IsActive:  true,
			},
		},
	}

	payload, err := json.Marshal(makePublicChatProduct(item))
	if err != nil {
		t.Fatalf("marshal public chat product: %v", err)
	}
	if strings.Contains(string(payload), `"stock"`) {
		t.Fatalf("public chat product exposes exact inventory: %s", payload)
	}
	if !strings.Contains(string(payload), `"availability":"in_stock"`) {
		t.Fatalf("public chat product omits availability: %s", payload)
	}
	if !strings.Contains(string(payload), `"sku":"CHAT-PRODUCT-DEFAULT"`) {
		t.Fatalf("public chat product omits selected SKU: %s", payload)
	}
}

func pointerTo(value uint) *uint {
	return &value
}
