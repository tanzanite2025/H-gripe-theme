package repository

import (
	"tanzanite/internal/domain/gallery"

	"gorm.io/gorm"
)

type GalleryRepository struct {
	db *gorm.DB
}

func NewGalleryRepository(db *gorm.DB) *GalleryRepository {
	return &GalleryRepository{db: db}
}

// Gallery 相关方法

// CreateGallery 创建图片库
func (r *GalleryRepository) CreateGallery(g *gallery.Gallery) error {
	return r.db.Create(g).Error
}

// FindGalleryByID 根据ID查找图片库
func (r *GalleryRepository) FindGalleryByID(id uint) (*gallery.Gallery, error) {
	var g gallery.Gallery
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("`order` ASC")
	}).First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// FindAllGalleries 查找所有图片库
func (r *GalleryRepository) FindAllGalleries(page, pageSize int) ([]gallery.Gallery, int64, error) {
	var galleries []gallery.Gallery
	var total int64

	if err := r.db.Model(&gallery.Gallery{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("`order` ASC").Limit(1) // 只加载封面图
	}).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&galleries).Error

	return galleries, total, err
}

// UpdateGallery 更新图片库
func (r *GalleryRepository) UpdateGallery(g *gallery.Gallery) error {
	return r.db.Save(g).Error
}

// DeleteGallery 删除图片库
func (r *GalleryRepository) DeleteGallery(id uint) error {
	// 先删除关联的图片
	if err := r.db.Where("gallery_id = ?", id).Delete(&gallery.GalleryImage{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&gallery.Gallery{}, id).Error
}

// GalleryImage 相关方法

// CreateGalleryImage 创建图片
func (r *GalleryRepository) CreateGalleryImage(img *gallery.GalleryImage) error {
	return r.db.Create(img).Error
}

// FindGalleryImageByID 根据ID查找图片
func (r *GalleryRepository) FindGalleryImageByID(id uint) (*gallery.GalleryImage, error) {
	var img gallery.GalleryImage
	err := r.db.First(&img, id).Error
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// FindImagesByGalleryID 查找图片库的所有图片
func (r *GalleryRepository) FindImagesByGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	var images []gallery.GalleryImage
	err := r.db.Where("gallery_id = ?", galleryID).
		Order("`order` ASC, created_at ASC").Find(&images).Error
	return images, err
}

// FindImagesByTags 根据标签查找图片
func (r *GalleryRepository) FindImagesByTags(tags []string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	var images []gallery.GalleryImage
	var total int64

	query := r.db.Model(&gallery.GalleryImage{})

	// 使用PostgreSQL的数组操作符查找包含任一标签的图片
	for i, tag := range tags {
		if i == 0 {
			query = query.Where("? = ANY(tags)", tag)
		} else {
			query = query.Or("? = ANY(tags)", tag)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&images).Error

	return images, total, err
}

// UpdateGalleryImage 更新图片
func (r *GalleryRepository) UpdateGalleryImage(img *gallery.GalleryImage) error {
	return r.db.Save(img).Error
}

// UpdateImageOrder 更新图片排序
func (r *GalleryRepository) UpdateImageOrder(id uint, order int) error {
	return r.db.Model(&gallery.GalleryImage{}).Where("id = ?", id).
		Update("`order`", order).Error
}

// DeleteGalleryImage 删除图片
func (r *GalleryRepository) DeleteGalleryImage(id uint) error {
	return r.db.Delete(&gallery.GalleryImage{}, id).Error
}

// BatchCreateImages 批量创建图片
func (r *GalleryRepository) BatchCreateImages(images []gallery.GalleryImage) error {
	return r.db.Create(&images).Error
}

// BatchDeleteImages 批量删除图片
func (r *GalleryRepository) BatchDeleteImages(ids []uint) error {
	return r.db.Delete(&gallery.GalleryImage{}, ids).Error
}

// GetImageCount 获取图片库的图片数量
func (r *GalleryRepository) GetImageCount(galleryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&gallery.GalleryImage{}).Where("gallery_id = ?", galleryID).Count(&count).Error
	return count, err
}

// SearchImages 搜索图片
func (r *GalleryRepository) SearchImages(keyword string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	var images []gallery.GalleryImage
	var total int64

	query := r.db.Model(&gallery.GalleryImage{}).
		Where("title ILIKE ? OR description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&images).Error

	return images, total, err
}
