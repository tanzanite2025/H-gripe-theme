package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead preserves the
// private container identity that ObjectKey alone deliberately does not retain.
// The caller must first authorize and select a persisted business attachment.
func ResolvePrivateBusinessAttachmentObjectKeyForAuthorizedRead(provider StorageService, reference string) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("private attachment provider unavailable")
	}
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || parsed.Path == "" || parsed.User != nil || parsed.Opaque != "" || strings.Contains(parsed.Path, "\\") ||
		(parsed.Scheme != "" && parsed.Scheme != "http" && parsed.Scheme != "https") || (parsed.Host != "" && parsed.Scheme == "") {
		return "", fmt.Errorf("invalid private attachment reference")
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid private attachment path")
		}
	}
	var key string
	if parsed.Host == "" {
		key = strings.TrimPrefix(strings.TrimPrefix(parsed.Path, "/uploads/"), "uploads/")
	} else {
		key, err = provider.ObjectKey(reference)
	}
	normalized, valid := NormalizeObjectKey(key)
	if err != nil || !valid || normalized != key || !IsPrivateObjectKey(key) {
		return "", fmt.Errorf("invalid private attachment identity")
	}
	if parsed.Host == "" {
		if parsed.Path != key && parsed.Path != "/uploads/"+key && parsed.Path != "uploads/"+key {
			return "", fmt.Errorf("invalid private attachment relative path")
		}
		return key, nil
	}
	var bases []string
	switch store := provider.(type) {
	case *localStorage:
		// Historical local uploads use a non-public /uploads/ locator even
		// though the bytes reside in the separate private filesystem root.
		bases = append(bases, strings.TrimRight(store.config.BaseURL, "/")+"/uploads", strings.TrimRight(store.privateStorage().config.BaseURL, "/")+"/uploads")
	case *s3StorageImpl:
		cfg := store.config
		if cfg.PrivateBucket == "" || cfg.PrivateBucket == cfg.Bucket {
			return "", fmt.Errorf("private container unavailable")
		}
		bases = append(bases, cfg.PrivateBaseURL, fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.PrivateBucket, cfg.Region))
		if cfg.Endpoint != "" {
			bases = append(bases, strings.TrimRight(cfg.Endpoint, "/")+"/"+cfg.PrivateBucket)
		}
	case *ossStorageImpl:
		cfg := store.config
		if cfg.PrivateBucket == "" || cfg.PrivateBucket == cfg.Bucket {
			return "", fmt.Errorf("private container unavailable")
		}
		bases = append(bases, cfg.PrivateBaseURL, "https://"+cfg.PrivateBucket+"."+strings.TrimPrefix(strings.TrimPrefix(cfg.Endpoint, "https://"), "http://"))
	default:
		// Adapter providers must retain their own exact canonical URL identity.
		canonical, parseErr := url.Parse(provider.GetURL(key))
		if parseErr == nil && canonical.Scheme == parsed.Scheme && strings.EqualFold(canonical.Host, parsed.Host) && canonical.Path == parsed.Path {
			return key, nil
		}
	}
	for _, base := range bases {
		if candidate, ok := ObjectKeyFromBaseURL(reference, base); ok && candidate == key {
			return key, nil
		}
	}
	return "", fmt.Errorf("private attachment belongs to a different storage container")
}
