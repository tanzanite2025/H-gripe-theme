package upload

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	CodeEmptyFile    = "empty_file"
	CodeFileTooLarge = "file_too_large"
	CodeInvalidType  = "invalid_type"
	CodeTooManyFiles = "too_many_files"
)

type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type FileRule struct {
	MaxSize             int64
	AllowedExtensions   []string
	AllowedContentTypes []string
}

type FilesRule struct {
	FileRule
	MaxFiles     int
	MaxTotalSize int64
}

var (
	ShowcaseImageRule = FilesRule{
		FileRule: FileRule{
			MaxSize:             5 << 20,
			AllowedExtensions:   []string{".webp"},
			AllowedContentTypes: []string{"image/webp"},
		},
		MaxFiles:     10,
		MaxTotalSize: 50 << 20,
	}
	SuggestionImageRule = FileRule{
		MaxSize:             5 << 20,
		AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
		AllowedContentTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
	}
	FAQAnswerImageRule = FileRule{
		MaxSize:             3 << 20,
		AllowedExtensions:   []string{".webp"},
		AllowedContentTypes: []string{"image/webp"},
	}
	WarrantyImageRule = FilesRule{
		FileRule: FileRule{
			MaxSize:             8 << 20,
			AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
			AllowedContentTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
		},
		MaxFiles:     10,
		MaxTotalSize: 80 << 20,
	}
	WarrantyVideoRule = FileRule{
		MaxSize:             50 << 20,
		AllowedExtensions:   []string{".mp4", ".mov", ".webm"},
		AllowedContentTypes: []string{"video/mp4", "video/quicktime", "video/webm"},
	}
	ProductImageRule = FileRule{
		MaxSize:             12 << 20,
		AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
		AllowedContentTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
	}
	ProductVideoRule = FileRule{
		MaxSize:             200 << 20,
		AllowedExtensions:   []string{".mp4", ".mov", ".webm"},
		AllowedContentTypes: []string{"video/mp4", "video/quicktime", "video/webm"},
	}
)

func ValidateFile(file *multipart.FileHeader, rule FileRule) error {
	if file == nil || file.Size <= 0 {
		return validationError(CodeEmptyFile, "empty_file: uploaded file is empty")
	}
	if rule.MaxSize > 0 && file.Size > rule.MaxSize {
		return validationError(CodeFileTooLarge, "file_too_large: %s exceeds %s", file.Filename, formatBytes(rule.MaxSize))
	}
	if !extensionAllowed(file.Filename, rule.AllowedExtensions) {
		return validationError(CodeInvalidType, "invalid_type: %s has an unsupported file extension", file.Filename)
	}

	contentType, err := detectContentType(file)
	if err != nil {
		return err
	}
	if !contentTypeAllowed(contentType, rule.AllowedContentTypes) {
		return validationError(CodeInvalidType, "invalid_type: %s content type %s is not allowed", file.Filename, contentType)
	}
	return nil
}

func ValidateWebPDimensions(file *multipart.FileHeader, expectedWidth, expectedHeight int) error {
	if file == nil {
		return validationError(CodeEmptyFile, "empty_file: uploaded file is empty")
	}
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to inspect WebP dimensions: %w", err)
	}
	defer func() { _ = src.Close() }()

	data, err := io.ReadAll(io.LimitReader(src, file.Size+1))
	if err != nil {
		return fmt.Errorf("failed to read WebP file: %w", err)
	}
	width, height, err := parseWebPDimensions(data)
	if err != nil {
		return validationError(CodeInvalidType, "invalid_type: unable to read WebP dimensions")
	}
	if width != expectedWidth || height != expectedHeight {
		return validationError(CodeInvalidType, "invalid_type: FAQ image must be exactly %dx%d pixels (received %dx%d)", expectedWidth, expectedHeight, width, height)
	}
	return nil
}

