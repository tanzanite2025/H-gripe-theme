package service

import (
	"mime/multipart"
	"strings"

	"commerce-platform/internal/pkg/upload"
)

func validateAgentCustomerServiceAttachmentForCurrentPrivateUploadRollout(file *multipart.FileHeader, privateUploadsEnabled bool) (string, int, int, error) {
	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return "", 0, 0, err
	}
	if strings.HasPrefix(mimeType, "video/") {
		if privateUploadsEnabled {
			return "", 0, 0, ErrUnsupportedMediaType
		}
		return "video", 0, 0, upload.ValidateFile(file, upload.ProductVideoRule)
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecCustomerServiceAttachment)); err != nil {
		return "", 0, 0, err
	}
	width, height, err := upload.ReadImageDimensions(file)
	return "image", width, height, err
}
