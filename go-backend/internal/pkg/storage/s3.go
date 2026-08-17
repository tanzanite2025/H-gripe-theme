package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3StorageImpl AWS S3 存储完整实现
type s3StorageImpl struct {
	config   *Config
	client   *s3.Client
	s3Config *aws.Config
}

// NewS3Storage 创建S3存储服务
func NewS3Storage(cfg *Config) (StorageService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}

	if cfg.Region == "" {
		return nil, fmt.Errorf("S3 region is required")
	}

	ctx := context.Background()

	// 创建AWS配置
	var awsConfig aws.Config
	var err error

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// 使用静态凭证
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		)
	} else {
		// 使用默认凭证链（环境变量、IAM角色等）
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建S3客户端
	var s3Client *s3.Client
	if cfg.Endpoint != "" {
		s3Client = s3.NewFromConfig(awsConfig, configureS3ClientOptions(cfg))
	} else {
		s3Client = s3.NewFromConfig(awsConfig)
	}

	return &s3StorageImpl{
		config:   cfg,
		client:   s3Client,
		s3Config: &awsConfig,
	}, nil
}

func configureS3ClientOptions(cfg *Config) func(*s3.Options) {
	return func(options *s3.Options) {
		if cfg == nil || strings.TrimSpace(cfg.Endpoint) == "" {
			return
		}

		// This SDK version uses the legacy EndpointResolver interface. Setting
		// only UsePathStyle would still send requests to AWS's default endpoint.
		options.EndpointResolver = s3.EndpointResolverFromURL(strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/"))
		options.UsePathStyle = true
	}
}

// Upload 上传文件到S3
func (s *s3StorageImpl) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return s.UploadWithPrefix(ctx, file, "")
}

func (s *s3StorageImpl) UploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	return s.uploadWithPrefix(ctx, file, prefix, "")
}

func (s *s3StorageImpl) UploadWithPrefixAndCacheControl(ctx context.Context, file *multipart.FileHeader, prefix string, cacheControl string) (string, error) {
	return s.uploadWithPrefix(ctx, file, prefix, cacheControl)
}

func (s *s3StorageImpl) uploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string, cacheControl string) (string, error) {
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

	// 上传到S3
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(filename),
		Body:        src,
		ContentType: aws.String(contentType),
		// ACL: types.ObjectCannedACLPublicRead, // 如果需要公开访问
	}
	if cacheControl = strings.TrimSpace(cacheControl); cacheControl != "" {
		input.CacheControl = aws.String(cacheControl)
	}
	_, err = s.client.PutObject(ctx, input)

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s.GetURL(filename), nil
}

func (s *s3StorageImpl) UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	// Do not send an ACL here: S3 buckets using Bucket Owner Enforced reject
	// ACL headers. Private-bucket policy is the durable access boundary.
	return s.UploadWithPrefix(ctx, file, prefix)
}

// UploadFromReader 从Reader上传到S3
func (s *s3StorageImpl) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return s.UploadFromReaderWithPrefix(ctx, reader, filename, "")
}

func (s *s3StorageImpl) UploadFromReaderWithPrefix(ctx context.Context, reader io.Reader, filename string, prefix string) (string, error) {
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, "")
}

func (s *s3StorageImpl) UploadFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename string, prefix string, cacheControl string) (string, error) {
	// 生成唯一文件名
	newFilename, err := generateObjectKey(filename, prefix)
	if err != nil {
		return "", err
	}

	// 检测内容类型
	contentType := detectContentType(filename)

	// 上传到S3
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(newFilename),
		Body:        reader,
		ContentType: aws.String(contentType),
	}
	if cacheControl = strings.TrimSpace(cacheControl); cacheControl != "" {
		input.CacheControl = aws.String(cacheControl)
	}
	_, err = s.client.PutObject(ctx, input)

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s.GetURL(newFilename), nil
}

// Delete 从S3删除文件
func (s *s3StorageImpl) Delete(ctx context.Context, url string) error {
	// 从URL提取文件key
	key, err := s.ObjectKey(url)
	if err != nil {
		return err
	}

	// 从S3删除对象
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// GetURL 获取S3文件URL
func (s *s3StorageImpl) GetURL(filename string) string {
	if s.config.BaseURL != "" {
		// 使用自定义域名或CDN
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), filename)
	}

	if s.config.Endpoint != "" {
		// 使用自定义端点（MinIO等）
		return fmt.Sprintf("%s/%s/%s", s.config.Endpoint, s.config.Bucket, filename)
	}

	// 使用标准S3 URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
		s.config.Bucket, s.config.Region, filename)
}

