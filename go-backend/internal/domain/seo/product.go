package seo

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/currency"
	productdomain "commerce-platform/internal/domain/product"
)

const (
	ProductMetaTitleSourceCustom       = "custom"
	ProductMetaTitleSourceProductName  = "product_name"
	ProductMetaDescriptionSourceCustom = "custom"
	ProductMetaDescriptionSourceShort  = "short_description"
	ProductMetaDescriptionSourceBody   = "description"
)

type ProductMetaFieldState struct {
	Value             string `json:"value"`
	Source            string `json:"source"`
	IsCustom          bool   `json:"is_custom"`
	FallbackActive    bool   `json:"fallback_active"`
	Length            int    `json:"length"`
	SoftLengthWarning bool   `json:"soft_length_warning"`
}

type ProductStructuredDataBrand struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type ProductStructuredDataOffer struct {
	Type          string  `json:"@type"`
	Price         float64 `json:"price"`
	PriceCurrency string  `json:"priceCurrency"`
	Availability  string  `json:"availability"`
	URL           string  `json:"url"`
}

type ProductStructuredDataVariant struct {
	Type          string   `json:"@type"`
	Name          string   `json:"name"`
	SKU           string   `json:"sku,omitempty"`
	Price         *float64 `json:"price,omitempty"`
	PriceCurrency string   `json:"priceCurrency,omitempty"`
	Availability  string   `json:"availability,omitempty"`
	URL           string   `json:"url"`
}

type ProductStructuredDataPreview struct {
	Context        string                         `json:"@context"`
	Type           string                         `json:"@type"`
	Name           string                         `json:"name"`
	Brand          *ProductStructuredDataBrand    `json:"brand,omitempty"`
	Description    string                         `json:"description,omitempty"`
	Image          []string                       `json:"image,omitempty"`
	SKU            string                         `json:"sku,omitempty"`
	URL            string                         `json:"url"`
	Offers         *ProductStructuredDataOffer    `json:"offers,omitempty"`
	ProductGroupID string                         `json:"productGroupID,omitempty"`
	VariesBy       []string                       `json:"variesBy,omitempty"`
	HasVariant     []ProductStructuredDataVariant `json:"hasVariant,omitempty"`
}

type ProductSEOBreadcrumbItem struct {
	Type string `json:"type"`
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
	Path string `json:"path"`
}

type ProductSEOBreadcrumbPreview struct {
	Status string                     `json:"status"`
	Reason string                     `json:"reason,omitempty"`
	Items  []ProductSEOBreadcrumbItem `json:"items"`
}

type ProductSEOPrimaryCategory struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Path string `json:"path"`
}

type ProductSEOReadiness struct {
	ProductName        string                       `json:"product_name"`
	Brand              string                       `json:"brand"`
	BrandConfigured    bool                         `json:"brand_configured"`
	SKU                string                       `json:"sku"`
	Price              *float64                     `json:"price,omitempty"`
	Currency           string                       `json:"currency"`
	Availability       string                       `json:"availability"`
	ImageCount         int                          `json:"image_count"`
	HasImage           bool                         `json:"has_image"`
	HasOffer           bool                         `json:"has_offer"`
	HasMetaTitle       bool                         `json:"has_meta_title"`
	HasMetaDescription bool                         `json:"has_meta_description"`
	ActiveVariantCount int                          `json:"active_variant_count"`
	VariantCount       int                          `json:"variant_count"`
	MetaTitle          ProductMetaFieldState        `json:"meta_title"`
	MetaDescription    ProductMetaFieldState        `json:"meta_description"`
	StructuredDataType string                       `json:"structured_data_type"`
	Ready              bool                         `json:"ready"`
	Missing            []string                     `json:"missing"`
	BlockingIssues     []string                     `json:"blocking_issues"`
	Warnings           []string                     `json:"warnings"`
	StructuredData     ProductStructuredDataPreview `json:"structured_data"`
	PrimaryCategory    *ProductSEOPrimaryCategory   `json:"primary_category,omitempty"`
	Breadcrumb         ProductSEOBreadcrumbPreview  `json:"breadcrumb"`
	BreadcrumbComplete bool                         `json:"breadcrumb_path_complete"`
	BreadcrumbSSR      string                       `json:"breadcrumb_ssr_status"`
	BreadcrumbReason   string                       `json:"breadcrumb_reason,omitempty"`
}

