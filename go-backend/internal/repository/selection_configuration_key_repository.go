package repository

import (
	"strings"

	selectionconfiguration "commerce-platform/internal/domain/selectionconfiguration"

	"gorm.io/gorm"
)

type SelectionConfigurationKeyRepository struct {
	db *gorm.DB
}

func NewSelectionConfigurationKeyRepository(db *gorm.DB) *SelectionConfigurationKeyRepository {
	return &SelectionConfigurationKeyRepository{db: db}
}

func (r *SelectionConfigurationKeyRepository) WithTx(tx *gorm.DB) *SelectionConfigurationKeyRepository {
	return &SelectionConfigurationKeyRepository{db: tx}
}

func (r *SelectionConfigurationKeyRepository) ListSelectionConfigurationKeysByKind(kind string, includeDisabled bool) ([]selectionconfiguration.SelectionConfigurationKey, error) {
	var keys []selectionconfiguration.SelectionConfigurationKey
	query := r.db.Model(&selectionconfiguration.SelectionConfigurationKey{}).Where("kind = ?", strings.TrimSpace(kind))
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	if err := query.Order("sort_order ASC, code ASC").Find(&keys).Error; err != nil {
		return nil, err
	}
	return keys, nil
}

func (r *SelectionConfigurationKeyRepository) FindSelectionConfigurationKeyByID(id uint) (*selectionconfiguration.SelectionConfigurationKey, error) {
	var key selectionconfiguration.SelectionConfigurationKey
	if err := r.db.First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *SelectionConfigurationKeyRepository) FindSelectionConfigurationKeyByKindAndCode(kind, code string) (*selectionconfiguration.SelectionConfigurationKey, error) {
	var key selectionconfiguration.SelectionConfigurationKey
	if err := r.db.Where("kind = ? AND code = ?", strings.TrimSpace(kind), strings.TrimSpace(code)).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *SelectionConfigurationKeyRepository) CreateSelectionConfigurationKey(key *selectionconfiguration.SelectionConfigurationKey) error {
	return r.db.Select(
		"kind",
		"code",
		"display_label",
		"description",
		"is_enabled",
		"sort_order",
	).Create(key).Error
}

func (r *SelectionConfigurationKeyRepository) SaveSelectionConfigurationKey(key *selectionconfiguration.SelectionConfigurationKey) error {
	return r.db.Save(key).Error
}
