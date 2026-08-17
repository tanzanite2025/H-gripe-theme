package cart

import (
	"commerce-platform/internal/api/v1/publicmedia"
	productdomain "commerce-platform/internal/domain/product"
)

type PublicCartSummary struct {
	ItemCount int              `json:"item_count"`
	Total     float64          `json:"total"`
	Items     []PublicCartItem `json:"items"`
}

type PublicCartItem struct {
	ID        uint               `json:"id"`
	CartID    uint               `json:"cart_id"`
	ProductID uint               `json:"product_id"`
	VariantID *uint              `json:"variant_id"`
	Quantity  int                `json:"quantity"`
	Price     float64            `json:"price"`
	Currency  string             `json:"currency"`
	Product   *PublicCartProduct `json:"product,omitempty"`
	Variant   *PublicCartVariant `json:"variant,omitempty"`
}

type PublicCartProduct struct {
	ID           uint              `json:"id"`
	Name         string            `json:"name"`
	Slug         string            `json:"slug"`
	ShortDesc    string            `json:"short_description"`
	Price        float64           `json:"price"`
	SalePrice    *float64          `json:"sale_price"`
	Availability string            `json:"availability"`
	Media        []PublicCartMedia `json:"media,omitempty"`
}

type PublicCartVariant struct {
	ID           uint     `json:"id"`
	ProductID    uint     `json:"product_id"`
	Title        string   `json:"title"`
	OptionValues string   `json:"option_values"`
	Price        float64  `json:"price"`
	SalePrice    *float64 `json:"sale_price"`
	IsDefault    bool     `json:"is_default"`
	Availability string   `json:"availability"`
}

type PublicCartMedia struct {
	MediaType     string                            `json:"media_type"`
	Role          string                            `json:"role"`
	URL           string                            `json:"url"`
	ThumbnailURL  string                            `json:"thumbnail_url,omitempty"`
	PosterURL     string                            `json:"poster_url,omitempty"`
	ImageVariants map[string]PublicCartImageVariant `json:"image_variants,omitempty"`
	Alt           string                            `json:"alt,omitempty"`
	Title         string                            `json:"title,omitempty"`
	SortOrder     int                               `json:"sort_order"`
	IsPrimary     bool                              `json:"is_primary"`
}

type PublicCartImageVariant struct {
	URL      string `json:"url"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

func PublicCartSummaryFromDomain(summary *productdomain.CartSummary, resolvers ...publicmedia.Resolver) PublicCartSummary {
	if summary == nil {
		return PublicCartSummary{Items: []PublicCartItem{}}
	}

	resolver := publicmediaResolver(resolvers)
	items := make([]PublicCartItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		publicItem := PublicCartItem{
			ID:        item.ID,
			CartID:    item.CartID,
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Currency:  item.Currency,
		}
		if item.Variant != nil {
			publicVariant := PublicCartVariant{
				ID:           item.Variant.ID,
				ProductID:    item.Variant.ProductID,
				Title:        item.Variant.Title,
				OptionValues: item.Variant.OptionValues,
				Price:        item.Variant.Price,
				SalePrice:    item.Variant.SalePrice,
				IsDefault:    item.Variant.IsDefault,
				Availability: string(cartAvailabilityForVariant(*item.Variant)),
			}
			publicItem.Variant = &publicVariant
		}
		if item.Product != nil {
			price, salePrice := item.Product.DisplayPrices()
			publicProduct := PublicCartProduct{
				ID:           item.Product.ID,
				Name:         item.Product.Name,
				Slug:         item.Product.Slug,
				ShortDesc:    item.Product.ShortDesc,
				Price:        price,
				SalePrice:    salePrice,
				Availability: string(cartAvailabilityForProduct(*item.Product)),
				Media:        publicCartMediaFromDomain(item.Product.Media, resolver),
			}
			if publicItem.Variant != nil {
				publicProduct.Availability = publicItem.Variant.Availability
			}
			publicItem.Product = &publicProduct
		}
		items = append(items, publicItem)
	}

	return PublicCartSummary{
		ItemCount: summary.ItemCount,
		Total:     summary.Total,
		Items:     items,
	}
}

func publicCartMediaFromDomain(items []productdomain.ProductMedia, resolver publicmedia.Resolver) []PublicCartMedia {
	media := make([]PublicCartMedia, 0, len(items))
	for _, item := range items {
		if !item.IsVisible || item.URL == "" {
			continue
		}
		media = append(media, PublicCartMedia{
			MediaType:     item.MediaType,
			Role:          item.Role,
			URL:           publicmedia.URL(resolver, item.URL),
			ThumbnailURL:  publicmedia.URL(resolver, item.ThumbnailURL),
			PosterURL:     publicmedia.URL(resolver, item.PosterURL),
			ImageVariants: publicCartImageVariantsFromDomain(productdomain.ParseProductMediaImageVariants(item.ImageVariantData), resolver),
			Alt:           item.Alt,
			Title:         item.Title,
			SortOrder:     item.SortOrder,
			IsPrimary:     item.IsPrimary,
		})
	}
	return media
}

func publicCartImageVariantsFromDomain(
	values map[string]productdomain.ProductMediaImageVariant,
	resolver publicmedia.Resolver,
) map[string]PublicCartImageVariant {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]PublicCartImageVariant, len(values))
	for preset, item := range values {
		if preset == "" || item.URL == "" {
			continue
		}
		result[preset] = PublicCartImageVariant{
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

func publicmediaResolver(resolvers []publicmedia.Resolver) publicmedia.Resolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}

func cartAvailabilityForProduct(item productdomain.Product) productAvailability {
	for _, variant := range item.ActiveVariants() {
		if variant.Stock > 0 {
			return productAvailabilityInStock
		}
	}
	return productAvailabilityOutOfStock
}

func cartAvailabilityForVariant(item productdomain.ProductVariant) productAvailability {
	if !item.IsActive || item.Stock <= 0 {
		return productAvailabilityOutOfStock
	}
	return productAvailabilityInStock
}

type productAvailability string

const (
	productAvailabilityInStock    productAvailability = "in_stock"
	productAvailabilityOutOfStock productAvailability = "out_of_stock"
)
