package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StorageService 存储服务接口
type StorageService interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (string, error)
	UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error)
	Delete(ctx context.Context, url string) error
	GetURL(filename string) string
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
	StorageTypeOSS   StorageType = "oss"
)

// Config 存储配置
type Config struct {
	Type      StorageType
	LocalPath string
	BaseURL   string
	// S3/OSS 配置
	Bucket          string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
}

// localStorage 本地存储实现
type localStorage struct {
	config *Config
}

// NewStorageService 创建存储服务
func NewStorageService(config *Config) (StorageService, error) {
	switch config.Type {
	case StorageTypeLocal:
		return newLocalStorage(config)
	case StorageTypeS3:
		return newS3Storage(config)
	case StorageTypeOSS:
		return newOSSStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// newLocalStorage 创建本地存储
func newLocalStorage(config *Config) (StorageService, error) {
	// 确保上传目录存在
	if err := os.MkdirAll(config.LocalPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &localStorage{
		config: config,
	}, nil
}

// Upload 上传文件
func (s *localStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 生成唯一文件名
	filename := s.generateFilename(file.Filename)

	// 确保目标目录存在
	destPath := filepath.Join(s.config.LocalPath, filename)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 创建目标文件
	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	// 复制文件内容
	if _, err := io.Copy(dest, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// 返回文件 URL
	return s.GetURL(filename), nil
}

// UploadFromReader 从 Reader 上传
func (s *localStorage) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	// 生成唯一文件名
	newFilename := s.generateFilename(filename)

	// 确保目标目录存在
	destPath := filepath.Join(s.config.LocalPath, newFilename)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 创建目标文件
	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	// 复制文件内容
	if _, err := io.Copy(dest, reader); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// 返回文件 URL
	return s.GetURL(newFilename), nil
}

// Delete 删除文件
func (s *localStorage) Delete(ctx context.Context, url string) error {
	// 从 URL 提取文件路径
	// 安全处理：确保文件在允许的目录内
	urlPath := strings.TrimPrefix(url, s.config.BaseURL+"/uploads/")

	// 清理路径，防止路径遍历攻击
	cleanPath := filepath.Clean(urlPath)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	// 构建完整文件路径
	filePath := filepath.Join(s.config.LocalPath, cleanPath)

	// 确保文件在允许的目录内
	absLocalPath, err := filepath.Abs(s.config.LocalPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute file path: %w", err)
	}

	if !strings.HasPrefix(absFilePath, absLocalPath) {
		return fmt.Errorf("invalid file path: outside allowed directory")
	}

	// 删除文件
	if err := os.Remove(absFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，视为成功
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL 获取文件 URL
func (s *localStorage) GetURL(filename string) string {
	return fmt.Sprintf("%s/uploads/%s", s.config.BaseURL, filename)
}

// generateFilename 生成唯一文件名
func (s *localStorage) generateFilename(originalFilename string) string {
	// 清理原始文件名，移除危险字符
	cleanName := filepath.Base(originalFilename)
	cleanName = strings.ReplaceAll(cleanName, "..", "")
	cleanName = strings.ReplaceAll(cleanName, "/", "")
	cleanName = strings.ReplaceAll(cleanName, "\\", "")

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(cleanName))

	// 生成 UUID
	id := uuid.New().String()

	// 生成日期路径 (YYYY/MM/DD)
	now := time.Now()
	datePath := now.Format("2006/01/02")

	// 组合文件名: YYYY/MM/DD/uuid.ext
	return filepath.Join(datePath, fmt.Sprintf("%s%s", id, ext))
}

// s3Storage S3 存储实现 (占位符)
type s3Storage struct {
	config *Config
}

func newS3Storage(config *Config) (StorageService, error) {
	// TODO: 实现 S3 存储
	// 需要安装: github.com/aws/aws-sdk-go-v2
	return &s3Storage{config: config}, nil
}

func (s *s3Storage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return "", fmt.Errorf("S3 storage not implemented yet")
}

func (s *s3Storage) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return "", fmt.Errorf("S3 storage not implemented yet")
}

func (s *s3Storage) Delete(ctx context.Context, url string) error {
	return fmt.Errorf("S3 storage not implemented yet")
}

func (s *s3Storage) GetURL(filename string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.config.Bucket, s.config.Region, filename)
}

// ossStorage 阿里云 OSS 存储实现 (占位符)
type ossStorage struct {
	config *Config
}

func newOSSStorage(config *Config) (StorageService, error) {
	// TODO: 实现 OSS 存储
	// 需要安装: github.com/aliyun/aliyun-oss-go-sdk
	return &ossStorage{config: config}, nil
}

func (s *ossStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return "", fmt.Errorf("OSS storage not implemented yet")
}

func (s *ossStorage) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return "", fmt.Errorf("OSS storage not implemented yet")
}

func (s *ossStorage) Delete(ctx context.Context, url string) error {
	return fmt.Errorf("OSS storage not implemented yet")
}

func (s *ossStorage) GetURL(filename string) string {
	return fmt.Sprintf("https://%s.%s.aliyuncs.com/%s", s.config.Bucket, s.config.Region, filename)
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = string(StorageTypeLocal)
	}

	return &Config{
		Type:            StorageType(storageType),
		LocalPath:       getEnv("STORAGE_LOCAL_PATH", "./uploads"),
		BaseURL:         getEnv("STORAGE_BASE_URL", "http://localhost:8080"),
		Bucket:          os.Getenv("STORAGE_BUCKET"),
		Region:          os.Getenv("STORAGE_REGION"),
		AccessKeyID:     os.Getenv("STORAGE_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("STORAGE_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("STORAGE_ENDPOINT"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ValidateFile 验证上传文件
func ValidateFile(file *multipart.FileHeader, maxSize int64, allowedTypes []string) error {
	// 检查文件大小
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// 检查文件类型
	if len(allowedTypes) > 0 {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		allowed := false
		for _, allowedType := range allowedTypes {
			if ext == allowedType {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("file type %s is not allowed", ext)
		}
	}

	return nil
}

// Common file type constants
var (
	ImageTypes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"}
	VideoTypes = []string{".mp4", ".avi", ".mov", ".wmv", ".flv"}
	DocTypes   = []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"}
)
