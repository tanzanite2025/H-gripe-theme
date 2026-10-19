package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (s *s3StorageImpl) OpenObjectByteRangeSource(ctx context.Context, key string) (ObjectByteRangeSource, error) {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	bucket, err := s.requireBucketForKey(key)
	if err != nil {
		return nil, err
	}
	head, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return nil, fmt.Errorf("inspect remote video metadata: %w", err)
	}
	metadata := ObjectByteRangeMetadata{Size: aws.ToInt64(head.ContentLength), ETag: aws.ToString(head.ETag)}
	if head.LastModified != nil {
		metadata.ModTime = *head.LastModified
	}
	if head.ContentLength == nil {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	if err := validatePinnedRemoteObjectByteRangeMetadata(metadata, aws.ToString(head.ContentEncoding)); err != nil {
		return nil, err
	}
	return &remoteObjectByteRangeSource{ctx: ctx, metadata: metadata, openRange: func(readContext context.Context, start, length int64) (io.ReadCloser, error) {
		object, err := s.client.GetObject(readContext, &s3.GetObjectInput{
			Bucket: aws.String(bucket), Key: aws.String(key), IfMatch: aws.String(metadata.ETag),
			Range: aws.String(fmt.Sprintf("bytes=%d-%d", start, start+length-1)),
		})
		if err != nil {
			return nil, fmt.Errorf("read pinned remote video range: %w", err)
		}
		if object.ContentLength == nil {
			if object.Body != nil {
				_ = object.Body.Close()
			}
			return nil, ErrRemoteObjectByteRangeUnavailable
		}
		return validateRemoteObjectByteRangeResponseBeforeExposingBody(object.Body, metadata, start, length, *object.ContentLength,
			aws.ToString(object.ContentRange), aws.ToString(object.ETag), aws.ToString(object.ContentEncoding))
	}}, nil
}

func (s *ossStorageImpl) OpenObjectByteRangeSource(ctx context.Context, key string) (ObjectByteRangeSource, error) {
	normalized, valid := NormalizeObjectKey(key)
	if !valid || normalized != key {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	head, err := bucket.GetObjectDetailedMeta(key, oss.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("inspect remote video metadata: %w", err)
	}
	size, err := strconv.ParseInt(strings.TrimSpace(head.Get("Content-Length")), 10, 64)
	if err != nil {
		return nil, ErrRemoteObjectByteRangeUnavailable
	}
	modTime, _ := time.Parse(http.TimeFormat, head.Get("Last-Modified"))
	metadata := ObjectByteRangeMetadata{Size: size, ModTime: modTime, ETag: head.Get("ETag")}
	if err := validatePinnedRemoteObjectByteRangeMetadata(metadata, head.Get("Content-Encoding")); err != nil {
		return nil, err
	}
	return &remoteObjectByteRangeSource{ctx: ctx, metadata: metadata, openRange: func(readContext context.Context, start, length int64) (io.ReadCloser, error) {
		var responseHeaders http.Header
		body, err := bucket.GetObject(key, oss.Range(start, start+length-1), oss.IfMatch(metadata.ETag), oss.WithContext(readContext), oss.GetResponseHeader(&responseHeaders))
		if err != nil {
			return nil, fmt.Errorf("read pinned remote video range: %w", err)
		}
		responseLength, lengthErr := strconv.ParseInt(strings.TrimSpace(responseHeaders.Get("Content-Length")), 10, 64)
		if lengthErr != nil {
			_ = body.Close()
			return nil, ErrRemoteObjectByteRangeUnavailable
		}
		return validateRemoteObjectByteRangeResponseBeforeExposingBody(body, metadata, start, length, responseLength,
			responseHeaders.Get("Content-Range"), responseHeaders.Get("ETag"), responseHeaders.Get("Content-Encoding"))
	}}, nil
}
