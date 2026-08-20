package product

import (
	"encoding/json"
	"strings"
	"testing"

	"commerce-platform/internal/domain/currency"
	productdomain "commerce-platform/internal/domain/product"
	"commerce-platform/internal/service"
)

func TestPublicProductFromDomainExposesPurchasableSkuFactsWithoutExactInventory(t *testing.T) {
	item := productdomain.Product{
		ID:                             11,
		ProductSpecificationTemplateID: pointerTo(uint(3)),
		ShippingTemplateID:             pointerTo(uint(7)),
		SKU:                            "PUBLIC-PRODUCT",
		Name:                           "Public Product",
		Slug:                           "public-product",
		Stock:                          47,
		Status:                         "active",
		ViewCount:                      146,
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
		`"product_specification_template_id"`,
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

func TestPublicProductSpecificationTemplateUsesBaseNameRegardlessOfLocale(t *testing.T) {
	item := productdomain.Product{
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{
			Name: "Wheelset",
		},
	}

	translated := PublicProductFromDomainWithLocale(item, "", "zh_cn")
	if translated.ProductSpecificationTemplate == nil || translated.ProductSpecificationTemplate.Name != "Wheelset" {
		t.Fatalf("expected base product specification template name, got %#v", translated.ProductSpecificationTemplate)
	}

	fallback := PublicProductFromDomainWithLocale(item, "", "fr")
	if fallback.ProductSpecificationTemplate == nil || fallback.ProductSpecificationTemplate.Name != "Wheelset" {
		t.Fatalf("expected product specification template name to remain stable, got %#v", fallback.ProductSpecificationTemplate)
	}
}

func TestPublicProductIncludesLocalizedRoutesWithoutCatalogFields(t *testing.T) {
	item := productdomain.Product{
		ID:   11,
		Name: "Localized Product",
		Slug: "localized-product",
	}
	routes := []productdomain.ProductTranslationRoute{
		{Locale: "en", Slug: "localized-product"},
		{Locale: "zh_cn", Slug: "本地化商品"},
	}

	publicProduct := PublicProductFromDomainWithLocaleAndRoutes(item, "", "en", routes)

	if len(publicProduct.LocalizedRoutes) != 2 {
		t.Fatalf("expected two localized routes, got %#v", publicProduct.LocalizedRoutes)
	}
	if publicProduct.LocalizedRoutes[1].Locale != "zh_cn" || publicProduct.LocalizedRoutes[1].Slug != "本地化商品" {
		t.Fatalf("unexpected localized route: %#v", publicProduct.LocalizedRoutes[1])
	}

	payload, err := json.Marshal(publicProduct)
	if err != nil {
		t.Fatalf("marshal public product: %v", err)
	}
	if strings.Contains(string(payload), `"parent_id"`) || strings.Contains(string(payload), `"translation_group_id"`) {
		t.Fatalf("public product response exposes translation storage fields: %s", payload)
	}
}

func TestPublicProductFromDomainExposesVariantOptionPresentationMetadata(t *testing.T) {
	optionValueID := uint(81)
	item := productdomain.Product{
		ID:     71,
		SKU:    "VISUAL-PRODUCT",
		Name:   "Visual Product",
		Slug:   "visual-product",
		Status: "active",
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{
			Name: "Finish Product",
			Slug: "finish_product",
			SpecDefinitions: []productdomain.SpecDefinition{
				{
					ID:              7,
					Group:           "Appearance",
					Name:            "Finish",
					Slug:            "finish",
					FieldType:       "select",
					Presentation:    "color",
					IsVariantOption: true,
					IsVisible:       true,
					SortOrder:       10,
				},
			},
		},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{
				ID:               optionValueID,
				SpecDefinitionID: 7,
				ValueKey:         "ruby_red",
				Label:            "Ruby Red",
				ColorHex:         "#8F2028",
				SwatchURL:        "/uploads/swatches/ruby-red.webp",
				SortOrder:        10,
				IsEnabled:        true,
			},
		},
		Media: []productdomain.ProductMedia{
			{
				ID:                   91,
				MediaType:            "image",
				Role:                 "gallery",
				VariantOptionValueID: &optionValueID,
				URL:                  "/uploads/products/ruby-red.webp",
				IsVisible:            true,
			},
		},
		Variants: []productdomain.ProductVariant{
			{
				ID:           72,
				SKU:          "VISUAL-RUBY-001",
				OptionValues: `{"finish":"ruby_red"}`,
				Price:        399,
				Stock:        5,
				IsDefault:    true,
				IsActive:     true,
			},
		},
	}

	publicProduct := PublicProductFromDomain(item)

	if len(publicProduct.VariantOptionValues) != 1 {
		t.Fatalf("expected one public option value, got %#v", publicProduct.VariantOptionValues)
	}
	option := publicProduct.VariantOptionValues[0]
	if option.SpecSlug != "finish" || option.ValueKey != "ruby_red" || option.Label != "Ruby Red" {
		t.Fatalf("unexpected public option metadata: %#v", option)
	}
	if option.ColorHex != "#8F2028" || option.SwatchURL != "/uploads/swatches/ruby-red.webp" {
		t.Fatalf("expected public swatch metadata, got %#v", option)
	}
	if len(publicProduct.Media) != 1 {
		t.Fatalf("expected one public media item, got %#v", publicProduct.Media)
	}
	if publicProduct.Media[0].VariantOptionValueID == nil || *publicProduct.Media[0].VariantOptionValueID != optionValueID {
		t.Fatalf("expected media to expose variant option value reference, got %#v", publicProduct.Media[0])
	}
	if publicProduct.ProductSpecificationTemplate == nil || len(publicProduct.ProductSpecificationTemplate.SpecDefinitions) != 1 || publicProduct.ProductSpecificationTemplate.SpecDefinitions[0].Presentation != "color" {
		t.Fatalf("expected public product specification template to expose color presentation, got %#v", publicProduct.ProductSpecificationTemplate)
	}
}

func TestPublicProductFromDomainCanonicalizesFirstPartyMediaURLs(t *testing.T) {
	resolver := service.NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30)
	item := productdomain.Product{
		ID:     91,
		SKU:    "MEDIA-PRODUCT",
		Name:   "Media Product",
		Slug:   "media-product",
		Status: "active",
		Brand: &productdomain.ProductBrand{
			Name:    "Internal Logo Brand",
			Slug:    "internal-logo-brand",
			LogoURL: "http://media.internal:8080/uploads/brands/logo.webp",
		},
		ProductSpecificationTemplate: &productdomain.ProductSpecificationTemplate{
			Name: "Wheelset",
			Slug: "wheelset",
			SpecDefinitions: []productdomain.SpecDefinition{
				{
					ID:              17,
					Name:            "Finish",
					Slug:            "finish",
					IsVisible:       true,
					IsVariantOption: true,
				},
			},
		},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{
				ID:               18,
				SpecDefinitionID: 17,
				ValueKey:         "black",
				Label:            "Black",
				SwatchURL:        "http://media.internal:8080/uploads/swatches/black.webp",
				IsEnabled:        true,
			},
		},
		Media: []productdomain.ProductMedia{
			{
				ID:           19,
				MediaType:    "image",
				URL:          "http://media.internal:8080/uploads/products/full.webp",
				ThumbnailURL: "http://media.internal:8080/uploads/products/thumb.webp",
				PosterURL:    "http://media.internal:8080/uploads/products/poster.webp",
				IsVisible:    true,
			},
		},
	}

	payload, err := json.Marshal(PublicProductFromDomain(item, resolver))
	if err != nil {
		t.Fatalf("marshal public product: %v", err)
	}
	body := string(payload)
	if strings.Contains(body, "media.internal") {
		t.Fatalf("public product response exposes internal media origin: %s", body)
	}
	for _, expected := range []string{
		`"url":"https://shop.example.test/uploads/products/full.webp"`,
		`"thumbnail_url":"https://shop.example.test/uploads/products/thumb.webp"`,
		`"poster_url":"https://shop.example.test/uploads/products/poster.webp"`,
		`"logo_url":"https://shop.example.test/uploads/brands/logo.webp"`,
		`"swatch_url":"https://shop.example.test/uploads/swatches/black.webp"`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("public product response missing canonical media %s: %s", expected, body)
		}
	}
}