// GetPresignedURL 获取预签名URL（用于临时访问私有文件）
func (s *s3StorageImpl) GetPresignedURL(ctx context.Context, filename string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(filename),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// extractKeyFromURL 从URL提取S3 key
func (s *s3StorageImpl) extractKeyFromURL(url string) string {
	// 处理自定义域名
	if s.config.BaseURL != "" {
		prefix := strings.TrimRight(s.config.BaseURL, "/") + "/"
		if strings.HasPrefix(url, prefix) {
			return strings.TrimPrefix(url, prefix)
		}
		return ""
	}

	// 处理标准S3 URL
	bucketPrefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.config.Bucket, s.config.Region)
	if strings.HasPrefix(url, bucketPrefix) {
		return strings.TrimPrefix(url, bucketPrefix)
	}

	// 处理自定义端点
	if s.config.Endpoint != "" {
		endpointPrefix := fmt.Sprintf("%s/%s/", s.config.Endpoint, s.config.Bucket)
		if strings.HasPrefix(url, endpointPrefix) {
			return strings.TrimPrefix(url, endpointPrefix)
		}
	}

	return ""
}

func (s *s3StorageImpl) ObjectKey(reference string) (string, error) {
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
		standardHost := fmt.Sprintf("%s.s3.%s.amazonaws.com", s.config.Bucket, s.config.Region)
		if strings.EqualFold(parsed.Host, standardHost) {
			if normalized, ok := NormalizeObjectKey(strings.TrimPrefix(parsed.Path, "/")); ok {
				return normalized, nil
			}
		}
		if s.config.Endpoint != "" {
			endpoint, endpointErr := url.Parse(strings.TrimRight(s.config.Endpoint, "/"))
			if endpointErr == nil && strings.EqualFold(parsed.Host, endpoint.Host) {
				expectedPrefix := "/" + strings.Trim(s.config.Bucket, "/") + "/"
				if strings.HasPrefix(parsed.Path, expectedPrefix) {
					if normalized, ok := NormalizeObjectKey(strings.TrimPrefix(parsed.Path, expectedPrefix)); ok {
						return normalized, nil
					}
				}
			}
		}
	}

	if key, ok := ObjectKeyFromReference(reference, s.config.BaseURL); ok {
		return key, nil
	}
	return "", fmt.Errorf("invalid object key")
}

func (s *s3StorageImpl) Open(ctx context.Context, key string) (*StoredObject, error) {
	normalizedKey, ok := NormalizeObjectKey(key)
	if !ok {
		return nil, fmt.Errorf("invalid object key")
	}

	object, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(normalizedKey),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open S3 object: %w", err)
	}
	if object.Body == nil {
		return nil, fmt.Errorf("failed to open S3 object: empty response body")
	}

	modTime := time.Time{}
	if object.LastModified != nil {
		modTime = *object.LastModified
	}
	return &StoredObject{
		ReadCloser: object.Body,
		Name:       filepath.Base(normalizedKey),
		MimeType:   aws.ToString(object.ContentType),
		Size:       object.ContentLength,
		ModTime:    modTime,
	}, nil
}

// detectContentType 检测文件内容类型
func detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		// 图片
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",

		// 视频
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",

		// 音频
		".mp3": "audio/mpeg",
		".wav": "audio/wav",
		".ogg": "audio/ogg",

		// 文档
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

		// 压缩包
		".zip": "application/zip",
		".rar": "application/x-rar-compressed",
		".7z":  "application/x-7z-compressed",
		".tar": "application/x-tar",
		".gz":  "application/gzip",

		// 其他
		".json": "application/json",
		".xml":  "application/xml",
		".txt":  "text/plain",
		".css":  "text/css",
		".js":   "application/javascript",
		".html": "text/html",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}

	return "application/octet-stream"
}

// ListObjects 列出S3中的对象
func (s *s3StorageImpl) ListObjects(ctx context.Context, prefix string, maxKeys int32) ([]string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.config.Bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: maxKeys,
	}

	result, err := s.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	keys := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		if obj.Key != nil {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}

// CopyObject 复制S3对象
func (s *s3StorageImpl) CopyObject(ctx context.Context, sourceKey, destKey string) error {
	normalizedSourceKey, ok := NormalizeObjectKey(sourceKey)
	if !ok {
		return fmt.Errorf("invalid source object key")
	}
	normalizedDestKey, ok := NormalizeObjectKey(destKey)
	if !ok {
		return fmt.Errorf("invalid destination object key")
	}
	copySource := fmt.Sprintf("%s/%s", s.config.Bucket, normalizedSourceKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(normalizedDestKey),
	})

	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}
