package service

import (
	"errors"
	"fmt"
	"strings"
	"tanzanite/internal/domain/product"
)

type ProductMediaInput struct {
	ID           *uint
	VariantID    *uint
	MediaAssetID *uint
	MediaType    string
	Role         string
	URL          string
	ThumbnailURL string
	PosterURL    string
	Alt          string
	Title        string
	Locale       string
	SortOrder    int
	IsPrimary    bool
	IsVisible    *bool
}

type ProductCreateInput struct {
	ProductTypeID *uint
	Name          string
	Slug          string
	Description   string
	ShortDesc     string
	Status        string
	Locale        string
	ParentID      *uint
	Featured      bool
	MetaTitle     string
	MetaDesc      string
	SpecValues    map[string]string
	Variants      []ProductVariantInput
	Media         []ProductMediaInput
}

type ProductUpdateInput struct {
	ProductTypeID       *uint
	UpdateProductTypeID bool
	Name                *string
	Slug                *string
	Description         *string
	ShortDesc           *string
	Status              *string
	Locale              *string
	ParentID            *uint
	UpdateParentID      bool
	Featured            *bool
	MetaTitle           *string
	MetaDesc            *string
	SpecValues          map[string]string
	UpdateSpecValues    bool
	Variants            []ProductVariantInput
	UpdateVariants      bool
	Media               []ProductMediaInput
	UpdateMedia         bool
}

func (s *ProductService) ListAdmin(page, pageSize int, status, locale, search, featured string) ([]product.Product, int64, error) {
	return s.productRepo.FindAllWithFilters(page, pageSize, status, locale, search, featured)
}

func (s *ProductService) GetAdminProduct(id uint) (*product.Product, error) {
	return s.findProduct(id)
}

func (s *ProductService) GetStats() (map[string]interface{}, error) {
	return s.productRepo.GetStats()
}

func (s *ProductService) CreateAdminProduct(input ProductCreateInput) (*product.Product, error) {
	specValues, err := s.buildSpecValues(input.ProductTypeID, input.SpecValues)
	if err != nil {
		return nil, err
	}

	variants, err := s.buildVariants(input.ProductTypeID, input.Variants)
	if err != nil {
		return nil, err
	}
	if err := s.ensureVariantSKUsAvailable(variants, 0); err != nil {
		return nil, err
	}
	mediaItems, err := s.buildProductMedia(input.Media)
	if err != nil {
		return nil, err
	}

	newProduct := &product.Product{
		ProductTypeID: input.ProductTypeID,
		SKU:           defaultVariantSKU(variants),
		Name:          input.Name,
		Slug:          input.Slug,
		Description:   input.Description,
		ShortDesc:     input.ShortDesc,
		Status:        input.Status,
		Locale:        input.Locale,
		ParentID:      input.ParentID,
		Featured:      input.Featured,
		MetaTitle:     input.MetaTitle,
		MetaDesc:      input.MetaDesc,
	}

	if err := s.productRepo.CreateWithSpecValuesVariantsAndMedia(newProduct, specValues, variants, mediaItems); err != nil {
		return nil, err
	}

	s.invalidateStorefrontHTMLCache("admin product create")

	return s.findProduct(newProduct.ID)
}

func (s *ProductService) UpdateAdminProduct(id uint, input ProductUpdateInput) (*product.Product, error) {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return nil, err
	}

	previousProduct := *existingProduct

	if input.UpdateProductTypeID {
		existingProduct.ProductTypeID = input.ProductTypeID
	}
	if input.Name != nil {
		existingProduct.Name = *input.Name
	}
	if input.Slug != nil {
		existingProduct.Slug = *input.Slug
	}
	if input.Description != nil {
		existingProduct.Description = *input.Description
	}
	if input.ShortDesc != nil {
		existingProduct.ShortDesc = *input.ShortDesc
	}
	if input.Status != nil {
		existingProduct.Status = *input.Status
	}
	if input.Locale != nil {
		existingProduct.Locale = *input.Locale
	}
	if input.UpdateParentID {
		existingProduct.ParentID = input.ParentID
	}
	if input.Featured != nil {
		existingProduct.Featured = *input.Featured
	}
	if input.MetaTitle != nil {
		existingProduct.MetaTitle = *input.MetaTitle
	}
	if input.MetaDesc != nil {
		existingProduct.MetaDesc = *input.MetaDesc
	}

	var specValues []product.ProductSpecValue
	if input.UpdateSpecValues {
		specValues, err = s.buildSpecValues(existingProduct.ProductTypeID, input.SpecValues)
		if err != nil {
			return nil, err
		}
	}

	var variants []product.ProductVariant
	if input.UpdateVariants {
		variants, err = s.buildVariants(existingProduct.ProductTypeID, input.Variants)
		if err != nil {
			return nil, err
		}
		if err := s.ensureVariantSKUsAvailable(variants, existingProduct.ID); err != nil {
			return nil, err
		}
	}

	var mediaItems []product.ProductMedia
	if input.UpdateMedia {
		mediaItems, err = s.buildProductMedia(input.Media)
		if err != nil {
			return nil, err
		}
	}

	if err := s.productRepo.UpdateWithSpecValuesVariantsAndMedia(existingProduct, specValues, input.UpdateSpecValues, variants, input.UpdateVariants, mediaItems, input.UpdateMedia); err != nil {
		return nil, err
	}

	s.clearProductCache(&previousProduct)
	s.clearProductCache(existingProduct)
	s.invalidateStorefrontHTMLCache("admin product update")

	return s.findProduct(existingProduct.ID)
}

