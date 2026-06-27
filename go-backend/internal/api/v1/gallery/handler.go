package gallery

import (
	"net/http"
	"strconv"
	"strings"

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
