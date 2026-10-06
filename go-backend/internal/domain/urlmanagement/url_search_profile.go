package urlmanagement

import (
	"time"

	"commerce-platform/internal/domain/seo"
	"gorm.io/datatypes"
)

// StorefrontURLSearchProfile is the controlled search overlay for a storefront
// URL. It keeps search keywords and display overrides separate from the route
// snapshot so URL management never becomes a second source of truth.
type StorefrontURLSearchProfile struct {
	ID             uint                             `gorm:"primaryKey" json:"id"`
	RouteEntryID   uint                             `gorm:"not null;uniqueIndex;index" json:"route_entry_id"`
	Enabled        bool                             `gorm:"not null;default:true;index" json:"enabled"`
	SearchWeight   int                              `gorm:"not null;default:100;index" json:"search_weight"`
	Keywords       datatypes.JSONSlice[string]      `gorm:"column:keywords_json;type:jsonb;not null;default:'[]'" json:"keywords"`
	DisplayTitle   string                           `gorm:"type:text;not null;default:''" json:"display_title"`
	DisplaySummary string                           `gorm:"type:text;not null;default:''" json:"display_summary"`
	CreatedAt      time.Time                        `json:"created_at"`
	UpdatedAt      time.Time                        `json:"updated_at"`
	RouteEntry     *seo.StorefrontRouteCatalogEntry `gorm:"foreignKey:RouteEntryID;references:ID" json:"route_entry,omitempty"`
}

func (StorefrontURLSearchProfile) TableName() string {
	return "storefront_url_search_profiles"
}

type StorefrontURLSearchProfileInput struct {
	Enabled        bool     `json:"enabled"`
	SearchWeight   int      `json:"search_weight"`
	Keywords       []string `json:"keywords"`
	DisplayTitle   string   `json:"display_title"`
	DisplaySummary string   `json:"display_summary"`
}

// StorefrontURLSearchProfileStats describes both the explicit admin
// configuration and the effective public search index for one locale scope.
// A route without a profile can still be included in the public index when
// the route itself is active, searchable, and indexable.
type StorefrontURLSearchProfileStats struct {
	TotalRoutes      int64 `json:"total_routes"`
	PublicIndexed    int64 `json:"public_indexed"`
	ExplicitProfiles int64 `json:"explicit_profiles"`
	ExplicitEnabled  int64 `json:"explicit_enabled"`
	KeywordCount     int64 `json:"keyword_count"`
	Unconfigured     int64 `json:"unconfigured"`
}
