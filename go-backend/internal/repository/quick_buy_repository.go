package repository

import (
	"time"

	"tanzanite/internal/domain/quickbuy"

	"gorm.io/gorm"
)

type QuickBuyRepository struct {
	db *gorm.DB
}

func NewQuickBuyRepository(db *gorm.DB) *QuickBuyRepository {
	return &QuickBuyRepository{db: db}
}

func (r *QuickBuyRepository) ListFlows() ([]quickbuy.Flow, error) {
	var flows []quickbuy.Flow
	err := r.db.
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC, id DESC")
		}).
		Order("sort_order ASC, id ASC").
		Find(&flows).Error
	return flows, err
}

func (r *QuickBuyRepository) FindFlowByID(id uint) (*quickbuy.Flow, error) {
	var flow quickbuy.Flow
	if err := r.db.
		Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("version_number DESC, id DESC")
		}).
		First(&flow, id).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *QuickBuyRepository) FindFlowBySlug(slug string) (*quickbuy.Flow, error) {
	var flow quickbuy.Flow
	if err := r.db.Where("slug = ?", slug).First(&flow).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

func (r *QuickBuyRepository) CreateFlowWithVersion(flow *quickbuy.Flow, version *quickbuy.Version) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(flow).Error; err != nil {
			return err
		}
		version.FlowID = flow.ID
		return createQuickBuyVersion(tx, version)
	})
}

func (r *QuickBuyRepository) UpdateFlow(flow *quickbuy.Flow) error {
	result := r.db.Model(&quickbuy.Flow{}).
		Where("id = ?", flow.ID).
		Updates(map[string]interface{}{
			"slug":          flow.Slug,
			"name":          flow.Name,
			"description":   flow.Description,
			"entry_surface": flow.EntrySurface,
			"is_enabled":    flow.IsEnabled,
			"sort_order":    flow.SortOrder,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *QuickBuyRepository) FindVersionByID(id uint) (*quickbuy.Version, error) {
	var version quickbuy.Version
	if err := preloadQuickBuyVersion(r.db).First(&version, id).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *QuickBuyRepository) FindLatestVersionNumber(flowID uint) (int, error) {
	var version quickbuy.Version
	err := r.db.
		Where("flow_id = ?", flowID).
		Order("version_number DESC").
		First(&version).Error
	if err != nil {
		if IsRecordNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	return version.VersionNumber, nil
}

func (r *QuickBuyRepository) CreateVersion(version *quickbuy.Version) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return createQuickBuyVersion(tx, version)
	})
}

func (r *QuickBuyRepository) ReplaceVersion(version *quickbuy.Version) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&quickbuy.Version{}).
			Where("id = ?", version.ID).
			Updates(map[string]interface{}{
				"starts_at": version.StartsAt,
				"ends_at":   version.EndsAt,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("flow_version_id = ?", version.ID).Delete(&quickbuy.Rule{}).Error; err != nil {
			return err
		}
		if err := tx.Where("step_id IN (?)",
			tx.Model(&quickbuy.Step{}).Select("id").Where("flow_version_id = ?", version.ID),
		).Delete(&quickbuy.StepFilter{}).Error; err != nil {
			return err
		}
		if err := tx.Where("step_id IN (?)",
			tx.Model(&quickbuy.Step{}).Select("id").Where("flow_version_id = ?", version.ID),
		).Delete(&quickbuy.StepProductType{}).Error; err != nil {
			return err
		}
		if err := tx.Where("flow_version_id = ?", version.ID).Delete(&quickbuy.Step{}).Error; err != nil {
			return err
		}
		return createQuickBuyVersionChildren(tx, version)
	})
}

