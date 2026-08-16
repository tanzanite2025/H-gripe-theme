package wheelsetfit

import (
	"time"

	"gorm.io/datatypes"
)

const (
	QuestionnaireSlug          = "wheelset-fit"
	WheelsetProductCategorySlug = "wheelset"
	SourceLocale               = "zh_cn"

	VersionStatusDraft     = "draft"
	VersionStatusPublished = "published"
	VersionStatusArchived  = "archived"

	InputModeSingleChoice = "single_choice"
)

type Questionnaire struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	Slug                string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	ProductCategorySlug string    `gorm:"size:120;not null;default:'wheelset'" json:"product_category_slug"`
	SourceLocale        string    `gorm:"size:32;not null;default:'zh_cn'" json:"source_locale"`
	IsEnabled           bool      `gorm:"not null;default:true" json:"is_enabled"`
	Versions            []Version `gorm:"foreignKey:QuestionnaireID;constraint:OnDelete:CASCADE" json:"versions,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (Questionnaire) TableName() string {
	return "wheelset_fit_questionnaires"
}

type Version struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	QuestionnaireID uint           `gorm:"not null;index" json:"questionnaire_id"`
	VersionNumber   int            `gorm:"not null" json:"version_number"`
	Status          string         `gorm:"size:24;not null;default:'draft';index" json:"status"`
	PublishedAt     *time.Time     `json:"published_at,omitempty"`
	PublishedBy     *uint          `gorm:"index" json:"published_by,omitempty"`
	Questionnaire   *Questionnaire `gorm:"foreignKey:QuestionnaireID" json:"questionnaire,omitempty"`
	Questions       []Question     `gorm:"foreignKey:QuestionnaireVersionID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (Version) TableName() string {
	return "wheelset_fit_questionnaire_versions"
}

type Question struct {
	ID                     uint                  `gorm:"primarykey" json:"id"`
	QuestionnaireVersionID uint                  `gorm:"not null;index" json:"questionnaire_version_id"`
	QuestionKey            string                `gorm:"size:120;not null" json:"question_key"`
	AnswerKey              string                `gorm:"size:120;not null" json:"answer_key"`
	SortOrder              int                   `gorm:"not null;default:10" json:"sort_order"`
	InputMode              string                `gorm:"size:32;not null;default:'single_choice'" json:"input_mode"`
	IsRequired             bool                  `gorm:"not null;default:true" json:"is_required"`
	AllowUnknown           bool                  `gorm:"not null;default:true" json:"allow_unknown"`
	IsEnabled              bool                  `gorm:"not null;default:true" json:"is_enabled"`
	SourceRevision         int                   `gorm:"not null;default:1" json:"source_revision"`
	Options                []Option              `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"options,omitempty"`
	Translations           []QuestionTranslation `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE" json:"translations,omitempty"`
	CreatedAt              time.Time             `json:"created_at"`
	UpdatedAt              time.Time             `json:"updated_at"`
}

func (Question) TableName() string {
	return "wheelset_fit_questions"
}

type Option struct {
	ID                   uint              `gorm:"primarykey" json:"id"`
	QuestionID           uint              `gorm:"not null;index" json:"question_id"`
	OptionKey            string            `gorm:"size:120;not null" json:"option_key"`
	AnswerValue          string            `gorm:"size:160;not null" json:"answer_value"`
	SortOrder            int               `gorm:"not null;default:10" json:"sort_order"`
	IsUnknown            bool              `gorm:"not null;default:false" json:"is_unknown"`
	IsEnabled            bool              `gorm:"not null;default:true" json:"is_enabled"`
	ProductFilterEffects datatypes.JSON    `gorm:"type:jsonb;not null;default:'{}'" json:"product_filter_effects"`
	SourceRevision       int               `gorm:"not null;default:1" json:"source_revision"`
	Translations         []OptionTranslation `gorm:"foreignKey:OptionID;constraint:OnDelete:CASCADE" json:"translations,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

func (Option) TableName() string {
	return "wheelset_fit_question_options"
}

type QuestionTranslation struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	QuestionID          uint      `gorm:"not null;index" json:"question_id"`
	Locale              string    `gorm:"size:32;not null" json:"locale"`
	Prompt              string    `gorm:"type:text;not null;default:''" json:"prompt"`
	HelpTitle           string    `gorm:"type:text;not null;default:''" json:"help_title"`
	HelpBody            string    `gorm:"type:text;not null;default:''" json:"help_body"`
	SourceRevision      int       `gorm:"not null;default:1" json:"source_revision"`
	TranslatedRevision  int       `gorm:"not null;default:0" json:"translated_revision"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (QuestionTranslation) TableName() string {
	return "wheelset_fit_question_translations"
}

type OptionTranslation struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	OptionID           uint      `gorm:"not null;index" json:"option_id"`
	Locale             string    `gorm:"size:32;not null" json:"locale"`
	Label              string    `gorm:"type:text;not null;default:''" json:"label"`
	Description        string    `gorm:"type:text;not null;default:''" json:"description"`
	SourceRevision     int       `gorm:"not null;default:1" json:"source_revision"`
	TranslatedRevision int       `gorm:"not null;default:0" json:"translated_revision"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (OptionTranslation) TableName() string {
	return "wheelset_fit_question_option_translations"
}
