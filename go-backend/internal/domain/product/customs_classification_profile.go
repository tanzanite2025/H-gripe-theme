package product

import (
	"time"

	"gorm.io/datatypes"
)

const (
	CustomsClassificationStatusDraft  = "draft"
	CustomsClassificationStatusActive = "active"
	CustomsClassificationStatusPaused = "paused"

	CustomsTradeRemedyRiskLevelNone     = "none"
	CustomsTradeRemedyRiskLevelLow      = "low"
	CustomsTradeRemedyRiskLevelMedium   = "medium"
	CustomsTradeRemedyRiskLevelHigh     = "high"
	CustomsTradeRemedyRiskLevelCritical = "critical"

	CustomsTradeRemedyRiskTagEUAntiDumpingAttention              = "eu_anti_dumping_attention"
	CustomsTradeRemedyRiskTagEUCompleteWheelsetAntiCircumvention = "eu_complete_wheelset_anti_circumvention"
	CustomsTradeRemedyRiskTagUSSection301List3Review             = "us_section_301_list_3_review"
)

// CustomsClassificationProfile stores reusable customs facts as independent
// master data. Products select a profile and orders keep an immutable snapshot.
type CustomsClassificationProfile struct {
	ID                           uint                        `gorm:"primarykey" json:"id"`
	Name                         string                      `gorm:"size:120;not null" json:"name"`
	Slug                         string                      `gorm:"size:140;uniqueIndex;not null" json:"slug"`
	ComponentKind                string                      `gorm:"size:64;not null;default:''" json:"component_kind"`
	Material                     string                      `gorm:"size:64;not null;default:''" json:"material"`
	HSCode                       string                      `gorm:"column:hs_code;size:12;not null" json:"hs_code"`
	CNCode                       string                      `gorm:"column:cn_code;size:12;not null;default:''" json:"cn_code"`
	CountryOfOrigin              string                      `gorm:"column:country_of_origin;size:2;not null;default:''" json:"country_of_origin"`
	CustomsDescription           string                      `gorm:"column:customs_description;size:255;not null;default:''" json:"customs_description"`
	Source                       string                      `gorm:"size:32;not null;default:''" json:"source"`
	SourceCode                   string                      `gorm:"size:64;not null;default:''" json:"source_code"`
	SourceURL                    string                      `gorm:"type:text;not null;default:''" json:"source_url"`
	SourceURLUS                  string                      `gorm:"column:source_url_us;type:text;not null;default:''" json:"source_url_us"`
	SourceURLEU                  string                      `gorm:"column:source_url_eu;type:text;not null;default:''" json:"source_url_eu"`
	SourceURLUK                  string                      `gorm:"column:source_url_uk;type:text;not null;default:''" json:"source_url_uk"`
	Notes                        string                      `gorm:"type:text;not null;default:''" json:"notes"`
	VerifiedAt                   *time.Time                  `gorm:"type:date" json:"verified_at"`
	ReviewDueAt                  *time.Time                  `gorm:"type:date" json:"review_due_at"`
	TradeRemedyRiskLevel         string                      `gorm:"column:trade_remedy_risk_level;size:16;not null;default:'none';index" json:"trade_remedy_risk_level"`
	TradeRemedyRiskTags          datatypes.JSONSlice[string] `gorm:"column:trade_remedy_risk_tags_json;type:jsonb;not null;default:'[]'" json:"trade_remedy_risk_tags"`
	TradeRemedyDeclarationAdvice string                      `gorm:"column:trade_remedy_declaration_advice;type:text;not null;default:''" json:"trade_remedy_declaration_advice"`
	Status                       string                      `gorm:"size:24;not null;default:'active';index" json:"status"`
	CreatedAt                    time.Time                   `json:"created_at"`
	UpdatedAt                    time.Time                   `json:"updated_at"`
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

func IsCustomsTradeRemedyRiskLevel(value string) bool {
	switch value {
	case CustomsTradeRemedyRiskLevelNone,
		CustomsTradeRemedyRiskLevelLow,
		CustomsTradeRemedyRiskLevelMedium,
		CustomsTradeRemedyRiskLevelHigh,
		CustomsTradeRemedyRiskLevelCritical:
		return true
	default:
		return false
	}
}

func IsCustomsTradeRemedyRiskTag(value string) bool {
	switch value {
	case CustomsTradeRemedyRiskTagEUAntiDumpingAttention,
		CustomsTradeRemedyRiskTagEUCompleteWheelsetAntiCircumvention,
		CustomsTradeRemedyRiskTagUSSection301List3Review:
		return true
	default:
		return false
	}
}