func BuildProductSEOReadiness(item productdomain.Product, brand, routePath string) ProductSEOReadiness {
	name := strings.TrimSpace(item.Name)
	brand = strings.TrimSpace(brand)
	activeVariants := item.ActiveVariants()
	defaultVariant := item.DefaultVariant()

	sku := strings.TrimSpace(item.SKU)
	price := item.Price
	currencyCode := currency.NormalizeCode(item.Currency)
	stock := item.Stock
	if defaultVariant != nil {
		sku = strings.TrimSpace(defaultVariant.SKU)
		price = defaultVariant.EffectivePrice()
		currencyCode = currency.NormalizeCode(defaultVariant.Currency)
		stock = defaultVariant.Stock
	}

	visibleImages := make([]string, 0, len(item.Media))
	for _, media := range item.Media {
		if !media.IsVisible || !strings.EqualFold(strings.TrimSpace(media.MediaType), "image") {
			continue
		}
		imageURL := strings.TrimSpace(media.URL)
		if imageURL == "" {
			continue
		}
		visibleImages = append(visibleImages, imageURL)
	}

	metaTitle := resolveProductMetaTitle(item.MetaTitle, name)
	metaDescription, metaDescriptionSource := resolveProductMetaDescription(item)
	titleSource := ProductMetaTitleSourceProductName
	if strings.TrimSpace(item.MetaTitle) != "" {
		titleSource = ProductMetaTitleSourceCustom
	}
	if metaTitle == "" {
		titleSource = ""
	}

	result := ProductSEOReadiness{
		ProductName:        name,
		Brand:              brand,
		BrandConfigured:    brand != "",
		SKU:                sku,
		Currency:           currencyCode,
		Availability:       resolveProductAvailability(item.Status, stock, defaultVariant != nil),
		ImageCount:         len(visibleImages),
		HasImage:           len(visibleImages) > 0,
		HasOffer:           price > 0 && currency.IsCatalogCode(currencyCode),
		HasMetaTitle:       strings.TrimSpace(item.MetaTitle) != "",
		HasMetaDescription: strings.TrimSpace(item.MetaDesc) != "",
		ActiveVariantCount: len(activeVariants),
		VariantCount:       len(item.Variants),
		MetaTitle: ProductMetaFieldState{
			Value:             metaTitle,
			Source:            titleSource,
			IsCustom:          strings.TrimSpace(item.MetaTitle) != "",
			FallbackActive:    strings.TrimSpace(item.MetaTitle) == "",
			Length:            runeLength(metaTitle),
			SoftLengthWarning: runeLength(metaTitle) > 60,
		},
		MetaDescription: ProductMetaFieldState{
			Value:             metaDescription,
			Source:            metaDescriptionSource,
			IsCustom:          strings.TrimSpace(item.MetaDesc) != "",
			FallbackActive:    strings.TrimSpace(item.MetaDesc) == "",
			Length:            runeLength(metaDescription),
			SoftLengthWarning: runeLength(metaDescription) > 160,
		},
	}
	if result.HasOffer {
		value := price
		result.Price = &value
	}

	result.Missing = make([]string, 0, 5)
	result.BlockingIssues = make([]string, 0, 5)
	result.Warnings = make([]string, 0, 5)
	if name == "" {
		result.Missing = append(result.Missing, "product_name")
		result.BlockingIssues = append(result.BlockingIssues, "product_name")
	}
	if !result.HasImage {
		result.Missing = append(result.Missing, "image")
		result.BlockingIssues = append(result.BlockingIssues, "image")
	}
	if !result.HasOffer {
		result.Missing = append(result.Missing, "price_or_currency")
		result.BlockingIssues = append(result.BlockingIssues, "price_or_currency")
	}
	if defaultVariant == nil && len(item.Variants) > 0 {
		result.Missing = append(result.Missing, "active_variant")
		result.BlockingIssues = append(result.BlockingIssues, "active_variant")
	}
	if strings.TrimSpace(item.Status) != "active" {
		result.BlockingIssues = append(result.BlockingIssues, "product_not_public")
	}
	if sku == "" {
		result.Missing = append(result.Missing, "sku")
		result.Warnings = append(result.Warnings, "sku")
	}
	if brand == "" {
		result.Missing = append(result.Missing, "brand")
		result.Warnings = append(result.Warnings, "brand")
	}
	if result.MetaTitle.FallbackActive {
		result.Warnings = append(result.Warnings, "meta_title_fallback")
	}
	if result.MetaDescription.FallbackActive {
		result.Warnings = append(result.Warnings, "meta_description_fallback")
	}
	if result.MetaTitle.SoftLengthWarning {
		result.Warnings = append(result.Warnings, "meta_title_length")
	}
	if result.MetaDescription.SoftLengthWarning {
		result.Warnings = append(result.Warnings, "meta_description_length")
	}

	result.StructuredData = buildProductStructuredDataPreview(
		item,
		brand,
		routePath,
		name,
		metaDescription,
		sku,
		price,
		currencyCode,
		result.Availability,
		visibleImages,
		activeVariants,
	)
	result.StructuredDataType = result.StructuredData.Type
	result.Ready = len(result.BlockingIssues) == 0
	return result
}

