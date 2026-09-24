package repository

import (
	"commerce-platform/internal/domain/notification"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationTemplateRepository struct {
	db *gorm.DB
}

func NewNotificationTemplateRepository(db *gorm.DB) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{db: db}
}

func (r *NotificationTemplateRepository) WithTx(tx *gorm.DB) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{db: tx}
}

func (r *NotificationTemplateRepository) FindByCodeLocale(code, locale string) (*notification.EmailTemplate, error) {
	var template notification.EmailTemplate
	if err := r.db.Where("code = ? AND locale = ?", code, locale).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *NotificationTemplateRepository) FindEnabledByCodeLocale(code, locale string) (*notification.EmailTemplate, error) {
	var template notification.EmailTemplate
	if err := r.db.Where("code = ? AND locale = ? AND is_enabled = ?", code, locale, true).First(&template).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *NotificationTemplateRepository) ListByCode(code string) ([]notification.EmailTemplate, error) {
	var templates []notification.EmailTemplate
	err := r.db.Where("code = ?", code).Order("locale ASC").Find(&templates).Error
	return templates, err
}

func (r *NotificationTemplateRepository) ListAll() ([]notification.EmailTemplate, error) {
	var templates []notification.EmailTemplate
	err := r.db.Order("code ASC, locale ASC").Find(&templates).Error
	return templates, err
}

func (r *NotificationTemplateRepository) FindByID(id uint) (*notification.EmailTemplate, error) {
	var template notification.EmailTemplate
	if err := r.db.First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *NotificationTemplateRepository) ListVersions(templateID uint) ([]notification.EmailTemplateVersion, error) {
	var versions []notification.EmailTemplateVersion
	err := r.db.Where("template_id = ?", templateID).Order("version DESC, id DESC").Find(&versions).Error
	return versions, err
}

func (r *NotificationTemplateRepository) FindVersion(templateID uint, version int) (*notification.EmailTemplateVersion, error) {
	var record notification.EmailTemplateVersion
	if err := r.db.Where("template_id = ? AND version = ?", templateID, version).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// SaveWithVersion updates the current locale row and appends its immutable
// audit snapshot atomically. The caller supplies the next version number after
// applying optimistic-lock policy in the service layer.
func (r *NotificationTemplateRepository) SaveWithVersion(template *notification.EmailTemplate, version *notification.EmailTemplateVersion) error {
	if err := template.Validate(); err != nil {
		return err
	}
	if version == nil || version.Version <= 0 {
		return notification.ErrTemplateVersionInvalid
	}
	if version.Version != template.Version {
		return notification.ErrTemplateVersionConflict
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if template.ID == 0 {
			if template.Version != 1 {
				return notification.ErrTemplateVersionConflict
			}
			if err := tx.Create(template).Error; err != nil {
				return err
			}
		} else {
			var current notification.EmailTemplate
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, template.ID).Error; err != nil {
				return err
			}
			if current.Code != template.Code || current.Locale != template.Locale {
				return notification.ErrTemplateIdentityImmutable
			}
			if template.Version != current.Version+1 {
				return notification.ErrTemplateVersionConflict
			}
			result := tx.Model(&notification.EmailTemplate{}).
				Where("id = ?", template.ID).
				Updates(map[string]interface{}{
					"name":               template.Name,
					"category":           template.Category,
					"subject_template":   template.SubjectTemplate,
					"body_html":          template.BodyHTML,
					"body_text":          template.BodyText,
					"allowed_variables":  template.AllowedVariables,
					"required_variables": template.RequiredVariables,
					"is_enabled":         template.IsEnabled,
					"version":            template.Version,
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		version.TemplateID = template.ID
		version.Code = template.Code
		version.Locale = template.Locale
		version.Name = template.Name
		version.SubjectTemplate = template.SubjectTemplate
		version.BodyHTML = template.BodyHTML
		version.BodyText = template.BodyText
		version.AllowedVariables = template.AllowedVariables
		version.RequiredVariables = template.RequiredVariables
		return tx.Create(version).Error
	})
}
