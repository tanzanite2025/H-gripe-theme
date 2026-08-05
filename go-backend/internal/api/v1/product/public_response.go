package product

import (
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
	ID           uint                     `json:"id"`
	Name         string                   `json:"name"`
	Slug         string                   `json:"slug"`
	SKU          string                   `json:"sku,omitempty"`
	Description  string                   `json:"description"`
	ShortDesc    string                   `json:"short_description"`
	Price        float64                  `json:"price"`
	SalePrice    *float64                 `json:"sale_price"`
	MetaTitle    string                   `json:"meta_title"`
	MetaDesc     string                   `json:"meta_description"`
	Availability Availability             `json:"availability"`
	Media        []PublicProductMedia     `json:"media,omitempty"`
	ProductType  *PublicProductType       `json:"product_type,omitempty"`
	SpecValues   []PublicProductSpecValue `json:"spec_values,omitempty"`
	Variants     []PublicProductVariant   `json:"variants,omitempty"`
}

type PublicProductMedia struct {
	MediaType    string `json:"media_type"`
	Role         string `json:"role"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	PosterURL    string `json:"poster_url,omitempty"`
	Alt          string `json:"alt,omitempty"`
	Title        string `json:"title,omitempty"`
	SortOrder    int    `json:"sort_order"`
	IsPrimary    bool   `json:"is_primary"`
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
	ID           uint         `json:"id"`
	SKU          string       `json:"sku,omitempty"`
	Title        string       `json:"title"`
	OptionValues string       `json:"option_values"`
	WeightGrams  int          `json:"weight_grams,omitempty"`
	Price        float64      `json:"price"`
	SalePrice    *float64     `json:"sale_price"`
	IsDefault    bool         `json:"is_default"`
	Availability Availability `json:"availability"`
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
	price, salePrice := item.DisplayPrices()

	variants := make([]PublicProductVariant, 0, len(item.ActiveVariants()))
	productAvailable := productStatusAllowsAvailability(item.Status)
	for _, variant := range item.ActiveVariants() {
		variants = append(variants, publicProductVariantFromDomain(variant, productAvailable))
	}

	media := make([]PublicProductMedia, 0, len(item.Media))
	for _, mediaItem := range item.Media {
		if !mediaItem.IsVisible || mediaItem.URL == "" {
			continue
		}
		media = append(media, PublicProductMedia{
			MediaType:    mediaItem.MediaType,
			Role:         mediaItem.Role,
			URL:          mediaItem.URL,
			ThumbnailURL: mediaItem.ThumbnailURL,
			PosterURL:    mediaItem.PosterURL,
			Alt:          mediaItem.Alt,
			Title:        mediaItem.Title,
			SortOrder:    mediaItem.SortOrder,
			IsPrimary:    mediaItem.IsPrimary,
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
		ID:           item.ID,
		Name:         item.Name,
		Slug:         item.Slug,
		SKU:          item.DisplaySKU(),
		Description:  item.Description,
		ShortDesc:    item.ShortDesc,
		Price:        price,
		SalePrice:    salePrice,
		MetaTitle:    item.MetaTitle,
		MetaDesc:     item.MetaDesc,
		Availability: availabilityForProduct(item),
		Media:        media,
		ProductType:  publicProductTypeFromDomain(item.ProductType),
		SpecValues:   specValues,
		Variants:     variants,
	}
}

func PublicProductsFromDomain(items []productdomain.Product) []PublicProduct {
	publicItems := make([]PublicProduct, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, PublicProductFromDomain(item))
	}
	return publicItems
}

func PublicProductVariantFromDomain(item productdomain.ProductVariant) PublicProductVariant {
	return publicProductVariantFromDomain(item, true)
}

func publicProductVariantFromDomain(item productdomain.ProductVariant, productAvailable bool) PublicProductVariant {
	return PublicProductVariant{
		ID:           item.ID,
		SKU:          item.SKU,
		Title:        item.Title,
		OptionValues: item.OptionValues,
		WeightGrams:  item.Weight,
		Price:        item.Price,
		SalePrice:    item.SalePrice,
		IsDefault:    item.IsDefault,
		Availability: availabilityForVariant(item, productAvailable),
	}
}

func publicProductTypeFromDomain(item *productdomain.ProductType) *PublicProductType {
	if item == nil {
		return nil
	}

	result := &PublicProductType{
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
		SortOrder:       item.SortOrder,
		IsVisible:       item.IsVisible,
		IsFilterable:    item.IsFilterable,
		Options:         item.Options,
	}
}

func PublicProductTypesFromDomain(items []productdomain.ProductType) []PublicProductTypeIndex {
	result := make([]PublicProductTypeIndex, 0, len(items))
	for _, item := range items {
		publicType := PublicProductTypeIndex{
			ID:              item.ID,
			Name:            item.Name,
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
