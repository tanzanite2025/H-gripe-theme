package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// PrivateCacheControlledReaderUploader guarantees private storage selection and
// object access policy as well as cache metadata. A generic reader uploader does
// not establish these guarantees and cannot substitute for this capability.
type PrivateCacheControlledReaderUploader interface {
	UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error)
}

func validatePrivateBucketUploadNamespaceAndConfiguration(cfg *Config, prefix string) error {
	if !isPrivateUploadPrefix(prefix) {
		return fmt.Errorf("private upload prefix must use a private namespace")
	}
	if cfg == nil || strings.TrimSpace(cfg.PrivateBucket) == "" ||
		strings.EqualFold(strings.TrimSpace(cfg.PrivateBucket), strings.TrimSpace(cfg.Bucket)) {
		return fmt.Errorf("private storage bucket is not configured")
	}
	if strings.TrimSpace(cfg.PrivateBaseURL) != "" &&
		strings.EqualFold(strings.TrimRight(strings.TrimSpace(cfg.PrivateBaseURL), "/"), strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")) {
		return fmt.Errorf("private storage base URL must be separate from public base URL")
	}
	return nil
}

func (s *localStorage) UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error) {
	if !isPrivateUploadPrefix(prefix) {
		return "", fmt.Errorf("private upload prefix must use a private namespace")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, privateUploadReaderWithRequestContext{ctx: ctx, reader: reader}, filename, prefix, cacheControl)
}

func (s *s3StorageImpl) UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error) {
	if err := validatePrivateBucketUploadNamespaceAndConfiguration(s.config, prefix); err != nil {
		return "", err
	}
	// Keep ACL headers absent for S3 bucket-owner-enforced configurations.
	// The separate private bucket must deny anonymous reads in its bucket policy.
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, cacheControl)
}

func (s *ossStorageImpl) UploadPrivateObjectFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename, prefix, cacheControl string) (string, error) {
	if err := validatePrivateBucketUploadNamespaceAndConfiguration(s.config, prefix); err != nil {
		return "", err
	}
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, cacheControl)
}
