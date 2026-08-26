package seo

type CategoryResourceUpdateRequest struct {
	Locale          string  `json:"locale"`
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
	Intro           *string `json:"intro"`
}