func (r *QuickBuyRepository) PublishVersion(versionID uint, publishedBy *uint, publishedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var version quickbuy.Version
		if err := tx.First(&version, versionID).Error; err != nil {
			return err
		}
		if err := tx.Model(&quickbuy.Version{}).
			Where("flow_id = ? AND id <> ? AND status = ?", version.FlowID, version.ID, quickbuy.FlowVersionStatusPublished).
			Updates(map[string]interface{}{"status": quickbuy.FlowVersionStatusArchived}).Error; err != nil {
			return err
		}
		result := tx.Model(&quickbuy.Version{}).
			Where("id = ?", version.ID).
			Updates(map[string]interface{}{
				"status":       quickbuy.FlowVersionStatusPublished,
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

func (r *QuickBuyRepository) ListPublishedVersions(surface string, now time.Time) ([]quickbuy.Version, error) {
	var versions []quickbuy.Version
	query := preloadQuickBuyVersion(r.db).
		Joins("JOIN quick_buy_flows ON quick_buy_flows.id = quick_buy_flow_versions.flow_id").
		Where("quick_buy_flow_versions.status = ?", quickbuy.FlowVersionStatusPublished).
		Where("quick_buy_flows.is_enabled = ?", true).
		Where("(quick_buy_flow_versions.starts_at IS NULL OR quick_buy_flow_versions.starts_at <= ?)", now).
		Where("(quick_buy_flow_versions.ends_at IS NULL OR quick_buy_flow_versions.ends_at > ?)", now)
	if surface != "" {
		query = query.Where("quick_buy_flows.entry_surface = ?", surface)
	}
	err := query.
		Order("quick_buy_flows.sort_order ASC").
		Order("quick_buy_flow_versions.published_at DESC NULLS LAST").
		Find(&versions).Error
	return versions, err
}

func (r *QuickBuyRepository) CreateSession(session *quickbuy.Session) error {
	return r.db.Create(session).Error
}

func (r *QuickBuyRepository) FindSessionByToken(token string) (*quickbuy.Session, error) {
	var session quickbuy.Session
	if err := preloadQuickBuySession(r.db).
		Where("session_token = ?", token).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *QuickBuyRepository) ReplaceSessionItems(sessionID uint, items []quickbuy.SessionItem, status, validationStatus string, subtotal float64, weightG int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("session_id = ?", sessionID).Delete(&quickbuy.SessionItem{}).Error; err != nil {
			return err
		}
		for index := range items {
			items[index].SessionID = sessionID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return tx.Model(&quickbuy.Session{}).
			Where("id = ?", sessionID).
			Updates(map[string]interface{}{
				"status":            status,
				"validation_status": validationStatus,
				"subtotal_snapshot": subtotal,
				"weight_snapshot_g": weightG,
			}).Error
	})
}

func preloadQuickBuyVersion(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Flow").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Steps.ProductTypes", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Steps.ProductTypes.ProductType.SpecDefinitions", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Steps.ProductTypes.ProductType.Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Preload("Steps.Filters", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Rules", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		})
}

func preloadQuickBuySession(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Flow").
		Preload("Version", func(db *gorm.DB) *gorm.DB {
			return preloadQuickBuyVersion(db)
		}).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		})
}

func createQuickBuyVersion(tx *gorm.DB, version *quickbuy.Version) error {
	steps := version.Steps
	rules := version.Rules
	version.Steps = nil
	version.Rules = nil
	if err := tx.Create(version).Error; err != nil {
		return err
	}
	version.Steps = steps
	version.Rules = rules
	return createQuickBuyVersionChildren(tx, version)
}

func createQuickBuyVersionChildren(tx *gorm.DB, version *quickbuy.Version) error {
	for stepIndex := range version.Steps {
		step := &version.Steps[stepIndex]
		step.FlowVersionID = version.ID
		productTypes := step.ProductTypes
		filters := step.Filters
		step.ProductTypes = nil
		step.Filters = nil
		if err := tx.Create(step).Error; err != nil {
			return err
		}
		for index := range productTypes {
			productTypes[index].StepID = step.ID
		}
		if len(productTypes) > 0 {
			if err := tx.Create(&productTypes).Error; err != nil {
				return err
			}
		}
		for index := range filters {
			filters[index].StepID = step.ID
		}
		if len(filters) > 0 {
			if err := tx.Create(&filters).Error; err != nil {
				return err
			}
		}
		step.ProductTypes = productTypes
		step.Filters = filters
	}
	for index := range version.Rules {
		version.Rules[index].FlowVersionID = version.ID
	}
	if len(version.Rules) > 0 {
		if err := tx.Create(&version.Rules).Error; err != nil {
			return err
		}
	}
	return nil
}
