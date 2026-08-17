package repository

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/media"

	"gorm.io/gorm"
)

type MediaDerivativePresetRepository struct {
	db *gorm.DB
}

func NewMediaDerivativePresetRepository(db *gorm.DB) *MediaDerivativePresetRepository {
	return &MediaDerivativePresetRepository{db: db}
}

func (r *MediaDerivativePresetRepository) WithTx(tx *gorm.DB) *MediaDerivativePresetRepository {
	return &MediaDerivativePresetRepository{db: tx}
}

func (r *MediaDerivativePresetRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	return r.db.Transaction(fn)
}

func (r *MediaDerivativePresetRepository) ListActive() ([]media.MediaDerivativePreset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if !r.hasPresetTable() {
		return nil, nil
	}

	presets := make([]media.MediaDerivativePreset, 0)
	err := r.db.
		Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&presets).Error
	return presets, err
}

func (r *MediaDerivativePresetRepository) ListAll() ([]media.MediaDerivativePreset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if !r.hasPresetTable() {
		return nil, nil
	}

	presets := make([]media.MediaDerivativePreset, 0)
	if err := r.db.
		Order("sort_order ASC, id ASC").
		Find(&presets).Error; err != nil {
		return nil, err
	}
	if len(presets) == 0 || !r.hasDerivativesTable() {
		return presets, nil
	}

	type derivativeUsage struct {
		Preset string
		Count  int64
	}
	usages := make([]derivativeUsage, 0)
	if err := r.db.
		Model(&media.MediaAssetDerivative{}).
		Select("preset, COUNT(*) AS count").
		Group("preset").
		Scan(&usages).Error; err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(usages))
	for _, usage := range usages {
		counts[strings.TrimSpace(usage.Preset)] = usage.Count
	}
	for index := range presets {
		presets[index].GeneratedDerivatives = counts[strings.TrimSpace(presets[index].Code)]
	}
	return presets, nil
}

func (r *MediaDerivativePresetRepository) CountActive() (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}
	if !r.hasPresetTable() {
		return 0, nil
	}
	var count int64
	err := r.db.Model(&media.MediaDerivativePreset{}).
		Where("enabled = ?", true).
		Count(&count).Error
	return count, err
}

// LockActiveConfiguration serializes configuration changes that affect the
// active preset budget on PostgreSQL. SQLite test databases already serialize
// writes, so no database-specific lock is needed there.
func (r *MediaDerivativePresetRepository) LockActiveConfiguration() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if r.db.Dialector != nil && r.db.Dialector.Name() == "postgres" {
		return r.db.Exec("SELECT pg_advisory_xact_lock(?)", int64(179_001)).Error
	}
	return nil
}

func (r *MediaDerivativePresetRepository) FindByID(id uint) (*media.MediaDerivativePreset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if id == 0 || !r.hasPresetTable() {
		return nil, gorm.ErrRecordNotFound
	}

	var preset media.MediaDerivativePreset
	if err := r.db.First(&preset, id).Error; err != nil {
		return nil, err
	}
	return &preset, nil
}

func (r *MediaDerivativePresetRepository) FindByCode(code string) (*media.MediaDerivativePreset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if code = strings.TrimSpace(code); code == "" || !r.hasPresetTable() {
		return nil, gorm.ErrRecordNotFound
	}

	var preset media.MediaDerivativePreset
	if err := r.db.Where("code = ?", code).First(&preset).Error; err != nil {
		return nil, err
	}
	return &preset, nil
}

func (r *MediaDerivativePresetRepository) Create(preset *media.MediaDerivativePreset) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if preset == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(preset).Error
}

func (r *MediaDerivativePresetRepository) Update(preset *media.MediaDerivativePreset) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if preset == nil || preset.ID == 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Save(preset).Error
}

func (r *MediaDerivativePresetRepository) Delete(preset *media.MediaDerivativePreset) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if preset == nil || preset.ID == 0 {
		return gorm.ErrInvalidData
	}
	return r.db.Delete(preset).Error
}

func (r *MediaDerivativePresetRepository) CountDerivativesByPreset(code string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, gorm.ErrInvalidDB
	}
	if code = strings.TrimSpace(code); code == "" || !r.hasDerivativesTable() {
		return 0, nil
	}

	var count int64
	err := r.db.
		Unscoped().
		Model(&media.MediaAssetDerivative{}).
		Where("preset = ?", code).
		Count(&count).Error
	return count, err
}

func (r *MediaDerivativePresetRepository) EnsureSystemDefaults(defaults []media.MediaDerivativePreset) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if !r.hasPresetTable() {
		return nil
	}

	for _, preset := range defaults {
		preset.Code = strings.TrimSpace(preset.Code)
		if preset.Code == "" {
			continue
		}
		var existing media.MediaDerivativePreset
		err := r.db.
			Where("code = ?", preset.Code).
			First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := r.db.Create(&preset).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *MediaDerivativePresetRepository) hasPresetTable() bool {
	return r != nil && r.db != nil && r.db.Migrator().HasTable(&media.MediaDerivativePreset{})
}

func (r *MediaDerivativePresetRepository) hasDerivativesTable() bool {
	return r != nil && r.db != nil && r.db.Migrator().HasTable(&media.MediaAssetDerivative{})
}
