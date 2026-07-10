package admin

import (
	"errors"
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type postCreateRequest struct {
	Title              string `json:"title" binding:"required"`
	Slug               string `json:"slug" binding:"required"`
	Content            string `json:"content"`
	Excerpt            string `json:"excerpt"`
	Status             string `json:"status" binding:"required,oneof=draft published archived"`
	Locale             string `json:"locale" binding:"required"`
	FeaturedImg        string `json:"featured_image"`
	Tags               string `json:"tags"`
	MetaTitle          string `json:"meta_title"`
	MetaDesc           string `json:"meta_description"`
	MetaKeywords       string `json:"meta_keywords"`
	CanonicalURL       string `json:"canonical_url"`
	TranslationGroupID *uint  `json:"translation_group_id"`
}

type postUpdateRequest struct {
	Title              *string `json:"title" binding:"omitempty,min=1"`
	Slug               *string `json:"slug" binding:"omitempty,min=1"`
	Content            *string `json:"content"`
	Excerpt            *string `json:"excerpt"`
	Status             *string `json:"status" binding:"omitempty,oneof=draft published archived"`
	Locale             *string `json:"locale" binding:"omitempty,min=1"`
	FeaturedImg        *string `json:"featured_image"`
	Tags               *string `json:"tags"`
	MetaTitle          *string `json:"meta_title"`
	MetaDesc           *string `json:"meta_description"`
	MetaKeywords       *string `json:"meta_keywords"`
	CanonicalURL       *string `json:"canonical_url"`
	TranslationGroupID *uint   `json:"translation_group_id"`
}

func respondPostServiceError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
	case errors.Is(err, service.ErrPostSlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Slug already exists for this locale"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
	}
}

func currentAdminUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	typedUserID, ok := userID.(uint)
	return typedUserID, ok
}
