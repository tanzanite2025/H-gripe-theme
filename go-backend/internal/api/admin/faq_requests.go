package admin

type createFAQRequest struct {
	Question          string `json:"question" binding:"required"`
	Answer            string `json:"answer" binding:"required"`
	AnswerImageURL    string `json:"answer_image_url"`
	AnswerImageAlt    string `json:"answer_image_alt"`
	AnswerImageWidth  int    `json:"answer_image_width"`
	AnswerImageHeight int    `json:"answer_image_height"`
	PageID            string `json:"page_id"`
	Category          string `json:"category" binding:"required"`
	Locale            string `json:"locale" binding:"required"`
	Status            string `json:"status" binding:"required,oneof=draft published"`
	Order             int    `json:"order"`
}

type updateFAQRequest struct {
	Question          string `json:"question"`
	Answer            string `json:"answer"`
	AnswerImageURL    string `json:"answer_image_url"`
	AnswerImageAlt    string `json:"answer_image_alt"`
	AnswerImageWidth  int    `json:"answer_image_width"`
	AnswerImageHeight int    `json:"answer_image_height"`
	PageID            string `json:"page_id"`
	Category          string `json:"category"`
	Locale            string `json:"locale"`
	Status            string `json:"status" binding:"omitempty,oneof=draft published"`
	Order             int    `json:"order"`
}

type faqPageRequest struct {
	RoutePath string `json:"route_path"`
	Domain    string `json:"domain"`
	Locale    string `json:"locale" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Subtitle  string `json:"subtitle"`
	Status    string `json:"status" binding:"required,oneof=active hidden"`
	SortOrder int    `json:"sort_order"`
}

type faqCategoryRequest struct {
	PageID      string `json:"page_id" binding:"required"`
	CategoryKey string `json:"category_key"`
	Name        string `json:"name" binding:"required"`
	Icon        string `json:"icon"`
	Locale      string `json:"locale" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=active hidden"`
	SortOrder   int    `json:"sort_order"`
}

type updateFAQOrderRequest struct {
	Order int `json:"order" binding:"required"`
}

type batchDeleteFAQRequest struct {
	FAQIDs []uint `json:"faq_ids" binding:"required,min=1"`
}