func TestPublicProductDisplayPriceDoesNotUseRuntimeConversion(t *testing.T) {
	salePrice := 80.0
	item := productdomain.Product{
		ID:       41,
		SKU:      "DISPLAY-CURRENCY-PRODUCT",
		Name:     "Display Currency Product",
		Slug:     "display-currency-product",
		Currency: "USD",
		Price:    120,
		Status:   "active",
		Variants: []productdomain.ProductVariant{
			{
				ID:        42,
				SKU:       "DISPLAY-CURRENCY-VAR",
				Title:     "Default",
				Currency:  "EUR",
				Price:     100,
				SalePrice: &salePrice,
				Stock:     4,
				IsDefault: true,
				IsActive:  true,
			},
		},
	}

	publicProduct := PublicProductFromDomainWithDisplayCurrency(item, "CNY")

	if publicProduct.Currency != "EUR" {
		t.Fatalf("expected public catalog currency to follow purchasable variant, got %q", publicProduct.Currency)
	}
	if publicProduct.Price != 100 {
		t.Fatalf("expected public price to follow purchasable variant, got %.2f", publicProduct.Price)
	}
	if publicProduct.SalePrice == nil || *publicProduct.SalePrice != 80 {
		t.Fatalf("expected public sale price to follow purchasable variant, got %#v", publicProduct.SalePrice)
	}
	if publicProduct.DisplayPrice != nil {
		t.Fatalf("expected product display price to require a stored snapshot, got %#v", publicProduct.DisplayPrice)
	}
	if publicProduct.Variants[0].Currency != "EUR" || publicProduct.Variants[0].Price != 100 {
		t.Fatalf("expected variant truth price to remain unchanged, got %#v", publicProduct.Variants[0])
	}
	if publicProduct.Variants[0].DisplayPrice != nil {
		t.Fatalf("expected variant display price to require a stored snapshot, got %#v", publicProduct.Variants[0].DisplayPrice)
	}
}

