package service

import "strings"

func normalizeMediaType(value string) (string, error) {
	mediaType := strings.ToLower(strings.TrimSpace(value))
	if mediaType == "" {
		mediaType = "image"
	}
	switch mediaType {
	case "image", "video":
		return mediaType, nil
	default:
		return "", ErrUnsupportedMediaType
	}
}

func normalizeOptionalMediaType(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeMediaType(value)
}

func normalizeMediaStatus(value string) (string, error) {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "" {
		status = "active"
	}
	switch status {
	case "active", "archived":
		return status, nil
	default:
		return "", ErrUnsupportedMediaStatus
	}
}

func normalizeOptionalMediaStatus(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeMediaStatus(value)
}

func normalizeVisibility(value string) (string, error) {
	visibility := strings.ToLower(strings.TrimSpace(value))
	if visibility == "" {
		visibility = "public"
	}
	switch visibility {
	case "public", "private":
		return visibility, nil
	default:
		return "", ErrUnsupportedVisibility
	}
}

func normalizeOptionalVisibility(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return normalizeVisibility(value)
}
