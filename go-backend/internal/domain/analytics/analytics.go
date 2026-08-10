package analytics

const Group = "analytics"

type Settings struct {
	GoogleAnalytics  string `json:"google_analytics"`
	GoogleTagManager string `json:"google_tag_manager"`
}

type UpdateRequest struct {
	Locale           string  `json:"locale"`
	GoogleAnalytics  *string `json:"google_analytics"`
	GoogleTagManager *string `json:"google_tag_manager"`
}
