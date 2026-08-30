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
)

// StorageService 存储服务接口
type StorageService interface {
	Upload(ctx context.Context, file *multipart.FileHeader) (string, error)
	UploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error)
	UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error)
	UploadFromReaderWithPrefix(ctx context.Context, reader io.Reader, filename string, prefix string) (string, error)
	Delete(ctx context.Context, url string) error
	GetURL(filename string) string
	ObjectKey(reference string) (string, error)
	CopyObject(ctx context.Context, sourceKey string, destKey string) error
}

// PrivateObjectUploader marks uploads that must remain private until an
// application-level moderation decision exposes them.
type PrivateObjectUploader interface {
	UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error)
}

// CacheControlledObjectUploader is an optional capability for immutable,
// first-party public objects. It keeps cache policy at the storage boundary
// without making unrelated uploads inherit long-lived browser/CDN caching.
type CacheControlledObjectUploader interface {
	UploadWithPrefixAndCacheControl(ctx context.Context, file *multipart.FileHeader, prefix string, cacheControl string) (string, error)
}

type CacheControlledReaderUploader interface {
	UploadFromReaderWithPrefixAndCacheControl(ctx context.Context, reader io.Reader, filename string, prefix string, cacheControl string) (string, error)
}

type StoredObject struct {
	ReadCloser io.ReadCloser
	Name       string
	MimeType   string
	Size       int64
	ModTime    time.Time
}

type ObjectOpener interface {
	Open(ctx context.Context, key string) (*StoredObject, error)
}

type PresignedURLProvider interface {
	GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
}

// StorageType 存储类型
type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
	StorageTypeOSS   StorageType = "oss"

	localDirPerm  os.FileMode = 0o755
	localFilePerm os.FileMode = 0o644
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
	if err := ensureLocalDirectory(config.LocalPath, config.LocalPath); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &localStorage{
		config: config,
	}, nil
}

// Upload 上传文件
func (s *localStorage) Upload(ctx context.Context, file *multipart.FileHeader) (string, error) {
	return s.UploadWithPrefix(ctx, file, "")
}

func (s *localStorage) UploadWithPrefix(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
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

	// 确保目标目录存在
	destPath, err := s.localPath(filename)
	if err != nil {
		return "", err
	}
	destDir := filepath.Dir(destPath)
	if err := ensureLocalDirectory(s.config.LocalPath, destDir); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// #nosec G304 -- destPath is constrained to the configured upload root by localPath.
	dest, err := createLocalUploadFile(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dest.Close() }()

	// 复制文件内容
	if _, err := io.Copy(dest, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// 返回文件 URL
	return s.GetURL(filename), nil
}

func (s *localStorage) UploadWithPrefixAndCacheControl(ctx context.Context, file *multipart.FileHeader, prefix string, _ string) (string, error) {
	// Local upload responses set the policy in uploads_route.go.
	return s.UploadWithPrefix(ctx, file, prefix)
}

func (s *localStorage) UploadWithPrefixPrivate(ctx context.Context, file *multipart.FileHeader, prefix string) (string, error) {
	// Local objects are protected by the public upload authorization route.
	return s.UploadWithPrefix(ctx, file, prefix)
}

// UploadFromReader 从 Reader 上传
func (s *localStorage) UploadFromReader(ctx context.Context, reader io.Reader, filename string) (string, error) {
	return s.UploadFromReaderWithPrefix(ctx, reader, filename, "")
}

func (s *localStorage) UploadFromReaderWithPrefix(ctx context.Context, reader io.Reader, filename string, prefix string) (string, error) {
	return s.UploadFromReaderWithPrefixAndCacheControl(ctx, reader, filename, prefix, "")
}

func (s *localStorage) UploadFromReaderWithPrefixAndCacheControl(_ context.Context, reader io.Reader, filename string, prefix string, _ string) (string, error) {
	// 生成唯一文件名
	newFilename, err := generateObjectKey(filename, prefix)
	if err != nil {
		return "", err
	}

	// 确保目标目录存在
	destPath, err := s.localPath(newFilename)
	if err != nil {
		return "", err
	}
	destDir := filepath.Dir(destPath)
	if err := ensureLocalDirectory(s.config.LocalPath, destDir); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	// #nosec G304 -- destPath is constrained to the configured upload root by localPath.
	dest, err := createLocalUploadFile(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dest.Close() }()

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
	urlPath, err := s.ObjectKey(url)
	if err != nil {
		return err
	}

	absFilePath, err := s.localPath(urlPath)
	if err != nil {
		return err
	}

	// 删除文件
	// #nosec G304 -- absFilePath is constrained to the configured upload root by localPath.
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
	cleanName, ok := NormalizeObjectKey(filename)
	if !ok {
		cleanName = strings.TrimPrefix(filepath.ToSlash(filename), "/")
	}
	return fmt.Sprintf("%s/uploads/%s", strings.TrimRight(s.config.BaseURL, "/"), cleanName)
}

func (s *localStorage) ObjectKey(reference string) (string, error) {
	key, ok := ObjectKeyFromReference(reference, s.config.BaseURL)
	if !ok {
		return "", fmt.Errorf("invalid object key")
	}
	return key, nil
}

func (s *localStorage) Open(ctx context.Context, key string) (*StoredObject, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filePath, err := s.localPath(key)
	if err != nil {
		return nil, err
	}

	// #nosec G304 -- filePath is constrained to the configured upload root by localPath.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}

	return &StoredObject{
		ReadCloser: file,
		Name:       filepath.Base(key),
		MimeType:   detectContentType(key),
		Size:       stat.Size(),
		ModTime:    stat.ModTime(),
	}, nil
}

func (s *localStorage) CopyObject(ctx context.Context, sourceKey string, destKey string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	normalizedSourceKey, ok := NormalizeObjectKey(sourceKey)
	if !ok {
		return fmt.Errorf("invalid source object key")
	}
	normalizedDestKey, ok := NormalizeObjectKey(destKey)
	if !ok {
		return fmt.Errorf("invalid destination object key")
	}

	sourcePath, err := s.localPath(normalizedSourceKey)
	if err != nil {
		return err
	}
	destPath, err := s.localPath(normalizedDestKey)
	if err != nil {
		return err
	}
	destDir := filepath.Dir(destPath)
	if err := ensureLocalDirectory(s.config.LocalPath, destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// #nosec G304 -- both paths are constrained to the configured upload root by localPath.
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = source.Close() }()

	// #nosec G304 -- destPath is constrained to the configured upload root by localPath.
	dest, err := createLocalUploadFile(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dest.Close() }()

	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	return nil
}

func (s *localStorage) localPath(name string) (string, error) {
	cleanName := filepath.Clean(name)
	if cleanName == "." || cleanName == ".." || filepath.IsAbs(cleanName) {
		return "", fmt.Errorf("invalid file path")
	}

	parentPrefix := ".." + string(os.PathSeparator)
	if strings.HasPrefix(cleanName, parentPrefix) {
		return "", fmt.Errorf("invalid file path: path traversal detected")
	}

	root, err := filepath.Abs(s.config.LocalPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute upload path: %w", err)
	}

	target, err := filepath.Abs(filepath.Join(root, cleanName))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute file path: %w", err)
	}

	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("failed to validate file path: %w", err)
	}

	if rel == ".." || strings.HasPrefix(rel, parentPrefix) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid file path: outside upload directory")
	}

	return target, nil
}

