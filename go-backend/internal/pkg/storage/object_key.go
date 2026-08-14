package storage

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func generateObjectKey(originalFilename string, prefix string) (string, error) {
	cleanName := filepath.Base(originalFilename)
	cleanName = strings.ReplaceAll(cleanName, "..", "")
	cleanName = strings.ReplaceAll(cleanName, "/", "")
	cleanName = strings.ReplaceAll(cleanName, "\\", "")

	ext := strings.ToLower(filepath.Ext(cleanName))
	id := uuid.New().String()
	datePath := time.Now().Format("2006/01/02")
	key := filepath.ToSlash(filepath.Join(datePath, fmt.Sprintf("%s%s", id, ext)))
	if strings.TrimSpace(prefix) == "" {
		return key, nil
	}

	cleanPrefix, ok := NormalizeObjectKey(prefix)
	if !ok {
		return "", fmt.Errorf("invalid object key prefix")
	}
	return cleanPrefix + "/" + key, nil
}

func NormalizeObjectKey(value string) (string, bool) {
	key := strings.Trim(strings.ReplaceAll(filepath.ToSlash(strings.TrimSpace(value)), "\\", "/"), "/")
	if key == "" {
		return "", false
	}
	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", false
		}
	}
	return key, true
}

func ObjectKeyFromReference(reference string, baseURL string) (string, bool) {
	value := strings.TrimSpace(reference)
	if value == "" {
		return "", false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", false
	}

	if parsed.IsAbs() || parsed.Host != "" {
		base, baseErr := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
		if baseErr == nil && base.Host != "" && strings.EqualFold(parsed.Host, base.Host) {
			basePath := strings.TrimRight(base.Path, "/")
			candidatePath := parsed.Path
			if basePath != "" {
				if candidatePath != basePath && !strings.HasPrefix(candidatePath, basePath+"/") {
					return "", false
				}
				candidatePath = strings.TrimPrefix(candidatePath, basePath)
			}
			candidatePath = strings.TrimPrefix(candidatePath, "/")
			if strings.HasPrefix(candidatePath, "uploads/") {
				return NormalizeObjectKey(strings.TrimPrefix(candidatePath, "uploads/"))
			}
		}
		return "", false
	}

	if markerIndex := strings.Index(parsed.Path, "/uploads/"); markerIndex >= 0 {
		return NormalizeObjectKey(parsed.Path[markerIndex+len("/uploads/"):])
	}
	return NormalizeObjectKey(parsed.Path)
}

func ObjectKeyFromBaseURL(reference string, baseURL string) (string, bool) {
	value := strings.TrimSpace(reference)
	configuredBaseURL := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if value == "" || configuredBaseURL == "" {
		return "", false
	}

	parsedReference, err := url.Parse(value)
	if err != nil || parsedReference.Host == "" {
		return "", false
	}
	parsedBase, err := url.Parse(configuredBaseURL)
	if err != nil || parsedBase.Host == "" || !strings.EqualFold(parsedReference.Host, parsedBase.Host) {
		return "", false
	}
	if parsedBase.Scheme != "" && !strings.EqualFold(parsedReference.Scheme, parsedBase.Scheme) {
		return "", false
	}

	basePath := strings.TrimRight(parsedBase.Path, "/")
	referencePath := parsedReference.Path
	if basePath != "" {
		if referencePath != basePath && !strings.HasPrefix(referencePath, basePath+"/") {
			return "", false
		}
		referencePath = strings.TrimPrefix(referencePath, basePath)
	}
	return NormalizeObjectKey(strings.TrimPrefix(referencePath, "/"))
}

func JoinObjectKey(prefix string, key string) (string, error) {
	normalizedKey, ok := NormalizeObjectKey(key)
	if !ok {
		return "", fmt.Errorf("invalid object key")
	}
	if strings.TrimSpace(prefix) == "" {
		return normalizedKey, nil
	}

	normalizedPrefix, ok := NormalizeObjectKey(prefix)
	if !ok {
		return "", fmt.Errorf("invalid object key prefix")
	}
	return path.Join(normalizedPrefix, normalizedKey), nil
}
