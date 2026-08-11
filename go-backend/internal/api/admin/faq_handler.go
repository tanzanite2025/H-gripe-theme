package admin

import (
	"commerce-platform/internal/service"
	"errors"
	"strings"
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
		errors.Is(err, service.ErrFAQLocaleImmutable) ||
		errors.Is(err, service.ErrFAQCategoryIdentityImmutable) ||
		strings.Contains(message, "required") ||
		strings.Contains(message, "does not exist") ||
		strings.Contains(message, "hidden") ||
		strings.Contains(message, "answer image")
}
