package gallery

import (
	"commerce-platform/internal/api/v1/publicmedia"
	gallerydomain "commerce-platform/internal/domain/gallery"
)

type publicGalleryProductLink struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Locale    string `json:"locale"`
}

type publicGalleryResponse struct {
	ID           uint                         `json:"id"`
	Name         string                       `json:"name"`
	Slug         string                       `json:"slug"`
	Description  string                       `json:"description"`
	CoverImage   string                       `json:"cover_image"`
	Images       []gallerydomain.GalleryImage `json:"images"`
	ProductLinks []publicGalleryProductLink   `json:"product_links"`
}

func publicGalleryFromDomain(item *gallerydomain.Gallery, resolvers ...publicmedia.Resolver) publicGalleryResponse {
	resolver := publicmediaResolver(resolvers)
	links := make([]publicGalleryProductLink, 0, len(item.ProductLinks))
	for _, link := range item.ProductLinks {
		if link.Product == nil || link.Product.Status != "active" {
			continue
		}
		links = append(links, publicGalleryProductLink{
			ProductID: link.ProductID,
			Name:      link.Product.Name,
			Slug:      link.Product.Slug,
			Locale:    link.Product.Locale,
		})
	}

	return publicGalleryResponse{
		ID:           item.ID,
		Name:         item.Name,
		Slug:         item.Slug,
		Description:  item.Description,
		CoverImage:   publicmedia.URL(resolver, item.CoverImage),
		Images:       publicGalleryImagesFromDomain(item.Images, resolver),
		ProductLinks: links,
	}
}

func publicGalleriesFromDomain(items []gallerydomain.Gallery, resolvers ...publicmedia.Resolver) []publicGalleryResponse {
	responses := make([]publicGalleryResponse, 0, len(items))
	for i := range items {
		responses = append(responses, publicGalleryFromDomain(&items[i], resolvers...))
	}
	return responses
}

func publicGalleryImagesFromDomain(items []gallerydomain.GalleryImage, resolver publicmedia.Resolver) []gallerydomain.GalleryImage {
	if len(items) == 0 {
		return items
	}

	result := make([]gallerydomain.GalleryImage, 0, len(items))
	for _, item := range items {
		publicItem := item
		publicItem.URL = publicmedia.URL(resolver, item.URL)
		publicItem.Thumbnail = publicmedia.URL(resolver, item.Thumbnail)
		result = append(result, publicItem)
	}
	return result
}

func publicmediaResolver(resolvers []publicmedia.Resolver) publicmedia.Resolver {
	if len(resolvers) == 0 {
		return nil
	}
	return resolvers[0]
}
