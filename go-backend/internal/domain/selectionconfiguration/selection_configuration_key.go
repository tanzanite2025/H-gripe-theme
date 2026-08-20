package selectionconfiguration

import "time"

const (
	SelectionConfigurationKeyKindQuestionKey = "question_key"
	SelectionConfigurationKeyKindAnswerKey   = "answer_key"
)

type SelectionConfigurationKey struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Kind         string    `gorm:"size:32;not null;index:idx_selection_configuration_keys_kind_code,unique" json:"kind"`
	Code         string    `gorm:"size:120;not null;index:idx_selection_configuration_keys_kind_code,unique" json:"code"`
	DisplayLabel string    `gorm:"type:text;not null;default:''" json:"display_label"`
	Description  string    `gorm:"type:text;not null;default:''" json:"description"`
	IsEnabled    bool      `gorm:"not null" json:"is_enabled"`
	SortOrder    int       `gorm:"not null;default:10" json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SelectionConfigurationKey) TableName() string {
	return "selection_configuration_keys"
}