func ensureLocalDirectory(root, dir string) error {
	if err := os.MkdirAll(dir, localDirPerm); err != nil {
		return err
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("failed to get absolute upload path: %w", err)
	}
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute directory path: %w", err)
	}

	rel, err := filepath.Rel(rootAbs, dirAbs)
	if err != nil {
		return fmt.Errorf("failed to validate directory path: %w", err)
	}
	parentPrefix := ".." + string(os.PathSeparator)
	if rel == ".." || strings.HasPrefix(rel, parentPrefix) || filepath.IsAbs(rel) {
		return fmt.Errorf("invalid directory path: outside upload directory")
	}

	current := rootAbs
	if err := bestEffortChmod(current, localDirPerm); err != nil {
		return err
	}
	if rel == "." {
		return nil
	}
	for _, part := range strings.Split(rel, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		if err := bestEffortChmod(current, localDirPerm); err != nil {
			return err
		}
	}
	return nil
}

func createLocalUploadFile(destPath string) (*os.File, error) {
	file, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, localFilePerm)
	if err != nil {
		return nil, err
	}
	if err := bestEffortChmodFile(file, localFilePerm); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

// Some managed volumes accept writes but reject chmod on mounted paths.
// Treat permission-denied chmods as best-effort so storage startup keeps moving.
func bestEffortChmod(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err != nil && !os.IsPermission(err) {
		return err
	}
	return nil
}

func bestEffortChmodFile(file *os.File, mode os.FileMode) error {
	if err := file.Chmod(mode); err != nil && !os.IsPermission(err) {
		return err
	}
	return nil
}

func newS3Storage(config *Config) (StorageService, error) {
	return NewS3Storage(config)
}

func newOSSStorage(config *Config) (StorageService, error) {
	return NewOSSStorage(config)
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

func LoadSiteLogoConfigFromEnv(base *Config) *Config {
	cfg := Config{}
	if base != nil {
		cfg = *base
	} else {
		cfg = *LoadConfigFromEnv()
	}

	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_TYPE")); value != "" {
		cfg.Type = StorageType(value)
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_LOCAL_PATH")); value != "" {
		cfg.LocalPath = value
	} else if cfg.Type == StorageTypeLocal {
		cfg.LocalPath = defaultSiteLogoLocalPath(cfg.LocalPath)
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_BASE_URL")); value != "" {
		cfg.BaseURL = value
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_BUCKET")); value != "" {
		cfg.Bucket = value
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_REGION")); value != "" {
		cfg.Region = value
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_ACCESS_KEY_ID")); value != "" {
		cfg.AccessKeyID = value
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_SECRET_ACCESS_KEY")); value != "" {
		cfg.SecretAccessKey = value
	}
	if value := strings.TrimSpace(os.Getenv("SITE_LOGO_STORAGE_ENDPOINT")); value != "" {
		cfg.Endpoint = value
	}
	return &cfg
}

func defaultSiteLogoLocalPath(localPath string) string {
	cleanPath := filepath.Clean(strings.TrimSpace(localPath))
	if cleanPath == "" || cleanPath == "." {
		return "./site-logo-uploads"
	}
	dir := filepath.Dir(cleanPath)
	if dir == "." {
		return "./site-logo-uploads"
	}
	return filepath.Join(dir, "site-logo-uploads")
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
