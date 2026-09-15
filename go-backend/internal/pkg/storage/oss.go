package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// ossStorageImpl 阿里云OSS存储完整实现
type ossStorageImpl struct {
	config        *Config
	client        *oss.Client
	bucket        *oss.Bucket
	privateBucket *oss.Bucket
}

// NewOSSStorage 创建OSS存储服务
func NewOSSStorage(cfg *Config) (StorageService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("OSS bucket is required")
	}

	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("OSS access key and secret are required")
	}

	// 构建OSS endpoint
	endpoint := cfg.Endpoint
	if endpoint == "" {
		// 默认endpoint格式：oss-{region}.aliyuncs.com
		if cfg.Region != "" {
			endpoint = fmt.Sprintf("https://oss-%s.aliyuncs.com", cfg.Region)
		} else {
			return nil, fmt.Errorf("OSS endpoint or region is required")
		}
	}

	// 创建OSS客户端
	client, err := oss.New(endpoint, cfg.AccessKeyID, cfg.SecretAccessKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取Bucket
	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get OSS bucket: %w", err)
	}

	var privateBucket *oss.Bucket
	if strings.TrimSpace(cfg.PrivateBucket) != "" && cfg.PrivateBucket != cfg.Bucket {
		privateBucket, err = client.Bucket(cfg.PrivateBucket)
		if err != nil {
			return nil, fmt.Errorf("failed to get private OSS bucket: %w", err)
		}
	}

	return &ossStorageImpl{
		config:        cfg,
		client:        client,
		bucket:        bucket,
		privateBucket: privateBucket,
	}, nil
}

// Upload 上传文件到OSS
func (s *ossStorageImpl) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return s.UploadWithPrefix(ctx, file, "")
}

func (s *ossStorageImpl) UploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	return s.uploadWithPrefix(ctx, file, prefix, false, "")
}

func (s *ossStorageImpl) UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	if !isPrivateUploadPrefix(prefix) {
		return "", fmt.Errorf("private upload prefix must use a private namespace")
	}
	if strings.TrimSpace(s.config.PrivateBucket) == "" || strings.TrimSpace(s.config.PrivateBucket) == strings.TrimSpace(s.config.Bucket) {
		return "", fmt.Errorf("private storage bucket is not configured")
	}
	if strings.TrimSpace(s.config.PrivateBaseURL) != "" && strings.EqualFold(strings.TrimRight(strings.TrimSpace(s.config.PrivateBaseURL), "/"), strings.TrimRight(strings.TrimSpace(s.config.BaseURL), "/")) {
		return "", fmt.Errorf("private storage base URL must be separate from public base URL")
	}
	return s.uploadWithPrefix(ctx, file, prefix, true, "")
}

func (s *ossStorageImpl) bucketForKey(key string) *oss.Bucket {
	if s != nil && IsPrivateObjectKey(key) {
		if s.privateBucket == nil {
			return nil
		}
		return s.privateBucket
	}
	return s.bucket
}

func (s *ossStorageImpl) UploadWithPrefixAndCacheControl(ctx context.Context, file *multipart.FileHeader, prefix string, cacheControl string) (string, error) {
	return s.uploadWithPrefix(ctx, file, prefix, false, cacheControl)
}

func (s *ossStorageImpl) uploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string, private bool, cacheControl string) (string, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = src.Close() }()

	// 生成唯一文件名
	filename, err := generateObjectKey(file.Filename, prefix)
	if err != nil {
		return "", err
	}

	// 检测内容类型
	contentType := detectContentType(file.Filename)

	// 上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	if private {
		options = append(options, oss.ObjectACL(oss.ACLPrivate))
	}
	if cacheControl = strings.TrimSpace(cacheControl); cacheControl != "" {
		options = append(options, oss.CacheControl(cacheControl))
	}

	// 上传到OSS
	bucket := s.bucketForKey(filename)
	if bucket == nil {
		return "", fmt.Errorf("private storage bucket is not configured")
	}
	err = bucket.PutObject(filename, src, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload to OSS: %w", err)
	}

	return s.getURL(filename, private), nil
}

// UploadFromReader 从Reader上传到OSS
func (s *ossStorageImpl) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return s.UploadFromReaderWithPrefix(ctx, reader, filename, "")
}

func (s *ossStorageImpl) UploadFromReaderWithPrefix(ctx context.Context, reader io.Reader, filename string, prefix string) (string, error) {
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, "")
}

func (s *ossStorageImpl) UploadFromReaderWithPrefixAndCacheControl(_ context.Context, reader io.Reader, filename string, prefix string, cacheControl string) (string, error) {
	// 生成唯一文件名
	newFilename, err := generateObjectKey(filename, prefix)
	if err != nil {
		return "", err
	}

	// 检测内容类型
	contentType := detectContentType(filename)

	// 上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
	}
	if cacheControl = strings.TrimSpace(cacheControl); cacheControl != "" {
		options = append(options, oss.CacheControl(cacheControl))
	}

	// 上传到OSS
	bucket := s.bucketForKey(newFilename)
	if bucket == nil {
		return "", fmt.Errorf("private storage bucket is not configured")
	}
	err = bucket.PutObject(newFilename, reader, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload to OSS: %w", err)
	}

	return s.GetURL(newFilename), nil
}

