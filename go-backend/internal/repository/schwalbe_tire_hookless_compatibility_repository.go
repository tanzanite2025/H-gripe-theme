package repository

import "time"

// SchwalbeTireHooklessCompatibility is a separately sourced Hookless fact.
// A row may target one Article No. or a model fallback. Missing rows are
// intentionally treated as unknown by the selector service.
type SchwalbeTireHooklessCompatibility struct {
	ScopeKey        string    `json:"scope_key"`
	ArticleNo       *string   `json:"article_no,omitempty"`
	ModelName       *string   `json:"model_name,omitempty"`
	Status          string    `json:"status"`
	SourceBasis     string    `json:"source_basis"`
	SourceVersion   string    `json:"source_version"`
	SourceURL       string    `json:"source_url"`
	SourceCheckedAt time.Time `json:"source_checked_at"`
}

// ListSchwalbeTireHooklessCompatibilities returns explicit compatibility
// records in article-first order. Older test databases and deployments can
// lack migration 369; an absent table is equivalent to no reviewed facts and
// must not make the ordinary catalog selector fail.
func (r *ProductRepository) ListSchwalbeTireHooklessCompatibilities() ([]SchwalbeTireHooklessCompatibility, error) {
	compatibilities := make([]SchwalbeTireHooklessCompatibility, 0)
	if r == nil || r.db == nil || !r.db.Migrator().HasTable("schwalbe_tire_hookless_compatibility") {
		return compatibilities, nil
	}

	err := r.db.Table("schwalbe_tire_hookless_compatibility").
		Select(`
			scope_key,
			article_no,
			model_name,
			status,
			source_basis,
			source_version,
			source_url,
			source_checked_at`).
		Order("CASE WHEN article_no IS NULL THEN 1 ELSE 0 END ASC").
		Order("article_no ASC").
		Order("model_name ASC").
		Scan(&compatibilities).Error
	if err != nil {
		return nil, err
	}
	return compatibilities, nil
}
