package repository

import "time"

type SiteQualityTargetStats struct {
	Total      int64 `json:"total"`
	Enabled    int64 `json:"enabled"`
	Due        int64 `json:"due"`
	Critical   int64 `json:"critical"`
	Standard   int64 `json:"standard"`
	Background int64 `json:"background"`
}

type SiteQualityJobStats struct {
	Total              int64      `json:"total"`
	Queued             int64      `json:"queued"`
	Processing         int64      `json:"processing"`
	Succeeded          int64      `json:"succeeded"`
	Failed             int64      `json:"failed"`
	DeadLetter         int64      `json:"dead_letter"`
	Claimable          int64      `json:"claimable"`
	StaleLeases        int64      `json:"stale_leases"`
	OldestQueuedAt     *time.Time `json:"oldest_queued_at,omitempty"`
	OldestProcessingAt *time.Time `json:"oldest_processing_at,omitempty"`
	LatestSuccessAt    *time.Time `json:"latest_success_at,omitempty"`
	LatestFailureAt    *time.Time `json:"latest_failure_at,omitempty"`
	LatestDeadLetterAt *time.Time `json:"latest_dead_letter_at,omitempty"`
}

type SiteQualityJobCleanupResult struct {
	Deleted    int64 `json:"deleted"`
	Failed     int64 `json:"failed"`
	DeadLetter int64 `json:"dead_letter"`
}

type SiteQualityProviderSlotStats struct {
	Provider        string     `json:"provider"`
	Configured      int        `json:"configured"`
	Total           int64      `json:"total"`
	Available       int64      `json:"available"`
	Locked          int64      `json:"locked"`
	StaleLocked     int64      `json:"stale_locked"`
	NextAvailableAt *time.Time `json:"next_available_at,omitempty"`
}
