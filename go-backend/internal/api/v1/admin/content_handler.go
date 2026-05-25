package admin

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type ContentHandler struct {
	postRepo *repository.PostRepository
}

func NewContentHandler(postRepo *repository.PostRepository) *ContentHandler {
	return &ContentHandler{
		postRepo: postRepo,
	}
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

	posts, total, err := h.postRepo.FindAllWithFilters(page, pageSize, status, locale, search, authorID)
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

	foundPost, err := h.postRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// 获取翻译版本
	if foundPost.TranslationGroupID != nil {
		translations, _ := h.postRepo.FindByTranslationGroup(*foundPost.TranslationGroupID)
		foundPost.Translations = translations
	}

	c.JSON(http.StatusOK, gin.H{
		"post": foundPost,
	})
}

// CreatePost 创建文章
// POST /api/admin/content/posts
func (h *ContentHandler) CreatePost(c *gin.Context) {
	var req struct {
		Title              string  `json:"title" binding:"required"`
		Slug               string  `json:"slug" binding:"required"`
		Content            string  `json:"content"`
		Excerpt            string  `json:"excerpt"`
		Status             string  `json:"status" binding:"required,oneof=draft published archived"`
		Locale             string  `json:"locale" binding:"required"`
		FeaturedImg        string  `json:"featured_image"`
		Tags               string  `json:"tags"`
		MetaTitle          string  `json:"meta_title"`
		MetaDesc           string  `json:"meta_description"`
		MetaKeywords       string  `json:"meta_keywords"`
		CanonicalURL       string  `json:"canonical_url"`
		TranslationGroupID *uint   `json:"translation_group_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID作为作者
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 检查 slug 是否已存在
	existingPost, _ := h.postRepo.FindBySlug(req.Slug, req.Locale)
	if existingPost != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Slug already exists for this locale"})
		return
	}

	// 创建文章
	newPost := &post.Post{
		Title:              req.Title,
		Slug:               req.Slug,
		Content:            req.Content,
		Excerpt:            req.Excerpt,
		Status:             req.Status,
		AuthorID:           userID.(uint),
		Locale:             req.Locale,
		FeaturedImg:        req.FeaturedImg,
		Tags:               req.Tags,
		MetaTitle:          req.MetaTitle,
		MetaDesc:           req.MetaDesc,
		MetaKeywords:       req.MetaKeywords,
		CanonicalURL:       req.CanonicalURL,
		TranslationGroupID: req.TranslationGroupID,
	}

	// 如果状态是已发布，设置发布时间
	if req.Status == "published" {
		now := time.Now()
		newPost.PublishedAt = &now
	}

	if err := h.postRepo.Create(newPost); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
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

	var req struct {
		Title              string  `json:"title"`
		Slug               string  `json:"slug"`
		Content            string  `json:"content"`
		Excerpt            string  `json:"excerpt"`
		Status             string  `json:"status" binding:"omitempty,oneof=draft published archived"`
		Locale             string  `json:"locale"`
		FeaturedImg        string  `json:"featured_image"`
		Tags               string  `json:"tags"`
		MetaTitle          string  `json:"meta_title"`
		MetaDesc           string  `json:"meta_description"`
		MetaKeywords       string  `json:"meta_keywords"`
		CanonicalURL       string  `json:"canonical_url"`
		TranslationGroupID *uint   `json:"translation_group_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取现有文章
	existingPost, err := h.postRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// 更新字段
	if req.Title != "" {
		existingPost.Title = req.Title
	}
	if req.Slug != "" && req.Slug != existingPost.Slug {
		// 检查新 slug 是否已被使用
		slugPost, _ := h.postRepo.FindBySlug(req.Slug, existingPost.Locale)
		if slugPost != nil && slugPost.ID != existingPost.ID {
			c.JSON(http.StatusConflict, gin.H{"error": "Slug already exists"})
			return
		}
		existingPost.Slug = req.Slug
	}
	if req.Content != "" {
		existingPost.Content = req.Content
	}
	if req.Excerpt != "" {
		existingPost.Excerpt = req.Excerpt
	}
	if req.Status != "" {
		// 如果从非发布状态改为发布状态，设置发布时间
		if req.Status == "published" && existingPost.Status != "published" {
			now := time.Now()
			existingPost.PublishedAt = &now
		}
		existingPost.Status = req.Status
	}
	if req.Locale != "" {
		existingPost.Locale = req.Locale
	}
	existingPost.FeaturedImg = req.FeaturedImg
	existingPost.Tags = req.Tags
	existingPost.MetaTitle = req.MetaTitle
	existingPost.MetaDesc = req.MetaDesc
	existingPost.MetaKeywords = req.MetaKeywords
	existingPost.CanonicalURL = req.CanonicalURL
	existingPost.TranslationGroupID = req.TranslationGroupID

	if err := h.postRepo.Update(existingPost); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post updated successfully",
		"post":    existingPost,
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

	if err := h.postRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
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

	if err := h.postRepo.UpdateStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post status updated successfully",
	})
}

// GetPostStats 获取文章统计
// GET /api/admin/content/posts/stats
func (h *ContentHandler) GetPostStats(c *gin.Context) {
	stats, err := h.postRepo.GetStats()
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

	foundPost, err := h.postRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var translations []post.Post
	if foundPost.TranslationGroupID != nil {
		translations, err = h.postRepo.FindByTranslationGroup(*foundPost.TranslationGroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get translations"})
			return
		}
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

	updated := 0
	for _, id := range req.PostIDs {
		if err := h.postRepo.UpdateStatus(id, req.Status); err == nil {
			updated++
		}
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

	deleted := 0
	for _, id := range req.PostIDs {
		if err := h.postRepo.Delete(id); err == nil {
			deleted++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch delete completed",
		"deleted": deleted,
		"total":   len(req.PostIDs),
	})
}
