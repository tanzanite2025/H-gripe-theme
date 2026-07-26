package admin

import (
	"errors"
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
	return errors.Is(err, service.ErrUnsupportedLocale) ||
		strings.Contains(message, "required") ||
		strings.Contains(message, "does not exist") ||
		strings.Contains(message, "hidden") ||
		strings.Contains(message, "answer image")
}
