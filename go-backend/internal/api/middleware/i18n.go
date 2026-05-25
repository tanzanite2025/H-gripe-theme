package middleware

import (
	"tanzanite/internal/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// I18n 国际化中间件
func I18n() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先级：
		// 1. URL路径 (/fr/products)
		// 2. Accept-Language Header
		// 3. Cookie
		// 4. 默认语言

		locale := "en"

		// 1. 从URL路径提取
		pathLocale := i18n.GetLocaleFromPath(c.Request.URL.Path)
		if pathLocale != "" && i18n.IsValidLocale(pathLocale) {
			locale = pathLocale
		} else {
			// 2. 从Header提取
			acceptLang := c.GetHeader("Accept-Language")
			if acceptLang != "" {
				headerLocale := i18n.NormalizeLocale(acceptLang)
				if i18n.IsValidLocale(headerLocale) {
					locale = headerLocale
				}
			}

			// 3. 从Cookie提取
			if cookieLocale, err := c.Cookie("locale"); err == nil {
				normalized := i18n.NormalizeLocale(cookieLocale)
				if i18n.IsValidLocale(normalized) {
					locale = normalized
				}
			}
		}

		// 存储到上下文
		c.Set("locale", locale)

		// 设置响应头
		c.Header("Content-Language", locale)

		c.Next()
	}
}

// GetLocale 从上下文获取语言
func GetLocale(c *gin.Context) string {
	if locale, exists := c.Get("locale"); exists {
		return locale.(string)
	}
	return "en"
}
