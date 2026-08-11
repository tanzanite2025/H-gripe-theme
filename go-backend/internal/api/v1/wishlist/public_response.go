package wishlist

import (
	productdomain "commerce-platform/internal/domain/product"
	wishlistdomain "commerce-platform/internal/domain/wishlist"
)

// PublicWishlistItem is the customer-facing contract for saved products.
// It deliberately excludes account metadata and exposes only the product
// fields needed to render a saved-product card and add it to the cart.
type PublicWishlistItem struct {
	ID        uint                   `json:"id"`
	ProductID uint                   `json:"product_id"`
	Product   *PublicWishlistProduct `json:"product,omitempty"`
}

type PublicWishlistProduct struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Price        float64  `json:"price"`
	SalePrice    *float64 `json:"sale_price"`
	Availability string   `json:"availability"`
	Thumbnail    string   `json:"thumbnail,omitempty"`
}

func publicWishlistResponses(items []wishlistdomain.Item) []PublicWishlistItem {
	responses := make([]PublicWishlistItem, 0, len(items))
	for _, item := range items {
		responses = append(responses, publicWishlistResponse(item))
	}
	return responses
}

func publicWishlistResponse(item wishlistdomain.Item) PublicWishlistItem {
	response := PublicWishlistItem{
		ID:        item.ID,
		ProductID: item.ProductID,
	}
	if item.Product == nil {
		return response
	}

	price, salePrice := item.Product.DisplayPrices()
	response.Product = &PublicWishlistProduct{
		ID:           item.Product.ID,
		Name:         item.Product.Name,
		Slug:         item.Product.Slug,
		Price:        price,
		SalePrice:    salePrice,
		Availability: string(wishlistAvailabilityForProduct(*item.Product)),
		Thumbnail:    wishlistThumbnail(*item.Product),
	}
	return response
}

func wishlistAvailabilityForProduct(item productdomain.Product) wishlistAvailability {
	for _, variant := range item.ActiveVariants() {
		if variant.Stock > 0 {
			return wishlistAvailabilityInStock
		}
	}
	return wishlistAvailabilityOutOfStock
}

func wishlistThumbnail(item productdomain.Product) string {
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && media.IsPrimary {
			return wishlistMediaURL(media)
		}
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible {
			return wishlistMediaURL(media)
		}
	}
	return ""
}

func wishlistMediaURL(item productdomain.ProductMedia) string {
	if item.ThumbnailURL != "" {
		return item.ThumbnailURL
	}
	return item.URL
}

type wishlistAvailability string

const (
	wishlistAvailabilityInStock    wishlistAvailability = "in_stock"
	wishlistAvailabilityOutOfStock wishlistAvailability = "out_of_stock"
)