// Delete 从OSS删除文件
func (s *ossStorageImpl) Delete(ctx context.Context, url string) error {
	// 从URL提取对象key
	key, err := s.ObjectKey(url)
	if err != nil {
		return err
	}

	// 从OSS删除对象
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return fmt.Errorf("private storage bucket is not configured")
	}
	err = bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete from OSS: %w", err)
	}

	return nil
}

// GetURL 获取OSS文件URL
func (s *ossStorageImpl) GetURL(filename string) string {
	return s.getURL(filename, false)
}

func (s *ossStorageImpl) getURL(filename string, forcePrivate bool) string {
	if (forcePrivate || IsPrivateObjectKey(filename)) && s.privateBucket == nil {
		return ""
	}
	baseURL := s.config.BaseURL
	if forcePrivate || IsPrivateObjectKey(filename) {
		if strings.TrimSpace(s.config.PrivateBaseURL) != "" {
			baseURL = s.config.PrivateBaseURL
		} else {
			// Never expose private objects through the public CDN. Fall back to
			// the private bucket's native endpoint URL below.
			baseURL = ""
		}
	}
	if baseURL != "" {
		// 使用自定义域名或CDN
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), filename)
	}

	// 使用标准OSS URL
	// 格式：https://{bucket}.{endpoint}/{object}
	endpoint := strings.TrimPrefix(s.config.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	bucket := s.config.Bucket
	if (forcePrivate || IsPrivateObjectKey(filename)) && strings.TrimSpace(s.config.PrivateBucket) != "" {
		bucket = s.config.PrivateBucket
	}
	return fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, filename)
}

// GetPresignedURL 获取OSS预签名URL（用于临时访问私有文件）
func (s *ossStorageImpl) GetPresignedURL(ctx context.Context, filename string, duration time.Duration) (string, error) {
	// 生成预签名URL
	bucket := s.bucketForKey(filename)
	if bucket == nil {
		return "", fmt.Errorf("private storage bucket is not configured")
	}
	signedURL, err := bucket.SignURL(filename, oss.HTTPGet, int64(duration.Seconds()))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return signedURL, nil
}

// extractKeyFromURL 从URL提取OSS key
func (s *ossStorageImpl) extractKeyFromURL(url string) string {
	// 处理私有和公共自定义域名（两者可能同时配置）
	if s.config.PrivateBaseURL != "" {
		prefix := strings.TrimRight(s.config.PrivateBaseURL, "/") + "/"
		if strings.HasPrefix(url, prefix) {
			return strings.TrimPrefix(url, prefix)
		}
	}
	if s.config.BaseURL != "" {
		prefix := strings.TrimRight(s.config.BaseURL, "/") + "/"
		if strings.HasPrefix(url, prefix) {
			return strings.TrimPrefix(url, prefix)
		}
	}

	// 处理标准OSS URL
	endpoint := strings.TrimPrefix(s.config.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	for _, bucket := range []string{s.config.PrivateBucket, s.config.Bucket} {
		if strings.TrimSpace(bucket) == "" {
			continue
		}
		ossPrefix := fmt.Sprintf("https://%s.%s/", bucket, endpoint)
		if strings.HasPrefix(url, ossPrefix) {
			return strings.TrimPrefix(url, ossPrefix)
		}
	}

	return ""
}

func (s *ossStorageImpl) ObjectKey(reference string) (string, error) {
	if key, ok := ObjectKeyFromBaseURL(reference, s.config.PrivateBaseURL); ok {
		return key, nil
	}
	if key, ok := ObjectKeyFromBaseURL(reference, s.config.BaseURL); ok {
		return key, nil
	}
	if key := s.extractKeyFromURL(reference); key != "" {
		if normalized, ok := NormalizeObjectKey(key); ok {
			return normalized, nil
		}
	}

	value := strings.TrimSpace(reference)
	if parsed, err := url.Parse(value); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		endpoint := strings.TrimPrefix(s.config.Endpoint, "https://")
		endpoint = strings.TrimPrefix(endpoint, "http://")
		for _, bucket := range []string{s.config.PrivateBucket, s.config.Bucket} {
			if strings.TrimSpace(bucket) == "" {
				continue
			}
			standardHost := fmt.Sprintf("%s.%s", bucket, endpoint)
			if strings.EqualFold(parsed.Host, standardHost) {
				if normalized, ok := NormalizeObjectKey(strings.TrimPrefix(parsed.Path, "/")); ok {
					return normalized, nil
				}
			}
		}
	}

	if key, ok := ObjectKeyFromReference(reference, s.config.BaseURL); ok {
		return key, nil
	}
	return "", fmt.Errorf("invalid object key")
}

