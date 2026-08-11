package content

import "commerce-platform/internal/service"

type Handler struct {
	postService *service.PostService
	faqService  *service.FAQService
}

func NewHandler(postService *service.PostService, faqService *service.FAQService) *Handler {
	return &Handler{
		postService: postService,
		faqService:  faqService,
	}
}
