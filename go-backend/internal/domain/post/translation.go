package post

// TranslationRoute is the public route projection for one published article
// translation. It intentionally excludes article content and SEO fields.
type TranslationRoute struct {
	Locale string
	Slug   string
	Tags   string
}
