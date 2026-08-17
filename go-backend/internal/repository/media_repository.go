package repository

import (
	"commerce-platform/internal/domain/media"
	productdomain "commerce-platform/internal/domain/product"
	"fmt"
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

type MediaImageDimensionFilter struct {
	Page                int
	PageSize            int
	Search              string
	State               string
	RequiredDerivatives []MediaDerivativeRequirement
}

type MediaDerivativeRequirement struct {
	Preset        string
	PresetVersion int
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
	listQuery := query
	if r.hasAssetDerivativesTable() {
		listQuery = listQuery.Preload("Derivatives")
	}
	err := listQuery.Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&assets).Error
	if err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *MediaRepository) ListImageDimensionAssets(
	filter MediaImageDimensionFilter,
) ([]media.MediaAsset, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, gorm.ErrInvalidDB
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	query := r.imageDimensionQuery(filter.Search, filter.State, filter.RequiredDerivatives)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var assets []media.MediaAsset
	listQuery := query
	if r.hasAssetDerivativesTable() {
		listQuery = listQuery.Preload("Derivatives")
	}
	err := listQuery.
		Order("media_assets.created_at DESC, media_assets.id DESC").
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&assets).Error
	return assets, total, err
}

func (r *MediaRepository) CountImageDimensionAssets(state string, requiredDerivatives []MediaDerivativeRequirement) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}
	var total int64
	err := r.imageDimensionQuery("", state, requiredDerivatives).Count(&total).Error
	return total, err
}

func (r *MediaRepository) UpdateAssetDimensions(id uint, width int, height int) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Model(&media.MediaAsset{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"width": width, "height": height}).Error
}

func (r *MediaRepository) imageDimensionQuery(search string, state string, requiredDerivatives []MediaDerivativeRequirement) *gorm.DB {
	query := r.db.Model(&media.MediaAsset{}).Where("media_assets.media_type = ?", "image")
	if search = strings.TrimSpace(search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(COALESCE(filename, '')) LIKE ? OR LOWER(COALESCE(original_filename, '')) LIKE ? OR LOWER(COALESCE(url, '')) LIKE ?",
			like, like, like,
		)
	}

	missingDimensions := "(media_assets.width <= 0 OR media_assets.height <= 0)"
	if !r.hasAssetDerivativesTable() {
		missingVariants := len(requiredDerivatives) > 0
		switch state {
		case "ready":
			if missingVariants {
				return query.Where("1 = 0")
			}
			return query.Where("NOT " + missingDimensions)
		case "missing_dimensions":
			return query.Where(missingDimensions)
		case "attention", "missing_variants":
			if missingVariants {
				return query
			}
			return query.Where(missingDimensions)
		default:
			return query
		}
	}

	missingVariants, missingVariantArgs := missingDerivativeRequirementsCondition(requiredDerivatives)
	switch state {
	case "attention":
		return query.Where(missingDimensions+" OR "+missingVariants, missingVariantArgs...)
	case "missing_dimensions":
		return query.Where(missingDimensions)
	case "missing_variants":
		return query.Where(missingVariants, missingVariantArgs...)
	case "ready":
		return query.Where("NOT "+missingDimensions).Where("NOT "+missingVariants, missingVariantArgs...)
	default:
		return query
	}
}

