package i18n

import (
	"net/http"
	"strconv"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler i18n API 处理器
type Handler struct {
	postService    *service.PostService
	sitemapService *service.SitemapService
}

// NewHandler 创建 i18n Handler
func NewHandler(postService *service.PostService, sitemapService *service.SitemapService) *Handler {
	return &Handler{
		postService:    postService,
		sitemapService: sitemapService,
	}
}

// Language 语言信息
type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	Enabled    bool   `json:"enabled"`
}

// SupportedLanguages 支持的语言列表（34种语言）
var SupportedLanguages = []Language{
	{Code: "en", Name: "English", NativeName: "English", Enabled: true},
	{Code: "zh", Name: "Chinese (Simplified)", NativeName: "简体中文", Enabled: true},
	{Code: "zh-TW", Name: "Chinese (Traditional)", NativeName: "繁體中文", Enabled: true},
	{Code: "es", Name: "Spanish", NativeName: "Español", Enabled: true},
	{Code: "fr", Name: "French", NativeName: "Français", Enabled: true},
	{Code: "de", Name: "German", NativeName: "Deutsch", Enabled: true},
	{Code: "ja", Name: "Japanese", NativeName: "日本語", Enabled: true},
	{Code: "ko", Name: "Korean", NativeName: "한국어", Enabled: true},
	{Code: "pt", Name: "Portuguese", NativeName: "Português", Enabled: true},
	{Code: "ru", Name: "Russian", NativeName: "Русский", Enabled: true},
	{Code: "ar", Name: "Arabic", NativeName: "العربية", Enabled: true},
	{Code: "it", Name: "Italian", NativeName: "Italiano", Enabled: true},
	{Code: "nl", Name: "Dutch", NativeName: "Nederlands", Enabled: true},
	{Code: "pl", Name: "Polish", NativeName: "Polski", Enabled: true},
	{Code: "tr", Name: "Turkish", NativeName: "Türkçe", Enabled: true},
	{Code: "vi", Name: "Vietnamese", NativeName: "Tiếng Việt", Enabled: true},
	{Code: "th", Name: "Thai", NativeName: "ไทย", Enabled: true},
	{Code: "id", Name: "Indonesian", NativeName: "Bahasa Indonesia", Enabled: true},
	{Code: "ms", Name: "Malay", NativeName: "Bahasa Melayu", Enabled: true},
	{Code: "hi", Name: "Hindi", NativeName: "हिन्दी", Enabled: true},
	{Code: "bn", Name: "Bengali", NativeName: "বাংলা", Enabled: true},
	{Code: "ta", Name: "Tamil", NativeName: "தமிழ்", Enabled: true},
	{Code: "te", Name: "Telugu", NativeName: "తెలుగు", Enabled: true},
	{Code: "mr", Name: "Marathi", NativeName: "मराठी", Enabled: true},
	{Code: "ur", Name: "Urdu", NativeName: "اردو", Enabled: true},
	{Code: "fa", Name: "Persian", NativeName: "فارسی", Enabled: true},
	{Code: "he", Name: "Hebrew", NativeName: "עברית", Enabled: true},
	{Code: "sv", Name: "Swedish", NativeName: "Svenska", Enabled: true},
	{Code: "no", Name: "Norwegian", NativeName: "Norsk", Enabled: true},
	{Code: "da", Name: "Danish", NativeName: "Dansk", Enabled: true},
	{Code: "fi", Name: "Finnish", NativeName: "Suomi", Enabled: true},
	{Code: "cs", Name: "Czech", NativeName: "Čeština", Enabled: true},
	{Code: "hu", Name: "Hungarian", NativeName: "Magyar", Enabled: true},
	{Code: "ro", Name: "Romanian", NativeName: "Română", Enabled: true},
}

// GetLanguages 获取支持的语言列表
// @Summary 获取支持的语言列表
// @Description 返回系统支持的所有语言
// @Tags i18n
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "languages: 语言列表"
// @Router /api/v1/i18n/languages [get]
func (h *Handler) GetLanguages(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"languages": SupportedLanguages,
		"total":     len(SupportedLanguages),
	})
}

