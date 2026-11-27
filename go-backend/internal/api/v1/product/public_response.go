package product

import (
	"strings"

	"commerce-platform/internal/api/v1/publicmedia"
	"commerce-platform/internal/domain/currency"
	productdomain "commerce-platform/internal/domain/product"
	reviewdomain "commerce-platform/internal/domain/review"
	"commerce-platform/internal/service"
)

// Availability is the coarse-grained stock state exposed to storefront clients.
// Exact inventory remains an internal value used by cart and checkout services.
type Availability string

const (
	AvailabilityInStock     Availability = "in_stock"
	AvailabilityMadeToOrder Availability = "made_to_order"
	AvailabilityOutOfStock  Availability = "out_of_stock"
)

// PublicProduct is the storefront contract. It intentionally contains only
// fields needed to render and purchase a product. Exact inventory counters,
// shipping configuration, internal relationships, and audit timestamps are kept
// out of this response.
type PublicProduct struct {
	ProductType              string                                    `json:"product_type,omitempty"`
	MPN                      string                                    `json:"mpn,omitempty"`
	GTIN                     string                                    `json:"gtin,omitempty"`
	CarbonRimSpecification   *productdomain.CarbonRimSpecification     `json:"carbon_rim_specification,omitempty"`
	BicycleTireSpecification *productdomain.BicycleTireSpecification   `json:"bicycle_tire_specification,omitempty"`
	WheelsetSpecification    *productdomain.WheelsetSpecification      `json:"wheelset_specification,omitempty"`
	SpokeSpecification       *productdomain.SpokesProductSpecification `json:"spoke_specification,omitempty"`
	GenericProductSchema     *productdomain.GenericProductSchema       `json:"generic_product_schema,omitempty"`
	ID                       uint                                      `json:"id"`
	Name                     string                                    `json:"name"`
	Slug                     string                                    `json:"slug"`
	LocalizedRoutes          []PublicProductTranslationRoute           `json:"localized_routes,omitempty"`
	Breadcrumb               *service.ProductBreadcrumbView            `json:"breadcrumb,omitempty"`
	SKU                      string                                    `json:"sku,omitempty"`
	Description              string                                    `json:"description"`
	ShortDesc                string                                    `json:"short_description"`
	Currency                 string                                    `json:"currency"`
	PriceDecimal             string                                    `json:"price_decimal"`
	DisplayPrice             *PublicDisplayPrice                       `json:"display_price,omitempty"`
	DisplayPrices            []PublicDisplayPrice                      `json:"display_prices,omitempty"`
	FulfillmentMode          string                                    `json:"fulfillment_mode,omitempty"`
	MetaTitle                string                                    `json:"meta_title"`
	MetaDesc                 string                                    `json:"meta_description"`
	Brand                    *PublicProductBrand                       `json:"brand,omitempty"`
	Category                 *PublicProductCategory                    `json:"product_category,omitempty"`
	AfterSalesTemplate       *PublicProductInformationTemplate         `json:"after_sales_template,omitempty"`
	PackagingTemplate        *PublicProductInformationTemplate         `json:"packaging_template,omitempty"`
	Availability             Availability                              `json:"availability"`
	Media                    []PublicProductMedia                      `json:"media,omitempty"`
	SpecValues               []PublicProductSpecValue                  `json:"spec_values,omitempty"`
	Variants                 []PublicProductVariant                    `json:"variants,omitempty"`
	SpokeRepairKitModels     []PublicSpokeRepairKitModel               `json:"spoke_repair_kit_models,omitempty"`
	ReviewSummary            *PublicProductReviewSummary               `json:"review_summary,omitempty"`
	ShippingDetails          *PublicProductShippingDetails             `json:"shipping_details,omitempty"`
}

// PublicSpokeRepairKitModel is the small product option contract exposed to
// buyers. Catalog verification and source timestamps remain backend data.
type PublicSpokeRepairKitModel struct {
	ID                uint   `json:"id"`
	BrandSlug         string `json:"brand_slug"`
	BrandName         string `json:"brand_name"`
	WheelsetModelSlug string `json:"wheelset_model_slug"`
	WheelsetModelName string `json:"wheelset_model_name"`
	LifecycleStatus   string `json:"lifecycle_status"`
	SortOrder         int    `json:"sort_order"`
}

