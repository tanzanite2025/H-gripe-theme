package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
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
		// 自定义端点（例如MinIO）
		s3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // MinIO需要使用路径样式
		})
	} else {
		s3Client = s3.NewFromConfig(awsConfig)
	}

	return &s3StorageImpl{
		config:   cfg,
		client:   s3Client,
		s3Config: &awsConfig,
	}, nil
}

// Upload 上传文件到S3
func (s *s3StorageImpl) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
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

	// 上传到S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(filename),
		Body:        src,
		ContentType: aws.String(contentType),
		// ACL: types.ObjectCannedACLPublicRead, // 如果需要公开访问
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s.GetURL(filename), nil
}

// UploadFromReader 从Reader上传到S3
func (s *s3StorageImpl) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	// 生成唯一文件名
	newFilename := s.generateFilename(filename)

	// 检测内容类型
	contentType := detectContentType(filename)

	// 上传到S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(newFilename),
		Body:        reader,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s.GetURL(newFilename), nil
}

// Delete 从S3删除文件
func (s *s3StorageImpl) Delete(ctx context.Context, url string) error {
	// 从URL提取文件key
	key := s.extractKeyFromURL(url)
	if key == "" {
		return fmt.Errorf("invalid URL: cannot extract key")
	}

	// 从S3删除对象
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
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

// generateFilename 生成唯一文件名
func (s *s3StorageImpl) generateFilename(originalFilename string) string {
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

// extractKeyFromURL 从URL提取S3 key
func (s *s3StorageImpl) extractKeyFromURL(url string) string {
	// 处理自定义域名
	if s.config.BaseURL != "" {
		return strings.TrimPrefix(url, s.config.BaseURL+"/")
	}

	// 处理标准S3 URL
	bucketPrefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.config.Bucket, s.config.Region)
	if strings.HasPrefix(url, bucketPrefix) {
		return strings.TrimPrefix(url, bucketPrefix)
	}

	// 处理自定义端点
	if s.config.Endpoint != "" {
		endpointPrefix := fmt.Sprintf("%s/%s/", s.config.Endpoint, s.config.Bucket)
		return strings.TrimPrefix(url, endpointPrefix)
	}

	return ""
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
		MaxKeys: aws.Int32(maxKeys),
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
	copySource := fmt.Sprintf("%s/%s", s.config.Bucket, sourceKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.config.Bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(destKey),
	})

	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	return nil
}