func buildProductStructuredDataPreview(
	item productdomain.Product,
	brand,
	routePath,
	name,
	description,
	sku string,
	price float64,
	currencyCode,
	availability string,
	images []string,
	activeVariants []productdomain.ProductVariant,
) ProductStructuredDataPreview {
	preview := ProductStructuredDataPreview{
		Context:     "https://schema.org",
		Type:        "Product",
		Name:        name,
		Description: description,
		Image:       images,
		SKU:         sku,
		URL:         routePath,
	}
	if brand != "" {
		preview.Brand = &ProductStructuredDataBrand{
			Type: "Brand",
			Name: brand,
		}
	}
	if len(activeVariants) >= 2 {
		preview.Type = "ProductGroup"
		preview.ProductGroupID = "product-" + strconv.FormatUint(uint64(item.ID), 10)
		preview.HasVariant = make([]ProductStructuredDataVariant, 0, len(activeVariants))
		for _, variant := range activeVariants {
			variantPrice := variant.EffectivePrice()
			variantCurrency := currency.NormalizeCode(variant.Currency)
			variantAvailability := "https://schema.org/OutOfStock"
			if variant.Stock > 0 {
				variantAvailability = "https://schema.org/InStock"
			}
			preview.HasVariant = append(preview.HasVariant, ProductStructuredDataVariant{
				Type:          "Product",
				Name:          firstNonEmpty(strings.TrimSpace(variant.Title), strings.TrimSpace(variant.SKU), name),
				SKU:           strings.TrimSpace(variant.SKU),
				Price:         positiveFloatPointer(variantPrice),
				PriceCurrency: variantCurrency,
				Availability:  variantAvailability,
				URL:           variantLandingPath(routePath, variant.ID),
			})
		}
		return preview
	}
	if price > 0 && currency.IsCatalogCode(currencyCode) && availability != "unknown" {
		preview.Offers = &ProductStructuredDataOffer{
			Type:          "Offer",
			Price:         price,
			PriceCurrency: currencyCode,
			Availability:  productSchemaAvailability(availability),
			URL:           routePath,
		}
	}
	return preview
}

func resolveProductMetaTitle(metaTitle, name string) string {
	if value := strings.TrimSpace(metaTitle); value != "" {
		return value
	}
	return strings.TrimSpace(name)
}

func resolveProductMetaDescription(item productdomain.Product) (string, string) {
	if value := cleanSEOText(item.MetaDesc); value != "" {
		return value, ProductMetaDescriptionSourceCustom
	}
	if value := cleanSEOText(item.ShortDesc); value != "" {
		return value, ProductMetaDescriptionSourceShort
	}
	if value := cleanSEOText(item.Description); value != "" {
		return value, ProductMetaDescriptionSourceBody
	}
	return "", ""
}

func resolveProductAvailability(status string, stock int, hasVariant bool) string {
	if strings.TrimSpace(status) != "active" {
		return "unknown"
	}
	if hasVariant && stock > 0 {
		return "in_stock"
	}
	if stock > 0 {
		return "in_stock"
	}
	return "out_of_stock"
}

func productSchemaAvailability(value string) string {
	if value == "in_stock" {
		return "https://schema.org/InStock"
	}
	return "https://schema.org/OutOfStock"
}

func variantLandingPath(routePath string, variantID uint) string {
	if strings.TrimSpace(routePath) == "" || variantID == 0 {
		return routePath
	}
	parsed, err := url.Parse(routePath)
	if err != nil {
		return routePath
	}
	query := parsed.Query()
	query.Set("variant", strconv.FormatUint(uint64(variantID), 10))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func cleanSEOText(value string) string {
	value = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func positiveFloatPointer(value float64) *float64 {
	if value <= 0 {
		return nil
	}
	return &value
}

func runeLength(value string) int {
	return len([]rune(value))
}
