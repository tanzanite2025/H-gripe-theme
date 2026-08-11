package admin

import (
	"time"

	gallerydomain "tanzanite/internal/domain/gallery"
)

type adminGalleryProductLinkResponse struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Locale    string `json:"locale"`
}

type adminGalleryResponse struct {
	ID           uint                              `json:"id"`
	Name         string                            `json:"name"`
	Title        string                            `json:"title"`
	Slug         string                            `json:"slug"`
	Description  string                            `json:"description"`
	CoverImage   string                            `json:"cover_image"`
	Images       []gallerydomain.GalleryImage      `json:"images"`
	ImageCount   int64                             `json:"image_count"`
	ProductLinks []adminGalleryProductLinkResponse `json:"product_links"`
	ViewCount    int                               `json:"view_count"`
	Status       string                            `json:"status"`
	CreatedAt    time.Time                         `json:"created_at"`
	UpdatedAt    time.Time                         `json:"updated_at"`
}

func adminGalleryFromDomain(item *gallerydomain.Gallery) adminGalleryResponse {
	imageCount := item.ImageCount
	if imageCount == 0 && len(item.Images) > 0 {
		imageCount = int64(len(item.Images))
	}

	links := make([]adminGalleryProductLinkResponse, 0, len(item.ProductLinks))
	for _, link := range item.ProductLinks {
		if link.Product == nil {
			continue
		}
		links = append(links, adminGalleryProductLinkResponse{
			ProductID: link.ProductID,
			Name:      link.Product.Name,
			Slug:      link.Product.Slug,
			Locale:    link.Product.Locale,
		})
	}

	return adminGalleryResponse{
		ID:           item.ID,
		Name:         item.Name,
		Title:        item.Name,
		Slug:         item.Slug,
		Description:  item.Description,
		CoverImage:   item.CoverImage,
		Images:       item.Images,
		ImageCount:   imageCount,
		ProductLinks: links,
		ViewCount:    item.ViewCount,
		Status:       item.Status,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}

func adminGalleriesFromDomain(items []gallerydomain.Gallery) []adminGalleryResponse {
	responses := make([]adminGalleryResponse, 0, len(items))
	for i := range items {
		responses = append(responses, adminGalleryFromDomain(&items[i]))
	}
	return responses
}
