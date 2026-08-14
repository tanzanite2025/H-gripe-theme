package showcase

import (
	"commerce-platform/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

const (
	showcaseMaxPage    = 100000
	showcaseMaxPerPage = 50
)

func (h *ShowcaseHandler) List(c *gin.Context) {
	kind := c.Query("type")
	if kind == "" {
		kind = "user"
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	page, perPage = normalizeShowcasePagination(page, perPage)

	items, err := h.service.ListPublic(kind, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]publicShowcaseItem, 0, len(items))
	for _, item := range items {
		imageReferences, err := decodePublicImageReferences(item.Images)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "showcase images are invalid"})
			return
		}
		galleryImages := make([]string, 0, len(imageReferences))
		for index := range imageReferences {
			galleryImages = append(galleryImages, h.publicImageURL(item.ID, index))
		}
		response = append(response, publicShowcaseItem{
			ID:            item.ID,
			Kind:          item.Kind,
			Title:         item.Title,
			Region:        item.Region,
			Location:      item.Location,
			Nickname:      item.Nickname,
			BikeModel:     item.BikeModel,
			Notes:         item.Notes,
			ProductRefs:   item.ProductRefs,
			GalleryImages: galleryImages,
			Status:        item.Status,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (h *ShowcaseHandler) ServePublicImageFile(c *gin.Context) {
	id, ok := parseShowcaseID(c)
	if !ok {
		return
	}
	imageIndex, ok := parseShowcaseImageIndex(c)
	if !ok {
		return
	}

	file, err := h.service.OpenPublicImageFile(c.Request.Context(), id, imageIndex)
	if err != nil {
		respondPublicShowcaseError(c, err)
		return
	}
	if strings.TrimSpace(file.RedirectURL) != "" {
		c.Header("Cache-Control", "public, max-age=300")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Redirect(http.StatusTemporaryRedirect, file.RedirectURL)
		return
	}
	if file.ReadCloser == nil {
		respondPublicShowcaseError(c, service.ErrShowcaseStorageUnavailable)
		return
	}
	defer func() { _ = file.ReadCloser.Close() }()

	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(http.StatusOK, file.Size, mimeType, file.ReadCloser, nil)
}

type publicShowcaseItem struct {
	ID            uint           `json:"id"`
	Kind          string         `json:"kind"`
	Title         string         `json:"title"`
	Region        string         `json:"region"`
	Location      string         `json:"location"`
	Nickname      string         `json:"nickname"`
	BikeModel     string         `json:"bike_model"`
	Notes         string         `json:"notes"`
	ProductRefs   datatypes.JSON `json:"product_refs"`
	GalleryImages []string       `json:"gallery_images"`
	Status        string         `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func decodePublicImageReferences(raw datatypes.JSON) ([]string, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return []string{}, nil
	}
	var references []string
	if err := json.Unmarshal(raw, &references); err != nil {
		return nil, err
	}
	return references, nil
}

func (h *ShowcaseHandler) publicImageURL(id uint, imageIndex int) string {
	path := fmt.Sprintf("/api/v1/showcase/%d/images/%d/file", id, imageIndex)
	return path
}

func respondPublicShowcaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrShowcaseNotFound),
		errors.Is(err, service.ErrShowcaseImageNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "showcase image not found"})
	case errors.Is(err, service.ErrShowcaseStorageUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "showcase image unavailable"})
	}
}

func parseShowcaseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return uint(id), true
}

func parseShowcaseImageIndex(c *gin.Context) (int, bool) {
	index, err := strconv.Atoi(c.Param("image_index"))
	if err != nil || index < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image index"})
		return 0, false
	}
	return index, true
}

func (h *ShowcaseHandler) AddComment(c *gin.Context) {
	var req struct {
		PhotoID  uint   `json:"photo_id" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Location string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "tpg_empty_comment"})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return
	}
	userID := userIDVal.(uint)

	author := "User"
	if usernameVal, ok := c.Get("username"); ok {
		author = usernameVal.(string)
	}

	comment, err := h.service.AddComment(req.PhotoID, userID, author, req.Content, req.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comment)
}

func (h *ShowcaseHandler) ListComments(c *gin.Context) {
	photoIDStr := c.Query("photo_id")
	photoID, err := strconv.Atoi(photoIDStr)
	if err != nil || photoID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo_id", "code": "tpg_invalid_photo_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	page, perPage = normalizeShowcasePagination(page, perPage)

	comments, err := h.service.ListComments(uint(photoID), page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type PhotoCommentResp struct {
		ID               uint   `json:"id"`
		Author           string `json:"author"`
		Content          string `json:"content"`
		DateGMT          string `json:"date_gmt"`
		DateGMTFormatted string `json:"dateGmtFormatted"`
		Location         string `json:"location"`
	}

	var res []PhotoCommentResp
	for _, c := range comments {
		dateStr := c.CreatedAt.Format("2006-01-02T15:04:05")
		res = append(res, PhotoCommentResp{
			ID:               c.ID,
			Author:           c.Author,
			Content:          c.Content,
			DateGMT:          dateStr,
			DateGMTFormatted: c.CreatedAt.Format("Jan 2, 2006"),
			Location:         c.Location,
		})
	}

	c.JSON(http.StatusOK, res)
}

func normalizeShowcasePagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if page > showcaseMaxPage {
		page = showcaseMaxPage
	}
	if perPage < 1 {
		perPage = 1
	}
	if perPage > showcaseMaxPerPage {
		perPage = showcaseMaxPerPage
	}
	return page, perPage
}
