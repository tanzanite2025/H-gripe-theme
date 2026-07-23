package admin

import (
	"strings"
	"tanzanite/internal/service"
)

type FAQHandler struct {
	faqService *service.FAQService
}

func NewFAQHandler(faqService *service.FAQService) *FAQHandler {
	return &FAQHandler{
		faqService: faqService,
	}
}

func isFAQValidationError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "required") ||
		strings.Contains(message, "does not exist") ||
		strings.Contains(message, "hidden") ||
		strings.Contains(message, "answer image")
}