func (s *ossStorageImpl) Open(ctx context.Context, key string) (*StoredObject, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	normalizedKey, ok := NormalizeObjectKey(key)
	if !ok {
		return nil, fmt.Errorf("invalid object key")
	}

	bucket := s.bucketForKey(normalizedKey)
	if bucket == nil {
		return nil, fmt.Errorf("private storage bucket is not configured")
	}
	headers, err := bucket.GetObjectMeta(normalizedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect OSS object: %w", err)
	}
	body, err := bucket.GetObject(normalizedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to open OSS object: %w", err)
	}

	size, _ := strconv.ParseInt(strings.TrimSpace(headers.Get("Content-Length")), 10, 64)
	modTime, _ := http.ParseTime(headers.Get("Last-Modified"))
	return &StoredObject{
		ReadCloser: body,
		Name:       path.Base(normalizedKey),
		MimeType:   strings.TrimSpace(headers.Get("Content-Type")),
		Size:       size,
		ModTime:    modTime,
	}, nil
}

// ListObjects 列出OSS中的对象
func (s *ossStorageImpl) ListObjects(ctx context.Context, prefix string, maxKeys int) ([]string, error) {
	marker := ""
	keys := make([]string, 0)
	bucket := s.bucket
	if IsPrivateObjectKey(prefix) || isPrivateUploadPrefix(prefix) {
		bucket = s.bucketForKey(prefix)
		if bucket == nil {
			return nil, fmt.Errorf("private storage bucket is not configured")
		}
	}

	for {
		lsRes, err := bucket.ListObjects(
			oss.Prefix(prefix),
			oss.MaxKeys(maxKeys),
			oss.Marker(marker),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range lsRes.Objects {
			keys = append(keys, obj.Key)
		}

		if !lsRes.IsTruncated {
			break
		}

		marker = lsRes.NextMarker
	}

	return keys, nil
}

// CopyObject 复制OSS对象
func (s *ossStorageImpl) CopyObject(ctx context.Context, sourceKey, destKey string) error {
	normalizedSourceKey, ok := NormalizeObjectKey(sourceKey)
	if !ok {
		return fmt.Errorf("invalid source object key")
	}
	normalizedDestKey, ok := NormalizeObjectKey(destKey)
	if !ok {
		return fmt.Errorf("invalid destination object key")
	}

	sourceBucket := s.bucketForKey(normalizedSourceKey)
	destBucket := s.bucketForKey(normalizedDestKey)
	if sourceBucket == nil || destBucket == nil {
		return fmt.Errorf("private storage bucket is not configured")
	}
	var err error
	if sourceBucket == destBucket {
		_, err = sourceBucket.CopyObject(normalizedSourceKey, normalizedDestKey)
	} else if sourceBucket == s.privateBucket {
		_, err = destBucket.CopyObjectFrom(s.config.PrivateBucket, normalizedSourceKey, normalizedDestKey)
	} else {
		_, err = destBucket.CopyObjectFrom(s.config.Bucket, normalizedSourceKey, normalizedDestKey)
	}
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}

// UploadMultipart 分片上传大文件到OSS
func (s *ossStorageImpl) UploadMultipart(ctx context.Context, reader io.Reader, filename string, chunkSize int64) (string, error) {
	// 生成唯一文件名
	newFilename, err := generateObjectKey(filename, "")
	if err != nil {
		return "", err
	}

	// 检测内容类型
	contentType := detectContentType(filename)

	// 初始化分片上传
	chunks, err := oss.SplitFileByPartSize(filename, chunkSize)
	if err != nil {
		return "", fmt.Errorf("failed to split file: %w", err)
	}

	// 上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
		oss.Routines(3), // 并发数
	}

	// 执行分片上传
	err = s.bucket.UploadFile(newFilename, filename, chunkSize, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload multipart: %w", err)
	}

	_ = chunks // 使用chunks避免未使用变量警告

	return s.GetURL(newFilename), nil
}

// GetObjectMeta 获取对象元信息
func (s *ossStorageImpl) GetObjectMeta(ctx context.Context, key string) (map[string]string, error) {
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return nil, fmt.Errorf("private storage bucket is not configured")
	}
	headers, err := bucket.GetObjectMeta(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get object meta: %w", err)
	}

	meta := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			meta[k] = v[0]
		}
	}

	return meta, nil
}

// IsObjectExist 检查对象是否存在
func (s *ossStorageImpl) IsObjectExist(ctx context.Context, key string) (bool, error) {
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return false, fmt.Errorf("private storage bucket is not configured")
	}
	exists, err := bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return exists, nil
}

// SetObjectACL 设置对象访问权限
func (s *ossStorageImpl) SetObjectACL(ctx context.Context, key string, acl oss.ACLType) error {
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return fmt.Errorf("private storage bucket is not configured")
	}
	err := bucket.SetObjectACL(key, acl)
	if err != nil {
		return fmt.Errorf("failed to set object ACL: %w", err)
	}

	return nil
}

// GetObjectToFile 下载对象到本地文件
func (s *ossStorageImpl) GetObjectToFile(ctx context.Context, key, filename string) error {
	bucket := s.bucketForKey(key)
	if bucket == nil {
		return fmt.Errorf("private storage bucket is not configured")
	}
	err := bucket.GetObjectToFile(key, filename)
	if err != nil {
		return fmt.Errorf("failed to download object: %w", err)
	}

	return nil
}