type PublicProductReviewSummary struct {
	ProductID     uint    `json:"product_id"`
	TotalReviews  int     `json:"total_reviews"`
	AverageRating float64 `json:"average_rating"`
	Rating5Count  int     `json:"rating_5_count"`
	Rating4Count  int     `json:"rating_4_count"`
	Rating3Count  int     `json:"rating_3_count"`
	Rating2Count  int     `json:"rating_2_count"`
	Rating1Count  int     `json:"rating_1_count"`
}

type PublicProductShippingDetails struct {
	Country       string `json:"country"`
	AmountDecimal string `json:"amount_decimal"`
	Currency      string `json:"currency"`
	FreeShipping  bool   `json:"free_shipping"`
	EtaMinDays    int    `json:"eta_min_days"`
	EtaMaxDays    int    `json:"eta_max_days"`
}

type PublicProductMedia struct {
	ID            uint                                 `json:"id,omitempty"`
	MediaType     string                               `json:"media_type"`
	Role          string                               `json:"role"`
	VariantID     *uint                                `json:"variant_id,omitempty"`
	URL           string                               `json:"url"`
	ThumbnailURL  string                               `json:"thumbnail_url,omitempty"`
	PosterURL     string                               `json:"poster_url,omitempty"`
	ImageVariants map[string]PublicProductImageVariant `json:"image_variants,omitempty"`
	Alt           string                               `json:"alt,omitempty"`
	Title         string                               `json:"title,omitempty"`
	SortOrder     int                                  `json:"sort_order"`
	IsPrimary     bool                                 `json:"is_primary"`
}

type PublicProductImageVariant struct {
	URL      string `json:"url"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

type PublicProductTranslationRoute struct {
	Locale string `json:"locale"`
	Slug   string `json:"slug"`
}

type PublicProductInformationTemplate struct {
	ID      uint   `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Locale  string `json:"locale"`
}

type PublicProductBrand struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	LogoURL    string `json:"logo_url,omitempty"`
	WebsiteURL string `json:"website_url,omitempty"`
}

type PublicProductCategory struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PublicSpecDefinition struct {
	Group     string `json:"group"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	FieldType string `json:"field_type"`
	Unit      string `json:"unit,omitempty"`
	IsVisible bool   `json:"is_visible"`
}

type PublicProductSpecValue struct {
	Value      string                `json:"value"`
	Definition *PublicSpecDefinition `json:"definition,omitempty"`
}

type PublicProductVariant struct {
	CarbonRimSpecification *productdomain.CarbonRimVariantSpecification `json:"carbon_rim_specification,omitempty"`
	ID                     uint                                         `json:"id"`
	SKU                    string                                       `json:"sku,omitempty"`
	Title                  string                                       `json:"title"`
	WeightGrams            int                                          `json:"weight_grams,omitempty"`
	Currency               string                                       `json:"currency"`
	PriceDecimal           string                                       `json:"price_decimal"`
	DisplayPrice           *PublicDisplayPrice                          `json:"display_price,omitempty"`
	DisplayPrices          []PublicDisplayPrice                         `json:"display_prices,omitempty"`
	IsDefault              bool                                         `json:"is_default"`
	Availability           Availability                                 `json:"availability"`
}

type PublicDisplayPrice struct {
	AmountDecimal  string  `json:"amount_decimal"`
	Currency       string  `json:"currency"`
	QuoteCurrency  string  `json:"quote_currency,omitempty"`
	Rate           float64 `json:"rate"`
	Source         string  `json:"source"`
	Converted      bool    `json:"converted"`
	FallbackReason string  `json:"fallback_reason,omitempty"`
}

type PublicProductAttribute struct {
	ID           uint                          `json:"id"`
	Name         string                        `json:"name"`
	Slug         string                        `json:"slug"`
	Type         string                        `json:"type"`
	IsFilterable bool                          `json:"is_filterable"`
	IsEnabled    bool                          `json:"is_enabled"`
	SortOrder    int                           `json:"sort_order"`
	Values       []PublicProductAttributeValue `json:"values,omitempty"`
}

type PublicProductAttributeValue struct {
	ID          uint   `json:"id"`
	AttributeID uint   `json:"attribute_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Value       string `json:"value,omitempty"`
	IsEnabled   bool   `json:"is_enabled"`
	SortOrder   int    `json:"sort_order"`
}

