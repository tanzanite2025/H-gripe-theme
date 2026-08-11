package repository

import (
	"commerce-platform/internal/domain/media"
	"strings"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) CreateAsset(asset *media.MediaAsset) error {
	return r.db.Create(asset).Error
}

// AssetStorageUsageByUploaderID returns storage attributed to an uploader.
// Legacy soft-deleted rows remain counted because their physical objects may
// still exist. New media-library deletes hard-delete both the object and row.
func (r *MediaRepository) AssetStorageUsageByUploaderID(uploaderID uint) (int64, error) {
	var total int64
	err := r.db.Unscoped().Model(&media.MediaAsset{}).
		Where("uploader_id = ?", uploaderID).
		Select("COALESCE(SUM(CASE WHEN size > 0 THEN size ELSE 0 END), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

type MediaAssetFilter struct {
	Page       int
	PageSize   int
	Search     string
	MediaType  string
	Status     string
	Visibility string
}

func (r *MediaRepository) ListAssets(filter MediaAssetFilter) ([]media.MediaAsset, int64, error) {
	query := r.db.Model(&media.MediaAsset{})

	if mediaType := strings.TrimSpace(filter.MediaType); mediaType != "" {
		query = query.Where("media_type = ?", mediaType)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if visibility := strings.TrimSpace(filter.Visibility); visibility != "" {
		query = query.Where("visibility = ?", visibility)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(COALESCE(filename, '')) LIKE ? OR LOWER(COALESCE(original_filename, '')) LIKE ? OR LOWER(COALESCE(alt, '')) LIKE ? OR LOWER(COALESCE(caption, '')) LIKE ? OR LOWER(COALESCE(url, '')) LIKE ?",
			like, like, like, like, like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var assets []media.MediaAsset
	err := query.Order("created_at DESC, id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *MediaRepository) FindAssetByID(id uint) (*media.MediaAsset, error) {
	var asset media.MediaAsset
	if err := r.db.First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *MediaRepository) FindAssetByStorageKey(key string) (*media.MediaAsset, error) {
	normalizedKey := strings.Trim(strings.ReplaceAll(key, "\\", "/"), "/")
	if normalizedKey == "" {
		return nil, gorm.ErrRecordNotFound
	}
	legacyKey := strings.ReplaceAll(normalizedKey, "/", "\\")
	likeSlash := "%/uploads/" + normalizedKey
	likeBackslash := "%/uploads/" + legacyKey

	var asset media.MediaAsset
	if err := r.db.Unscoped().
		Where("storage_key = ? OR storage_key = ? OR url LIKE ? OR url LIKE ?", normalizedKey, legacyKey, likeSlash, likeBackslash).
		First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *MediaRepository) UpdateAsset(asset *media.MediaAsset) error {
	return r.db.Save(asset).Error
}

func (r *MediaRepository) DeleteAsset(id uint) error {
	return r.db.Delete(&media.MediaAsset{}, id).Error
}

func (r *MediaRepository) HardDeleteAsset(id uint) error {
	return r.db.Unscoped().Delete(&media.MediaAsset{}, id).Error
}
