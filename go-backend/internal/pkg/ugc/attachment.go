package ugc

import (
	"errors"
	"net/url"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultAttachmentReferenceMaxBytes = 2048

var (
	ErrAttachmentTooMany     = errors.New("too many attachments")
	ErrAttachmentTooLong     = errors.New("attachment reference is too long")
	ErrAttachmentInvalidURL  = errors.New("attachment reference must point to a public upload")
	ErrAttachmentInvalidType = errors.New("attachment type is not allowed")
)

var DefaultImageAttachmentExtensions = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

type AttachmentReference struct {
	RawValue   string
	Value      string
	StorageKey string
	SourceHost string
}

func NormalizeUploadAttachmentReferences(values []string, maxAttachments int, allowedExtensions []string) ([]AttachmentReference, error) {
	refs := make([]AttachmentReference, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}

		key, sourceHost, err := uploadStorageKeyFromReference(value, allowedExtensions)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[key]; exists {
			continue
		}
		if maxAttachments > 0 && len(refs)+1 > maxAttachments {
			return nil, ErrAttachmentTooMany
		}

		seen[key] = struct{}{}
		refs = append(refs, AttachmentReference{
			RawValue:   value,
			Value:      "/uploads/" + key,
			StorageKey: key,
			SourceHost: sourceHost,
		})
	}

	return refs, nil
}

func NormalizeUploadImageAttachmentReferences(values []string, maxAttachments int) ([]AttachmentReference, error) {
	return NormalizeUploadAttachmentReferences(values, maxAttachments, DefaultImageAttachmentExtensions)
}

func uploadStorageKeyFromReference(value string, allowedExtensions []string) (string, string, error) {
	if len(value) > DefaultAttachmentReferenceMaxBytes {
		return "", "", ErrAttachmentTooLong
	}
	if !utf8.ValidString(value) || strings.Contains(value, "\x00") || strings.Contains(value, "\\") || strings.ContainsFunc(value, unicode.IsSpace) {
		return "", "", ErrAttachmentInvalidURL
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", "", ErrAttachmentInvalidURL
	}

	if parsed.Scheme != "" {
		scheme := strings.ToLower(parsed.Scheme)
		if scheme != "http" && scheme != "https" {
			return "", "", ErrAttachmentInvalidURL
		}
		if parsed.Host == "" || parsed.User != nil {
			return "", "", ErrAttachmentInvalidURL
		}
		key, err := uploadStorageKeyFromPath(parsed.EscapedPath(), allowedExtensions)
		return key, strings.ToLower(parsed.Host), err
	}
	if parsed.Host != "" || strings.HasPrefix(value, "//") {
		return "", "", ErrAttachmentInvalidURL
	}
	if strings.Contains(value, ":") {
		return "", "", ErrAttachmentInvalidURL
	}

	pathValue := value
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		pathValue = parsed.EscapedPath()
	}
	key, err := uploadStorageKeyFromPath(pathValue, allowedExtensions)
	return key, "", err
}

func uploadStorageKeyFromPath(pathValue string, allowedExtensions []string) (string, error) {
	unescaped, err := url.PathUnescape(pathValue)
	if err != nil {
		return "", ErrAttachmentInvalidURL
	}
	unescaped = strings.TrimSpace(unescaped)
	if unescaped == "" || strings.Contains(unescaped, "\x00") || strings.Contains(unescaped, "\\") || strings.ContainsFunc(unescaped, unicode.IsSpace) {
		return "", ErrAttachmentInvalidURL
	}

	candidate := strings.TrimLeft(unescaped, "/")
	if strings.HasPrefix(candidate, "uploads/") {
		candidate = strings.TrimPrefix(candidate, "uploads/")
	} else if strings.HasPrefix(unescaped, "/") {
		return "", ErrAttachmentInvalidURL
	}

	return cleanUploadStorageKey(candidate, allowedExtensions)
}

func cleanUploadStorageKey(key string, allowedExtensions []string) (string, error) {
	key = strings.Trim(key, "/")
	if key == "" {
		return "", ErrAttachmentInvalidURL
	}

	for _, segment := range strings.Split(key, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrAttachmentInvalidURL
		}
	}

	clean := path.Clean(key)
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." {
		return "", ErrAttachmentInvalidURL
	}

	if !allowedAttachmentExtension(path.Ext(clean), allowedExtensions) {
		return "", ErrAttachmentInvalidType
	}
	return clean, nil
}

func allowedAttachmentExtension(ext string, allowedExtensions []string) bool {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" || len(allowedExtensions) == 0 {
		return false
	}
	for _, allowed := range allowedExtensions {
		if ext == allowed {
			return true
		}
	}
	return false
}
