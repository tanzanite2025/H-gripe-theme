package service

import (
	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/repository"
)

type GalleryService struct {
	repo *repository.GalleryRepository
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
