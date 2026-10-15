package service

import (
	"fmt"
	"strings"
	"unicode"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/pkg/faqcontent"
	"commerce-platform/internal/repository"
)

func (s *FAQService) normalizeFAQContent(item *faq.FAQ, mutations ...repository.FAQMediaFieldMutation) error {
	mutation := repository.FAQMediaFieldMutation{UpdateAnswer: true, UpdateAnswerImage: true}
	if len(mutations) > 0 {
		mutation = mutations[0]
	}
	if mutation.UpdateAnswer {
		answer, err := faqcontent.SanitizeAnswer(item.Answer)
		if err != nil {
			return err
		}
		if !faqcontent.HasVisibleText(answer) {
			return fmt.Errorf("answer is required")
		}
		item.Answer = answer
	}
	if !mutation.UpdateAnswerImage {
		return nil
	}

	item.AnswerImageURL = strings.TrimSpace(item.AnswerImageURL)
	item.AnswerImageAlt = strings.TrimSpace(item.AnswerImageAlt)
	if item.AnswerImageURL == "" {
		item.AnswerImageAlt = ""
		item.AnswerImageWidth = 0
		item.AnswerImageHeight = 0
		return nil
	}
	// Identity, MIME type and actual dimensions are checked from the locked
	// media asset in the repository, never inferred from URL substrings.
	if item.AnswerImageWidth != 800 || item.AnswerImageHeight != 800 {
		return fmt.Errorf("answer image must be exactly 800x800 pixels")
	}
	if item.AnswerImageAlt == "" {
		return fmt.Errorf("answer image alt text is required")
	}
	return nil
}

func normalizeFAQStatus(status, fallback string) string {
	status = strings.TrimSpace(status)
	switch status {
	case "active", "hidden", "published", "draft":
		return status
	default:
		return fallback
	}
}

func slugifyFAQKey(value string) string {
	var builder strings.Builder
	previousDash := false
	for _, r := range strings.ToLower(value) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			previousDash = false
			continue
		}
		if !previousDash && builder.Len() > 0 {
			builder.WriteByte('-')
		}
		previousDash = true
	}
	return strings.Trim(builder.String(), "-")
}

func normalizeRoutePath(routePath string) string {
	routePath = strings.TrimSpace(routePath)
	if routePath == "" {
		return "/"
	}
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}
	routePath = strings.TrimRight(routePath, "/")
	if routePath == "" {
		return "/"
	}
	return routePath
}
