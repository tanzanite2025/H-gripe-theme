package repository

import (
	"time"

	selectionassistant "commerce-platform/internal/domain/selectionassistant"

	"gorm.io/gorm"
)

type SelectionAssistantRepository struct {
	db *gorm.DB
}

func NewSelectionAssistantRepository(db *gorm.DB) *SelectionAssistantRepository {
	return &SelectionAssistantRepository{db: db}
}

func (r *SelectionAssistantRepository) ListFlows() ([]selectionassistant.Flow, error) {
	var flows []selectionassistant.Flow
	err := r.db.
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC, id DESC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&flows).Error
	return flows, err
}

func (r *SelectionAssistantRepository) FindFlowByID(id uint) (*selectionassistant.Flow, error) {
	var flow selectionassistant.Flow
	if err := r.db.
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC, id DESC")
		}).
		First(&flow, id).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *SelectionAssistantRepository) FindEnabledPublishedFlowBySlug(slug string) (*selectionassistant.Flow, error) {
	var flow selectionassistant.Flow
	if err := r.db.
		Where("slug = ? AND is_enabled = TRUE", slug).
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.
				Where("status = ?", selectionassistant.FlowVersionStatusPublished).
				Order("version_number DESC, id DESC")
		}).
		First(&flow).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *SelectionAssistantRepository) CreateFlowWithVersion(flow *selectionassistant.Flow, version *selectionassistant.Version) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(flow).Error; err != nil {
			return err
		}
		version.FlowID = flow.ID
		return tx.Create(version).Error
	})
}

func (r *SelectionAssistantRepository) UpdateFlow(flow *selectionassistant.Flow) error {
	result := r.db.Model(&selectionassistant.Flow{}).
		Where("id = ?", flow.ID).
		Updates(map[string]interface{}{
			"slug":                  flow.Slug,
			"name":                  flow.Name,
			"description":           flow.Description,
			"product_category_slug": flow.ProductCategorySlug,
			"is_enabled":            flow.IsEnabled,
			"sort_order":            flow.SortOrder,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SelectionAssistantRepository) FindVersionByID(id uint) (*selectionassistant.Version, error) {
	var version selectionassistant.Version
	if err := r.db.
		Preload("Flow").
		First(&version, id).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *SelectionAssistantRepository) FindDraftVersion(flowID uint) (*selectionassistant.Version, error) {
	var version selectionassistant.Version
	if err := r.db.
		Where("flow_id = ? AND status = ?", flowID, selectionassistant.FlowVersionStatusDraft).
		Order("version_number DESC, id DESC").
		First(&version).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *SelectionAssistantRepository) FindLatestVersionNumber(flowID uint) (int, error) {
	var version selectionassistant.Version
	err := r.db.
		Where("flow_id = ?", flowID).
		Order("version_number DESC, id DESC").
		First(&version).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return version.VersionNumber, nil
}

func (r *SelectionAssistantRepository) CreateVersion(version *selectionassistant.Version) error {
	return r.db.Create(version).Error
}

func (r *SelectionAssistantRepository) UpdateDraftVersion(version *selectionassistant.Version) error {
	result := r.db.Model(&selectionassistant.Version{}).
		Where("id = ? AND status = ?", version.ID, selectionassistant.FlowVersionStatusDraft).
		Updates(map[string]interface{}{
			"config": version.Config,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *SelectionAssistantRepository) PublishVersion(versionID uint, publishedBy *uint, publishedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var version selectionassistant.Version
		if err := tx.First(&version, versionID).Error; err != nil {
			return err
		}
		if err := tx.Model(&selectionassistant.Version{}).
			Where("flow_id = ? AND id <> ? AND status = ?", version.FlowID, version.ID, selectionassistant.FlowVersionStatusPublished).
			Updates(map[string]interface{}{"status": selectionassistant.FlowVersionStatusArchived}).Error; err != nil {
			return err
		}
		result := tx.Model(&selectionassistant.Version{}).
			Where("id = ? AND status = ?", version.ID, selectionassistant.FlowVersionStatusDraft).
			Updates(map[string]interface{}{
				"status":       selectionassistant.FlowVersionStatusPublished,
				"published_at": publishedAt,
				"published_by": publishedBy,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}
