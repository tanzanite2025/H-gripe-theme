package upload

import (
	"mime/multipart"
)

// ValidateTotalSize validates the combined size of a mixed upload. It is
// separate from ValidateFiles because different file classes can have
// different per-file rules while sharing one request-level attachment cap.
func ValidateTotalSize(files []*multipart.FileHeader, maxTotalSize int64) error {
	if maxTotalSize <= 0 {
		return nil
	}

	var totalSize int64
	for _, file := range files {
		if file == nil {
			continue
		}
		if file.Size < 0 || file.Size > maxTotalSize-totalSize {
			return validationError(CodeFileTooLarge, "file_too_large: total upload size exceeds %s", formatBytes(maxTotalSize))
		}
		totalSize += file.Size
	}
	return nil
}
