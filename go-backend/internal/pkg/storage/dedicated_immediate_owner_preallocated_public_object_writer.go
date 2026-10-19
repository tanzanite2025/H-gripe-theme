package storage

import (
	"context"
	"errors"
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

// This capability cannot write shared media, private attachments or showcase
// drafts. Its caller must commit an exact-key cleanup receipt before any IO.
type PreallocatedDedicatedImmediateOwnerObjectWriter interface {
	WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(context.Context, io.Reader, string, string) (string, error)
}

const DedicatedImmediateOwnerUploadCacheControl = "public, max-age=31536000, immutable"

func AllocateDedicatedImmediateOwnerUploadObjectKey(namespace, filename string) (string, error) {
	extension := strings.ToLower(path.Ext(filename))
	key := namespace + "/preallocated-uploads/" + uuid.NewString() + extension
	if !IsPreallocatedDedicatedImmediateOwnerObjectKey(key) {
		return "", fmt.Errorf("invalid dedicated immediate owner upload identity")
	}
	return key, nil
}

func IsPreallocatedDedicatedImmediateOwnerObjectKey(key string) bool {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return false
	}
	namespace := DedicatedPublicObjectNamespaceForStorageKey(key)
	if namespace != "site-logo" && namespace != "site-favicon" && namespace != "customer-service/avatars" {
		return false
	}
	name := strings.TrimPrefix(key, namespace+"/preallocated-uploads/")
	if name == key || strings.Contains(name, "/") {
		return false
	}
	extension := path.Ext(name)
	switch extension {
	case ".webp", ".png", ".jpg", ".jpeg", ".ico":
	default:
		return false
	}
	idText := strings.TrimSuffix(name, extension)
	id, err := uuid.Parse(idText)
	return err == nil && id != uuid.Nil && id.String() == idText
}

func validateDedicatedImmediateOwnerPreallocatedWrite(key, cacheControl string) error {
	if !IsPreallocatedDedicatedImmediateOwnerObjectKey(key) || cacheControl != DedicatedImmediateOwnerUploadCacheControl {
		return fmt.Errorf("invalid dedicated immediate owner upload key or cache policy")
	}
	return nil
}

func (s *localStorage) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	if err := validateDedicatedImmediateOwnerPreallocatedWrite(key, cacheControl); err != nil {
		return "", err
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
	if writeErr != nil || closeErr != nil {
		return "", fmt.Errorf("write dedicated public object: %w", errors.Join(writeErr, closeErr))
	}
	return s.GetURL(key), nil
}

func (s *s3StorageImpl) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	if err := validateDedicatedImmediateOwnerPreallocatedWrite(key, cacheControl); err != nil {
		return "", err
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader,
		ContentType: aws.String(detectContentType(key)), CacheControl: aws.String(cacheControl)})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key, cacheControl string) (string, error) {
	if err := validateDedicatedImmediateOwnerPreallocatedWrite(key, cacheControl); err != nil {
		return "", err
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("dedicated public media bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.CacheControl(cacheControl), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
