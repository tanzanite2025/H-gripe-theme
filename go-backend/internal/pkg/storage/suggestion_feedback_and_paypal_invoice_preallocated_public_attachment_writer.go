package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const SuggestionAndPayPalAttachmentUploadCacheControl = "private, no-store"

type PreallocatedSuggestionOrPayPalAttachmentWriter interface {
	WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(context.Context, io.Reader, string) (string, error)
}

func (s *localStorage) WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, _, _, valid := SuggestionOrPayPalAttachmentObjectIdentity(key); !valid {
		return "", fmt.Errorf("invalid suggestion or PayPal attachment upload key")
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

func (s *s3StorageImpl) WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, _, _, valid := SuggestionOrPayPalAttachmentObjectIdentity(key); !valid {
		return "", fmt.Errorf("invalid suggestion or PayPal attachment upload key")
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: reader,
		ContentType: aws.String(detectContentType(key)), CacheControl: aws.String(SuggestionAndPayPalAttachmentUploadCacheControl)})
	if err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}

func (s *ossStorageImpl) WriteSuggestionOrPayPalAttachmentAtPreallocatedKey(ctx context.Context, reader io.Reader, key string) (string, error) {
	if _, _, _, valid := SuggestionOrPayPalAttachmentObjectIdentity(key); !valid {
		return "", fmt.Errorf("invalid suggestion or PayPal attachment upload key")
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return "", fmt.Errorf("suggestion or PayPal attachment public bucket is not configured")
	}
	if err := bucket.PutObject(key, reader, oss.ContentType(detectContentType(key)), oss.CacheControl(SuggestionAndPayPalAttachmentUploadCacheControl), oss.WithContext(ctx)); err != nil {
		return "", err
	}
	return s.GetURL(key), nil
}
