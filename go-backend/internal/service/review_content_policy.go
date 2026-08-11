package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/review"
	"commerce-platform/internal/pkg/ugc"
)

const (
	maxReviewTitleRunes   = 200
	maxReviewContentRunes = 5000
	maxReviewImages       = 6
)

var (
	ErrReviewRequired        = errors.New("review is required")
	ErrReviewTitleRequired   = errors.New("review title is required")
	ErrReviewContentRequired = errors.New("review content is required")
	ErrReviewContentTooLong  = errors.New("review content is too long")
	ErrReviewTitleTooLong    = errors.New("review title is too long")
	ErrReviewImagesInvalid   = errors.New("review images contain an invalid URL")
	ErrReviewTooManyImages   = errors.New("review includes too many images")
)

func normalizeReviewSubmission(item *review.Review) error {
	if item == nil {
		return ErrReviewRequired
	}

	title, err := ugc.PlainText(item.Title, maxReviewTitleRunes)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrReviewTitleTooLong
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(title) == "" {
		return ErrReviewTitleRequired
	}

	content, err := ugc.PlainText(item.Content, maxReviewContentRunes)
	if errors.Is(err, ugc.ErrTextTooLong) {
		return ErrReviewContentTooLong
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return ErrReviewContentRequired
	}
	images, err := normalizeReviewImages(item.Images)
	if err != nil {
		return err
	}

	item.Title = title
	item.Content = content
	item.Images = images
	return nil
}

func normalizeReviewForPublic(item *review.Review) {
	if item == nil {
		return
	}
	item.Title = normalizeReviewPublicText(item.Title)
	item.Content = normalizeReviewPublicText(item.Content)
	item.ReplyContent = normalizeReviewPublicText(item.ReplyContent)
	images, err := normalizeReviewImages(item.Images)
	if err != nil {
		item.Images = "[]"
		return
	}
	item.Images = images
}

func normalizeReviewPublicText(value string) string {
	normalized, err := ugc.PlainText(value, 0)
	if err != nil {
		return ""
	}
	return normalized
}

func normalizeReviewImages(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "[]", nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return "", ErrReviewImagesInvalid
	}
	if len(values) > maxReviewImages {
		return "", ErrReviewTooManyImages
	}

	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !isAllowedReviewImageURL(value) {
			return "", ErrReviewImagesInvalid
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	encoded, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func isAllowedReviewImageURL(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil {
		return false
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		return parsed.Host != ""
	case "":
		return parsed.Host == "" && strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//")
	default:
		return false
	}
}