func PublicProductFromDomain(item productdomain.Product, resolvers ...publicmedia.Resolver) PublicProduct {
	return PublicProductFromDomainWithLocale(item, "", "", resolvers...)
}

func PublicProductReviewSummaryFromDomain(item *reviewdomain.ReviewSummary) *PublicProductReviewSummary {
	if item == nil {
		return nil
	}
	return &PublicProductReviewSummary{
		ProductID:     item.ProductID,
		TotalReviews:  item.TotalReviews,
		AverageRating: item.AverageRating,
		Rating5Count:  item.Rating5Count,
		Rating4Count:  item.Rating4Count,
		Rating3Count:  item.Rating3Count,
		Rating2Count:  item.Rating2Count,
		Rating1Count:  item.Rating1Count,
	}
}

func PublicProductFromDomainWithDisplayCurrency(item productdomain.Product, displayCurrency string, resolvers ...publicmedia.Resolver) PublicProduct {
	return PublicProductFromDomainWithLocale(item, displayCurrency, "", resolvers...)
}

func PublicProductFromDomainWithLocale(item productdomain.Product, displayCurrency, locale string, resolvers ...publicmedia.Resolver) PublicProduct {
	return PublicProductFromDomainWithLocaleAndRoutes(item, displayCurrency, locale, nil, resolvers...)
}

func PublicProductFromDomainWithLocaleAndRoutes(item productdomain.Product, displayCurrency, locale string, translationRoutes []productdomain.ProductTranslationRoute, resolvers ...publicmedia.Resolver) PublicProduct {
	resolver := publicmediaResolver(resolvers)
	price := item.DisplayPriceDecimal()
	priceCurrency := item.DisplayPriceCurrency()
	displayPriceData := item.DisplayPriceData
	if startingVariant := item.StartingPriceVariant(); startingVariant != nil {
		displayPriceData = startingVariant.DisplayPriceData
	}
	displayPrices := publicDisplayPricesFromSnapshots(displayPriceData)

	variants := make([]PublicProductVariant, 0, len(item.ActiveVariants()))
	productAvailable := productStatusAllowsAvailability(item.Status)
	fulfillmentMode := productdomain.NormalizeFulfillmentMode(item.FulfillmentMode)
	for _, variant := range item.ActiveVariants() {
		variants = append(variants, publicProductVariantFromDomainWithFulfillmentMode(variant, productAvailable, displayCurrency, fulfillmentMode))
	}

	media := make([]PublicProductMedia, 0, len(item.Media))
	for _, mediaItem := range item.Media {
		if !mediaItem.IsVisible || mediaItem.URL == "" {
			continue
		}
		media = append(media, PublicProductMedia{
			ID:            mediaItem.ID,
			MediaType:     mediaItem.MediaType,
			Role:          mediaItem.Role,
			VariantID:     mediaItem.VariantID,
			URL:           publicmedia.URL(resolver, mediaItem.URL),
			ThumbnailURL:  publicmedia.URL(resolver, mediaItem.ThumbnailURL),
			PosterURL:     publicmedia.URL(resolver, mediaItem.PosterURL),
			ImageVariants: publicProductImageVariantsFromDomain(productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData), resolver),
			Alt:           mediaItem.Alt,
			Title:         mediaItem.Title,
			SortOrder:     mediaItem.SortOrder,
			IsPrimary:     mediaItem.IsPrimary,
		})
	}

	var specValues []PublicProductSpecValue
	if item.ProductType == productdomain.ProductTypeWheelset {
		specValues = publicWheelsetSpecificationValues(item.WheelsetSpecification, locale)
	} else if item.ProductType == productdomain.ProductTypeSpoke {
		specValues = publicSpokesProductSpecificationValues(item.SpokeSpecification, locale)
	} else {
		specValues = publicCarbonRimSpecificationValues(item.CarbonRimSpecification, locale)
	}

	spokeRepairKitModels := make([]PublicSpokeRepairKitModel, 0, len(item.SpokeRepairKitModels))
	for _, model := range item.SpokeRepairKitModels {
		spokeRepairKitModels = append(spokeRepairKitModels, PublicSpokeRepairKitModel{
			ID:                model.ID,
			BrandSlug:         model.BrandSlug,
			BrandName:         model.BrandName,
			WheelsetModelSlug: model.WheelsetModelSlug,
			WheelsetModelName: model.WheelsetModelName,
			LifecycleStatus:   model.LifecycleStatus,
			SortOrder:         model.SortOrder,
		})
	}
	mpn, gtin := wheelsetStructuredDataProductIdentifiers(item)
	var wheelsetSpecification *productdomain.WheelsetSpecification
	var spokeSpecification *productdomain.SpokesProductSpecification
	var genericProductSchema *productdomain.GenericProductSchema
	if item.ProductType == productdomain.ProductTypeWheelset {
		wheelsetSpecification = item.WheelsetSpecification
	} else if item.ProductType == productdomain.ProductTypeSpoke {
		spokeSpecification = item.SpokeSpecification
		mpn, gtin = spokesStructuredDataProductIdentifiers(item)
	} else if item.ProductType == "" {
		genericProductSchema = item.GenericProductSchema
	}

	return PublicProduct{
		ProductType:              item.ProductType,
		MPN:                      mpn,
		GTIN:                     gtin,
		CarbonRimSpecification:   item.CarbonRimSpecification,
		BicycleTireSpecification: item.BicycleTireSpecification,
		WheelsetSpecification:    wheelsetSpecification,
		SpokeSpecification:       spokeSpecification,
		GenericProductSchema:     genericProductSchema,
		ID:                       item.ID,
		Name:                     item.Name,
		Slug:                     item.Slug,
		LocalizedRoutes:          publicProductTranslationRoutesFromDomain(translationRoutes),
		SKU:                      item.DisplaySKU(),
		Description:              item.Description,
		ShortDesc:                item.ShortDesc,
		Currency:                 priceCurrency,
		PriceDecimal:             price,
		DisplayPrice:             displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:            displayPrices,
		FulfillmentMode:          publicFulfillmentMode(fulfillmentMode),
		MetaTitle:                item.MetaTitle,
		MetaDesc:                 item.MetaDesc,
		Brand:                    publicProductBrandFromDomain(item.Brand, resolver),
		Category:                 publicProductCategoryFromDomainWithLocale(item.ProductCategory, locale),
		AfterSalesTemplate:       publicProductInformationTemplateFromDomain(item.AfterSalesTemplate),
		PackagingTemplate:        publicProductInformationTemplateFromDomain(item.PackagingTemplate),
		Availability:             availabilityForProduct(item),
		Media:                    media,
		SpecValues:               specValues,
		Variants:                 variants,
		SpokeRepairKitModels:     spokeRepairKitModels,
	}
}

