package wishlist

import (
	"commerce-platform/internal/api/v1/publicmedia"
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

func publicWishlistResponses(items []wishlistdomain.Item, resolvers ...publicmedia.Resolver) []PublicWishlistItem {
	responses := make([]PublicWishlistItem, 0, len(items))
	for _, item := range items {
		responses = append(responses, publicWishlistResponse(item, resolvers...))
	}
	return responses
}

func publicWishlistResponse(item wishlistdomain.Item, resolvers ...publicmedia.Resolver) PublicWishlistItem {
	response := PublicWishlistItem{
		ID:        item.ID,
		ProductID: item.ProductID,
	}
	if item.Product == nil {
		return response
	}

	resolver := publicmediaResolver(resolvers)
	price, salePrice := item.Product.DisplayPrices()
	response.Product = &PublicWishlistProduct{
		ID:           item.Product.ID,
		Name:         item.Product.Name,
		Slug:         item.Product.Slug,
		Price:        price,
		SalePrice:    salePrice,
		Availability: string(wishlistAvailabilityForProduct(*item.Product)),
		Thumbnail:    wishlistThumbnail(*item.Product, resolver),
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

func wishlistThumbnail(item productdomain.Product, resolver publicmedia.Resolver) string {
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible && media.IsPrimary {
			return publicmedia.URL(resolver, wishlistMediaURL(media))
		}
	}
	for _, media := range item.Media {
		if media.MediaType == "image" && media.IsVisible {
			return publicmedia.URL(resolver, wishlistMediaURL(media))
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

func publicmediaResolver(resolvers []publicmedia.Resolver) publicmedia.Resolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}

type wishlistAvailability string

const (
	wishlistAvailabilityInStock    wishlistAvailability = "in_stock"
	wishlistAvailabilityOutOfStock wishlistAvailability = "out_of_stock"
)
