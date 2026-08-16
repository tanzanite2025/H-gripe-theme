package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	customerServiceAvatarMaxRequestBytes = 3 << 20
	customerServiceAvatarMultipartMemory = 1 << 20
)

type CustomerServiceAvatarHandler struct {
	avatarService *service.CustomerServiceAvatarService
}

func NewCustomerServiceAvatarHandler(avatarService *service.CustomerServiceAvatarService) *CustomerServiceAvatarHandler {
	return &CustomerServiceAvatarHandler{avatarService: avatarService}
}

func (h *CustomerServiceAvatarHandler) UploadAvatar(c *gin.Context) {
	if h == nil || h.avatarService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer-service avatar storage is unavailable"})
		return
	}
	userID, ok := customerServiceAvatarUserID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, customerServiceAvatarMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(customerServiceAvatarMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "customer-service avatar upload is too large", "code": "request_too_large"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "code": upload.CodeEmptyFile})
		return
	}

	avatarURL, err := h.avatarService.Upload(c.Request.Context(), userID, file)
	if err != nil {
		respondCustomerServiceAvatarError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"avatar": avatarURL})
}

func (h *CustomerServiceAvatarHandler) DeleteAvatar(c *gin.Context) {
	if h == nil || h.avatarService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer-service avatar storage is unavailable"})
		return
	}
	userID, ok := customerServiceAvatarUserID(c)
	if !ok {
		return
	}
	if err := h.avatarService.Remove(c.Request.Context(), userID); err != nil {
		respondCustomerServiceAvatarError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar": ""})
}

func customerServiceAvatarUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public chat agent user id"})
		return 0, false
	}
	return uint(id), true
}

func respondCustomerServiceAvatarError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerServiceAvatarProfileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrCustomerServiceAvatarStorageUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "customer-service avatar storage is unavailable"})
	case upload.ErrorCode(err) != "invalid_upload":
		c.JSON(upload.HTTPStatus(err), gin.H{"error": err.Error(), "code": upload.ErrorCode(err)})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update customer-service avatar"})
	}
}