func (r *MediaRepository) FindAssetByID(id uint) (*media.MediaAsset, error) {
	var asset media.MediaAsset
	query := r.db
	if r.hasAssetDerivativesTable() {
		query = query.Preload("Derivatives")
	}
	if err := query.First(&asset, id).Error; err != nil {
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
	likeSlash := "%/uploads/" + escapeLikePattern(normalizedKey)
	likeBackslash := "%/uploads/" + escapeLikePattern(legacyKey)

	var asset media.MediaAsset
	if err := r.db.Unscoped().
		Where("storage_key = ? OR storage_key = ? OR url LIKE ? ESCAPE '\\' OR url LIKE ? ESCAPE '\\'", normalizedKey, legacyKey, likeSlash, likeBackslash).
		First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func escapeLikePattern(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}

func (r *MediaRepository) UpdateAsset(asset *media.MediaAsset) error {
	return r.db.Save(asset).Error
}

func (r *MediaRepository) CreateAssetDerivatives(derivatives []media.MediaAssetDerivative) error {
	if len(derivatives) == 0 {
		return nil
	}
	return r.db.Create(&derivatives).Error
}

func (r *MediaRepository) ReplaceAssetDerivatives(derivatives []media.MediaAssetDerivative) ([]string, error) {
	if len(derivatives) == 0 {
		return nil, nil
	}
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	oldURLs := make([]string, 0, len(derivatives))
	err := r.db.Transaction(func(tx *gorm.DB) error {
		for _, derivative := range derivatives {
			if derivative.MediaAssetID == 0 || strings.TrimSpace(derivative.Preset) == "" {
				continue
			}
			var existing []media.MediaAssetDerivative
			if err := tx.
				Where("media_asset_id = ? AND preset = ?", derivative.MediaAssetID, derivative.Preset).
				Find(&existing).Error; err != nil {
				return err
			}
			for _, item := range existing {
				if strings.TrimSpace(item.URL) != "" {
					oldURLs = append(oldURLs, item.URL)
				}
			}
			if err := tx.
				Where("media_asset_id = ? AND preset = ?", derivative.MediaAssetID, derivative.Preset).
				Delete(&media.MediaAssetDerivative{}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&derivatives).Error
	})
	if err != nil {
		return nil, err
	}
	return oldURLs, nil
}

func (r *MediaRepository) FindAssetDerivatives(assetID uint) ([]media.MediaAssetDerivative, error) {
	if assetID == 0 {
		return nil, nil
	}
	if !r.hasAssetDerivativesTable() {
		return nil, nil
	}
	var derivatives []media.MediaAssetDerivative
	err := r.db.
		Where("media_asset_id = ?", assetID).
		Order("width ASC, id ASC").
		Find(&derivatives).Error
	return derivatives, err
}

func (r *MediaRepository) ListImageAssetsMissingDerivatives(afterID uint, limit int, requiredDerivatives []MediaDerivativeRequirement) ([]media.MediaAsset, error) {
	if !r.hasAssetDerivativesTable() || len(requiredDerivatives) == 0 {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	missingVariants, missingVariantArgs := missingDerivativeRequirementsCondition(requiredDerivatives)

	var assets []media.MediaAsset
	query := r.db.Model(&media.MediaAsset{}).
		Preload("Derivatives").
		Where("media_assets.media_type = ?", "image").
		Where(missingVariants, missingVariantArgs...).
		Order("media_assets.id ASC").
		Limit(limit)
	if afterID > 0 {
		query = query.Where("media_assets.id > ?", afterID)
	}

	if err := query.Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

// ListImageAssetsForDerivativeRebuild returns every image asset in stable ID
// order. A rebuild must visit ready assets as well so disabled or replaced
// presets are removed from persisted product-media snapshots.
func (r *MediaRepository) ListImageAssetsForDerivativeRebuild(afterID uint, limit int) ([]media.MediaAsset, error) {
	if r == nil || r.db == nil || !r.hasAssetDerivativesTable() {
		return nil, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}

	var assets []media.MediaAsset
	query := r.db.Model(&media.MediaAsset{}).
		Preload("Derivatives").
		Where("media_assets.media_type = ?", "image").
		Order("media_assets.id ASC").
		Limit(limit)
	if afterID > 0 {
		query = query.Where("media_assets.id > ?", afterID)
	}
	if err := query.Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *MediaRepository) FindDerivativeByStorageKey(key string) (*media.MediaAssetDerivative, error) {
	normalizedKey := strings.Trim(strings.ReplaceAll(key, "\\", "/"), "/")
	if normalizedKey == "" {
		return nil, gorm.ErrRecordNotFound
	}
	if !r.hasAssetDerivativesTable() {
		return nil, gorm.ErrRecordNotFound
	}

	var derivative media.MediaAssetDerivative
	if err := r.db.
		Preload("Asset").
		Where("storage_key = ?", normalizedKey).
		First(&derivative).Error; err != nil {
		return nil, err
	}
	return &derivative, nil
}

func (r *MediaRepository) UpdateProductMediaImageVariantsForAsset(
	assetID uint,
	variants map[string]productdomain.ProductMediaImageVariant,
	thumbnail string,
) (int64, error) {
	if assetID == 0 || !r.hasProductMediaImageVariantsColumn() {
		return 0, nil
	}

	var rowsAffected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var items []productdomain.ProductMedia
		if err := tx.
			Where("media_asset_id = ? AND media_type = ?", assetID, "image").
			Find(&items).Error; err != nil {
			return err
		}

		nextVariants := productdomain.ProductMediaImageVariantsJSON(variants)
		nextThumbnail := strings.TrimSpace(thumbnail)
		for index := range items {
			item := &items[index]
			updates := map[string]interface{}{
				"image_variants": nextVariants,
			}
			if shouldSynchronizeGeneratedThumbnail(item, nextThumbnail) {
				updates["thumbnail_url"] = nextThumbnail
			}

			result := tx.Model(&productdomain.ProductMedia{}).
				Where("id = ?", item.ID).
				Updates(updates)
			if result.Error != nil {
				return result.Error
			}
			rowsAffected += result.RowsAffected
		}
		return nil
	})
	return rowsAffected, err
}

func shouldSynchronizeGeneratedThumbnail(item *productdomain.ProductMedia, nextThumbnail string) bool {
	if item == nil {
		return false
	}
	currentThumbnail := strings.TrimSpace(item.ThumbnailURL)
	if currentThumbnail == "" {
		return currentThumbnail != nextThumbnail
	}

	currentVariants := productdomain.ParseProductMediaImageVariants(item.ImageVariantData)
	previousGeneratedThumbnail := generatedProductMediaThumbnail(currentVariants)
	return previousGeneratedThumbnail != "" && currentThumbnail == previousGeneratedThumbnail && currentThumbnail != nextThumbnail
}

func generatedProductMediaThumbnail(variants map[string]productdomain.ProductMediaImageVariant) string {
	for _, preset := range []string{"thumbnail", "card", "large"} {
		if item, ok := variants[preset]; ok && strings.TrimSpace(item.URL) != "" {
			return strings.TrimSpace(item.URL)
		}
	}
	return ""
}

func (r *MediaRepository) hasAssetDerivativesTable() bool {
	return r != nil && r.db != nil && r.db.Migrator().HasTable(&media.MediaAssetDerivative{})
}

func (r *MediaRepository) hasProductMediaImageVariantsColumn() bool {
	return r != nil &&
		r.db != nil &&
		r.db.Migrator().HasTable(&productdomain.ProductMedia{}) &&
		r.db.Migrator().HasColumn(&productdomain.ProductMedia{}, "image_variants")
}

func missingDerivativeRequirementsCondition(requiredDerivatives []MediaDerivativeRequirement) (string, []interface{}) {
	if len(requiredDerivatives) == 0 {
		return "(1 = 0)", nil
	}

	parts := make([]string, 0, len(requiredDerivatives))
	args := make([]interface{}, 0, len(requiredDerivatives)*2)
	for index, requirement := range requiredDerivatives {
		preset := strings.TrimSpace(requirement.Preset)
		if preset == "" {
			continue
		}
		version := requirement.PresetVersion
		if version <= 0 {
			version = 1
		}
		alias := fmt.Sprintf("derivative_%d", index)
		parts = append(parts, fmt.Sprintf(`NOT EXISTS (
			SELECT 1 FROM media_asset_derivatives AS %[1]s
			WHERE %[1]s.media_asset_id = media_assets.id
				AND %[1]s.deleted_at IS NULL
				AND %[1]s.preset = ?
				AND %[1]s.preset_version = ?
				AND TRIM(COALESCE(%[1]s.url, '')) <> ''
		)`, alias))
		args = append(args, preset, version)
	}
	if len(parts) == 0 {
		return "(1 = 0)", nil
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func (r *MediaRepository) DeleteAsset(id uint) error {
	return r.db.Delete(&media.MediaAsset{}, id).Error
}

func (r *MediaRepository) HardDeleteAsset(id uint) error {
	return r.db.Unscoped().Delete(&media.MediaAsset{}, id).Error
}
