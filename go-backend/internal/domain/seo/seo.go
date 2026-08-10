package seo

const Group = "seo"

type Settings struct {
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`
}

type UpdateRequest struct {
	Locale          string  `json:"locale"`
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
}

type ArticleResourceUpdateRequest struct {
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
	CanonicalURL    *string `json:"canonical_url"`
}

type ProductResourceUpdateRequest struct {
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
}
