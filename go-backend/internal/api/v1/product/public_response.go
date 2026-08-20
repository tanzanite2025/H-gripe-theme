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
	AvailabilityInStock    Availability = "in_stock"
	AvailabilityOutOfStock Availability = "out_of_stock"
)

// PublicProduct is the storefront contract. It intentionally contains only
// fields needed to render and purchase a product. Exact inventory counters,
// shipping configuration, internal relationships, and audit timestamps are kept
// out of this response.
type PublicProduct struct {
	ID                           uint                                `json:"id"`
	Name                         string                              `json:"name"`
	Slug                         string                              `json:"slug"`
	LocalizedRoutes              []PublicProductTranslationRoute     `json:"localized_routes,omitempty"`
	SKU                          string                              `json:"sku,omitempty"`
	Description                  string                              `json:"description"`
	ShortDesc                    string                              `json:"short_description"`
	Currency                     string                              `json:"currency"`
	Price                        float64                             `json:"price"`
	SalePrice                    *float64                            `json:"sale_price"`
	DisplayPrice                 *PublicDisplayPrice                 `json:"display_price,omitempty"`
	DisplayPrices                []PublicDisplayPrice                `json:"display_prices,omitempty"`
	MetaTitle                    string                              `json:"meta_title"`
	MetaDesc                     string                              `json:"meta_description"`
	Brand                        *PublicProductBrand                 `json:"brand,omitempty"`
	AfterSalesTemplate           *PublicProductInformationTemplate   `json:"after_sales_template,omitempty"`
	PackagingTemplate            *PublicProductInformationTemplate   `json:"packaging_template,omitempty"`
	Availability                 Availability                        `json:"availability"`
	Media                        []PublicProductMedia                `json:"media,omitempty"`
	ProductSpecificationTemplate *PublicProductSpecificationTemplate `json:"product_specification_template,omitempty"`
	SpecValues                   []PublicProductSpecValue            `json:"spec_values,omitempty"`
	Variants                     []PublicProductVariant              `json:"variants,omitempty"`
	VariantOptionValues          []PublicVariantOptionValue          `json:"variant_option_values,omitempty"`
	ReviewSummary                *PublicProductReviewSummary         `json:"review_summary,omitempty"`
	ShippingDetails              *PublicProductShippingDetails       `json:"shipping_details,omitempty"`
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
	Country      string  `json:"country"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	FreeShipping bool    `json:"free_shipping"`
	EtaMinDays   int     `json:"eta_min_days"`
	EtaMaxDays   int     `json:"eta_max_days"`
}

type PublicProductMedia struct {
	ID                   uint                                 `json:"id,omitempty"`
	MediaType            string                               `json:"media_type"`
	Role                 string                               `json:"role"`
	VariantID            *uint                                `json:"variant_id,omitempty"`
	VariantOptionValueID *uint                                `json:"variant_option_value_id,omitempty"`
	URL                  string                               `json:"url"`
	ThumbnailURL         string                               `json:"thumbnail_url,omitempty"`
	PosterURL            string                               `json:"poster_url,omitempty"`
	ImageVariants        map[string]PublicProductImageVariant `json:"image_variants,omitempty"`
	Alt                  string                               `json:"alt,omitempty"`
	Title                string                               `json:"title,omitempty"`
	SortOrder            int                                  `json:"sort_order"`
	IsPrimary            bool                                 `json:"is_primary"`
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

type PublicProductSpecificationTemplate struct {
	Name            string                 `json:"name"`
	Slug            string                 `json:"slug"`
	SpecDefinitions []PublicSpecDefinition `json:"spec_definitions,omitempty"`
}

type PublicSpecDefinition struct {
	Group           string `json:"group"`
	Name            string `json:"name"`
	Slug            string `json:"slug"`
	FieldType       string `json:"field_type"`
	Unit            string `json:"unit,omitempty"`
	IsVariantOption bool   `json:"is_variant_option"`
	Presentation    string `json:"presentation"`
	SortOrder       int    `json:"sort_order"`
	IsVisible       bool   `json:"is_visible"`
	IsFilterable    bool   `json:"is_filterable"`
	Options         string `json:"options,omitempty"`
}

type PublicProductSpecValue struct {
	Value      string                `json:"value"`
	Definition *PublicSpecDefinition `json:"definition,omitempty"`
}

type PublicProductVariant struct {
	ID            uint                 `json:"id"`
	SKU           string               `json:"sku,omitempty"`
	Title         string               `json:"title"`
	OptionValues  string               `json:"option_values"`
	WeightGrams   int                  `json:"weight_grams,omitempty"`
	Currency      string               `json:"currency"`
	Price         float64              `json:"price"`
	SalePrice     *float64             `json:"sale_price"`
	DisplayPrice  *PublicDisplayPrice  `json:"display_price,omitempty"`
	DisplayPrices []PublicDisplayPrice `json:"display_prices,omitempty"`
	IsDefault     bool                 `json:"is_default"`
	Availability  Availability         `json:"availability"`
}

type PublicVariantOptionValue struct {
	ID               uint   `json:"id"`
	SpecDefinitionID uint   `json:"spec_definition_id"`
	SpecSlug         string `json:"spec_slug"`
	ValueKey         string `json:"value_key"`
	Label            string `json:"label"`
	ColorHex         string `json:"color_hex,omitempty"`
	SwatchURL        string `json:"swatch_url,omitempty"`
	SortOrder        int    `json:"sort_order"`
	IsEnabled        bool   `json:"is_enabled"`
}

type PublicDisplayPrice struct {
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	QuoteCurrency  string  `json:"quote_currency,omitempty"`
	Rate           float64 `json:"rate"`
	Source         string  `json:"source"`
	Converted      bool    `json:"converted"`
	FallbackReason string  `json:"fallback_reason,omitempty"`
}

// PublicProductSpecificationTemplateIndex is the small catalog-taxonomy contract used by
// category navigation and filters. It does not expose the editable admin
// fields or audit trail of product specification templates.
type PublicProductSpecificationTemplateIndex struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
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
	price, salePrice := item.DisplayPrices()
	priceCurrency := item.DisplayPriceCurrency()
	displayPrices := publicDisplayPricesFromSnapshots(item.DisplayPriceData)

	variants := make([]PublicProductVariant, 0, len(item.ActiveVariants()))
	productAvailable := productStatusAllowsAvailability(item.Status)
	for _, variant := range item.ActiveVariants() {
		variants = append(variants, publicProductVariantFromDomainWithDisplayCurrency(variant, productAvailable, displayCurrency))
	}

	media := make([]PublicProductMedia, 0, len(item.Media))
	for _, mediaItem := range item.Media {
		if !mediaItem.IsVisible || mediaItem.URL == "" {
			continue
		}
		media = append(media, PublicProductMedia{
			ID:                   mediaItem.ID,
			MediaType:            mediaItem.MediaType,
			Role:                 mediaItem.Role,
			VariantID:            mediaItem.VariantID,
			VariantOptionValueID: mediaItem.VariantOptionValueID,
			URL:                  publicmedia.URL(resolver, mediaItem.URL),
			ThumbnailURL:         publicmedia.URL(resolver, mediaItem.ThumbnailURL),
			PosterURL:            publicmedia.URL(resolver, mediaItem.PosterURL),
			ImageVariants:        publicProductImageVariantsFromDomain(productdomain.ParseProductMediaImageVariants(mediaItem.ImageVariantData), resolver),
			Alt:                  mediaItem.Alt,
			Title:                mediaItem.Title,
			SortOrder:            mediaItem.SortOrder,
			IsPrimary:            mediaItem.IsPrimary,
		})
	}

	definitionsByID := make(map[uint]productdomain.SpecDefinition)
	if item.ProductSpecificationTemplate != nil {
		for _, definition := range item.ProductSpecificationTemplate.SpecDefinitions {
			definitionsByID[definition.ID] = definition
		}
	}
	variantOptionValues := make([]PublicVariantOptionValue, 0, len(item.VariantOptionValues))
	for _, optionValue := range item.VariantOptionValues {
		definition, ok := definitionsByID[optionValue.SpecDefinitionID]
		if !ok || !definition.IsVisible || !definition.IsVariantOption || !optionValue.IsEnabled {
			continue
		}
		variantOptionValues = append(variantOptionValues, PublicVariantOptionValue{
			ID:               optionValue.ID,
			SpecDefinitionID: optionValue.SpecDefinitionID,
			SpecSlug:         definition.Slug,
			ValueKey:         optionValue.ValueKey,
			Label:            optionValue.Label,
			ColorHex:         optionValue.ColorHex,
			SwatchURL:        publicmedia.URL(resolver, optionValue.SwatchURL),
			SortOrder:        optionValue.SortOrder,
			IsEnabled:        optionValue.IsEnabled,
		})
	}

	specValues := make([]PublicProductSpecValue, 0, len(item.SpecValues))
	for _, value := range item.SpecValues {
		if value.SpecDefinition == nil || !value.SpecDefinition.IsVisible {
			continue
		}
		definition := publicSpecDefinitionFromDomain(*value.SpecDefinition)
		specValues = append(specValues, PublicProductSpecValue{
			Value:      value.Value,
			Definition: &definition,
		})
	}

	return PublicProduct{
		ID:                           item.ID,
		Name:                         item.Name,
		Slug:                         item.Slug,
		LocalizedRoutes:              publicProductTranslationRoutesFromDomain(translationRoutes),
		SKU:                          item.DisplaySKU(),
		Description:                  item.Description,
		ShortDesc:                    item.ShortDesc,
		Currency:                     priceCurrency,
		Price:                        price,
		SalePrice:                    salePrice,
		DisplayPrice:                 displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:                displayPrices,
		MetaTitle:                    item.MetaTitle,
		MetaDesc:                     item.MetaDesc,
		Brand:                        publicProductBrandFromDomain(item.Brand, resolver),
		AfterSalesTemplate:           publicProductInformationTemplateFromDomain(item.AfterSalesTemplate),
		PackagingTemplate:            publicProductInformationTemplateFromDomain(item.PackagingTemplate),
		Availability:                 availabilityForProduct(item),
		Media:                        media,
		ProductSpecificationTemplate: publicProductSpecificationTemplateFromDomain(item.ProductSpecificationTemplate),
		SpecValues:                   specValues,
		Variants:                     variants,
		VariantOptionValues:          variantOptionValues,
	}
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
	displayPrices := publicDisplayPricesFromSnapshots(item.DisplayPriceData)
	return PublicProductVariant{
		ID:            item.ID,
		SKU:           item.SKU,
		Title:         item.Title,
		OptionValues:  item.OptionValues,
		WeightGrams:   item.Weight,
		Currency:      item.Currency,
		Price:         item.Price,
		SalePrice:     item.SalePrice,
		DisplayPrice:  displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices: displayPrices,
		IsDefault:     item.IsDefault,
		Availability:  availabilityForVariant(item, productAvailable),
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
			Amount:         snapshot.Amount,
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

func publicProductSpecificationTemplateFromDomain(item *productdomain.ProductSpecificationTemplate) *PublicProductSpecificationTemplate {
	if item == nil {
		return nil
	}

	result := &PublicProductSpecificationTemplate{
		Name:            item.Name,
		Slug:            item.Slug,
		SpecDefinitions: make([]PublicSpecDefinition, 0, len(item.SpecDefinitions)),
	}
	for _, definition := range item.SpecDefinitions {
		if !definition.IsVisible {
			continue
		}
		result.SpecDefinitions = append(result.SpecDefinitions, publicSpecDefinitionFromDomain(definition))
	}
	return result
}

func publicSpecDefinitionFromDomain(item productdomain.SpecDefinition) PublicSpecDefinition {
	return PublicSpecDefinition{
		Group:           item.Group,
		Name:            item.Name,
		Slug:            item.Slug,
		FieldType:       item.FieldType,
		Unit:            item.Unit,
		IsVariantOption: item.IsVariantOption,
		Presentation:    item.Presentation,
		SortOrder:       item.SortOrder,
		IsVisible:       item.IsVisible,
		IsFilterable:    item.IsFilterable,
		Options:         item.Options,
	}
}

func PublicProductSpecificationTemplatesFromDomain(items []productdomain.ProductSpecificationTemplate) []PublicProductSpecificationTemplateIndex {
	return PublicProductSpecificationTemplatesFromDomainWithLocale(items, "")
}

func PublicProductSpecificationTemplatesFromDomainWithLocale(items []productdomain.ProductSpecificationTemplate, _ string) []PublicProductSpecificationTemplateIndex {
	result := make([]PublicProductSpecificationTemplateIndex, 0, len(items))
	for _, item := range items {
		publicType := PublicProductSpecificationTemplateIndex{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
		}
		result = append(result, publicType)
	}
	return result
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
	for _, variant := range item.ActiveVariants() {
		if variant.Stock > 0 {
			return AvailabilityInStock
		}
	}
	return AvailabilityOutOfStock
}

func availabilityForVariant(item productdomain.ProductVariant, productAvailable bool) Availability {
	if !productAvailable || !item.IsActive || item.Stock <= 0 {
		return AvailabilityOutOfStock
	}
	return AvailabilityInStock
}

func productStatusAllowsAvailability(status string) bool {
	return status == "" || status == "active"
}
