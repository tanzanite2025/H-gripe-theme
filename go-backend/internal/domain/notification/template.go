package notification

import (
	"errors"
	"strings"
	"time"

	"gorm.io/datatypes"
)

const (
	TemplateCategoryOrder      = "order"
	TemplateCategoryAfterSales = "after_sales"
	TemplateCategorySystem     = "system"
)

var (
	ErrTemplateCodeRequired      = errors.New("notification template code is required")
	ErrTemplateLocaleRequired    = errors.New("notification template locale is required")
	ErrTemplateCategoryInvalid   = errors.New("notification template category is invalid")
	ErrTemplateNameRequired      = errors.New("notification template name is required")
	ErrTemplateSubjectRequired   = errors.New("notification template subject is required")
	ErrTemplateBodyRequired      = errors.New("notification template body is required")
	ErrTemplateVersionInvalid    = errors.New("notification template version must be positive")
	ErrTemplateVersionConflict   = errors.New("notification template version conflict")
	ErrTemplateIdentityImmutable = errors.New("notification template code and locale are immutable")
)

// EmailTemplate is the persisted, locale-specific template pointer. The
// current body is kept on this row for fast reads; every write must also add a
// matching EmailTemplateVersion snapshot through the repository transaction.
type EmailTemplate struct {
	ID                uint                   `gorm:"primarykey" json:"id"`
	Code              string                 `gorm:"size:64;not null;uniqueIndex:idx_email_template_code_locale" json:"code"`
	Locale            string                 `gorm:"size:16;not null;default:'en';uniqueIndex:idx_email_template_code_locale" json:"locale"`
	Category          string                 `gorm:"size:32;not null;default:'order';index" json:"category"`
	Name              string                 `gorm:"size:128;not null" json:"name"`
	SubjectTemplate   string                 `gorm:"size:255;not null" json:"subject_template"`
	BodyHTML          string                 `gorm:"type:text;not null" json:"body_html"`
	BodyText          string                 `gorm:"type:text;not null" json:"body_text"`
	AllowedVariables  datatypes.JSON         `gorm:"type:jsonb;not null;default:'[]'" json:"allowed_variables"`
	RequiredVariables datatypes.JSON         `gorm:"type:jsonb;not null;default:'[]'" json:"required_variables"`
	IsEnabled         bool                   `gorm:"not null;default:true;index" json:"is_enabled"`
	Version           int                    `gorm:"not null;default:1" json:"version"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Versions          []EmailTemplateVersion `gorm:"foreignKey:TemplateID" json:"versions,omitempty"`
}

func (EmailTemplate) TableName() string {
	return "email_templates"
}

func (t *EmailTemplate) Validate() error {
	if t == nil || strings.TrimSpace(t.Code) == "" {
		return ErrTemplateCodeRequired
	}
	if strings.TrimSpace(t.Locale) == "" {
		return ErrTemplateLocaleRequired
	}
	switch strings.TrimSpace(t.Category) {
	case TemplateCategoryOrder, TemplateCategoryAfterSales, TemplateCategorySystem:
	default:
		return ErrTemplateCategoryInvalid
	}
	if strings.TrimSpace(t.Name) == "" {
		return ErrTemplateNameRequired
	}
	if strings.TrimSpace(t.SubjectTemplate) == "" {
		return ErrTemplateSubjectRequired
	}
	if strings.TrimSpace(t.BodyHTML) == "" && strings.TrimSpace(t.BodyText) == "" {
		return ErrTemplateBodyRequired
	}
	if t.Version <= 0 {
		return ErrTemplateVersionInvalid
	}
	return nil
}

// EmailTemplateVersion is an immutable audit snapshot of a template write.
type EmailTemplateVersion struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	TemplateID        uint           `gorm:"not null;index;uniqueIndex:idx_email_template_version_number" json:"template_id"`
	Code              string         `gorm:"size:64;not null;index" json:"code"`
	Locale            string         `gorm:"size:16;not null;index" json:"locale"`
	Version           int            `gorm:"not null;uniqueIndex:idx_email_template_version_number" json:"version"`
	Name              string         `gorm:"size:128;not null" json:"name"`
	SubjectTemplate   string         `gorm:"size:255;not null" json:"subject_template"`
	BodyHTML          string         `gorm:"type:text;not null" json:"body_html"`
	BodyText          string         `gorm:"type:text;not null" json:"body_text"`
	AllowedVariables  datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"allowed_variables"`
	RequiredVariables datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"required_variables"`
	ChangedByUserID   *uint          `gorm:"index" json:"changed_by_user_id,omitempty"`
	ChangeReason      string         `gorm:"size:255;not null;default:''" json:"change_reason"`
	CreatedAt         time.Time      `json:"created_at"`
	Template          *EmailTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}

func (EmailTemplateVersion) TableName() string {
	return "email_template_versions"
}
