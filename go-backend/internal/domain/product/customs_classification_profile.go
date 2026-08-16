package product

import "time"

const (
	CustomsClassificationStatusDraft  = "draft"
	CustomsClassificationStatusActive = "active"
	CustomsClassificationStatusPaused = "paused"
)

// CustomsClassificationProfile stores reusable customs facts for a product family.
// Product rows may copy these values later, while orders keep their immutable item snapshot.
type CustomsClassificationProfile struct {
	ID                             uint                          `gorm:"primarykey" json:"id"`
	ProductSpecificationTemplateID *uint                         `gorm:"index" json:"product_specification_template_id,omitempty"`
	ProductSpecificationTemplate   *ProductSpecificationTemplate `gorm:"foreignKey:ProductSpecificationTemplateID;constraint:OnDelete:SET NULL" json:"product_specification_template,omitempty"`
	Name                           string                        `gorm:"size:120;not null" json:"name"`
	Slug                           string                        `gorm:"size:140;uniqueIndex;not null" json:"slug"`
	ComponentKind                  string                        `gorm:"size:64;not null;default:''" json:"component_kind"`
	Material                       string                        `gorm:"size:64;not null;default:''" json:"material"`
	HSCode                         string                        `gorm:"column:hs_code;size:12;not null" json:"hs_code"`
	CNCode                         string                        `gorm:"column:cn_code;size:12;not null;default:''" json:"cn_code"`
	CountryOfOrigin                string                        `gorm:"column:country_of_origin;size:2;not null;default:''" json:"country_of_origin"`
	CustomsDescription             string                        `gorm:"column:customs_description;size:255;not null;default:''" json:"customs_description"`
	Source                         string                        `gorm:"size:32;not null;default:''" json:"source"`
	SourceCode                     string                        `gorm:"size:64;not null;default:''" json:"source_code"`
	SourceURL                      string                        `gorm:"type:text;not null;default:''" json:"source_url"`
	Notes                          string                        `gorm:"type:text;not null;default:''" json:"notes"`
	Status                         string                        `gorm:"size:24;not null;default:'active';index" json:"status"`
	CreatedAt                      time.Time                     `json:"created_at"`
	UpdatedAt                      time.Time                     `json:"updated_at"`
}

func (CustomsClassificationProfile) TableName() string {
	return "customs_classification_profiles"
}

func IsCustomsClassificationStatus(value string) bool {
	switch value {
	case CustomsClassificationStatusDraft, CustomsClassificationStatusActive, CustomsClassificationStatusPaused:
		return true
	default:
		return false
	}
}
