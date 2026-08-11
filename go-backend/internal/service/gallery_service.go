package service

import (
	"errors"
	"strings"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
)

var (
	ErrGalleryMediaAssetRequired    = errors.New("gallery image media asset is required")
	ErrGalleryMediaAssetNotFound    = errors.New("gallery image media asset not found")
	ErrGalleryMediaAssetInvalid     = errors.New("gallery image media asset is not usable")
	ErrGalleryMediaAssetUnavailable = errors.New("gallery media asset repository is unavailable")
	ErrGalleryImageTitleRequired    = errors.New("gallery image title is required")
)

type GalleryService struct {
	repo      *repository.GalleryRepository
	mediaRepo *repository.MediaRepository
}

type GalleryAdminCreateInput struct {
	Title       string
	Description string
	Slug        string
	ProductIDs  []uint
	Images      []GalleryImageAdminCreateInput
}

type GalleryImageAdminCreateInput struct {
	MediaAssetID uint
	Title        string
	Description  string
	Tags         string
	Order        int
}

type GalleryAdminUpdateInput struct {
	Title       *string
	Description *string
	Slug        *string
	ProductIDs  *[]uint
}

type GalleryImageAdminUpdateInput struct {
	Title        *string
	Description  *string
	MediaAssetID *uint
	Tags         *string
	Order        *int
}

func NewGalleryService(repo *repository.GalleryRepository, mediaRepos ...*repository.MediaRepository) *GalleryService {
	var mediaRepo *repository.MediaRepository
	if len(mediaRepos) > 0 {
		mediaRepo = mediaRepos[0]
	}
	return &GalleryService{repo: repo, mediaRepo: mediaRepo}
}

// Gallery 相关方法

// CreateGallery 创建图片库
func (s *GalleryService) CreateGallery(g *gallery.Gallery) error {
	return s.repo.CreateGallery(g)
}

func (s *GalleryService) CreateAdminGallery(input GalleryAdminCreateInput) (*gallery.Gallery, error) {
	newGallery := &gallery.Gallery{
		Name:        input.Title,
		Description: input.Description,
		Slug:        input.Slug,
	}

	images, err := s.prepareAdminGalleryImages(input.Images)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateGalleryWithProductLinksAndImages(newGallery, input.ProductIDs, images); err != nil {
		return nil, err
	}

	return s.GetGalleryByID(newGallery.ID)
}

// GetGalleryByID 根据ID获取图片库
func (s *GalleryService) GetGalleryByID(id uint) (*gallery.Gallery, error) {
	return s.repo.FindGalleryByID(id)
}

func (s *GalleryService) GetPublicGalleryByID(id uint) (*gallery.Gallery, error) {
	galleryItem, err := s.repo.FindGalleryByIDAndStatus(id, "published")
	if err != nil {
		if IsRecordNotFound(err) {
			return nil, ErrGalleryNotFound
		}
		return nil, err
	}
	return galleryItem, nil
}

func (s *GalleryService) GetPublicGalleryBySlug(slug string) (*gallery.Gallery, error) {
	galleryItem, err := s.repo.FindGalleryBySlugAndStatus(slug, "published")
	if err != nil {
		if IsRecordNotFound(err) {
			return nil, ErrGalleryNotFound
		}
		return nil, err
	}
	return galleryItem, nil
}

// GetAllGalleries 获取所有图片库
func (s *GalleryService) GetAllGalleries(page, pageSize int) ([]gallery.Gallery, int64, error) {
	return s.repo.FindAllGalleries(page, pageSize)
}

func (s *GalleryService) GetPublicGalleries(page, pageSize int) ([]gallery.Gallery, int64, error) {
	return s.repo.FindAllGalleriesByStatus("published", page, pageSize)
}

// UpdateGallery 更新图片库
func (s *GalleryService) UpdateGallery(g *gallery.Gallery) error {
	return s.repo.UpdateGallery(g)
}

func (s *GalleryService) UpdateAdminGallery(id uint, input GalleryAdminUpdateInput) (*gallery.Gallery, error) {
	existingGallery, err := s.GetGalleryByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil && *input.Title != "" {
		existingGallery.Name = *input.Title
	}
	if input.Description != nil {
		existingGallery.Description = *input.Description
	}
	if input.Slug != nil && *input.Slug != "" {
		existingGallery.Slug = *input.Slug
	}

	if input.ProductIDs != nil {
		if err := s.repo.UpdateGalleryWithProductLinks(existingGallery, *input.ProductIDs); err != nil {
			return nil, err
		}
	} else if err := s.repo.UpdateGallery(existingGallery); err != nil {
		return nil, err
	}

	return s.GetGalleryByID(id)
}