func TestPublicProductDisplayPriceUsesStoredSnapshotForRequestedCurrency(t *testing.T) {
	displayPrices := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{
			Amount:        96.8,
			Currency:      "USD",
			QuoteCurrency: "USD",
			Rate:          0.1385,
			Source:        "direct_rate",
			Converted:     true,
		},
	}, "CNY")
	item := productdomain.Product{
		ID:               51,
		SKU:              "SNAPSHOT-PRODUCT",
		Name:             "Snapshot Product",
		Slug:             "snapshot-product",
		Currency:         "CNY",
		Price:            699,
		DisplayPriceData: displayPrices,
		Status:           "active",
		Variants: []productdomain.ProductVariant{
			{
				ID:               52,
				SKU:              "SNAPSHOT-VAR",
				Title:            "Default",
				Currency:         "CNY",
				Price:            699,
				DisplayPriceData: displayPrices,
				Stock:            4,
				IsDefault:        true,
				IsActive:         true,
			},
		},
	}

	publicProduct := PublicProductFromDomainWithDisplayCurrency(item, "USD")

	if publicProduct.DisplayPrice == nil {
		t.Fatal("expected product display price from stored snapshot")
	}
	if publicProduct.DisplayPrice.Currency != "USD" || publicProduct.DisplayPrice.Amount != 96.8 {
		t.Fatalf("expected stored USD display price, got %#v", publicProduct.DisplayPrice)
	}
	if len(publicProduct.DisplayPrices) != 1 || publicProduct.DisplayPrices[0].QuoteCurrency != "USD" {
		t.Fatalf("expected product display_prices to include stored USD snapshot, got %#v", publicProduct.DisplayPrices)
	}
	if len(publicProduct.Variants) != 1 || publicProduct.Variants[0].DisplayPrice == nil {
		t.Fatalf("expected variant display price from stored snapshot, got %#v", publicProduct.Variants)
	}
	if publicProduct.Variants[0].DisplayPrice.Currency != "USD" || publicProduct.Variants[0].DisplayPrice.Amount != 96.8 {
		t.Fatalf("expected variant stored USD display price, got %#v", publicProduct.Variants[0].DisplayPrice)
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
