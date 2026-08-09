package product

import (
	"strings"
	"tanzanite/internal/domain/currency"
	productdomain "tanzanite/internal/domain/product"
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
	ID                  uint                              `json:"id"`
	Name                string                            `json:"name"`
	Slug                string                            `json:"slug"`
	SKU                 string                            `json:"sku,omitempty"`
	Description         string                            `json:"description"`
	ShortDesc           string                            `json:"short_description"`
	Currency            string                            `json:"currency"`
	Price               float64                           `json:"price"`
	SalePrice           *float64                          `json:"sale_price"`
	DisplayPrice        *PublicDisplayPrice               `json:"display_price,omitempty"`
	DisplayPrices       []PublicDisplayPrice              `json:"display_prices,omitempty"`
	MetaTitle           string                            `json:"meta_title"`
	MetaDesc            string                            `json:"meta_description"`
	AfterSalesTemplate  *PublicProductInformationTemplate `json:"after_sales_template,omitempty"`
	PackagingTemplate   *PublicProductInformationTemplate `json:"packaging_template,omitempty"`
	Availability        Availability                      `json:"availability"`
	Media               []PublicProductMedia              `json:"media,omitempty"`
	ProductType         *PublicProductType                `json:"product_type,omitempty"`
	SpecValues          []PublicProductSpecValue          `json:"spec_values,omitempty"`
	Variants            []PublicProductVariant            `json:"variants,omitempty"`
	VariantOptionValues []PublicVariantOptionValue        `json:"variant_option_values,omitempty"`
}

type PublicProductMedia struct {
	ID                   uint   `json:"id,omitempty"`
	MediaType            string `json:"media_type"`
	Role                 string `json:"role"`
	VariantID            *uint  `json:"variant_id,omitempty"`
	VariantOptionValueID *uint  `json:"variant_option_value_id,omitempty"`
	URL                  string `json:"url"`
	ThumbnailURL         string `json:"thumbnail_url,omitempty"`
	PosterURL            string `json:"poster_url,omitempty"`
	Alt                  string `json:"alt,omitempty"`
	Title                string `json:"title,omitempty"`
	SortOrder            int    `json:"sort_order"`
	IsPrimary            bool   `json:"is_primary"`
}

type PublicProductInformationTemplate struct {
	ID      uint   `json:"id"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Locale  string `json:"locale"`
}

type PublicProductType struct {
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

// PublicProductTypeIndex is the small catalog-taxonomy contract used by
// category navigation and filters. It does not expose the editable admin
// fields or audit trail of product types.
type PublicProductTypeIndex struct {
	ID              uint                    `json:"id"`
	Name            string                  `json:"name"`
	Slug            string                  `json:"slug"`
	SpecDefinitions []PublicProductTypeSpec `json:"spec_definitions,omitempty"`
}

type PublicProductTypeSpec struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	FieldType    string `json:"field_type"`
	Presentation string `json:"presentation"`
	IsFilterable bool   `json:"is_filterable"`
	SortOrder    int    `json:"sort_order"`
	Options      string `json:"options,omitempty"`
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

func PublicProductFromDomain(item productdomain.Product) PublicProduct {
	return PublicProductFromDomainWithLocale(item, "", "")
}

func PublicProductFromDomainWithDisplayCurrency(item productdomain.Product, displayCurrency string) PublicProduct {
	return PublicProductFromDomainWithLocale(item, displayCurrency, "")
}

func PublicProductFromDomainWithLocale(item productdomain.Product, displayCurrency, locale string) PublicProduct {
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
			URL:                  mediaItem.URL,
			ThumbnailURL:         mediaItem.ThumbnailURL,
			PosterURL:            mediaItem.PosterURL,
			Alt:                  mediaItem.Alt,
			Title:                mediaItem.Title,
			SortOrder:            mediaItem.SortOrder,
			IsPrimary:            mediaItem.IsPrimary,
		})
	}

	definitionsByID := make(map[uint]productdomain.SpecDefinition)
	if item.ProductType != nil {
		for _, definition := range item.ProductType.SpecDefinitions {
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
			SwatchURL:        optionValue.SwatchURL,
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
		ID:                  item.ID,
		Name:                item.Name,
		Slug:                item.Slug,
		SKU:                 item.DisplaySKU(),
		Description:         item.Description,
		ShortDesc:           item.ShortDesc,
		Currency:            priceCurrency,
		Price:               price,
		SalePrice:           salePrice,
		DisplayPrice:        displayPriceForCurrency(displayCurrency, displayPrices),
		DisplayPrices:       displayPrices,
		MetaTitle:           item.MetaTitle,
		MetaDesc:            item.MetaDesc,
		AfterSalesTemplate:  publicProductInformationTemplateFromDomain(item.AfterSalesTemplate),
		PackagingTemplate:   publicProductInformationTemplateFromDomain(item.PackagingTemplate),
		Availability:        availabilityForProduct(item),
		Media:               media,
		ProductType:         publicProductTypeFromDomainWithLocale(item.ProductType, locale),
		SpecValues:          specValues,
		Variants:            variants,
		VariantOptionValues: variantOptionValues,
	}
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

func PublicProductsFromDomain(items []productdomain.Product) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, "", "")
}

func PublicProductsFromDomainWithDisplayCurrency(items []productdomain.Product, displayCurrency string) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, displayCurrency, "")
}

func PublicProductsFromDomainWithLocale(items []productdomain.Product, locale string) []PublicProduct {
	return PublicProductsFromDomainWithLocaleAndDisplayCurrency(items, "", locale)
}

func PublicProductsFromDomainWithLocaleAndDisplayCurrency(items []productdomain.Product, displayCurrency, locale string) []PublicProduct {
	publicItems := make([]PublicProduct, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, PublicProductFromDomainWithLocale(item, displayCurrency, locale))
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

func publicProductTypeFromDomain(item *productdomain.ProductType) *PublicProductType {
	return publicProductTypeFromDomainWithLocale(item, "")
}

func publicProductTypeFromDomainWithLocale(item *productdomain.ProductType, locale string) *PublicProductType {
	if item == nil {
		return nil
	}

	result := &PublicProductType{
		Name:            item.NameForLocale(locale),
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

func PublicProductTypesFromDomain(items []productdomain.ProductType) []PublicProductTypeIndex {
	return PublicProductTypesFromDomainWithLocale(items, "")
}

func PublicProductTypesFromDomainWithLocale(items []productdomain.ProductType, locale string) []PublicProductTypeIndex {
	result := make([]PublicProductTypeIndex, 0, len(items))
	for _, item := range items {
		publicType := PublicProductTypeIndex{
			ID:              item.ID,
			Name:            item.NameForLocale(locale),
			Slug:            item.Slug,
			SpecDefinitions: make([]PublicProductTypeSpec, 0, len(item.SpecDefinitions)),
		}
		for _, definition := range item.SpecDefinitions {
			if !definition.IsVisible {
				continue
			}
			publicType.SpecDefinitions = append(publicType.SpecDefinitions, PublicProductTypeSpec{
				ID:           definition.ID,
				Name:         definition.Name,
				Slug:         definition.Slug,
				FieldType:    definition.FieldType,
				Presentation: definition.Presentation,
				IsFilterable: definition.IsFilterable,
				SortOrder:    definition.SortOrder,
				Options:      definition.Options,
			})
		}
		result = append(result, publicType)
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
