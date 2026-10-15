package storage

import (
	"fmt"
	"net/url"
	"strings"
)

// TrustedBusinessMediaReferenceObjectKey resolves identity without fetching the
// URL. An empty key means an external URL or an unmanaged static path. Only the
// configured provider, the configured site and root /uploads/ paths are trusted;
// a foreign host containing /uploads/ is never an alias of local storage.
func TrustedBusinessMediaReferenceObjectKey(reference string, provider StorageService, siteURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil || parsed.Path == "" || parsed.User != nil || strings.Contains(parsed.Path, "\\") || parsed.Opaque != "" {
		return "", fmt.Errorf("invalid business media URL")
	}
	if parsed.Host != "" || parsed.Scheme != "" {
		if parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return "", fmt.Errorf("business media URL must use HTTP or HTTPS")
		}
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("business media URL cannot contain directory traversal")
		}
	}
	parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
	if isConfiguredPrivateStorageOriginForBusinessMediaReference(parsed.String(), provider) {
		return "", fmt.Errorf("private storage addresses cannot be published as business media")
	}
	if parsed.Host == "" {
		if strings.HasPrefix(parsed.Path, "/uploads/") || strings.HasPrefix(parsed.Path, "uploads/") {
			key, valid := NormalizeObjectKey(strings.TrimPrefix(strings.TrimPrefix(parsed.Path, "/"), "uploads/"))
			if !valid {
				return "", fmt.Errorf("invalid managed media object key")
			}
			return key, nil
		}
		return "", nil
	}
	if provider != nil {
		if key, err := provider.ObjectKey(parsed.String()); err == nil {
			if normalized, valid := NormalizeObjectKey(key); valid {
				return normalized, nil
			}
		}
	}
	if site, err := url.Parse(strings.TrimSpace(siteURL)); err == nil && site.Host != "" && strings.EqualFold(site.Scheme, parsed.Scheme) {
		if key, valid := ObjectKeyFromReference(parsed.String(), siteURL); valid {
			return key, nil
		}
	}
	return "", nil
}

// ObjectKey intentionally supports both public and private storage elsewhere.
// Public business binding must also preserve bucket identity when two buckets
// happen to contain the same full key.
func isConfiguredPrivateStorageOriginForBusinessMediaReference(reference string, provider StorageService) bool {
	var config *Config
	var privateNativeBaseURLs []string
	switch store := provider.(type) {
	case *localStorage:
		config = store.config
	case *s3StorageImpl:
		config = store.config
		if config.PrivateBucket != "" && config.PrivateBucket != config.Bucket {
			privateNativeBaseURLs = append(privateNativeBaseURLs, fmt.Sprintf("https://%s.s3.%s.amazonaws.com", config.PrivateBucket, config.Region))
			if config.Endpoint != "" {
				privateNativeBaseURLs = append(privateNativeBaseURLs, strings.TrimRight(config.Endpoint, "/")+"/"+config.PrivateBucket)
			}
		}
	case *ossStorageImpl:
		config = store.config
		if config.PrivateBucket != "" && config.PrivateBucket != config.Bucket {
			endpoint := strings.TrimPrefix(strings.TrimPrefix(config.Endpoint, "https://"), "http://")
			privateNativeBaseURLs = append(privateNativeBaseURLs, "https://"+config.PrivateBucket+"."+endpoint)
		}
	}
	if config == nil {
		return false
	}
	if strings.TrimSpace(config.PrivateBaseURL) != "" && config.PrivateBaseURL != config.BaseURL {
		privateNativeBaseURLs = append(privateNativeBaseURLs, config.PrivateBaseURL)
	}
	for _, base := range privateNativeBaseURLs {
		// Both HTTP and HTTPS identify the same private container. Scheme
		// changes must not turn a private origin into an unmanaged external URL.
		for _, scheme := range []string{"https://", "http://"} {
			candidate := scheme + strings.TrimPrefix(strings.TrimPrefix(base, "https://"), "http://")
			if _, matches := ObjectKeyFromBaseURL(reference, candidate); matches {
				return true
			}
		}
	}
	return false
}
