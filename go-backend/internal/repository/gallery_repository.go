package repository

import (
	"strings"
	"tanzanite/internal/domain/gallery"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *GalleryRepository) CreateGalleryWithProductLinks(g *gallery.Gallery, productIDs []uint) error {
	return r.CreateGalleryWithProductLinksAndImages(g, productIDs, nil)
}

func (r *GalleryRepository) CreateGalleryWithProductLinksAndImages(g *gallery.Gallery, productIDs []uint, images []gallery.GalleryImage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(g).Error; err != nil {
			return err
		}
		if err := r.replaceGalleryProductLinksTx(tx, g.ID, productIDs); err != nil {
			return err
		}
		if len(images) == 0 {
			return nil
		}
		for i := range images {
			images[i].GalleryID = g.ID
		}
		return tx.Create(&images).Error
	})
}

// FindGalleryByID 根据ID查找图片库
func (r *GalleryRepository) FindGalleryByID(id uint) (*gallery.Gallery, error) {
	var g gallery.Gallery
	query := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}})
	})
	query = r.preloadProductLinks(query)
	err := query.First(&g, id).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// FindGalleryByIDAndStatus 根据ID和状态查找图片库
func (r *GalleryRepository) FindGalleryByIDAndStatus(id uint, status string) (*gallery.Gallery, error) {
	var g gallery.Gallery
	query := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}})
	})
	query = r.preloadProductLinks(query)
	err := query.Where("status = ?", status).First(&g, id).Error
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
	query := r.db.Order("created_at DESC").Offset(offset).Limit(pageSize)
	query = r.preloadProductLinks(query)
	err := query.Find(&galleries).Error
	if err != nil {
		return nil, 0, err
	}
	if err := r.loadFirstImages(&galleries); err != nil {
		return nil, 0, err
	}
	if err := r.loadImageCounts(&galleries); err != nil {
		return nil, 0, err
	}

	return galleries, total, err
}

// FindAllGalleriesByStatus 查找指定状态的图片库
func (r *GalleryRepository) FindAllGalleriesByStatus(status string, page, pageSize int) ([]gallery.Gallery, int64, error) {
	var galleries []gallery.Gallery
	var total int64

	query := r.db.Model(&gallery.Gallery{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query = query.Order("created_at DESC").Offset(offset).Limit(pageSize)
	query = r.preloadProductLinks(query)
	err := query.Find(&galleries).Error
	if err != nil {
		return nil, 0, err
	}
	if err := r.loadFirstImages(&galleries); err != nil {
		return nil, 0, err
	}
	if err := r.loadImageCounts(&galleries); err != nil {
		return nil, 0, err
	}

	return galleries, total, err
}

// UpdateGallery 更新图片库
func (r *GalleryRepository) UpdateGallery(g *gallery.Gallery) error {
	return r.db.Save(g).Error
}

func (r *GalleryRepository) UpdateGalleryWithProductLinks(g *gallery.Gallery, productIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(g).Error; err != nil {
			return err
		}
		return r.replaceGalleryProductLinksTx(tx, g.ID, productIDs)
	})
}

// DeleteGallery 删除图片库
func (r *GalleryRepository) DeleteGallery(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("gallery_id = ?", id).Delete(&gallery.GalleryImage{}).Error; err != nil {
			return err
		}
		if tx.Migrator().HasTable(&gallery.GalleryProductLink{}) {
			if err := tx.Where("gallery_id = ?", id).Delete(&gallery.GalleryProductLink{}).Error; err != nil {
				return err
			}
		}
		result := tx.Delete(&gallery.Gallery{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *GalleryRepository) FindGalleryBySlugAndStatus(slug, status string) (*gallery.Gallery, error) {
	var g gallery.Gallery
	query := r.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}})
	})
	query = r.preloadProductLinks(query)
	err := query.Where("slug = ? AND status = ?", slug, status).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
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

func (r *GalleryRepository) FindGalleryImageByIDAndGalleryID(imageID, galleryID uint) (*gallery.GalleryImage, error) {
	var img gallery.GalleryImage
	err := r.db.
		Where("id = ? AND gallery_id = ?", imageID, galleryID).
		First(&img).Error
	if err != nil {
		return nil, err
	}
	return &img, nil
}

// FindImagesByGalleryID 查找图片库的所有图片
func (r *GalleryRepository) FindImagesByGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	var images []gallery.GalleryImage
	err := r.db.Where("gallery_id = ?", galleryID).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("created_at ASC").
		Find(&images).Error
	return images, err
}

