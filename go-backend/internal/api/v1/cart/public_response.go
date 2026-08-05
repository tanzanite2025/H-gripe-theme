package cart

import (
	productdomain "tanzanite/internal/domain/product"
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

func PublicCartSummaryFromDomain(summary *productdomain.CartSummary) PublicCartSummary {
	if summary == nil {
		return PublicCartSummary{Items: []PublicCartItem{}}
	}

	items := make([]PublicCartItem, 0, len(summary.Items))
	for _, item := range summary.Items {
		publicItem := PublicCartItem{
			ID:        item.ID,
			CartID:    item.CartID,
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
			Price:     item.Price,
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
				Media:        publicCartMediaFromDomain(item.Product.Media),
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

func publicCartMediaFromDomain(items []productdomain.ProductMedia) []PublicCartMedia {
	media := make([]PublicCartMedia, 0, len(items))
	for _, item := range items {
		if !item.IsVisible || item.URL == "" {
			continue
		}
		media = append(media, PublicCartMedia{
			MediaType:    item.MediaType,
			Role:         item.Role,
			URL:          item.URL,
			ThumbnailURL: item.ThumbnailURL,
			PosterURL:    item.PosterURL,
			Alt:          item.Alt,
			Title:        item.Title,
			SortOrder:    item.SortOrder,
			IsPrimary:    item.IsPrimary,
		})
	}
	return media
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
