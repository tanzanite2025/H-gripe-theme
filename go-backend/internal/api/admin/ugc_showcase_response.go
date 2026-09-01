package admin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"

	"gorm.io/datatypes"
)

type adminShowcaseListItem struct {
	ID             uint           `json:"id"`
	UserID         uint           `json:"user_id"`
	Kind           string         `json:"kind"`
	Title          string         `json:"title"`
	OrderID        *uint          `json:"order_id,omitempty"`
	Region         string         `json:"region"`
	Location       string         `json:"location"`
	Nickname       string         `json:"nickname"`
	Notes          string         `json:"notes"`
	ProductRefs    datatypes.JSON `json:"product_refs"`
	GalleryImages  []string       `json:"gallery_images"`
	ImageFiles     []imageFileRef `json:"image_files"`
	ImageCount     int            `json:"image_count"`
	Status         string         `json:"status"`
	RejectedReason string         `json:"rejected_reason"`
	ApprovedAt     *time.Time     `json:"approved_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type imageFileRef struct {
	Index   int    `json:"index"`
	FileURL string `json:"file_url"`
}

func buildAdminShowcaseListItems(items []ugcshowcasedomain.UGCShowcase) []adminShowcaseListItem {
	responseItems := make([]adminShowcaseListItem, 0, len(items))
	for _, item := range items {
		galleryImages, imageFiles := buildAdminShowcaseImageFileReferences(item.ID, item.Images)
		responseItems = append(responseItems, adminShowcaseListItem{
			ID:             item.ID,
			UserID:         item.UserID,
			Kind:           item.Kind,
			Title:          item.Title,
			OrderID:        item.OrderID,
			Region:         item.Region,
			Location:       item.Location,
			Nickname:       item.Nickname,
			Notes:          item.Notes,
			ProductRefs:    item.ProductRefs,
			GalleryImages:  galleryImages,
			ImageFiles:     imageFiles,
			ImageCount:     len(galleryImages),
			Status:         item.Status,
			RejectedReason: item.RejectedReason,
			ApprovedAt:     item.ApprovedAt,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		})
	}
	return responseItems
}

func buildAdminShowcaseImageFileReferences(showcaseID uint, rawImages datatypes.JSON) ([]string, []imageFileRef) {
	imageReferences := decodeAdminShowcaseImageReferences(rawImages)
	galleryImages := make([]string, 0, len(imageReferences))
	imageFiles := make([]imageFileRef, 0, len(imageReferences))
	for index := range imageReferences {
		fileURL := fmt.Sprintf("/api/admin/showcase/%d/images/%d/file", showcaseID, index)
		galleryImages = append(galleryImages, fileURL)
		imageFiles = append(imageFiles, imageFileRef{
			Index:   index,
			FileURL: fileURL,
		})
	}
	return galleryImages, imageFiles
}

func decodeAdminShowcaseImageReferences(rawImages datatypes.JSON) []string {
	if len(rawImages) == 0 || strings.TrimSpace(string(rawImages)) == "" {
		return []string{}
	}
	var imageReferences []string
	if err := json.Unmarshal(rawImages, &imageReferences); err != nil {
		return []string{}
	}

	filtered := make([]string, 0, len(imageReferences))
	for _, imageReference := range imageReferences {
		if strings.TrimSpace(imageReference) != "" {
			filtered = append(filtered, imageReference)
		}
	}
	return filtered
}