func (s *ProductService) buildProductMedia(input []ProductMediaInput) ([]product.ProductMedia, error) {
	if len(input) == 0 {
		return nil, nil
	}

	items := make([]product.ProductMedia, 0, len(input))
	primaryByType := make(map[string]bool)
	for index, item := range input {
		mediaType := strings.ToLower(strings.TrimSpace(item.MediaType))
		if mediaType == "" {
			mediaType = "image"
		}
		if mediaType != "image" && mediaType != "video" {
			return nil, fmt.Errorf("%w: unsupported media type %q", ErrProductMediaInvalid, item.MediaType)
		}

		role := strings.ToLower(strings.TrimSpace(item.Role))
		if role == "" {
			if item.IsPrimary {
				role = "primary"
			} else if mediaType == "video" {
				role = "video"
			} else {
				role = "gallery"
			}
		}

		url := strings.TrimSpace(item.URL)
		if url == "" {
			return nil, fmt.Errorf("%w: media url is required", ErrProductMediaInvalid)
		}

		isVisible := true
		if item.IsVisible != nil {
			isVisible = *item.IsVisible
		}

		isPrimary := item.IsPrimary || role == "primary"
		if isPrimary {
			if primaryByType[mediaType] {
				return nil, fmt.Errorf("%w: only one primary %s media is allowed", ErrProductMediaInvalid, mediaType)
			}
			primaryByType[mediaType] = true
		}

		id := uint(0)
		if item.ID != nil {
			id = *item.ID
		}
		items = append(items, product.ProductMedia{
			ID:           id,
			VariantID:    item.VariantID,
			MediaAssetID: item.MediaAssetID,
			MediaType:    mediaType,
			Role:         role,
			URL:          url,
			ThumbnailURL: strings.TrimSpace(item.ThumbnailURL),
			PosterURL:    strings.TrimSpace(item.PosterURL),
			Alt:          strings.TrimSpace(item.Alt),
			Title:        strings.TrimSpace(item.Title),
			Locale:       strings.TrimSpace(item.Locale),
			SortOrder:    item.SortOrder,
			IsPrimary:    isPrimary,
			IsVisible:    isVisible,
		})

		if items[len(items)-1].SortOrder == 0 && index > 0 {
			items[len(items)-1].SortOrder = index * 10
		}
	}

	if !primaryByType["image"] {
		for i := range items {
			if items[i].MediaType == "image" && items[i].IsVisible {
				items[i].IsPrimary = true
				items[i].Role = "primary"
				break
			}
		}
	}

	return items, nil
}

func (s *ProductService) Delete(id uint) error {
	return s.deleteProductByID(id, true)
}

func (s *ProductService) deleteProductByID(id uint, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product delete")
	}

	return nil
}

func (s *ProductService) UpdateStatus(id uint, status string) error {
	return s.updateProductStatusByID(id, status, true)
}

func (s *ProductService) updateProductStatusByID(id uint, status string, shouldInvalidateHTML bool) error {
	existingProduct, err := s.findProduct(id)
	if err != nil {
		return err
	}

	if err := s.productRepo.UpdateStatus(id, status); err != nil {
		return err
	}

	s.clearProductCache(existingProduct)
	if shouldInvalidateHTML {
		s.invalidateStorefrontHTMLCache("admin product status update")
	}

	return nil
}

func (s *ProductService) BatchUpdateStatus(ids []uint, status string) (int, error) {
	updated := 0
	for _, id := range ids {
		if err := s.updateProductStatusByID(id, status, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return updated, err
		}
		updated++
	}
	if updated > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch status update")
	}

	return updated, nil
}

func (s *ProductService) BatchDelete(ids []uint) (int, error) {
	deleted := 0
	for _, id := range ids {
		if err := s.deleteProductByID(id, false); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				continue
			}
			return deleted, err
		}
		deleted++
	}
	if deleted > 0 {
		s.invalidateStorefrontHTMLCache("admin product batch delete")
	}

	return deleted, nil
}
