package service

import (
	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/repository"
)

type GalleryService struct {
	repo *repository.GalleryRepository
}

type GalleryAdminUpdateInput struct {
	Title       string
	Description string
	Slug        string
}

type GalleryImageAdminUpdateInput struct {
	Title       string
	Description string
	URL         string
	Thumbnail   string
	Tags        string
	Order       int
}

func NewGalleryService(repo *repository.GalleryRepository) *GalleryService {
	return &GalleryService{repo: repo}
}

// Gallery 相关方法

// CreateGallery 创建图片库
func (s *GalleryService) CreateGallery(g *gallery.Gallery) error {
	return s.repo.CreateGallery(g)
}

// GetGalleryByID 根据ID获取图片库
func (s *GalleryService) GetGalleryByID(id uint) (*gallery.Gallery, error) {
	return s.repo.FindGalleryByID(id)
}

// GetAllGalleries 获取所有图片库
func (s *GalleryService) GetAllGalleries(page, pageSize int) ([]gallery.Gallery, int64, error) {
	return s.repo.FindAllGalleries(page, pageSize)
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

	if input.Title != "" {
		existingGallery.Name = input.Title
	}
	if input.Description != "" {
		existingGallery.Description = input.Description
	}
	if input.Slug != "" {
		existingGallery.Slug = input.Slug
	}

	if err := s.UpdateGallery(existingGallery); err != nil {
		return nil, err
	}

	return existingGallery, nil
}

// DeleteGallery 删除图片库
func (s *GalleryService) DeleteGallery(id uint) error {
	return s.repo.DeleteGallery(id)
}

// GalleryImage 相关方法

// CreateGalleryImage 创建图片
func (s *GalleryService) CreateGalleryImage(img *gallery.GalleryImage) error {
	return s.repo.CreateGalleryImage(img)
}

// GetGalleryImageByID 根据ID获取图片
func (s *GalleryService) GetGalleryImageByID(id uint) (*gallery.GalleryImage, error) {
	return s.repo.FindGalleryImageByID(id)
}

// GetImagesByGalleryID 获取图片库的所有图片
func (s *GalleryService) GetImagesByGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	return s.repo.FindImagesByGalleryID(galleryID)
}

// GetImagesByTags 根据标签获取图片
func (s *GalleryService) GetImagesByTags(tags []string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.FindImagesByTags(tags, page, pageSize)
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

	if input.Title != "" {
		existingImage.Title = input.Title
	}
	if input.Description != "" {
		existingImage.Description = input.Description
	}
	if input.URL != "" {
		existingImage.URL = input.URL
	}
	if input.Thumbnail != "" {
		existingImage.Thumbnail = input.Thumbnail
	}
	if input.Tags != "" {
		existingImage.Tags = input.Tags
	}
	existingImage.Order = input.Order

	if err := s.UpdateGalleryImage(existingImage); err != nil {
		return nil, err
	}

	return existingImage, nil
}

// UpdateImageOrder 更新图片排序
func (s *GalleryService) UpdateImageOrder(id uint, order int) error {
	return s.repo.UpdateImageOrder(id, order)
}

// DeleteGalleryImage 删除图片
func (s *GalleryService) DeleteGalleryImage(id uint) error {
	return s.repo.DeleteGalleryImage(id)
}

// BatchCreateImages 批量创建图片
func (s *GalleryService) BatchCreateImages(images []gallery.GalleryImage) error {
	return s.repo.BatchCreateImages(images)
}

// BatchDeleteImages 批量删除图片
func (s *GalleryService) BatchDeleteImages(ids []uint) error {
	return s.repo.BatchDeleteImages(ids)
}

// GetImageCount 获取图片库的图片数量
func (s *GalleryService) GetImageCount(galleryID uint) (int64, error) {
	return s.repo.GetImageCount(galleryID)
}

// SearchImages 搜索图片
func (s *GalleryService) SearchImages(keyword string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	return s.repo.SearchImages(keyword, page, pageSize)
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
