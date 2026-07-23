package service

import (
	"fmt"
	"strings"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/pkg/faqcontent"
	"unicode"
)

func (s *FAQService) normalizeFAQContent(item *faq.FAQ) error {
	answer, err := faqcontent.SanitizeAnswer(item.Answer)
	if err != nil {
		return err
	}
	if !faqcontent.HasVisibleText(answer) {
		return fmt.Errorf("answer is required")
	}
	item.Answer = answer

	item.AnswerImageURL = strings.TrimSpace(item.AnswerImageURL)
	item.AnswerImageAlt = strings.TrimSpace(item.AnswerImageAlt)
	if item.AnswerImageURL == "" {
		item.AnswerImageAlt = ""
		item.AnswerImageWidth = 0
		item.AnswerImageHeight = 0
		return nil
	}
	if !strings.HasPrefix(item.AnswerImageURL, "/uploads/") && !strings.Contains(item.AnswerImageURL, "/uploads/") {
		return fmt.Errorf("answer image must be a managed upload")
	}
	if !strings.HasSuffix(strings.ToLower(strings.Split(item.AnswerImageURL, "?")[0]), ".webp") {
		return fmt.Errorf("answer image must be WebP")
	}
	if item.AnswerImageWidth != 800 || item.AnswerImageHeight != 800 {
		return fmt.Errorf("answer image must be exactly 800x800 pixels")
	}
	if item.AnswerImageAlt == "" {
		return fmt.Errorf("answer image alt text is required")
	}
	return nil
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	locale = strings.ReplaceAll(locale, "_", "-")
	if locale == "" {
		return "en"
	}
	if strings.HasPrefix(locale, "zh-") {
		return "zh"
	}
	return locale
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