// GetPostTranslations 获取文章的所有翻译版本
// @Summary 获取文章的所有翻译版本
// @Description 根据文章ID获取该文章的所有语言版本
// @Tags i18n
// @Accept json
// @Produce json
// @Param post_id path int true "文章ID"
// @Success 200 {object} map[string]interface{} "translations: 翻译列表"
// @Failure 400 {object} map[string]interface{} "error: 错误信息"
// @Failure 404 {object} map[string]interface{} "error: 错误信息"
// @Router /api/v1/i18n/translations/{post_id} [get]
func (h *Handler) GetPostTranslations(c *gin.Context) {
	// 获取文章ID
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	// 获取翻译版本
	translations, err := h.postService.GetTranslations(uint(postID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// 构建响应
	translationMap := make(map[string]interface{})
	for _, t := range translations {
		translationMap[t.Locale] = gin.H{
			"id":            t.ID,
			"title":         t.Title,
			"slug":          t.Slug,
			"locale":        t.Locale,
			"published_at":  t.PublishedAt,
			"url":           buildPostURL(t.Locale, t.Slug),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"post_id":      postID,
		"translations": translationMap,
		"count":        len(translations),
	})
}

// GetHreflangSitemap 获取 Hreflang Sitemap
// @Summary 获取 Hreflang Sitemap
// @Description 生成包含 hreflang 标签的 XML Sitemap
// @Tags i18n
// @Produce xml
// @Success 200 {string} string "XML Sitemap"
// @Failure 500 {object} map[string]interface{} "error: 错误信息"
// @Router /sitemap-hreflang.xml [get]
func (h *Handler) GetHreflangSitemap(c *gin.Context) {
	sitemap, err := h.sitemapService.GenerateHreflangSitemap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sitemap"})
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, sitemap)
}

// GetLocaleSitemap 获取指定语言的 Sitemap
// @Summary 获取指定语言的 Sitemap
// @Description 生成指定语言的 XML Sitemap
// @Tags i18n
// @Produce xml
// @Param locale path string true "语言代码"
// @Success 200 {string} string "XML Sitemap"
// @Failure 500 {object} map[string]interface{} "error: 错误信息"
// @Router /sitemap-{locale}.xml [get]
func (h *Handler) GetLocaleSitemap(c *gin.Context) {
	locale := c.Param("locale")

	sitemap, err := h.sitemapService.GenerateSimpleSitemap(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sitemap"})
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, sitemap)
}

// GetSitemapIndex 获取 Sitemap 索引
// @Summary 获取 Sitemap 索引
// @Description 生成 Sitemap 索引文件
// @Tags i18n
// @Produce xml
// @Success 200 {string} string "XML Sitemap Index"
// @Failure 500 {object} map[string]interface{} "error: 错误信息"
// @Router /sitemap.xml [get]
func (h *Handler) GetSitemapIndex(c *gin.Context) {
	// 获取启用的语言列表
	enabledLocales := make([]string, 0)
	for _, lang := range SupportedLanguages {
		if lang.Enabled {
			enabledLocales = append(enabledLocales, lang.Code)
		}
	}

	sitemap, err := h.sitemapService.GenerateSitemapIndex(enabledLocales)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sitemap index"})
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, sitemap)
}

// DetectLanguage 检测用户语言偏好
// @Summary 检测用户语言偏好
// @Description 根据 Accept-Language 头和 Cookie 检测用户语言偏好
// @Tags i18n
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "detected_locale: 检测到的语言"
// @Router /api/v1/i18n/detect [get]
func (h *Handler) DetectLanguage(c *gin.Context) {
	// 1. 优先从 Cookie 获取
	if locale, err := c.Cookie("locale"); err == nil && locale != "" {
		if isValidLocale(locale) {
			c.JSON(http.StatusOK, gin.H{
				"detected_locale": locale,
				"source":          "cookie",
			})
			return
		}
	}

	// 2. 从 Accept-Language 头获取
	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang != "" {
		locale := parseAcceptLanguage(acceptLang)
		if isValidLocale(locale) {
			c.JSON(http.StatusOK, gin.H{
				"detected_locale": locale,
				"source":          "header",
			})
			return
		}
	}

	// 3. 默认返回英文
	c.JSON(http.StatusOK, gin.H{
		"detected_locale": "en",
		"source":          "default",
	})
}

// SetLanguage 设置用户语言偏好
// @Summary 设置用户语言偏好
// @Description 设置用户的语言偏好到 Cookie
// @Tags i18n
// @Accept json
// @Produce json
// @Param locale body string true "语言代码"
// @Success 200 {object} map[string]interface{} "message: 成功信息"
// @Failure 400 {object} map[string]interface{} "error: 错误信息"
// @Router /api/v1/i18n/set-language [post]
func (h *Handler) SetLanguage(c *gin.Context) {
	var req struct {
		Locale string `json:"locale" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 验证语言代码
	if !isValidLocale(req.Locale) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported locale"})
		return
	}

	// 设置 Cookie (有效期 1 年)
	c.SetCookie("locale", req.Locale, 365*24*60*60, "/", "", false, false)

	c.JSON(http.StatusOK, gin.H{
		"message": "Language preference saved",
		"locale":  req.Locale,
	})
}

// 辅助函数

// buildPostURL 构建文章 URL
func buildPostURL(locale, slug string) string {
	if locale == "en" {
		return "/blog/" + slug
	}
	return "/" + locale + "/blog/" + slug
}

// isValidLocale 验证语言代码是否有效
func isValidLocale(locale string) bool {
	for _, lang := range SupportedLanguages {
		if lang.Code == locale && lang.Enabled {
			return true
		}
	}
	return false
}

// parseAcceptLanguage 解析 Accept-Language 头
func parseAcceptLanguage(acceptLang string) string {
	// 简单实现：取第一个语言代码
	// 完整实现应该解析 q 值并排序
	if len(acceptLang) >= 2 {
		locale := acceptLang[:2]
		return locale
	}
	return "en"
}
