package admin

import (
	"net/http"
	"strings"

	"tanzanite/internal/pkg/upload"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const productTypeImageMaxRequestBytes = 4 << 20

type ProductTypeImageHandler struct {
	productService *service.ProductService
	mediaService   *service.MediaService
}

func NewProductTypeImageHandler(
	productService *service.ProductService,
	mediaService *service.MediaService,
) *ProductTypeImageHandler {
	return &ProductTypeImageHandler{
		productService: productService,
		mediaService:   mediaService,
	}
}

func (h *ProductTypeImageHandler) UploadImage(c *gin.Context) {
	id, ok := parseProductTypeID(c)
	if !ok {
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, productTypeImageMaxRequestBytes)
	file, err := c.FormFile("file")
	if err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "product type image is too large",
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

	if err := upload.ValidateFile(file, upload.ProductTypeImageRule); err != nil {
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
	productType, err := h.productService.UpdateProductTypeImage(id, &assetID, asset.URL)
	if err != nil {
		_ = h.mediaService.DeleteAsset(c.Request.Context(), asset.ID, service.MediaAssetDeleteConfirmation(asset.ID))
		respondProductTypeServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": productType})
}

func (h *ProductTypeImageHandler) DeleteImage(c *gin.Context) {
	id, ok := parseProductTypeID(c)
	if !ok {
		return
	}

	productType, err := h.productService.UpdateProductTypeImage(id, nil, "")
	if err != nil {
		respondProductTypeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": productType})
}
