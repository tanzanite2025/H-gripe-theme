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
	"github.com/google/uuid"
)

const BusinessDraftMediaObjectPrefix = "business-draft-media"

// This capability only writes keys allocated inside one draft receipt namespace.
// The service must commit the receipt before invoking it.
type PreallocatedBusinessDraftMediaObjectWriter interface {
	WriteBusinessDraftMediaObjectAtPreallocatedKey(context.Context, io.Reader, string) (string, error)
}

func BusinessDraftMediaReceiptIDFromObjectKey(key string) (string, bool) {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return "", false
	}
	parts := strings.Split(key, "/")
	if len(parts) != 3 || parts[0] != BusinessDraftMediaObjectPrefix {
		return "", false
	}
	id, err := uuid.Parse(parts[1])
	if err != nil || id.String() != parts[1] || id == uuid.Nil {
		return "", false
	}
	// Generated object names contain no URL syntax, escaping or subdirectories.
	for _, ch := range parts[2] {
		if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') && ch != '-' && ch != '.' {
			return "", false
		}
	}
	if parts[2] == "" || parts[2] == "." || parts[2] == ".." {
		return "", false
	}
	return parts[1], true
}

func AllocateBusinessDraftMediaObjectKey(receiptID, extension string) (string, error) {
	key := BusinessDraftMediaObjectPrefix + "/" + receiptID + "/" + uuid.NewString() + strings.ToLower(extension)
	if _, valid := BusinessDraftMediaReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid business draft media object identity")
	}
	return key, nil
}

func (s *localStorage) WriteBusinessDraftMediaObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := BusinessDraftMediaReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid business draft media object key")
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

func (s *s3StorageImpl) WriteBusinessDraftMediaObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := BusinessDraftMediaReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid business draft media object key")
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

func (s *ossStorageImpl) WriteBusinessDraftMediaObjectAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, valid := BusinessDraftMediaReceiptIDFromObjectKey(key); !valid {
		return "", fmt.Errorf("invalid business draft media object key")
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("business draft media bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
