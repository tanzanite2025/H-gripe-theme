package v1

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type wpCompatFeaturedImage struct {
	URL    string `json:"url"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`
	Alt    string `json:"alt,omitempty"`
}

type wpCompatTranslation struct {
	ID   uint   `json:"id"`
	Slug string `json:"slug"`
}

type wpCompatPostSummary struct {
	ID            uint                           `json:"id"`
	Lang          string                         `json:"lang"`
	Group         string                         `json:"group"`
	Slug          string                         `json:"slug"`
	Title         string                         `json:"title"`
	Excerpt       string                         `json:"excerpt"`
	Date          string                         `json:"date"`
	FeaturedImage *wpCompatFeaturedImage         `json:"featuredImage"`
	Categories    []string                       `json:"categories"`
	Translations  map[string]wpCompatTranslation `json:"translations"`
}

type wpCompatPostDetail struct {
	wpCompatPostSummary
	ContentHTML  string `json:"contentHtml"`
	CanonicalURL string `json:"canonicalUrl"`
}

func registerWordPressCompatRoutes(r *gin.Engine, postService *service.PostService) {
	// Only expose compatibility routes for modules that have actually moved to Go.
	// Other WordPress plugin routes should continue to be served by WordPress until migrated.
	wp := r.Group("/wp-json/tanzanite/v1")
	{
		wp.GET("/posts", listCompatBlogPosts(postService))
		wp.GET("/post", getCompatBlogPost(postService))
		wp.GET("/translations", getCompatBlogTranslations(postService))
	}
}

func listCompatBlogPosts(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := compatLocale(c)
		status := c.DefaultQuery("status", "published")
		category := strings.ToLower(strings.TrimSpace(c.Query("category")))
		page := parseCompatInt(c, "page", 1)
		perPage := parseCompatInt(c, "per_page", parseCompatInt(c, "page_size", 5))

		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 5
		}

		fetchPage := page
		fetchSize := perPage
		if category != "" {
			fetchPage = 1
			fetchSize = 500
		}

		posts, total, err := postService.List(locale, status, fetchPage, fetchSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if category != "" {
			posts = filterCompatPostsByCategory(posts, category)
			total = int64(len(posts))
			posts = sliceCompatPosts(posts, page, perPage)
		}

		items := make([]wpCompatPostSummary, 0, len(posts))
		for _, item := range posts {
			items = append(items, makeCompatPostSummary(item, category, nil))
		}

		c.JSON(http.StatusOK, gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
			"items":    items,
		})
	}
}

func getCompatBlogPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := compatLocale(c)
		slug := strings.TrimSpace(c.Query("slug"))
		if slug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing slug"})
			return
		}

		item, err := postService.GetBySlug(slug, locale)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
			return
		}

		translations := loadCompatTranslations(postService, item)
		c.JSON(http.StatusOK, makeCompatPostDetail(*item, c.Query("category"), translations))
	}
}

func getCompatBlogTranslations(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		group := strings.TrimSpace(c.Query("group"))
		if group == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing group"})
			return
		}

		groupID, err := parseCompatGroupID(group)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"group":        group,
				"translations": map[string]wpCompatTranslation{},
			})
			return
		}

		posts, err := postService.GetTranslationsByGroup(groupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"group":        group,
			"translations": makeCompatTranslationMap(posts),
		})
	}
}

func compatLocale(c *gin.Context) string {
	if lang := strings.TrimSpace(c.Query("lang")); lang != "" {
		return lang
	}
	return middleware.GetLocale(c)
}

func parseCompatInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

func parseCompatGroupID(group string) (uint, error) {
	group = strings.TrimPrefix(group, "post-")
	group = strings.TrimPrefix(group, "grp-")
	value, err := strconv.ParseUint(group, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(value), nil
}

func sliceCompatPosts(posts []post.Post, page, perPage int) []post.Post {
	start := (page - 1) * perPage
	if start >= len(posts) {
		return []post.Post{}
	}
	end := start + perPage
	if end > len(posts) {
		end = len(posts)
	}
	return posts[start:end]
}

func filterCompatPostsByCategory(posts []post.Post, category string) []post.Post {
	filtered := make([]post.Post, 0, len(posts))
	for _, item := range posts {
		if hasCompatCategory(item, category) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func hasCompatCategory(item post.Post, category string) bool {
	if category == "" {
		return true
	}
	for _, itemCategory := range compatCategories(item, "") {
		if itemCategory == category {
			return true
		}
	}
	return false
}

func makeCompatPostSummary(
	item post.Post,
	fallbackCategory string,
	translations map[string]wpCompatTranslation,
) wpCompatPostSummary {
	if translations == nil {
		translations = map[string]wpCompatTranslation{
			item.Locale: {ID: item.ID, Slug: item.Slug},
		}
	}

	return wpCompatPostSummary{
		ID:            item.ID,
		Lang:          item.Locale,
		Group:         compatGroup(item),
		Slug:          item.Slug,
		Title:         item.Title,
		Excerpt:       item.Excerpt,
		Date:          compatPostDate(item),
		FeaturedImage: compatFeaturedImage(item),
		Categories:    compatCategories(item, fallbackCategory),
		Translations:  translations,
	}
}

func makeCompatPostDetail(
	item post.Post,
	fallbackCategory string,
	translations map[string]wpCompatTranslation,
) wpCompatPostDetail {
	return wpCompatPostDetail{
		wpCompatPostSummary: makeCompatPostSummary(item, fallbackCategory, translations),
		ContentHTML:         item.Content,
		CanonicalURL:        item.CanonicalURL,
	}
}

func compatGroup(item post.Post) string {
	if item.TranslationGroupID != nil {
		return fmt.Sprintf("%d", *item.TranslationGroupID)
	}
	return fmt.Sprintf("post-%d", item.ID)
}

func compatPostDate(item post.Post) string {
	if item.PublishedAt != nil {
		return item.PublishedAt.Format(time.RFC3339)
	}
	return item.CreatedAt.Format(time.RFC3339)
}

func compatFeaturedImage(item post.Post) *wpCompatFeaturedImage {
	if strings.TrimSpace(item.FeaturedImg) == "" {
		return nil
	}
	return &wpCompatFeaturedImage{URL: item.FeaturedImg}
}

func compatCategories(item post.Post, fallback string) []string {
	tags := strings.ToLower(item.Tags)
	categories := make([]string, 0, 2)
	for _, candidate := range []string{"news", "wheelsbuild"} {
		if fallback == candidate || strings.Contains(tags, candidate) {
			categories = append(categories, candidate)
		}
	}
	if len(categories) == 0 && fallback != "" {
		categories = append(categories, fallback)
	}
	return categories
}

func loadCompatTranslations(
	postService *service.PostService,
	item *post.Post,
) map[string]wpCompatTranslation {
	posts, err := postService.GetTranslations(item.ID)
	if err != nil || len(posts) == 0 {
		return map[string]wpCompatTranslation{
			item.Locale: {ID: item.ID, Slug: item.Slug},
		}
	}
	return makeCompatTranslationMap(posts)
}

func makeCompatTranslationMap(posts []post.Post) map[string]wpCompatTranslation {
	translations := make(map[string]wpCompatTranslation, len(posts))
	for _, item := range posts {
		translations[item.Locale] = wpCompatTranslation{
			ID:   item.ID,
			Slug: item.Slug,
		}
	}
	return translations
}
