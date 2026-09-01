package admin

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strings"
	"unicode/utf8"

	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	mediaMaxRequestBytes = 82 << 20
	mediaMultipartMemory = 4 << 20
	mediaMaxAltRunes     = 500
	mediaMaxCaptionRunes = 5000
)

func (h *MediaHandler) UploadAsset(c *gin.Context) {
	h.uploadAsset(c, "")
}

func (h *MediaHandler) UploadRefundReturnPolicyImage(c *gin.Context) {
	h.uploadAsset(c, string(upload.SpecRefundReturnImage))
}

func (h *MediaHandler) uploadAsset(c *gin.Context, forcedPurpose string) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, mediaMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(mediaMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "media upload request is too large", "code": "request_too_large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart upload", "code": "invalid_upload"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	requestedType := c.PostForm("media_type")
	requestedPurpose := c.PostForm("image_purpose")
	if forcedPurpose != "" {
		requestedType = "image"
		requestedPurpose = forcedPurpose
	}
	mediaType, err := validateMediaUpload(
		file,
		requestedType,
		requestedPurpose,
	)
	if err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}

	width, height := 0, 0
	if mediaType == "image" {
		width, height, err = upload.ReadImageDimensions(file)
		if err != nil {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "unable to read image dimensions",
				"code":  upload.CodeInvalidType,
			})
			return
		}
	}

	alt := strings.TrimSpace(c.PostForm("alt"))
	caption := strings.TrimSpace(c.PostForm("caption"))
	if utf8.RuneCountInString(alt) > mediaMaxAltRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "alt text is too long", "code": "metadata_too_long"})
		return
	}
	if utf8.RuneCountInString(caption) > mediaMaxCaptionRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caption is too long", "code": "metadata_too_long"})
		return
	}

	asset, err := h.mediaService.UploadAsset(c.Request.Context(), service.MediaUploadInput{
		File:       file,
		MediaType:  mediaType,
		Alt:        alt,
		Caption:    caption,
		UploaderID: currentUserID(c),
		Width:      width,
		Height:     height,
	})
	if err != nil {
		respondMediaError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Response{Code: 0, Message: "Media asset uploaded successfully", Data: gin.H{"asset": asset}})
}

func isRequestBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesError *http.MaxBytesError
	return errors.As(err, &maxBytesError) || strings.Contains(strings.ToLower(err.Error()), "request body too large")
}

func validateMediaUpload(file *multipart.FileHeader, requestedType, requestedPurpose string) (string, error) {
	purpose := strings.TrimSpace(requestedPurpose)
	switch strings.ToLower(strings.TrimSpace(requestedType)) {
	case "image":
		if purpose != "" {
			if _, ok := upload.RuleForSpec(purpose); !ok {
				return "", upload.ValidateSpecFile(file, purpose)
			}
			return "image", upload.ValidateSpecFile(file, purpose)
		}
		return "image", upload.ValidateFile(file, upload.ProductImageRule)
	case "video":
		return "video", upload.ValidateFile(file, upload.ProductVideoRule)
	default:
		if purpose != "" {
			if _, ok := upload.RuleForSpec(purpose); !ok {
				return "", upload.ValidateSpecFile(file, purpose)
			}
			return "image", upload.ValidateSpecFile(file, purpose)
		}
		if err := upload.ValidateFile(file, upload.ProductImageRule); err == nil {
			return "image", nil
		}
		if err := upload.ValidateFile(file, upload.ProductVideoRule); err == nil {
			return "video", nil
		} else {
			return "", err
		}
	}
}