// DeleteGallery 删除图片库
func (s *GalleryService) DeleteGallery(id uint) error {
	return s.repo.DeleteGallery(id)
}

func (s *GalleryService) ReplaceGalleryProductLinks(id uint, productIDs []uint) error {
	return s.repo.ReplaceGalleryProductLinks(id, productIDs)
}

// GalleryImage 相关方法

// CreateGalleryImage 创建图片
func (s *GalleryService) CreateGalleryImage(img *gallery.GalleryImage) error {
	if _, err := s.GetGalleryByID(img.GalleryID); err != nil {
		return err
	}
	if img.MediaAssetID == nil || *img.MediaAssetID == 0 {
		return ErrGalleryMediaAssetRequired
	}
	asset, err := s.findUsableMediaAsset(*img.MediaAssetID)
	if err != nil {
		return err
	}
	img.URL = asset.URL
	if img.Thumbnail == "" {
		img.Thumbnail = asset.URL
	}
	return s.repo.CreateGalleryImage(img)
}

func (s *GalleryService) prepareAdminGalleryImages(inputs []GalleryImageAdminCreateInput) ([]gallery.GalleryImage, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	images := make([]gallery.GalleryImage, 0, len(inputs))
	for _, input := range inputs {
		if input.MediaAssetID == 0 {
			return nil, ErrGalleryMediaAssetRequired
		}
		if strings.TrimSpace(input.Title) == "" {
			return nil, ErrGalleryImageTitleRequired
		}

		asset, err := s.findUsableMediaAsset(input.MediaAssetID)
		if err != nil {
			return nil, err
		}

		mediaAssetID := input.MediaAssetID
		images = append(images, gallery.GalleryImage{
			MediaAssetID: &mediaAssetID,
			URL:          asset.URL,
			Thumbnail:    asset.URL,
			Title:        strings.TrimSpace(input.Title),
			Description:  input.Description,
			Tags:         input.Tags,
			Order:        input.Order,
		})
	}

	return images, nil
}

// GetGalleryImageByID 根据ID获取图片
func (s *GalleryService) GetGalleryImageByID(id uint) (*gallery.GalleryImage, error) {
	return s.repo.FindGalleryImageByID(id)
}

func (s *GalleryService) GetGalleryImageByGalleryID(galleryID, imageID uint) (*gallery.GalleryImage, error) {
	if _, err := s.GetGalleryByID(galleryID); err != nil {
		return nil, err
	}
	return s.repo.FindGalleryImageByIDAndGalleryID(imageID, galleryID)
}

// GetImagesByGalleryID 获取图片库的所有图片
func (s *GalleryService) GetImagesByGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	if _, err := s.GetGalleryByID(galleryID); err != nil {
		return nil, err
	}
	return s.repo.FindImagesByGalleryID(galleryID)
}

func (s *GalleryService) GetPublicImagesByGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	if _, err := s.GetPublicGalleryByID(galleryID); err != nil {
		return nil, err
	}
	return s.repo.FindImagesByPublishedGalleryID(galleryID)
}

// GetImagesByTags 根据标签获取图片
func (s *GalleryService) GetImagesByTags(tags []string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.FindImagesByTags(tags, page, pageSize)
}

func (s *GalleryService) GetPublicImagesByTags(tags []string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.FindImagesByTagsInPublishedGalleries(tags, page, pageSize)
}

// UpdateGalleryImage 更新图片
func (s *GalleryService) UpdateGalleryImage(img *gallery.GalleryImage) error {
	return s.repo.UpdateGalleryImage(img)
}

func (s *GalleryService) UpdateAdminGalleryImage(id uint, input GalleryImageAdminUpdateInput) (*gallery.GalleryImage, error) {
	existingImage, err := s.GetGalleryImageByID(id)
	if err != nil {
		return nil, err
	}
	return s.updateAdminGalleryImage(existingImage, input)
}

