package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type GalleryHandler struct {
	galleryService *service.GalleryService
}

func NewGalleryHandler(galleryService *service.GalleryService) *GalleryHandler {
	return &GalleryHandler{
		galleryService: galleryService,
	}
}

func respondGalleryServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGalleryMediaAssetRequired),
		errors.Is(err, service.ErrGalleryMediaAssetInvalid),
		errors.Is(err, service.ErrGalleryImageTitleRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrGalleryMediaAssetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrGalleryNotFound),
		errors.Is(err, service.ErrGalleryMediaAssetUnavailable),
		service.IsRecordNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gallery operation failed"})
	}
}

// ListGalleries 获取图库列表
// GET /api/admin/galleries
func (h *GalleryHandler) ListGalleries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	galleries, total, err := h.galleryService.GetAllGalleries(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch galleries"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"galleries": adminGalleriesFromDomain(galleries),
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetGallery 获取图库详情
// GET /api/admin/galleries/:id
func (h *GalleryHandler) GetGallery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	galleryItem, err := h.galleryService.GetGalleryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gallery": adminGalleryFromDomain(galleryItem),
	})
}

// CreateGallery 创建图库
// POST /api/admin/galleries
func (h *GalleryHandler) CreateGallery(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Slug        string `json:"slug" binding:"required"`
		ProductIDs  []uint `json:"product_ids"`
		Images      []struct {
			MediaAssetID uint   `json:"media_asset_id"`
			Title        string `json:"title"`
			Description  string `json:"description"`
			Tags         string `json:"tags"`
			Order        int    `json:"order"`
		} `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdInput := service.GalleryAdminCreateInput{
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		ProductIDs:  req.ProductIDs,
	}
	for _, image := range req.Images {
		createdInput.Images = append(createdInput.Images, service.GalleryImageAdminCreateInput{
			MediaAssetID: image.MediaAssetID,
			Title:        image.Title,
			Description:  image.Description,
			Tags:         image.Tags,
			Order:        image.Order,
		})
	}
	createdGallery, err := h.galleryService.CreateAdminGallery(createdInput)
	if err != nil {
		respondGalleryServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Gallery created successfully",
		"gallery": adminGalleryFromDomain(createdGallery),
	})
}

// UpdateGallery 更新图库
// PUT /api/admin/galleries/:id
func (h *GalleryHandler) UpdateGallery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Slug        *string `json:"slug"`
		ProductIDs  *[]uint `json:"product_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedGallery, err := h.galleryService.UpdateAdminGallery(uint(id), service.GalleryAdminUpdateInput{
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		ProductIDs:  req.ProductIDs,
	})
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
			return
		}
		respondGalleryServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gallery updated successfully",
		"gallery": adminGalleryFromDomain(updatedGallery),
	})
}

// DeleteGallery 删除图库
// DELETE /api/admin/galleries/:id
func (h *GalleryHandler) DeleteGallery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	if err := h.galleryService.DeleteGallery(uint(id)); err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gallery deleted successfully",
	})
}

// ListImages 获取图库的图片列表
// GET /api/admin/galleries/:id/images
func (h *GalleryHandler) ListImages(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	images, err := h.galleryService.GetImagesByGalleryID(uint(id))
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"images": images,
	})
}

// CreateImage 创建图片
// POST /api/admin/galleries/:id/images
func (h *GalleryHandler) CreateImage(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	var req struct {
		Title        string `json:"title" binding:"required"`
		Description  string `json:"description"`
		MediaAssetID *uint  `json:"media_asset_id" binding:"required"`
		Tags         string `json:"tags"`
		Order        int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newImage := &gallery.GalleryImage{
		GalleryID:    uint(galleryID),
		Title:        req.Title,
		Description:  req.Description,
		MediaAssetID: req.MediaAssetID,
		Tags:         req.Tags,
		Order:        req.Order,
	}

	if err := h.galleryService.CreateGalleryImage(newImage); err != nil {
		respondGalleryServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Image created successfully",
		"image":   newImage,
	})
}

// UpdateImage 更新图片
// PUT /api/admin/galleries/:id/images/:imageId
func (h *GalleryHandler) UpdateImage(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	var req struct {
		Title        *string `json:"title"`
		Description  *string `json:"description"`
		MediaAssetID *uint   `json:"media_asset_id"`
		Tags         *string `json:"tags"`
		Order        *int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedImage, err := h.galleryService.UpdateAdminGalleryImageForGallery(uint(galleryID), uint(imageID), service.GalleryImageAdminUpdateInput{
		Title:        req.Title,
		Description:  req.Description,
		MediaAssetID: req.MediaAssetID,
		Tags:         req.Tags,
		Order:        req.Order,
	})
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		respondGalleryServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image updated successfully",
		"image":   updatedImage,
	})
}

// DeleteImage 删除图片
// DELETE /api/admin/galleries/:id/images/:imageId
func (h *GalleryHandler) DeleteImage(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	if err := h.galleryService.DeleteGalleryImageForGallery(uint(galleryID), uint(imageID)); err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}

// BatchDeleteImages 批量删除图片
// POST /api/admin/galleries/:id/images/batch-delete
func (h *GalleryHandler) BatchDeleteImages(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	var req struct {
		ImageIDs []uint `json:"image_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted, err := h.galleryService.BatchDeleteGalleryImages(uint(galleryID), req.ImageIDs)
	if err != nil {
		if service.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "One or more images were not found in this gallery"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch delete images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
	})
}
