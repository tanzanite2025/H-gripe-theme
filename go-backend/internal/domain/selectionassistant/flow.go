package selectionassistant

import (
	"time"

	"gorm.io/datatypes"
)

const (
	FlowVersionStatusDraft     = "draft"
	FlowVersionStatusPublished = "published"
	FlowVersionStatusArchived  = "archived"

	ConfigKind = "product_selection_assistant"

	NodeTypeQuestion = "question"
	NodeTypeTerminal = "terminal"
	NodeTypeSupport  = "support"

	WheelsetProductCategorySlug = "wheelset"
)

type Flow struct {
	ID                  uint      `gorm:"primarykey" json:"id"`
	Slug                string    `gorm:"size:120;uniqueIndex;not null" json:"slug"`
	Name                string    `gorm:"size:160;not null" json:"name"`
	Description         string    `gorm:"type:text;not null;default:''" json:"description"`
	ProductCategorySlug string    `gorm:"size:120;not null;default:'wheelset';index" json:"product_category_slug"`
	IsEnabled           bool      `gorm:"not null;default:true;index" json:"is_enabled"`
	SortOrder           int       `gorm:"not null;default:100" json:"sort_order"`
	Versions            []Version `gorm:"foreignKey:FlowID" json:"versions,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (Flow) TableName() string {
	return "selection_assistant_flows"
}

type Version struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	FlowID        uint           `gorm:"not null;index" json:"flow_id"`
	VersionNumber int            `gorm:"not null;default:1" json:"version_number"`
	Status        string         `gorm:"size:24;not null;default:'draft';index" json:"status"`
	Config        datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"config"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
	PublishedBy   *uint          `gorm:"index" json:"published_by,omitempty"`
	Flow          *Flow          `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (Version) TableName() string {
	return "selection_assistant_flow_versions"
}

type Config struct {
	Kind             string           `json:"kind"`
	SchemaVersion    int              `json:"schema_version"`
	EntryNodeKey     string           `json:"entry_node_key"`
	BaseProductQuery BaseProductQuery `json:"base_product_query"`
	Nodes            []Node           `json:"nodes"`
}

type BaseProductQuery struct {
	CategorySlug string `json:"category_slug"`
}

type Node struct {
	Key       string            `json:"key"`
	Type      string            `json:"type"`
	Prompt    map[string]string `json:"prompt,omitempty"`
	Helper    map[string]string `json:"helper,omitempty"`
	HelpTitle map[string]string `json:"help_title,omitempty"`
	HelpBody  map[string]string `json:"help_body,omitempty"`
	Options   []Option          `json:"options,omitempty"`
	Editor    EditorPosition    `json:"editor,omitempty"`
}

type Option struct {
	Key           string            `json:"key"`
	Label         map[string]string `json:"label,omitempty"`
	Description   map[string]string `json:"description,omitempty"`
	AnswerEffects map[string]string `json:"answer_effects,omitempty"`
	QueryEffects  QueryEffects      `json:"query_effects,omitempty"`
	NextNodeKey   string            `json:"next_node_key,omitempty"`
}

type QueryEffects struct {
	Keyword     string              `json:"keyword,omitempty"`
	SpecFilters map[string][]string `json:"spec_filters,omitempty"`
}

type EditorPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}