func spokesStructuredDataProductIdentifiers(item productdomain.Product) (string, string) {
	if item.ProductType != productdomain.ProductTypeSpoke || item.SpokeSpecification == nil {
		return "", ""
	}
	gtin := strings.TrimSpace(item.StandardProductIdentifier)
	if !productdomain.IsValidGS1GlobalTradeItemNumber(gtin) {
		gtin = ""
	}
	return strings.TrimSpace(item.SpokeSpecification.ManufacturerPartNumber), gtin
}

func wheelsetStructuredDataProductIdentifiers(item productdomain.Product) (string, string) {
	if item.ProductType != productdomain.ProductTypeWheelset {
		return "", ""
	}

	mpn := strings.TrimSpace(item.ManufacturerProductIdentifier)
	gtin := strings.TrimSpace(item.StandardProductIdentifier)
	if !productdomain.IsValidGS1GlobalTradeItemNumber(gtin) {
		gtin = ""
	}
	return mpn, gtin
}

func publicProductCategoryFromDomainWithLocale(item *productdomain.ProductCategory, locale string) *PublicProductCategory {
	if item == nil || strings.TrimSpace(item.Slug) == "" {
		return nil
	}
	return &PublicProductCategory{ID: item.ID, Name: item.LocalizedDisplayName(locale), Slug: item.Slug}
}

