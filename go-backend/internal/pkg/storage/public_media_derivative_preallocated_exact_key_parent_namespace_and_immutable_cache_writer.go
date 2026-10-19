package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// Public derivatives use exact immutable generation keys in their parent asset namespace.
// This writer cannot enter original, dedicated-owner or private directories.
type PreallocatedPublicMediaDerivativeObjectWriter interface {
	WritePublicMediaDerivativeObjectAtPreallocatedKey(context.Context, io.Reader, string) (string, error)
}

const PublicMediaDerivativeImmutableCacheControl = "public, max-age=31536000, immutable"

func PreallocatedPublicMediaDerivativeParentAssetID(key string) (uint, bool) {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return 0, false
	}
	parts := strings.Split(key, "/")
	if len(parts) != 6 || parts[0] != "media-derivatives" || parts[4] != "preallocated-generations" {
		return 0, false
	}
	parent, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil || parent == 0 || uint64(uint(parent)) != parent || strconv.FormatUint(parent, 10) != parts[1] {
		return 0, false
	}
	if len(parts[2]) < 2 || len(parts[2]) > 40 {
		return 0, false
	}
	for _, character := range parts[2] {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '_' || character == '-') {
			return 0, false
		}
	}
	version, err := strconv.Atoi(strings.TrimPrefix(parts[3], "v"))
	if err != nil || version <= 0 || "v"+strconv.Itoa(version) != parts[3] {
		return 0, false
	}
	extension := path.Ext(parts[5])
	if extension != ".jpg" && extension != ".png" {
		return 0, false
	}
	id, err := uuid.Parse(strings.TrimSuffix(parts[5], extension))
	if err != nil || id == uuid.Nil || id.String()+extension != parts[5] {
		return 0, false
	}
	return uint(parent), true
}
func IsPreallocatedPublicMediaDerivativeObjectKey(key string) bool {
	_, valid := PreallocatedPublicMediaDerivativeParentAssetID(key)
	return valid
}
func AllocatePublicMediaDerivativeObjectKey(parent uint, preset string, version int, extension string) (string, error) {
	key := fmt.Sprintf("media-derivatives/%d/%s/v%d/preallocated-generations/%s%s", parent, preset, version, uuid.NewString(), extension)
	if !IsPreallocatedPublicMediaDerivativeObjectKey(key) {
		return "", fmt.Errorf("invalid public media derivative identity")
	}
	return key, nil
}

func (s *localStorage) WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPublicMediaDerivativeObjectKey(key) {
		return "", fmt.Errorf("invalid public media derivative object key")
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

func (s *s3StorageImpl) WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPublicMediaDerivativeObjectKey(key) {
		return "", fmt.Errorf("invalid public media derivative object key")
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader, ContentType: aws.String(detectContentType(key)), CacheControl: aws.String(PublicMediaDerivativeImmutableCacheControl)})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WritePublicMediaDerivativeObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if !IsPreallocatedPublicMediaDerivativeObjectKey(key) {
		return "", fmt.Errorf("invalid public media derivative object key")
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("public media derivative bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.CacheControl(PublicMediaDerivativeImmutableCacheControl), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
