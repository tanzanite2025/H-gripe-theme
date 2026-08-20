package sitequality

import "time"

const (
	SiteQualityStrategyMobile  = "mobile"
	SiteQualityStrategyDesktop = "desktop"

	SiteQualityRunStatusSuccess = "success"
	SiteQualityRunStatusFailed  = "failed"
)

// SiteQualityRun is the durable evidence captured for a Lighthouse runner
// request. Raw provider responses stay server-side while normalized findings
// are exposed through the service-center API.
type SiteQualityRun struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	TargetID                 *uint     `gorm:"index" json:"target_id,omitempty"`
	JobID                    *uint     `gorm:"index" json:"job_id,omitempty"`
	TargetURL                string    `gorm:"type:text;not null;index" json:"target_url"`
	CanonicalURL             string    `gorm:"type:text;not null;default:'';index" json:"canonical_url"`
	FinalURL                 string    `gorm:"type:text;not null;default:''" json:"final_url"`
	Strategy                 string    `gorm:"size:16;not null;index" json:"strategy"`
	Status                   string    `gorm:"size:16;not null;index" json:"status"`
	InitiatedByUserID        uint      `gorm:"not null;default:0;index" json:"initiated_by_user_id"`
	Provider                 string    `gorm:"size:32;not null;default:'lighthouse_runner'" json:"provider"`
	LighthouseVersion        string    `gorm:"size:64;not null;default:''" json:"lighthouse_version,omitempty"`
	EnvironmentJSON          string    `gorm:"type:jsonb;not null;default:'{}'" json:"environment,omitempty"`
	ReleaseID                string    `gorm:"size:128;not null;default:'';index" json:"release_id,omitempty"`
	PerformanceScore         *int      `json:"performance_score,omitempty"`
	AccessibilityScore       *int      `json:"accessibility_score,omitempty"`
	BestPracticesScore       *int      `json:"best_practices_score,omitempty"`
	SEOScore                 *int      `json:"seo_score,omitempty"`
	FirstContentfulPaintMS   *float64  `json:"first_contentful_paint_ms,omitempty"`
	LargestContentfulPaintMS *float64  `json:"largest_contentful_paint_ms,omitempty"`
	InteractionToNextPaintMS *float64  `json:"interaction_to_next_paint_ms,omitempty"`
	CumulativeLayoutShift    *float64  `json:"cumulative_layout_shift,omitempty"`
	TotalBlockingTimeMS      *float64  `json:"total_blocking_time_ms,omitempty"`
	SpeedIndexMS             *float64  `json:"speed_index_ms,omitempty"`
	IssuesJSON               string    `gorm:"type:jsonb;not null;default:'[]'" json:"-"`
	RawResponseJSON          string    `gorm:"type:jsonb;not null;default:'{}'" json:"-"`
	ErrorMessage             string    `gorm:"type:text;not null;default:''" json:"error_message,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (SiteQualityRun) TableName() string {
	return "site_quality_runs"
}

// SiteQualityRunArchive mirrors the immutable run shape without retaining
// foreign keys to the hot table.
type SiteQualityRunArchive struct {
	SiteQualityRun
}

func (SiteQualityRunArchive) TableName() string {
	return "site_quality_runs_archive"
}