// FindImagesByPublishedGalleryID 查找已发布图片库的图片
func (r *GalleryRepository) FindImagesByPublishedGalleryID(galleryID uint) ([]gallery.GalleryImage, error) {
	var images []gallery.GalleryImage
	err := r.db.Joins("JOIN galleries ON galleries.id = gallery_images.gallery_id").
		Where("gallery_images.gallery_id = ? AND galleries.status = ?", galleryID, "published").
		Order(clause.OrderByColumn{Column: clause.Column{Table: "gallery_images", Name: "order"}}).
		Order("gallery_images.created_at ASC").
		Find(&images).Error
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

// FindImagesByTagsInPublishedGalleries 根据标签查找已发布图片库中的图片
func (r *GalleryRepository) FindImagesByTagsInPublishedGalleries(tags []string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	var images []gallery.GalleryImage
	var total int64

	query := r.db.Model(&gallery.GalleryImage{}).
		Joins("JOIN galleries ON galleries.id = gallery_images.gallery_id").
		Where("galleries.status = ?", "published")

	tagConditions := make([]string, 0, len(tags))
	tagArgs := make([]interface{}, 0, len(tags))
	for _, tag := range tags {
		tagConditions = append(tagConditions, "gallery_images.tags LIKE ?")
		tagArgs = append(tagArgs, "%"+tag+"%")
	}
	if len(tagConditions) > 0 {
		query = query.Where("("+strings.Join(tagConditions, " OR ")+")", tagArgs...)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("gallery_images.created_at DESC").Offset(offset).Limit(pageSize).Find(&images).Error

	return images, total, err
}

// UpdateGalleryImage 更新图片
func (r *GalleryRepository) UpdateGalleryImage(img *gallery.GalleryImage) error {
	return r.db.Save(img).Error
}

// UpdateImageOrder 更新图片排序
func (r *GalleryRepository) UpdateImageOrder(id uint, order int) error {
	return r.db.Model(&gallery.GalleryImage{}).Where("id = ?", id).
		Update("order", order).Error
}

// DeleteGalleryImage 删除图片
func (r *GalleryRepository) DeleteGalleryImage(id uint) error {
	return r.db.Delete(&gallery.GalleryImage{}, id).Error
}

func (r *GalleryRepository) DeleteGalleryImageByIDAndGalleryID(imageID, galleryID uint) error {
	result := r.db.
		Where("id = ? AND gallery_id = ?", imageID, galleryID).
		Delete(&gallery.GalleryImage{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// BatchCreateImages 批量创建图片
func (r *GalleryRepository) BatchCreateImages(images []gallery.GalleryImage) error {
	return r.db.Create(&images).Error
}

// BatchDeleteImages 批量删除图片
func (r *GalleryRepository) BatchDeleteImages(ids []uint) error {
	return r.db.Delete(&gallery.GalleryImage{}, ids).Error
}

func (r *GalleryRepository) BatchDeleteImagesByGalleryID(galleryID uint, ids []uint) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	uniqueIDs := make([]uint, 0, len(ids))
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}
	if len(uniqueIDs) == 0 {
		return 0, gorm.ErrRecordNotFound
	}

	var matchingCount int64
	if err := r.db.Model(&gallery.GalleryImage{}).
		Where("gallery_id = ? AND id IN ?", galleryID, uniqueIDs).
		Count(&matchingCount).Error; err != nil {
		return 0, err
	}
	if matchingCount != int64(len(uniqueIDs)) {
		return 0, gorm.ErrRecordNotFound
	}

	result := r.db.
		Where("gallery_id = ? AND id IN ?", galleryID, uniqueIDs).
		Delete(&gallery.GalleryImage{})
	return result.RowsAffected, result.Error
}

func (r *GalleryRepository) ReplaceGalleryProductLinks(galleryID uint, productIDs []uint) error {
	if !r.db.Migrator().HasTable(&gallery.GalleryProductLink{}) {
		return gorm.ErrInvalidData
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.replaceGalleryProductLinksTx(tx, galleryID, productIDs)
	})
}

func (r *GalleryRepository) replaceGalleryProductLinksTx(tx *gorm.DB, galleryID uint, productIDs []uint) error {
	if !tx.Migrator().HasTable(&gallery.GalleryProductLink{}) {
		return gorm.ErrInvalidData
	}

	if err := tx.Where("gallery_id = ?", galleryID).Delete(&gallery.GalleryProductLink{}).Error; err != nil {
		return err
	}
	if len(productIDs) == 0 {
		return nil
	}

	var existingIDs []uint
	if err := tx.Table("products").
		Where("id IN ? AND status = ? AND deleted_at IS NULL", productIDs, "active").
		Pluck("id", &existingIDs).Error; err != nil {
		return err
	}
	existing := make(map[uint]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		existing[id] = struct{}{}
	}

	links := make([]gallery.GalleryProductLink, 0, len(productIDs))
	seen := make(map[uint]struct{}, len(productIDs))
	for _, productID := range productIDs {
		if _, ok := existing[productID]; !ok {
			return gorm.ErrRecordNotFound
		}
		if _, ok := seen[productID]; ok {
			continue
		}
		seen[productID] = struct{}{}
		links = append(links, gallery.GalleryProductLink{
			GalleryID: galleryID,
			ProductID: productID,
			SortOrder: len(links),
		})
	}
	return tx.Create(&links).Error
}

func (r *GalleryRepository) loadFirstImages(galleries *[]gallery.Gallery) error {
	if galleries == nil || len(*galleries) == 0 {
		return nil
	}

	galleryIDs := make([]uint, 0, len(*galleries))
	for _, item := range *galleries {
		galleryIDs = append(galleryIDs, item.ID)
	}

	var images []gallery.GalleryImage
	err := r.db.Raw(`
		SELECT *
		FROM (
			SELECT gallery_images.*,
				ROW_NUMBER() OVER (
					PARTITION BY gallery_id
					ORDER BY "order" ASC, created_at ASC, id ASC
				) AS row_number
			FROM gallery_images
			WHERE gallery_id IN ? AND deleted_at IS NULL
		) ranked_gallery_images
		WHERE row_number = 1
	`, galleryIDs).Scan(&images).Error
	if err != nil {
		return err
	}

	imageByGalleryID := make(map[uint]gallery.GalleryImage, len(images))
	for _, image := range images {
		imageByGalleryID[image.GalleryID] = image
	}
	for i := range *galleries {
		if image, ok := imageByGalleryID[(*galleries)[i].ID]; ok {
			(*galleries)[i].Images = []gallery.GalleryImage{image}
		}
	}
	return nil
}

func (r *GalleryRepository) loadImageCounts(galleries *[]gallery.Gallery) error {
	if galleries == nil || len(*galleries) == 0 {
		return nil
	}

	galleryIDs := make([]uint, 0, len(*galleries))
	for _, item := range *galleries {
		galleryIDs = append(galleryIDs, item.ID)
	}

	type imageCountRow struct {
		GalleryID uint
		Count     int64
	}
	var rows []imageCountRow
	if err := r.db.Model(&gallery.GalleryImage{}).
		Select("gallery_id, COUNT(*) AS count").
		Where("gallery_id IN ?", galleryIDs).
		Group("gallery_id").
		Scan(&rows).Error; err != nil {
		return err
	}

	countByGalleryID := make(map[uint]int64, len(rows))
	for _, row := range rows {
		countByGalleryID[row.GalleryID] = row.Count
	}
	for i := range *galleries {
		(*galleries)[i].ImageCount = countByGalleryID[(*galleries)[i].ID]
	}
	return nil
}

func (r *GalleryRepository) preloadProductLinks(query *gorm.DB) *gorm.DB {
	if !r.db.Migrator().HasTable(&gallery.GalleryProductLink{}) {
		return query
	}
	return query.Preload("ProductLinks", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Preload("ProductLinks.Product", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "slug", "locale", "status")
	})
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

// SearchImagesInPublishedGalleries 搜索已发布图片库中的图片
func (r *GalleryRepository) SearchImagesInPublishedGalleries(keyword string, page, pageSize int) ([]gallery.GalleryImage, int64, error) {
	var images []gallery.GalleryImage
	var total int64

	query := r.db.Model(&gallery.GalleryImage{}).
		Joins("JOIN galleries ON galleries.id = gallery_images.gallery_id").
		Where("galleries.status = ?", "published").
		Where("gallery_images.title ILIKE ? OR gallery_images.description ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("gallery_images.created_at DESC").Offset(offset).Limit(pageSize).Find(&images).Error

	return images, total, err
}