func ValidateFiles(files []*multipart.FileHeader, rule FilesRule) error {
	if rule.MaxFiles > 0 && len(files) > rule.MaxFiles {
		return validationError(CodeTooManyFiles, "too_many_files: maximum %d files allowed", rule.MaxFiles)
	}

	var totalSize int64
	for _, file := range files {
		if file != nil {
			totalSize += file.Size
		}
		if rule.MaxTotalSize > 0 && totalSize > rule.MaxTotalSize {
			return validationError(CodeFileTooLarge, "file_too_large: total upload size exceeds %s", formatBytes(rule.MaxTotalSize))
		}
		if err := ValidateFile(file, rule.FileRule); err != nil {
			return err
		}
	}
	return nil
}

func ErrorCode(err error) string {
	var validationErr *ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.Code
	}
	return "invalid_upload"
}

func HTTPStatus(err error) int {
	switch ErrorCode(err) {
	case CodeFileTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeInvalidType:
		return http.StatusUnsupportedMediaType
	default:
		return http.StatusBadRequest
	}
}

func detectContentType(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to inspect uploaded file: %w", err)
	}
	defer func() { _ = src.Close() }()

	header := make([]byte, 512)
	n, err := io.ReadFull(src, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to inspect uploaded file: %w", err)
	}
	contentType := http.DetectContentType(header[:n])
	if contentType == "application/octet-stream" {
		if videoContentType := sniffVideoContentType(header[:n]); videoContentType != "" {
			return videoContentType, nil
		}
	}
	return contentType, nil
}

func sniffVideoContentType(header []byte) string {
	if len(header) >= 12 && bytes.Equal(header[4:8], []byte("ftyp")) {
		brand := strings.TrimSpace(string(header[8:12]))
		if brand == "qt" {
			return "video/quicktime"
		}
		return "video/mp4"
	}

	if len(header) >= 4 && bytes.Equal(header[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}) {
		if bytes.Contains(bytes.ToLower(header), []byte("webm")) {
			return "video/webm"
		}
	}
	return ""
}

func parseWebPDimensions(data []byte) (int, int, error) {
	if len(data) < 16 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
		return 0, 0, fmt.Errorf("invalid WebP container")
	}

	chunkType := string(data[12:16])
	switch chunkType {
	case "VP8X":
		if len(data) < 30 {
			return 0, 0, fmt.Errorf("truncated VP8X chunk")
		}
		width := 1 + (int(data[24]) | int(data[25])<<8 | int(data[26])<<16)
		height := 1 + (int(data[27]) | int(data[28])<<8 | int(data[29])<<16)
		return width, height, nil
	case "VP8L":
		if len(data) < 25 || data[20] != 0x2f {
			return 0, 0, fmt.Errorf("invalid VP8L header")
		}
		bits := uint32(data[21]) | uint32(data[22])<<8 | uint32(data[23])<<16 | uint32(data[24])<<24
		width := 1 + int(bits&0x3fff)
		height := 1 + int((bits>>14)&0x3fff)
		return width, height, nil
	case "VP8 ":
		if len(data) < 30 {
			return 0, 0, fmt.Errorf("truncated VP8 chunk")
		}
		for index := 20; index+6 < len(data) && index < 64; index++ {
			if data[index] == 0x9d && data[index+1] == 0x01 && data[index+2] == 0x2a {
				width := int(data[index+3]) | int(data[index+4])<<8
				height := int(data[index+5]) | int(data[index+6])<<8
				return width & 0x3fff, height & 0x3fff, nil
			}
		}
	}
	return 0, 0, fmt.Errorf("unsupported WebP chunk")
}

func extensionAllowed(filename string, allowed []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, item := range allowed {
		if ext == strings.ToLower(item) {
			return true
		}
	}
	return false
}

func contentTypeAllowed(contentType string, allowed []string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	for _, item := range allowed {
		if contentType == strings.ToLower(item) {
			return true
		}
	}
	return false
}

func validationError(code string, format string, args ...interface{}) error {
	return &ValidationError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func formatBytes(size int64) string {
	if size%(1<<20) == 0 {
		return fmt.Sprintf("%dMB", size/(1<<20))
	}
	return fmt.Sprintf("%d bytes", size)
}
