package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// The caller persists the allocated key before invoking this capability.
// Writes use private storage and no-store metadata without generating another key.
type PreallocatedPrivateObjectWriter interface {
	WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, objectKey string) (string, error)
}

type privateUploadReaderWithRequestContext struct {
	ctx    context.Context
	reader io.Reader
}

func (r privateUploadReaderWithRequestContext) Read(data []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(data)
}

func AllocatePrivateObjectKeyForPendingUpload(originalFilename, prefix string) (string, error) {
	if !isPrivateUploadPrefix(prefix) {
		return "", fmt.Errorf("private upload prefix must use a private namespace")
	}
	return generateObjectKey(originalFilename, prefix)
}

func validatePreallocatedPrivateObjectStorageKey(key string) (string, error) {
	normalized, ok := NormalizeObjectKey(key)
	if !ok || normalized != key || !IsPrivateObjectKey(normalized) {
		return "", fmt.Errorf("invalid preallocated private object key")
	}
	return normalized, nil
}

func (s *localStorage) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, objectKey string) (string, error) {
	key, err := validatePreallocatedPrivateObjectStorageKey(objectKey)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	private := s.privateStorage()
	destination, err := private.localPath(key)
	if err != nil {
		return "", err
	}
	if err := ensureLocalDirectory(private.config.LocalPath, filepath.Dir(destination)); err != nil {
		return "", err
	}
	// #nosec G304 -- private.localPath constrains destination to the private root.
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, localFilePerm)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(file, privateUploadReaderWithRequestContext{ctx: ctx, reader: reader})
	closeErr := file.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return private.GetURL(key), nil
}

func (s *s3StorageImpl) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, objectKey string) (string, error) {
	key, err := validatePreallocatedPrivateObjectStorageKey(objectKey)
	if err != nil {
		return "", err
	}
	if err := validatePrivateBucketUploadNamespaceAndConfiguration(s.config, strings.TrimSuffix(key, "/"+filepath.Base(key))); err != nil {
		return "", err
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader,
		ContentType: aws.String(detectContentType(key)), CacheControl: aws.String("private, no-store")})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WritePrivateObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, objectKey string) (string, error) {
	key, err := validatePreallocatedPrivateObjectStorageKey(objectKey)
	if err != nil {
		return "", err
	}
	if err := validatePrivateBucketUploadNamespaceAndConfiguration(s.config, strings.TrimSuffix(key, "/"+filepath.Base(key))); err != nil {
		return "", err
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("private storage bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.CacheControl("private, no-store"), oss.ObjectACL(oss.ACLPrivate), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
