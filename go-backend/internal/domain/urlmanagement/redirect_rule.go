package urlmanagement

import "time"

const (
	RedirectRuleStateDraft     = "draft"
	RedirectRuleStatePublished = "published"
	RedirectRuleStateDisabled  = "disabled"
)

// StorefrontRedirectRule is an Admin-managed exact storefront path redirect.
// Static aliases remain owned by the storefront route registry.
type StorefrontRedirectRule struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	SourcePath    string     `gorm:"size:2048;not null;uniqueIndex" json:"source_path"`
	TargetPath    string     `gorm:"size:2048;not null" json:"target_path"`
	StatusCode    int        `gorm:"not null;default:301" json:"status_code"`
	State         string     `gorm:"size:32;not null;default:'draft';index" json:"state"`
	Reason        string     `gorm:"type:text;not null;default:''" json:"reason"`
	CreatedByID   uint       `gorm:"not null;default:0" json:"created_by_id"`
	PublishedByID *uint      `gorm:"index" json:"published_by_id,omitempty"`
	PublishedAt   *time.Time `gorm:"index" json:"published_at,omitempty"`
	DisabledAt    *time.Time `gorm:"index" json:"disabled_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (StorefrontRedirectRule) TableName() string {
	return "storefront_redirect_rules"
}

type StorefrontRedirectRuleInput struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	StatusCode int    `json:"status_code"`
	Reason     string `json:"reason"`
}

type StorefrontPublishedRedirect struct {
	SourcePath string `json:"source_path"`
	TargetPath string `json:"target_path"`
	StatusCode int    `json:"status_code"`
}