func publicProductBrandFromDomain(item *productdomain.ProductBrand, resolver publicmedia.Resolver) *PublicProductBrand {
	if item == nil || strings.TrimSpace(item.Name) == "" {
		return nil
	}
	return &PublicProductBrand{
		ID:         item.ID,
		Name:       item.Name,
		Slug:       item.Slug,
		LogoURL:    publicmedia.URL(resolver, item.LogoURL),
		WebsiteURL: strings.TrimSpace(item.WebsiteURL),
	}
}

func publicProductImageVariantsFromDomain(
	values map[string]productdomain.ProductMediaImageVariant,
	resolver publicmedia.Resolver,
) map[string]PublicProductImageVariant {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]PublicProductImageVariant, len(values))
	for preset, item := range values {
		if strings.TrimSpace(preset) == "" || strings.TrimSpace(item.URL) == "" {
			continue
		}
		result[preset] = PublicProductImageVariant{
			URL:      publicmedia.URL(resolver, item.URL),
			Width:    item.Width,
			Height:   item.Height,
			MimeType: item.MimeType,
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func publicProductInformationTemplateFromDomain(item *productdomain.ProductInformationTemplate) *PublicProductInformationTemplate {
	if item == nil || strings.TrimSpace(item.Content) == "" {
		return nil
	}
	return &PublicProductInformationTemplate{
		ID:      item.ID,
		Kind:    item.Kind,
		Name:    item.Name,
		Content: item.Content,
		Locale:  item.Locale,
	}
}

func PublicProductsFromDomain(items []productdomain.Product, resolvers ...publicmedia.Resolver) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, "", "", resolvers...)
}

func PublicProductsFromDomainWithDisplayCurrency(items []productdomain.Product, displayCurrency string, resolvers ...publicmedia.Resolver) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, displayCurrency, "", resolvers...)
}

func PublicProductsFromDomainWithLocale(items []productdomain.Product, locale string, resolvers ...publicmedia.Resolver) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, "", locale, resolvers...)
}

func PublicProductsFromDomainWithLocaleAndDisplayCurrency(items []productdomain.Product, displayCurrency, locale string, resolvers ...publicmedia.Resolver) []PublicProduct {
	publicItems := make([]PublicProduct, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, PublicProductFromDomainWithLocale(item, displayCurrency, locale, resolvers...))
	}
	return publicItems
}

func PublicProductVariantFromDomain(item productdomain.ProductVariant) PublicProductVariant {
	return publicProductVariantFromDomain(item, true)
}

func publicProductVariantFromDomain(item productdomain.ProductVariant, productAvailable bool) PublicProductVariant {
	return publicProductVariantFromDomainWithDisplayCurrency(item, productAvailable, "")
}

func publicProductVariantFromDomainWithDisplayCurrency(item productdomain.ProductVariant, productAvailable bool, displayCurrency string) PublicProductVariant {
	return publicProductVariantFromDomainWithFulfillmentMode(item, productAvailable, displayCurrency, productdomain.FulfillmentModeStock)
}

func publicProductVariantFromDomainWithFulfillmentMode(item productdomain.ProductVariant, productAvailable bool, displayCurrency, fulfillmentMode string) PublicProductVariant {
	displayPrices := publicDisplayPricesFromSnapshots(item.DisplayPriceData)
	priceMoney, _ := item.PriceMoney()
	price, _ := priceMoney.FormatMajor()
	return PublicProductVariant{
		ID:                     item.ID,
		SKU:                    item.SKU,
		Title:                  item.Title,
		CarbonRimSpecification: item.CarbonRimSpecification,
		WeightGrams:            item.Weight,
		Currency:               item.Currency,
		PriceDecimal:           price,
		DisplayPrice:           displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:          displayPrices,
		IsDefault:              item.IsDefault,
		Availability:           availabilityForVariant(item, productAvailable, fulfillmentMode),
	}
}

func publicDisplayPricesFromSnapshots(raw []byte) []PublicDisplayPrice {
	snapshots := currency.ParseDisplayPriceSnapshots(raw)
	if len(snapshots) == 0 {
		return nil
	}
	result := make([]PublicDisplayPrice, 0, len(snapshots))
	for _, snapshot := range snapshots {
		result = append(result, PublicDisplayPrice{
			AmountDecimal:  snapshot.AmountDecimal,
			Currency:       snapshot.Currency,
			QuoteCurrency:  snapshot.QuoteCurrency,
			Rate:           snapshot.Rate,
			Source:         snapshot.Source,
			Converted:      snapshot.Converted,
			FallbackReason: snapshot.FallbackReason,
		})
	}
	return result
}

