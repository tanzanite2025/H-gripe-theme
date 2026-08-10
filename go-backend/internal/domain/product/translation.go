package product

type ProductTranslation struct {
	ID       uint   `json:"id"`
	ParentID *uint  `json:"parent_id"`
	Locale   string `json:"locale"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	SKU      string `json:"sku"`
	Status   string `json:"status"`
	IsRoot   bool   `json:"is_root"`
}

type ProductTranslationGroup struct {
	RootID         uint                 `json:"root_id"`
	SourceID       uint                 `json:"source_id"`
	Translations   []ProductTranslation `json:"translations"`
	MissingLocales []string             `json:"missing_locales"`
}

// ProductTranslationRoute is the public route projection for one localized
// product page. It intentionally excludes catalog fields and internal IDs.
type ProductTranslationRoute struct {
	Locale string
	Slug   string
}
