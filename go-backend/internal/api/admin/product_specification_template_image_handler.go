package admin

import (
	"net/http"
	"strings"

	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const productSpecificationTemplateImageMaxRequestBytes = 4 << 20

type ProductSpecificationTemplateImageHandler struct {
	productService *service.ProductService
	mediaService   *service.MediaService
}

func NewProductSpecificationTemplateImageHandler(
	productService *service.ProductService,
	mediaService *service.MediaService,
) *ProductSpecificationTemplateImageHandler {
	return &ProductSpecificationTemplateImageHandler{
		productService: productService,
		mediaService:   mediaService,
	}
}

func (h *ProductSpecificationTemplateImageHandler) UploadImage(c *gin.Context) {
	id, ok := parseProductSpecificationTemplateID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, productSpecificationTemplateImageMaxRequestBytes)
	file, err := c.FormFile("file")
	if err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "product specification template image is too large",
				"code":  upload.CodeFileTooLarge,
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required", "code": upload.CodeEmptyFile})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	if err := upload.ValidateFile(file, upload.ProductSpecificationTemplateImageRule); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{
			"error": err.Error(),
			"code":  upload.ErrorCode(err),
		})
		return
	}

	asset, err := h.mediaService.UploadAsset(c.Request.Context(), service.MediaUploadInput{
		File:       file,
		MediaType:  "image",
		Alt:        strings.TrimSpace(c.PostForm("alt")),
		UploaderID: currentUserID(c),
	})
	if err != nil {
		respondMediaError(c, err)
		return
	}

	assetID := asset.ID
	productSpecificationTemplate, err := h.productService.UpdateProductSpecificationTemplateImage(id, &assetID, asset.URL)
	if err != nil {
		h.productService.CleanupDetachedProductSpecificationTemplateImageAsset(asset.ID, "product specification template image relation update failed")
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": productSpecificationTemplate})
}

func (h *ProductSpecificationTemplateImageHandler) DeleteImage(c *gin.Context) {
	id, ok := parseProductSpecificationTemplateID(c)
	if !ok {
		return
	}

	productSpecificationTemplate, err := h.productService.UpdateProductSpecificationTemplateImage(id, nil, "")
	if err != nil {
		respondProductSpecificationTemplateServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productSpecificationTemplate})
}
