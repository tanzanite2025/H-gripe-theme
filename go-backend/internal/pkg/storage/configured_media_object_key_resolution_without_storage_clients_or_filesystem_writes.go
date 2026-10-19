package storage

import "fmt"

// ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess uses the same URL
// parsers as live storage, without constructing clients, opening objects or
// creating local directories. It deliberately exposes no storage mutation API.
func ResolveConfiguredPublicMediaObjectKeyWithoutStorageAccess(reference string, config *Config, siteURL string) (string, error) {
	if config == nil {
		return "", fmt.Errorf("media storage identity configuration is required")
	}
	var provider StorageService
	switch config.Type {
	case StorageTypeLocal:
		provider = &localStorage{config: config}
	case StorageTypeS3:
		provider = &s3StorageImpl{config: config}
	case StorageTypeOSS:
		provider = &ossStorageImpl{config: config}
	default:
		return "", fmt.Errorf("unsupported media storage identity type")
	}
	return TrustedBusinessMediaReferenceObjectKey(reference, provider, siteURL)
}
