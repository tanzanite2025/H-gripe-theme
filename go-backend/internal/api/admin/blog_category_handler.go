package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type BlogCategoryHandler struct {
	categoryService *service.BlogCategoryService
}

type blogCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
	Locale      string `json:"locale" binding:"required"`
	SortOrder   int    `json:"sort_order"`
}

func NewBlogCategoryHandler(categoryService *service.BlogCategoryService) *BlogCategoryHandler {
	return &BlogCategoryHandler{categoryService: categoryService}
}

func (h *BlogCategoryHandler) List(c *gin.Context) {
	categories, err := h.categoryService.List(c.Query("locale"))
	if err != nil {
		respondBlogCategoryError(c, err, "Failed to fetch blog categories")
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *BlogCategoryHandler) Create(c *gin.Context) {
	var req blogCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := h.categoryService.Create(service.BlogCategoryInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, Locale: req.Locale, SortOrder: req.SortOrder,
	})
	if err != nil {
		respondBlogCategoryError(c, err, "Failed to create blog category")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": category})
}

func (h *BlogCategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	var req blogCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	category, err := h.categoryService.Update(uint(id), service.BlogCategoryInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description, Locale: req.Locale, SortOrder: req.SortOrder,
	})
	if err != nil {
		respondBlogCategoryError(c, err, "Failed to update blog category")
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": category})
}

func (h *BlogCategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}
	if err := h.categoryService.Delete(uint(id)); err != nil {
		respondBlogCategoryError(c, err, "Failed to delete blog category")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Blog category deleted successfully"})
}

func respondBlogCategoryError(c *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrBlogCategoryNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Blog category not found"})
	case errors.Is(err, service.ErrBlogCategorySlugExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Category slug already exists for this locale"})
	case errors.Is(err, service.ErrBlogCategoryInUse):
		c.JSON(http.StatusConflict, gin.H{"error": "Category is still used by posts"})
	case errors.Is(err, service.ErrBlogCategoryInvalid), errors.Is(err, service.ErrUnsupportedLocale):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallbackMessage})
	}
}
