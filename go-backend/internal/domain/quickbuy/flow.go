package quickbuy

import (
	"time"

	"commerce-platform/internal/domain/product"

	"gorm.io/datatypes"
)

const (
	FlowVersionStatusDraft     = "draft"
	FlowVersionStatusPublished = "published"
	FlowVersionStatusArchived  = "archived"

	SelectionModeSingle   = "single"
	SelectionModeMultiple = "multiple"
	SelectionModeQuantity = "quantity"
	SelectionModeAuto     = "auto"
)

type Flow struct {
	ID           uint              `gorm:"primarykey" json:"id"`
	Slug         string            `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Name         string            `gorm:"size:160;not null" json:"name"`
	Description  string            `gorm:"type:text;not null;default:''" json:"description"`
	HelpText     string            `gorm:"type:text;not null;default:''" json:"help_text"`
	EntrySurface string            `gorm:"size:80;not null;default:'dock';index" json:"entry_surface"`
	IsEnabled    bool              `gorm:"not null;default:true;index" json:"is_enabled"`
	SortOrder    int               `gorm:"not null;default:100" json:"sort_order"`
	Translations []FlowTranslation `gorm:"foreignKey:FlowID;constraint:OnDelete:CASCADE" json:"translations,omitempty"`
	Versions     []Version         `gorm:"foreignKey:FlowID" json:"versions,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func (Flow) TableName() string {
	return "quick_buy_flows"
}

type FlowTranslation struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	FlowID    uint      `gorm:"not null;uniqueIndex:idx_quick_buy_flow_translations_flow_locale" json:"flow_id"`
	Locale    string    `gorm:"size:32;not null;uniqueIndex:idx_quick_buy_flow_translations_flow_locale;index" json:"locale"`
	HelpText  string    `gorm:"type:text;not null;default:''" json:"help_text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FlowTranslation) TableName() string {
	return "quick_buy_flow_translations"
}

type Version struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	FlowID        uint       `gorm:"not null;index" json:"flow_id"`
	VersionNumber int        `gorm:"not null;default:1" json:"version_number"`
	Status        string     `gorm:"size:24;not null;default:'draft';index" json:"status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	PublishedBy   *uint      `gorm:"index" json:"published_by,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
	Flow          *Flow      `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	Steps         []Step     `gorm:"foreignKey:FlowVersionID" json:"steps,omitempty"`
	Rules         []Rule     `gorm:"foreignKey:FlowVersionID" json:"rules,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Version) TableName() string {
	return "quick_buy_flow_versions"
}

type Step struct {
	ID              uint              `gorm:"primarykey" json:"id"`
	FlowVersionID   uint              `gorm:"not null;index" json:"flow_version_id"`
	StepKey         string            `gorm:"size:120;not null" json:"step_key"`
	Name            string            `gorm:"size:160;not null" json:"name"`
	SortOrder       int               `gorm:"not null;default:100" json:"sort_order"`
	SelectionMode   string            `gorm:"size:24;not null;default:'single'" json:"selection_mode"`
	IsRequired      bool              `gorm:"not null;default:true" json:"is_required"`
	MinSelect       int               `gorm:"not null;default:0" json:"min_select"`
	MaxSelect       int               `gorm:"not null;default:1" json:"max_select"`
	DefaultQuantity int               `gorm:"not null;default:1" json:"default_quantity"`
	AllowSkip       bool              `gorm:"not null;default:false" json:"allow_skip"`
	ProductTypes    []StepProductType `gorm:"foreignKey:StepID" json:"product_types,omitempty"`
	Filters         []StepFilter      `gorm:"foreignKey:StepID" json:"filters,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

func (Step) TableName() string {
	return "quick_buy_steps"
}

type StepProductType struct {
	ID            uint                 `gorm:"primarykey" json:"id"`
	StepID        uint                 `gorm:"not null;index" json:"step_id"`
	ProductTypeID uint                 `gorm:"not null;index" json:"product_type_id"`
	IsPrimary     bool                 `gorm:"not null;default:false" json:"is_primary"`
	SortOrder     int                  `gorm:"not null;default:100" json:"sort_order"`
	ProductType   *product.ProductType `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

func (StepProductType) TableName() string {
	return "quick_buy_step_product_types"
}

type StepFilter struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	StepID           uint           `gorm:"not null;index" json:"step_id"`
	FilterType       string         `gorm:"size:40;not null" json:"filter_type"`
	SpecDefinitionID *uint          `gorm:"index" json:"spec_definition_id,omitempty"`
	Operator         string         `gorm:"size:24;not null;default:'eq'" json:"operator"`
	Value            datatypes.JSON `gorm:"type:jsonb;not null;default:'null'" json:"value"`
	SortOrder        int            `gorm:"not null;default:100" json:"sort_order"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (StepFilter) TableName() string {
	return "quick_buy_step_filters"
}

type Rule struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	FlowVersionID uint           `gorm:"not null;index" json:"flow_version_id"`
	RuleKey       string         `gorm:"size:120;not null" json:"rule_key"`
	RuleType      string         `gorm:"size:40;not null" json:"rule_type"`
	SourceStepKey string         `gorm:"size:120;not null;default:''" json:"source_step_key"`
	SourceSpecKey string         `gorm:"size:120;not null;default:''" json:"source_spec_key"`
	TargetStepKey string         `gorm:"size:120;not null;default:''" json:"target_step_key"`
	TargetSpecKey string         `gorm:"size:120;not null;default:''" json:"target_spec_key"`
	Rule          datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"rule"`
	Severity      string         `gorm:"size:16;not null;default:'error'" json:"severity"`
	MessageKey    string         `gorm:"size:160;not null;default:''" json:"message_key"`
	IsEnabled     bool           `gorm:"not null;default:true" json:"is_enabled"`
	SortOrder     int            `gorm:"not null;default:100" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (Rule) TableName() string {
	return "quick_buy_compatibility_rules"
}
