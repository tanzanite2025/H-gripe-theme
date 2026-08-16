package seo

import "time"

const (
	RouteSourceStatic  = "static"
	RouteSourceProduct = "product"
	RouteSourceBlog    = "blog"
	RouteSourceAlias   = "alias"

	RouteEntryStatusActive    = "active"
	RouteEntryStatusAlias     = "alias"
	RouteEntryStatusDuplicate = "duplicate"
	RouteEntryStatusStale     = "stale"

	RouteCheckStatusOK              = "ok"
	RouteCheckStatusRedirect        = "redirect"
	RouteCheckStatusNotFound        = "not_found"
	RouteCheckStatusServerError     = "server_error"
	RouteCheckStatusCanonicalMisfit = "canonical_mismatch"
	RouteCheckStatusError           = "error"
)

// StorefrontRouteCatalogEntry is the backend-owned snapshot of a storefront
// URL. It keeps route identity separate from the content resource that
// currently resolves the route.
type StorefrontRouteCatalogEntry struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	RouteKey          string     `gorm:"size:255;not null;uniqueIndex" json:"route_key"`
	Path              string     `gorm:"type:text;not null;index" json:"path"`
	Locale            string     `gorm:"size:20;not null;index" json:"locale"`
	SourceType        string     `gorm:"size:32;not null;index" json:"source_type"`
	SourceID          *uint      `gorm:"index" json:"source_id,omitempty"`
	SourceKey         string     `gorm:"size:255;index" json:"source_key"`
	Title             string     `gorm:"type:text" json:"title"`
	Summary           string     `gorm:"type:text" json:"summary"`
	CanonicalPath     string     `gorm:"type:text" json:"canonical_path"`
	IsAlias           bool       `gorm:"not null;default:false;index" json:"is_alias"`
	IsSearchable      bool       `gorm:"not null;default:true;index" json:"is_searchable"`
	IsCheckable       bool       `gorm:"not null;default:true;index" json:"is_checkable"`
	IsIndexable       bool       `gorm:"not null;default:true" json:"is_indexable"`
	EntryStatus       string     `gorm:"size:32;not null;default:'active';index" json:"entry_status"`
	DuplicateGroupKey string     `gorm:"size:255;index" json:"duplicate_group_key"`
	ManifestVersion   string     `gorm:"size:64" json:"manifest_version"`
	LastSeenAt        time.Time  `gorm:"not null;index" json:"last_seen_at"`
	LastCheckStatus   string     `gorm:"size:32;index" json:"last_check_status"`
	LastHTTPStatus    int        `json:"last_http_status"`
	LastFinalURL      string     `gorm:"type:text" json:"last_final_url"`
	LastCanonicalURL  string     `gorm:"type:text" json:"last_canonical_url"`
	LastResponseMS    int        `json:"last_response_ms"`
	LastRedirectCount int        `json:"last_redirect_count"`
	LastContentHash   string     `gorm:"size:64" json:"last_content_hash"`
	LastCheckError    string     `gorm:"type:text" json:"last_check_error"`
	LastCheckedAt     *time.Time `json:"last_checked_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (StorefrontRouteCatalogEntry) TableName() string {
	return "storefront_route_catalog_entries"
}

type StorefrontRouteCheckResult struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RouteEntryID  uint      `gorm:"not null;index" json:"route_entry_id"`
	CheckedAt     time.Time `gorm:"not null;index" json:"checked_at"`
	HTTPStatus    int       `json:"http_status"`
	FinalURL      string    `gorm:"type:text" json:"final_url"`
	CanonicalURL  string    `gorm:"type:text" json:"canonical_url"`
	ResponseMS    int       `json:"response_ms"`
	RedirectCount int       `json:"redirect_count"`
	ContentHash   string    `gorm:"size:64;index" json:"content_hash"`
	Status        string    `gorm:"size:32;not null;index" json:"status"`
	ErrorMessage  string    `gorm:"type:text" json:"error_message"`
}

func (StorefrontRouteCheckResult) TableName() string {
	return "storefront_route_check_results"
}

type StorefrontRouteCatalogStats struct {
	Total             int64      `json:"total"`
	Active            int64      `json:"active"`
	Alias             int64      `json:"alias"`
	Duplicate         int64      `json:"duplicate"`
	Stale             int64      `json:"stale"`
	NeedsAttention    int64      `json:"needs_attention"`
	Checked           int64      `json:"checked"`
	Unchecked         int64      `json:"unchecked"`
	OK                int64      `json:"ok"`
	Redirects         int64      `json:"redirects"`
	NotFound          int64      `json:"not_found"`
	ServerErrors      int64      `json:"server_errors"`
	CanonicalMismatch int64      `json:"canonical_mismatch"`
	Errors            int64      `json:"errors"`
	Searchable        int64      `json:"searchable"`
	Checkable         int64      `json:"checkable"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	ManifestVersion   string     `json:"manifest_version"`
}

type StorefrontRouteManifest struct {
	Version string                         `json:"version"`
	Routes  []StorefrontRouteManifestRoute `json:"routes"`
}

type StorefrontRouteManifestRoute struct {
	Key           string `json:"key"`
	Path          string `json:"path"`
	Label         string `json:"label"`
	Description   string `json:"description"`
	CanonicalPath string `json:"canonical_path"`
	IsAlias       bool   `json:"is_alias"`
	IsSearchable  bool   `json:"is_searchable"`
	IsCheckable   bool   `json:"is_checkable"`
	IsIndexable   bool   `json:"is_indexable"`
}
