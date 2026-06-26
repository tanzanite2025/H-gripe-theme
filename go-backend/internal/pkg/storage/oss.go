package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
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
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 生成唯一文件名
	filename := s.generateFilename(file.Filename)

	// 检测内容类型
	contentType := detectContentType(file.Filename)

	// 上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
		// oss.ObjectACL(oss.ACLPublicRead), // 如果需要公开访问
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
	// 生成唯一文件名
	newFilename := s.generateFilename(filename)

	// 检测内容类型
	contentType := detectContentType(filename)

	// 上传选项
	options := []oss.Option{
		oss.ContentType(contentType),
	}

	// 上传到OSS
	err := s.bucket.PutObject(newFilename, reader, options...)
	if err != nil {
		return "", fmt.Errorf("failed to upload to OSS: %w", err)
	}

	return s.GetURL(newFilename), nil
}

// Delete 从OSS删除文件
func (s *ossStorageImpl) Delete(ctx context.Context, url string) error {
	// 从URL提取对象key
	key := s.extractKeyFromURL(url)
	if key == "" {
		return fmt.Errorf("invalid URL: cannot extract key")
	}

	// 从OSS删除对象
	err := s.bucket.DeleteObject(key)
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

// generateFilename 生成唯一文件名
func (s *ossStorageImpl) generateFilename(originalFilename string) string {
	// 清理原始文件名
	cleanName := filepath.Base(originalFilename)
	ext := strings.ToLower(filepath.Ext(cleanName))

	// 生成UUID
	id := uuid.New().String()

	// 生成日期路径
	now := time.Now()
	datePath := now.Format("2006/01/02")

	// 组合文件名
	return filepath.ToSlash(filepath.Join(datePath, fmt.Sprintf("%s%s", id, ext)))
}

// extractKeyFromURL 从URL提取OSS key
func (s *ossStorageImpl) extractKeyFromURL(url string) string {
	// 处理自定义域名
	if s.config.BaseURL != "" {
		return strings.TrimPrefix(url, s.config.BaseURL+"/")
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
	sourceObject := fmt.Sprintf("%s/%s", s.config.Bucket, sourceKey)

	_, err := s.bucket.CopyObject(sourceObject, destKey)
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}

// UploadMultipart 分片上传大文件到OSS
func (s *ossStorageImpl) UploadMultipart(ctx context.Context, reader io.Reader, filename string, chunkSize int64) (string, error) {
	// 生成唯一文件名
	newFilename := s.generateFilename(filename)

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
