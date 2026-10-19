package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Permanent uploads retain their trusted feature directory and never inherit draft expiry.
// This restricted writer cannot enter private or dedicated-owner directories.
type PreallocatedPermanentMediaOriginalObjectWriter interface {
	WritePermanentMediaOriginalObjectAtPreallocatedKey(context.Context, io.Reader, string) (string, error)
}

func IsPermanentMediaOriginalUploadPrefix(prefix string) bool {
	switch prefix {
	case "", "catalog/product-images", "catalog/product-videos", "catalog/product-detail-media", "catalog/product-variant-swatches", "content/page-media", "shared/reusable-brand-and-instruction-media", "customer-service/reusable-reply-media", "customer-service/public-conversation-attachments", "workbench-feed":
		return true
	}
	return false
}
func IsPreallocatedPermanentMediaOriginalObjectKey(key string) bool {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return false
	}
	directory := path.Dir(key)
	const segment = "preallocated-permanent-originals"
	if directory != segment && !strings.HasSuffix(directory, "/"+segment) {
		return false
	}
	prefix := strings.TrimSuffix(strings.TrimSuffix(directory, segment), "/")
	if !IsPermanentMediaOriginalUploadPrefix(prefix) {
		return false
	}
	filename := path.Base(key)
	extension := path.Ext(filename)
	switch extension {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".mp4", ".webm", ".mov":
	default:
		return false
	}
	id, err := uuid.Parse(strings.TrimSuffix(filename, extension))
	return err == nil && id != uuid.Nil && id.String()+extension == filename
}
func AllocatePermanentMediaOriginalObjectKey(prefix, extension string) (string, error) {
	if !IsPermanentMediaOriginalUploadPrefix(prefix) {
		return "", fmt.Errorf("unsupported permanent media original prefix")
	}
	key := path.Join(prefix, "preallocated-permanent-originals", uuid.NewString()+strings.ToLower(extension))
	if !IsPreallocatedPermanentMediaOriginalObjectKey(key) {
		return "", fmt.Errorf("invalid permanent media original identity")
	}
	return key, nil
}

func (s *localStorage) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPermanentMediaOriginalObjectKey(key) {
		return "", fmt.Errorf("invalid permanent media original object key")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	destination, err := s.localPath(key)
	if err != nil {
		return "", err
	}
	if err := ensureLocalDirectory(s.config.LocalPath, filepath.Dir(destination)); err != nil {
		return "", err
	}
	// #nosec G304 -- localPath confines the destination to the configured root.
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, localFilePerm)
	if err != nil {
		return "", err
	}
	_, writeErr := io.Copy(file, privateUploadReaderWithRequestContext{ctx: ctx, reader: reader})
	closeErr := file.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if closeErr != nil {
		return "", closeErr
	}
	return s.GetURL(key), nil
}

func (s *s3StorageImpl) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPermanentMediaOriginalObjectKey(key) {
		return "", fmt.Errorf("invalid permanent media original object key")
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader, ContentType: aws.String(detectContentType(key))})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WritePermanentMediaOriginalObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPermanentMediaOriginalObjectKey(key) {
		return "", fmt.Errorf("invalid permanent media original object key")
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("permanent media original bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
