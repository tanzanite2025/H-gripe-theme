package repository

import (
	"tanzanite/internal/domain/setting"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

// Get 获取设置
func (r *SettingRepository) Get(key, locale string) (*setting.Setting, error) {
	var s setting.Setting
	err := r.db.Where("key = ? AND locale = ?", key, locale).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByGroup 获取分组设置
func (r *SettingRepository) GetByGroup(group, locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where("group = ? AND locale = ?", group, locale).Find(&settings).Error
	return settings, err
}

// Set 设置值
func (r *SettingRepository) Set(s *setting.Setting) error {
	var existing setting.Setting
	err := r.db.Where("key = ? AND locale = ?", s.Key, s.Locale).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(s).Error
	}

	existing.Value = s.Value
	existing.Type = s.Type
	existing.Group = s.Group
	return r.db.Save(&existing).Error
}

// Delete 删除设置
func (r *SettingRepository) Delete(key, locale string) error {
	return r.db.Where("key = ? AND locale = ?", key, locale).Delete(&setting.Setting{}).Error
}

// GetAll 获取所有设置
func (r *SettingRepository) GetAll(locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where("locale = ?", locale).Find(&settings).Error
	return settings, err
}

// GetAllPublic 获取所有公开设置
func (r *SettingRepository) GetAllPublic(locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where("locale = ? AND is_public = ?", locale, true).Find(&settings).Error
	return settings, err
}

// GetByKeys 批量获取设置
func (r *SettingRepository) GetByKeys(keys []string, locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where("key IN ? AND locale = ?", keys, locale).Find(&settings).Error
	return settings, err
}

// BatchSet 批量设置
func (r *SettingRepository) BatchSet(settings []setting.Setting) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, s := range settings {
			var existing setting.Setting
			err := tx.Where("key = ? AND locale = ?", s.Key, s.Locale).First(&existing).Error

			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&s).Error; err != nil {
					return err
				}
			} else {
				existing.Value = s.Value
				existing.Type = s.Type
				existing.Group = s.Group
				existing.IsPublic = s.IsPublic
				existing.Description = s.Description
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// GetGroups 获取所有分组
func (r *SettingRepository) GetGroups() ([]string, error) {
	var groups []string
	err := r.db.Model(&setting.Setting{}).Distinct("group").Pluck("group", &groups).Error
	return groups, err
}
