package seo

import (
	"net/http"
	"strconv"
	"time"

	postdomain "tanzanite/internal/domain/post"
	seodomain "tanzanite/internal/domain/seo"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ArticlesHandler struct {
	seoResources *service.SEOResourceService
	audit        seoAuditRecorder
}

func NewArticlesHandler(seoResources *service.SEOResourceService) *ArticlesHandler {
	return &ArticlesHandler{seoResources: seoResources}
}

func (h *ArticlesHandler) ConfigureAuditService(recorder seoAuditRecorder) {
	if h == nil {
		return
	}
	h.audit = recorder
}

func (h *ArticlesHandler) Get(c *gin.Context) {
	page, pageSize := resourcePagination(c)
	articles, total, err := h.seoResources.ListArticles(
		page,
		pageSize,
		c.Query("status"),
		c.Query("locale"),
		c.Query("search"),
	)
	if err != nil {
		writeResourceError(c, err, service.ErrPostNotFound)
		return
	}

	items := make([]gin.H, 0, len(articles))
	for _, article := range articles {
		items = append(items, gin.H{
			"id":               article.ID,
			"title":            article.Title,
			"slug":             article.Slug,
			"route_path":       seodomain.BuildArticleRoute(article.Locale, article.Slug, article.Tags).Path,
			"status":           article.Status,
			"locale":           article.Locale,
			"tags":             article.Tags,
			"meta_title":       article.MetaTitle,
			"meta_description": article.MetaDesc,
			"canonical_url":    article.CanonicalURL,
			"created_at":       article.CreatedAt,
			"published_at":     article.PublishedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *ArticlesHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	var request seodomain.ArticleResourceUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startedAt := time.Now().UTC()
	existing, _ := h.seoResources.GetArticle(uint(id))
	article, err := h.seoResources.UpdateArticle(uint(id), request)
	if err != nil {
		recordSEOAudit(h.audit, c, seoAuditEvent{
			StartedAt:    startedAt,
			Resource:     seoAuditResourceArticle,
			ResourceID:   uint(id),
			Status:       seoAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     articleAuditValue(existing),
			Changes: seoFieldChanges(map[string]interface{}{
				"meta_title":       request.MetaTitle != nil,
				"meta_description": request.MetaDescription != nil,
				"canonical_url":    request.CanonicalURL != nil,
			}),
		})
		writeResourceError(c, err, service.ErrPostNotFound)
		return
	}

	recordSEOAudit(h.audit, c, seoAuditEvent{
		StartedAt:  startedAt,
		Resource:   seoAuditResourceArticle,
		ResourceID: article.ID,
		Status:     seoAuditStatusOK,
		Changes: seoFieldChanges(map[string]interface{}{
			"meta_title":       request.MetaTitle != nil,
			"meta_description": request.MetaDescription != nil,
			"canonical_url":    request.CanonicalURL != nil,
		}),
		OldValue: articleAuditValue(existing),
		NewValue: articleAuditValue(article),
	})

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":               article.ID,
			"title":            article.Title,
			"slug":             article.Slug,
			"route_path":       seodomain.BuildArticleRoute(article.Locale, article.Slug, article.Tags).Path,
			"status":           article.Status,
			"locale":           article.Locale,
			"tags":             article.Tags,
			"meta_title":       article.MetaTitle,
			"meta_description": article.MetaDesc,
			"canonical_url":    article.CanonicalURL,
		},
	})
}

func articleAuditValue(article *postdomain.Post) map[string]string {
	if article == nil {
		return nil
	}
	return map[string]string{
		"meta_title":       article.MetaTitle,
		"meta_description": article.MetaDesc,
		"canonical_url":    article.CanonicalURL,
	}
}
