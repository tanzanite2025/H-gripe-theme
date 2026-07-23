package service

import "tanzanite/internal/domain/faq"

type FAQAdminUpdateInput struct {
	Question          string
	Answer            string
	AnswerImageURL    string
	AnswerImageAlt    string
	AnswerImageWidth  int
	AnswerImageHeight int
	AnswerImageSet    bool
	PageID            string
	Category          string
	Locale            string
	Status            string
	Order             int
}

type FAQPageAdminInput struct {
	RoutePath string
	Domain    string
	Locale    string
	Title     string
	Subtitle  string
	Status    string
	SortOrder int
}

type FAQCategoryAdminInput struct {
	PageID      string
	CategoryKey string
	Name        string
	Icon        string
	Locale      string
	Status      string
	SortOrder   int
}

type FAQCategoryAdminView struct {
	faq.FAQCategory
	FAQCount int64 `json:"faq_count"`
}

type FAQPageAdminView struct {
	faq.FAQPage
	FAQCount   int64                  `json:"faq_count"`
	Categories []FAQCategoryAdminView `json:"categories"`
}

type FAQPublicItem struct {
	ID                string   `json:"id"`
	Question          string   `json:"question"`
	Answer            string   `json:"answer"`
	AnswerImageURL    string   `json:"answerImageUrl,omitempty"`
	AnswerImageAlt    string   `json:"answerImageAlt,omitempty"`
	AnswerImageWidth  int      `json:"answerImageWidth,omitempty"`
	AnswerImageHeight int      `json:"answerImageHeight,omitempty"`
	Tags              []string `json:"tags"`
}

type FAQPublicCategory struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Icon  string          `json:"icon,omitempty"`
	Items []FAQPublicItem `json:"items"`
}

type FAQPublicPageData struct {
	PageID     string              `json:"pageId"`
	Title      string              `json:"title,omitempty"`
	Subtitle   string              `json:"subtitle,omitempty"`
	Categories []FAQPublicCategory `json:"categories"`
}
