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
	config *Config
	client *oss.Client
	bucket *oss.Bucket
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

	return &ossStorageImpl{
		config: cfg,
		client: client,
		bucket: bucket,
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
	return s.uploadWithPrefix(ctx, file, prefix, true, "")
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
	err = s.bucket.PutObject(filename, src, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload to OSS: %w", err)
	}

	return s.GetURL(filename), nil
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
	err = s.bucket.PutObject(newFilename, reader, options...)
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
	err = s.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("failed to delete from OSS: %w", err)
	}

	return nil
}

// GetURL 获取OSS文件URL
func (s *ossStorageImpl) GetURL(filename string) string {
	if s.config.BaseURL != "" {
		// 使用自定义域名或CDN
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), filename)
	}

	// 使用标准OSS URL
	// 格式：https://{bucket}.{endpoint}/{object}
	endpoint := strings.TrimPrefix(s.config.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", s.config.Bucket, endpoint, filename)
}

// GetPresignedURL 获取OSS预签名URL（用于临时访问私有文件）
func (s *ossStorageImpl) GetPresignedURL(ctx context.Context, filename string, duration time.Duration) (string, error) {
	// 生成预签名URL
	signedURL, err := s.bucket.SignURL(filename, oss.HTTPGet, int64(duration.Seconds()))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return signedURL, nil
}

// extractKeyFromURL 从URL提取OSS key
func (s *ossStorageImpl) extractKeyFromURL(url string) string {
	// 处理自定义域名
	if s.config.BaseURL != "" {
		prefix := strings.TrimRight(s.config.BaseURL, "/") + "/"
		if strings.HasPrefix(url, prefix) {
			return strings.TrimPrefix(url, prefix)
		}
		return ""
	}

	// 处理标准OSS URL
	endpoint := strings.TrimPrefix(s.config.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	ossPrefix := fmt.Sprintf("https://%s.%s/", s.config.Bucket, endpoint)

	if strings.HasPrefix(url, ossPrefix) {
		return strings.TrimPrefix(url, ossPrefix)
	}

	return ""
}

func (s *ossStorageImpl) ObjectKey(reference string) (string, error) {
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
		standardHost := fmt.Sprintf("%s.%s", s.config.Bucket, endpoint)
		if strings.EqualFold(parsed.Host, standardHost) {
			if normalized, ok := NormalizeObjectKey(strings.TrimPrefix(parsed.Path, "/")); ok {
				return normalized, nil
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

	headers, err := s.bucket.GetObjectMeta(normalizedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect OSS object: %w", err)
	}
	body, err := s.bucket.GetObject(normalizedKey)
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

	for {
		lsRes, err := s.bucket.ListObjects(
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

	_, err := s.bucket.CopyObject(normalizedSourceKey, normalizedDestKey)
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
	headers, err := s.bucket.GetObjectMeta(key)
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
	exists, err := s.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return exists, nil
}

// SetObjectACL 设置对象访问权限
func (s *ossStorageImpl) SetObjectACL(ctx context.Context, key string, acl oss.ACLType) error {
	err := s.bucket.SetObjectACL(key, acl)
	if err != nil {
		return fmt.Errorf("failed to set object ACL: %w", err)
	}

	return nil
}

// GetObjectToFile 下载对象到本地文件
func (s *ossStorageImpl) GetObjectToFile(ctx context.Context, key, filename string) error {
	err := s.bucket.GetObjectToFile(key, filename)
	if err != nil {
		return fmt.Errorf("failed to download object: %w", err)
	}

	return nil
}
