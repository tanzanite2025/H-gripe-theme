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

const VisualShowcaseDraftUploadCacheControl = "public, max-age=31536000, immutable"

type PreallocatedVisualShowcaseDraftObjectWriter interface {
	WriteVisualShowcaseDraftObjectAtPreallocatedKey(context.Context, io.Reader, string) (string, error)
}

func AllocateVisualShowcaseDraftUploadObjectKey(prefix, filename string) (string, error) {
	key := prefix + "/upload-drafts/" + uuid.NewString() + strings.ToLower(path.Ext(filename))
	if _, valid := VisualShowcaseDraftUploadReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid visual showcase draft upload identity")
	}
	return key, nil
}

func VisualShowcaseDraftUploadReceiptIDFromObjectKey(key string) (string, bool) {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return "", false
	}
	parts := strings.Split(key, "/")
	if len(parts) != 5 || parts[0] != "visual-showcase" || parts[3] != "upload-drafts" {
		return "", false
	}
	for _, segment := range parts[1:3] {
		if segment == "" {
			return "", false
		}
		for _, ch := range segment {
			if (ch < 'a' || ch > 'z') && (ch < '0' || ch > '9') && ch != '-' && ch != '_' {
				return "", false
			}
		}
	}
	extension := path.Ext(parts[4])
	switch extension {
	case ".png", ".jpg", ".jpeg", ".webp":
	default:
		return "", false
	}
	idText := strings.TrimSuffix(parts[4], extension)
	id, err := uuid.Parse(idText)
	if err != nil || id == uuid.Nil || id.String() != idText {
		return "", false
	}
	return idText, true
}

func (s *localStorage) WriteVisualShowcaseDraftObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := VisualShowcaseDraftUploadReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid visual showcase draft upload key")
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
	// #nosec G304 -- localPath confines this path to the configured public root.
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, localFilePerm)
	if err != nil {
		return "", err
	}
	_, writeErr := io.Copy(file, privateUploadReaderWithRequestContext{ctx: ctx, reader: reader})
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *s3StorageImpl) WriteVisualShowcaseDraftObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := VisualShowcaseDraftUploadReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid visual showcase draft upload key")
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader,
		ContentType: aws.String(detectContentType(key)), CacheControl: aws.String(VisualShowcaseDraftUploadCacheControl)})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WriteVisualShowcaseDraftObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := VisualShowcaseDraftUploadReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid visual showcase draft upload key")
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("visual showcase public bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.CacheControl(VisualShowcaseDraftUploadCacheControl), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
