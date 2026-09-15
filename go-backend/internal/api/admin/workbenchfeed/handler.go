package workbenchfeed

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	feed "commerce-platform/internal/workbenchfeed"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	maxUploadRequestBytes = 52 << 20
	multipartMemory       = 8 << 20
)

type Handler struct {
	service *feed.Service
}

func NewHandler(service *feed.Service) *Handler {
	return &Handler{service: service}
}

type entryRequest struct {
	PublishedAt    *string                `json:"published_at"`
	Locale         string                 `json:"locale"`
	Content        string                 `json:"content"`
	Tags           []string               `json:"tags"`
	Status         string                 `json:"status"`
	Media          []mediaRequest         `json:"media"`
	TaggedProducts []taggedProductRequest `json:"tagged_products"`
}

type mediaRequest struct {
	FilePath   string `json:"file_path"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FileSizeKB int    `json:"file_size_kb"`
	Caption    string `json:"caption"`
	SortOrder  int    `json:"sort_order"`
}

type taggedProductRequest struct {
	ProductID    uint   `json:"product_id"`
	VariantID    *uint  `json:"variant_id"`
	DirectAction string `json:"direct_action"`
	SortOrder    int    `json:"sort_order"`
}

type statusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) List(c *gin.Context) {
	if !h.available(c) {
		return
	}
	page, pageSize := parseAdminPagination(c)
	result, err := h.service.List(feed.ListInput{
		Page:     page,
		PageSize: pageSize,
		Locale:   c.Query("locale"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
		Admin:    true,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": result})
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok || !h.available(c) {
		return
	}
	entry, err := h.service.Get(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entry": entry}})
}

func (h *Handler) Create(c *gin.Context) {
	if !h.available(c) {
		return
	}
	var request entryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.Create(toEntryInput(request, currentUserID(c)))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": gin.H{"entry": entry}})
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok || !h.available(c) {
		return
	}
	var request entryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.Update(id, toEntryInput(request, currentUserID(c)))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entry": entry}})
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok || !h.available(c) {
		return
	}
	var request statusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.service.UpdateStatus(id, request.Status, currentUserID(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"entry": entry}})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok || !h.available(c) {
		return
	}
	if err := h.service.Delete(id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "workbench feed entry deleted"})
}

func (h *Handler) ProductOptions(c *gin.Context) {
	if !h.available(c) {
		return
	}
	page, pageSize := parseAdminPagination(c)
	options, total, err := h.service.ListProductOptions(feed.CatalogQuery{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"options": options,
			"pagination": gin.H{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": (int(total) + pageSize - 1) / pageSize,
			},
		},
	})
}

func (h *Handler) UploadMedia(c *gin.Context) {
	if !h.available(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadRequestBytes)
	if err := c.Request.ParseMultipartForm(multipartMemory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart upload"})
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
	media, err := h.service.UploadMedia(c.Request.Context(), currentUserID(c), file, c.PostForm("caption"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": gin.H{"media": media}})
}

func (h *Handler) available(c *gin.Context) bool {
	if h != nil && h.service != nil {
		return true
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "workbench feed is unavailable"})
	return false
}

func toEntryInput(request entryRequest, actorID uint) feed.EntryInput {
	input := feed.EntryInput{
		Locale:         request.Locale,
		Content:        request.Content,
		Tags:           request.Tags,
		Status:         request.Status,
		ActorID:        actorID,
		Media:          make([]feed.MediaInput, 0, len(request.Media)),
		TaggedProducts: make([]feed.TaggedProductInput, 0, len(request.TaggedProducts)),
	}
	if request.PublishedAt != nil {
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*request.PublishedAt)); err == nil {
			input.PublishedAt = &parsed
		}
	}
	for _, item := range request.Media {
		input.Media = append(input.Media, feed.MediaInput{
			FilePath: item.FilePath, Width: item.Width, Height: item.Height,
			FileSizeKB: item.FileSizeKB, Caption: item.Caption, SortOrder: item.SortOrder,
		})
	}
	for _, item := range request.TaggedProducts {
		input.TaggedProducts = append(input.TaggedProducts, feed.TaggedProductInput{
			ProductID: item.ProductID, VariantID: item.VariantID,
			DirectAction: item.DirectAction, SortOrder: item.SortOrder,
		})
	}
	return input
}

func parseAdminPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workbench feed entry ID"})
		return 0, false
	}
	return uint(id), true
}

func currentUserID(c *gin.Context) uint {
	return c.GetUint("user_id")
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, feed.ErrEntryNotFound), errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "workbench feed entry not found"})
	case errors.Is(err, feed.ErrInvalidEntry), errors.Is(err, feed.ErrInvalidStatus), errors.Is(err, feed.ErrInvalidDirectAction),
		errors.Is(err, feed.ErrMediaLimit), errors.Is(err, feed.ErrProductLimit), errors.Is(err, feed.ErrProductNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "workbench feed operation failed"})
	}
}
