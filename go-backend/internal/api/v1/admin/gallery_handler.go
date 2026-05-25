package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type GalleryHandler struct {
	galleryRepo *repository.GalleryRepository
}

func NewGalleryHandler(galleryRepo *repository.GalleryRepository) *GalleryHandler {
	return &GalleryHandler{
		galleryRepo: galleryRepo,
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

	galleries, total, err := h.galleryRepo.FindAllGalleries(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch galleries"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"galleries": galleries,
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

	galleryItem, err := h.galleryRepo.FindGalleryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"gallery": galleryItem,
	})
}

// CreateGallery 创建图库
// POST /api/admin/galleries
func (h *GalleryHandler) CreateGallery(c *gin.Context) {
	var req struct {
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		Slug        string `json:"slug" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newGallery := &gallery.Gallery{
		Name:        req.Title,
		Description: req.Description,
		Slug:        req.Slug,
	}

	if err := h.galleryRepo.CreateGallery(newGallery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gallery"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Gallery created successfully",
		"gallery": newGallery,
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
		Title       string `json:"title"`
		Description string `json:"description"`
		Slug        string `json:"slug"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingGallery, err := h.galleryRepo.FindGalleryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	if req.Title != "" {
		existingGallery.Name = req.Title
	}
	if req.Description != "" {
		existingGallery.Description = req.Description
	}
	if req.Slug != "" {
		existingGallery.Slug = req.Slug
	}

	if err := h.galleryRepo.UpdateGallery(existingGallery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update gallery"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gallery updated successfully",
		"gallery": existingGallery,
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

	if err := h.galleryRepo.DeleteGallery(uint(id)); err != nil {
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

	images, err := h.galleryRepo.FindImagesByGalleryID(uint(id))
	if err != nil {
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
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
		URL         string `json:"url" binding:"required"`
		Thumbnail   string `json:"thumbnail"`
		Tags        string `json:"tags"`
		Order       int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newImage := &gallery.GalleryImage{
		GalleryID:   uint(galleryID),
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Thumbnail:   req.Thumbnail,
		Tags:        req.Tags,
		Order:       req.Order,
	}

	if err := h.galleryRepo.CreateGalleryImage(newImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create image"})
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
	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Thumbnail   string `json:"thumbnail"`
		Tags        string `json:"tags"`
		Order       int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingImage, err := h.galleryRepo.FindGalleryImageByID(uint(imageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	if req.Title != "" {
		existingImage.Title = req.Title
	}
	if req.Description != "" {
		existingImage.Description = req.Description
	}
	if req.URL != "" {
		existingImage.URL = req.URL
	}
	if req.Thumbnail != "" {
		existingImage.Thumbnail = req.Thumbnail
	}
	if req.Tags != "" {
		existingImage.Tags = req.Tags
	}
	existingImage.Order = req.Order

	if err := h.galleryRepo.UpdateGalleryImage(existingImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image updated successfully",
		"image":   existingImage,
	})
}

// DeleteImage 删除图片
// DELETE /api/admin/galleries/:id/images/:imageId
func (h *GalleryHandler) DeleteImage(c *gin.Context) {
	imageID, err := strconv.ParseUint(c.Param("imageId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	if err := h.galleryRepo.DeleteGalleryImage(uint(imageID)); err != nil {
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
	var req struct {
		ImageIDs []uint `json:"image_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.galleryRepo.BatchDeleteImages(req.ImageIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to batch delete images"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": len(req.ImageIDs),
	})
}
