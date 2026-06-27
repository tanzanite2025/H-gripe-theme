package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ContentHandler struct {
	postService *service.PostService
}

func NewContentHandler(postService *service.PostService) *ContentHandler {
	return &ContentHandler{
		postService: postService,
	}
}

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

// ListPosts 获取文章列表
// GET /api/admin/content/posts
func (h *ContentHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	locale := c.Query("locale")
	search := c.Query("search")
	authorID := c.Query("author_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	posts, total, err := h.postService.ListAdmin(page, pageSize, status, locale, search, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetPost 获取文章详情
// GET /api/admin/content/posts/:id
func (h *ContentHandler) GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	foundPost, err := h.postService.GetAdminPost(uint(id))
	if err != nil {
		respondPostServiceError(c, err, "Failed to fetch post")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"post": foundPost,
	})
}

// CreatePost 创建文章
// POST /api/admin/content/posts
func (h *ContentHandler) CreatePost(c *gin.Context) {
	var req postCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, ok := currentAdminUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	newPost, err := h.postService.CreateAdminPost(service.PostCreateInput{
		Title:              req.Title,
		Slug:               req.Slug,
		Content:            req.Content,
		Excerpt:            req.Excerpt,
		Status:             req.Status,
		AuthorID:           userID,
		Locale:             req.Locale,
		FeaturedImg:        req.FeaturedImg,
		Tags:               req.Tags,
		MetaTitle:          req.MetaTitle,
		MetaDesc:           req.MetaDesc,
		MetaKeywords:       req.MetaKeywords,
		CanonicalURL:       req.CanonicalURL,
		TranslationGroupID: req.TranslationGroupID,
	})
	if err != nil {
		respondPostServiceError(c, err, "Failed to create post")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post created successfully",
		"post":    newPost,
	})
}

// UpdatePost 更新文章
// PUT /api/admin/content/posts/:id
func (h *ContentHandler) UpdatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req postUpdateRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var raw map[string]json.RawMessage
	if err := c.ShouldBindBodyWith(&raw, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, updateTranslationGroupID := raw["translation_group_id"]

	updatedPost, err := h.postService.UpdateAdminPost(uint(id), service.PostUpdateInput{
		Title:                    req.Title,
		Slug:                     req.Slug,
		Content:                  req.Content,
		Excerpt:                  req.Excerpt,
		Status:                   req.Status,
		Locale:                   req.Locale,
		FeaturedImg:              req.FeaturedImg,
		Tags:                     req.Tags,
		MetaTitle:                req.MetaTitle,
		MetaDesc:                 req.MetaDesc,
		MetaKeywords:             req.MetaKeywords,
		CanonicalURL:             req.CanonicalURL,
		TranslationGroupID:       req.TranslationGroupID,
		UpdateTranslationGroupID: updateTranslationGroupID,
	})
	if err != nil {
		respondPostServiceError(c, err, "Failed to update post")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    updatedPost,
	})
}

// DeletePost 删除文章
// DELETE /api/admin/content/posts/:id
func (h *ContentHandler) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.Delete(uint(id)); err != nil {
		respondPostServiceError(c, err, "Failed to delete post")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post deleted successfully",
	})
}

// UpdatePostStatus 更新文章状态
// PATCH /api/admin/content/posts/:id/status
func (h *ContentHandler) UpdatePostStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=draft published archived"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.postService.UpdateStatus(uint(id), req.Status); err != nil {
		respondPostServiceError(c, err, "Failed to update post status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post status updated successfully",
	})
}

// GetPostStats 获取文章统计
// GET /api/admin/content/posts/stats
func (h *ContentHandler) GetPostStats(c *gin.Context) {
	stats, err := h.postService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTranslations 获取文章的所有翻译版本
// GET /api/admin/content/posts/:id/translations
func (h *ContentHandler) GetTranslations(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	translations, err := h.postService.GetTranslations(uint(id))
	if err != nil {
		respondPostServiceError(c, err, "Failed to get translations")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"translations": translations,
	})
}

// BatchUpdateStatus 批量更新文章状态
// POST /api/admin/content/posts/batch-status
func (h *ContentHandler) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		PostIDs []uint `json:"post_ids" binding:"required,min=1"`
		Status  string `json:"status" binding:"required,oneof=draft published archived"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.postService.BatchUpdateStatus(req.PostIDs, req.Status)
	if err != nil {
		respondPostServiceError(c, err, "Failed to batch update post status")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch update completed",
		"updated": updated,
		"total":   len(req.PostIDs),
	})
}

// BatchDelete 批量删除文章
// POST /api/admin/content/posts/batch-delete
func (h *ContentHandler) BatchDelete(c *gin.Context) {
	var req struct {
		PostIDs []uint `json:"post_ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleted, err := h.postService.BatchDelete(req.PostIDs)
	if err != nil {
		respondPostServiceError(c, err, "Failed to batch delete posts")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.PostIDs),
	})
}
