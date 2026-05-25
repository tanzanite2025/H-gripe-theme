package gallery

import (
	"net/http"
	"strconv"
	"strings"

	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/service"

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

// ========== 公开 API ==========

// GetGalleries 获取图片库列表
// GET /api/v1/galleries
func (h *GalleryHandler) GetGalleries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	galleries, total, err := h.galleryService.GetAllGalleries(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data":        galleries,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetGalleryByID 获取单个图片库
// GET /api/v1/galleries/:id
func (h *GalleryHandler) GetGalleryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	gallery, err := h.galleryService.GetGalleryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gallery})
}

// GetGalleryImages 获取图片库的所有图片
// GET /api/v1/galleries/:id/images
func (h *GalleryHandler) GetGalleryImages(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	images, err := h.galleryService.GetImagesByGalleryID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  images,
		"total": len(images),
	})
}

// SearchImages 搜索图片
// GET /api/v1/galleries/images/search?q=keyword
func (h *GalleryHandler) SearchImages(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search keyword is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	images, total, err := h.galleryService.SearchImages(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data":        images,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"keyword":     keyword,
	})
}

// GetImagesByTags 根据标签获取图片
// GET /api/v1/galleries/images/tags?tags=tag1,tag2
func (h *GalleryHandler) GetImagesByTags(c *gin.Context) {
	tagsStr := c.Query("tags")
	if tagsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tags parameter is required"})
		return
	}

	tags := strings.Split(tagsStr, ",")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	images, total, err := h.galleryService.GetImagesByTags(tags, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"data":        images,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
		"tags":        tags,
	})
}

// ========== 管理员 API ==========

// CreateGallery 创建图片库
// POST /api/v1/admin/galleries
func (h *GalleryHandler) CreateGallery(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Slug        string `json:"slug" binding:"required"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认状态
	if req.Status == "" {
		req.Status = "published"
	}

	gallery := &gallery.Gallery{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Status:      req.Status,
	}

	if err := h.galleryService.CreateGallery(gallery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Gallery created successfully",
		"data":    gallery,
	})
}

// UpdateGallery 更新图片库
// PUT /api/v1/admin/galleries/:id
func (h *GalleryHandler) UpdateGallery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	// 检查图片库是否存在
	existingGallery, err := h.galleryService.GetGalleryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		CoverImage  string `json:"cover_image"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		existingGallery.Name = req.Name
	}
	if req.Slug != "" {
		existingGallery.Slug = req.Slug
	}
	if req.Description != "" {
		existingGallery.Description = req.Description
	}
	if req.CoverImage != "" {
		existingGallery.CoverImage = req.CoverImage
	}
	if req.Status != "" {
		existingGallery.Status = req.Status
	}

	if err := h.galleryService.UpdateGallery(existingGallery); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Gallery updated successfully",
		"data":    existingGallery,
	})
}

// DeleteGallery 删除图片库
// DELETE /api/v1/admin/galleries/:id
func (h *GalleryHandler) DeleteGallery(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	if err := h.galleryService.DeleteGallery(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gallery deleted successfully"})
}

// CreateGalleryImage 创建图片
// POST /api/v1/admin/galleries/:id/images
func (h *GalleryHandler) CreateGalleryImage(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	// 检查图片库是否存在
	if _, err := h.galleryService.GetGalleryByID(uint(galleryID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	var req struct {
		URL         string `json:"url" binding:"required"`
		Thumbnail   string `json:"thumbnail"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Alt         string `json:"alt"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Size        int64  `json:"size"`
		Tags        string `json:"tags"`
		Order       int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image := &gallery.GalleryImage{
		GalleryID:   uint(galleryID),
		URL:         req.URL,
		Thumbnail:   req.Thumbnail,
		Title:       req.Title,
		Description: req.Description,
		Alt:         req.Alt,
		Width:       req.Width,
		Height:      req.Height,
		Size:        req.Size,
		Tags:        req.Tags,
		Order:       req.Order,
	}

	if err := h.galleryService.CreateGalleryImage(image); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Image created successfully",
		"data":    image,
	})
}

// UpdateGalleryImage 更新图片
// PUT /api/v1/admin/galleries/images/:id
func (h *GalleryHandler) UpdateGalleryImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	// 检查图片是否存在
	existingImage, err := h.galleryService.GetGalleryImageByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	var req struct {
		URL         string `json:"url"`
		Thumbnail   string `json:"thumbnail"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Alt         string `json:"alt"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Size        int64  `json:"size"`
		Tags        string `json:"tags"`
		Order       int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.URL != "" {
		existingImage.URL = req.URL
	}
	if req.Thumbnail != "" {
		existingImage.Thumbnail = req.Thumbnail
	}
	if req.Title != "" {
		existingImage.Title = req.Title
	}
	if req.Description != "" {
		existingImage.Description = req.Description
	}
	if req.Alt != "" {
		existingImage.Alt = req.Alt
	}
	if req.Width > 0 {
		existingImage.Width = req.Width
	}
	if req.Height > 0 {
		existingImage.Height = req.Height
	}
	if req.Size > 0 {
		existingImage.Size = req.Size
	}
	if req.Tags != "" {
		existingImage.Tags = req.Tags
	}
	if req.Order >= 0 {
		existingImage.Order = req.Order
	}

	if err := h.galleryService.UpdateGalleryImage(existingImage); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image updated successfully",
		"data":    existingImage,
	})
}

// DeleteGalleryImage 删除图片
// DELETE /api/v1/admin/galleries/images/:id
func (h *GalleryHandler) DeleteGalleryImage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
		return
	}

	if err := h.galleryService.DeleteGalleryImage(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

// BatchCreateImages 批量创建图片
// POST /api/v1/admin/galleries/:id/images/batch
func (h *GalleryHandler) BatchCreateImages(c *gin.Context) {
	galleryID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gallery ID"})
		return
	}

	// 检查图片库是否存在
	if _, err := h.galleryService.GetGalleryByID(uint(galleryID)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gallery not found"})
		return
	}

	var req struct {
		Images []struct {
			URL         string `json:"url" binding:"required"`
			Thumbnail   string `json:"thumbnail"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Alt         string `json:"alt"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			Size        int64  `json:"size"`
			Tags        string `json:"tags"`
			Order       int    `json:"order"`
		} `json:"images" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	images := make([]gallery.GalleryImage, len(req.Images))
	for i, img := range req.Images {
		images[i] = gallery.GalleryImage{
			GalleryID:   uint(galleryID),
			URL:         img.URL,
			Thumbnail:   img.Thumbnail,
			Title:       img.Title,
			Description: img.Description,
			Alt:         img.Alt,
			Width:       img.Width,
			Height:      img.Height,
			Size:        img.Size,
			Tags:        img.Tags,
			Order:       img.Order,
		}
	}

	if err := h.galleryService.BatchCreateImages(images); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Images created successfully",
		"count":   len(images),
	})
}

// BatchDeleteImages 批量删除图片
// DELETE /api/v1/admin/galleries/images/batch
func (h *GalleryHandler) BatchDeleteImages(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.galleryService.BatchDeleteImages(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Images deleted successfully",
		"count":   len(req.IDs),
	})
}

// BatchUpdateOrder 批量更新图片排序
// POST /api/v1/admin/galleries/images/batch-order
func (h *GalleryHandler) BatchUpdateOrder(c *gin.Context) {
	var req struct {
		Orders map[uint]int `json:"orders" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.galleryService.BatchUpdateOrder(req.Orders); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image orders updated successfully",
		"count":   len(req.Orders),
	})
}
