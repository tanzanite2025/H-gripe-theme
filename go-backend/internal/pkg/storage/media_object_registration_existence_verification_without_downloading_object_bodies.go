package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ObjectExistenceVerifier checks the provider origin, including its private
// location, rather than a possibly stale public URL or CDN response.
type ObjectExistenceVerifier interface {
	VerifyStoredObjectExists(context.Context, string) error
}

func (s *localStorage) VerifyStoredObjectExists(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	target := s
	if IsPrivateObjectKey(key) {
		target = s.privateStorage()
	}
	filePath, err := target.localPath(key)
	if err != nil {
		return err
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("stored object is not a regular file")
	}
	return nil
}

func (s *s3StorageImpl) VerifyStoredObjectExists(ctx context.Context, key string) error {
	normalized, valid := NormalizeObjectKey(key)
	if !valid {
		return fmt.Errorf("invalid object key")
	}
	bucket, err := s.requireBucketForKey(normalized)
	if err != nil {
		return err
	}
	_, err = s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(normalized)})
	return err
}

func (s *ossStorageImpl) VerifyStoredObjectExists(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	normalized, valid := NormalizeObjectKey(key)
	if !valid {
		return fmt.Errorf("invalid object key")
	}
	bucket := s.bucketForKey(normalized)
	if bucket == nil {
		return fmt.Errorf("private storage bucket is not configured")
	}
	_, err := bucket.GetObjectMeta(normalized, oss.WithContext(ctx))
	return err
}
