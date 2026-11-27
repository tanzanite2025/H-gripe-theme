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
		SKU:                            "PUBLIC-PRODUCT",
		Name:                           "Public Product",
		Slug:                           "public-product",
		Stock:                          47,
		Status:                         "active",
		ViewCount:                      146,
		Variants: []productdomain.ProductVariant{
			{
				ID:                  12,
				ProductID:           11,
				ShippingTemplateIDs: []uint{7},
				SKU:                 "PUBLIC-PRODUCT-DEFAULT",
				Stock:               3,
				Weight:              1400,
				IsActive:            true,
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
		`"shipping_template_ids"`,
		`"sale_price_decimal"`,
		`"sale_price_minor"`,
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

func TestPublicProductFromDomainExposesValidatedMpnAndGtinOnlyForWheelsets(t *testing.T) {
	wheelset := productdomain.Product{
		ProductType:                   productdomain.ProductTypeWheelset,
		MerchantProductIdentifier:     "INTERNAL-MERCHANT-WHEELSET-001",
		ManufacturerProductIdentifier: " FACTORY-WHEELSET-001 ",
		StandardProductIdentifier:     " 00012345678905 ",
	}
	publicWheelset := PublicProductFromDomain(wheelset)
	if publicWheelset.MPN != "FACTORY-WHEELSET-001" {
		t.Fatalf("wheelset MPN = %q", publicWheelset.MPN)
	}
	if publicWheelset.GTIN != "00012345678905" {
		t.Fatalf("wheelset GTIN = %q", publicWheelset.GTIN)
	}

	serializedWheelset, err := json.Marshal(publicWheelset)
	if err != nil {
		t.Fatalf("marshal public wheelset: %v", err)
	}
	for _, field := range []string{`"mpn":"FACTORY-WHEELSET-001"`, `"gtin":"00012345678905"`} {
		if !strings.Contains(string(serializedWheelset), field) {
			t.Fatalf("public wheelset omits %s: %s", field, serializedWheelset)
		}
	}
	for _, field := range []string{`"merchant_product_identifier"`, `"manufacturer_product_identifier"`, `"standard_product_identifier"`, `INTERNAL-MERCHANT-WHEELSET-001`} {
		if strings.Contains(string(serializedWheelset), field) {
			t.Fatalf("public wheelset exposes raw internal field %s: %s", field, serializedWheelset)
		}
	}

	wheelset.StandardProductIdentifier = "00012345678904"
	if invalidGTIN := PublicProductFromDomain(wheelset).GTIN; invalidGTIN != "" {
		t.Fatalf("invalid wheelset GTIN was exposed: %q", invalidGTIN)
	}
	wheelset.ManufacturerProductIdentifier = ""
	wheelset.StandardProductIdentifier = ""
	serializedWheelset, err = json.Marshal(PublicProductFromDomain(wheelset))
	if err != nil {
		t.Fatalf("marshal wheelset without public identifiers: %v", err)
	}
	if strings.Contains(string(serializedWheelset), `"mpn"`) || strings.Contains(string(serializedWheelset), `"gtin"`) {
		t.Fatalf("absent identifiers must be omitted without using merchant identifier as fallback: %s", serializedWheelset)
	}

	ordinaryProduct := productdomain.Product{
		ManufacturerProductIdentifier: "FACTORY-OTHER-001",
		StandardProductIdentifier:     "4006381333931",
		WheelsetSpecification: &productdomain.WheelsetSpecification{
			SchemaVersion: 1,
			FrontWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{24}, LacingPatterns: []string{"2x"},
			},
			RearWheel: productdomain.WheelsetProductWheelSpecification{
				WheelSize: "700C", SpokeHoleCounts: []int{24}, LacingPatterns: []string{"2x"},
			},
		},
	}
	publicOrdinaryProduct := PublicProductFromDomain(ordinaryProduct)
	if publicOrdinaryProduct.MPN != "" || publicOrdinaryProduct.GTIN != "" {
		t.Fatalf("non-wheelset identifiers were exposed as WHEELSET Schema fields: %#v", publicOrdinaryProduct)
	}
	if publicOrdinaryProduct.WheelsetSpecification != nil {
		t.Fatal("non-wheelset wheelset specification was exposed in the public response")
	}
}

func TestPublicProductFromDomainExposesGenericSchemaOnlyForOrdinaryProducts(t *testing.T) {
	weight := 42.25
	ordinary := productdomain.Product{
		GenericProductSchema: &productdomain.GenericProductSchema{
			SchemaVersion: 1,
			Model:         "BC-100",
			Material:      "Aluminium",
			Color:         "Black",
			Size:          "Standard",
			WeightG:       &weight,
		},
	}
	publicOrdinary := PublicProductFromDomain(ordinary)
	if publicOrdinary.GenericProductSchema == nil || publicOrdinary.GenericProductSchema.Model != "BC-100" {
		t.Fatalf("ordinary product generic schema was not exposed: %#v", publicOrdinary.GenericProductSchema)
	}

	dedicated := ordinary
	dedicated.ProductType = productdomain.ProductTypeWheelset
	publicDedicated := PublicProductFromDomain(dedicated)
	if publicDedicated.GenericProductSchema != nil {
		t.Fatalf("dedicated product leaked ordinary generic schema: %#v", publicDedicated.GenericProductSchema)
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

func TestPublicProductFromDomainExposesMadeToOrderAvailabilityWithoutStock(t *testing.T) {
	item := productdomain.Product{
		ID:              41,
		Name:            "Custom Product",
		Slug:            "custom-product",
		Status:          "active",
		FulfillmentMode: productdomain.FulfillmentModeMadeToOrder,
		Variants: []productdomain.ProductVariant{
			{
				ID:       42,
				IsActive: true,
				Stock:    0,
			},
		},
	}

	publicProduct := PublicProductFromDomain(item)
	if publicProduct.Availability != AvailabilityMadeToOrder {
		t.Fatalf("made-to-order availability = %q", publicProduct.Availability)
	}
	if publicProduct.FulfillmentMode != productdomain.FulfillmentModeMadeToOrder {
		t.Fatalf("made-to-order fulfillment mode = %q", publicProduct.FulfillmentMode)
	}
	if len(publicProduct.Variants) != 1 || publicProduct.Variants[0].Availability != AvailabilityMadeToOrder {
		t.Fatalf("made-to-order variant availability = %#v", publicProduct.Variants)
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
					ID:           7,
					Group:        "Appearance",
					Name:         "Finish",
					Slug:         "finish",
					FieldType:    "select",
					Presentation: "color",
					Role:         "variant",
					IsVisible:    true,
					SortOrder:    10,
				},
			},
		},
		VariantOptionValues: []productdomain.ProductVariantOptionValue{
			{
				ID:                 optionValueID,
				SpecDefinitionID:   7,
				ValueKey:           "ruby_red",
				Label:              "Ruby Red",
				ColorHex:           "#8F2028",
				SwatchURL:          "/uploads/swatches/ruby-red.webp",
				SortOrder:          10,
				IsEnabled:          true,
				CustomOptionPolicy: &productdomain.ProductCustomOptionPolicy{IsDefault: true, PriceDeltaMinor: 250, InventoryPolicy: "none"},
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
				PriceMinor:   39900,
				Stock:        5,
				IsDefault:    true,
				IsActive:     true,
			},
		},
	}

	publicProduct := PublicProductFromDomain(item)
	if len(publicProduct.Media) != 1 {
		t.Fatalf("expected product media to remain public, got %#v", publicProduct.Media)
	}
	payload, err := json.Marshal(publicProduct)
	if err != nil {
		t.Fatalf("marshal public product: %v", err)
	}
	for _, retiredField := range []string{"product_specification_template", "variant_option_values", "option_value_relations", "variant_option_value_id", "option_values", "spec_definition_id", "swatch_url"} {
		if strings.Contains(string(payload), `"`+retiredField+`"`) {
			t.Fatalf("public product response exposes retired template option field %q: %s", retiredField, payload)
		}
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
			Name: "Road Bike",
			Slug: "road-bike",
			SpecDefinitions: []productdomain.SpecDefinition{
				{
					ID:        17,
					Name:      "Finish",
					Slug:      "finish",
					IsVisible: true,
					Role:      "variant",
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
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("public product response missing canonical media %s: %s", expected, body)
		}
	}
}

func TestPublicProductDisplayPriceDoesNotUseRuntimeConversion(t *testing.T) {
	item := productdomain.Product{
		ID:         41,
		SKU:        "DISPLAY-CURRENCY-PRODUCT",
		Name:       "Display Currency Product",
		Slug:       "display-currency-product",
		Currency:   "USD",
		PriceMinor: 12000,
		Status:     "active",
		Variants: []productdomain.ProductVariant{
			{
				ID:         42,
				SKU:        "DISPLAY-CURRENCY-VAR",
				Title:      "Default",
				Currency:   "EUR",
				PriceMinor: 10000,
				Stock:      4,
				IsDefault:  true,
				IsActive:   true,
			},
		},
	}

	publicProduct := PublicProductFromDomainWithDisplayCurrency(item, "CNY")

	if publicProduct.Currency != "EUR" {
		t.Fatalf("expected public catalog currency to follow purchasable variant, got %q", publicProduct.Currency)
	}
	if publicProduct.PriceDecimal != "100.00" {
		t.Fatalf("expected public price to follow purchasable variant, got %q", publicProduct.PriceDecimal)
	}
	if publicProduct.DisplayPrice != nil {
		t.Fatalf("expected product display price to require a stored snapshot, got %#v", publicProduct.DisplayPrice)
	}
	if publicProduct.Variants[0].Currency != "EUR" || publicProduct.Variants[0].PriceDecimal != "100.00" {
		t.Fatalf("expected variant truth price to remain unchanged, got %#v", publicProduct.Variants[0])
	}
	if publicProduct.Variants[0].DisplayPrice != nil {
		t.Fatalf("expected variant display price to require a stored snapshot, got %#v", publicProduct.Variants[0].DisplayPrice)
	}
}

func TestPublicProductDisplayPriceUsesStoredSnapshotForRequestedCurrency(t *testing.T) {
	displayPrices := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{
			AmountDecimal: "96.80",
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
		PriceMinor:       69900,
		DisplayPriceData: displayPrices,
		Status:           "active",
		Variants: []productdomain.ProductVariant{
			{
				ID:               52,
				SKU:              "SNAPSHOT-VAR",
				Title:            "Default",
				Currency:         "CNY",
				PriceMinor:       69900,
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
	if publicProduct.DisplayPrice.Currency != "USD" || publicProduct.DisplayPrice.AmountDecimal != "96.80" {
		t.Fatalf("expected stored USD display price, got %#v", publicProduct.DisplayPrice)
	}
	if len(publicProduct.DisplayPrices) != 1 || publicProduct.DisplayPrices[0].QuoteCurrency != "USD" {
		t.Fatalf("expected product display_prices to include stored USD snapshot, got %#v", publicProduct.DisplayPrices)
	}
	if len(publicProduct.Variants) != 1 || publicProduct.Variants[0].DisplayPrice == nil {
		t.Fatalf("expected variant display price from stored snapshot, got %#v", publicProduct.Variants)
	}
	if publicProduct.Variants[0].DisplayPrice.Currency != "USD" || publicProduct.Variants[0].DisplayPrice.AmountDecimal != "96.80" {
		t.Fatalf("expected variant stored USD display price, got %#v", publicProduct.Variants[0].DisplayPrice)
	}
	payload, err := json.Marshal(publicProduct)
	if err != nil {
		t.Fatalf("marshal display price response: %v", err)
	}
	body := string(payload)
	if !strings.Contains(body, `"amount_decimal":"96.80"`) {
		t.Fatalf("expected decimal display amount in public response: %s", body)
	}
	if strings.Contains(body, `"amount":`) {
		t.Fatalf("public display price must not expose numeric amount: %s", body)
	}
}

func TestPublicProductPriceAndSnapshotUseLowestEffectivePriceVariant(t *testing.T) {
	defaultDisplayPrices := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{AmountDecimal: "125.00", Currency: "USD", QuoteCurrency: "USD", Rate: 1.25, Source: "default_variant"},
	}, "EUR")
	startingDisplayPrices := currency.DisplayPriceSnapshotsJSON([]currency.DisplayPriceSnapshot{
		{AmountDecimal: "99.00", Currency: "USD", QuoteCurrency: "USD", Rate: 1.1, Source: "starting_variant"},
	}, "CNY")
	item := productdomain.Product{
		ID:               61,
		SKU:              "LEGACY-PRODUCT",
		Name:             "Multi Variant Product",
		Slug:             "multi-variant-product",
		Currency:         "CNY",
		PriceMinor:       99900,
		DisplayPriceData: defaultDisplayPrices,
		Status:           "active",
		Variants: []productdomain.ProductVariant{
			{
				ID:               62,
				SKU:              "DEFAULT-HIGHER",
				Title:            "Default",
				Currency:         "CNY",
				PriceMinor:       10000,
				DisplayPriceData: defaultDisplayPrices,
				Stock:            4,
				IsDefault:        true,
				IsActive:         true,
			},
			{
				ID:               63,
				SKU:              "LOWEST-EFFECTIVE",
				Title:            "Lowest effective price",
				Currency:         "CNY",
				PriceMinor:       9000,
				DisplayPriceData: startingDisplayPrices,
				Stock:            2,
				IsActive:         true,
			},
		},
	}

	publicProduct := PublicProductFromDomainWithDisplayCurrency(item, "USD")

	if publicProduct.PriceDecimal != "90.00" {
		t.Fatalf("expected catalog price from lowest active variant, got price=%v", publicProduct.PriceDecimal)
	}
	if publicProduct.Currency != "CNY" {
		t.Fatalf("expected product currency from lowest effective price variant, got %q", publicProduct.Currency)
	}
	if publicProduct.DisplayPrice == nil || publicProduct.DisplayPrice.AmountDecimal != "99.00" || publicProduct.DisplayPrice.Source != "starting_variant" {
		t.Fatalf("expected requested display price from lowest effective price variant snapshot, got %#v", publicProduct.DisplayPrice)
	}
	if len(publicProduct.DisplayPrices) != 1 || publicProduct.DisplayPrices[0].Source != "starting_variant" {
		t.Fatalf("expected product display price snapshots from lowest effective price variant, got %#v", publicProduct.DisplayPrices)
	}
	if publicProduct.SKU != "DEFAULT-HIGHER" {
		t.Fatalf("expected product SKU to remain anchored to default variant, got %q", publicProduct.SKU)
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