func (s *GalleryService) UpdateAdminGalleryImageForGallery(galleryID, imageID uint, input GalleryImageAdminUpdateInput) (*gallery.GalleryImage, error) {
	existingImage, err := s.GetGalleryImageByGalleryID(galleryID, imageID)
	if err != nil {
		return nil, err
	}
	return s.updateAdminGalleryImage(existingImage, input)
}

func (s *GalleryService) updateAdminGalleryImage(existingImage *gallery.GalleryImage, input GalleryImageAdminUpdateInput) (*gallery.GalleryImage, error) {
	if input.Title != nil && *input.Title != "" {
		existingImage.Title = *input.Title
	}
	if input.Description != nil {
		existingImage.Description = *input.Description
	}
	if input.MediaAssetID != nil {
		if *input.MediaAssetID == 0 {
			return nil, ErrGalleryMediaAssetRequired
		}
		asset, err := s.findUsableMediaAsset(*input.MediaAssetID)
		if err != nil {
			return nil, err
		}
		existingImage.MediaAssetID = input.MediaAssetID
		existingImage.URL = asset.URL
		existingImage.Thumbnail = asset.URL
	}
	if input.Tags != nil {
		existingImage.Tags = *input.Tags
	}
	if input.Order != nil {
		existingImage.Order = *input.Order
	}

	if err := s.UpdateGalleryImage(existingImage); err != nil {
		return nil, err
	}

	return existingImage, nil
}

func (s *GalleryService) findUsableMediaAsset(id uint) (*media.MediaAsset, error) {
	if s.mediaRepo == nil {
		return nil, ErrGalleryMediaAssetUnavailable
	}
	asset, err := s.mediaRepo.FindAssetByID(id)
	if err != nil {
		if IsRecordNotFound(err) {
			return nil, ErrGalleryMediaAssetNotFound
		}
		return nil, err
	}
	if asset.MediaType != "" && asset.MediaType != "image" {
		return nil, ErrGalleryMediaAssetInvalid
	}
	if asset.Status != "" && asset.Status != "active" {
		return nil, ErrGalleryMediaAssetInvalid
	}
	if asset.Visibility != "" && asset.Visibility != "public" {
		return nil, ErrGalleryMediaAssetInvalid
	}
	if asset.URL == "" {
		return nil, ErrGalleryMediaAssetInvalid
	}
	return asset, nil
}

// UpdateImageOrder 更新图片排序
func (s *GalleryService) UpdateImageOrder(id uint, order int) error {
	return s.repo.UpdateImageOrder(id, order)
}

// DeleteGalleryImage 删除图片
func (s *GalleryService) DeleteGalleryImage(id uint) error {
	return s.repo.DeleteGalleryImage(id)
}

func (s *GalleryService) DeleteGalleryImageForGallery(galleryID, imageID uint) error {
	if _, err := s.GetGalleryByID(galleryID); err != nil {
		return err
	}
	return s.repo.DeleteGalleryImageByIDAndGalleryID(imageID, galleryID)
}

// BatchCreateImages 批量创建图片
func (s *GalleryService) BatchCreateImages(images []gallery.GalleryImage) error {
	return s.repo.BatchCreateImages(images)
}

// BatchDeleteImages 批量删除图片
func (s *GalleryService) BatchDeleteImages(ids []uint) error {
	return s.repo.BatchDeleteImages(ids)
}

func (s *GalleryService) BatchDeleteGalleryImages(galleryID uint, ids []uint) (int64, error) {
	if _, err := s.GetGalleryByID(galleryID); err != nil {
		return 0, err
	}
	return s.repo.BatchDeleteImagesByGalleryID(galleryID, ids)
}

// GetImageCount 获取图片库的图片数量
func (s *GalleryService) GetImageCount(galleryID uint) (int64, error) {
	return s.repo.GetImageCount(galleryID)
}

// SearchImages 搜索图片
func (s *GalleryService) SearchImages(keyword string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.SearchImages(keyword, page, pageSize)
}

func (s *GalleryService) SearchPublicImages(keyword string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.SearchImagesInPublishedGalleries(keyword, page, pageSize)
}

// BatchUpdateOrder 批量更新图片排序
func (s *GalleryService) BatchUpdateOrder(orders map[uint]int) error {
	for id, order := range orders {
		if err := s.repo.UpdateImageOrder(id, order); err != nil {
			return err
		}
	}
	return nil
}
