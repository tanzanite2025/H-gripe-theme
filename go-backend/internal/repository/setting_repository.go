package repository

import (
	"commerce-platform/internal/domain/setting"

	"gorm.io/gorm"
)

type SettingRepository struct {
	db                             *gorm.DB
	businessMediaReferenceResolver PublicBusinessMediaReferenceObjectKeyResolver
}

func NewSettingRepository(db *gorm.DB) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) WithTx(tx *gorm.DB) *SettingRepository {
	return &SettingRepository{db: tx, businessMediaReferenceResolver: r.businessMediaReferenceResolver}
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

// GetPublic 获取公开设置
func (r *SettingRepository) GetPublic(key, locale string) (*setting.Setting, error) {
	var s setting.Setting
	err := r.db.Where("key = ? AND locale = ? AND is_public = ?", key, locale, true).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByGroup 获取分组设置
func (r *SettingRepository) GetByGroup(group, locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where(map[string]interface{}{
		"group":  group,
		"locale": locale,
	}).Find(&settings).Error
	return settings, err
}

// GetPublicByGroup 获取公开分组设置
func (r *SettingRepository) GetPublicByGroup(group, locale string) ([]setting.Setting, error) {
	var settings []setting.Setting
	err := r.db.Where(map[string]interface{}{
		"group":     group,
		"locale":    locale,
		"is_public": true,
	}).Find(&settings).Error
	return settings, err
}

// Set 设置值
func (r *SettingRepository) Set(s *setting.Setting) error {
	if s == nil {
		return gorm.ErrInvalidData
	}
	values := []setting.Setting{*s}
	if err := r.persistSettingsWithOneTransactionalMediaBindingPlan(values, false); err != nil {
		return err
	}
	*s = values[0]
	return nil
}

// Delete 删除设置
func (r *SettingRepository) Delete(key, locale string) error {
	if !setting.IsDedicatedSiteIdentityOwnerSettingKey(key) {
		return r.db.Where("key = ? AND locale = ?", key, locale).Delete(&setting.Setting{}).Error
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := lockSiteIdentitySettingNamespacesBeforeSortedSettingRows(tx, []settingKeyAndLocaleIdentity{{Key: key, Locale: locale}}); err != nil {
			return err
		}
		return tx.Where("key = ? AND locale = ?", key, locale).Delete(&setting.Setting{}).Error
	})
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
	return r.persistSettingsWithOneTransactionalMediaBindingPlan(settings, true)
}

// GetGroups 获取所有分组
func (r *SettingRepository) GetGroups() ([]string, error) {
	var groups []string
	err := r.db.Model(&setting.Setting{}).Distinct().Pluck("group", &groups).Error
	return groups, err
}

// GetPublicGroups 获取包含公开设置的分组
func (r *SettingRepository) GetPublicGroups() ([]string, error) {
	var groups []string
	err := r.db.Model(&setting.Setting{}).Where("is_public = ?", true).Distinct().Pluck("group", &groups).Error
	return groups, err
}
