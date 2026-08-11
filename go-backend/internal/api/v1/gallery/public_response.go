package gallery

import gallerydomain "commerce-platform/internal/domain/gallery"

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

func publicGalleryFromDomain(item *gallerydomain.Gallery) publicGalleryResponse {
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
		CoverImage:   item.CoverImage,
		Images:       item.Images,
		ProductLinks: links,
	}
}

func publicGalleriesFromDomain(items []gallerydomain.Gallery) []publicGalleryResponse {
	responses := make([]publicGalleryResponse, 0, len(items))
	for i := range items {
		responses = append(responses, publicGalleryFromDomain(&items[i]))
	}
	return responses
}