func displayPriceForCurrency(displayCurrency string, displayPrices []PublicDisplayPrice) *PublicDisplayPrice {
	if displayCurrency != "" {
		for i := range displayPrices {
			if currency.NormalizeCode(displayPrices[i].Currency) == currency.NormalizeCode(displayCurrency) {
				return &displayPrices[i]
			}
		}
	}
	return nil
}

func publicmediaResolver(resolvers []publicmedia.Resolver) publicmedia.Resolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}

func PublicProductCategoryListWithMedia(
	item *service.ProductCategoryListView,
	resolver publicmedia.Resolver,
) *service.ProductCategoryListView {
	if item == nil {
		return nil
	}

	result := *item
	result.Tree = publicProductCategoryViewsWithMedia(item.Tree, resolver)
	result.Flat = publicProductCategoryViewsWithMedia(item.Flat, resolver)
	return &result
}

func PublicProductCategoryViewWithMedia(
	item service.ProductCategoryView,
	resolver publicmedia.Resolver,
) service.ProductCategoryView {
	item.ImageURL = publicmedia.URL(resolver, item.ImageURL)
	return item
}

func publicProductCategoryViewsWithMedia(
	items []service.ProductCategoryView,
	resolver publicmedia.Resolver,
) []service.ProductCategoryView {
	if len(items) == 0 {
		return items
	}

	result := make([]service.ProductCategoryView, 0, len(items))
	for _, item := range items {
		publicItem := item
		publicItem.ImageURL = publicmedia.URL(resolver, item.ImageURL)
		publicItem.Children = publicProductCategoryViewsWithMedia(item.Children, resolver)
		result = append(result, publicItem)
	}
	return result
}

func PublicProductAttributesFromDomain(items []productdomain.ProductAttribute) []PublicProductAttribute {
	result := make([]PublicProductAttribute, 0, len(items))
	for _, item := range items {
		values := make([]PublicProductAttributeValue, 0, len(item.Values))
		for _, value := range item.Values {
			if !value.IsEnabled {
				continue
			}
			values = append(values, PublicProductAttributeValue{
				ID:          value.ID,
				AttributeID: value.AttributeID,
				Name:        value.Name,
				Slug:        value.Slug,
				Value:       value.Value,
				IsEnabled:   value.IsEnabled,
				SortOrder:   value.SortOrder,
			})
		}
		result = append(result, PublicProductAttribute{
			ID:           item.ID,
			Name:         item.Name,
			Slug:         item.Slug,
			Type:         item.Type,
			IsFilterable: item.IsFilterable,
			IsEnabled:    item.IsEnabled,
			SortOrder:    item.SortOrder,
			Values:       values,
		})
	}
	return result
}

func availabilityForProduct(item productdomain.Product) Availability {
	if !productStatusAllowsAvailability(item.Status) {
		return AvailabilityOutOfStock
	}
	if productdomain.NormalizeFulfillmentMode(item.FulfillmentMode) == productdomain.FulfillmentModeMadeToOrder && len(item.ActiveVariants()) > 0 {
		return AvailabilityMadeToOrder
	}
	for _, variant := range item.ActiveVariants() {
		if variant.Stock > 0 {
			return AvailabilityInStock
		}
	}
	return AvailabilityOutOfStock
}

func availabilityForVariant(item productdomain.ProductVariant, productAvailable bool, fulfillmentMode string) Availability {
	if !productAvailable || !item.IsActive {
		return AvailabilityOutOfStock
	}
	if productdomain.NormalizeFulfillmentMode(fulfillmentMode) == productdomain.FulfillmentModeMadeToOrder {
		return AvailabilityMadeToOrder
	}
	if item.Stock <= 0 {
		return AvailabilityOutOfStock
	}
	return AvailabilityInStock
}

func productStatusAllowsAvailability(status string) bool {
	return status == "" || status == "active"
}

func publicFulfillmentMode(value string) string {
	if productdomain.NormalizeFulfillmentMode(value) == productdomain.FulfillmentModeMadeToOrder {
		return productdomain.FulfillmentModeMadeToOrder
	}
	return ""
}
